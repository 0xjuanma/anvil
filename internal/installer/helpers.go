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
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/utils"
)

// runCommandWithSpinner executes a system command with spinner feedback
func runCommandWithSpinner(spinnerMsg, errorMsg string, command string, args ...string) error {
	spinner := charm.NewDotsSpinner(spinnerMsg)
	spinner.Start()

	result, err := system.RunCommand(command, args...)
	if err != nil || !result.Success {
		spinner.Error(errorMsg)
		if result.Error != "" {
			return fmt.Errorf("%s: %s", errorMsg, result.Error)
		}
		return fmt.Errorf("%s: %w", errorMsg, err)
	}

	spinner.Success(spinnerMsg)
	return nil
}

// ensureApplicationsDirectory ensures the Applications directory exists and returns its path
func ensureApplicationsDirectory() (string, error) {
	homeDir, _ := system.HomeDir()
	applicationsDir := filepath.Join(homeDir, "Applications")
	if err := utils.EnsureDirectory(applicationsDir); err != nil {
		return "", fmt.Errorf("failed to create Applications directory: %w", err)
	}
	return applicationsDir, nil
}

// ensureLinuxApplicationsDirectory ensures the Linux applications directory exists and returns its path
func ensureLinuxApplicationsDirectory(appName string) (string, error) {
	homeDir, _ := system.HomeDir()
	destDir := filepath.Join(homeDir, ".local", "share", "applications", appName)
	if err := utils.EnsureDirectory(filepath.Dir(destDir)); err != nil {
		return "", fmt.Errorf("failed to create destination directory: %w", err)
	}
	return destDir, nil
}

// ensureExtractDirectory creates and returns an extract directory path
func ensureExtractDirectory(filePath, appName string) (string, error) {
	extractDir := filepath.Join(filepath.Dir(filePath), appName+"-extracted")
	if err := utils.EnsureDirectory(extractDir); err != nil {
		return "", fmt.Errorf("failed to create extract directory: %w", err)
	}
	return extractDir, nil
}

// copyAppToApplications copies an application to the Applications directory
func copyAppToApplications(appPath, destPath string) error {
	if err := utils.CopyDirectorySimple(appPath, destPath); err != nil {
		return fmt.Errorf("failed to copy application: %w", err)
	}
	return nil
}

