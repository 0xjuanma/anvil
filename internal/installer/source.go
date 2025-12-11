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
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/utils"
)

// InstallFromSource installs an application from a source URL or command
func InstallFromSource(appName, source string) error {
	// Check if source is a shell command (curl/wget style) or a URL
	if isShellCommand(source) {
		return installFromCommand(appName, source)
	}
	return installFromURL(appName, source)
}

// isShellCommand checks if the source is a shell command rather than a URL
func isShellCommand(source string) bool {
	trimmed := strings.TrimSpace(source)
	// Check for common shell command patterns
	return strings.HasPrefix(trimmed, "sh -c") ||
		strings.HasPrefix(trimmed, "bash -c") ||
		strings.HasPrefix(trimmed, "curl") ||
		strings.HasPrefix(trimmed, "wget") ||
		strings.Contains(trimmed, "$(curl") ||
		strings.Contains(trimmed, "$(wget")
}

// installFromCommand executes a shell command to install an application
func installFromCommand(appName, command string) error {
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Installing %s from command", appName))
	spinner.Start()

	cmd, err := parseShellCommand(command)
	if err != nil {
		spinner.Error(fmt.Sprintf("Invalid command for %s", appName))
		return fmt.Errorf("invalid command: %w", err)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		spinner.Error(fmt.Sprintf("Failed to install %s", appName))
		return fmt.Errorf("command execution failed: %w", err)
	}

	spinner.Success(fmt.Sprintf("%s installed successfully", appName))
	return nil
}

// parseShellCommand parses a shell command string into an exec.Cmd
func parseShellCommand(command string) (*exec.Cmd, error) {
	trimmed := strings.TrimSpace(command)

	// Handle sh -c or bash -c commands
	if strings.HasPrefix(trimmed, "sh -c") || strings.HasPrefix(trimmed, "bash -c") {
		shell := "sh"
		if strings.HasPrefix(trimmed, "bash") {
			shell = "bash"
		}
		cmdStr := extractCommandFromShC(trimmed)
		return exec.Command(shell, "-c", cmdStr), nil
	}

	// Direct command execution
	parts := strings.Fields(trimmed)
	if len(parts) == 0 {
		return nil, fmt.Errorf("empty command")
	}
	return exec.Command(parts[0], parts[1:]...), nil
}

// extractCommandFromShC extracts the command string from "sh -c 'command'" format
func extractCommandFromShC(fullCommand string) string {
	// Find "sh -c" or "bash -c" and extract everything after it
	trimmed := strings.TrimSpace(fullCommand)

	// Look for the pattern: "sh -c" or "bash -c" followed by quoted string
	shPattern := "sh -c"
	bashPattern := "bash -c"

	var startIdx int
	if strings.HasPrefix(trimmed, shPattern) {
		startIdx = len(shPattern)
	} else if strings.HasPrefix(trimmed, bashPattern) {
		startIdx = len(bashPattern)
	} else {
		return fullCommand
	}

	// Skip whitespace after "sh -c" or "bash -c"
	remaining := strings.TrimSpace(trimmed[startIdx:])

	// Handle quoted strings - remove outer quotes if present
	remaining = strings.TrimSpace(remaining)
	if len(remaining) > 0 {
		// Check if it starts and ends with matching quotes
		if (remaining[0] == '"' && remaining[len(remaining)-1] == '"') ||
			(remaining[0] == '\'' && remaining[len(remaining)-1] == '\'') {
			remaining = remaining[1 : len(remaining)-1]
		}
	}

	return remaining
}

// installFromURL installs an application from a URL
func installFromURL(appName, sourceURL string) error {
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Downloading %s from source", appName))
	spinner.Start()

	downloadedFile, err := downloadFile(sourceURL, appName)
	if err != nil {
		spinner.Error(fmt.Sprintf("Failed to download %s", appName))
		return fmt.Errorf("failed to download %s: %w", appName, err)
	}
	spinner.Success(fmt.Sprintf("Downloaded %s", appName))

	spinner = charm.NewDotsSpinner(fmt.Sprintf("Installing %s", appName))
	spinner.Start()

	if err := installDownloadedFile(downloadedFile, appName); err != nil {
		spinner.Error(fmt.Sprintf("Failed to install %s", appName))
		return fmt.Errorf("failed to install %s: %w", appName, err)
	}

	spinner.Success(fmt.Sprintf("%s installed successfully", appName))
	return nil
}

// downloadFile downloads a file from URL to a temporary location
func downloadFile(fileURL, appName string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", fileURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "anvil-cli/1.0")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP error %d: %s", resp.StatusCode, resp.Status)
	}

	homeDir, _ := system.HomeDir()
	downloadsDir := filepath.Join(homeDir, "Downloads", "anvil-downloads")
	if err := utils.EnsureDirectory(downloadsDir); err != nil {
		return "", fmt.Errorf("failed to create downloads directory: %w", err)
	}

	fileName := getFileNameFromURL(fileURL, appName)
	filePath := filepath.Join(downloadsDir, fileName)

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	if _, err := io.Copy(file, resp.Body); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filePath, nil
}

// getFileNameFromURL extracts filename from URL or uses app name
func getFileNameFromURL(fileURL, appName string) string {
	parsedURL, err := url.Parse(fileURL)
	if err == nil && parsedURL.Path != "" {
		fileName := filepath.Base(parsedURL.Path)
		if fileName != "" && fileName != "/" {
			return fileName
		}
	}

	ext := getExtensionFromURL(fileURL)
	return fmt.Sprintf("%s%s", appName, ext)
}

// getExtensionFromURL tries to detect file extension from URL
func getExtensionFromURL(fileURL string) string {
	parsedURL, err := url.Parse(fileURL)
	if err == nil {
		path := strings.ToLower(parsedURL.Path)
		extensions := []string{".dmg", ".pkg", ".zip", ".tar.gz", ".deb", ".rpm", ".AppImage", ".tar.bz2"}
		for _, ext := range extensions {
			if strings.HasSuffix(path, ext) {
				return ext
			}
		}
	}
	return ".zip"
}

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
	case ".dmg":
		return installDMG(filePath, appName)
	case ".pkg":
		return installPKG(filePath)
	case ".zip":
		return installZIP(filePath, appName)
	default:
		return fmt.Errorf("unsupported file type: %s (supported: .dmg, .pkg, .zip)", ext)
	}
}

// installOnLinux handles installation on Linux
func installOnLinux(filePath, appName string) error {
	ext := strings.ToLower(filepath.Ext(filePath))
	baseName := strings.ToLower(filepath.Base(filePath))

	if strings.HasSuffix(baseName, ".tar.gz") {
		return installTarGz(filePath, appName)
	} else if strings.HasSuffix(baseName, ".tar.bz2") {
		return installTarBz2(filePath, appName)
	}

	switch ext {
	case ".deb":
		return installDEB(filePath)
	case ".rpm":
		return installRPM(filePath)
	case ".AppImage":
		return installAppImage(filePath, appName)
	case ".zip":
		return installZIP(filePath, appName)
	default:
		return fmt.Errorf("unsupported file type: %s (supported: .deb, .rpm, .AppImage, .zip, .tar.gz, .tar.bz2)", ext)
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

// handleExtractedContentsMacOS handles extracted contents on macOS
func handleExtractedContentsMacOS(extractDir, appName string) error {
	appPath := findAppInDirectory(extractDir, appName)
	if appPath == "" {
		return fmt.Errorf("failed to find .app in extracted contents")
	}

	applicationsDir, err := ensureApplicationsDirectory()
	if err != nil {
		return err
	}

	appNameFromPath := filepath.Base(appPath)
	destPath := filepath.Join(applicationsDir, appNameFromPath)

	return copyAppToApplications(appPath, destPath)
}

// handleExtractedContentsLinux handles extracted contents on Linux
func handleExtractedContentsLinux(extractDir, appName string) error {
	entries, err := os.ReadDir(extractDir)
	if err != nil {
		return fmt.Errorf("failed to read extract directory: %w", err)
	}

	if len(entries) == 1 && entries[0].IsDir() {
		appDir := filepath.Join(extractDir, entries[0].Name())
		destDir, err := ensureLinuxApplicationsDirectory(entries[0].Name())
		if err != nil {
			return err
		}
		return utils.CopyDirectorySimple(appDir, destDir)
	}

	destDir, err := ensureLinuxApplicationsDirectory(appName)
	if err != nil {
		return err
	}

	return utils.CopyDirectorySimple(extractDir, destDir)
}

// extractMountPath extracts the mount path from hdiutil output
func extractMountPath(output string) string {
	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.Contains(line, "/Volumes/") {
			parts := strings.Fields(line)
			for _, part := range parts {
				if strings.HasPrefix(part, "/Volumes/") {
					return part
				}
			}
		}
	}
	return ""
}

// findAppInDirectory searches for .app bundle in directory
func findAppInDirectory(dir, appName string) string {
	var foundApp string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() && strings.HasSuffix(path, ".app") {
			baseName := strings.ToLower(strings.TrimSuffix(filepath.Base(path), ".app"))
			searchName := strings.ToLower(appName)
			if baseName == searchName || strings.Contains(baseName, searchName) {
				foundApp = path
				return filepath.SkipDir
			}
		}
		return nil
	})
	return foundApp
}

// SourceURL returns the source URL for an app if it exists
func SourceURL(appName string) (string, bool, error) {
	cfg, err := config.LoadConfig()
	if err != nil {
		return "", false, fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Sources == nil {
		return "", false, nil
	}

	sourceURL, exists := cfg.Sources[appName]
	if !exists || sourceURL == "" {
		return "", false, nil
	}

	return sourceURL, true, nil
}
