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
	settingsPath := config.AnvilConfigPath()

	output.PrintStage("Preparing to push anvil configuration...")
	output.PrintInfo("Repository: %s", anvilConfig.GitHub.ConfigRepo)
	output.PrintInfo("Branch: %s", anvilConfig.GitHub.Branch)
	output.PrintInfo("Settings file: %s", settingsPath)

	// Stage 2: Sanitize settings to remove PII (git config section)
	output.PrintStage("Sanitizing git config (masking username, email, SSH key path)...")
	sanitizedPath, cleanup, err := config.CreateSanitizedTempFile(anvilConfig)
	if err != nil {
		output.PrintError("Failed to sanitize settings: %v", err)
		return errors.NewConfigurationError(constants.OpPush, "sanitize-config", err)
	}
	// Ensure cleanup happens regardless of success or failure
	defer cleanup()

	output.PrintSuccess("Git config masked for security. Local values preserved.")

	// Common push workflow stages (use sanitized path for diff preview)
	ctx := context.Background()
	anvilSettingsPath := fmt.Sprintf("%s/%s", constants.ANVIL_CONFIG_DIR, constants.ANVIL_CONFIG_FILE)
	githubClient, diffSummary, err := executeCommonPushStagesForAnvil(anvilConfig, sanitizedPath, anvilSettingsPath[1:], ctx)
	if err != nil {
		return err
	}
	if githubClient == nil || diffSummary == nil {
		return nil // User cancelled
	}

	// Stage 4: Push configuration (use sanitized path)
	output.PrintStage("Pushing configuration to repository...")
	result, err := githubClient.PushAnvilConfig(ctx, sanitizedPath)
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
