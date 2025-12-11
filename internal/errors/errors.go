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

// Package errors provides structured error types for the Anvil CLI.
// It defines error categories and helper functions for creating typed errors
// that can be programmatically inspected using errors.Is and errors.As.
package errors

import (
	"fmt"
	"strings"
)

// ErrorType represents different categories of errors
type ErrorType int

const (
	// ErrorTypeGeneral represents general errors
	ErrorTypeGeneral ErrorType = iota
	// ErrorTypePlatform represents platform-specific errors
	ErrorTypePlatform
	// ErrorTypeValidation represents validation errors
	ErrorTypeValidation
	// ErrorTypeConfiguration represents configuration errors
	ErrorTypeConfiguration
	// ErrorTypeInstallation represents installation errors
	ErrorTypeInstallation
	// ErrorTypeNetwork represents network-related errors
	ErrorTypeNetwork
	// ErrorTypeFileSystem represents file system errors
	ErrorTypeFileSystem
)

// String returns a string representation of the error type
func (et ErrorType) String() string {
	switch et {
	case ErrorTypePlatform:
		return "platform"
	case ErrorTypeValidation:
		return "validation"
	case ErrorTypeConfiguration:
		return "configuration"
	case ErrorTypeInstallation:
		return "installation"
	case ErrorTypeNetwork:
		return "network"
	case ErrorTypeFileSystem:
		return "filesystem"
	default:
		return "general"
	}
}

// AnvilError represents a structured error with operation, command, and type context
type AnvilError struct {
	Op      string    // The operation being performed (init, setup, config, etc.)
	Command string    // The specific command or subcommand
	Type    ErrorType // The category of error
	Err     error     // The underlying error
	Context string    // Additional context information
}

// Error implements the error interface with improved formatting
func (e *AnvilError) Error() string {
	var parts []string

	// Build the error prefix
	if e.Command != "" {
		parts = append(parts, fmt.Sprintf("anvil %s %s", e.Op, e.Command))
	} else {
		parts = append(parts, fmt.Sprintf("anvil %s", e.Op))
	}

	// Add error type if not general
	if e.Type != ErrorTypeGeneral {
		parts = append(parts, fmt.Sprintf("[%s]", e.Type.String()))
	}

	// Add context if available
	if e.Context != "" {
		parts = append(parts, fmt.Sprintf("(%s)", e.Context))
	}

	// Join with the error message
	prefix := strings.Join(parts, " ")
	return fmt.Sprintf("%s: %v", prefix, e.Err)
}

// Unwrap returns the underlying error
func (e *AnvilError) Unwrap() error {
	return e.Err
}

// Is checks if the error matches the target error type
func (e *AnvilError) Is(target error) bool {
	if t, ok := target.(*AnvilError); ok {
		return e.Type == t.Type && e.Op == t.Op && e.Command == t.Command
	}
	return false
}

// NewAnvilErrorWithType creates a new AnvilError with specified type.
// Use this for custom error types or when you need full control over error creation.
func NewAnvilErrorWithType(op, command string, errType ErrorType, err error) *AnvilError {
	return &AnvilError{
		Op:      op,
		Command: command,
		Type:    errType,
		Err:     err,
	}
}

// NewValidationError creates a validation error.
// Use this for input validation failures.
func NewValidationError(op, command string, err error) *AnvilError {
	return NewAnvilErrorWithType(op, command, ErrorTypeValidation, err)
}

// NewConfigurationError creates a configuration error.
// Use this for configuration file or settings issues.
func NewConfigurationError(op, command string, err error) *AnvilError {
	return NewAnvilErrorWithType(op, command, ErrorTypeConfiguration, err)
}

// NewInstallationError creates an installation error.
// Use this for tool or package installation failures.
func NewInstallationError(op, command string, err error) *AnvilError {
	return NewAnvilErrorWithType(op, command, ErrorTypeInstallation, err)
}

// NewNetworkError creates a network error.
// Use this for network-related failures (API calls, downloads, etc.).
func NewNetworkError(op, command string, err error) *AnvilError {
	return NewAnvilErrorWithType(op, command, ErrorTypeNetwork, err)
}

// NewFileSystemError creates a file system error.
// Use this for file or directory operation failures.
func NewFileSystemError(op, command string, err error) *AnvilError {
	return NewAnvilErrorWithType(op, command, ErrorTypeFileSystem, err)
}
