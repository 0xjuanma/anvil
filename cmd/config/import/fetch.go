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

package importcmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/0xjuanma/anvil/internal/constants"
	"gopkg.in/yaml.v2"
)

// fetchFile downloads a file from URL or copies from local path to a temporary file.
func fetchFile(sourcePath string) (string, func(), error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if isURL(sourcePath) {
		return fetchFromURL(ctx, sourcePath)
	}

	// Handle local file
	return fetchFromLocal(sourcePath)
}

// isURL checks if the given string is a URL.
func isURL(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}

// fetchFromURL downloads file from URL to a temporary file.
func fetchFromURL(ctx context.Context, fileURL string) (string, func(), error) {
	// Create HTTP request with context
	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return "", nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set user agent
	req.Header.Set("User-Agent", "anvil-cli/1.0")

	// Execute request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "anvil-import-*.yaml")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}

	// Copy content to temporary file
	_, err = io.Copy(tempFile, resp.Body)
	tempFile.Close()
	if err != nil {
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	cleanup := func() {
		os.Remove(tempFile.Name())
	}

	return tempFile.Name(), cleanup, nil
}

// fetchFromLocal copies local file to temporary file for consistent handling.
func fetchFromLocal(filePath string) (string, func(), error) {
	// Validate file exists
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			return "", nil, fmt.Errorf("file does not exist: %s", filePath)
		}
		return "", nil, fmt.Errorf("cannot access file: %w", err)
	}

	// Create temporary file
	tempFile, err := os.CreateTemp("", "anvil-import-*.yaml")
	if err != nil {
		return "", nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	tempFile.Close()

	// Copy file content
	sourceData, err := os.ReadFile(filePath)
	if err != nil {
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to read source file: %w", err)
	}

	if err := os.WriteFile(tempFile.Name(), sourceData, constants.FilePerm); err != nil {
		os.Remove(tempFile.Name())
		return "", nil, fmt.Errorf("failed to write temp file: %w", err)
	}

	cleanup := func() {
		os.Remove(tempFile.Name())
	}

	return tempFile.Name(), cleanup, nil
}

// parseImportFile parses the import file and extracts only group data.
func parseImportFile(filePath string) (*ImportConfig, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read import file: %w", err)
	}

	// Parse as generic map first to extract only groups
	var rawData map[string]interface{}
	if err := yaml.Unmarshal(data, &rawData); err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}

	// Extract only groups section
	groupsData, exists := rawData["groups"]
	if !exists {
		return &ImportConfig{Groups: make(map[string][]string)}, nil
	}

	// Convert to proper structure
	groupsMap, ok := groupsData.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("groups section has invalid format")
	}

	importConfig := &ImportConfig{
		Groups: make(map[string][]string),
	}

	for groupName, groupTools := range groupsMap {
		groupNameStr, ok := groupName.(string)
		if !ok {
			continue // Skip invalid group names
		}

		toolsList, ok := groupTools.([]interface{})
		if !ok {
			continue // Skip invalid tool lists
		}

		var tools []string
		for _, tool := range toolsList {
			if toolStr, ok := tool.(string); ok {
				tools = append(tools, toolStr)
			}
		}

		if len(tools) > 0 {
			importConfig.Groups[groupNameStr] = tools
		}
	}

	return importConfig, nil
}
