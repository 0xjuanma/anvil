package system

import (
	"os"
	"runtime"
)

// IsMacOS returns true if the current OS is macOS
func IsMacOS() bool {
	return runtime.GOOS == "darwin"
}

// IsLinux returns true if the current OS is Linux
func IsLinux() bool {
	return runtime.GOOS == "linux"
}

// HomeDir returns the user's home directory
func HomeDir() (string, error) {
	return os.UserHomeDir()
}
