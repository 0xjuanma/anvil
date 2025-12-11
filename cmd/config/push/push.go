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

// Package push provides functionality to push configuration files to
// a GitHub repository with automated branch creation and diff preview.
package push

import (
	"context"
	"fmt"
	"os"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/github"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

// PushOperationOptions contains options for performing a push operation.
type PushOperationOptions struct {
	GitHubClient *github.GitHubClient
	AppName      string
	ConfigPath   string
	DiffSummary  *github.DiffSummary
	AnvilConfig  *config.AnvilConfig
}

var PushCmd = &cobra.Command{
	Use:   "push [app-name]",
	Short: "Push configuration files to GitHub repository",
	Long:  constants.PUSH_COMMAND_LONG_DESCRIPTION,
	Args:  cobra.MaximumNArgs(1), // Accept 0 or 1 argument
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPushCommand(cmd, args)
	},
}

// isNewAppAddition checks if this is a new app that exists locally but not in remote.
func isNewAppAddition(appName string, anvilConfig *config.AnvilConfig) bool {
	// Check if app exists in local configs but not in remote
	if localPath, exists := anvilConfig.Configs[appName]; exists {
		if _, err := os.Stat(localPath); err == nil {
			// App exists locally and is configured
			return true
		}
	}
	return false
}

// runPushCommand executes the configuration push process.
func runPushCommand(cmd *cobra.Command, args []string) error {
	// Option 2: App-specific config push
	if len(args) > 0 {
		appName := args[0]
		return pushAppConfig(appName)
	}

	// Option 1: Anvil config push
	return pushAnvilConfig()
}

// pushAppConfig pushes application-specific configuration to the repository.
func pushAppConfig(appName string) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintHeader(fmt.Sprintf("Push '%s' Configuration", appName))

	// Stage 1: Load and validate configuration
	anvilConfig, err := loadAndValidateConfig()
	if err != nil {
		return err
	}

	// Stage 2: Resolve app location
	configPath, err := resolveAppLocation(appName, anvilConfig)
	if err != nil {
		return err
	}

	// Show new app information if this is a new addition
	if isNewAppAddition(appName, anvilConfig) {
		showNewAppInfo(appName, configPath)
	}

	// Common push workflow stages
	ctx := context.Background()
	githubClient, diffSummary, err := executeCommonPushStages(anvilConfig, appName, configPath, ctx)
	if err != nil {
		return err
	}
	if diffSummary == nil {
		return nil // User cancelled
	}

	// Stage 7: Push configuration
	opts := PushOperationOptions{
		GitHubClient: githubClient,
		AppName:      appName,
		ConfigPath:   configPath,
		DiffSummary:  diffSummary,
		AnvilConfig:  anvilConfig,
	}
	return performPushOperation(ctx, opts)
}

// loadAndValidateConfig loads and validates the anvil configuration.
func loadAndValidateConfig() (*config.AnvilConfig, error) {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage(constants.SpinnerLoadingConfig)

	anvilConfig, err := config.LoadConfig()
	if err != nil {
		return nil, errors.NewConfigurationError(constants.OpPush, "load-config", err)
	}

	// Validate GitHub configuration
	if anvilConfig.GitHub.ConfigRepo == "" {
		return nil, errors.NewConfigurationError(constants.OpPush, "missing-repo",
			fmt.Errorf(constants.ErrGitHubRepoNotSet, constants.ANVIL_CONFIG_FILE))
	}

	output.PrintSuccess("Configuration loaded successfully")
	return anvilConfig, nil
}

// resolveAppLocation resolves the app configuration location.
func resolveAppLocation(appName string, anvilConfig *config.AnvilConfig) (string, error) {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Resolving app configuration location...")

	configPath, locationSource, err := config.ResolveAppLocation(appName)
	if err != nil {
		// Check if this is a new app addition
		if isNewAppAddition(appName, anvilConfig) {
			output.PrintInfo("🆕 New app '%s' detected - will be added to repository", appName)
			// Get the configured path for new apps
			if localPath, exists := anvilConfig.Configs[appName]; exists {
				configPath = localPath
			} else {
				return "", handleAppLocationError(appName, err)
			}
		} else {
			return "", handleAppLocationError(appName, err)
		}
	}

	// Handle different location sources
	if locationSource == config.LocationTemp {
		output.PrintWarning("App '%s' found in temp directory but not configured in settings\n", appName)
		output.PrintInfo("💡 To push app configurations, you need to configure the local path in %s:\n", constants.ANVIL_CONFIG_FILE)
		output.PrintInfo("configs:")
		output.PrintInfo("  %s: /path/to/your/%s/configs\n", appName, appName)
		output.PrintInfo("This ensures anvil knows where to find your local configurations.")
		output.PrintInfo("The temp directory (%s) contains pulled configs for review only.", configPath)
		return "", fmt.Errorf("app config path not configured in settings")
	}

	output.PrintSuccess("App configuration location resolved")
	output.PrintInfo("Config path: %s", configPath)
	return configPath, nil
}

// setupAuthentication sets up GitHub authentication.
func setupAuthentication(anvilConfig *config.AnvilConfig) (*github.GitHubClient, error) {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage(constants.SpinnerSettingUpAuth)

	var token string
	if anvilConfig.GitHub.TokenEnvVar != "" {
		token = os.Getenv(anvilConfig.GitHub.TokenEnvVar)
		if token == "" {
			output.PrintWarning("GitHub token not found in environment variable: %s", anvilConfig.GitHub.TokenEnvVar)
			output.PrintInfo("Proceeding with SSH authentication if available...")
		} else {
			output.PrintSuccess("GitHub token found in environment")
		}
	}

	// Create GitHub client
	githubClient := github.NewGitHubClient(
		anvilConfig.GitHub.ConfigRepo,
		anvilConfig.GitHub.Branch,
		anvilConfig.GitHub.LocalPath,
		token,
		anvilConfig.Git.SSHKeyPath,
		anvilConfig.Git.Username,
		anvilConfig.Git.Email,
	)

	return githubClient, nil
}

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

// pushAnvilConfig pushes the anvil settings.yaml to the repository.
func pushAnvilConfig() error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintHeader("Push Anvil Configuration")

	// Stage 1: Load and validate configuration
	anvilConfig, err := loadAndValidateConfig()
	if err != nil {
		return err
	}

	// Get settings file path
	settingsPath := config.GetAnvilConfigPath()

	output.PrintStage("Preparing to push anvil configuration...")
	output.PrintInfo("Repository: %s", anvilConfig.GitHub.ConfigRepo)
	output.PrintInfo("Branch: %s", anvilConfig.GitHub.Branch)
	output.PrintInfo("Settings file: %s", settingsPath)

	// Common push workflow stages
	ctx := context.Background()
	anvilSettingsPath := fmt.Sprintf("%s/%s", constants.ANVIL_CONFIG_DIR, constants.ANVIL_CONFIG_FILE)
	githubClient, diffSummary, err := executeCommonPushStagesForAnvil(anvilConfig, settingsPath, anvilSettingsPath[1:], ctx)
	if err != nil {
		return err
	}
	if githubClient == nil || diffSummary == nil {
		return nil // User cancelled
	}

	// Stage 4: Push configuration
	output.PrintStage("Pushing configuration to repository...")
	result, err := githubClient.PushAnvilConfig(ctx, settingsPath)
	if err != nil {
		output.PrintError("Push failed: %v", err)
		return cleanupOnError(ctx, githubClient, errors.NewInstallationError(constants.OpPush, "push-config", err))
	}

	// Check if no changes were detected (result will be nil)
	if result == nil {
		output.PrintSuccess(constants.StatusUpToDate)
		return nil
	}

	output.PrintSuccess(constants.StatusPushedSuccessfully)
	displaySuccessMessage(constants.ANVIL, result, diffSummary, anvilConfig)

	return nil
}
