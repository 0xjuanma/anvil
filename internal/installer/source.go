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

package installer

import (
	"fmt"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
)

// InstallFromSource installs an application from a source URL or command
func InstallFromSource(appName, source string) error {
	// Check if source is a shell command (curl/wget style) or a URL
	if isShellCommand(source) {
		return installFromCommand(appName, source)
	}
	return installFromURL(appName, source)
}


// installFromURL installs an application from a URL
func installFromURL(appName, sourceURL string) error {
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Downloading %s from source", appName))
	spinner.Start()

	downloadedFile, err := downloadFile(sourceURL, appName)
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to download %s", appName))
		return fmt.Errorf("failed to download %s: %w", appName, err)
	}
	spinner.Success(fmt.Sprintf("Downloaded %s", appName))

	spinner = charm.NewDotsSpinner(fmt.Sprintf("Installing %s", appName))
	spinner.Start()

		if err := installDownloadedFile(downloadedFile, appName); err != nil {
			// Check if extraction succeeded but installation failed
			if extractErr, ok := err.(*ExtractionSucceededError); ok {
				spinner.Warning("Extraction succeeded, but automatic installation failed")
				// Provide helpful feedback about where the app was extracted
				fmt.Printf("\n✓ %s was successfully downloaded and extracted to:\n", appName)
				fmt.Printf("  %s\n", extractErr.ExtractDir)
				fmt.Printf("\nPlease manually move the application to your Applications folder.\n\n")
				// Return error so caller can handle it appropriately (won't fall back to brew)
				return extractErr
			}
			spinner.Error(fmt.Sprintf("Failed to install %s", appName))
			return fmt.Errorf("failed to install %s: %w", appName, err)
		}

	spinner.Success(fmt.Sprintf("%s installed successfully", appName))
	return nil
}


// SourceURL returns the source URL for an app if it exists
func SourceURL(appName string) (string, bool, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", false, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Sources == nil {
		return "", false, nil
	}

	sourceURL, exists := cfg.Sources[appName]
	if !exists || sourceURL == "" {
		return "", false, nil
	}

	return sourceURL, true, nil
}
