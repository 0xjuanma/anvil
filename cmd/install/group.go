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

package install

import (
	"context"
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/installer"
	"github.com/0xjuanma/palantir"
)

// installGroup installs all tools in a group.
func installGroup(opts InstallGroupOptions) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Installing '%s' group", opts.GroupName))

	if len(opts.Tools) == 0 {
		return errors.NewInstallationError(constants.OpInstall, opts.GroupName,
			fmt.Errorf("group '%s' has no tools defined", opts.GroupName))
	}

	// Deduplicate tools within the group and update settings if needed
	deduplicatedTools, err := deduplicateGroupTools(opts.GroupName, opts.Tools)
	if err != nil {
		o.PrintWarning("Failed to deduplicate group tools: %v", err)
	} else {
		opts.Tools = deduplicatedTools
	}

	o.PrintInfo("Installing %d tools: %s", len(opts.Tools), strings.Join(opts.Tools, ", "))

	if opts.Concurrent {
		return installGroupConcurrent(opts)
	}

	return installGroupSerial(opts.GroupName, opts.Tools, opts.DryRun)
}

// deduplicateGroupTools removes duplicate tools within a group and updates the settings file.
func deduplicateGroupTools(groupName string, tools []string) ([]string, error) {
	seen := make(map[string]struct{}, len(tools))
	deduplicatedTools := make([]string, 0, len(tools))
	var duplicatesFound []string

	// Deduplicate
	for _, tool := range tools {
		if _, exists := seen[tool]; !exists {
			seen[tool] = struct{}{}
			deduplicatedTools = append(deduplicatedTools, tool)
		} else {
			duplicatesFound = append(duplicatesFound, tool)
		}
	}

	// Return original list if no duplicates found
	if len(duplicatesFound) == 0 {
		return tools, nil
	}

	o := palantir.GetGlobalOutputHandler()
	o.PrintWarning("Found duplicates in group '%s': %s", groupName, strings.Join(duplicatesFound, ", "))
	o.PrintInfo("Removing duplicates from settings file...")

	// Update the configuration with deduplicated tools
	if err := config.UpdateGroupTools(groupName, deduplicatedTools); err != nil {
		return tools, fmt.Errorf("failed to update group with deduplicated tools: %w", err)
	}

	o.PrintSuccess(fmt.Sprintf("Successfully removed %d duplicate(s) from group '%s'", len(duplicatesFound), groupName))
	return deduplicatedTools, nil
}

// installGroupConcurrent installs tools concurrently.
func installGroupConcurrent(opts InstallGroupOptions) error {
	o := palantir.GetGlobalOutputHandler()

	// Create new output handler to send into concurrent installer
	outputHandler := palantir.NewDefaultOutputHandler()
	concurrentInstaller := installer.NewConcurrentInstaller(opts.MaxWorkers, outputHandler, opts.DryRun)

	if opts.Timeout > 0 {
		concurrentInstaller.SetTimeout(opts.Timeout)
	}

	// Create context with potential cancellation
	ctx := context.Background()
	stats, err := concurrentInstaller.InstallTools(ctx, opts.Tools)

	// Track successfully installed apps
	if !opts.DryRun && stats != nil && stats.SuccessfulTools > 0 {
		o.PrintInfo("Updating settings to track installed apps...")
		o.PrintInfo("Group installation tracking not implemented yet")
	}

	return err
}

// installGroupSerial installs tools serially using unified installation logic.
func installGroupSerial(groupName string, tools []string, dryRun bool) error {
	o := palantir.GetGlobalOutputHandler()

	successCount := 0
	var installErrors []string

	// Initialize tool statuses
	toolStatuses := make([]toolStatus, len(tools))
	for i, tool := range tools {
		toolStatuses[i] = toolStatus{
			name:   tool,
			status: toolStatusPending,
			emoji:  "⋯",
		}
	}

	for i, tool := range tools {
		// Update status to installing
		toolStatuses[i].status = toolStatusInstalling
		toolStatuses[i].emoji = "⠋"

		// Print dashboard
		printInstallDashboard(groupName, toolStatuses, i+1, len(tools))

		// Use unified installation logic
		_, err := installSingleToolUnified(tool, dryRun)

		if err != nil {
			toolStatuses[i].status = toolStatusFailed
			toolStatuses[i].emoji = "✗"
			errorMsg := fmt.Sprintf("%s: %v", tool, err)
			installErrors = append(installErrors, errorMsg)
			o.PrintError("%s: %v", tool, err)
		} else {
			toolStatuses[i].status = toolStatusDone
			toolStatuses[i].emoji = "✓"
			successCount++
		}

		// Print final dashboard state
		printInstallDashboard(groupName, toolStatuses, i+1, len(tools))
	}

	return reportGroupInstallationResults(groupName, successCount, len(tools), installErrors)
}
