package download

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	file "github.com/flopp/parkrun-milestones/internal/file"
)

func AlwaysDownload(url string, filePath string) error {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

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

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("reading response body from '%s': %w", url, err)
	}
	// A connection hiccup can yield a 200 with no/truncated body; refuse it
	// rather than letting it silently overwrite a good cache entry.
	if len(body) == 0 {
		return fmt.Errorf("empty response body from '%s'", url)
	}

	if err := os.MkdirAll(filepath.Dir(filePath), 0770); err != nil {
		return err
	}

	// Write to a temp file and rename into place so a failed/partial write
	// can never leave a truncated file at filePath.
	tmpFile, err := os.CreateTemp(filepath.Dir(filePath), filepath.Base(filePath)+".tmp*")
	if err != nil {
		return err
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath) // no-op once renamed below

	if _, err := tmpFile.Write(body); err != nil {
		tmpFile.Close()
		return err
	}
	if err := tmpFile.Close(); err != nil {
		return err
	}

	return os.Rename(tmpPath, filePath)
}

func DownloadFileMaxMtime(url string, filePath string, maxMtime time.Time) error {
	mtime, err := file.GetMtime(filePath)
	if err == nil {
		if mtime.After(maxMtime) {
			return nil
		}
	}

	return AlwaysDownload(url, filePath)
}

func DownloadFile(url string, filePath string, maxAge time.Duration) error {
	if mtime, err := file.GetMtime(filePath); err == nil && mtime.After(time.Now().Add(-maxAge)) {
		return nil
	}

	return AlwaysDownload(url, filePath)
}
