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

// Package pull provides functionality to pull configuration files from
// a GitHub repository to a temporary location for review.
package pull

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/github"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

var PullCmd = &cobra.Command{
	Use:   "pull [directory]",
	Short: "Pull configuration files from a specific directory in GitHub repository",
	Long:  constants.PULL_COMMAND_LONG_DESCRIPTION,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runPullCommand(cmd, args)
	},
}

const branchConfigErrorPrefix = "Branch Configuration Error"

// handleBranchConfigError handles branch configuration errors with helpful context.
func handleBranchConfigError(err error, cfg *config.AnvilConfig, stage string, output palantir.OutputHandler) error {
	if !strings.Contains(err.Error(), branchConfigErrorPrefix) {
		return err
	}

	fmt.Println("")
	output.PrintError("%s", err.Error())
	fmt.Println("")

	if stage == "validate" {
		output.PrintInfo("The repository exists but the configured branch is not available.")
		output.PrintInfo("    You may need to:")
		output.PrintInfo("    • Update the branch in your %s", constants.ANVIL_CONFIG_FILE)
		output.PrintInfo("    • Or check the available branches in your repository")
	} else {
		output.PrintInfo("The repository exists but the configured branch is not available.")
		output.PrintInfo("    You may need to:")
		output.PrintInfo("    • Update the branch in your %s", constants.ANVIL_CONFIG_FILE)
		output.PrintInfo("    • Or delete the local repository at: %s", cfg.GitHub.LocalPath)
		output.PrintInfo("      (It will be re-cloned with the correct branch)")
	}

	return fmt.Errorf("%s failed due to branch configuration issue", stage)
}

// runPullCommand executes the configuration pull process for a specific directory.
func runPullCommand(cmd *cobra.Command, args []string) error {
	// Setup: Determine target and load config
	targetDir, cfg, err := setupPullCommand(cmd, args)
	if err != nil {
		return err
	}

	// Stage 1: Authentication
	githubClient, ctx, err := setupPullAuthentication(cfg)
	if err != nil {
		return err
	}

	// Stage 2: Validate repository
	if err := validatePullRepository(ctx, githubClient, cfg); err != nil {
		return err
	}

	// Stage 3: Clone/update repository
	if err := ensurePullRepository(ctx, githubClient, cfg); err != nil {
		return err
	}

	// Stage 4: Pull latest changes
	if err := pullLatestChanges(ctx, githubClient, cfg); err != nil {
		return err
	}

	// Stage 5: Copy directory
	return copyPullDirectory(cfg, targetDir)
}

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
func setupPullAuthentication(cfg *config.AnvilConfig) (*github.GitHubClient, context.Context, error) {
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

	githubClient := github.NewGitHubClient(
		cfg.GitHub.ConfigRepo,
		cfg.GitHub.Branch,
		cfg.GitHub.LocalPath,
		token,
		cfg.Git.SSHKeyPath,
		cfg.Git.Username,
		cfg.Git.Email,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	// Store cancel in a way that can be deferred - we'll need to handle this differently
	// For now, we'll let the context timeout handle cleanup
	_ = cancel

	return githubClient, ctx, nil
}

// validatePullRepository validates repository access.
func validatePullRepository(ctx context.Context, githubClient *github.GitHubClient, cfg *config.AnvilConfig) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage("Stage 2: Validating repository access...")
	spinner := charm.NewCircleSpinner("Validating repository access and branch configuration")
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
	spinner := charm.NewDotsSpinner("Cloning or updating repository")
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
	spinner := charm.NewDotsSpinner("Pulling latest changes")
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

	displaySuccessMessage(targetDir, tempDir, cfg)
	return nil
}

// displaySuccessMessage displays a success message after the pull operation.
func displaySuccessMessage(targetDir, tempDir string, cfg *config.AnvilConfig) {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Pull Complete!")
	o.PrintInfo("Configuration directory '%s' has been pulled from: %s", targetDir, cfg.GitHub.ConfigRepo)
	o.PrintInfo("Files are available at: %s", tempDir)

	// List the files that were copied
	if err := listCopiedFiles(tempDir); err == nil {
		// Files listed successfully
	} else {
		o.PrintWarning("Could not list copied files: %v", err)
	}
}

// validateGitHubConfig validates that GitHub configuration is properly set up.
func validateGitHubConfig(cfg *config.AnvilConfig) error {
	if cfg.GitHub.ConfigRepo == "" {
		return errors.NewConfigurationError(constants.OpPull, "validate-config",
			fmt.Errorf("github.config_repo is not configured. Please edit %s/%s and set github.config_repo to your repository (e.g., 'username/dotfiles')",
				config.GetAnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE))
	}

	if cfg.GitHub.Branch == "" {
		return errors.NewConfigurationError(constants.OpPull, "validate-config",
			fmt.Errorf(`github.branch is not configured.

📝 To fix this:
  1. Edit your %s file at: %s/%s
  2. Set the 'github.branch' field to your repository's default branch
  3. Common branch names: 'main', 'master', 'develop'
  
Example:
  github:
    branch: "main"  # ← Set this to your repository's default branch`,
				constants.ANVIL_CONFIG_FILE, config.GetAnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE))
	}

	if cfg.GitHub.LocalPath == "" {
		return errors.NewConfigurationError(constants.OpPull, "validate-config",
			fmt.Errorf("github.local_path is not configured"))
	}

	output := palantir.GetGlobalOutputHandler()
	// Provide guidance about branch configuration
	if cfg.GitHub.Branch != "main" && cfg.GitHub.Branch != "master" {
		output.PrintWarning("Note: You're using branch '%s'. Make sure this branch exists in your repository.", cfg.GitHub.Branch)
		output.PrintInfo("💡 Common default branches are 'main' or 'master'")
	}

	// Check if git is available
	if cfg.Git.Username == "" || cfg.Git.Email == "" {
		output.PrintWarning(fmt.Sprintf("Git user configuration is incomplete. Consider setting git.username and git.email in %s", constants.ANVIL_CONFIG_FILE))
	}

	return nil
}

// copyDirectoryToTemp copies a specific directory from the repo to a temporary location.
func copyDirectoryToTemp(cfg *config.AnvilConfig, targetDir string) (string, error) {
	// Source directory in the cloned repo
	sourceDir := filepath.Join(cfg.GitHub.LocalPath, targetDir)

	// Check if source directory exists
	if _, err := os.Stat(sourceDir); os.IsNotExist(err) {
		return "", errors.NewConfigurationError(constants.OpPull, "source-directory",
			fmt.Errorf("directory '%s' does not exist in repository %s", targetDir, cfg.GitHub.ConfigRepo))
	}

	// Create temp directory inside anvil config
	tempBasedir := filepath.Join(config.GetAnvilConfigDirectory(), "temp")
	if err := utils.EnsureDirectory(tempBasedir); err != nil {
		return "", errors.NewFileSystemError(constants.OpPull, "create-temp-dir", err)
	}

	// Destination directory
	destDir := filepath.Join(tempBasedir, targetDir)

	// Remove existing destination if it exists
	if err := os.RemoveAll(destDir); err != nil {
		return "", errors.NewFileSystemError(constants.OpPull, "remove-existing", err)
	}

	// Copy directory recursively
	if err := utils.CopyDirectorySimple(sourceDir, destDir); err != nil {
		return "", errors.NewFileSystemError(constants.OpPull, "copy-directory", err)
	}

	return destDir, nil
}

// listCopiedFiles lists the files that were copied to the temp directory.
func listCopiedFiles(tempDir string) error {
	fmt.Println("")
	palantir.GetGlobalOutputHandler().PrintInfo("Copied files:")

	return filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, only show files
		if !info.IsDir() {
			relPath, err := filepath.Rel(tempDir, path)
			if err != nil {
				relPath = path
			}
			palantir.GetGlobalOutputHandler().PrintInfo("  • %s", relPath)
		}

		return nil
	})
}

func init() {
	// Add flags for additional functionality
	PullCmd.Flags().Bool("force", false, "Force pull even if local changes exist")
	PullCmd.Flags().String("branch", "", "Override the branch to pull from")
}
