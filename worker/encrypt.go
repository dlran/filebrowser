package worker

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/hibiken/asynq"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
)

type encryptParams struct {
	Src      []string `json:"src"`
	Password string   `json:"password"`
	Root     string   `json:"root"`
	Mode     bool     `json:"mode"`
}

func encryptFileHandler(ctx context.Context, task *asynq.Task) error {
	taskID, ok := asynq.GetTaskID(ctx)
	if !ok {
		log.Println("failed to get task id")
		return nil
	}
	log.Printf("Processing task ID:%s", taskID)

	var payload encryptParams
	err := json.Unmarshal(task.Payload(), &payload)
	if err != nil {
		log.Fatalf("Error unmarshaling: %v", err)
	}

	for _, item := range payload.Src {
		srcPath := filepath.Join(payload.Root, item)
		password := payload.Password
		encryptMode := payload.Mode
		isEncryptFile := filepath.Ext(srcPath) == ".enc"
		if (isEncryptFile && encryptMode) || (!isEncryptFile && !encryptMode) {
			log.Println("Mode invalid")
			continue
		}

		var mode, salt, outputPath string
		if encryptMode {
			mode = "-e"
			salt = " -salt"
			outputPath = srcPath + ".enc"
		} else {
			mode = "-d"
			salt = ""
			outputPath = strings.TrimSuffix(srcPath, ".enc")
		}

		shellStr := fmt.Sprintf(`
            set -euo pipefail
            in="%s"
            out="%s"

            if [ -e "$out" ]; then
              echo "Error: file already exists: $out" >&2
              exit 1
            fi

            openssl enc %s -aes-256-cbc -pbkdf2 -iter 1000000%s -in "$in" -out "$out" -k '%s'

            rm -- "$in"
            `,
			srcPath, outputPath, mode, salt, password)
		// re := regexp.MustCompile(`-k '\w+'`)
		// pcmd := re.ReplaceAllString(shellStr, "-k xxxxxx")
		// log.Println(pcmd)
		log.Println("Processing", srcPath)

		cmd := exec.Command("bash", "-c", shellStr)

		var outBuf bytes.Buffer
		cmd.Stdout = &outBuf
		cmd.Stderr = &outBuf

		if err := cmd.Run(); err != nil {
			log.Printf("shell script failed: %v\n: %s", err, outBuf.String())
			return err
		}

		log.Printf("%s successful: %s\n", srcPath, outBuf.String())
	}
	res := fmt.Sprintf("%d files complete", len(payload.Src))
	if _, err := task.ResultWriter().Write([]byte(res)); err != nil {
		return err
	}
	return nil
}
