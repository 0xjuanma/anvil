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

package cmd

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/spf13/cobra"
)

// customHelpFunc provides an enhanced help display using Charm UI
func customHelpFunc(cmd *cobra.Command, args []string) {
	// Show logo for root command
	if cmd.Name() == constants.ANVIL {
		fmt.Println(constants.AnvilLogo)
		fmt.Println()
	}

	// Description box - format multiline descriptions with indentation
	if cmd.Long != "" {
		// Remove the logo from long description if present (already shown above)
		description := strings.ReplaceAll(cmd.Long, constants.AnvilLogo, "")
		description = strings.TrimSpace(description)

		// Split into paragraphs and add indentation
		var formattedDesc strings.Builder
		formattedDesc.WriteString("\n")
		paragraphs := strings.Split(description, "\n\n")
		for i, para := range paragraphs {
			lines := strings.Split(para, "\n")
			for _, line := range lines {
				formattedDesc.WriteString("  " + strings.TrimSpace(line) + "\n")
			}
			if i < len(paragraphs)-1 {
				formattedDesc.WriteString("\n")
			}
		}
		formattedDesc.WriteString("\n")

		fmt.Println(charm.RenderBox("About", formattedDesc.String(), "#FF6B9D", false))
	} else if cmd.Short != "" {
		fmt.Println(charm.RenderBox("", "\n  "+cmd.Short+"\n", "#FF6B9D", false))
	}

	// Usage section
	if cmd.HasAvailableSubCommands() {
		usageContent := fmt.Sprintf("\n  %s [command] [flags]\n", cmd.Name())
		fmt.Println(charm.RenderBox("Usage", usageContent, "#00D9FF", false))
	} else {
		usageContent := fmt.Sprintf("\n  %s\n", cmd.UseLine())
		fmt.Println(charm.RenderBox("Usage", usageContent, "#00D9FF", false))
	}

	// Available Commands
	if cmd.HasAvailableSubCommands() {
		var commandsContent strings.Builder
		commandsContent.WriteString("\n")

		for _, subCmd := range cmd.Commands() {
			if !subCmd.Hidden {
				commandsContent.WriteString(fmt.Sprintf("  %-12s %s\n", subCmd.Name(), subCmd.Short))
			}
		}
		commandsContent.WriteString("\n")

		fmt.Println(charm.RenderBox("Available Commands", commandsContent.String(), "#00FF87", false))
	}

	// Flags
	if cmd.HasAvailableFlags() {
		var flagsContent strings.Builder
		flagsContent.WriteString("\n")
		flagsContent.WriteString(cmd.Flags().FlagUsages())

		fmt.Println(charm.RenderBox("Flags", flagsContent.String(), "#FFD700", false))
	}

	// Footer
	fmt.Println()
	if cmd.HasAvailableSubCommands() {
		fmt.Println("  💡 Use 'anvil [command] --help' for more information about a command")
	}
	fmt.Println()
}
