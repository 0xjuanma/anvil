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
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/validators"
	"github.com/0xjuanma/palantir"
)

// showAvailableChecks displays all available checks organized by category.
func showAvailableChecks(engine *validators.DoctorEngine) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Available Health Checks")

	o.PrintInfo("🏷️  CATEGORIES (run all checks in a group):\n")

	checks := engine.ListChecks()

	categoryDescriptions := map[string]string{
		"environment":   "Verify anvil initialization and directory structure",
		"dependencies":  "Check required tools and Homebrew installation",
		"configuration": "Validate git and GitHub settings",
		"connectivity":  "Test GitHub access and repository connections",
	}

	totalChecks := 0
	for _, category := range allCategories {
		if checkNames, exists := checks[category]; exists {
			o.PrintStage(fmt.Sprintf("anvil doctor %s", category))
			o.PrintInfo("    %s", categoryDescriptions[category])
			o.PrintInfo("    Includes: %s", strings.Join(checkNames, ", "))
			o.PrintInfo("    (%d checks)\n", len(checkNames))
			totalChecks += len(checkNames)
		}
	}

	o.PrintInfo("🔍 SPECIFIC CHECKS (run individual validators):\n")

	for _, category := range allCategories {
		if checkNames, exists := checks[category]; exists {
			o.PrintStage(fmt.Sprintf("%s checks:", strings.Title(category)))
			for _, checkName := range checkNames {
				o.PrintInfo("  anvil doctor %s", checkName)
			}
			fmt.Println("")
		}
	}

	o.PrintInfo("💡 USAGE EXAMPLES:\n")
	o.PrintInfo("  anvil doctor                    # Run all %d checks", totalChecks)
	o.PrintInfo("  anvil doctor environment        # Run 3 environment checks")
	o.PrintInfo("  anvil doctor git-config         # Run only git configuration check")
	o.PrintInfo("  anvil doctor --fix              # Auto-fix detected issues")
	o.PrintInfo("  anvil doctor dependencies --fix # Auto-fix dependency issues")

	return nil
}

// runSingleCheck executes a specific health check.
func runSingleCheck(engine *validators.DoctorEngine, checkName string, verbose bool) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Running Check: %s", checkName))

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	spinner := charm.NewLineSpinner(fmt.Sprintf("Executing %s check", checkName))
	spinner.Start()

	result := engine.RunCheckWithProgress(ctx, checkName, verbose)

	if result.Status == validators.PASS {
		spinner.Success(fmt.Sprintf("%s check passed", checkName))
	} else if result.Status == validators.WARN {
		spinner.Warning(fmt.Sprintf("%s check completed with warnings", checkName))
	} else {
		spinner.Error(fmt.Sprintf("%s check failed", checkName))
	}

	displayResults([]*validators.ValidationResult{result}, verbose)

	if result.Status == validators.FAIL {
		return errors.NewValidationError(constants.OpDoctor, checkName, fmt.Errorf(result.Message))
	}

	return nil
}

// runCategoryChecks executes all checks in a specific category.
func runCategoryChecks(engine *validators.DoctorEngine, category string, verbose bool) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Running %s Health Checks", strings.Title(category)))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get validators for this category to show count
	categoryValidators := engine.GetValidatorsByCategory(category)
	if len(categoryValidators) == 0 {
		o.PrintError("Category '%s' not found", category)
		return errors.NewValidationError(constants.OpDoctor, category, fmt.Errorf("category not found"))
	}

	spinner := charm.NewLineSpinner(fmt.Sprintf("Executing %d checks in %s category", len(categoryValidators), category))
	spinner.Start()

	results := engine.RunCategoryWithProgress(ctx, category, verbose)

	// Count status
	passed, warned, failed := 0, 0, 0
	for _, result := range results {
		if result.Status == validators.PASS {
			passed++
		} else if result.Status == validators.WARN {
			warned++
		} else if result.Status == validators.FAIL {
			failed++
		}
	}

	if failed > 0 {
		spinner.Error(fmt.Sprintf("%s checks completed: %d failed", category, failed))
	} else if warned > 0 {
		spinner.Warning(fmt.Sprintf("%s checks completed: %d warnings", category, warned))
	} else {
		spinner.Success(fmt.Sprintf("All %s checks passed", category))
	}

	displayResults(results, verbose)
	printSummary(results)

	// Check if any critical failures
	for _, result := range results {
		if result.Status == validators.FAIL {
			return errors.NewValidationError(constants.OpDoctor, category, fmt.Errorf("validation failures detected"))
		}
	}

	return nil
}

// checkStatus represents the status of an individual health check
type checkStatus struct {
	name    string
	status  string // "pending", "checking", "pass", "warn", "fail"
	emoji   string
	message string
}

// runAllChecks executes all available health checks.
func runAllChecks(engine *validators.DoctorEngine, verbose bool) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Running Anvil Health Check")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Get all validators and initialize status tracking
	allValidators := engine.GetAllValidators()
	totalChecks := len(allValidators)

	// Group validators by category
	categories := make(map[string][]string)
	for _, cat := range allCategories {
		categories[cat] = []string{}
	}

	for _, v := range allValidators {
		categories[v.Category()] = append(categories[v.Category()], v.Name())
	}

	// Initialize check statuses
	checkStatuses := make(map[string]*checkStatus)
	for _, v := range allValidators {
		checkStatuses[v.Name()] = &checkStatus{
			name:   v.Name(),
			status: "pending",
			emoji:  "⋯",
		}
	}

	// Run checks with a spinner
	spinner := charm.NewLineSpinner(fmt.Sprintf("Running %d health checks", totalChecks))
	spinner.Start()

	results := engine.RunAll(ctx)

	// Count results for spinner message
	passed, warned, failed := 0, 0, 0
	for _, result := range results {
		switch result.Status {
		case validators.PASS:
			passed++
		case validators.WARN:
			warned++
		case validators.FAIL:
			failed++
		}
	}

	// Update spinner based on results
	if failed > 0 {
		spinner.Error(fmt.Sprintf("Completed: %d passed, %d failed", passed, failed))
	} else if warned > 0 {
		spinner.Warning(fmt.Sprintf("Completed: %d passed, %d warnings", passed, warned))
	} else {
		spinner.Success(fmt.Sprintf("All %d checks passed!", totalChecks))
	}

	// Update statuses based on results
	for _, result := range results {
		if cs, exists := checkStatuses[result.Name]; exists {
			cs.message = result.Message
			switch result.Status {
			case validators.PASS:
				cs.status = "pass"
				cs.emoji = "✓"
			case validators.WARN:
				cs.status = "warn"
				cs.emoji = "⚠"
			case validators.FAIL:
				cs.status = "fail"
				cs.emoji = "✗"
			case validators.SKIP:
				cs.status = "skip"
				cs.emoji = "○"
			}
		}
	}

	// Print organized results by category
	fmt.Println()
	for _, category := range allCategories {
		if checkNames, exists := categories[category]; exists && len(checkNames) > 0 {
			printCategoryResults(category, checkNames, checkStatuses, results, verbose)
		}
	}

	printSummary(results)

	return nil
}
