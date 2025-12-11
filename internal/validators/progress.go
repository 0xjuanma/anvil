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

package validators

import (
	"context"
	"fmt"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/palantir"
)

// RunAllWithProgress executes all registered validators with progress feedback
func (d *DoctorEngine) RunAllWithProgress(ctx context.Context, verbose bool) []*ValidationResult {
	config, err := config.LoadConfig()
	if err != nil {
		// If we can't load config, create a minimal result for critical failure
		return []*ValidationResult{{
			Name:     "config-load",
			Category: "environment",
			Status:   FAIL,
			Message:  "Failed to load configuration",
			Details:  []string{err.Error()},
			FixHint:  "Run 'anvil init' to initialize your environment",
			AutoFix:  false,
		}}
	}

	validators := d.registry.GetAllValidators()
	return d.runValidatorsWithProgress(ctx, config, validators, verbose)
}

// RunCategoryWithProgress executes validators in a specific category with progress feedback
func (d *DoctorEngine) RunCategoryWithProgress(ctx context.Context, category string, verbose bool) []*ValidationResult {
	config, err := config.LoadConfig()
	if err != nil {
		return []*ValidationResult{{
			Name:     "config-load",
			Category: category,
			Status:   FAIL,
			Message:  "Failed to load configuration",
			Details:  []string{err.Error()},
			FixHint:  "Run 'anvil init' to initialize your environment",
			AutoFix:  false,
		}}
	}

	validators := d.registry.GetValidatorsByCategory(category)
	if len(validators) == 0 {
		return []*ValidationResult{{
			Name:     "category-not-found",
			Category: category,
			Status:   FAIL,
			Message:  fmt.Sprintf("Category '%s' not found", category),
			FixHint:  "Use 'anvil doctor --list' to see available categories",
			AutoFix:  false,
		}}
	}

	return d.runValidatorsWithProgress(ctx, config, validators, verbose)
}

// RunCheckWithProgress executes a specific validator with progress feedback
func (d *DoctorEngine) RunCheckWithProgress(ctx context.Context, checkName string, verbose bool) *ValidationResult {
	config, err := config.LoadConfig()
	if err != nil {
		return &ValidationResult{
			Name:     checkName,
			Category: "unknown",
			Status:   FAIL,
			Message:  "Failed to load configuration",
			Details:  []string{err.Error()},
			FixHint:  "Run 'anvil init' to initialize your environment",
			AutoFix:  false,
		}
	}

	validator, exists := d.registry.GetValidator(checkName)
	if !exists {
		return &ValidationResult{
			Name:     checkName,
			Category: "unknown",
			Status:   FAIL,
			Message:  fmt.Sprintf("Check '%s' not found", checkName),
			FixHint:  "Use 'anvil doctor --list' to see available checks",
			AutoFix:  false,
		}
	}

	// Show progress for single check
	o := palantir.GetGlobalOutputHandler()
	o.PrintInfo("🔍 Running %s check...", validator.Name())
	if verbose {
		o.PrintInfo("   Description: %s", validator.Description())
		o.PrintInfo("   Category: %s", validator.Category())
	}

	result := validator.Validate(ctx, config)

	// Show immediate result
	statusEmoji := getStatusEmoji(result.Status)
	o.PrintInfo("%s %s", statusEmoji, result.Message)

	return result
}

// runValidatorsWithProgress executes a list of validators with progress feedback
func (d *DoctorEngine) runValidatorsWithProgress(ctx context.Context, config *config.AnvilConfig, validators []Validator, verbose bool) []*ValidationResult {
	var results []*ValidationResult
	totalValidators := len(validators)
	o := palantir.GetGlobalOutputHandler()
	for i, validator := range validators {
		// Show progress
		o.PrintProgress(i+1, totalValidators, fmt.Sprintf("Running %s", validator.Name()))

		if verbose {
			o.PrintInfo("   Description: %s", validator.Description())
			o.PrintInfo("   Category: %s", validator.Category())
		}

		result := validator.Validate(ctx, config)
		results = append(results, result)

		// Show immediate result
		statusEmoji := getStatusEmoji(result.Status)
		if verbose {
			o.PrintInfo("   Result: %s %s", statusEmoji, result.Message)
			if len(result.Details) > 0 {
				for _, detail := range result.Details {
					o.PrintInfo("      %s", detail)
				}
			}
		} else {
			o.PrintInfo("   %s %s", statusEmoji, result.Message)
		}
	}

	fmt.Println("")
	o.PrintSuccess("All validation checks completed")

	return results
}
