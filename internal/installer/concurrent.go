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

// Package installer provides concurrent installation capabilities for the Anvil CLI.
//
// This package implements worker pool patterns for concurrent tool installation,
// with proper progress tracking, error handling, and resource management.
package installer

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/0xjuanma/anvil/internal/brew"
	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
)

// InstallationResult represents the result of a single tool installation
type InstallationResult struct {
	ToolName  string
	Success   bool
	Error     error
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
}

// InstallationStats provides statistics about the installation process
type InstallationStats struct {
	TotalTools      int
	SuccessfulTools int
	FailedTools     int
	TotalDuration   time.Duration
	AverageDuration time.Duration
	MaxDuration     time.Duration
	MinDuration     time.Duration
	ConcurrentJobs  int
}

// ConcurrentInstaller handles concurrent tool installation
type ConcurrentInstaller struct {
	maxWorkers    int
	output        palantir.OutputHandler
	dryRun        bool
	timeout       time.Duration
	retryAttempts int
}

// NewConcurrentInstaller creates a new concurrent installer
func NewConcurrentInstaller(maxWorkers int, output palantir.OutputHandler, dryRun bool) *ConcurrentInstaller {
	if maxWorkers <= 0 {
		maxWorkers = runtime.NumCPU()
	}

	return &ConcurrentInstaller{
		maxWorkers:    maxWorkers,
		output:        output,
		dryRun:        dryRun,
		timeout:       constants.ToolInstallTimeout,
		retryAttempts: constants.DefaultRetryAttempts,
	}
}

// InstallTools installs multiple tools concurrently
func (ci *ConcurrentInstaller) InstallTools(ctx context.Context, tools []string) (*InstallationStats, error) {
	if len(tools) == 0 {
		return nil, fmt.Errorf("no tools provided for installation")
	}

	startTime := time.Now()
	ci.output.PrintHeader(fmt.Sprintf("Installing %d tools concurrently (max %d workers)", len(tools), ci.maxWorkers))

	// Create channels for work distribution
	toolChan := make(chan string, len(tools))
	resultChan := make(chan InstallationResult, len(tools))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < ci.maxWorkers; i++ {
		wg.Add(1)
		go ci.worker(ctx, i+1, toolChan, resultChan, &wg)
	}

	// Send tools to workers
	for _, tool := range tools {
		toolChan <- tool
	}
	close(toolChan)

	// Wait for all workers to complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	results := make([]InstallationResult, 0, len(tools))
	for result := range resultChan {
		results = append(results, result)
		ci.printProgress(result, len(results), len(tools))
	}

	// Calculate statistics
	stats := ci.calculateStats(results, startTime)

	// Print summary
	ci.printSummary(stats, results)

	// Return error if any installations failed
	if stats.FailedTools > 0 {
		return stats, errors.NewInstallationError(constants.OpInstall, "concurrent",
			fmt.Errorf("failed to install %d of %d tools", stats.FailedTools, stats.TotalTools))
	}

	return stats, nil
}

// worker processes tools from the channel
func (ci *ConcurrentInstaller) worker(ctx context.Context, workerID int, toolChan <-chan string, resultChan chan<- InstallationResult, wg *sync.WaitGroup) {
	defer wg.Done()

	for tool := range toolChan {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			resultChan <- InstallationResult{
				ToolName:  tool,
				Success:   false,
				Error:     ctx.Err(),
				StartTime: time.Now(),
				EndTime:   time.Now(),
			}
			return
		default:
		}

		// Install the tool with timeout
		result := ci.installWithTimeout(ctx, tool, workerID)
		resultChan <- result
	}
}

// installWithTimeout installs a single tool with timeout and retry logic
func (ci *ConcurrentInstaller) installWithTimeout(ctx context.Context, tool string, workerID int) InstallationResult {
	startTime := time.Now()

	// Create context with timeout
	toolCtx, cancel := context.WithTimeout(ctx, ci.timeout)
	defer cancel()

	var lastErr error

	// Retry logic
	for attempt := 0; attempt <= ci.retryAttempts; attempt++ {
		if attempt > 0 {
			ci.output.PrintInfo("Worker %d: Retrying %s (attempt %d/%d)", workerID, tool, attempt+1, ci.retryAttempts+1)
			time.Sleep(time.Second * time.Duration(attempt)) // Exponential backoff
		}

		// Use unified availability checking logic (ensures consistency with other installation methods)
		if brew.IsApplicationAvailable(tool) {
			ci.output.PrintAlreadyAvailable("Worker %d: %s is already available", workerID, tool)
			return InstallationResult{
				ToolName:  tool,
				Success:   true,
				StartTime: startTime,
				EndTime:   time.Now(),
				Duration:  time.Since(startTime),
			}
		}

		// Handle dry-run consistently with other installation methods
		if ci.dryRun {
			ci.output.PrintInfo("Worker %d: Would install %s", workerID, tool)
			return InstallationResult{
				ToolName:  tool,
				Success:   true,
				StartTime: startTime,
				EndTime:   time.Now(),
				Duration:  time.Since(startTime),
			}
		}

		// Install the tool
		err := ci.installSingleTool(toolCtx, tool, workerID)
		if err == nil {
			endTime := time.Now()
			ci.output.PrintSuccess(fmt.Sprintf("Worker %d: %s installed successfully", workerID, tool))
			return InstallationResult{
				ToolName:  tool,
				Success:   true,
				StartTime: startTime,
				EndTime:   endTime,
				Duration:  endTime.Sub(startTime),
			}
		}

		lastErr = err

		// Check if context was cancelled
		select {
		case <-toolCtx.Done():
			return InstallationResult{
				ToolName:  tool,
				Success:   false,
				Error:     fmt.Errorf("timeout installing %s after %v", tool, ci.timeout),
				StartTime: startTime,
				EndTime:   time.Now(),
				Duration:  time.Since(startTime),
			}
		default:
		}
	}

	// All retries failed
	return InstallationResult{
		ToolName:  tool,
		Success:   false,
		Error:     fmt.Errorf("failed to install %s after %d attempts: %w", tool, ci.retryAttempts+1, lastErr),
		StartTime: startTime,
		EndTime:   time.Now(),
		Duration:  time.Since(startTime),
	}
}

// installSingleTool installs a single tool (similar to the original logic)
func (ci *ConcurrentInstaller) installSingleTool(ctx context.Context, tool string, workerID int) error {
	// Check if source is configured for this app (user explicitly configured it)
	sourceURL, exists, sourceErr := SourceURL(tool)
	if sourceErr != nil {
		ci.output.PrintWarning("Worker %d: Failed to check source URL for %s: %v", workerID, tool, sourceErr)
		// Fall back to brew if we can't check source
		return brew.InstallPackageDirectly(tool)
	}

	// If source exists, try it first (user explicitly configured it)
	if exists && sourceURL != "" {
		ci.output.PrintInfo("Worker %d: Installing %s from configured source", workerID, tool)
		if err := InstallFromSource(tool, sourceURL); err != nil {
			// Check if extraction succeeded but installation failed
			if _, ok := err.(*ExtractionSucceededError); ok {
				// Extraction succeeded, don't fall back to brew
				// User message already shown in InstallFromSource
				return nil
			}
			// Source installation failed, fall back to brew
			ci.output.PrintInfo("Worker %d: Source installation failed, falling back to brew for %s", workerID, tool)
			return brew.InstallPackageDirectly(tool)
		}
		// Source installation succeeded, continue with post-install steps
	} else {
		// No source configured, use brew (default for majority of apps)
		if err := brew.InstallPackageDirectly(tool); err != nil {
			return errors.NewInstallationError(constants.OpInstall, tool, err)
		}
	}

	// Handle special cases for specific tools
	if tool == "zsh" {
		spinner := charm.NewLineSpinner(fmt.Sprintf("Worker %d: Installing Oh My Zsh", workerID))
		spinner.Start()
		ohMyZshScript := `sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended`
		if err := ci.printPostInstallInstructions(ohMyZshScript); err != nil {
			spinner.Warning(fmt.Sprintf("Worker %d: Oh My Zsh setup skipped", workerID))
		} else {
			spinner.Success(fmt.Sprintf("Worker %d: Oh My Zsh installed", workerID))
		}
	}

	// Handle config check for git
	if tool == "git" {
		if err := ci.checkToolConfiguration(tool); err != nil {
			ci.output.PrintWarning("Worker %d: Configuration check failed for %s: %v", workerID, tool, err)
		}
	}

	return nil
}

// printPostInstallInstructions prints post-install instructions for a tool
func (ci *ConcurrentInstaller) printPostInstallInstructions(script string) error {
	ci.output.PrintInfo("To complete setup, run:")
	ci.output.PrintInfo("  %s", script)
	return nil
}

// checkToolConfiguration checks if a tool is properly configured
func (ci *ConcurrentInstaller) checkToolConfiguration(toolName string) error {
	switch toolName {
	case constants.PkgGit:
		return ci.checkGitConfiguration()
	default:
		return nil
	}
}

// checkGitConfiguration checks if git is properly configured
func (ci *ConcurrentInstaller) checkGitConfiguration() error {
	config, err := config.LoadConfig()
	if err == nil && (config.Git.Username == "" || config.Git.Email == "") {
		ci.output.PrintInfo("Git installed successfully")
		ci.output.PrintWarning("Consider configuring git with:")
		ci.output.PrintInfo("  git config --global user.name 'Your Name'")
		ci.output.PrintInfo("  git config --global user.email 'your.email@example.com'")
	}
	return nil
}

// printProgress prints installation progress
func (ci *ConcurrentInstaller) printProgress(result InstallationResult, completed, total int) {
	status := "✓"
	if !result.Success {
		status = "✗"
	}

	ci.output.PrintProgress(completed, total,
		fmt.Sprintf("%s %s (%v)", status, result.ToolName, result.Duration.Round(time.Millisecond)))
}

// calculateStats calculates installation statistics
func (ci *ConcurrentInstaller) calculateStats(results []InstallationResult, startTime time.Time) *InstallationStats {
	stats := &InstallationStats{
		TotalTools:     len(results),
		TotalDuration:  time.Since(startTime),
		ConcurrentJobs: ci.maxWorkers,
	}

	var durations []time.Duration
	for _, result := range results {
		durations = append(durations, result.Duration)

		if result.Success {
			stats.SuccessfulTools++
		} else {
			stats.FailedTools++
		}
	}

	if len(durations) > 0 {
		// Calculate average duration
		var totalDuration time.Duration
		for _, d := range durations {
			totalDuration += d
		}
		stats.AverageDuration = totalDuration / time.Duration(len(durations))

		// Find min and max durations
		stats.MinDuration = durations[0]
		stats.MaxDuration = durations[0]
		for _, d := range durations {
			if d < stats.MinDuration {
				stats.MinDuration = d
			}
			if d > stats.MaxDuration {
				stats.MaxDuration = d
			}
		}
	}

	return stats
}

// printSummary prints installation summary
func (ci *ConcurrentInstaller) printSummary(stats *InstallationStats, results []InstallationResult) {
	ci.output.PrintHeader("Concurrent Installation Complete")
	ci.output.PrintInfo("Successfully installed %d of %d tools", stats.SuccessfulTools, stats.TotalTools)
	ci.output.PrintInfo("Total time: %v (avg: %v per tool)",
		stats.TotalDuration.Round(time.Millisecond),
		stats.AverageDuration.Round(time.Millisecond))
	ci.output.PrintInfo("Used %d concurrent workers", stats.ConcurrentJobs)

	if stats.FailedTools > 0 {
		ci.output.PrintWarning("Failed installations:")
		for _, result := range results {
			if !result.Success {
				ci.output.PrintError("  • %s: %v", result.ToolName, result.Error)
			}
		}
	}

	// Performance comparison estimate
	if stats.TotalTools > 1 {
		estimatedSerialTime := stats.AverageDuration * time.Duration(stats.TotalTools)
		speedup := float64(estimatedSerialTime) / float64(stats.TotalDuration)
		ci.output.PrintSuccess(fmt.Sprintf("Estimated speedup: %.1fx faster than serial installation", speedup))
	}
}

// SetTimeout sets the timeout for individual tool installations
func (ci *ConcurrentInstaller) SetTimeout(timeout time.Duration) {
	ci.timeout = timeout
}

// SetRetryAttempts sets the number of retry attempts for failed installations
func (ci *ConcurrentInstaller) SetRetryAttempts(attempts int) {
	ci.retryAttempts = attempts
}
