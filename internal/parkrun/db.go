package parkrun

import (
	"fmt"
	"os"
	"path"
	"time"

	download "github.com/flopp/parkrun-milestones/internal/download"
	file "github.com/flopp/parkrun-milestones/internal/file"
)

var MaxFileAge time.Duration = 24 * time.Hour

func CachePath(format string, a ...any) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(base, "parkrun-milestones", fmt.Sprintf(format, a...)), nil
}

func DownloadAndRead(url string, fileName string) (string, error) {
	filePath, err := CachePath(fileName)
	if err != nil {
		return "", err
	}

	if err := download.DownloadFile(url, filePath, MaxFileAge); err != nil {
		return "", err
	}

	buf, err := file.ReadFile(filePath)
	if err != nil {
		return "", err
	}

	return buf, err
}
