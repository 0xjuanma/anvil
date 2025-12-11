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
	"fmt"
	"os"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/github"
	"github.com/0xjuanma/palantir"
)

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
	githubClient := github.NewGitHubClient(github.GitHubClientOptions{
		RepoURL:    anvilConfig.GitHub.ConfigRepo,
		Branch:     anvilConfig.GitHub.Branch,
		LocalPath:  anvilConfig.GitHub.LocalPath,
		Token:      token,
		SSHKeyPath: anvilConfig.Git.SSHKeyPath,
		Username:   anvilConfig.Git.Username,
		Email:      anvilConfig.Git.Email,
	})

	return githubClient, nil
}
