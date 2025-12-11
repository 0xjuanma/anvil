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

package installer

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/0xjuanma/anvil/internal/terminal/charm"
)

// isShellCommand checks if the source is a shell command rather than a URL
func isShellCommand(source string) bool {
	trimmed := strings.TrimSpace(source)
	// Check for common shell command patterns
	return strings.HasPrefix(trimmed, "sh -c") ||
		strings.HasPrefix(trimmed, "bash -c") ||
		strings.HasPrefix(trimmed, "curl") ||
		strings.HasPrefix(trimmed, "wget") ||
		strings.Contains(trimmed, "$(curl") ||
		strings.Contains(trimmed, "$(wget")
}

// installFromCommand executes a shell command to install an application
func installFromCommand(appName, command string) error {
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Installing %s from command", appName))
	spinner.Start()

	cmd, err := parseShellCommand(command)
	if err != nil {
		spinner.Error(fmt.Sprintf("Invalid command for %s", appName))
		return fmt.Errorf("invalid command: %w", err)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		spinner.Error(fmt.Sprintf("Failed to install %s", appName))
		return fmt.Errorf("command execution failed: %w", err)
	}

	spinner.Success(fmt.Sprintf("%s installed successfully", appName))
	return nil
}

// parseShellCommand parses a shell command string into an exec.Cmd
func parseShellCommand(command string) (*exec.Cmd, error) {
	trimmed := strings.TrimSpace(command)

	// Handle sh -c or bash -c commands
	if strings.HasPrefix(trimmed, "sh -c") || strings.HasPrefix(trimmed, "bash -c") {
		shell := "sh"
		if strings.HasPrefix(trimmed, "bash") {
			shell = "bash"
		}
		cmdStr := extractCommandFromShC(trimmed)
		return exec.Command(shell, "-c", cmdStr), nil
	}

	// Direct command execution
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}

	return exec.Command(parts[0], parts[1:]...), nil
}

// extractCommandFromShC extracts the command string from "sh -c 'command'" format
func extractCommandFromShC(fullCommand string) string {
	// Find the position after "sh -c" or "bash -c"
	prefix := "sh -c"
	if strings.HasPrefix(fullCommand, "bash") {
		prefix = "bash -c"
	}

	// Find the command part (after the shell and -c)
	startIdx := len(prefix)
	for startIdx < len(fullCommand) && (fullCommand[startIdx] == ' ' || fullCommand[startIdx] == '\'') {
		startIdx++
	}

	// Extract the command, handling quotes
	command := fullCommand[startIdx:]
	command = strings.TrimSpace(command)

	// Remove surrounding quotes if present
	if len(command) >= 2 {
		if (command[0] == '\'' && command[len(command)-1] == '\'') ||
			(command[0] == '"' && command[len(command)-1] == '"') {
			command = command[1 : len(command)-1]
		}
	}

	return command
}
