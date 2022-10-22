package file

import "os"

func ReadFile(filePath string) (string, error) {
	buf, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}
