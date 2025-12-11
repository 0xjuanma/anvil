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

// Package cmd provides the root command and command structure for the Anvil CLI.
// It sets up the command hierarchy and provides the main entry point for all
// Anvil subcommands.
package cmd

import (
	"fmt"

	"github.com/0xjuanma/anvil/cmd/clean"
	"github.com/0xjuanma/anvil/cmd/config"
	"github.com/0xjuanma/anvil/cmd/doctor"
	"github.com/0xjuanma/anvil/cmd/initcmd"
	"github.com/0xjuanma/anvil/cmd/install"
	"github.com/0xjuanma/anvil/cmd/update"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   constants.ANVIL,
	Short: "🔥 One CLI to rule them all.",
	Long:  fmt.Sprintf("%s\n\n%s", constants.AnvilLogo, constants.ANVIL_LONG_DESCRIPTION),
	RunE: func(cmd *cobra.Command, args []string) error {
		// Check if version flag was used
		if versionFlag, _ := cmd.Flags().GetBool("version"); versionFlag {
			showVersionInfo()
			return nil
		}

		showWelcomeBanner()
		return nil
	},
}

// Execute runs the root command and handles any errors that occur during
// command execution. This is the main entry point called by main.main().
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return fmt.Errorf("root command execution failed: %w", err)
	}
	return nil
}

func init() {
	rootCmd.AddCommand(initcmd.InitCmd)
	rootCmd.AddCommand(install.InstallCmd)
	rootCmd.AddCommand(config.ConfigCmd)
	rootCmd.AddCommand(doctor.DoctorCmd)
	rootCmd.AddCommand(clean.CleanCmd)
	rootCmd.AddCommand(update.UpdateCmd)

	// Add version flag
	rootCmd.Flags().BoolP("version", "v", false, "Show version information")

	// Set custom help template
	rootCmd.SetHelpFunc(customHelpFunc)
}
