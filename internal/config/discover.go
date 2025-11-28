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

package config

import (
	"os"
	"strings"

	"github.com/0xjuanma/anvil/internal/brew"
	"github.com/0xjuanma/anvil/internal/system"
)

// RunDiscoverLogic discovers apps and tools installed on the system and adds them to the "discovered-apps" group if not tracked
func RunDiscoverLogic() error {
	// 1. Use Homebrew to discover tools(using --formulae flag)
	homebrewTools, err := discoverHomebrewTools()
	if err != nil {
		return err
	}

	// 2. Use Applications folder to discover apps
	macOSApps := []string{}
	if system.IsMacOS() {
		macOSApps, err = macOSAppDiscovery()
		if err != nil {
			return err
		}
	}

	// 3. Iterate over all discovered apps and add them to the "discovered-apps" group if not tracked
	for _, app := range append(homebrewTools, macOSApps...) {
		tracked, err := IsAppTracked(app)
		if err != nil || tracked {
			continue
		}

		if err := AddAppToGroup("discovered-apps", app); err != nil {
			return err
		}
	}

	return nil
}

// discoverHomebrewTools discovers tools installed via Homebrew using the --formula flag
func discoverHomebrewTools() ([]string, error) {
	tools := []string{}
	homebrewTools, err := brew.GetInstalledPackages()
	if err != nil {
		return nil, err
	}

	for _, tool := range homebrewTools {
		tools = append(tools, tool.Name)
	}

	return tools, nil
}

// macOSAppDiscovery discovers apps in the /Applications folder
func macOSAppDiscovery() ([]string, error) {
	apps := []string{}

	defaultApps := []string{
		"calculator", "calendar", "chess", "contacts",
		"dictionary", "facetime", "finder", "font-book",
		"image-capture", "keychain-access", "mail", "maps",
		"messages", "music", "news", "notes",
		"photo-booth", "photos", "preview", "quicktime-player",
		"reminders", "safari", "stickies", "system-preferences",
		"system-settings", "textedit", "time-machine", "tv",
		"utilities",
	}
	defaultAppSet := make(map[string]struct{}, len(defaultApps))
	for _, app := range defaultApps {
		defaultAppSet[app] = struct{}{}
	}

	entries, err := os.ReadDir("/Applications")
	if err != nil {
		return nil, err
	}

	appNameToPackage := func(name string) string {
		name = strings.TrimSuffix(name, ".app")
		name = strings.ToLower(name)
		return strings.ReplaceAll(name, " ", "-")
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
			continue
		}

		packageName := appNameToPackage(entry.Name())
		if _, exists := defaultAppSet[packageName]; exists {
			continue
		}

		apps = append(apps, packageName)
	}

	return apps, nil
}
