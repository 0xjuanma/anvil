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

package github

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/utils"
)

// copyConfigToRepo copies a file or directory to the repository
func (gc *GitHubClient) copyConfigToRepo(sourcePath, targetDir string) error {
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return fmt.Errorf("failed to stat source path %s: %w", sourcePath, err)
	}

	if sourceInfo.IsDir() {
		// Copy directory contents to target directory
		return gc.copyDirectoryContents(sourcePath, targetDir)
	} else {
		// Copy single file to target directory
		fileName := filepath.Base(sourcePath)
		targetFile := filepath.Join(targetDir, fileName)
		return utils.CopyFileSimple(sourcePath, targetFile)
	}
}

// copyDirectoryContents recursively copies directory contents
func (gc *GitHubClient) copyDirectoryContents(sourceDir, targetDir string) error {
	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Get relative path
		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetDir, relPath)

		if info.IsDir() {
			return os.MkdirAll(targetPath, info.Mode())
		} else {
			return utils.CopyFileSimple(path, targetPath)
		}
	})
}

// getCommittedFiles returns a list of files that were committed in the target directory
func (gc *GitHubClient) getCommittedFiles(targetDir, appName string) ([]string, error) {
	var files []string

	err := filepath.Walk(targetDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			// Get relative path from the repo root
			relPath, err := filepath.Rel(gc.LocalPath, path)
			if err != nil {
				return err
			}
			files = append(files, relPath)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk target directory: %w", err)
	}

	if len(files) == 0 {
		// Fallback to just showing the app directory
		files = []string{fmt.Sprintf("%s/", appName)}
	}

	return files, nil
}
