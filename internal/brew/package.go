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

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
)

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

// InstalledPackages returns a list of installed packages (leaves only)
func InstalledPackages() ([]BrewPackage, error) {
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
		line = strings.TrimSpace(line)
		// Filter out empty lines and potential noise/status messages (e.g. "✔︎ JSON API...")
		if line != "" && !strings.Contains(line, "JSON API") && !strings.Contains(line, "✔") {
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

// PackageInfo gets information about a package
func PackageInfo(packageName string) (*BrewPackage, error) {
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
