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

// GetAllValidators returns all registered validators
func (d *DoctorEngine) GetAllValidators() []Validator {
	return d.registry.GetAllValidators()
}

// GetValidatorsByCategory returns validators for a specific category
func (d *DoctorEngine) GetValidatorsByCategory(category string) []Validator {
	return d.registry.GetValidatorsByCategory(category)
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
