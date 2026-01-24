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

package pull

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/github"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

// setupPullCommand determines the target directory and loads configuration.
func setupPullCommand(cmd *cobra.Command, args []string) (string, *config.AnvilConfig, error) {
	targetDir := constants.ANVIL
	if len(args) > 0 {
		targetDir = args[0]
	}

	cfg, err := config.LoadConfig()
	if err != nil {
		return "", nil, errors.NewConfigurationError(constants.OpPull, "load-config", err)
	}

	if err := validateGitHubConfig(cfg); err != nil {
		return "", nil, err
	}

	output := palantir.GetGlobalOutputHandler()
	output.PrintHeader(fmt.Sprintf("Pull '%s' Configuration", targetDir))
	output.PrintInfo("Repository: %s", cfg.GitHub.ConfigRepo)
	output.PrintInfo("Branch: %s", cfg.GitHub.Branch)
	output.PrintInfo("Target directory: %s", targetDir)
	fmt.Println("")

	return targetDir, cfg, nil
}

// setupPullAuthentication sets up authentication and creates GitHub client.
func setupPullAuthentication(cfg *config.AnvilConfig) (*github.GitHubClient, context.Context, context.CancelFunc, error) {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Checking authentication...")
	token := ""
	if cfg.GitHub.TokenEnvVar != "" {
		token = os.Getenv(cfg.GitHub.TokenEnvVar)
		if token != "" {
			output.PrintSuccess(fmt.Sprintf("GitHub token found in environment variable: %s", cfg.GitHub.TokenEnvVar))
		} else {
			output.PrintWarning("No GitHub token found in %s - will attempt SSH authentication", cfg.GitHub.TokenEnvVar)
		}
	}

	githubClient := github.NewGitHubClient(github.GitHubClientOptions{
		RepoURL:    cfg.GitHub.ConfigRepo,
		Branch:     cfg.GitHub.Branch,
		LocalPath:  cfg.GitHub.LocalPath,
		Token:      token,
		SSHKeyPath: cfg.Git.SSHKeyPath,
		Username:   cfg.Git.Username,
		Email:      cfg.Git.Email,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	return githubClient, ctx, cancel, nil
}

// validatePullRepository validates repository access.
func validatePullRepository(ctx context.Context, githubClient *github.GitHubClient, cfg *config.AnvilConfig) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Stage 2: Validating repository access...")
	spinner := charm.NewCircleSpinner(constants.SpinnerValidatingRepository)
	spinner.Start()
	if err := githubClient.ValidateRepository(ctx); err != nil {
		spinner.Error("Repository validation failed")
		if branchErr := handleBranchConfigError(err, cfg, "validate", output); branchErr != nil {
			return branchErr
		}
		return fmt.Errorf("failed to validate repository: %w", err)
	}
	spinner.Success("Repository access confirmed")
	return nil
}

// ensurePullRepository clones or updates the repository.
func ensurePullRepository(ctx context.Context, githubClient *github.GitHubClient, cfg *config.AnvilConfig) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Stage 3: Cloning or updating repository...")
	spinner := charm.NewDotsSpinner(constants.SpinnerCloningRepository)
	spinner.Start()
	if err := githubClient.CloneRepository(ctx); err != nil {
		spinner.Error("Clone failed")
		if branchErr := handleBranchConfigError(err, cfg, "clone", output); branchErr != nil {
			return branchErr
		}
		return fmt.Errorf("failed to clone repository: %w", err)
	}
	spinner.Success("Repository ready")
	return nil
}

// pullLatestChanges pulls the latest changes from the repository.
func pullLatestChanges(ctx context.Context, githubClient *github.GitHubClient, cfg *config.AnvilConfig) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Stage 4: Pulling latest changes...")
	spinner := charm.NewDotsSpinner(constants.SpinnerPullingChanges)
	spinner.Start()
	if err := githubClient.PullChanges(ctx); err != nil {
		spinner.Error("Pull failed")
		if branchErr := handleBranchConfigError(err, cfg, "pull", output); branchErr != nil {
			return branchErr
		}
		return fmt.Errorf("failed to pull changes: %w", err)
	}
	spinner.Success("Repository updated")
	return nil
}

// copyPullDirectory copies the configuration directory to temp location.
func copyPullDirectory(cfg *config.AnvilConfig, targetDir string) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Stage 5: Copying configuration directory...")
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Copying %s directory", targetDir))
	spinner.Start()
	tempDir, err := copyDirectoryToTemp(cfg, targetDir)
	if err != nil {
		spinner.Error("Failed to copy configuration")
		return err
	}
	spinner.Success("Configuration directory copied to temp location")

	// Stage 6: Regenerate git config if pulling anvil settings
	if targetDir == constants.ANVIL {
		if err := regenerateGitConfigInPulledSettings(tempDir); err != nil {
			output.PrintWarning("Could not regenerate git config: %v", err)
			// Don't fail the operation, just warn
		}
	}

	displaySuccessMessage(targetDir, tempDir, cfg)
	return nil
}

// regenerateGitConfigInPulledSettings regenerates the git config section in pulled anvil settings.
// This replaces any masked/placeholder values with the local system's git configuration.
func regenerateGitConfigInPulledSettings(tempDir string) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Stage 6: Regenerating git config from system...")

	// Path to the pulled settings.yaml
	settingsPath := fmt.Sprintf("%s/%s", tempDir, constants.ANVIL_CONFIG_FILE)

	// Load the pulled settings
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		return fmt.Errorf("failed to read pulled settings: %w", err)
	}

	var pulledConfig config.AnvilConfig
	if err := yaml.Unmarshal(data, &pulledConfig); err != nil {
		return fmt.Errorf("failed to parse pulled settings: %w", err)
	}

	// Regenerate git config from system if masked
	regenerated, err := config.RegenerateGitConfigIfMasked(&pulledConfig)
	if err != nil {
		return fmt.Errorf("failed to regenerate git config: %w", err)
	}

	if regenerated {
		// Save the updated config back to the temp location
		updatedData, err := yaml.Marshal(&pulledConfig)
		if err != nil {
			return fmt.Errorf("failed to marshal updated config: %w", err)
		}

		if err := os.WriteFile(settingsPath, updatedData, constants.FilePerm); err != nil {
			return fmt.Errorf("failed to write updated config: %w", err)
		}

		output.PrintSuccess("Git config regenerated from system")
	} else {
		output.PrintInfo("Git config already populated, no regeneration needed")
	}

	return nil
}
