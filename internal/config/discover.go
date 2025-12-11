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
	"garageband":         {},
	"image-capture":      {},
	"imovie":             {},
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
	"iterm":                   "iterm2",
	"zoom.us":                 "zoom",
	"1password 7":             "1password",
	"alfred 5":                "alfred",
	"alfred 4":                "alfred",
	"pgadmin 4":               "pgadmin4",
	"dbeaver":                 "dbeaver-community",
	"alttab":                  "alt-tab",
	"adobe acrobat reader dc": "adobe-acrobat-reader",
	"parallels desktop":       "parallels",
	"cleanmymac x":            "cleanmymac",
	"bartender 5":             "bartender",
	"bartender 4":             "bartender",
	"logi options+":           "logi-options-plus",
	"hands off!":              "hands-off",
	"box":                     "box-drive",
	"pcloud":                  "pcloud-drive",
	"superduper!":             "superduper",
	"vlc media player":        "vlc",
	"epic games launcher":     "epic-games",
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
	homebrewTools, err := brew.InstalledPackages()
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
		// Skip dotfiles, non-directories, and non-app bundles
		if strings.HasPrefix(entry.Name(), ".") || !entry.IsDir() || !strings.HasSuffix(entry.Name(), ".app") {
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
	lowerName := strings.ToLower(cleanName)

	// Check aliases first
	if pkg, ok := appAliases[lowerName]; ok {
		return pkg
	}

	// Fallback to standard/basic normalization
	return strings.ReplaceAll(lowerName, " ", "-")
}
