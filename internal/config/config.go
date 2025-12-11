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
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/anvil/internal/version"
	"gopkg.in/yaml.v2"
)

//go:embed settings-sample.yaml
var sampleConfigData []byte

// Configuration cache to avoid repeated file I/O operations
var (
	configCache      *AnvilConfig
	configCacheMutex sync.RWMutex
)

var builtInGroups = []string{"dev", "essentials"}

// AnvilGroups represents grouped tool configurations
type AnvilGroups map[string][]string

// GitConfig represents git configuration
type GitConfig struct {
	Username   string `yaml:"username"`
	Email      string `yaml:"email"`
	SSHKeyPath string `yaml:"ssh_key_path,omitempty"` // Reference to SSH private key
}

// GitHubConfig represents GitHub repository configuration for config sync
type GitHubConfig struct {
	ConfigRepo  string `yaml:"config_repo"`             // GitHub repository URL for configs (e.g., "username/dotfiles")
	Branch      string `yaml:"branch"`                  // Branch to use (default: "main")
	LocalPath   string `yaml:"local_path"`              // Local path where configs are stored/synced
	Token       string `yaml:"token,omitempty"`         // GitHub token (use env var reference)
	TokenEnvVar string `yaml:"token_env_var,omitempty"` // Environment variable name for token
}

// AnvilTools represents tool configurations
type AnvilTools struct {
	RequiredTools []string `yaml:"required_tools"`
	InstalledApps []string `yaml:"installed_apps"` // Tracks individually installed applications
}

// getCachedConfig returns the cached configuration or loads it if not cached
func getCachedConfig() (*AnvilConfig, error) {
	configCacheMutex.RLock()
	if configCache != nil {
		configCacheMutex.RUnlock()
		return configCache, nil
	}
	configCacheMutex.RUnlock()

	configCacheMutex.Lock()
	defer configCacheMutex.Unlock()

	// Double-check after acquiring write lock
	if configCache != nil {
		return configCache, nil
	}

	var err error
	configCache, err = LoadConfig()
	return configCache, err
}

// withConfig executes a function with the cached config, handling common error patterns
func withConfig(fn func(*AnvilConfig) error) error {
	config, err := getCachedConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return fn(config)
}

// withConfigAndSave executes a function with the cached config and saves it
func withConfigAndSave(fn func(*AnvilConfig) error) error {
	config, err := getCachedConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if err := fn(config); err != nil {
		return err
	}
	return SaveConfig(config)
}

// ensureMap initializes a map if it's nil
func ensureMap(m interface{}) {
	switch v := m.(type) {
	case *map[string][]string:
		if *v == nil {
			*v = make(map[string][]string)
		}
	case *map[string]string:
		if *v == nil {
			*v = make(map[string]string)
		}
	}
}

// PopulateGitConfigFromSystem populates git configuration from local git settings and auto-detects SSH keys
func PopulateGitConfigFromSystem(gitConfig *GitConfig) error {
	// Always populate username from local git config
	if gitUser, err := system.RunCommand(constants.GitCommand, constants.GitConfig, constants.GitGlobal, constants.GitUserName); err == nil && gitUser.Success {
		gitConfig.Username = strings.TrimSpace(gitUser.Output)
	}

	// Always populate email from local git config
	if gitEmail, err := system.RunCommand(constants.GitCommand, constants.GitConfig, constants.GitGlobal, constants.GitUserEmail); err == nil && gitEmail.Success {
		gitConfig.Email = strings.TrimSpace(gitEmail.Output)
	}

	// Auto-detect SSH key path from common locations
	homeDir, _ := system.HomeDir()
	sshDir := filepath.Join(homeDir, ".ssh")

	// Common SSH key names in order of preference
	commonKeyNames := []string{
		"id_ed25519",
		"id_ed25519_personal",
		"id_rsa",
		"id_rsa_personal",
		"id_ecdsa",
	}

	// Find the first existing SSH key
	for _, keyName := range commonKeyNames {
		keyPath := filepath.Join(sshDir, keyName)
		if _, err := os.Stat(keyPath); err == nil {
			gitConfig.SSHKeyPath = keyPath
			break
		}
	}

	// If no common keys found, use the default path (will be created if needed)
	if gitConfig.SSHKeyPath == "" {
		gitConfig.SSHKeyPath = filepath.Join(sshDir, "id_ed25519")
	}

	return nil
}

// invalidateCache clears the configuration cache
func invalidateCache() {
	configCacheMutex.Lock()
	defer configCacheMutex.Unlock()
	configCache = nil
}

// LoadSampleConfig loads the sample configuration from the assets file
func LoadSampleConfig() (*AnvilConfig, error) {
	return LoadSampleConfigWithVersion("")
}

// CreateDirectories creates necessary directories for anvil
func CreateDirectories() error {
	configDir := AnvilConfigDirectory()

	// Only create the main config directory
	if err := utils.EnsureDirectory(configDir); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", configDir, err)
	}

	return nil
}

// GenerateDefaultSettings generates the default settings.yaml file
func GenerateDefaultSettings() error {
	return GenerateDefaultSettingsWithVersion(version.Version())
}

// GenerateDefaultSettingsWithVersion generates the default settings.yaml file with a specific version
func GenerateDefaultSettingsWithVersion(version string) error {
	configPath := AnvilConfigPath()

	// Check if settings.yaml already exists
	if _, err := os.Stat(configPath); err == nil {
		return nil // File already exists, don't overwrite
	}

	// Load the sample configuration with version
	config, err := LoadSampleConfigWithVersion(version)
	if err != nil {
		return fmt.Errorf("failed to load sample config: %w", err)
	}

	// Marshal to YAML
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config to YAML: %w", err)
	}

	// Write to file
	if err := os.WriteFile(configPath, data, constants.FilePerm); err != nil {
		return fmt.Errorf("failed to write %s: %w", constants.ANVIL_CONFIG_FILE, err)
	}

	return nil
}

// CheckEnvironmentConfigurations checks local environment configurations
func CheckEnvironmentConfigurations() []string {
	var warnings []string

	// Check Git configuration
	if gitUser, err := system.RunCommand(constants.GitCommand, constants.GitConfig, constants.GitGlobal, constants.GitUserName); err != nil || !gitUser.Success || strings.TrimSpace(gitUser.Output) == "" {
		warnings = append(warnings, fmt.Sprintf("Configure git user.name: %s %s %s %s 'Your Name'", constants.GitCommand, constants.GitConfig, constants.GitGlobal, constants.GitUserName))
	}

	if gitEmail, err := system.RunCommand(constants.GitCommand, constants.GitConfig, constants.GitGlobal, constants.GitUserEmail); err != nil || !gitEmail.Success || strings.TrimSpace(gitEmail.Output) == "" {
		warnings = append(warnings, fmt.Sprintf("Configure git user.email: %s %s %s %s 'your.email@example.com'", constants.GitCommand, constants.GitConfig, constants.GitGlobal, constants.GitUserEmail))
	}

	// Check SSH keys
	homeDir, _ := system.HomeDir()
	sshDir := filepath.Join(homeDir, constants.SSHDir)
	if _, err := os.Stat(sshDir); os.IsNotExist(err) {
		warnings = append(warnings, "Set up SSH keys for GitHub: ssh-keygen -t ed25519 -C 'your.email@example.com'")
	} else {
		// Check for common SSH key files
		keyFiles := []string{"id_rsa", "id_ed25519", "id_ecdsa"}
		hasKey := false
		for _, keyFile := range keyFiles {
			if _, err := os.Stat(filepath.Join(sshDir, keyFile)); err == nil {
				hasKey = true
				break
			}
		}
		if !hasKey {
			warnings = append(warnings, fmt.Sprintf("No SSH keys found in ~/%s - consider generating SSH keys for GitHub", constants.SSHDir))
		}
	}

	// Check for common environment variables
	envVars := []string{constants.EnvEditor, constants.EnvShell}
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			warnings = append(warnings, fmt.Sprintf("Consider setting %s environment variable", envVar))
		}
	}

	return warnings
}

// AddInstalledApp adds an app to the installed apps list if it's not already there
func AddInstalledApp(appName string) error {
	return withConfigAndSave(func(config *AnvilConfig) error {
		// Check if already tracked anywhere
		if tracked, _ := IsAppTracked(appName); tracked {
			return nil
		}

		config.Tools.InstalledApps = append(config.Tools.InstalledApps, appName)
		return nil
	})
}

// InstalledApps returns the list of individually installed applications
func InstalledApps() ([]string, error) {
	var apps []string
	err := withConfig(func(config *AnvilConfig) error {
		apps = config.Tools.InstalledApps
		return nil
	})
	return apps, err
}

// IsAppTracked checks if an app is being tracked in any category
func IsAppTracked(appName string) (bool, error) {
	var found bool
	err := withConfig(func(config *AnvilConfig) error {
		// Check in all tool lists
		for _, tool := range append(config.Tools.RequiredTools, config.Tools.InstalledApps...) {
			if tool == appName {
				found = true
				return nil
			}
		}

		// Check in groups
		for _, tools := range config.Groups {
			for _, tool := range tools {
				if tool == appName {
					found = true
					return nil
				}
			}
		}
		return nil
	})
	return found, err
}

// RemoveInstalledApp removes an app from the installed apps list
func RemoveInstalledApp(appName string) error {
	return withConfigAndSave(func(config *AnvilConfig) error {
		for i, app := range config.Tools.InstalledApps {
			if app == appName {
				config.Tools.InstalledApps = append(config.Tools.InstalledApps[:i], config.Tools.InstalledApps[i+1:]...)
				break
			}
		}
		return nil
	})
}

// LocationSource represents where an app config location was found
type LocationSource int

const (
	LocationConfigs LocationSource = iota // Found in configs section of settings.yaml
	LocationTemp                          // Found in temp directory (pulled but not configured)
)

// String returns a string representation of the location source
func (ls LocationSource) String() string {
	switch ls {
	case LocationConfigs:
		return "configs"
	case LocationTemp:
		return "temp"
	default:
		return "unknown"
	}
}

// AppConfigPath checks if an app has a configured local path in the configs section
func AppConfigPath(appName string) (string, bool, error) {
	config, err := getCachedConfig()
	if err != nil {
		return "", false, fmt.Errorf("failed to load config: %w", err)
	}

	if config.Configs == nil {
		return "", false, nil
	}

	path, exists := config.Configs[appName]
	if !exists {
		return "", false, nil
	}

	// Verify the path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", false, fmt.Errorf("configured path for %s does not exist: %s", appName, path)
	}

	return path, true, nil
}

// TempAppPath checks if an app directory exists in the temp directory (from previous pull)
func TempAppPath(appName string) (string, bool, error) {
	tempPath := filepath.Join(AnvilConfigDirectory(), "temp", appName)
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		return "", false, nil
	}

	return tempPath, true, nil
}

// ResolveAppLocation finds the config location for an app following the priority order
func ResolveAppLocation(appName string) (string, LocationSource, error) {
	// Priority 1: Check configs section in settings.yaml
	if path, found, err := AppConfigPath(appName); err != nil {
		return "", LocationConfigs, err
	} else if found {
		return path, LocationConfigs, nil
	}

	// Priority 2: Check temp directory (pulled configs)
	if path, found, err := TempAppPath(appName); err != nil {
		return "", LocationTemp, err
	} else if found {
		return path, LocationTemp, nil
	}

	// Not found anywhere
	return "", LocationConfigs, fmt.Errorf("app '%s' not found in configs or temp directory", appName)
}

// SetAppConfigPath sets the config path for an app in the configs section
func SetAppConfigPath(appName, configPath string) error {
	return withConfigAndSave(func(config *AnvilConfig) error {
		ensureMap(&config.Configs)
		config.Configs[appName] = configPath
		return nil
	})
}

// ConfiguredApps returns a list of all apps that have configured paths
func ConfiguredApps() ([]string, error) {
	var apps []string
	err := withConfig(func(config *AnvilConfig) error {
		for appName := range config.Configs {
			apps = append(apps, appName)
		}
		return nil
	})
	return apps, err
}
