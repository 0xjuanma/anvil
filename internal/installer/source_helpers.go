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

// handleExtractedContentsMacOS handles extracted contents on macOS
func handleExtractedContentsMacOS(extractDir, appName string) error {
	appPath := findAppInDirectory(extractDir, appName)
	if appPath == "" {
		return fmt.Errorf("failed to find .app in extracted contents")
	}

	applicationsDir, err := ensureApplicationsDirectory()
	if err != nil {
		return err
	}

	appNameFromPath := filepath.Base(appPath)
	destPath := filepath.Join(applicationsDir, appNameFromPath)

	return copyAppToApplications(appPath, destPath)
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
