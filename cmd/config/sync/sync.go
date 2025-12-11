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

// Package sync provides functionality to sync pulled configuration files
// to their local destinations with automatic archiving of existing files.
package sync

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

var SyncCmd = &cobra.Command{
	Use:   "sync [app-name]",
	Short: "Sync pulled configuration files to their local destinations",
	Long:  constants.SYNC_COMMAND_LONG_DESCRIPTION,
	Args:  cobra.MaximumNArgs(1), // Accept 0 or 1 argument
	RunE: func(cmd *cobra.Command, args []string) error {
		return runSyncCommand(cmd, args)
	},
}

// runSyncCommand executes the configuration sync process.
func runSyncCommand(cmd *cobra.Command, args []string) error {
	// Check for dry-run flag
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	// If no arguments provided, sync the anvil settings
	if len(args) == 0 {
		return syncAnvilSettings(dryRun)
	}

	// Sync specific app config
	appName := args[0]
	return syncAppConfig(appName, dryRun)
}

// syncAnvilSettings syncs the main anvil settings.yaml file.
func syncAnvilSettings(dryRun bool) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Configuration Sync: Anvil settings")

	tempSettingsPath := filepath.Join(config.GetAnvilConfigDirectory(), "temp", constants.ANVIL, constants.ANVIL_CONFIG_FILE)
	if _, err := os.Stat(tempSettingsPath); os.IsNotExist(err) {
		o.PrintError("Pulled anvil settings not found\n")
		o.PrintInfo("💡 No pulled settings found at: %s", tempSettingsPath)
		o.PrintInfo("🔧 To fix this:")
		o.PrintInfo("   • Run 'anvil config pull anvil' to download settings")
		o.PrintInfo("   • Ensure your repository has an 'anvil' directory with settings.yaml")
		return fmt.Errorf("config not pulled yet")
	}

	currentSettingsPath := config.GetAnvilConfigPath()

	o.PrintInfo("Source: %s", tempSettingsPath)
	o.PrintInfo("Destination: %s\n", currentSettingsPath)

	if dryRun {
		o.PrintInfo("Dry run - would sync anvil settings")
		return nil
	}

	opts := SyncOptions{
		ArchivePrefix:  "anvil-settings",
		SourcePath:     tempSettingsPath,
		DestPath:       currentSettingsPath,
		ConfirmMsg:     fmt.Sprintf("Sync local %s? Old copy will be archived.", constants.ANVIL_CONFIG_FILE),
		SpinnerMsg:     "Syncing anvil settings",
		SpinnerSuccess: "[Anvil] settings synced successfully",
		SuccessMsg:     "Sync done!",
	}
	return performSync(opts)
}

// syncAppConfig syncs configuration files for a specific app.
func syncAppConfig(appName string, dryRun bool) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintHeader(fmt.Sprintf("Configuration Sync: %s", appName))

	cfg, err := config.LoadConfig()
	if err != nil {
		return errors.NewConfigurationError(constants.OpSync, "load-config", err)
	}

	tempAppPath := filepath.Join(config.GetAnvilConfigDirectory(), "temp", appName)
	if _, err := os.Stat(tempAppPath); os.IsNotExist(err) {
		output.PrintError("Pulled %s configuration not found\n", appName)
		output.PrintInfo("💡 No pulled config found at: %s", tempAppPath)
		output.PrintInfo("🔧 To fix this:")
		output.PrintInfo("   • Run 'anvil config pull %s' to download configuration", appName)
		output.PrintInfo("   • Ensure your repository has a '%s' directory", appName)
		return fmt.Errorf("config not pulled yet")
	}

	if cfg.Configs == nil {
		return fmt.Errorf("no configs section found in %s", constants.ANVIL_CONFIG_FILE)
	}

	localConfigPath, exists := cfg.Configs[appName]
	if !exists {
		output.PrintError("App config path not configured\n")
		output.PrintInfo("💡 The app '%s' doesn't have a local config path defined", appName)
		output.PrintInfo("🔧 To fix this:")
		output.PrintInfo("   • Edit your %s file", constants.ANVIL_CONFIG_FILE)
		output.PrintInfo("   • Add the following to the 'configs' section:\n")
		output.PrintInfo("configs:")
		output.PrintInfo("  %s: \"/path/to/%s/config\"\n", appName, appName)
		output.PrintInfo("Example paths:")
		output.PrintInfo("  • ~/.config/%s", appName)
		output.PrintInfo("  • ~/Library/Application Support/%s", strings.Title(appName))
		return fmt.Errorf("app config path not defined")
	}

	output.PrintInfo("Source: %s", tempAppPath)
	output.PrintInfo("Destination: %s\n", localConfigPath)

	if dryRun {
		output.PrintInfo("Dry run - would sync %s configuration", appName)
		return nil
	}

	opts := SyncOptions{
		ArchivePrefix:  fmt.Sprintf("%s-configs", appName),
		SourcePath:     tempAppPath,
		DestPath:       localConfigPath,
		ConfirmMsg:     fmt.Sprintf("Sync %s configs? Old copy will be archived.", appName),
		SpinnerMsg:     fmt.Sprintf("Syncing %s configuration", appName),
		SpinnerSuccess: fmt.Sprintf("[%s] configuration synced successfully", strings.Title(appName)),
		SuccessMsg:     "Sync done!",
	}
	return performSync(opts)
}

// performSync executes the core sync operation for any config type.
func performSync(opts SyncOptions) error {
	output := palantir.GetGlobalOutputHandler()

	archivePath, err := createArchiveDirectory(opts.ArchivePrefix)
	if err != nil {
		return fmt.Errorf("failed to create archive directory: %w", err)
	}

	output.PrintInfo("Archive: %s\n", archivePath)

	if os.Getenv("ANVIL_TEST_MODE") != "true" {
		if !output.Confirm(opts.ConfirmMsg) {
			output.PrintInfo("Sync cancelled")
			return nil
		}
	}

	fmt.Println("")

	spinner := charm.NewDotsSpinner(opts.SpinnerMsg)
	spinner.Start()

	if err := archiveExistingConfig(opts.ArchivePrefix, opts.DestPath, archivePath); err != nil {
		spinner.Error("Failed to archive existing config")
		return fmt.Errorf("failed to archive existing config: %w", err)
	}

	sourceInfo, err := os.Stat(opts.SourcePath)
	if err != nil {
		spinner.Error("Failed to read source")
		return fmt.Errorf("failed to read source: %w", err)
	}

	if sourceInfo.IsDir() {
		err = utils.CopyDirectorySimple(opts.SourcePath, opts.DestPath)
	} else {
		err = utils.CopyFileSimple(opts.SourcePath, opts.DestPath)
	}

	if err != nil {
		spinner.Error("Failed to copy new config")
		return fmt.Errorf("failed to copy new config: %w", err)
	}

	spinner.Success(opts.SpinnerSuccess)

	output.PrintSuccess(opts.SuccessMsg)
	output.PrintInfo("Old configs archived to: %s", archivePath)
	output.PrintInfo("Manual recovery possible from archive directory (no auto-recover yet)")

	return nil
}

func init() {
	SyncCmd.Flags().Bool("dry-run", false, "Show what would be synced without making changes")
}
