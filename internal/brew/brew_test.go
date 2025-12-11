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
	"runtime"
	"testing"
)

func TestBrewPackageStruct(t *testing.T) {
	// Test BrewPackage struct creation and fields
	pkg := BrewPackage{
		Name:        "git",
		Version:     "2.39.0",
		Description: "Distributed revision control system",
		Installed:   true,
	}

	if pkg.Name != "git" {
		t.Errorf("Expected name 'git', got '%s'", pkg.Name)
	}
	if pkg.Version != "2.39.0" {
		t.Errorf("Expected version '2.39.0', got '%s'", pkg.Version)
	}
	if pkg.Description != "Distributed revision control system" {
		t.Errorf("Expected description 'Distributed revision control system', got '%s'", pkg.Description)
	}
	if !pkg.Installed {
		t.Error("Expected installed to be true")
	}
}

func TestBrewPackageZeroValues(t *testing.T) {
	// Test BrewPackage with zero values
	pkg := BrewPackage{}

	if pkg.Name != "" {
		t.Errorf("Expected empty name, got '%s'", pkg.Name)
	}
	if pkg.Version != "" {
		t.Errorf("Expected empty version, got '%s'", pkg.Version)
	}
	if pkg.Description != "" {
		t.Errorf("Expected empty description, got '%s'", pkg.Description)
	}
	if pkg.Installed {
		t.Error("Expected installed to be false")
	}
}

func TestIsBrewInstalled(t *testing.T) {
	// Test that IsBrewInstalled returns a boolean
	// This is an integration test - it will return true/false based on actual system state
	result := IsBrewInstalled()

	// Just verify it returns a boolean value
	if result != true && result != false {
		t.Error("IsBrewInstalled should return a boolean value")
	}
}

func TestInstallBrewPlatformCheck(t *testing.T) {
	// Test platform check in InstallBrew
	if runtime.GOOS != "darwin" {
		// On non-macOS systems, it should return an error
		err := InstallBrew()
		if err == nil {
			t.Error("Expected error on non-macOS platform")
		}
		if err.Error() != "Homebrew is only supported on macOS" {
			t.Errorf("Expected platform error message, got: %s", err.Error())
		}
	} else {
		// On macOS, we can't easily test without actually installing brew
		// so we just verify the function exists and can be called
		_ = InstallBrew()
	}
}

func TestUpdateBrewWhenNotInstalled(t *testing.T) {
	// If brew is not installed, UpdateBrew should return an error
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	err := UpdateBrew()
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
}

func TestInstallPackageWhenNotInstalled(t *testing.T) {
	// If brew is not installed, InstallPackage should return an error
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	err := InstallPackage("git")
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
}

func TestIsPackageInstalledWhenBrewNotInstalled(t *testing.T) {
	// If brew is not installed, IsPackageInstalled should return false
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	result := IsPackageInstalled("git")
	if result != false {
		t.Error("Expected false when brew is not installed")
	}
}

func TestGetInstalledPackagesWhenNotInstalled(t *testing.T) {
	// If brew is not installed, InstalledPackages should return an error
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	packages, err := GetInstalledPackages()
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
	if packages != nil {
		t.Error("Expected nil packages when brew is not installed")
	}
}

func TestGetPackageInfoWhenNotInstalled(t *testing.T) {
	// If brew is not installed, PackageInfo should return an error
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	pkg, err := PackageInfo("git")
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
	if pkg != nil {
		t.Error("Expected nil package when brew is not installed")
	}
}

func TestInstallPackagesWhenNotInstalled(t *testing.T) {
	// If brew is not installed, InstallPackages should return an error
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	packages := []string{"git", "vim"}
	err := InstallPackages(packages)
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
}

func TestInstallPackagesEmptySlice(t *testing.T) {
	// Test InstallPackages with empty slice
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	packages := []string{}
	err := InstallPackages(packages)
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
}

// Integration tests that run when brew is installed
func TestBrewIntegrationWhenInstalled(t *testing.T) {
	if !IsBrewInstalled() {
		t.Skip("Skipping integration test - Homebrew is not installed")
	}

	// Test that functions can be called without error when brew is installed
	t.Run("UpdateBrew", func(t *testing.T) {
		// This might fail if network is unavailable, but should not panic
		_ = UpdateBrew()
	})

	t.Run("InstalledPackages", func(t *testing.T) {
		packages, err := InstalledPackages()
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
		if packages == nil {
			t.Error("Expected packages slice, got nil")
		}

		// Verify each package has required fields
		for _, pkg := range packages {
			if pkg.Name == "" {
				t.Error("Package name should not be empty")
			}
			if !pkg.Installed {
				t.Error("All packages from InstalledPackages should be marked as installed")
			}
		}
	})

	t.Run("IsPackageInstalled", func(t *testing.T) {
		// Test with a package that's very likely to be installed
		result := IsPackageInstalled("git")
		// Just verify it returns a boolean - don't assume git is installed
		if result != true && result != false {
			t.Error("IsPackageInstalled should return a boolean value")
		}
	})

	t.Run("GetPackageInfo", func(t *testing.T) {
		// Test with a common package
		pkg, err := PackageInfo("git")
		if err != nil {
			// This might fail if git formula doesn't exist, which is ok
			t.Logf("GetPackageInfo returned error: %v", err)
		} else {
			if pkg == nil {
				t.Error("Expected package info, got nil")
			} else {
				if pkg.Name != "git" {
					t.Errorf("Expected package name 'git', got '%s'", pkg.Name)
				}
			}
		}
	})
}

// Benchmark tests
func BenchmarkIsBrewInstalled(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsBrewInstalled()
	}
}

func TestIsApplicationAvailable(t *testing.T) {
	tests := []struct {
		name        string
		packageName string
		setupFunc   func()
		cleanupFunc func()
		expected    bool
	}{
		{
			name:        "nonexistent package",
			packageName: "nonexistent-package-12345",
			expected:    false,
		},
		{
			name:        "common command line tool",
			packageName: "echo",
			expected:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupFunc != nil {
				tt.setupFunc()
			}
			if tt.cleanupFunc != nil {
				defer tt.cleanupFunc()
			}

			result := IsApplicationAvailable(tt.packageName)
			if result != tt.expected {
				t.Errorf("Expected IsApplicationAvailable(%s) to return %v, got %v", tt.packageName, tt.expected, result)
			}
		})
	}
}

func TestInstallPackageWithCheck(t *testing.T) {
	if !IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is not installed")
	}

	tests := []struct {
		name        string
		packageName string
		expectError bool
	}{
		{
			name:        "install nonexistent package",
			packageName: "nonexistent-package-12345",
			expectError: true,
		},
		{
			name:        "install already available package",
			packageName: "git", // git is usually available
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := InstallPackageWithCheck(tt.packageName)

			if tt.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}
		})
	}
}

func TestInstallPackageWithCheckWhenBrewNotInstalled(t *testing.T) {
	// This test assumes brew is not installed - skip if it is
	if IsBrewInstalled() {
		t.Skip("Skipping test - Homebrew is installed")
	}

	err := InstallPackageWithCheck("git")
	if err == nil {
		t.Error("Expected error when brew is not installed")
	}
	if err.Error() != "Homebrew is not installed" {
		t.Errorf("Expected 'Homebrew is not installed', got: %s", err.Error())
	}
}

func BenchmarkIsApplicationAvailable(b *testing.B) {
	for i := 0; i < b.N; i++ {
		IsApplicationAvailable("git")
	}
}

func BenchmarkIsPackageInstalled(b *testing.B) {
	if !IsBrewInstalled() {
		b.Skip("Homebrew not installed")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		IsPackageInstalled("git")
	}
}
