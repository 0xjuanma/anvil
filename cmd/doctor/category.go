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

package doctor

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/validators"
	"github.com/0xjuanma/palantir"
)

// categoryStats holds statistics for a category.
type categoryStats struct {
	passed  int
	warned  int
	failed  int
	skipped int
	total   int
}

// getCategoryBreakdown returns stats broken down by category.
func getCategoryBreakdown(results []*validators.ValidationResult) map[string]categoryStats {
	breakdown := make(map[string]categoryStats)

	for _, result := range results {
		stats := breakdown[result.Category]
		stats.total++

		switch result.Status {
		case validators.PASS:
			stats.passed++
		case validators.WARN:
			stats.warned++
		case validators.FAIL:
			stats.failed++
		case validators.SKIP:
			stats.skipped++
		}

		breakdown[result.Category] = stats
	}

	return breakdown
}

func getStatusEmoji(status validators.ValidationStatus) string {
	switch status {
	case validators.PASS:
		return "✅"
	case validators.WARN:
		return "⚠️ "
	case validators.FAIL:
		return "❌"
	case validators.SKIP:
		return "⏭️"
	default:
		return "❓"
	}
}

func getCategoryStatus(passed, warned, failed, skipped int) string {
	if failed > 0 {
		return "[FAIL]"
	} else if warned > 0 {
		return "[WARN]"
	} else {
		return "[PASS]"
	}
}

// displayCategory shows results for a specific category.
func displayCategory(category string, results []*validators.ValidationResult, verbose bool) {
	// Count statuses
	passed, warned, failed, skipped := 0, 0, 0, 0
	for _, result := range results {
		switch result.Status {
		case validators.PASS:
			passed++
		case validators.WARN:
			warned++
		case validators.FAIL:
			failed++
		case validators.SKIP:
			skipped++
		}
	}

	// Choose category status
	categoryStatus := getCategoryStatus(passed, warned, failed, skipped)
	o := palantir.GetGlobalOutputHandler()
	o.PrintStage(fmt.Sprintf("%s %s", categoryStatus, strings.Title(category)))

	for _, result := range results {
		statusEmoji := getStatusEmoji(result.Status)
		o.PrintInfo("  %s %s", statusEmoji, result.Message)

		if verbose && len(result.Details) > 0 {
			for _, detail := range result.Details {
				o.PrintInfo("      %s", detail)
			}
		}

		if result.Status != validators.PASS && result.FixHint != "" {
			o.PrintInfo("      💡 %s", result.FixHint)
		}
	}

	fmt.Println("")
}

// printCategoryResults prints results for a single category in a clean format.
func printCategoryResults(category string, checkNames []string, statuses map[string]*checkStatus, results []*validators.ValidationResult, verbose bool) {
	// Count status for this category
	passed, warned, failed := 0, 0, 0
	for _, name := range checkNames {
		if cs, exists := statuses[name]; exists {
			switch cs.status {
			case "pass":
				passed++
			case "warn":
				warned++
			case "fail":
				failed++
			}
		}
	}

	categoryStatus := getCategoryStatus(passed, warned, failed, 0)
	categoryTitle := strings.Title(category)

	// Print category header with emoji
	fmt.Printf("  %s %s\n", categoryStatus, charm.RenderHighlight(categoryTitle, "#00D9FF"))

	// Print each check result
	for _, name := range checkNames {
		if cs, exists := statuses[name]; exists {
			// Get full result for this check
			var result *validators.ValidationResult
			for _, r := range results {
				if r.Name == name {
					result = r
					break
				}
			}

			if result != nil {
				fmt.Printf("    %s %s\n", cs.emoji, result.Message)

				// Show fix hint for failed/warned checks
				if (result.Status == validators.FAIL || result.Status == validators.WARN) && result.FixHint != "" && !verbose {
					fmt.Printf("      💡 %s\n", result.FixHint)
				}

				// Show details in verbose mode
				if verbose && len(result.Details) > 0 {
					for _, detail := range result.Details {
						fmt.Printf("        %s\n", detail)
					}
				}
			}
		}
	}

	fmt.Println()
}
