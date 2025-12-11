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
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/utils"
)

// createArchiveDirectory creates a timestamped archive directory.
func createArchiveDirectory(prefix string) (string, error) {

	// Create timestamp
	timestamp := time.Now().Format("2006-01-02-15-04-05")
	archiveName := fmt.Sprintf("%s-%s", prefix, timestamp)
	archivePath := filepath.Join(config.AnvilConfigDirectory(), "archive", archiveName)

	// Create archive directory
	if err := utils.EnsureDirectory(archivePath); err != nil {
		return "", err
	}

	return archivePath, nil
}

// archiveExistingConfig archives the existing configuration.
func archiveExistingConfig(configType, sourcePath, archivePath string) error {
	// Check if source exists
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		// Nothing to archive
		return nil
	}

	// Determine destination in archive
	var destPath string
	if configType == "anvil-settings" {
		destPath = filepath.Join(archivePath, constants.ANVIL_CONFIG_FILE)
	} else {
		// For app configs, preserve the directory structure
		destPath = archivePath
	}

	// Copy to archive
	sourceInfo, err := os.Stat(sourcePath)
	if err != nil {
		return err
	}

	if sourceInfo.IsDir() {
		return utils.CopyDirectorySimple(sourcePath, destPath)
	} else {
		// Ensure parent directory exists
		if err := utils.EnsureDirectory(filepath.Dir(destPath)); err != nil {
			return err
		}
		return utils.CopyFileSimple(sourcePath, destPath)
	}
}
