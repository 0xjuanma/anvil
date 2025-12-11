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
