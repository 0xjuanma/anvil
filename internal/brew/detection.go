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
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
)

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
