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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xjuanma/anvil/internal/brew"
	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/installer"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/tools"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

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
				palantir.GetGlobalOutputHandler().PrintError("Failed to load application data: %v", err)
				return
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
			return
		}

		if len(args) == 0 {
			return nil
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
		return installGroup(target, tools, dryRun, concurrent, maxWorkers, timeout)
	}

	// If not a group, treat as individual application
	return installIndividualApp(target, dryRun, cmd)
}

// installGroup installs all tools in a group.
func installGroup(groupName string, tools []string, dryRun bool, concurrent bool, maxWorkers int, timeout time.Duration) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Installing '%s' group", groupName))

	if len(tools) == 0 {
		return errors.NewInstallationError(constants.OpInstall, groupName,
			fmt.Errorf("group '%s' has no tools defined", groupName))
	}

	// Deduplicate tools within the group and update settings if needed
	deduplicatedTools, err := deduplicateGroupTools(groupName, tools)
	if err != nil {
		o.PrintWarning("Failed to deduplicate group tools: %v", err)
	} else {
		tools = deduplicatedTools
	}

	o.PrintInfo("Installing %d tools: %s", len(tools), strings.Join(tools, ", "))

	if concurrent {
		return installGroupConcurrent(groupName, tools, dryRun, maxWorkers, timeout)
	}

	return installGroupSerial(groupName, tools, dryRun)
}

// deduplicateGroupTools removes duplicate tools within a group and updates the settings file.
func deduplicateGroupTools(groupName string, tools []string) ([]string, error) {
	seen := make(map[string]struct{}, len(tools))
	deduplicatedTools := make([]string, 0, len(tools))
	var duplicatesFound []string

	// Deduplicate
	for _, tool := range tools {
		if _, exists := seen[tool]; !exists {
			seen[tool] = struct{}{}
			deduplicatedTools = append(deduplicatedTools, tool)
		} else {
			duplicatesFound = append(duplicatesFound, tool)
		}
	}

	// Return original list if no duplicates found
	if len(duplicatesFound) == 0 {
		return tools, nil
	}

	o := palantir.GetGlobalOutputHandler()
	o.PrintWarning("Found duplicates in group '%s': %s", groupName, strings.Join(duplicatesFound, ", "))
	o.PrintInfo("Removing duplicates from settings file...")

	// Update the configuration with deduplicated tools
	if err := config.UpdateGroupTools(groupName, deduplicatedTools); err != nil {
		return tools, fmt.Errorf("failed to update group with deduplicated tools: %w", err)
	}

	o.PrintSuccess(fmt.Sprintf("Successfully removed %d duplicate(s) from group '%s'", len(duplicatesFound), groupName))
	return deduplicatedTools, nil
}

// installGroupConcurrent installs tools concurrently.
func installGroupConcurrent(groupName string, tools []string, dryRun bool, maxWorkers int, timeout time.Duration) error {
	o := palantir.GetGlobalOutputHandler()

	// Create new output handler to send into concurrent installer
	outputHandler := palantir.NewDefaultOutputHandler()
	concurrentInstaller := installer.NewConcurrentInstaller(maxWorkers, outputHandler, dryRun)

	if timeout > 0 {
		concurrentInstaller.SetTimeout(timeout)
	}

	// Create context with potential cancellation
	ctx := context.Background()
	stats, err := concurrentInstaller.InstallTools(ctx, tools)

	// Track successfully installed apps
	if !dryRun && stats != nil && stats.SuccessfulTools > 0 {
		o.PrintInfo("Updating settings to track installed apps...")
		o.PrintInfo("Group installation tracking not implemented yet")
	}

	return err
}

// toolStatus represents the status of a tool installation.
type toolStatus struct {
	name   string
	status string // pending, installing, done, failed
	emoji  string
}

const (
	toolStatusPending   = "pending"
	toolStatusInstalling = "installing"
	toolStatusDone      = "done"
	toolStatusFailed    = "failed"
)

// installGroupSerial installs tools serially using unified installation logic.
func installGroupSerial(groupName string, tools []string, dryRun bool) error {
	o := palantir.GetGlobalOutputHandler()

	successCount := 0
	var installErrors []string

	// Initialize tool statuses
	toolStatuses := make([]toolStatus, len(tools))
	for i, tool := range tools {
		toolStatuses[i] = toolStatus{
			name:   tool,
			status: toolStatusPending,
			emoji:  "⋯",
		}
	}

	for i, tool := range tools {
		// Update status to installing
		toolStatuses[i].status = toolStatusInstalling
		toolStatuses[i].emoji = "⠋"

		// Print dashboard
		printInstallDashboard(groupName, toolStatuses, i+1, len(tools))

		// Use unified installation logic
		_, err := installSingleToolUnified(tool, dryRun)

		if err != nil {
			toolStatuses[i].status = toolStatusFailed
			toolStatuses[i].emoji = "✗"
			errorMsg := fmt.Sprintf("%s: %v", tool, err)
			installErrors = append(installErrors, errorMsg)
			o.PrintError("%s: %v", tool, err)
		} else {
			toolStatuses[i].status = toolStatusDone
			toolStatuses[i].emoji = "✓"
			successCount++
		}

		// Print final dashboard state
		printInstallDashboard(groupName, toolStatuses, i+1, len(tools))
	}

	return reportGroupInstallationResults(groupName, successCount, len(tools), installErrors)
}

// printInstallDashboard displays the current installation progress.
func printInstallDashboard(groupName string, statuses []toolStatus, current, total int) {
	var content strings.Builder
	content.WriteString("\n")

	// Show each tool with its status
		for i, status := range statuses {
			var statusText string
			switch status.status {
			case toolStatusDone:
				statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, "Installed")
			case toolStatusFailed:
				statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, "Failed")
			case toolStatusInstalling:
				statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, "Installing...")
			default:
				statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, "Pending")
			}

		content.WriteString(fmt.Sprintf("  [%d/%d] %s\n", i+1, total, statusText))
	}

	content.WriteString("\n")

	// Calculate progress
	percentage := (current * 100) / total
	barWidth := 30
	filled := (percentage * barWidth) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	content.WriteString(fmt.Sprintf("  Progress: %d%% %s\n", percentage, bar))

	// Clear previous output and print new dashboard
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
	fmt.Println(charm.RenderBox(fmt.Sprintf("Installing '%s' group (%d tools)", groupName, total), content.String(), "#00D9FF", false))
}

// installIndividualApp installs a single application using unified installation logic.
func installIndividualApp(appName string, dryRun bool, cmd *cobra.Command) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Installing '%s'", appName))

	// Validate app name is not empty
	if appName == "" {
		return errors.NewInstallationError(constants.OpInstall, appName,
			fmt.Errorf("application name cannot be empty"))
	}

	wasNewlyInstalled, err := installSingleToolUnified(appName, dryRun)
	if err != nil {
		return errors.NewInstallationError(constants.OpInstall, appName,
			fmt.Errorf("failed to install '%s'. Please verify the name is correct. You can search for packages using 'brew search %s'", appName, appName))
	}

	// Only track the app in settings if it was newly installed and not dry-run
	if !dryRun && wasNewlyInstalled {
		// Check if --group-name flag is provided
		groupName, _ := cmd.Flags().GetString("group-name")
		if groupName != "" {
			// Add app to the specified group
			if err := config.AddAppToGroup(groupName, appName); err != nil {
				o.PrintWarning("Failed to add %s to group '%s': %v", appName, groupName, err)
				// Continue with normal tracking as fallback
				return trackAppInSettings(appName)
			}
			o.PrintSuccess(fmt.Sprintf("Added %s to group '%s'", appName, groupName))
			return nil
		} else {
			// Normal tracking in installed_apps
			return trackAppInSettings(appName)
		}
	}

	return nil
}

// installSingleTool installs a single tool, handling special cases dynamically.
func installSingleTool(toolName string) error {
	o := palantir.GetGlobalOutputHandler()

	// Check if source is configured for this app (user explicitly configured it)
	sourceURL, exists, sourceErr := installer.GetSourceURL(toolName)
	if sourceErr != nil {
		o.PrintWarning("Failed to check source URL for %s: %v", toolName, sourceErr)
		// Fall back to brew if we can't check source
		return brew.InstallPackageDirectly(toolName)
	}

	// If source exists, try it first (user explicitly configured it)
	if exists && sourceURL != "" {
		o.PrintInfo("Installing %s from configured source", toolName)
		if err := installer.InstallFromSource(toolName, sourceURL); err != nil {
			// Source installation failed, fall back to brew
			o.PrintInfo("Source installation failed, falling back to brew for %s", toolName)
			return brew.InstallPackageDirectly(toolName)
		}
		// Source installation succeeded, continue with post-install steps
	} else {
		// No source configured, use brew (default for majority of apps)
		if err := brew.InstallPackageDirectly(toolName); err != nil {
			return err
		}
	}

	// Handle config check for git
	if toolName == "git" {
		if err := checkToolConfiguration(toolName); err != nil {
			o.PrintWarning("Configuration check failed for %s: %v", toolName, err)
		}
	}

	return nil
}

// installSingleToolUnified provides unified installation logic for all installation modes.
func installSingleToolUnified(toolName string, dryRun bool) (wasNewlyInstalled bool, err error) {
	o := palantir.GetGlobalOutputHandler()

	// ALWAYS check availability first using the latest IsApplicationAvailable logic
	if brew.IsApplicationAvailable(toolName) {
		o.PrintAlreadyAvailable("%s is already available on the system", toolName)
		return false, nil
	}

	// Handle installation based on mode
	if dryRun {
		o.PrintInfo("Would install: %s", toolName)
		return true, nil
	}

	// Perform real installation using existing logic
	if err := installSingleTool(toolName); err != nil {
		return false, err
	}

	o.PrintSuccess(fmt.Sprintf("%s installed successfully", toolName))
	return true, nil
}

// trackAppInSettings handles adding newly installed apps to settings.
func trackAppInSettings(appName string) error {
	o := palantir.GetGlobalOutputHandler()
	// Check if already tracked to avoid duplicates
	if isTracked, err := config.IsAppTracked(appName); err != nil {
		o.PrintWarning("Failed to check if %s is already tracked: %v", appName, err)
		return nil // Don't fail installation for tracking issues
	} else if isTracked {
		return nil
	}

	spinner := charm.NewDotsSpinner(fmt.Sprintf("Tracking %s in settings", appName))
	spinner.Start()

	if err := config.AddInstalledApp(appName); err != nil {
		spinner.Warning("Failed to update settings")
		o.PrintWarning("Failed to update settings file: %v", err)
		return nil // Don't fail installation for tracking issues
	}

	spinner.Success(fmt.Sprintf("%s tracked in settings", appName))
	return nil
}

// reportGroupInstallationResults provides unified error reporting for group installations.
func reportGroupInstallationResults(groupName string, successCount, totalCount int, installErrors []string) error {
	// Print summary
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Group Installation Complete")
	o.PrintInfo("Successfully installed %d of %d tools", successCount, totalCount)

	if len(installErrors) > 0 {
		o.PrintWarning("Some installations failed:")
		for _, err := range installErrors {
			o.PrintError("  • %s", err)
		}
		return errors.NewInstallationError(constants.OpInstall, groupName,
			fmt.Errorf("failed to install %d tools", len(installErrors)))
	}

	return nil
}

// checkToolConfiguration checks if a tool is properly configured.
func checkToolConfiguration(toolName string) error {
	switch toolName {
	case constants.PkgGit:
		return checkGitConfiguration()
	default:
		return nil
	}
}

// checkGitConfiguration checks if git is properly configured.
func checkGitConfiguration() error {
	config, err := config.LoadConfig()
	if err == nil && (config.Git.Username == "" || config.Git.Email == "") {
		o := palantir.GetGlobalOutputHandler()
		o.PrintInfo("Git installed successfully")
		o.PrintWarning("Consider configuring git with:")
		o.PrintInfo("  git config --global user.name 'Your Name'")
		o.PrintInfo("  git config --global user.email 'your.email@example.com'")
	}
	return nil
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
