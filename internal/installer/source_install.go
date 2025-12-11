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

package installer

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/utils"
)

// installDownloadedFile installs the downloaded file based on its type and OS
func installDownloadedFile(filePath, appName string) error {
	if system.IsMacOS() {
		return installOnMacOS(filePath, appName)
	} else if system.IsLinux() {
		return installOnLinux(filePath, appName)
	}
	return fmt.Errorf("unsupported operating system")
}

// installOnMacOS handles installation on macOS
func installOnMacOS(filePath, appName string) error {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case constants.ExtDMG:
		return installDMG(filePath, appName)
	case constants.ExtPKG:
		return installPKG(filePath)
	case constants.ExtZIP:
		return installZIP(filePath, appName)
	default:
		return fmt.Errorf("unsupported file type: %s (supported: %s, %s, %s)", ext, constants.ExtDMG, constants.ExtPKG, constants.ExtZIP)
	}
}

// installOnLinux handles installation on Linux
func installOnLinux(filePath, appName string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	if strings.HasSuffix(baseName, constants.ExtTarGz) {
		return installTarGz(filePath, appName)
	} else if strings.HasSuffix(baseName, constants.ExtTarBz2) {
		return installTarBz2(filePath, appName)
	}

	switch ext {
	case constants.ExtDEB:
		return installDEB(filePath)
	case constants.ExtRPM:
		return installRPM(filePath)
	case constants.ExtAppImage:
		return installAppImage(filePath, appName)
	case constants.ExtZIP:
		return installZIP(filePath, appName)
	default:
		return fmt.Errorf("unsupported file type: %s (supported: %s, %s, %s, %s, %s, %s, %s)", ext, constants.ExtDEB, constants.ExtRPM, constants.ExtAppImage, constants.ExtZIP, constants.ExtTarGz, constants.ExtTarBz2)
	}
}

// installDMG mounts DMG, copies .app to Applications, and unmounts
func installDMG(filePath, appName string) error {
	mountResult, err := system.RunCommand("hdiutil", "attach", filePath, "-nobrowse", "-quiet")
	if err != nil || !mountResult.Success {
		return fmt.Errorf("failed to mount DMG: %s", mountResult.Error)
	}

	mountPath := extractMountPath(mountResult.Output)
	if mountPath == "" {
		system.RunCommand("hdiutil", "detach", mountPath, "-quiet")
		return fmt.Errorf("failed to extract mount path from DMG")
	}

	defer func() {
		system.RunCommand("hdiutil", "detach", mountPath, "-quiet")
	}()

	spinner := charm.NewDotsSpinner("Finding application")
	spinner.Start()
	appPath := findAppInDirectory(mountPath, appName)
	if appPath == "" {
		spinner.Error("Application not found")
		return fmt.Errorf("failed to find .app in DMG")
	}
	spinner.Success("Application found")

	applicationsDir, err := ensureApplicationsDirectory()
	if err != nil {
		return err
	}

	appNameFromPath := filepath.Base(appPath)
	destPath := filepath.Join(applicationsDir, appNameFromPath)

	if err := utils.CopyDirectorySimple(appPath, destPath); err != nil {
		return fmt.Errorf("failed to copy application: %w", err)
	}

	spinner = charm.NewDotsSpinner("Installing to Applications")
	spinner.Success("Application installed")
	return nil
}

// installPKG installs a .pkg file using installer command
func installPKG(filePath string) error {
	return runCommandWithSpinner(
		"Installing package",
		"Failed to install package",
		"sudo", "installer", "-pkg", filePath, "-target", "/",
	)
}

// installZIP extracts ZIP and handles contents
func installZIP(filePath, appName string) error {
	extractDir, err := ensureExtractDirectory(filePath, appName)
	if err != nil {
		return err
	}

	if err := runCommandWithSpinner(
		"Extracting ZIP",
		"Failed to extract ZIP",
		"unzip", "-q", filePath, "-d", extractDir,
	); err != nil {
		return err
	}

	if system.IsMacOS() {
		return handleExtractedContentsMacOS(extractDir, appName)
	}
	return handleExtractedContentsLinux(extractDir, appName)
}

// installDEB installs a .deb package
func installDEB(filePath string) error {
	if err := runCommandWithSpinner(
		"Installing DEB package",
		"Failed to install DEB package",
		"sudo", "dpkg", "-i", filePath,
	); err != nil {
		return err
	}

	// Attempt dependency resolution (non-critical)
	if result, err := system.RunCommand("sudo", "apt-get", "-f", "install", "-y"); err != nil || !result.Success {
		spinner := charm.NewDotsSpinner("Installing DEB package")
		spinner.Warning("Dependency resolution had issues")
	}

	return nil
}

// installRPM installs an .rpm package
func installRPM(filePath string) error {
	var command string
	var args []string

	if system.CommandExists("dnf") {
		command = "sudo"
		args = []string{"dnf", "install", "-y", filePath}
	} else if system.CommandExists("yum") {
		command = "sudo"
		args = []string{"yum", "install", "-y", filePath}
	} else {
		command = "sudo"
		args = []string{"rpm", "-i", filePath}
	}

	return runCommandWithSpinner(
		"Installing RPM package",
		"Failed to install RPM package",
		command, args...,
	)
}

// installAppImage makes AppImage executable and optionally installs it
func installAppImage(filePath, appName string) error {
	appImageDir, err := ensureApplicationsDirectory()
	if err != nil {
		return err
	}

	destPath := filepath.Join(appImageDir, filepath.Base(filePath))
	if err := utils.CopyFileSimple(filePath, destPath); err != nil {
		return fmt.Errorf("failed to copy AppImage: %w", err)
	}

	return runCommandWithSpinner(
		"Setting up AppImage",
		"Failed to make AppImage executable",
		"chmod", "+x", destPath,
	)
}

// installTarGz extracts and installs .tar.gz archive
func installTarGz(filePath, appName string) error {
	return installTarArchive(filePath, appName, "tar", "-xzf")
}

// installTarBz2 extracts and installs .tar.bz2 archive
func installTarBz2(filePath, appName string) error {
	return installTarArchive(filePath, appName, "tar", "-xjf")
}

// installTarArchive extracts tar archive and handles contents
func installTarArchive(filePath, appName, command, flags string) error {
	extractDir, err := ensureExtractDirectory(filePath, appName)
	if err != nil {
		return err
	}

	if err := runCommandWithSpinner(
		"Extracting archive",
		"Failed to extract archive",
		command, flags, filePath, "-C", extractDir,
	); err != nil {
		return err
	}

	return handleExtractedContentsLinux(extractDir, appName)
}
