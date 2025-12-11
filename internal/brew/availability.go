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

package brew

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/system"
)

// IsApplicationAvailable checks if an application is available on the system
// Optimized approach: Fastest operations first, slowest operations last
func IsApplicationAvailable(packageName string) bool {
	// Step 1: For known casks, check if app exists in /Applications (fastest - no system calls) - macOS only
	if system.IsMacOS() && isKnownCask(packageName) {
		if checkKnownCaskInApplications(packageName) {
			return true
		}
	}

	// Step 2: For known formulas, check PATH (fast - single system call)
	if isKnownFormula(packageName) {
		result, err := system.RunCommand("which", packageName)
		if err == nil && result.Success {
			return true
		}
		// Skip spotlight search for known formulas - they should only be in PATH
		return false
	}

	// Step 3: For unknown packages, check most likely /Applications path first (fast - single filesystem check) - macOS only
	if system.IsMacOS() && searchApplication(fmt.Sprintf("%s.app", packageName)) {
		return true
	}

	// Step 4: For unknown packages, check PATH (fast - single system call)
	result, err := system.RunCommand("which", packageName)
	if err == nil && result.Success {
		return true
	}

	// Step 5: Check if installed via Homebrew (slower - brew command)
	if IsPackageInstalled(packageName) {
		return true
	}

	// Step 6: Fallback - Spotlight search (slowest - system-wide search) - macOS only
	if system.IsMacOS() {
		return spotlightSearch(packageName)
	}

	return false
}

// checkKnownCaskInApplications checks if a known cask app exists in /Applications.
// Uses optimized app name generation for faster lookups.
func checkKnownCaskInApplications(packageName string) bool {
	// Use optimized app name generation for known casks
	appNames := generateOptimizedAppNames(packageName)
	for _, appName := range appNames {
		if searchApplication(appName) {
			return true
		}
	}
	return false
}

// searchApplication checks if an app exists in /Applications directory.
func searchApplication(appName string) bool {
	result, err := system.RunCommand("test", "-d", fmt.Sprintf("/Applications/%s", appName))
	if err == nil && result.Success {
		return true
	}

	return false
}

// generateOptimizedAppNames creates optimized app names for known packages.
// Uses special cases for common apps and fallback generation for others.
func generateOptimizedAppNames(packageName string) []string {
	// Use special cases first (most common)
	specialCases := map[string][]string{
		"visual-studio-code":    {"Visual Studio Code.app"},
		"google-chrome":         {"Google Chrome.app"},
		"1password":             {"1Password.app", "1Password 7 - Password Manager.app"},
		"iterm2":                {"iTerm.app"},
		"firefox":               {"Firefox.app"},
		"slack":                 {"Slack.app"},
		"docker-desktop":        {"Docker.app"},
		"postman":               {"Postman.app"},
		"vlc":                   {"VLC.app"},
		"spotify":               {"Spotify.app"},
		"discord":               {"Discord.app"},
		"zoom":                  {"zoom.us.app"},
		"notion":                {"Notion.app"},
		"cursor":                {"Cursor.app"},
		"raycast":               {"Raycast.app"},
		"alfred":                {"Alfred 5.app", "Alfred 4.app"},
		"obsidian":              {"Obsidian.app"},
		"rectangle":             {"Rectangle.app"},
		"brave-browser":         {"Brave Browser.app"},
		"microsoft-edge":        {"Microsoft Edge.app"},
		"arc":                   {"Arc.app"},
		"steam":                 {"Steam.app"},
		"telegram":              {"Telegram.app"},
		"signal":                {"Signal.app"},
		"whatsapp":              {"WhatsApp.app"},
		"obs":                   {"OBS.app"},
		"gimp":                  {"GIMP.app"},
		"inkscape":              {"Inkscape.app"},
		"mongodb-compass":       {"MongoDB Compass.app"},
		"dbeaver-community":     {"DBeaver.app"},
		"pgadmin4":              {"pgAdmin 4.app"},
		"db-browser-for-sqlite": {"DB Browser for SQLite.app"},
		"kitty":                 {"kitty.app"},
		"alacritty":             {"Alacritty.app"},
		"wezterm":               {"WezTerm.app"},
		"iina":                  {"IINA.app"},
		"stats":                 {"Stats.app"},
		"betterdisplay":         {"BetterDisplay.app"},
		"alt-tab":               {"AltTab.app"},
		"karabiner-elements":    {"Karabiner-Elements.app"},
		"bitwarden":             {"Bitwarden.app"},
		"claude":                {"Claude.app"},
		"utm":                   {"UTM.app"},
		"adobe-acrobat-reader":  {"Adobe Acrobat Reader DC.app"},
		"appcleaner":            {"AppCleaner.app"},
		"vscodium":              {"VSCodium.app"},
		"insomnia":              {"Insomnia.app"},
		"claude-code":           {"Claude Code.app"},
	}

	if special, exists := specialCases[packageName]; exists {
		return special
	}

	// Fallback to generic generation
	var names []string
	names = append(names, packageName+".app")

	// Handle hyphenated names
	if strings.Contains(packageName, "-") {
		spacedName := strings.ReplaceAll(packageName, "-", " ")
		// Simple title case: capitalize first letter of each word
		words := strings.Fields(spacedName)
		for i, word := range words {
			if len(word) > 0 {
				words[i] = strings.ToUpper(word[:1]) + strings.ToLower(word[1:])
			}
		}
		names = append(names, strings.Join(words, " ")+".app")
	}

	return names
}

// spotlightSearch uses macOS Spotlight to find applications system-wide.
// This is the slowest availability check and is used as a last resort.
func spotlightSearch(packageName string) bool {
	// Use mdfind to search for applications containing the package name
	query := fmt.Sprintf("kMDItemKind == 'Application' && kMDItemFSName == '*%s*'", packageName)
	result, err := system.RunCommand("mdfind", query)

	if err != nil {
		return false
	}

	// If mdfind returns any results, the app exists somewhere
	return strings.TrimSpace(result.Output) != ""
}
