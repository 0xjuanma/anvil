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
	"sort"

	"github.com/0xjuanma/anvil/internal/config"
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

