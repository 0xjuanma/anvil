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
	displayInitBanner()

	if err := validateAndInstallInitTools(); err != nil {
		return err
	}

	if err := createInitDirectories(); err != nil {
		return err
	}

	if err := generateInitSettings(); err != nil {
		return err
	}

	warnings := checkInitEnvironment()

	discoverFlag, _ := cmd.Flags().GetBool("discover")
	if discoverFlag {
		if err := runInitDiscovery(); err != nil {
			return err
		}
	}

	return displayInitCompletion(warnings)
}

// displayInitBanner displays the initialization banner.
func displayInitBanner() {
	fmt.Println(charm.RenderBox("🔨 ANVIL INITIALIZATION", "", "#00D9FF", true))
	fmt.Println()
}

// validateAndInstallInitTools validates and installs required tools.
func validateAndInstallInitTools() error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintStage("Stage 1: Tool Validation")
	spinner := charm.NewCircleSpinner(constants.SpinnerValidatingTools)
	spinner.Start()
	if err := tools.ValidateAndInstallTools(); err != nil {
		spinner.Error("Tool validation failed")
		return errors.NewValidationError(constants.OpInit, "validate-tools", err)
	}
	spinner.Success("All required tools are available")
	return nil
}

// createInitDirectories creates necessary directories.
func createInitDirectories() error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintStage("Stage 2: Directory Creation")
	spinner := charm.NewDotsSpinner(constants.SpinnerCreatingDirectories)
	spinner.Start()
	if err := config.CreateDirectories(); err != nil {
		spinner.Error("Failed to create directories")
		return errors.NewFileSystemError(constants.OpInit, "create-directories", err)
	}
	spinner.Success("Directories created successfully")
	return nil
}

// generateInitSettings generates default settings.yaml.
func generateInitSettings() error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintStage("Stage 3: Settings Generation")
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Generating default %s", constants.ANVIL_CONFIG_FILE))
	spinner.Start()
	if err := config.GenerateDefaultSettings(); err != nil {
		spinner.Error("Failed to generate settings")
		return errors.NewConfigurationError(constants.OpInit, "generate-settings", err)
	}
	spinner.Success(fmt.Sprintf("Default %s generated", constants.ANVIL_CONFIG_FILE))
	return nil
}

// checkInitEnvironment checks local environment configurations.
func checkInitEnvironment() []string {
	o := palantir.GetGlobalOutputHandler()
	o.PrintStage("Stage 4: Environment Check")
	spinner := charm.NewDotsSpinner(constants.SpinnerCheckingEnvironment)
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
	return warnings
}

// runInitDiscovery runs the app discovery logic.
func runInitDiscovery() error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintStage("Stage 5: App Discovery Logic")
	spinner := charm.NewDotsSpinner(constants.SpinnerRunningDiscovery)
	spinner.Start()
	if err := config.RunDiscoverLogic(); err != nil {
		spinner.Error("Failed to run app discovery logic")
		return errors.NewConfigurationError(constants.OpInit, "run-discover-logic", err)
	}
	spinner.Success("App discovery logic completed")
	return nil
}

// displayInitCompletion displays completion message and next steps.
func displayInitCompletion(warnings []string) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Initialization Complete!")
	o.PrintInfo("Anvil has been successfully initialized and is ready to use.")
	o.PrintInfo("Configuration files have been created in: %s", config.GetAnvilConfigPath())

	if len(warnings) > 0 {
		fmt.Println("")
		o.PrintInfo("Recommended next steps to complete your setup:")
		for _, warning := range warnings {
			o.PrintInfo("  • %s", warning)
		}
		fmt.Println("")
		o.PrintInfo("These steps are optional but recommended for the best experience.")
	}

	fmt.Println("")
	o.PrintInfo("You can now use:")
	o.PrintInfo("  • 'anvil install [group]' to install development tool groups")
	o.PrintInfo("  • 'anvil install [app]' to install any individual application")
	o.PrintInfo("  • Edit %s/%s to customize your configuration", config.GetAnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE)

	o.PrintWarning("Configuration Management Setup Required:")
	o.PrintInfo("  • Edit the 'github.config_repo' field in %s to enable config pull/push", constants.ANVIL_CONFIG_FILE)
	o.PrintInfo("  • Example: 'github.config_repo: username/dotfiles'")
	o.PrintInfo("  • Set GITHUB_TOKEN environment variable for authentication")
	o.PrintInfo("  • Run 'anvil doctor' once added to validate configuration")

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
