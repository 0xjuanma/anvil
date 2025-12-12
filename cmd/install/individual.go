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
	"fmt"

	"github.com/0xjuanma/anvil/internal/brew"
	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/installer"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

// installIndividualApp installs a single application using unified installation logic.
func installIndividualApp(appName string, dryRun bool, cmd *cobra.Command) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Installing '%s'", appName))

	// Validate app name is not empty
	if appName == "" {
		return errors.NewInstallationError(constants.OpInstall, appName,
			fmt.Errorf("application name cannot be empty"))
	}

	wasNewlyInstalled, err := installSingleToolUnified(appName, dryRun)
	if err != nil {
		return errors.NewInstallationError(constants.OpInstall, appName,
			fmt.Errorf("failed to install '%s'. Please verify the name is correct. You can search for packages using 'brew search %s'", appName, appName))
	}

	// Only track the app in settings if it was newly installed and not dry-run
	if !dryRun && wasNewlyInstalled {
		// Check if --group-name flag is provided
		groupName, _ := cmd.Flags().GetString("group-name")
		if groupName != "" {
			// Add app to the specified group
			if err := config.AddAppToGroup(groupName, appName); err != nil {
				o.PrintWarning("Failed to add %s to group '%s': %v", appName, groupName, err)
				// Continue with normal tracking as fallback
				return trackAppInSettings(appName)
			}
			o.PrintSuccess(fmt.Sprintf("Added %s to group '%s'", appName, groupName))
			return nil
		} else {
			// Normal tracking in installed_apps
			return trackAppInSettings(appName)
		}
	}

	return nil
}

// installSingleTool installs a single tool, handling special cases dynamically.
func installSingleTool(toolName string) error {
	o := palantir.GetGlobalOutputHandler()

	// Check if source is configured for this app (user explicitly configured it)
	sourceURL, exists, sourceErr := installer.SourceURL(toolName)
	if sourceErr != nil {
		o.PrintWarning("Failed to check source URL for %s: %v", toolName, sourceErr)
		// Fall back to brew if we can't check source
		return brew.InstallPackageDirectly(toolName)
	}

	// If source exists, try it first (user explicitly configured it)
	if exists && sourceURL != "" {
		o.PrintInfo("Installing %s from configured source", toolName)
		if err := installer.InstallFromSource(toolName, sourceURL); err != nil {
			// Check if extraction succeeded but installation failed
			if _, ok := err.(*installer.ExtractionSucceededError); ok {
				// Extraction succeeded, don't fall back to brew
				// User message already shown in InstallFromSource
				return nil
			}
			// Source installation failed, fall back to brew
			o.PrintInfo("Source installation failed, falling back to brew for %s", toolName)
			return brew.InstallPackageDirectly(toolName)
		}
		// Source installation succeeded, continue with post-install steps
	} else {
		// No source configured, use brew (default for majority of apps)
		if err := brew.InstallPackageDirectly(toolName); err != nil {
			return err
		}
	}

	// Handle config check for git
	if toolName == "git" {
		if err := checkToolConfiguration(toolName); err != nil {
			o.PrintWarning("Configuration check failed for %s: %v", toolName, err)
		}
	}

	return nil
}

// installSingleToolUnified provides unified installation logic for all installation modes.
func installSingleToolUnified(toolName string, dryRun bool) (wasNewlyInstalled bool, err error) {
	o := palantir.GetGlobalOutputHandler()

	// ALWAYS check availability first using the latest IsApplicationAvailable logic
	if brew.IsApplicationAvailable(toolName) {
		o.PrintAlreadyAvailable("%s is already available on the system", toolName)
		return false, nil
	}

	// Handle installation based on mode
	if dryRun {
		o.PrintInfo("Would install: %s", toolName)
		return true, nil
	}

	// Perform real installation using existing logic
	if err := installSingleTool(toolName); err != nil {
		return false, err
	}

	o.PrintSuccess(fmt.Sprintf("%s installed successfully", toolName))
	return true, nil
}
