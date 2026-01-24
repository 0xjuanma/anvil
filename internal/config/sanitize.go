package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/constants"
	"gopkg.in/yaml.v2"
)

// PII masking placeholder values
const (
	RedactedUsername   = "REDACTED_USERNAME"
	RedactedEmail      = "REDACTED_EMAIL"
	RedactedSSHKeyPath = "REDACTED_SSH_KEY_PATH"
)

// SanitizeSettingsForPush creates a deep copy of the config with masked git section.
// This ensures no PII (username, email, SSH key path) is pushed to the remote repository.
func SanitizeSettingsForPush(original *AnvilConfig) (*AnvilConfig, error) {
	if original == nil {
		return nil, fmt.Errorf("cannot sanitize nil config")
	}

	// Create a deep copy by marshaling and unmarshaling
	data, err := yaml.Marshal(original)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal config for copy: %w", err)
	}

	var sanitized AnvilConfig
	if err := yaml.Unmarshal(data, &sanitized); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config copy: %w", err)
	}

	// Mask the git config section
	MaskGitConfig(&sanitized.Git)

	return &sanitized, nil
}

// MaskGitConfig replaces PII values in GitConfig with placeholder values.
// This masks username, email, and ssh_key_path to prevent accidental commits.
func MaskGitConfig(gitConfig *GitConfig) {
	if gitConfig == nil {
		return
	}

	// Mask all git config fields that contain PII
	gitConfig.Username = RedactedUsername
	gitConfig.Email = RedactedEmail
	gitConfig.SSHKeyPath = RedactedSSHKeyPath
}

// CreateSanitizedTempFile writes a sanitized version of the config to a temporary file.
// Returns the path to the temporary file. Caller is responsible for cleanup.
func CreateSanitizedTempFile(original *AnvilConfig) (string, func(), error) {
	// Create sanitized copy
	sanitized, err := SanitizeSettingsForPush(original)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sanitize config: %w", err)
	}

	// Create temp directory if it doesn't exist
	tempDir := filepath.Join(AnvilConfigDirectory(), "temp", ".sanitized")
	if err := os.MkdirAll(tempDir, constants.DirPerm); err != nil {
		return "", nil, fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Create temp file path
	tempFilePath := filepath.Join(tempDir, constants.ANVIL_CONFIG_FILE)

	// Marshal sanitized config to YAML
	data, err := yaml.Marshal(sanitized)
	if err != nil {
		return "", nil, fmt.Errorf("failed to marshal sanitized config: %w", err)
	}

	// Write to temp file
	if err := os.WriteFile(tempFilePath, data, constants.FilePerm); err != nil {
		return "", nil, fmt.Errorf("failed to write sanitized config: %w", err)
	}

	// Create cleanup function
	cleanup := func() {
		os.Remove(tempFilePath)
		// Try to remove the temp directory if empty
		os.Remove(tempDir)
	}

	return tempFilePath, cleanup, nil
}

// IsGitConfigMasked checks if the git config contains masked/placeholder values.
// Returns true if any of the git config fields contain redacted placeholders.
func IsGitConfigMasked(gitConfig *GitConfig) bool {
	if gitConfig == nil {
		return false
	}

	return gitConfig.Username == RedactedUsername ||
		gitConfig.Email == RedactedEmail ||
		gitConfig.SSHKeyPath == RedactedSSHKeyPath
}

// RegenerateGitConfigIfMasked checks if git config is masked and regenerates it from system.
// Returns true if regeneration was performed, false if config was already valid.
func RegenerateGitConfigIfMasked(config *AnvilConfig) (bool, error) {
	if config == nil {
		return false, fmt.Errorf("cannot regenerate git config for nil config")
	}

	// Check if any values are masked
	if !IsGitConfigMasked(&config.Git) {
		return false, nil
	}

	// Regenerate from system
	if err := PopulateGitConfigFromSystem(&config.Git); err != nil {
		return false, fmt.Errorf("failed to regenerate git config: %w", err)
	}

	return true, nil
}
