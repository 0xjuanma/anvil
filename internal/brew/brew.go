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

// Package brew provides Homebrew package management and installation capabilities.
// It handles package installation, availability checking, and Homebrew setup
// for both macOS and Linux systems.
package brew

import (
	"fmt"
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

// IsBrewInstalledAtPath checks if Homebrew is installed at known platform-specific paths.
// Checks common installation locations before falling back to PATH lookup.
func IsBrewInstalledAtPath() bool {
	var brewPaths []string

	// Platform-specific Homebrew paths
	if system.IsMacOS() {
		brewPaths = []string{
			constants.BrewPathAppleSilicon, // Apple Silicon
			constants.BrewPathIntel,        // Intel
		}
	} else {
		// Linux Homebrew paths
		brewPaths = []string{
			constants.BrewPathLinuxStandard, // Standard Linux Homebrew
			constants.BrewPathLinuxUser,     // User-local Linux Homebrew
			constants.BrewPathLinuxAlt,      // Alternative Linux path
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
	time.Sleep(constants.SpinnerDelay)
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

