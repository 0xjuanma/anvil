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
	"time"

	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/validators"
	"github.com/0xjuanma/palantir"
)

// runFixCheck attempts to fix a specific check.
func runFixCheck(engine *validators.DoctorEngine, checkName string) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader(fmt.Sprintf("Fixing Check: %s", checkName))

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// First run the check to see current status
	result := engine.RunCheck(ctx, checkName)
	displayResults([]*validators.ValidationResult{result}, false)

	if result.Status == validators.PASS && checkName != "git-config" {
		o.PrintSuccess("Check is already passing, no fix needed")
		return nil
	}

	if !result.AutoFix {
		o.PrintWarning("This check cannot be automatically fixed")
		o.PrintInfo("Manual fix required: %s", result.FixHint)
		return nil
	}

	// Confirm with user
	if !o.Confirm(fmt.Sprintf("Attempt to fix '%s'?", checkName)) {
		o.PrintInfo("Fix cancelled by user")
		return nil
	}

	// Attempt fix
	spinner := charm.NewDotsSpinner(fmt.Sprintf("Attempting to fix %s", checkName))
	spinner.Start()
	if err := engine.FixCheck(ctx, checkName); err != nil {
		spinner.Error("Fix failed")
		o.PrintError("Fix failed: %v", err)
		return err
	}

	spinner.Success("Fix completed!")

	// Verify fix
	spinner = charm.NewLineSpinner("Verifying fix")
	spinner.Start()
	newResult := engine.RunCheck(ctx, checkName)
	spinner.Success("Verification complete")
	displayResults([]*validators.ValidationResult{newResult}, false)

	if newResult.Status == validators.PASS {
		o.PrintSuccess("✅ Check is now passing!")
	} else {
		o.PrintWarning("⚠️  Check still has issues after fix attempt")
	}

	return nil
}

// runFixAll attempts to fix all auto-fixable issues.
func runFixAll(engine *validators.DoctorEngine, category string) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Auto-fixing Issues")

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Run checks to find fixable issues
	var results []*validators.ValidationResult
	if category != "" {
		results = engine.RunCategory(ctx, category)
	} else {
		results = engine.RunAll(ctx)
	}

	fixableIssues := validators.GetFixableIssues(results)
	if len(fixableIssues) == 0 {
		if category != "" {
			o.PrintSuccess(fmt.Sprintf("No auto-fixable issues found in %s category!", category))
		} else {
			o.PrintSuccess("No auto-fixable issues found!")
		}
		return nil
	}

	if category != "" {
		o.PrintInfo("Found %d auto-fixable issues in %s category:", len(fixableIssues), category)
	} else {
		o.PrintInfo("Found %d auto-fixable issues:", len(fixableIssues))
	}
	for _, issue := range fixableIssues {
		o.PrintInfo("  • %s: %s", issue.Name, issue.Message)
	}

	confirmMessage := "Attempt to fix all auto-fixable issues?"
	if category != "" {
		confirmMessage = fmt.Sprintf("Attempt to fix all auto-fixable issues in %s category?", category)
	}

	if !o.Confirm(confirmMessage) {
		o.PrintInfo("Fix cancelled by user")
		return nil
	}

	var fixedCount, failedCount int
	for _, issue := range fixableIssues {
		o.PrintInfo("Fixing %s...", issue.Name)
		if err := engine.FixCheck(ctx, issue.Name); err != nil {
			o.PrintError("Failed to fix %s: %v", issue.Name, err)
			failedCount++
		} else {
			o.PrintSuccess(fmt.Sprintf("Fixed %s", issue.Name))
			fixedCount++
		}
	}

	o.PrintInfo("Fix complete: %d succeeded, %d failed", fixedCount, failedCount)
	return nil
}
