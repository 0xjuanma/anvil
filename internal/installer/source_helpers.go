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
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xjuanma/anvil/internal/utils"
)

// ExtractionSucceededError indicates that extraction succeeded but installation failed
// This allows us to provide helpful feedback about where the app was extracted
type ExtractionSucceededError struct {
	ExtractDir string
	AppName    string
	Reason     string
}

func (e *ExtractionSucceededError) Error() string {
	return fmt.Sprintf("extraction succeeded but installation failed: %s", e.Reason)
}

// handleExtractedContentsMacOS handles extracted contents on macOS
// Returns ExtractionSucceededError if extraction succeeded but moving to Applications failed
func handleExtractedContentsMacOS(extractDir, appName string) error {
	appPath := findAppInDirectory(extractDir, appName)
	if appPath == "" {
		// Extraction succeeded but we can't find the .app
		return &ExtractionSucceededError{
			ExtractDir: extractDir,
			AppName:    appName,
			Reason:     "failed to find .app in extracted contents",
		}
	}

	applicationsDir, err := ensureApplicationsDirectory()
	if err != nil {
		// Extraction succeeded but we can't create Applications directory
		return &ExtractionSucceededError{
			ExtractDir: extractDir,
			AppName:    appName,
			Reason:     fmt.Sprintf("failed to create Applications directory: %v", err),
		}
	}

	appNameFromPath := filepath.Base(appPath)
	destPath := filepath.Join(applicationsDir, appNameFromPath)

	if err := copyAppToApplications(appPath, destPath); err != nil {
		// Extraction succeeded but copying to Applications failed
		return &ExtractionSucceededError{
			ExtractDir: extractDir,
			AppName:    appName,
			Reason:     fmt.Sprintf("failed to copy application to Applications: %v", err),
		}
	}

	return nil
}

// handleExtractedContentsLinux handles extracted contents on Linux
func handleExtractedContentsLinux(extractDir, appName string) error {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extract directory: %w", err)
	}

	if len(entries) == 1 && entries[0].IsDir() {
		appDir := filepath.Join(extractDir, entries[0].Name())
		destDir, err := ensureLinuxApplicationsDirectory(entries[0].Name())
		if err != nil {
			return err
		}
		return utils.CopyDirectorySimple(appDir, destDir)
	}

	destDir, err := ensureLinuxApplicationsDirectory(appName)
	if err != nil {
		return err
	}

	return utils.CopyDirectorySimple(extractDir, destDir)
}

// extractMountPath extracts the mount path from hdiutil output
func extractMountPath(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/Volumes/") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "/Volumes/") {
					return part
				}
			}
		}
	}
	return ""
}

// findAppInDirectory searches for .app bundle in directory
func findAppInDirectory(dir, appName string) string {
	var foundApp string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.HasSuffix(path, ".app") {
			baseName := strings.ToLower(strings.TrimSuffix(filepath.Base(path), ".app"))
			searchName := strings.ToLower(appName)
			if baseName == searchName || strings.Contains(baseName, searchName) {
				foundApp = path
				return filepath.SkipDir
			}
		}
		return nil
	})
	return foundApp
}
