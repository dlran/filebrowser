package img

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"os/exec"
	"path/filepath"
)

func CopyExif(srcPath, dstPath string) error {
	cmd := exec.Command("exiftool", "-overwrite_original", "-TagsFromFile", srcPath, "-all:all", "-ThumbnailImage=", dstPath)

	var outBuf bytes.Buffer
	cmd.Stdout = &outBuf
	cmd.Stderr = &outBuf

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("exiftool failed: %v\n: %s", err, outBuf.String())
	}

	log.Println("copy exif successful:", outBuf.String())

	return nil
}

func ExtractFrame(srcPath, fps string) error {
	dir := filepath.Dir(srcPath)
	fileNameWithExt := filepath.Base(srcPath)

	fileName := strings.TrimSuffix(fileNameWithExt, filepath.Ext(fileNameWithExt))

	newDirPath := filepath.Join(dir, fileName)

	err := os.MkdirAll(newDirPath, 0755)
	if err != nil {
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

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("ffmpeg failed: %v\n: %s", err, outBuf.String())
	}

	log.Println("ffmpeg successful:", outBuf.String())

	return nil
}

