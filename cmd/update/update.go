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

// Package update provides functionality for updating Anvil to the latest
// version by downloading and executing the installation script from GitHub.
package update

import (
	"context"
	"fmt"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

// UpdateCmd represents the update command.
var UpdateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Anvil to the latest version",
	Long:  constants.UPDATE_COMMAND_LONG_DESCRIPTION,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runUpdateCommand(cmd)
	},
}

// runUpdateCommand executes the update process.
func runUpdateCommand(cmd *cobra.Command) error {
	o := palantir.GetGlobalOutputHandler()

	o.PrintHeader("Updating Anvil to Latest Version")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	result, err := updateAnvil(cmd.Context(), dryRun)

	if err != nil {
		return errors.NewInstallationError(constants.OpUpdate, constants.ANVIL,
			fmt.Errorf("failed to execute update script: %w", err))
	}

	// For dry-run mode, result will be nil. Return early
	if dryRun {
		return nil
	}

	if !result.Success {
		return errors.NewInstallationError(constants.OpUpdate, constants.ANVIL,
			fmt.Errorf("update script failed with exit code %d: %s", result.ExitCode, result.Output))
	}

	o.PrintSuccess("Anvil has been successfully updated!")
	o.PrintInfo("Run 'anvil --version' to verify the new version")
	o.PrintInfo("You may need to restart your terminal session for changes to take effect")

	return nil
}

// updateAnvil updates Anvil to the latest version by downloading and
// executing the installation script from GitHub releases with fallback to main branch.
func updateAnvil(ctx context.Context, dryRun bool) (*system.CommandResult, error) {
	o := palantir.GetGlobalOutputHandler()

	if dryRun {
		o.PrintInfo("Dry run mode - would update Anvil to the latest version")
		o.PrintInfo("Command that would be executed:")
		o.PrintInfo("curl -sSL %s | bash", constants.UpdateReleaseScriptURL)
		return nil, nil
	}

	// Check if curl is available
	if !system.CommandExists("curl") {
		return nil, errors.NewAnvilErrorWithType(constants.OpUpdate, "curl", errors.ErrorTypeInstallation,
			fmt.Errorf("curl is required for updating Anvil but is not available"))
	}

	o.PrintStage("Downloading and executing update script...")
	o.PrintInfo("Fetching latest version from GitHub releases...")

	// Try downloading from releases first, fallback to main branch if that fails
	// This handles cases where the install.sh in releases hasn't been updated yet
	updateScript := fmt.Sprintf(`set -e
		if ! curl -sfSL %s -o /tmp/anvil-install.sh 2>/dev/null; then
			echo "⚠️  Install script not found in releases, trying main branch..."
			curl -sfSL %s -o /tmp/anvil-install.sh || {
				echo "❌ Failed to download install script"
				exit 1
			}
		fi
		bash /tmp/anvil-install.sh
		rm -f /tmp/anvil-install.sh`, constants.UpdateReleaseScriptURL, constants.UpdateMainScriptURL)

	// Execute the update command using the existing system package
	result, err := system.RunCommandWithTimeout(
		ctx,
		"bash",
		"-c",
		updateScript,
	)

	return result, err
}
func init() {
	UpdateCmd.Flags().Bool("dry-run", false, "Show what would be updated without actually updating")
}
