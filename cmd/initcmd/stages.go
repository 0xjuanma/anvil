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

package initcmd

import (
	"fmt"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/tools"
	"github.com/0xjuanma/palantir"
)

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
