package system

import (
	"os"
	"runtime"
)

// getType returns the type of the operating system
func getType() string {
	return runtime.GOOS
}

// IsMacOS returns true if the current OS is macOS
func IsMacOS() bool {
	return getType() == "darwin"
}

// IsLinux returns true if the current OS is Linux
func IsLinux() bool {
	return getType() == "linux"
}

// HomeDir returns the user's home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}
