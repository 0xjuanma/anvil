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
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
)

// PushConfigResult represents the result of a config push operation
type PushConfigResult struct {
	BranchName     string
	CommitMessage  string
	RepositoryURL  string
	FilesCommitted []string
}

// verifyRepositoryPrivacy ensures the repository is private before allowing push operations
func (gc *GitHubClient) verifyRepositoryPrivacy(ctx context.Context) error {
	// First test git access using the client's authentication method
	authenticatedURL := gc.getCloneURL()
	result, err := system.RunCommandWithTimeout(ctx, "git", "ls-remote", authenticatedURL, "HEAD")

	if err != nil || !result.Success {
		return fmt.Errorf("🚨 SECURITY BLOCK: Cannot verify repository privacy - authentication failed\n"+
			"Repository: %s\n"+
			"Anvil REQUIRES private repositories for configuration data.\n"+
			"Configure proper authentication (GITHUB_TOKEN or SSH keys) before pushing", gc.RepoURL)
	}

	// Test if repository is publicly accessible (this should FAIL for private repos)
	repoURL := fmt.Sprintf("https://github.com/%s", gc.RepoURL)
	httpResult, httpErr := system.RunCommandWithTimeout(ctx, "curl", "-s", "-f", "-I", repoURL)

	if httpErr == nil && httpResult.Success {
		// 🚨 CRITICAL: Repository is public - BLOCK the push
		output := palantir.GetGlobalOutputHandler()
		output.PrintError("🚨 SECURITY VIOLATION: Configuration push BLOCKED")
		output.PrintError("")
		output.PrintError("Repository '%s' is PUBLIC", gc.RepoURL)
		output.PrintError("❌ Configuration files contain sensitive data")
		output.PrintError("❌ PUBLIC repositories expose API keys, paths, and personal information")
		output.PrintError("❌ This could lead to security breaches and data leaks")
		output.PrintError("")
		output.PrintError("🔒 REQUIRED ACTION: Make repository PRIVATE")
		output.PrintError("   Visit: https://github.com/%s/settings", gc.RepoURL)
		output.PrintError("   Go to: Danger Zone → Change repository visibility → Private")
		output.PrintError("")
		output.PrintError("🛡️  Anvil will NEVER push configuration data to public repositories")

		return fmt.Errorf("SECURITY BLOCK: Repository is public. Configuration push denied for security")
	}

	// Repository appears to be private and git access works - safe to proceed
	palantir.GetGlobalOutputHandler().PrintSuccess("Repository privacy verified - safe to push configuration data")
	return nil
}

// PushConfig pushes configuration files to the repository (unified function for both anvil and app configs)
func (gc *GitHubClient) PushConfig(ctx context.Context, appName, configPath string) (*PushConfigResult, error) {
	// 🚨 CRITICAL SECURITY CHECK: Verify repository is private before ANY push operations
	if err := gc.verifyRepositoryPrivacy(ctx); err != nil {
		return nil, err
	}

	// Ensure repository is ready
	if err := gc.ensureRepositoryReady(ctx); err != nil {
		return nil, err
	}

	// Check if there are differences before proceeding
	targetPath := fmt.Sprintf("%s/", appName) // App configs go in a directory named after the app
	output := palantir.GetGlobalOutputHandler()

	// Check for changes and handle new vs existing apps
	shouldProceed, err := gc.checkForChanges(ctx, appName, configPath, targetPath)
	if err != nil {
		return nil, err
	}
	if !shouldProceed {
		return nil, nil // No changes to push
	}

	output.PrintInfo("Differences detected between local and remote %s configuration", appName)

	// Perform the push operation
	return gc.performPushOperation(ctx, appName, configPath)
}


// performPushOperation executes the actual push operation
func (gc *GitHubClient) performPushOperation(ctx context.Context, appName, configPath string) (*PushConfigResult, error) {
	// Generate branch name with timestamp
	branchName := generateTimestampedBranchName("config-push")

	// Create and checkout new branch
	if err := gc.createAndCheckoutBranch(ctx, branchName); err != nil {
		return nil, err
	}

	// Copy configs to repo
	targetDir := filepath.Join(gc.LocalPath, appName)
	if err := utils.EnsureDirectory(targetDir); err != nil {
		return nil, errors.NewFileSystemError(constants.OpPush, "mkdir-app", err)
	}

	// Copy the config path (file or directory) to the target directory
	if err := gc.copyConfigToRepo(configPath, targetDir); err != nil {
		return nil, err
	}

	// Commit changes
	commitMessage := fmt.Sprintf("anvil[push]: %s", appName)
	if err := gc.commitChanges(ctx, commitMessage); err != nil {
		return nil, err
	}

	// Push branch
	if err := gc.pushBranch(ctx, branchName); err != nil {
		return nil, err
	}

	// Determine files committed
	filesCommitted, err := gc.getCommittedFiles(targetDir, appName)
	if err != nil {
		filesCommitted = []string{fmt.Sprintf("%s/", appName)} // Fallback
	}

	result := &PushConfigResult{
		BranchName:     branchName,
		CommitMessage:  commitMessage,
		RepositoryURL:  gc.getRepositoryURL(),
		FilesCommitted: filesCommitted,
	}

	return result, nil
}

// PushAppConfig is a wrapper for backwards compatibility - delegates to unified PushConfig
func (gc *GitHubClient) PushAppConfig(ctx context.Context, appName, configPath string) (*PushConfigResult, error) {
	return gc.PushConfig(ctx, appName, configPath)
}

// PushAnvilConfig is a wrapper for backwards compatibility - delegates to unified PushConfig
func (gc *GitHubClient) PushAnvilConfig(ctx context.Context, settingsPath string) (*PushConfigResult, error) {
	return gc.PushConfig(ctx, constants.ANVIL, settingsPath)
}

