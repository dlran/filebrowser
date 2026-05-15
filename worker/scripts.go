package worker

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/hibiken/asynq"
)

type ScriptTaskPayload struct {
	Src  []string `json:"src"`
	Root string   `json:"root"`
}

func executeScriptTaskHandler(ctx context.Context, task *asynq.Task) error {
	var payload ScriptTaskPayload

	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		return err
	}

	for _, scriptRelativePath := range payload.Src {
		scriptPath := filepath.Join(payload.Root, scriptRelativePath)

		ext := strings.ToLower(filepath.Ext(scriptPath))

		if ext != ".py" {
			log.Printf("skip unsupported script: %s", scriptPath)
			continue
		}

		cmd, err := buildScriptCommand(ctx, scriptPath)
		if err != nil {
			log.Printf("build command failed: path=%s err=%v", scriptPath, err)
			continue
		}

		err = runScriptCommandRealtimeLog(ctx, cmd, scriptPath)
		if err != nil {

			// inspector.CancelProcessing()
			// 导致的主动取消
			if errors.Is(ctx.Err(), context.Canceled) {
				log.Printf(
					"script canceled: path=%s",
					scriptPath,
				)

				return nil
			}

			log.Printf(
				"script execute failed: path=%s err=%v",
				scriptPath,
				err,
			)

			continue
		}
	}

	res := fmt.Sprintf("%d files complete", len(payload.Src))

	if _, err := task.ResultWriter().Write([]byte(res)); err != nil {
		return err
	}

	return nil
}

func buildScriptCommand(
	ctx context.Context,
	scriptPath string,
) (*exec.Cmd, error) {

	ext := strings.ToLower(filepath.Ext(scriptPath))

	switch ext {

	case ".py":

		cmd := exec.Command(
			"python3",
			"-u",
			scriptPath,
		)

		// 创建独立进程组
		// 后续可终止整个子进程树
		cmd.SysProcAttr = &syscall.SysProcAttr{
			Setpgid: true,
		}

		return cmd, nil
	}

	return nil, exec.ErrNotFound
}

func runScriptCommandRealtimeLog(
	ctx context.Context,
	cmd *exec.Cmd,
	scriptPath string,
) error {

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(2)

	go streamLog(
		&wg,
		stdoutPipe,
		scriptPath,
		"stdout",
	)

	go streamLog(
		&wg,
		stderrPipe,
		scriptPath,
		"stderr",
	)

	waitDone := make(chan error, 1)

	go func() {
		waitDone <- cmd.Wait()
	}()

	select {

	case err := <-waitDone:

		wg.Wait()

		return err

	case <-ctx.Done():

		log.Printf(
			"script cancel signal received: %s",
			scriptPath,
		)

		// 优雅退出
		err = syscall.Kill(
			-cmd.Process.Pid,
			syscall.SIGTERM,
		)

		if err != nil {
			log.Printf(
				"send SIGTERM failed: path=%s err=%v",
				scriptPath,
				err,
			)
		}

		select {

		// Python 已自行退出
		case err := <-waitDone:

			wg.Wait()

			return err

		// 超时强制 kill
		case <-time.After(10 * time.Second):

			log.Printf(
				"force kill script: %s",
				scriptPath,
			)

			err = syscall.Kill(
				-cmd.Process.Pid,
				syscall.SIGKILL,
			)

			if err != nil {
				log.Printf(
					"send SIGKILL failed: path=%s err=%v",
					scriptPath,
					err,
				)
			}

			wg.Wait()

			return ctx.Err()
		}
	}
}

func streamLog(
	wg *sync.WaitGroup,
	reader io.Reader,
	scriptPath string,
	streamType string,
) {
	defer wg.Done()

	scanner := bufio.NewScanner(reader)

	for scanner.Scan() {

		log.Printf(
			"[script][%s][%s] %s",
			streamType,
			scriptPath,
			scanner.Text(),
		)
	}

	if err := scanner.Err(); err != nil {

		log.Printf(
			"log stream error: path=%s stream=%s err=%v",
			scriptPath,
			streamType,
			err,
		)
	}
}
