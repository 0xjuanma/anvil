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

// Package utils provides utility functions for file operations, formatting, and rendering.
// It includes file system operations, text formatting helpers, and tree/list rendering utilities.
package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/0xjuanma/anvil/internal/constants"
)

// CopyOptions holds options for file and directory copying operations
type CopyOptions struct {
	// Common options
	Overwrite     bool
	PreservePerms bool
	FileMode      os.FileMode

	// Directory-specific options (ignored for files)
	IncludeHidden bool
	DirMode       os.FileMode
	Merge         bool

	// File-specific options (ignored for directories)
	CreateDirs bool
}

// DefaultCopyOptions returns default options for file and directory copying
func DefaultCopyOptions() CopyOptions {
	return CopyOptions{
		Overwrite:     true,
		PreservePerms: false,
		FileMode:      constants.FilePerm,
		IncludeHidden: true,
		DirMode:       constants.DirPerm,
		Merge:         true,
		CreateDirs:    true,
	}
}

// CopyFile copies a file from src to dst with configurable options.
func CopyFile(src, dst string, options CopyOptions) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source file error: %w", err)
	}

	if srcInfo.IsDir() {
		return fmt.Errorf("source path is a directory, not a file: %s", src)
	}

	// Check if destination exists and handle overwrite
	if _, err := os.Stat(dst); err == nil && !options.Overwrite {
		return fmt.Errorf("destination exists: %s", dst)
	}

	if options.CreateDirs {
		if err := EnsureDirectory(filepath.Dir(dst)); err != nil {
			return fmt.Errorf("failed to create destination directory: %w", err)
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	fileMode := options.FileMode
	if options.PreservePerms {
		fileMode = srcInfo.Mode()
	}

	dstFile, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileMode)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	return nil
}

// CopyFileSimple copies a file using default options.
func CopyFileSimple(src, dst string) error {
	return CopyFile(src, dst, DefaultCopyOptions())
}

// CopyDirectory recursively copies a directory from src to dst with configurable options.
func CopyDirectory(src, dst string, options CopyOptions) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("source directory error: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("source path is not a directory: %s", src)
	}

	if options.Overwrite && !options.Merge {
		if err := os.RemoveAll(dst); err != nil {
			return fmt.Errorf("failed to remove existing destination: %w", err)
		}
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("walk %s: %w", path, err)
		}

		if !options.IncludeHidden && isHidden(info.Name()) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}
		destPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			dirMode := options.DirMode
			if options.PreservePerms {
				dirMode = info.Mode()
			}
			return os.MkdirAll(destPath, dirMode)
		}

		fileOptions := CopyOptions{
			CreateDirs:    true,
			Overwrite:     options.Overwrite,
			PreservePerms: options.PreservePerms,
			FileMode:      options.FileMode,
		}
		return CopyFile(path, destPath, fileOptions)
	})
}

// CopyDirectorySimple copies a directory using default options.
func CopyDirectorySimple(src, dst string) error {
	return CopyDirectory(src, dst, DefaultCopyOptions())
}

// isHidden checks if a file/directory name represents a hidden item
func isHidden(name string) bool {
	return len(name) > 0 && name[0] == '.'
}

// EnsureDirectory creates a directory and all necessary parent directories
func EnsureDirectory(path string) error {
	return os.MkdirAll(path, constants.DirPerm)
}
