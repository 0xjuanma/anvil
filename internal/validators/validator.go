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

// Package validators provides a validation framework for the Anvil CLI doctor command.
// It defines validator interfaces, result types, and a registry system for managing
// health checks across different categories (environment, dependencies, configuration, connectivity).
package validators

import (
	"context"
	"fmt"
	"sort"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/palantir"
)

// ValidationStatus represents the result status of a validation check
type ValidationStatus int

const (
	PASS ValidationStatus = iota
	WARN
	FAIL
	SKIP
)

func (vs ValidationStatus) String() string {
	switch vs {
	case PASS:
		return "PASS"
	case WARN:
		return "WARN"
	case FAIL:
		return "FAIL"
	case SKIP:
		return "SKIP"
	default:
		return "UNKNOWN"
	}
}

// ValidationResult represents the result of a validation check
type ValidationResult struct {
	Name     string           `json:"name"`
	Category string           `json:"category"`
	Status   ValidationStatus `json:"status"`
	Message  string           `json:"message"`
	Details  []string         `json:"details,omitempty"`
	FixHint  string           `json:"fix_hint,omitempty"`
	AutoFix  bool             `json:"auto_fix"`
}

// Validator interface defines the contract for all validation checks
type Validator interface {
	Name() string
	Category() string
	Description() string
	Validate(ctx context.Context, config *config.AnvilConfig) *ValidationResult
	CanFix() bool
	Fix(ctx context.Context, config *config.AnvilConfig) error
}

// ValidationRegistry manages all available validators
type ValidationRegistry struct {
	validators map[string]Validator
	categories map[string][]string
}

// NewValidationRegistry creates a new validator registry
func NewValidationRegistry() *ValidationRegistry {
	return &ValidationRegistry{
		validators: make(map[string]Validator),
		categories: make(map[string][]string),
	}
}

// Register adds a validator to the registry
func (vr *ValidationRegistry) Register(validator Validator) {
	name := validator.Name()
	category := validator.Category()

	vr.validators[name] = validator
	vr.categories[category] = append(vr.categories[category], name)
}

// GetValidator retrieves a validator by name
func (vr *ValidationRegistry) GetValidator(name string) (Validator, bool) {
	validator, exists := vr.validators[name]
	return validator, exists
}

// GetValidatorsByCategory retrieves all validators in a category
func (vr *ValidationRegistry) GetValidatorsByCategory(category string) []Validator {
	var validators []Validator
	if names, exists := vr.categories[category]; exists {
		for _, name := range names {
			if validator, ok := vr.validators[name]; ok {
				validators = append(validators, validator)
			}
		}
	}
	return validators
}

// GetAllValidators returns all registered validators
func (vr *ValidationRegistry) GetAllValidators() []Validator {
	var validators []Validator
	for _, validator := range vr.validators {
		validators = append(validators, validator)
	}
	return validators
}

// GetCategories returns all available categories
func (vr *ValidationRegistry) GetCategories() []string {
	var categories []string
	for category := range vr.categories {
		categories = append(categories, category)
	}
	sort.Strings(categories)
	return categories
}

// ListChecks returns a map of categories to validator names
func (vr *ValidationRegistry) ListChecks() map[string][]string {
	result := make(map[string][]string)
	for category, names := range vr.categories {
		sorted := make([]string, len(names))
		copy(sorted, names)
		sort.Strings(sorted)
		result[category] = sorted
	}
	return result
}

// DoctorEngine manages the validation process
type DoctorEngine struct {
	registry *ValidationRegistry
	output   palantir.OutputHandler
}

// NewDoctorEngine creates a new doctor engine
func NewDoctorEngine(output palantir.OutputHandler) *DoctorEngine {
	engine := &DoctorEngine{
		registry: NewValidationRegistry(),
		output:   output,
	}

	// Register all validators
	engine.registerDefaultValidators()

	return engine
}

// RunAll executes all registered validators
func (d *DoctorEngine) RunAll(ctx context.Context) []*ValidationResult {
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
	return d.runValidators(ctx, config, validators)
}

// RunCategory executes validators in a specific category
func (d *DoctorEngine) RunCategory(ctx context.Context, category string) []*ValidationResult {
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

	return d.runValidators(ctx, config, validators)
}

// RunCheck executes a specific validator
func (d *DoctorEngine) RunCheck(ctx context.Context, checkName string) *ValidationResult {
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

	return validator.Validate(ctx, config)
}

// FixCheck attempts to fix a specific validation issue
func (d *DoctorEngine) FixCheck(ctx context.Context, checkName string) error {
	config, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	validator, exists := d.registry.GetValidator(checkName)
	if !exists {
		return fmt.Errorf("check '%s' not found", checkName)
	}

	if !validator.CanFix() {
		return fmt.Errorf("check '%s' cannot be automatically fixed", checkName)
	}

	return validator.Fix(ctx, config)
}

// ListChecks returns available categories and checks
func (d *DoctorEngine) ListChecks() map[string][]string {
	return d.registry.ListChecks()
}

// runValidators executes a list of validators and returns results
func (d *DoctorEngine) runValidators(ctx context.Context, config *config.AnvilConfig, validators []Validator) []*ValidationResult {
	var results []*ValidationResult

	for _, validator := range validators {
		result := validator.Validate(ctx, config)
		results = append(results, result)
	}

	return results
}

// registerDefaultValidators registers all built-in validators
func (d *DoctorEngine) registerDefaultValidators() {
	// Environment validators
	d.registry.Register(&InitRunValidator{})
	d.registry.Register(&SettingsFileValidator{})
	d.registry.Register(&DirectoryStructureValidator{})

	// Dependency validators
	d.registry.Register(&BrewValidator{})
	d.registry.Register(&RequiredToolsValidator{})

	// Configuration validators
	d.registry.Register(&GitConfigValidator{})
	d.registry.Register(&GitHubConfigValidator{})
	d.registry.Register(&SyncConfigValidator{})

	// Connectivity validators
	d.registry.Register(&GitHubAccessValidator{})
	d.registry.Register(&RepositoryValidator{})
	d.registry.Register(&GitConnectivityValidator{})
}

// GetSummary creates a summary of validation results
func GetSummary(results []*ValidationResult) (passed, warned, failed, skipped int) {
	for _, result := range results {
		switch result.Status {
		case PASS:
			passed++
		case WARN:
			warned++
		case FAIL:
			failed++
		case SKIP:
			skipped++
		}
	}
	return
}

// GetFixableIssues returns results that can be automatically fixed
func GetFixableIssues(results []*ValidationResult) []*ValidationResult {
	var fixable []*ValidationResult
	for _, result := range results {
		if result.AutoFix && result.Status != PASS {
			fixable = append(fixable, result)
		}
	}
	return fixable
}

// FormatResultsTable creates a formatted table of results grouped by category
func FormatResultsTable(results []*ValidationResult) map[string][]*ValidationResult {
	categories := make(map[string][]*ValidationResult)
	for _, result := range results {
		categories[result.Category] = append(categories[result.Category], result)
	}
	return categories
}

// GetAllValidators returns all registered validators
func (d *DoctorEngine) GetAllValidators() []Validator {
	return d.registry.GetAllValidators()
}

// GetValidatorsByCategory returns validators for a specific category
func (d *DoctorEngine) GetValidatorsByCategory(category string) []Validator {
	return d.registry.GetValidatorsByCategory(category)
}

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

// getStatusEmoji returns the appropriate emoji for a validation status
func getStatusEmoji(status ValidationStatus) string {
	switch status {
	case PASS:
		return "✅"
	case WARN:
		return "⚠️"
	case FAIL:
		return "❌"
	case SKIP:
		return "⏭️ "
	default:
		return "❓"
	}
}
