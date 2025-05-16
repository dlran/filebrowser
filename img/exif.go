package img

import (
	"fmt"
	"log"
	"bytes"
	"os/exec"
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

	log.Println("Copy Exif Successful:", outBuf.String())

	return nil
}

