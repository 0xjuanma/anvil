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
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/utils"
)

// DiffSummary contains diff information using Git's native output
type DiffSummary struct {
	GitStatOutput string // Git's native --stat output
	FullDiff      string // Full diff for small changes
	TotalFiles    int    // Simple count of changed files
}

// DiffPreview generates diff preview for both anvil and app configs before pushing
func (gc *GitHubClient) DiffPreview(ctx context.Context, sourcePath, targetPath string) (*DiffSummary, error) {
	// First, ensure repository is ready
	if err := gc.ensureRepositoryReady(ctx); err != nil {
		return nil, err
	}

	// Use the same change detection logic as the actual push
	// Check if the target directory exists in the repo
	repoTargetPath := filepath.Join(gc.LocalPath, targetPath)
	if _, err := os.Stat(repoTargetPath); os.IsNotExist(err) {
		// Target doesn't exist in repo - this is a new app
		// Verify the local path actually exists and has content
		if localInfo, err := os.Stat(sourcePath); err == nil {
			if localInfo.IsDir() {
				// Check if directory has files
				entries, err := os.ReadDir(sourcePath)
				if err == nil && len(entries) > 0 {
					// New app with content - generate diff
					return gc.generateGitDiff(ctx, sourcePath, targetPath)
				}
			} else if localInfo.Size() > 0 {
				// New file with content - generate diff
				return gc.generateGitDiff(ctx, sourcePath, targetPath)
			}
		}
		// No content or invalid path
		return &DiffSummary{GitStatOutput: "", FullDiff: "", TotalFiles: 0}, nil
	} else {
		// Target exists in repo - check for changes using existing logic
		hasChanges, err := gc.hasAppConfigChanges(sourcePath, targetPath)
		if err != nil {
			return nil, fmt.Errorf("failed to check for config changes: %w", err)
		}

		if !hasChanges {
			return &DiffSummary{GitStatOutput: "", FullDiff: "", TotalFiles: 0}, nil
		}

		return gc.generateGitDiff(ctx, sourcePath, targetPath)
	}
}

// generateNewAppDiff generates a diff preview for new apps without modifying repository state
func (gc *GitHubClient) generateNewAppDiff(ctx context.Context, sourcePath, targetPath string) (*DiffSummary, error) {
	// For new apps, we'll simulate the diff by analyzing the local files
	// This avoids modifying the repository state during preview

	localInfo, err := os.Stat(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat source path: %w", err)
	}

	var fileCount int
	var statOutput string

	if localInfo.IsDir() {
		// Count files in directory
		entries, err := os.ReadDir(sourcePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read directory: %w", err)
		}

		fileCount = len(entries)
		if fileCount > 0 {
			// Generate a simulated stat output for new files
			statOutput = fmt.Sprintf(" %s/", targetPath)
			for _, entry := range entries {
				if !entry.IsDir() {
					statOutput += fmt.Sprintf("\n %s/%s | 1 +\n", targetPath, entry.Name())
				}
			}
			statOutput += fmt.Sprintf("\n %d files changed, %d insertions(+)\n", fileCount, fileCount)
		}
	} else {
		// Single file
		fileCount = 1
		fileName := filepath.Base(sourcePath)
		statOutput = fmt.Sprintf(" %s/%s | 1 +\n 1 file changed, 1 insertion(+)\n", targetPath, fileName)
	}

	return &DiffSummary{
		GitStatOutput: statOutput,
		FullDiff:      "", // No full diff for new apps in preview
		TotalFiles:    fileCount,
	}, nil
}

// generateGitDiff handles diff generation using Git's native capabilities (simplified)
func (gc *GitHubClient) generateGitDiff(ctx context.Context, sourcePath, targetPath string) (*DiffSummary, error) {
	// Change to repo directory
	originalDir, err := os.Getwd()
	if err != nil {
		return nil, errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return nil, errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Check if this is a new app (target doesn't exist in repo)
	repoTargetPath := filepath.Join(gc.LocalPath, targetPath)
	isNewApp := false
	if _, err := os.Stat(repoTargetPath); os.IsNotExist(err) {
		isNewApp = true
	}

	if isNewApp {
		// For new apps, generate diff without copying files to avoid modifying repo state
		return gc.generateNewAppDiff(ctx, sourcePath, targetPath)
	}

	// For existing apps, use the original logic
	// Setup files based on target path type
	if strings.HasSuffix(targetPath, ".yaml") || strings.HasSuffix(targetPath, ".yml") {
		// Single file (anvil settings)
		repoFilePath := filepath.Join(gc.LocalPath, targetPath)
		if err := utils.EnsureDirectory(filepath.Dir(repoFilePath)); err != nil {
			return nil, errors.NewFileSystemError(constants.OpPush, "mkdir", err)
		}
		if err := utils.CopyFileSimple(sourcePath, repoFilePath); err != nil {
			return nil, errors.NewFileSystemError(constants.OpPush, "copy-file", err)
		}
	} else {
		// Directory (app configs)
		targetDir := filepath.Join(gc.LocalPath, targetPath)
		if err := utils.EnsureDirectory(targetDir); err != nil {
			return nil, errors.NewFileSystemError(constants.OpPush, "mkdir", err)
		}
		if err := gc.copyConfigToRepo(sourcePath, targetDir); err != nil {
			return nil, err
		}
	}

	// Stage the target files
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "add", targetPath); err != nil {
		return nil, errors.NewInstallationError(constants.OpPush, "git-add", err)
	}

	// Get Git's native stat output
	statResult, err := system.RunCommandWithTimeout(ctx, constants.GitCommand,
		"diff", "--cached", "--stat", "--stat-width=80")
	if err != nil {
		return nil, errors.NewInstallationError(constants.OpPush, "git-diff-stat", err)
	}

	// Get full diff only for small single files
	var fullDiff string
	if gc.isSingleSmallFile(statResult.Output) {
		diffResult, err := system.RunCommandWithTimeout(ctx, constants.GitCommand,
			"diff", "--cached", "--no-color")
		if err == nil {
			fullDiff = diffResult.Output
		}
	}

	// Always reset staging area
	system.RunCommandWithTimeout(ctx, constants.GitCommand, "reset", "HEAD")

	// Revert any changes we made to the working directory during diff generation
	// This safely reverts only the changes we made, without removing existing files
	system.RunCommandWithTimeout(ctx, constants.GitCommand, "checkout", "--", targetPath)

	// Clean up any untracked files that were created during diff generation
	system.RunCommandWithTimeout(ctx, constants.GitCommand, "clean", "-fd", targetPath)

	return &DiffSummary{
		GitStatOutput: statResult.Output,
		FullDiff:      fullDiff,
		TotalFiles:    gc.extractFileCount(statResult.Output),
	}, nil
}
