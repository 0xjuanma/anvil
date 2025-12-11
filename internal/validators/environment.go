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

package validators

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/utils"
)

// InitRunValidator checks if anvil init has been run successfully
type InitRunValidator struct{}

func (v *InitRunValidator) Name() string     { return "init-run" }
func (v *InitRunValidator) Category() string { return "environment" }
func (v *InitRunValidator) Description() string {
	return "Verify anvil initialization has been completed"
}
func (v *InitRunValidator) CanFix() bool { return false }

func (v *InitRunValidator) Validate(ctx context.Context, cfg *config.AnvilConfig) *ValidationResult {
	configPath := config.AnvilConfigPath()

	// Check if settings.yaml exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return &ValidationResult{
			Name:     v.Name(),
			Category: v.Category(),
			Status:   FAIL,
			Message:  "Anvil has not been initialized",
			Details:  []string{"Settings file not found at " + configPath},
			FixHint:  "Run 'anvil init' to set up your environment",
			AutoFix:  false,
		}
	}

	// Check if basic required directories exist
	anvilDir := filepath.Dir(configPath)
	if _, err := os.Stat(anvilDir); os.IsNotExist(err) {
		return &ValidationResult{
			Name:     v.Name(),
			Category: v.Category(),
			Status:   FAIL,
			Message:  "Anvil directory structure missing",
			Details:  []string{"Directory not found: " + anvilDir},
			FixHint:  "Run 'anvil init' to recreate directory structure",
			AutoFix:  false,
		}
	}

	return &ValidationResult{
		Name:     v.Name(),
		Category: v.Category(),
		Status:   PASS,
		Message:  "Anvil initialization complete",
		Details:  []string{"Settings file found at " + configPath},
		AutoFix:  false,
	}
}

func (v *InitRunValidator) Fix(ctx context.Context, cfg *config.AnvilConfig) error {
	return fmt.Errorf("automatic initialization not supported, run 'anvil init' manually")
}

// SettingsFileValidator validates the settings.yaml file
type SettingsFileValidator struct{}

func (v *SettingsFileValidator) Name() string     { return "settings-file" }
func (v *SettingsFileValidator) Category() string { return "environment" }
func (v *SettingsFileValidator) Description() string {
	return "Validate settings.yaml file exists and is valid"
}
func (v *SettingsFileValidator) CanFix() bool { return false }

func (v *SettingsFileValidator) Validate(ctx context.Context, cfg *config.AnvilConfig) *ValidationResult {
	configPath := config.AnvilConfigPath()

	// Check file exists
	info, err := os.Stat(configPath)
	if os.IsNotExist(err) {
		return &ValidationResult{
			Name:     v.Name(),
			Category: v.Category(),
			Status:   FAIL,
			Message:  "Settings file does not exist",
			Details:  []string{"Expected file: " + configPath},
			FixHint:  "Run 'anvil init' to create settings file",
			AutoFix:  false,
		}
	}

	// Check file permissions
	if info.Mode().Perm() != constants.FilePerm {
		return &ValidationResult{
			Name:     v.Name(),
			Category: v.Category(),
			Status:   WARN,
			Message:  "Settings file has incorrect permissions",
			Details:  []string{fmt.Sprintf("Current: %v, Expected: %v", info.Mode().Perm(), constants.FilePerm)},
			FixHint:  fmt.Sprintf("Run 'chmod %o %s'", constants.FilePerm, configPath),
			AutoFix:  true,
		}
	}

	// Check if file is readable and valid YAML
	_, err = config.LoadConfig()
	if err != nil {
		return &ValidationResult{
			Name:     v.Name(),
			Category: v.Category(),
			Status:   FAIL,
			Message:  "Settings file is not valid YAML",
			Details:  []string{err.Error()},
			FixHint:  "Check YAML syntax in " + configPath,
			AutoFix:  false,
		}
	}

	return &ValidationResult{
		Name:     v.Name(),
		Category: v.Category(),
		Status:   PASS,
		Message:  "Settings file is valid",
		Details:  []string{"File: " + configPath},
		AutoFix:  false,
	}
}

func (v *SettingsFileValidator) Fix(ctx context.Context, cfg *config.AnvilConfig) error {
	configPath := config.AnvilConfigPath()

	// Fix file permissions
	if err := os.Chmod(configPath, constants.FilePerm); err != nil {
		return fmt.Errorf("failed to fix file permissions: %w", err)
	}

	return nil
}

// DirectoryStructureValidator validates the anvil directory structure
type DirectoryStructureValidator struct{}

func (v *DirectoryStructureValidator) Name() string     { return "directory-structure" }
func (v *DirectoryStructureValidator) Category() string { return "environment" }
func (v *DirectoryStructureValidator) Description() string {
	return "Verify anvil directory structure is correct"
}
func (v *DirectoryStructureValidator) CanFix() bool { return true }

func (v *DirectoryStructureValidator) Validate(ctx context.Context, cfg *config.AnvilConfig) *ValidationResult {
	anvilDir := config.AnvilConfigDirectory()

	// Required directories
	requiredDirs := []string{
		anvilDir,
		filepath.Join(anvilDir, "temp"),
	}

	var missingDirs []string
	for _, dir := range requiredDirs {
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			missingDirs = append(missingDirs, dir)
		}
	}

	if len(missingDirs) > 0 {
		return &ValidationResult{
			Name:     v.Name(),
			Category: v.Category(),
			Status:   FAIL,
			Message:  "Directory structure incomplete",
			Details:  missingDirs,
			FixHint:  "Missing directories will be created automatically",
			AutoFix:  true,
		}
	}

	// Check directory permissions
	for _, dir := range requiredDirs {
		info, err := os.Stat(dir)
		if err != nil {
			continue
		}
		if info.Mode().Perm() != constants.DirPerm {
			return &ValidationResult{
				Name:     v.Name(),
				Category: v.Category(),
				Status:   WARN,
				Message:  "Directory has incorrect permissions",
				Details:  []string{fmt.Sprintf("Directory: %s, Current: %v, Expected: %v", dir, info.Mode().Perm(), constants.DirPerm)},
				FixHint:  "Directory permissions will be corrected",
				AutoFix:  true,
			}
		}
	}

	return &ValidationResult{
		Name:     v.Name(),
		Category: v.Category(),
		Status:   PASS,
		Message:  "Directory structure is correct",
		Details:  []string{fmt.Sprintf("Base directory: %s", anvilDir)},
		AutoFix:  false,
	}
}

func (v *DirectoryStructureValidator) Fix(ctx context.Context, cfg *config.AnvilConfig) error {
	anvilDir := config.AnvilConfigDirectory()

	// Create required directories
	requiredDirs := []string{
		anvilDir,
		filepath.Join(anvilDir, "temp"),
	}

	for _, dir := range requiredDirs {
		if err := utils.EnsureDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	return nil
}
