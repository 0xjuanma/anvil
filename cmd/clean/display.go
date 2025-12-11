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

package clean

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xjuanma/palantir"
)

// displayCleanPreview shows what will be cleaned.
func displayCleanPreview(output palantir.OutputHandler, itemsToClean []string) {
	output.PrintInfo("Found %d root directories to clean:", len(itemsToClean))
	output.PrintInfo("Directory structure to be cleaned:")

	// Build and display tree structure for each directory
	for _, itemPath := range itemsToClean {
		itemName := filepath.Base(itemPath)
		if info, err := os.Stat(itemPath); err == nil && info.IsDir() {
			// Count items in directory
			count, treeOutput := buildDirectoryTree(itemPath, itemName)
			output.PrintInfo("  📁 %s (%d)", itemName, count)
			fmt.Print(treeOutput)
		} else {
			output.PrintInfo("  📁 %s", itemName)
		}
	}
}

// handleUserConfirmation handles user confirmation and returns true if should proceed.
func handleUserConfirmation(output palantir.OutputHandler, force, dryRun bool, itemCount int) bool {
	// Confirm deletion unless force flag is used
	if !force && !dryRun {
		confirmMsg := fmt.Sprintf("Are you sure you want to clean the contents of these %d root directories? This action cannot be undone", itemCount)
		if !output.Confirm(confirmMsg) {
			output.PrintInfo("Clean operation cancelled.")
			return false
		}
	}
	return true
}

// displayCleanResult shows the result of cleaning a specific item.
func displayCleanResult(output palantir.OutputHandler, itemPath string) {
	itemName := filepath.Base(itemPath)
	if info, err := os.Stat(itemPath); err == nil && info.IsDir() {
		if itemName == "dotfiles" {
			output.PrintSuccess("Removed dotfiles directory completely")
		} else {
			output.PrintSuccess("Cleaned contents of directory " + itemName)
		}
	} else {
		output.PrintSuccess("Cleaned " + itemName)
	}
}

// buildDirectoryTree builds a simple tree showing only immediate contents of a directory.
func buildDirectoryTree(dirPath, dirName string) (int, string) {
	var output strings.Builder
	var count int

	// Read only the immediate contents of the directory
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return 0, ""
	}

	// Count and display only immediate entries
	for _, entry := range entries {
		count++
		output.WriteString(fmt.Sprintf("    ├── %s\n", entry.Name()))
	}

	return count, output.String()
}
