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

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/palantir"
)

// commitChanges adds and commits all changes in the repository
func (gc *GitHubClient) commitChanges(ctx context.Context, commitMessage string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Configure git user if provided
	if err := gc.configureGitUser(ctx); err != nil {
		return err
	}

	// Add all changes
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "add", "."); err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-add", err)
	}

	// Check if there are changes to commit
	result, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "diff", "--cached", "--exit-code")
	if err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-diff-check", err)
	}

	if result.ExitCode == 0 {
		// Exit code 0 means no differences
		return fmt.Errorf("no changes to commit")
	}

	// Exit code 1 means there are differences - proceed with commit
	palantir.GetGlobalOutputHandler().PrintInfo("Changes detected, proceeding with commit...")

	// Commit changes
	if _, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "commit", "-m", commitMessage); err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-commit", err)
	}

	palantir.GetGlobalOutputHandler().PrintSuccess(fmt.Sprintf("Committed changes: %s", commitMessage))
	return nil
}

// pushBranch pushes the current branch to origin
func (gc *GitHubClient) pushBranch(ctx context.Context, branchName string) error {
	originalDir, err := os.Getwd()
	if err != nil {
		return errors.NewFileSystemError(constants.OpPush, "getwd", err)
	}
	defer os.Chdir(originalDir)

	if err := os.Chdir(gc.LocalPath); err != nil {
		return errors.NewFileSystemError(constants.OpPush, "chdir", err)
	}

	// Push branch to origin
	result, err := system.RunCommandWithTimeout(ctx, constants.GitCommand, "push", "--set-upstream", "origin", branchName)
	if err != nil {
		return errors.NewInstallationError(constants.OpPush, "git-push",
			fmt.Errorf("failed to push branch: %s, error: %w", result.Error, err))
	}

	palantir.GetGlobalOutputHandler().PrintSuccess(fmt.Sprintf("Pushed branch '%s' to origin", branchName))
	return nil
}
