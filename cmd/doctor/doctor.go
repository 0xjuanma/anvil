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

// Package doctor provides health check functionality for the Anvil CLI,
// validating environment, dependencies, configuration, and connectivity.
package doctor

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/validators"
	"github.com/0xjuanma/palantir"
	"github.com/spf13/cobra"
)

var allCategories = []string{"environment", "dependencies", "configuration", "connectivity"}

var DoctorCmd = &cobra.Command{
	Use:   "doctor [category|check]",
	Short: "Run health checks and validate anvil environment",
	Long:  constants.DOCTOR_COMMAND_LONG_DESCRIPTION,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runDoctorCommand(cmd, args)
	},
}

// runDoctorCommand executes the doctor validation process.
func runDoctorCommand(cmd *cobra.Command, args []string) error {
	// Get command flags
	listChecks, _ := cmd.Flags().GetBool("list")
	fix, _ := cmd.Flags().GetBool("fix")
	verbose, _ := cmd.Flags().GetBool("verbose")

	// Create doctor engine with terminal output
	engine := validators.NewDoctorEngine(palantir.GetGlobalOutputHandler())

	// Handle list command
	if listChecks {
		return showAvailableChecks(engine)
	}

	// Handle fix command
	if fix {
		if len(args) > 0 {
			return runFixCheck(engine, args[0])
		} else {
			return runFixAll(engine, "")
		}
	}

	// Handle positional arguments
	if len(args) == 0 {
		// No arguments: run all checks
		return runAllChecks(engine, verbose)
	}

	target := args[0]

	// Check if it's a category first
	for _, category := range allCategories {
		if target == category {
			return runCategoryChecks(engine, category, verbose)
		}
	}

	// Otherwise treat it as a specific check
	return runSingleCheck(engine, target, verbose)
}

// displayResults shows validation results in a formatted table.
func displayResults(results []*validators.ValidationResult, verbose bool) {
	categories := validators.FormatResultsTable(results)

	for _, category := range allCategories {
		if categoryResults, exists := categories[category]; exists {
			displayCategory(category, categoryResults, verbose)
		}
	}
}

// printSummary shows overall health check summary.
func printSummary(results []*validators.ValidationResult) {
	passed, warned, failed, _ := validators.GetSummary(results)
	total := len(results)

	// Get category breakdown
	categoryStats := getCategoryBreakdown(results)

	// Build dashboard box
	var dashboardContent strings.Builder
	dashboardContent.WriteString("\n")

	// Category status bars
	for _, category := range allCategories {
		if stats, exists := categoryStats[category]; exists {
			status := getCategoryStatus(stats.passed, stats.warned, stats.failed, stats.skipped)

			// Calculate percentage for progress bar
			percentage := 0
			if stats.total > 0 {
				percentage = (stats.passed * 100) / stats.total
			}

			// Create progress bar (20 chars wide)
			barWidth := 20
			filled := (percentage * barWidth) / 100
			bar := strings.Repeat("━", filled) + strings.Repeat("━", barWidth-filled)

			categoryTitle := strings.Title(category)
			dashboardContent.WriteString(fmt.Sprintf("  %-6s %-15s %s  %d/%d passing\n",
				status, categoryTitle, bar, stats.passed, stats.total))
		}
	}

	dashboardContent.WriteString("\n")
	dashboardContent.WriteString(fmt.Sprintf("  Overall: %d/%d checks passing\n", passed, total))
	dashboardContent.WriteString("\n")

	// Render dashboard box
	fmt.Println(charm.RenderBox("Summary", dashboardContent.String(), "#00D9FF", true))

	// Show fixable issues in a separate box
	fixableIssues := validators.GetFixableIssues(results)
	if len(fixableIssues) > 0 {
		var fixContent strings.Builder
		for _, issue := range fixableIssues {
			fixContent.WriteString(fmt.Sprintf("  • %s\n", issue.Name))
		}
		fixContent.WriteString("\n")
		fixContent.WriteString("  Run 'anvil doctor --fix' to automatically fix them\n")

		fmt.Println(charm.RenderBox("🔧 Auto-fixable Issues", fixContent.String(), "#FFD700", true))
	}

	// Overall status badge
	fmt.Println()
	if failed > 0 {
		fmt.Println("  " + charm.RenderBadge("ISSUES FOUND", "#FF5F87"))
	} else if warned > 0 {
		fmt.Println("  " + charm.RenderBadge("MINOR ISSUES", "#FFD700"))
	} else {
		fmt.Println("  " + charm.RenderBadge("HEALTHY", "#00FF87"))
	}
	fmt.Println()
}

func init() {
	// Add flags for enhanced doctor functionality
	DoctorCmd.Flags().Bool("list", false, "List all available health checks")
	DoctorCmd.Flags().Bool("fix", false, "Attempt to automatically fix issues")
	DoctorCmd.Flags().Bool("verbose", false, "Show detailed output")
}
