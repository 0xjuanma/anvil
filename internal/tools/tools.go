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

// Package tools provides tool validation and management for the Anvil CLI.
// It handles required tool checking, installation validation, and tool status reporting.
package tools

import (
	"fmt"

	"github.com/0xjuanma/anvil/internal/brew"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/palantir"
)

// Tool represents a macOS system tool
type Tool struct {
	Name        string
	Command     string
	Required    bool
	InstallWith string
	Description string
}

// RequiredTools returns the list of required tools for anvil on macOS
// Note: Homebrew is handled separately as a prerequisite in ValidateAndInstallTools()
func RequiredTools() []Tool {
	return []Tool{
		{
			Name:        "Git",
			Command:     constants.GitCommand,
			Required:    true,
			InstallWith: "brew",
			Description: "Version control system",
		},
		{
			Name:        "cURL",
			Command:     constants.CurlCommand,
			Required:    true,
			InstallWith: "system",
			Description: "Command line tool for transferring data",
		},
	}
}

// ValidateAndInstallTools validates and installs required tools on macOS
func ValidateAndInstallTools() error {

	// Phase 1: Install Homebrew as a prerequisite (required for other tool installations)
	if err := brew.EnsureBrewIsInstalled(); err != nil {
		return fmt.Errorf("tools: %w", err)
	}

	// Phase 2: Validate and install other required tools (using Homebrew when needed)
	requiredTools := RequiredTools()
	for _, tool := range requiredTools {
		if err := validateTool(tool); err != nil {
			return fmt.Errorf("failed to validate required tool %s: %w", tool.Name, err)
		}
	}

	return nil
}

// validateTool validates a single tool on macOS
func validateTool(tool Tool) error {
	o := palantir.GetGlobalOutputHandler()
	if system.CommandExists(tool.Command) {
		o.PrintInfo("✓ %s is available", tool.Name)
		return nil
	}

	if !tool.Required {
		o.PrintWarning("○ %s is not installed (optional)", tool.Name)
		return nil
	}

	// Try to install the tool
	o.PrintInfo("Installing %s...", tool.Name)

	switch tool.InstallWith {
	case "brew":
		if err := brew.InstallPackage(tool.Command); err != nil {
			return fmt.Errorf("failed to install %s with brew: %w", tool.Name, err)
		}
	case "system":
		// cURL should be available
		return fmt.Errorf("%s is not available on this system", tool.Name)
	default:
		return fmt.Errorf("unknown installation method for %s", tool.Name)
	}

	// Verify installation
	if !system.CommandExists(tool.Command) {
		return fmt.Errorf("%s was not successfully installed", tool.Name)
	}

	o.PrintSuccess(fmt.Sprintf("%s installed successfully", tool.Name))
	return nil
}

// ToolInfo returns information about a specific tool
func ToolInfo(toolName string) (*Tool, error) {
	allTools := RequiredTools()

	for _, tool := range allTools {
		if tool.Name == toolName || tool.Command == toolName {
			return &tool, nil
		}
	}

	return nil, fmt.Errorf("tool %s not found", toolName)
}

// CheckToolsStatus checks the status of all tools on macOS
func CheckToolsStatus() (map[string]bool, error) {
	if !system.IsMacOS() {
		// TODO: Implement tool status check for Linux
		return nil, fmt.Errorf("tool status check only supported on macOS")
	}

	status := make(map[string]bool)

	allTools := RequiredTools()
	for _, tool := range allTools {
		status[tool.Name] = system.CommandExists(tool.Command)
	}

	return status, nil
}
