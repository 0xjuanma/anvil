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
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
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
	githubClient, ctx, cancel, err := setupPullAuthentication(cfg)
	if err != nil {
		return err
	}
	defer cancel()

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

func init() {
	// Add flags for additional functionality
	PullCmd.Flags().Bool("force", false, "Force pull even if local changes exist")
	PullCmd.Flags().String("branch", "", "Override the branch to pull from")
}
