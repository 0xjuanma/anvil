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

package sync

import (
	"os"
	"path/filepath"
	"testing"
)

func setupTestEnv(t *testing.T) (anvilDir, archiveDir string, cleanup func()) {
	t.Helper()

	originalAnvilDir := os.Getenv("ANVIL_CONFIG_DIR")
	os.Setenv("ANVIL_TEST_MODE", "true")

	tempDir := t.TempDir()
	anvilDir = filepath.Join(tempDir, ".anvil")
	archiveDir = filepath.Join(anvilDir, "archive")

	if err := os.MkdirAll(archiveDir, 0755); err != nil {
		t.Fatal(err)
	}

	os.Setenv("ANVIL_CONFIG_DIR", anvilDir)

	cleanup = func() {
		os.Setenv("ANVIL_TEST_MODE", "")
		if originalAnvilDir != "" {
			os.Setenv("ANVIL_CONFIG_DIR", originalAnvilDir)
		} else {
			os.Unsetenv("ANVIL_CONFIG_DIR")
		}
	}

	return anvilDir, archiveDir, cleanup
}

func TestPerformSync_SingleFile(t *testing.T) {
	anvilDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceFile := filepath.Join(anvilDir, "source.yaml")
	destFile := filepath.Join(anvilDir, "dest.yaml")

	if err := os.WriteFile(sourceFile, []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}

	opts := SyncOptions{
		ArchivePrefix:  "test-sync",
		SourcePath:     sourceFile,
		DestPath:       destFile,
		ConfirmMsg:     "Confirm sync?",
		SpinnerMsg:     "Syncing...",
		SpinnerSuccess: "Synced",
		SuccessMsg:     "Success",
	}
	err := performSync(opts)

	if err != nil {
		t.Fatalf("performSync failed: %v", err)
	}

	content, err := os.ReadFile(destFile)
	if err != nil {
		t.Fatalf("Failed to read destination: %v", err)
	}
	if string(content) != "new content" {
		t.Errorf("Content mismatch: got %q, want %q", string(content), "new content")
	}
}

func TestPerformSync_Directory_PreservesLocalFiles(t *testing.T) {
	anvilDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceDir := filepath.Join(anvilDir, "source")
	destDir := filepath.Join(anvilDir, "dest")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sourceDir, "remote.conf"), []byte("remote"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "local-only.conf"), []byte("local"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "shared.conf"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "shared.conf"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	err := performSync(
		"test-dir-sync",
		sourceDir,
		destDir,
		"Confirm sync?",
		"Syncing...",
		"Synced",
		"Success",
	)

	if err != nil {
		t.Fatalf("performSync failed: %v", err)
	}

	tests := []struct {
		file    string
		content string
		desc    string
	}{
		{"remote.conf", "remote", "remote file copied"},
		{"local-only.conf", "local", "local-only file preserved"},
		{"shared.conf", "new", "shared file updated from remote"},
	}

	for _, tt := range tests {
		t.Run(tt.file, func(t *testing.T) {
			content, err := os.ReadFile(filepath.Join(destDir, tt.file))
			if err != nil {
				t.Fatalf("%s not found: %v", tt.desc, err)
			}
			if string(content) != tt.content {
				t.Errorf("%s: got %q, want %q", tt.desc, string(content), tt.content)
			}
		})
	}
}

func TestPerformSync_CreatesArchive(t *testing.T) {
	anvilDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceDir := filepath.Join(anvilDir, "source")
	destDir := filepath.Join(anvilDir, "dest")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "old.txt"), []byte("old"), 0644); err != nil {
		t.Fatal(err)
	}

	err := performSync(
		"archive-test",
		sourceDir,
		destDir,
		"Confirm?",
		"Syncing...",
		"Done",
		"Success",
	)

	if err != nil {
		t.Fatalf("performSync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "new.txt")); err != nil {
		t.Error("New file not synced")
	}
	if _, err := os.Stat(filepath.Join(destDir, "old.txt")); err != nil {
		t.Error("Old file not preserved during sync")
	}
}

func TestPerformSync_NestedDirectories(t *testing.T) {
	anvilDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceDir := filepath.Join(anvilDir, "source")
	destDir := filepath.Join(anvilDir, "dest")

	sourceNested := filepath.Join(sourceDir, "nested", "deep")
	destNested := filepath.Join(destDir, "nested")

	if err := os.MkdirAll(sourceNested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destNested, 0755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(filepath.Join(sourceNested, "remote.txt"), []byte("remote"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destNested, "local.txt"), []byte("local"), 0644); err != nil {
		t.Fatal(err)
	}

	err := performSync(
		"nested-test",
		sourceDir,
		destDir,
		"Confirm?",
		"Syncing...",
		"Done",
		"Success",
	)

	if err != nil {
		t.Fatalf("performSync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "nested", "deep", "remote.txt")); err != nil {
		t.Error("Remote nested file not copied")
	}
	if _, err := os.Stat(filepath.Join(destDir, "nested", "local.txt")); err != nil {
		t.Error("Local nested file not preserved")
	}
}

func TestPerformSync_SourceNotExists(t *testing.T) {
	anvilDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceFile := filepath.Join(anvilDir, "nonexistent.yaml")
	destFile := filepath.Join(anvilDir, "dest.yaml")

	err := performSync(
		"error-test",
		sourceFile,
		destFile,
		"Confirm?",
		"Syncing...",
		"Done",
		"Success",
	)

	if err == nil {
		t.Error("Expected error for non-existent source, got nil")
	}
}

func TestPerformSync_WithExistingDestination(t *testing.T) {
	anvilDir, _, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceDir := filepath.Join(anvilDir, "source")
	destDir := filepath.Join(anvilDir, "dest")

	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "new.txt"), []byte("new content"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(destDir, "old.txt"), []byte("old content"), 0644); err != nil {
		t.Fatal(err)
	}

	err := performSync(
		"overwrite-test",
		sourceDir,
		destDir,
		"Confirm?",
		"Syncing...",
		"Done",
		"Success",
	)

	if err != nil {
		t.Fatalf("performSync failed: %v", err)
	}

	if _, err := os.Stat(filepath.Join(destDir, "new.txt")); err != nil {
		t.Error("New file not copied")
	}
	if _, err := os.Stat(filepath.Join(destDir, "old.txt")); err != nil {
		t.Error("Old file not preserved")
	}
}

func TestArchiveExistingConfig_File(t *testing.T) {
	anvilDir, archiveDir, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceFile := filepath.Join(anvilDir, "settings.yaml")
	if err := os.WriteFile(sourceFile, []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(archiveDir, "test-archive")
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		t.Fatal(err)
	}

	err := archiveExistingConfig("anvil-settings", sourceFile, archivePath)
	if err != nil {
		t.Fatalf("archiveExistingConfig failed: %v", err)
	}

	archivedFile := filepath.Join(archivePath, "settings.yaml")
	if _, err := os.Stat(archivedFile); err != nil {
		t.Error("Archived file not created")
	}
}

func TestArchiveExistingConfig_Directory(t *testing.T) {
	anvilDir, archiveDir, cleanup := setupTestEnv(t)
	defer cleanup()

	sourceDir := filepath.Join(anvilDir, "config")
	if err := os.MkdirAll(sourceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sourceDir, "file.txt"), []byte("content"), 0644); err != nil {
		t.Fatal(err)
	}

	archivePath := filepath.Join(archiveDir, "test-archive")
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		t.Fatal(err)
	}

	err := archiveExistingConfig("test-configs", sourceDir, archivePath)
	if err != nil {
		t.Fatalf("archiveExistingConfig failed: %v", err)
	}
}

func TestArchiveExistingConfig_SourceNotExists(t *testing.T) {
	_, archiveDir, cleanup := setupTestEnv(t)
	defer cleanup()

	archivePath := filepath.Join(archiveDir, "test-archive")
	if err := os.MkdirAll(archivePath, 0755); err != nil {
		t.Fatal(err)
	}

	err := archiveExistingConfig("test-config", "/nonexistent", archivePath)
	if err != nil {
		t.Errorf("Expected nil for non-existent source, got: %v", err)
	}
}

func TestCreateArchiveDirectory(t *testing.T) {
	_, _, cleanup := setupTestEnv(t)
	defer cleanup()

	archivePath, err := createArchiveDirectory("test-prefix")
	if err != nil {
		t.Fatalf("createArchiveDirectory failed: %v", err)
	}

	if _, err := os.Stat(archivePath); err != nil {
		t.Error("Archive directory not created")
	}

	if !filepath.IsAbs(archivePath) {
		t.Error("Archive path is not absolute")
	}
}
