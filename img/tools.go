package img

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type Exif struct {
	FileType             string  `json:"FileType"`
	FileTypeExtension    string  `json:"FileTypeExtension"`
	MIMEType             string  `json:"MIMEType"`
	Make                 string  `json:"Make"`
	Model                string  `json:"Model"`
	Orientation          string  `json:"Orientation"`
	Rotate               string  `json:"Rotate"`
	Software             string  `json:"Software"`
	ExposureTime         string  `json:"ExposureTime"`
	FNumber              float64 `json:"FNumber"`
	ExposureProgram      string  `json:"ExposureProgram"`
	ISO                  int     `json:"ISO"`
	ExposureCompensation string  `json:"ExposureCompensation"`
	MeteringMode         string  `json:"MeteringMode"`
	Flash                string  `json:"Flash"`
	FocalLength          string  `json:"FocalLength"`
	ExposureMode         string  `json:"ExposureMode"`
	WhiteBalance         string  `json:"WhiteBalance"`
	LensModel            string  `json:"LensModel"`
	ImageWidth           int     `json:"ImageWidth"`
	ImageHeight          int     `json:"ImageHeight"`
	GPSAltitude          string  `json:"GPSAltitude"`
	GPSLatitude          string  `json:"GPSLatitude"`
	GPSLongitude         string  `json:"GPSLongitude"`
	UserComment          string  `json:"UserComment"`
	Duration             string  `json:"Duration"`
	ContentIdentifier    string  `json:"ContentIdentifier"`
	DateTimeOriginal     string  `json:"DateTimeOriginal"`
	CreationDate         string  `json:"CreationDate"`
	CreateDate           string  `json:"CreateDate"`
	ModifyDate           string  `json:"ModifyDate"`
	FileModifyDate       string  `json:"FileModifyDate"`
}

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

func ExtractExif(
	srcPath string,
	exifOutput *Exif,
) error {
	cmd := exec.Command(
		"exiftool", "-json", "-a", "-u", srcPath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("exiftool failed: %v\n: %s", err, string(output))
	}

	log.Println("executed successful:", string(output))

	var exifData []Exif
	if err := json.Unmarshal(output, &exifData); err != nil {
		log.Printf("Failed to parse exif json: %v", err)
		return err
	}
	*exifOutput = exifData[0]
	return nil
}
