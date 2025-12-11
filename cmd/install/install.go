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

// Package install provides functionality for installing development tools
// and applications dynamically via Homebrew, supporting both individual
// and group-based installation with serial or concurrent execution.
package install

import (
	"fmt"
	"time"

	"github.com/0xjuanma/anvil/internal/brew"
	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/tools"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

// InstallGroupOptions contains options for installing a group of tools.
type InstallGroupOptions struct {
	GroupName  string
	Tools      []string
	DryRun     bool
	Concurrent bool
	MaxWorkers int
	Timeout    time.Duration
}

// InstallCmd represents the install command.
var InstallCmd = &cobra.Command{
	Use:   "install [group-name|app-name] [--group-name group]",
	Short: "Install development tools and applications dynamically via Homebrew",
	Long:  constants.INSTALL_COMMAND_LONG_DESCRIPTION,
	Args: func(cmd *cobra.Command, args []string) error {
		// Allow no arguments if --list or --tree flag is used
		listFlag, _ := cmd.Flags().GetBool("list")
		treeFlag, _ := cmd.Flags().GetBool("tree")
		if listFlag || treeFlag {
			return nil
		}
		// Otherwise, require exactly one argument
		return cobra.ExactArgs(1)(cmd, args)
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check for tree or list flag
		treeFlag, _ := cmd.Flags().GetBool("tree")
		listFlag, _ := cmd.Flags().GetBool("list")

		if treeFlag || listFlag {
			// Load and prepare data once
			groups, builtInGroupNames, customGroupNames, installedApps, err := tools.LoadAndPrepareAppData()
			if err != nil {
				return fmt.Errorf("failed to load application data: %w", err)
			}

			// Choose rendering based on flag
			var content string
			var title = "Available Applications"
			if treeFlag {
				content = utils.RenderTreeView(groups, builtInGroupNames, customGroupNames, installedApps)
				title = fmt.Sprintf("%s (Tree View)", title)
			} else {
				content = utils.RenderListView(groups, builtInGroupNames, customGroupNames, installedApps)
				title = fmt.Sprintf("%s (List View)", title)
			}

			// Display in box
			fmt.Println(charm.RenderBox(title, content, "#00D9FF", false))
			return nil
		}

		if len(args) == 0 {
			return fmt.Errorf("target required (group name or app name)")
		}
		return runInstallCommand(cmd, args[0])
	},
}

// runInstallCommand executes the dynamic install process.
func runInstallCommand(cmd *cobra.Command, target string) error {
	o := palantir.GetGlobalOutputHandler()

	// Check for dry-run flag
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if dryRun {
		o.PrintInfo("Dry run mode - no actual installations will be performed")
	}

	// Check for concurrent flag
	concurrent, _ := cmd.Flags().GetBool("concurrent")
	maxWorkers, _ := cmd.Flags().GetInt("workers")
	timeout, _ := cmd.Flags().GetDuration("timeout")

	// Ensure Homebrew is installed
	if err := brew.EnsureBrewIsInstalled(); err != nil {
		return fmt.Errorf("install: %w", err)
	}

	// Try to get group tools first
	if tools, err := config.GetGroupTools(target); err == nil {
		opts := InstallGroupOptions{
			GroupName:  target,
			Tools:      tools,
			DryRun:     dryRun,
			Concurrent: concurrent,
			MaxWorkers: maxWorkers,
			Timeout:    timeout,
		}
		return installGroup(opts)
	}

	// If not a group, treat as individual application
	return installIndividualApp(target, dryRun, cmd)
}

func init() {
	// Add flags for additional functionality
	InstallCmd.Flags().Bool("dry-run", false, "Show what would be installed without installing")
	InstallCmd.Flags().Bool("list", false, "List all available groups")
	InstallCmd.Flags().Bool("tree", false, "Display all applications in a tree format")
	InstallCmd.Flags().Bool("update", false, "Update Homebrew before installation")
	InstallCmd.Flags().String("group-name", "", "Add the installed app to a group (creates group if it doesn't exist)")

	// Add concurrent installation flags
	InstallCmd.Flags().Bool("concurrent", false, "Enable concurrent installation for improved performance")
	InstallCmd.Flags().Int("workers", 0, "Number of concurrent workers (default: number of CPU cores)")
	InstallCmd.Flags().Duration("timeout", 0, "Timeout for individual tool installations (default: 10 minutes)")
}
