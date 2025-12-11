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
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
)

// validateGitHubConfig validates that GitHub configuration is properly set up.
func validateGitHubConfig(cfg *config.AnvilConfig) error {
	if cfg.GitHub.ConfigRepo == "" {
		return errors.NewConfigurationError(constants.OpPull, "validate-config",
			fmt.Errorf("github.config_repo is not configured. Please edit %s/%s and set github.config_repo to your repository (e.g., 'username/dotfiles')",
				config.AnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE))
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
				constants.ANVIL_CONFIG_FILE, config.AnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE))
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
	tempBasedir := filepath.Join(config.AnvilConfigDirectory(), "temp")
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
