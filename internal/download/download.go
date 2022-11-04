package download

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	file "github.com/flopp/parkrun-milestones/internal/file"
)

func download(url string, filePath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}
	req.Header.Add("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/104.0.0.0 Safari/537.36")
	client := &http.Client{}
	response, err := client.Do(req)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	statusOK := response.StatusCode >= 200 && response.StatusCode < 300
	if !statusOK {
		return fmt.Errorf("Non-OK HTTP status: %d", response.StatusCode)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0770); err != nil {
		return err
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, response.Body)
	return err
}

func DownloadFileMaxMtime(url string, filePath string, maxMtime time.Time) error {
	mtime, err := file.GetMtime(filePath)
	if err == nil {
		if mtime.After(maxMtime) {
			return nil
		}
	}

	return download(url, filePath)
}

func DownloadFile(url string, filePath string, maxAge time.Duration) error {
	if mtime, err := file.GetMtime(filePath); err == nil && mtime.After(time.Now().Add(-maxAge)) {
		return nil
	}

	return download(url, filePath)
}
