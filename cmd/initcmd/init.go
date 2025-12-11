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
	"github.com/0xjuanma/anvil/internal/constants"
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

func init() {
	// Add flags for additional functionality
	InitCmd.Flags().Bool("discover", false, "Run the app/package discovery logic")
}
