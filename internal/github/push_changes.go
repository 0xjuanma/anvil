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
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xjuanma/palantir"
)

// checkForChanges determines if there are changes to push and handles new vs existing apps
func (gc *GitHubClient) checkForChanges(ctx context.Context, appName, configPath, targetPath string) (bool, error) {
	repoTargetPath := filepath.Join(gc.LocalPath, targetPath)

	// Check if target exists in repo
	if _, err := os.Stat(repoTargetPath); os.IsNotExist(err) {
		return gc.handleNewApp(appName, configPath)
	}

	// Target exists in repo - check for changes
	return gc.handleExistingApp(appName, configPath, targetPath)
}

// handleNewApp handles the case where this is a new app not yet in the repository
func (gc *GitHubClient) handleNewApp(appName, configPath string) (bool, error) {
	output := palantir.GetGlobalOutputHandler()

	// Verify the local path actually exists and has content
	localInfo, err := os.Stat(configPath)
	if err != nil {
		return false, fmt.Errorf("local config path is invalid: %w", err)
	}

	if localInfo.IsDir() {
		// Check if directory has files
		entries, err := os.ReadDir(configPath)
		if err == nil && len(entries) > 0 {
			output.PrintInfo("New app '%s' detected - will be added to repository", appName)
			return true, nil
		}
	} else if localInfo.Size() > 0 {
		output.PrintInfo("New app '%s' detected - will be added to repository", appName)
		return true, nil
	}

	// No content to push
	output.PrintSuccess("Configuration is up-to-date!")
	output.PrintInfo("Local %s configs match the remote repository.", appName)
	output.PrintInfo("No changes to push.")
	return false, nil
}

// handleExistingApp handles the case where the app already exists in the repository
func (gc *GitHubClient) handleExistingApp(appName, configPath, targetPath string) (bool, error) {
	output := palantir.GetGlobalOutputHandler()

	// Check for changes
	hasChanges, err := gc.hasAppConfigChanges(configPath, targetPath)
	if err != nil {
		return false, fmt.Errorf("failed to check for config changes: %w", err)
	}

	if !hasChanges {
		output.PrintSuccess("Configuration is up-to-date!")
		output.PrintInfo("Local %s configs match the remote repository.", appName)
		output.PrintInfo("No changes to push.")
		return false, nil
	}

	return true, nil
}

// hasAppConfigChanges checks if the local app config differs from the remote
func (gc *GitHubClient) hasAppConfigChanges(localConfigPath, targetPath string) (bool, error) {
	// Check if the target directory exists in the repo
	repoTargetPath := filepath.Join(gc.LocalPath, targetPath)

	// If target doesn't exist in repo, this is a new app
	if _, err := os.Stat(repoTargetPath); os.IsNotExist(err) {
		// Verify the local path actually exists and has content
		if localInfo, err := os.Stat(localConfigPath); err == nil {
			if localInfo.IsDir() {
				// Check if directory has files
				entries, err := os.ReadDir(localConfigPath)
				if err == nil && len(entries) > 0 {
					return true, nil // New app with content
				}
			} else if localInfo.Size() > 0 {
				return true, nil // New file with content
			}
		}
		return false, fmt.Errorf("local config path is empty or invalid")
	}

	// Compare the local config with the repo version
	return gc.hasFileOrDirChanges(localConfigPath, repoTargetPath)
}

// hasFileOrDirChanges compares a local file or directory with a repo version
func (gc *GitHubClient) hasFileOrDirChanges(localPath, repoPath string) (bool, error) {
	localInfo, err := os.Stat(localPath)
	if err != nil {
		return false, fmt.Errorf("failed to stat local path %s: %w", localPath, err)
	}

	repoInfo, err := os.Stat(repoPath)
	if err != nil {
		if os.IsNotExist(err) {
			return true, nil // Repo version doesn't exist, so there are changes
		}
		return false, fmt.Errorf("failed to stat repo path %s: %w", repoPath, err)
	}

	// If one is a file and the other is a directory, there are changes
	if localInfo.IsDir() != repoInfo.IsDir() {
		return true, nil
	}

	if localInfo.IsDir() {
		// Compare directories recursively
		return gc.hasDirectoryChanges(localPath, repoPath)
	} else {
		// Compare files
		return gc.hasFileChanges(localPath, repoPath)
	}
}

// hasDirectoryChanges recursively compares two directories
func (gc *GitHubClient) hasDirectoryChanges(localDir, repoDir string) (bool, error) {
	// Get all files in both directories
	localFiles := make(map[string]os.FileInfo)
	repoFiles := make(map[string]os.FileInfo)

	// Walk local directory
	err := filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(localDir, path)
		if err != nil {
			return err
		}
		localFiles[relPath] = info
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to walk local directory: %w", err)
	}

	// Walk repo directory
	err = filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(repoDir, path)
		if err != nil {
			return err
		}
		repoFiles[relPath] = info
		return nil
	})
	if err != nil {
		return false, fmt.Errorf("failed to walk repo directory: %w", err)
	}

	// Check if file lists differ
	if len(localFiles) != len(repoFiles) {
		return true, nil
	}

	// Compare each file
	for relPath, localInfo := range localFiles {
		_, exists := repoFiles[relPath]
		if !exists {
			return true, nil
		}

		// Skip directories for content comparison
		if localInfo.IsDir() {
			continue
		}

		// Compare file contents
		localFilePath := filepath.Join(localDir, relPath)
		repoFilePath := filepath.Join(repoDir, relPath)
		hasChanges, err := gc.hasFileChanges(localFilePath, repoFilePath)
		if err != nil {
			return false, err
		}
		if hasChanges {
			return true, nil
		}
	}

	return false, nil
}

// hasFileChanges compares two files for differences
func (gc *GitHubClient) hasFileChanges(localFile, repoFile string) (bool, error) {
	localContent, err := os.ReadFile(localFile)
	if err != nil {
		return false, fmt.Errorf("failed to read local file %s: %w", localFile, err)
	}

	repoContent, err := os.ReadFile(repoFile)
	if err != nil {
		return false, fmt.Errorf("failed to read repo file %s: %w", repoFile, err)
	}

	return !bytes.Equal(localContent, repoContent), nil
}
