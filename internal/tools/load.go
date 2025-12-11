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

package tools

import (
	"fmt"
	"sort"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/palantir"
)

// LoadAndPrepareAppData loads all application data and prepares it for rendering
// This function is copied from the install package to maintain consistency
func LoadAndPrepareAppData() (groups map[string][]string, builtInGroupNames []string, customGroupNames []string, installedApps []string, err error) {
	// Load groups from config
	groups, err = config.AvailableGroups()
	if err != nil {
		err = errors.NewConfigurationError(constants.OpShow, "load-data",
			fmt.Errorf("failed to load groups: %w", err))
		return
	}

	// Get built-in group names
	builtInGroupNames = config.BuiltInGroups()

	// Extract and sort custom group names
	for groupName := range groups {
		if !config.IsBuiltInGroup(groupName) {
			customGroupNames = append(customGroupNames, groupName)
		}
	}
	sort.Strings(customGroupNames)

	// Load and sort installed apps
	installedApps, err = config.InstalledApps()
	if err != nil {
		// Don't fail on installed apps error, just log warning
		palantir.GetGlobalOutputHandler().PrintWarning("Failed to load installed apps: %v", err)
		installedApps = []string{}
		err = nil // Reset error since we can continue
	} else {
		sort.Strings(installedApps)
	}

	return
}
