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

// Package importcmd provides functionality to import tool groups from local
// files or remote URLs into the anvil configuration.
package importcmd

import (
	"fmt"
	"sort"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

var ImportCmd = &cobra.Command{
	Use:   "import [file-or-url]",
	Short: "Import groups from a local file or URL",
	Long:  "Import tool groups from a local YAML file or remote URL into your anvil configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runImportCommand(cmd, args[0])
	},
}

// ImportConfig represents the structure for importing configurations.
type ImportConfig struct {
	Groups map[string][]string `yaml:"groups"`
}

// runImportCommand executes the group import process.
func runImportCommand(cmd *cobra.Command, importPath string) error {
	output := palantir.GetGlobalOutputHandler()
	output.PrintHeader("Import Groups from File")

	// Stage 1: Fetch and validate source file
	output.PrintStage("Stage 1: Fetching source file...")
	spinner := charm.NewCircleSpinner("Fetching import file")
	spinner.Start()
	tempFile, cleanup, err := fetchFile(importPath)
	if err != nil {
		spinner.Error("Failed to fetch source file")
		return errors.NewFileSystemError(constants.OpConfig, "fetch-file", err)
	}
	defer cleanup()
	spinner.Success("Source file fetched successfully")

	// Stage 2: Parse and validate import data
	output.PrintStage("Parsing import file...")
	importData, err := parseImportFile(tempFile)
	if err != nil {
		return errors.NewConfigurationError(constants.OpConfig, "parse-import", err)
	}

	if len(importData.Groups) == 0 {
		return errors.NewConfigurationError(constants.OpConfig, "no-groups",
			fmt.Errorf("no valid groups found in import file"))
	}
	output.PrintSuccess("Import file parsed successfully")

	// Stage 3: Validate group structure
	output.PrintStage("Validating group structure...")
	if err := validateImportGroups(importData.Groups); err != nil {
		return errors.NewConfigurationError(constants.OpConfig, "validate-groups", err)
	}
	output.PrintSuccess("Group structure validation passed")

	// Stage 4: Check for conflicts with existing groups
	output.PrintStage("Checking for conflicts...")
	currentConfig, err := config.LoadConfig()
	if err != nil {
		return errors.NewConfigurationError(constants.OpConfig, "load-config", err)
	}

	conflicts := checkGroupConflicts(importData.Groups, currentConfig.Groups)
	if len(conflicts) > 0 {
		return errors.NewConfigurationError(constants.OpConfig, "group-conflicts",
			fmt.Errorf("groups already exist: %s", strings.Join(conflicts, ", ")))
	}
	output.PrintSuccess("No conflicts detected")

	// Stage 5: Display import summary
	output.PrintStage("Preparing import summary...")
	displayImportSummary(importData.Groups)

	// Stage 6: Confirm import
	if !output.Confirm("Proceed with importing these groups?") {
		output.PrintInfo("Import cancelled by user")
		return nil
	}

	// Stage 7: Import groups
	output.PrintStage("Stage 7: Importing groups...")
	spinner = charm.NewDotsSpinner(fmt.Sprintf("Importing %d groups", len(importData.Groups)))
	spinner.Start()
	if err := importGroups(currentConfig, importData.Groups); err != nil {
		spinner.Error("Failed to import groups")
		return errors.NewConfigurationError(constants.OpConfig, "import-groups", err)
	}
	spinner.Success(fmt.Sprintf("Successfully imported %d groups", len(importData.Groups)))

	output.PrintInfo("\n✨ Import completed! %d groups have been added to your configuration.", len(importData.Groups))
	return nil
}

// validateImportGroups validates the structure of imported groups.
func validateImportGroups(groups map[string][]string) error {
	if len(groups) == 0 {
		return fmt.Errorf("no groups found to import")
	}

	validator := config.NewConfigValidator(nil)

	for groupName, tools := range groups {
		// Validate group name
		if err := validator.ValidateGroupName(groupName); err != nil {
			return fmt.Errorf("invalid group name '%s': %w", groupName, err)
		}

		// Validate group is not empty
		if len(tools) == 0 {
			return fmt.Errorf("group '%s' cannot be empty", groupName)
		}

		// Validate each tool name
		for _, tool := range tools {
			if err := validator.ValidateAppName(tool); err != nil {
				return fmt.Errorf("invalid tool '%s' in group '%s': %w", tool, groupName, err)
			}
		}
	}

	return nil
}

// checkGroupConflicts checks if any imported groups already exist.
func checkGroupConflicts(importGroups map[string][]string, existingGroups config.AnvilGroups) []string {
	var conflicts []string
	for groupName := range importGroups {
		if _, exists := existingGroups[groupName]; exists {
			conflicts = append(conflicts, groupName)
		}
	}
	sort.Strings(conflicts)
	return conflicts
}

// displayImportSummary shows a tree view of groups that will be imported.
func displayImportSummary(groups map[string][]string) {
	output := palantir.GetGlobalOutputHandler()
	fmt.Println("")
	output.PrintInfo("📋 Import Summary:")
	output.PrintInfo("═══════════════════")

	// Sort group names for consistent output
	var groupNames []string
	for groupName := range groups {
		groupNames = append(groupNames, groupName)
	}
	sort.Strings(groupNames)

	totalGroups := len(groups)
	totalApps := 0

	for _, groupName := range groupNames {
		tools := groups[groupName]
		totalApps += len(tools)

		// Display group with tree structure
		output.PrintInfo("├── 📁 %s (%d tools)", groupName, len(tools))

		// Sort tools for consistent output
		sortedTools := make([]string, len(tools))
		copy(sortedTools, tools)
		sort.Strings(sortedTools)

		for i, tool := range sortedTools {
			if i == len(sortedTools)-1 {
				output.PrintInfo("│   └── 🔧 %s", tool)
			} else {
				output.PrintInfo("│   ├── 🔧 %s", tool)
			}
		}
		output.PrintInfo("│")
	}

	output.PrintInfo("Total: %d groups, %d applications", totalGroups, totalApps)
	fmt.Println("")
}

// importGroups adds the imported groups to the current configuration.
func importGroups(currentConfig *config.AnvilConfig, importGroups map[string][]string) error {
	// Add new groups to existing configuration
	for groupName, tools := range importGroups {
		currentConfig.Groups[groupName] = tools
	}

	// Save updated configuration
	return config.SaveConfig(currentConfig)
}
