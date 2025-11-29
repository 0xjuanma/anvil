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
	"sync"
	"time"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
)

var (
	// Cache brew installation status to avoid repeated checks
	brewInstalledCache *bool
	brewCacheMutex     sync.RWMutex
)

// BrewPackage represents a brew package
type BrewPackage struct {
	Name        string
	Version     string
	Description string
	Installed   bool
}

// EnsureBrewIsInstalled ensures Homebrew is installed
func EnsureBrewIsInstalled() error {
	if !IsBrewInstalled() {
		palantir.GetGlobalOutputHandler().PrintInfo("Homebrew not found. Installing Homebrew...")

		var err error
		if system.IsMacOS() {
			err = InstallBrew()
		} else {
			err = InstallBrewLinux()
		}

		if err != nil {
			return fmt.Errorf("failed to install Homebrew: %w", err)
		}
		palantir.GetGlobalOutputHandler().PrintSuccess("Homebrew installed successfully")
	}

	return nil
}

// IsBrewInstalled checks if Homebrew is installed (with caching)
func IsBrewInstalled() bool {
	// Check cache first
	brewCacheMutex.RLock()
	if brewInstalledCache != nil {
		result := *brewInstalledCache
		brewCacheMutex.RUnlock()
		return result
	}
	brewCacheMutex.RUnlock()

	// Not in cache, check and cache the result
	brewCacheMutex.Lock()
	defer brewCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if brewInstalledCache != nil {
		return *brewInstalledCache
	}

	result := system.CommandExists(constants.BrewCommand)
	brewInstalledCache = &result
	return result
}

// IsBrewInstalledAtPath checks if Homebrew is installed at known paths
func IsBrewInstalledAtPath() bool {
	var brewPaths []string

	// Platform-specific Homebrew paths
	if system.IsMacOS() {
		brewPaths = []string{
			"/opt/homebrew/bin/brew", // Apple Silicon
			"/usr/local/bin/brew",    // Intel
		}
	} else {
		// Linux Homebrew paths
		brewPaths = []string{
			"/home/linuxbrew/.linuxbrew/bin/brew", // Standard Linux Homebrew
			"~/.linuxbrew/bin/brew",               // User-local Linux Homebrew
			"/opt/homebrew/bin/brew",              // Alternative Linux path
		}
	}

	for _, path := range brewPaths {
		result, err := system.RunCommand("test", "-x", path)
		if err == nil && result.Success {
			return true
		}
	}

	return system.CommandExists("brew")
}

// InstallBrew installs Homebrew if not already installed
func InstallBrew() error {
	if IsBrewInstalled() {
		return nil
	}

	// Check for Xcode Command Line Tools on macOS only
	if system.IsMacOS() {
		spinner := charm.NewDotsSpinner("Checking Xcode Command Line Tools")
		spinner.Start()

		xcodeResult, err := system.RunCommand("xcode-select", "-p")
		if err != nil || !xcodeResult.Success {
			spinner.Error("Xcode Command Line Tools not found")
			return fmt.Errorf("Xcode Command Line Tools required for Homebrew installation. Install with: xcode-select --install")
		}
		spinner.Success("Xcode Command Line Tools verified")
	}

	palantir.GetGlobalOutputHandler().PrintInfo("Installing Homebrew (this may take a few minutes)")
	palantir.GetGlobalOutputHandler().PrintInfo("You may be prompted for your password to complete the installation")
	fmt.Println()

	spinner := charm.NewDotsSpinner("Preparing Homebrew installation")
	spinner.Start()
	time.Sleep(200 * time.Millisecond)
	spinner.Stop()

	fmt.Print("\r\033[K→ Enter password when prompted: ")

	installScript := `echo | /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

	spinner = charm.NewDotsSpinner("Installing")
	spinner.Start()
	err := system.RunInteractiveCommand("/bin/bash", "-c", installScript)
	spinner.Stop()
	fmt.Println()

	if err != nil {
		palantir.GetGlobalOutputHandler().PrintError("Homebrew installation failed")
		return fmt.Errorf("failed to install Homebrew: %w", err)
	}

	spinner = charm.NewDotsSpinner("Verifying Homebrew installation")
	spinner.Start()

	if !IsBrewInstalledAtPath() {
		spinner.Error("Homebrew installation verification failed")
		return fmt.Errorf("Homebrew installation completed but brew command not accessible")
	}

	spinner.Success("Homebrew installed successfully")
	return nil
}

// UpdateBrew updates Homebrew and its formulae
func UpdateBrew() error {
	if !IsBrewInstalled() {
		return fmt.Errorf("Homebrew is not installed")
	}

	spinner := charm.NewDotsSpinner("Updating Homebrew")
	spinner.Start()

	result, err := system.RunCommand(constants.BrewCommand, constants.BrewUpdate)
	if err != nil {
		spinner.Error("Failed to update Homebrew")
		return fmt.Errorf("failed to run brew update: %w", err)
	}

	if !result.Success {
		spinner.Error("Homebrew update failed")
		return fmt.Errorf("brew update failed: %s", result.Error)
	}

	spinner.Success("Homebrew updated successfully")
	return nil
}

// InstallPackage installs a package using Homebrew
func InstallPackage(packageName string) error {
	if !IsBrewInstalled() {
		return fmt.Errorf("Homebrew is not installed")
	}

	palantir.GetGlobalOutputHandler().PrintInfo("Installing %s...", packageName)

	result, err := system.RunCommand(constants.BrewCommand, constants.BrewInstall, packageName)
	if err != nil {
		return fmt.Errorf("failed to run brew install: %w", err)
	}

	if !result.Success {
		// Include actual brew output for better diagnostics
		var errorDetails string
		if result.Output != "" {
			errorDetails = fmt.Sprintf("brew output: %s", strings.TrimSpace(result.Output))
		} else {
			errorDetails = fmt.Sprintf("system error: %s", result.Error)
		}
		return fmt.Errorf("failed to install %s: %s", packageName, errorDetails)
	}

	return nil
}

// IsPackageInstalled checks if a package is installed (both formulas and casks)
func IsPackageInstalled(packageName string) bool {
	if !IsBrewInstalled() {
		return false
	}

	// Use single brew list command to check both formulas and casks
	result, err := system.RunCommand(constants.BrewCommand, constants.BrewList, packageName)
	if err == nil && result.Success {
		return true
	}

	return false
}

// GetInstalledPackages returns a list of installed packages (leaves only)
func GetInstalledPackages() ([]BrewPackage, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("Homebrew is not installed")
	}

	// Use 'brew leaves' to get only top-level packages (not dependencies)
	// Adding '--installed-on-request' to filter out dependencies that might have become leaves
	result, err := system.RunCommand(constants.BrewCommand, "leaves", "--installed-on-request")
	if err != nil {
		return nil, fmt.Errorf("failed to run brew leaves: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get installed packages: %s", result.Error)
	}

	var packages []BrewPackage
	lines := strings.Split(strings.TrimSpace(result.Output), "\n")

	for _, line := range lines {
		if line != "" {
			packages = append(packages, BrewPackage{
				Name:      line,
				Installed: true,
			})
		}
	}

	return packages, nil
}

// InstallPackages installs multiple packages
func InstallPackages(packages []string) error {
	if !IsBrewInstalled() {
		return fmt.Errorf("Homebrew is not installed")
	}

	for i, pkg := range packages {
		palantir.GetGlobalOutputHandler().PrintProgress(i+1, len(packages), fmt.Sprintf("Installing %s", pkg))

		if IsPackageInstalled(pkg) {
			palantir.GetGlobalOutputHandler().PrintInfo("%s is already installed", pkg)
			continue
		}

		if err := InstallPackageWithCheck(pkg); err != nil {
			return fmt.Errorf("failed to install %s: %w", pkg, err)
		}
	}

	return nil
}

// GetPackageInfo gets information about a package
func GetPackageInfo(packageName string) (*BrewPackage, error) {
	if !IsBrewInstalled() {
		return nil, fmt.Errorf("Homebrew is not installed")
	}

	result, err := system.RunCommand(constants.BrewCommand, constants.BrewInfo, packageName)
	if err != nil {
		return nil, fmt.Errorf("failed to run brew info: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("failed to get info for %s: %s", packageName, result.Error)
	}

	// Parse the output to extract package information
	lines := strings.Split(result.Output, "\n")
	if len(lines) == 0 {
		return nil, fmt.Errorf("no information found for %s", packageName)
	}

	pkg := &BrewPackage{
		Name:      packageName,
		Installed: IsPackageInstalled(packageName),
	}

	// Extract version and description from the first line
	firstLine := lines[0]
	if strings.Contains(firstLine, ":") {
		parts := strings.Split(firstLine, ":")
		if len(parts) > 1 {
			pkg.Description = strings.TrimSpace(parts[1])
		}
	}

	return pkg, nil
}

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

// checkKnownCaskInApplications checks if a known cask app exists in /Applications
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

// searchApplication checks if an app exists in /Applications
func searchApplication(appName string) bool {
	result, err := system.RunCommand("test", "-d", fmt.Sprintf("/Applications/%s", appName))
	if err == nil && result.Success {
		return true
	}

	return false
}

// isKnownCask checks if a package is a known cask from our lookup table
func isKnownCask(packageName string) bool {
	if isCask, exists := knownBrewPackages[packageName]; exists {
		return isCask
	}
	return false
}

// isKnownFormula checks if a package is a known formula from our lookup table
func isKnownFormula(packageName string) bool {
	if isCask, exists := knownBrewPackages[packageName]; exists {
		return !isCask // If it exists and is not a cask, it's a formula
	}
	return false
}

// generateOptimizedAppNames creates optimized app names for known packages
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
		names = append(names, strings.Title(spacedName)+".app")
	}

	return names
}

// spotlightSearch uses macOS Spotlight to find applications system-wide
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

// isCaskPackage determines if a package is a Homebrew cask using optimized lookup
func isCaskPackage(packageName string) bool {
	// Step 1: Check static lookup table (fastest - covers 95% of common packages)
	if isCask, exists := knownBrewPackages[packageName]; exists {
		return isCask
	}

	// Step 2: Check runtime cache
	caskCacheMutex.RLock()
	if isCask, cached := caskCache[packageName]; cached {
		caskCacheMutex.RUnlock()
		return isCask
	}
	caskCacheMutex.RUnlock()

	// Step 3: Dynamic detection (expensive - only for unknown packages)
	isCask := detectCaskDynamically(packageName)

	// Cache the result
	caskCacheMutex.Lock()
	caskCache[packageName] = isCask
	caskCacheMutex.Unlock()

	return isCask
}

// detectCaskDynamically performs expensive brew search for unknown packages
func detectCaskDynamically(packageName string) bool {
	result, err := system.RunCommand(constants.BrewCommand, "search", "--cask", packageName)
	if err != nil {
		return false
	}

	lines := strings.Split(strings.TrimSpace(result.Output), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip headers, empty lines, and error messages
		if line == "" || strings.Contains(line, "==>") || strings.Contains(line, "Error:") || strings.Contains(line, "Warning:") {
			continue
		}
		// Only consider exact matches for casks to avoid false positives
		if line == packageName {
			return true
		}
	}

	// Default to false (formula) if not a cask
	return false
}

// InstallPackageWithCheck installs a package only if it's not already available
func InstallPackageWithCheck(packageName string) error {
	if !IsBrewInstalled() {
		return fmt.Errorf("Homebrew is not installed")
	}

	if IsApplicationAvailable(packageName) {
		palantir.GetGlobalOutputHandler().PrintAlreadyAvailable("%s is already available on the system", packageName)
		return nil
	}

	return InstallPackageDirectly(packageName)
}

// InstallPackageDirectly installs a package without checking availability first
// Used when availability has already been verified by the caller
func InstallPackageDirectly(packageName string) error {
	if !IsBrewInstalled() {
		return fmt.Errorf("Homebrew is not installed")
	}

	isCask := isCaskPackage(packageName)
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Installing %s", packageName))
	spinner.Start()

	var result *system.CommandResult
	var err error

	if isCask {
		result, err = system.RunCommand(constants.BrewCommand, constants.BrewInstall, "--cask", packageName)
	} else {
		result, err = system.RunCommand(constants.BrewCommand, constants.BrewInstall, packageName)
	}

	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to install %s", packageName))
		return fmt.Errorf("failed to run brew install: %w", err)
	}

	if !result.Success {
		if strings.Contains(result.Error, "already an App at") {
			spinner.Warning(fmt.Sprintf("%s already installed manually", packageName))
			return nil
		}

		spinner.Error(fmt.Sprintf("Failed to install %s", packageName))
		if result.Output != "" {
			return fmt.Errorf("brew: %s", strings.TrimSpace(result.Output))
		} else {
			return fmt.Errorf("installation failed: %s", result.Error)
		}
	}

	spinner.Success(fmt.Sprintf("%s installed successfully", packageName))
	return nil
}
