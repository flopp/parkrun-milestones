package file

import (
	"os"
	"time"
)

func GetMtime(filePath string) (time.Time, error) {
	stat, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, err
	}

	return stat.ModTime(), nil
}
