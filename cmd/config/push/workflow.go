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

package push

import (
	"context"
	"fmt"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/github"
	"github.com/0xjuanma/palantir"
)

// executeCommonPushStages executes the common stages shared between pushAppConfig and pushAnvilConfig.
// Returns githubClient, diffSummary, and error. If user cancels, returns nil diffSummary.
func executeCommonPushStages(anvilConfig *config.AnvilConfig, appName, configPath string, ctx context.Context) (*github.GitHubClient, *github.DiffSummary, error) {
	output := palantir.GetGlobalOutputHandler()

	// Security warning
	showSecurityWarning(anvilConfig.GitHub.ConfigRepo)

	// Authentication setup
	githubClient, err := setupAuthentication(anvilConfig)
	if err != nil {
		return nil, nil, err
	}

	// Prepare and show diff
	diffSummary, err := prepareDiffPreview(githubClient, appName, configPath, ctx)
	if err != nil {
		return nil, nil, err
	}

	// User confirmation
	if !handleUserConfirmation(output, appName, githubClient, ctx) {
		return nil, nil, nil
	}

	return githubClient, diffSummary, nil
}

// prepareDiffPreview prepares and shows the diff preview.
func prepareDiffPreview(githubClient *github.GitHubClient, appName, configPath string, ctx context.Context) (*github.DiffSummary, error) {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage(fmt.Sprintf("Preparing to push %s configuration...", appName))
	output.PrintInfo("Repository: %s", githubClient.RepoURL)
	output.PrintInfo("Branch: %s", githubClient.Branch)
	output.PrintInfo("App: %s", appName)
	output.PrintInfo("Local config path: %s", configPath)

	// Add diff output before confirmation
	output.PrintStage(constants.SpinnerAnalyzingChanges)
	targetPath := fmt.Sprintf("%s/", appName)
	diffSummary, err := githubClient.GetDiffPreview(ctx, configPath, targetPath)
	if err != nil {
		output.PrintWarning("Unable to generate diff preview: %v", err)
		return nil, nil
	}

	showDiffOutput(diffSummary)
	return diffSummary, nil
}

// handleUserConfirmation handles user confirmation for the push operation.
func handleUserConfirmation(output palantir.OutputHandler, appName string, githubClient *github.GitHubClient, ctx context.Context) bool {
	output.PrintStage("Requesting user confirmation...")
	if !output.Confirm(fmt.Sprintf("Do you want to push your %s configurations to the repository?", appName)) {
		output.PrintInfo("Push cancelled by user")
		cleanupOnError(ctx, githubClient, nil)
		return false
	}
	return true
}

// cleanupOnError handles cleanup of staged changes when an error occurs.
func cleanupOnError(ctx context.Context, client *github.GitHubClient, err error) error {
	if err != nil {
		if cleanupErr := client.CleanupStagedChanges(ctx); cleanupErr != nil {
			palantir.GetGlobalOutputHandler().PrintWarning("Failed to cleanup staged changes: %v", cleanupErr)
		}
	}
	return err
}

// performPushOperation executes the actual push operation.
func performPushOperation(ctx context.Context, opts PushOperationOptions) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage(fmt.Sprintf("Pushing %s configuration to repository...", opts.AppName))

	result, err := opts.GitHubClient.PushAppConfig(ctx, opts.AppName, opts.ConfigPath)
	if err != nil {
		return cleanupOnError(ctx, opts.GitHubClient, errors.NewInstallationError(constants.OpPush, "push-app-config", err))
	}

	// Check if no changes were detected (result will be nil)
	if result == nil {
		// Configuration was up-to-date, success message already shown in PushAppConfig
		return nil
	}

	displaySuccessMessage(opts.AppName, result, opts.DiffSummary, opts.AnvilConfig)
	return nil
}

// executeCommonPushStagesForAnvil executes common stages for anvil config push.
func executeCommonPushStagesForAnvil(anvilConfig *config.AnvilConfig, settingsPath, remotePath string, ctx context.Context) (*github.GitHubClient, *github.DiffSummary, error) {
	output := palantir.GetGlobalOutputHandler()

	// Security warning
	showSecurityWarning(anvilConfig.GitHub.ConfigRepo)

	// Authentication setup
	githubClient, err := setupAuthentication(anvilConfig)
	if err != nil {
		return nil, nil, err
	}

	// Prepare and show diff
	output.PrintStage(constants.SpinnerAnalyzingChanges)
	diffSummary, err := githubClient.GetDiffPreview(ctx, settingsPath, remotePath)
	if err != nil {
		output.PrintWarning("Unable to generate diff preview: %v", err)
	} else {
		showDiffOutput(diffSummary)
	}

	// User confirmation
	output.PrintStage("Requesting user confirmation...")
	if !output.Confirm("Do you want to push your anvil settings to the repository?") {
		output.PrintInfo("Push cancelled by user")
		cleanupOnError(ctx, githubClient, nil)
		return nil, nil, nil
	}

	return githubClient, diffSummary, nil
}
