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
	"github.com/0xjuanma/palantir"
)

var defaultAppSet = map[string]struct{}{
	"calculator":         {},
	"calendar":           {},
	"chess":              {},
	"contacts":           {},
	"dictionary":         {},
	"facetime":           {},
	"finder":             {},
	"font-book":          {},
	"image-capture":      {},
	"keychain-access":    {},
	"mail":               {},
	"maps":               {},
	"messages":           {},
	"music":              {},
	"news":               {},
	"notes":              {},
	"photo-booth":        {},
	"photos":             {},
	"preview":            {},
	"quicktime-player":   {},
	"reminders":          {},
	"safari":             {},
	"stickies":           {},
	"system-preferences": {},
	"system-settings":    {},
	"textedit":           {},
	"time-machine":       {},
	"tv":                 {},
}

var appAliases = map[string]string{
	"iTerm":                   "iterm2",
	"Zoom.us":                 "zoom",
	"1Password 7":             "1password",
	"Alfred 5":                "alfred",
	"Alfred 4":                "alfred",
	"pgAdmin 4":               "pgadmin4",
	"DBeaver":                 "dbeaver-community",
	"AltTab":                  "alt-tab",
	"Adobe Acrobat Reader DC": "adobe-acrobat-reader",
	"Parallels Desktop":       "parallels",
	"CleanMyMac X":            "cleanmymac",
	"Bartender 5":             "bartender",
	"Bartender 4":             "bartender",
	"Logi Options+":           "logi-options-plus",
	"Hands Off!":              "hands-off",
	"Box":                     "box-drive",
	"pCloud":                  "pcloud-drive",
	"SuperDuper!":             "superduper",
	"VLC Media Player":        "vlc",
	"Epic Games Launcher":     "epic-games",
}

// RunDiscoverLogic discovers apps and tools installed on the system and adds them to the "discovered-apps" group if not tracked
func RunDiscoverLogic() error {
	// 1. Use Homebrew to discover tools(using --formulae flag)
	homebrewTools, err := discoverHomebrewTools()
	if err != nil {
		// Log error but continue with macOS app discovery
		palantir.GetGlobalOutputHandler().PrintWarning("Failed to discover Homebrew tools: %v", err)
	}

	// 2. Use Applications folder to discover apps
	macOSApps := []string{}
	if system.IsMacOS() {
		macOSApps, err = discoverMacOSApps()
		if err != nil {
			return err
		}
	}

	// 3. Filter tracked apps
	var appsToAdd []string
	for _, app := range append(homebrewTools, macOSApps...) {
		tracked, err := IsAppTracked(app)
		if err != nil || tracked {
			continue
		}
		appsToAdd = append(appsToAdd, app)
	}

	// 4. Add all discovered apps to the "discovered-apps" group in bulk
	if len(appsToAdd) > 0 {
		if err := AddAppsToGroup("discovered-apps", appsToAdd); err != nil {
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

// discoverMacOSApps discovers apps in the /Applications folder
func discoverMacOSApps() ([]string, error) {
	apps := []string{}

	entries, err := os.ReadDir("/Applications")
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
			continue
		}

		packageName := convertAppNameToPackage(entry.Name())
		if _, exists := defaultAppSet[packageName]; exists {
			continue
		}

		apps = append(apps, packageName)
	}

	return apps, nil
}

// convertAppNameToPackage converts a macOS .app name to a package name
func convertAppNameToPackage(name string) string {
	cleanName := strings.TrimSuffix(name, ".app")

	// Check explicit aliases first
	if pkg, ok := appAliases[cleanName]; ok {
		return pkg
	}

	// Fallback to standard/basic normalization
	name = strings.ToLower(cleanName)
	return strings.ReplaceAll(name, " ", "-")
}
