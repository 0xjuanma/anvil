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

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
)

// trackAppInSettings handles adding newly installed apps to settings.
func trackAppInSettings(appName string) error {
	o := palantir.GetGlobalOutputHandler()
	// Check if already tracked to avoid duplicates
	if isTracked, err := config.IsAppTracked(appName); err != nil {
		o.PrintWarning("Failed to check if %s is already tracked: %v", appName, err)
		return nil // Don't fail installation for tracking issues
	} else if isTracked {
		return nil
	}

	spinner := charm.NewDotsSpinner(fmt.Sprintf("Tracking %s in settings", appName))
	spinner.Start()

	if err := config.AddInstalledApp(appName); err != nil {
		spinner.Warning("Failed to update settings")
		o.PrintWarning("Failed to update settings file: %v", err)
		return nil // Don't fail installation for tracking issues
	}

	spinner.Success(fmt.Sprintf("%s tracked in settings", appName))
	return nil
}

// reportGroupInstallationResults provides unified error reporting for group installations.
func reportGroupInstallationResults(groupName string, successCount, totalCount int, installErrors []string) error {
	// Print summary
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Group Installation Complete")
	o.PrintInfo("Successfully installed %d of %d tools", successCount, totalCount)

	if len(installErrors) > 0 {
		o.PrintWarning("Some installations failed:")
		for _, err := range installErrors {
			o.PrintError("  • %s", err)
		}
		return errors.NewInstallationError(constants.OpInstall, groupName,
			fmt.Errorf("failed to install %d tools", len(installErrors)))
	}

	return nil
}
