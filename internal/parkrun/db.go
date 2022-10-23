package parkrun

import (
	"fmt"
	"os"
	"path"
	"time"
)

var MaxFileAge time.Duration = 24 * time.Hour

func CachePath(format string, a ...any) (string, error) {
	base, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	return path.Join(base, "parkrun-milestones", fmt.Sprintf(format, a...)), nil
}
