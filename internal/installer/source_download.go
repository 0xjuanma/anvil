/*
Copyright © 2022 Juanma Roca juanmaxroca@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package installer

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/utils"
)

// downloadFile downloads a file from URL to a temporary location
func downloadFile(fileURL, appName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "anvil-cli/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	homeDir, _ := system.HomeDir()
	// Organize downloads by app name: ~/Downloads/anvil-downloads/{appName}/
	downloadsDir := filepath.Join(homeDir, constants.DownloadsDirName, constants.AnvilDownloadsSubdir, appName)
	if err := utils.EnsureDirectory(downloadsDir); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	fileName := getFileNameFromURL(fileURL, appName)
	filePath := filepath.Join(downloadsDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// getFileNameFromURL extracts filename from URL or uses app name
func getFileNameFromURL(fileURL, appName string) string {
	parsedURL, err := url.Parse(fileURL)
	if err == nil && parsedURL.Path != "" {
		fileName := filepath.Base(parsedURL.Path)
		if fileName != "" && fileName != "/" {
			return fileName
		}
	}

	ext := getExtensionFromURL(fileURL)
	return fmt.Sprintf("%s%s", appName, ext)
}

// getExtensionFromURL tries to detect file extension from URL
func getExtensionFromURL(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err == nil {
		path := strings.ToLower(parsedURL.Path)
		for _, ext := range constants.SupportedFileExtensions {
			if strings.HasSuffix(path, ext) {
				return ext
			}
		}
	}
	return constants.ExtDefault
}
