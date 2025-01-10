package requestsgo

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Download fetches data from a remote source and stores it in a file
func (c *Client) Download(url string, method string, body io.Reader, dst string) error {
	if c.cfg.CreateDirs {
		dirs := filepath.Base(dst)
		if err := os.MkdirAll(dirs, os.ModePerm); err != nil {
			return fmt.Errorf("creating dirs failed: %w", err)
		}
	}

	resp, err := c.Request(url, method, body)
	if err != nil {
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	out, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("opening dst failed: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("copying data failed: %w", err)
	}

	return nil
}

// GetDownload wraps Download, simplifying sending GET requests
func (c *Client) GetDownload(url string, dst string) error {
	return c.Download(url, http.MethodGet, nil, dst)
}
