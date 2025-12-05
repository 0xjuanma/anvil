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

package show

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

var ShowCmd = &cobra.Command{
	Use:   "show [directory]",
	Short: "Show configuration files from anvil settings or pulled directories",
	Long:  constants.SHOW_COMMAND_LONG_DESCRIPTION,
	Args:  cobra.MaximumNArgs(1), // Accept 0 or 1 argument
	Run: func(cmd *cobra.Command, args []string) {
		if err := runShowCommand(cmd, args); err != nil {
			palantir.GetGlobalOutputHandler().PrintError("Show failed: %v", err)
			return
		}
	},
	Example: `  anvil config show                    # Show full anvil settings
  anvil config show --groups          # Show only groups
  anvil config show --configs         # Show only config sources
  anvil config show --sources         # Show only installation sources
  anvil config show --git             # Show only git configuration
  anvil config show --github          # Show only GitHub configuration
  anvil config show myapp             # Show pulled configuration for 'myapp'`,
}

func init() {
	ShowCmd.Flags().Bool("raw", false, "Show raw file content without formatting")
	ShowCmd.Flags().BoolP("groups", "g", false, "Show only groups (only applicable for anvil settings)")
	ShowCmd.Flags().BoolP("configs", "c", false, "Show only config source directories (only applicable for anvil settings)")
	ShowCmd.Flags().BoolP("sources", "s", false, "Show only installation sources (only applicable for anvil settings)")
	ShowCmd.Flags().Bool("git", false, "Show only git configuration (only applicable for anvil settings)")
	ShowCmd.Flags().Bool("github", false, "Show only GitHub configuration (only applicable for anvil settings)")
}

// runShowCommand executes the configuration show process
func runShowCommand(cmd *cobra.Command, args []string) error {
	raw, _ := cmd.Flags().GetBool("raw")
	groups, _ := cmd.Flags().GetBool("groups")
	configs, _ := cmd.Flags().GetBool("configs")
	sources, _ := cmd.Flags().GetBool("sources")
	git, _ := cmd.Flags().GetBool("git")
	github, _ := cmd.Flags().GetBool("github")

	// If no arguments provided, show the anvil config file
	if len(args) == 0 {
		// Check if any specific section flags are set
		if groups || configs || sources || git || github {
			return showAnvilSettingsSection(groups, configs, sources, git, github)
		}
		return showAnvilSettings(raw)
	}

	// Show specific pulled configuration directory
	targetDir := args[0]
	return showPulledConfig(targetDir)
}

func checkSettingsFileExists(o palantir.OutputHandler, configPath string) error {
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		o.PrintError("Anvil settings file not found at: %s", configPath)
		o.PrintInfo("💡 Run 'anvil init' to create the initial settings file")
		return fmt.Errorf("settings file not found")
	}
	return nil
}

// showAnvilSettings displays the main anvil config file
func showAnvilSettings(raw bool) error {
	o := palantir.GetGlobalOutputHandler()

	// Stage 1: Locate settings file
	configPath := config.GetAnvilConfigPath()

	// Check settings file
	err := checkSettingsFileExists(o, configPath)
	if err != nil {
		return err
	}

	// Stage 2: Read and display content
	content, err := os.ReadFile(configPath)
	if err != nil {
		return errors.NewFileSystemError(constants.OpShow, "read-settings", err)
	}

	// If raw flag is set, show raw content
	if raw {
		o.PrintHeader("Anvil Settings Configuration (Raw)")
		o.PrintInfo("File: %s\n", configPath)
		fmt.Print(string(content))
		return nil
	}

	// Enhanced view with box
	var boxContent strings.Builder
	boxContent.WriteString("\n")
	boxContent.WriteString(fmt.Sprintf("  Location: %s\n", configPath))
	boxContent.WriteString("\n")

	// Parse and display YAML content with slight indentation
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if line != "" {
			boxContent.WriteString("  " + line + "\n")
		}
	}

	boxContent.WriteString("\n")

	// Display in box
	fmt.Println(charm.RenderBox(fmt.Sprintf("anvil %s", constants.ANVIL_CONFIG_FILE), boxContent.String(), "#00FF87", false))

	// Footer with helpful info
	fmt.Println()
	fmt.Println("  💡 Edit with: nano " + configPath)
	fmt.Println("  💡 Show raw: anvil config show --raw")
	fmt.Println()

	return nil
}

// showPulledConfig displays configuration files from a pulled directory
func showPulledConfig(targetDir string) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Configuration Directory: %s", targetDir))

	// Stage 1: Load anvil configuration
	o.PrintStage("Loading anvil configuration...")
	o.PrintSuccess("Configuration loaded")

	// Stage 2: Locate pulled configuration directory
	o.PrintStage("Locating pulled configuration directory...")
	tempDir := filepath.Join(config.GetAnvilConfigDirectory(), "temp", targetDir)

	// Check if the directory exists
	if _, err := os.Stat(tempDir); os.IsNotExist(err) {
		o.PrintError("Configuration directory '%s' not found\n", targetDir)
		o.PrintInfo("💡 This could be because:")
		o.PrintInfo("   • The app name is incorrect")
		o.PrintInfo("   • The configuration was never pulled")
		o.PrintInfo("   • Use 'anvil config pull %s' to pull this configuration first", targetDir)
		fmt.Println("")

		// Show available pulled configurations
		tempBasePath := filepath.Join(config.GetAnvilConfigDirectory(), "temp")
		if entries, err := os.ReadDir(tempBasePath); err == nil && len(entries) > 0 {
			o.PrintInfo("Available pulled configurations:")
			for _, entry := range entries {
				if entry.IsDir() {
					o.PrintInfo("  • %s", entry.Name())
				}
			}
		} else {
			o.PrintInfo("No configurations have been pulled yet.")
			o.PrintInfo("Use 'anvil config pull <directory>' to pull configurations from your repository.")
		}

		return fmt.Errorf("configuration directory not found")
	}
	o.PrintSuccess("Configuration directory located")
	o.PrintInfo("Directory: %s\n", tempDir)

	// Stage 3: Display directory contents
	o.PrintStage("Reading configuration files...")
	err := showDirectoryTree(tempDir, targetDir)
	if err != nil {
		return err
	}
	o.PrintSuccess("Configuration files displayed")

	return nil
}

// showSingleFile displays the content of a single configuration file
func showSingleFile(filePath, targetDir string) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Configuration: %s", targetDir))
	o.PrintInfo("File: %s\n", filepath.Base(filePath))

	// Read and display the file content
	content, err := os.ReadFile(filePath)
	if err != nil {
		return errors.NewFileSystemError(constants.OpShow, "read-config-file", err)
	}

	fmt.Print(string(content))
	return nil
}

func init() {
}
