package img

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"os/exec"
	"path/filepath"
	"strings"
    "regexp"
)

func CopyExif(srcPath, dstPath string) error {
    re := regexp.MustCompile(`(?i)\.(mp4|mov)$`)
    var commands string
    if re.MatchString(dstPath) {
        commands = fmt.Sprintf("exiftool -overwrite_original -TagsFromFile %s -DateTimeOriginal -GPS* %s", srcPath, dstPath)
    } else {
        commands = fmt.Sprintf("exiftool -overwrite_original -TagsFromFile %s -all:all -ThumbnailImage= %s && exiftool -overwrite_original -ContentIdentifier= %s", srcPath, dstPath, dstPath)
    }
	commandList := strings.Split(commands, "&&")

    for _, cmdStr := range commandList {
        cmdStr = strings.TrimSpace(cmdStr)
		if cmdStr == "" {
			continue
		}
        parts := strings.Fields(cmdStr)
		if len(parts) == 0 {
			continue
		}

        name := parts[0]
        args := parts[1:]
        cmd := exec.Command(name, args...)

        var outBuf bytes.Buffer
        cmd.Stdout = &outBuf
        cmd.Stderr = &outBuf

        err := cmd.Run()
        if err != nil {
            return fmt.Errorf("error executing: %v\n: %s", err, outBuf.String())
        }

        log.Println("executed successful:", outBuf.String())
    }

	return nil
}

func ExtractFrame(srcPath, fps string) error {
	dir := filepath.Dir(srcPath)
	fileNameWithExt := filepath.Base(srcPath)

	fileName := strings.TrimSuffix(fileNameWithExt, filepath.Ext(fileNameWithExt))

	newDirPath := filepath.Join(dir, fileName)

	if err := os.MkdirAll(newDirPath, 0755); err != nil {
		return fmt.Errorf("Mkdir failed: %v", err)
	}

	outputFilePath := filepath.Join(newDirPath, fmt.Sprintf("%s_%%04d.jpg", fileName))

	if fps == "" {
		fps = "1"
	}

	cmd := exec.Command("ffmpeg", "-i", srcPath,
		"-vf",
		fmt.Sprintf("fps=%s", fps),
		outputFilePath,
	)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg failed: %v\n: %s", err, outBuf.String())
	}

	log.Println("ffmpeg successful:", outBuf.String())

	return nil
}

