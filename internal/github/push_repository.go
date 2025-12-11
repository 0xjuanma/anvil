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
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/palantir"
)

// ensureRepositoryReady ensures the repository is cloned and up to date
func (gc *GitHubClient) ensureRepositoryReady(ctx context.Context) error {
	// Clone repository if it doesn't exist
	if err := gc.CloneRepository(ctx); err != nil {
		return fmt.Errorf("failed to clone repository: %w", err)
	}

	// Switch back to main branch and pull latest changes
	if err := gc.switchToMainBranch(ctx); err != nil {
		return fmt.Errorf("failed to switch to main branch: %w", err)
	}

	if err := gc.PullChanges(ctx); err != nil {
		return fmt.Errorf("failed to pull latest changes: %w", err)
	}

	// Ensure repository is in a clean state before starting push operations
	if err := gc.ensureCleanState(ctx); err != nil {
		return fmt.Errorf("failed to ensure clean repository state: %w", err)
	}

	return nil
}

// switchToMainBranch switches to the main branch specified in config
func (gc *GitHubClient) switchToMainBranch(ctx context.Context) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Checkout main branch
	_, err = system.RunCommandWithTimeout(ctx, constants.GitCommand, "checkout", gc.Branch)
	if err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-checkout-main", err)
	}

	return nil
}

// createAndCheckoutBranch creates a new branch and checks it out
func (gc *GitHubClient) createAndCheckoutBranch(ctx context.Context, branchName string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Create and checkout new branch
	_, err = system.RunCommandWithTimeout(ctx, constants.GitCommand, "checkout", "-b", branchName)
	if err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-checkout-new-branch", err)
	}

	palantir.GetGlobalOutputHandler().PrintInfo("Created and switched to branch: %s", branchName)
	return nil
}

// ensureCleanState ensures the repository is in a clean state before push operations
func (gc *GitHubClient) ensureCleanState(ctx context.Context) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Check if there are any staged changes
	stagedResult, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "diff", "--cached", "--exit-code")
	if err != nil && stagedResult.ExitCode != 0 {
		// There are staged changes, reset them
		if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "reset", "HEAD"); err != nil {
			return errors.NewInstallationError(constants.OpPush, "git-reset", err)
		}
	}

	// Check if there are any untracked files
	statusResult, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "status", "--porcelain")
	if err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-status", err)
	}

	// If there are untracked files, clean them
	if strings.TrimSpace(statusResult.Output) != "" {
		if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "clean", "-fd"); err != nil {
			return errors.NewInstallationError(constants.OpPush, "git-clean", err)
		}
	}

	return nil
}

// CleanupStagedChanges removes any staged changes from the repository
// This is called when a push operation is cancelled to ensure clean state
func (gc *GitHubClient) CleanupStagedChanges(ctx context.Context) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Reset any staged changes
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "reset", "HEAD"); err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-reset", err)
	}

	// Clean any untracked files that might have been created during diff preview
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "clean", "-fd"); err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-clean", err)
	}

	// Also reset any working directory changes that might have been left behind
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "checkout", "--", "."); err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-checkout", err)
	}

	// Switch back to main branch to ensure we're in a clean state
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "checkout", gc.Branch); err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-checkout-main", err)
	}

	return nil
}
