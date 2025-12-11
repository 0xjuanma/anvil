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

// Package system provides system command execution and OS detection utilities.
// It handles command execution with timeouts, environment setup, and platform detection.
package system

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
)

// CommandResult represents the result of a command execution
type CommandResult struct {
	Command  string
	ExitCode int
	Output   string
	Error    string
	Success  bool
}

// RunCommand executes a system command with a default timeout
func RunCommand(command string, args ...string) (*CommandResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultCommandTimeout)
	defer cancel()
	return RunCommandWithTimeout(ctx, command, args...)
}

// RunCommandWithTimeout executes a system command with the given context
func RunCommandWithTimeout(ctx context.Context, command string, args ...string) (*CommandResult, error) {
	cmd := exec.CommandContext(ctx, command, args...)

	// For git commands, ensure non-interactive mode to prevent credential prompts
	if command == constants.GitCommand {
		cmd.Env = append(os.Environ(), constants.GitNonInteractiveEnvVars()...)
	}

	// Capture both stdout and stderr
	output, err := cmd.CombinedOutput()

	result := &CommandResult{
		Command: strings.Join(append([]string{command}, args...), " "),
		Output:  string(output),
		Success: err == nil,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		result.Error = err.Error()
	}

	return result, nil
}

// CommandExists checks if a command exists in the system PATH
func CommandExists(command string) bool {
	_, err := exec.LookPath(command)
	return err == nil
}

// CommandPath returns the full path to a command if it exists
func CommandPath(command string) (string, error) {
	return exec.LookPath(command)
}

// RunCommandInDirectory executes a command in a specific directory
func RunCommandInDirectory(dir, command string, args ...string) (*CommandResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.DefaultCommandTimeout)
	defer cancel()
	return RunCommandInDirectoryWithTimeout(ctx, dir, command, args...)
}

// RunCommandInDirectoryWithTimeout executes a command in a specific directory with context
func RunCommandInDirectoryWithTimeout(ctx context.Context, dir, command string, args ...string) (*CommandResult, error) {
	cmd := exec.CommandContext(ctx, command, args...)
	cmd.Dir = dir

	// For git commands, ensure non-interactive mode to prevent credential prompts
	if command == constants.GitCommand {
		cmd.Env = append(os.Environ(), constants.GitNonInteractiveEnvVars()...)
	}

	output, err := cmd.CombinedOutput()

	result := &CommandResult{
		Command: fmt.Sprintf("cd %s && %s", dir, strings.Join(append([]string{command}, args...), " ")),
		Output:  string(output),
		Success: err == nil,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		}
		result.Error = err.Error()
	}

	return result, nil
}

// EnvironmentVariable gets an environment variable with a default value
func EnvironmentVariable(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetEnvironmentVariable sets an environment variable
func SetEnvironmentVariable(key, value string) error {
	return os.Setenv(key, value)
}

// RunInteractiveCommand executes a command with stdin connected for user interaction.
// Output is captured and only shown on error for clean UX.
// Sets INTERACTIVE=1 environment variable to force interactive mode even when TTY detection fails.
func RunInteractiveCommand(command string, args ...string) error {
	var outputBuffer bytes.Buffer

	cmd := exec.Command(command, args...)

	cmd.Stdin = os.Stdin

	// Capture stdout and stderr for clean output
	cmd.Stdout = &outputBuffer
	cmd.Stderr = &outputBuffer

	// Preserve existing environment and add INTERACTIVE=1
	// This forces tools like Homebrew installer to run in interactive mode
	// even when they can't detect a TTY through the subprocess chain
	cmd.Env = append(os.Environ(), "INTERACTIVE=1")

	err := cmd.Run()

	// If command failed, show the captured output for debugging
	if err != nil {
		output := strings.TrimSpace(outputBuffer.String())
		if output != "" {
			fmt.Println("\nCommand output:")
			fmt.Println(strings.Repeat("-", 80))
			fmt.Println(output)
			fmt.Println(strings.Repeat("-", 80))
		}
	}

	return err
}
