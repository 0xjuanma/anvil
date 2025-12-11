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

// Package initcmd provides initialization functionality for setting up the
// Anvil CLI environment, including tool validation and configuration generation.
package initcmd

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/tools"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

// InitCmd represents the init command for Anvil CLI environment setup.
var InitCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Anvil CLI environment",
	Long:  constants.INIT_COMMAND_LONG_DESCRIPTION,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runInitCommand(cmd)
	},
}

// runInitCommand executes the complete initialization process for Anvil CLI environment.
func runInitCommand(cmd *cobra.Command) error {
	// Display initialization banner
	fmt.Println(charm.RenderBox("🔨 ANVIL INITIALIZATION", "", "#00D9FF", true))
	fmt.Println()

	o := palantir.GetGlobalOutputHandler()

	// Stage 1: Tool validation and installation
	o.PrintStage("Stage 1: Tool Validation")
	spinner := charm.NewCircleSpinner("Validating and installing required tools")
	spinner.Start()
	if err := tools.ValidateAndInstallTools(); err != nil {
		spinner.Error("Tool validation failed")
		return errors.NewValidationError(constants.OpInit, "validate-tools", err)
	}
	spinner.Success("All required tools are available")

	// Stage 2: Create necessary directories
	o.PrintStage("Stage 2: Directory Creation")
	spinner = charm.NewDotsSpinner("Creating necessary directories")
	spinner.Start()
	if err := config.CreateDirectories(); err != nil {
		spinner.Error("Failed to create directories")
		return errors.NewFileSystemError(constants.OpInit, "create-directories", err)
	}
	spinner.Success("Directories created successfully")

	// Stage 3: Generate default settings.yaml
	o.PrintStage("Stage 3: Settings Generation")
	spinner = charm.NewDotsSpinner(fmt.Sprintf("Generating default %s", constants.ANVIL_CONFIG_FILE))
	spinner.Start()
	if err := config.GenerateDefaultSettings(); err != nil {
		spinner.Error("Failed to generate settings")
		return errors.NewConfigurationError(constants.OpInit, "generate-settings", err)
	}
	spinner.Success(fmt.Sprintf("Default %s generated", constants.ANVIL_CONFIG_FILE))

	// Stage 4: Check local environment configurations
	o.PrintStage("Stage 4: Environment Check")
	spinner = charm.NewDotsSpinner("Checking local environment configurations")
	spinner.Start()
	warnings := config.CheckEnvironmentConfigurations()
	if len(warnings) > 0 {
		spinner.Warning("Environment configuration warnings found")
		for _, warning := range warnings {
			o.PrintWarning("  - %s", warning)
		}
	} else {
		spinner.Success("Environment configurations are properly set")
	}
	// Stage 5: Run the discovery logic only if the --discover flag is provided
	discoverFlag, _ := cmd.Flags().GetBool("discover")
	if discoverFlag {
		o.PrintStage("Stage 5: App Discovery Logic")
		spinner = charm.NewDotsSpinner("Running discovery logic")
		spinner.Start()
		if err := config.RunDiscoverLogic(); err != nil {
			spinner.Error("Failed to run app discovery logic")
			return errors.NewConfigurationError(constants.OpInit, "run-discover-logic", err)
		}
		spinner.Success("App discovery logic completed")
	}

	// Stage 6: Print completion message and next steps
	o.PrintHeader("Initialization Complete!")
	o.PrintInfo("Anvil has been successfully initialized and is ready to use.")
	o.PrintInfo("Configuration files have been created in: %s", config.GetAnvilConfigPath())

	// Provide specific guidance if there are configuration warnings
	if len(warnings) > 0 {
		fmt.Println("")
		o.PrintInfo("Recommended next steps to complete your setup:")
		for _, warning := range warnings {
			o.PrintInfo("  • %s", warning)
		}
		fmt.Println("")
		o.PrintInfo("These steps are optional but recommended for the best experience.")
	}

	// Final usage guidance
	fmt.Println("")
	o.PrintInfo("You can now use:")
	o.PrintInfo("  • 'anvil install [group]' to install development tool groups")
	o.PrintInfo("  • 'anvil install [app]' to install any individual application")
	o.PrintInfo("  • Edit %s/%s to customize your configuration", config.GetAnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE)

	// GitHub configuration warning
	o.PrintWarning("Configuration Management Setup Required:")
	o.PrintInfo("  • Edit the 'github.config_repo' field in %s to enable config pull/push", constants.ANVIL_CONFIG_FILE)
	o.PrintInfo("  • Example: 'github.config_repo: username/dotfiles'")
	o.PrintInfo("  • Set GITHUB_TOKEN environment variable for authentication")
	o.PrintInfo("  • Run 'anvil doctor' once added to validate configuration")

	// Show available groups dynamically
	if groups, err := config.GetAvailableGroups(); err == nil {
		builtInGroups := config.GetBuiltInGroups()
		fmt.Println("")
		o.PrintInfo("Available groups: %s", strings.Join(builtInGroups, ", "))
		if len(groups) > len(builtInGroups) {
			o.PrintInfo("Custom groups: %d defined", len(groups)-len(builtInGroups))
		}
	} else {
		o.PrintInfo("Available groups: dev, essentials")
	}
	o.PrintInfo("Example: 'anvil install essentials' or 'anvil install firefox'")

	return nil
}

func init() {
	// Add flags for additional functionality
	InitCmd.Flags().Bool("discover", false, "Run the app/package discovery logic")
}
