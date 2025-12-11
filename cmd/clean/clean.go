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

// Package clean provides functionality to clean all content inside .anvil
// directories while preserving the directory structure and main config file.
package clean

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean all content inside .anvil directories",
	Long:  constants.CLEAN_COMMAND_LONG_DESCRIPTION,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runCleanCommand(cmd, args)
	},
}

// runCleanCommand executes the clean process, scanning for items and
// performing cleanup after user confirmation.
func runCleanCommand(cmd *cobra.Command, args []string) error {
	// Get command flags
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")
	output := palantir.GetGlobalOutputHandler()
	output.PrintHeader("Cleaning Anvil Directories")

	// Get anvil directory path
	anvilDir, err := getAnvilDirectoryPath()
	if err != nil {
		return err
	}

	// Check if .anvil directory exists
	if _, err := os.Stat(anvilDir); os.IsNotExist(err) {
		output.PrintWarning("Directory %s does not exist. Nothing to clean.", anvilDir)
		return nil
	}

	// Get items to clean
	itemsToClean, err := getItemsToClean(anvilDir)
	if err != nil {
		return err
	}

	if len(itemsToClean) == 0 {
		output.PrintSuccess(fmt.Sprintf("No root directories found to clean. Only %s exists.", constants.ANVIL_CONFIG_FILE))
		return nil
	}

	// Display what will be cleaned
	displayCleanPreview(output, itemsToClean)

	// Handle user confirmation and dry run
	if !handleUserConfirmation(output, force, dryRun, len(itemsToClean)) {
		return nil
	}

	if dryRun {
		output.PrintInfo("DRY RUN: Would clean contents of %d root directories", len(itemsToClean))
		return nil
	}

	// Perform the actual cleaning
	return performCleaning(output, itemsToClean)
}

// getAnvilDirectoryPath returns the path to the .anvil directory.
func getAnvilDirectoryPath() (string, error) {
	homeDir, err := system.GetHomeDir()
	if err != nil {
		return "", &errors.AnvilError{
			Op:      "clean",
			Command: "clean",
			Type:    errors.ErrorTypeFileSystem,
			Err:     fmt.Errorf("failed to get home directory: %w", err),
		}
	}
	return filepath.Join(homeDir, constants.ANVIL_CONFIG_DIR), nil
}

// getItemsToClean scans the anvil directory and returns items to clean.
func getItemsToClean(anvilDir string) ([]string, error) {
	output := palantir.GetGlobalOutputHandler()
	output.PrintStage(fmt.Sprintf("Scanning %s directory for content to clean", constants.ANVIL_CONFIG_DIR))

	spinner := charm.NewCircleSpinner(fmt.Sprintf("Scanning %s directory", constants.ANVIL_CONFIG_DIR))
	spinner.Start()

	// Get all items in .anvil directory
	items, err := os.ReadDir(anvilDir)
	if err != nil {
		spinner.Error("Failed to scan directory")
		return nil, &errors.AnvilError{
			Op:      "clean",
			Command: "clean",
			Type:    errors.ErrorTypeFileSystem,
			Err:     fmt.Errorf("failed to read .anvil directory: %w", err),
		}
	}

	var itemsToClean []string
	for _, item := range items {
		// Skip Anvil config file
		if item.Name() == constants.ANVIL_CONFIG_FILE {
			continue
		}

		itemPath := filepath.Join(anvilDir, item.Name())
		itemsToClean = append(itemsToClean, itemPath)
	}

	spinner.Success(fmt.Sprintf("Found %d items to clean", len(itemsToClean)))
	return itemsToClean, nil
}

// performCleaning executes the actual cleaning process for all items.
func performCleaning(output palantir.OutputHandler, itemsToClean []string) error {
	output.PrintStage("Cleaning directories and files")

	spinner := charm.NewDotsSpinner(fmt.Sprintf("Cleaning %d items", len(itemsToClean)))
	spinner.Start()

	// Clean each item
	cleanedCount := 0
	for _, itemPath := range itemsToClean {
		if err := cleanItem(itemPath); err != nil {
			output.PrintWarning("Failed to clean %s: %v", filepath.Base(itemPath), err)
			continue
		}
		cleanedCount++
		displayCleanResult(output, itemPath)
	}

	if cleanedCount == len(itemsToClean) {
		spinner.Success(fmt.Sprintf("Successfully cleaned %d items", cleanedCount))
	} else if cleanedCount > 0 {
		spinner.Warning(fmt.Sprintf("Cleaned %d/%d items (some failed)", cleanedCount, len(itemsToClean)))
	} else {
		spinner.Error("Failed to clean items")
	}

	output.PrintInfo("Successfully cleaned contents of %d/%d root directories", cleanedCount, len(itemsToClean))

	if cleanedCount < len(itemsToClean) {
		output.PrintWarning("Some root directories could not be cleaned. Check the warnings above.")
	}

	return nil
}

// cleanItem removes the contents of a directory or the file itself.
// Special handling applies to the dotfiles directory which is removed entirely.
func cleanItem(itemPath string) error {
	info, err := os.Stat(itemPath)
	if err != nil {
		return fmt.Errorf("failed to stat item: %w", err)
	}

	if info.IsDir() {
		itemName := filepath.Base(itemPath)

		// Special handling for dotfiles directory - remove it completely
		if itemName == "dotfiles" {
			// Remove the entire dotfiles directory to ensure clean git repository state
			if err := os.RemoveAll(itemPath); err != nil {
				return fmt.Errorf("failed to remove dotfiles directory: %w", err)
			}
			return nil
		}

		// For other directories (temp/, archive/), remove contents but preserve the directory structure
		// This is important for directories that are needed by the tool but can be empty
		entries, err := os.ReadDir(itemPath)
		if err != nil {
			return fmt.Errorf("failed to read directory contents: %w", err)
		}

		for _, entry := range entries {
			entryPath := filepath.Join(itemPath, entry.Name())
			if entry.IsDir() {
				// Remove subdirectory and all its contents
				if err := os.RemoveAll(entryPath); err != nil {
					return fmt.Errorf("failed to remove subdirectory %s: %w", entry.Name(), err)
				}
			} else {
				// Remove file
				if err := os.Remove(entryPath); err != nil {
					return fmt.Errorf("failed to remove file %s: %w", entry.Name(), err)
				}
			}
		}
	} else {
		// Remove single file
		if err := os.Remove(itemPath); err != nil {
			return fmt.Errorf("failed to remove file: %w", err)
		}
	}

	return nil
}

func init() {
	CleanCmd.Flags().BoolP("dry-run", "n", false, "Show what would be cleaned without actually deleting")
	CleanCmd.Flags().BoolP("force", "f", false, "Skip confirmation prompt")
}
