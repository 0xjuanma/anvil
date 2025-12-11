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

package config

import "fmt"

// GroupTools returns the tools for a specific group
func GroupTools(groupName string) ([]string, error) {
	var result []string
	err := withConfig(func(config *AnvilConfig) error {
		// Check if the group exists in the Groups map
		if tools, exists := config.Groups[groupName]; exists {
			result = tools
			return nil
		}
		return fmt.Errorf("group '%s' not found", groupName)
	})
	return result, err
}

// AvailableGroups returns all available groups
func AvailableGroups() (map[string][]string, error) {
	var groups map[string][]string
	err := withConfig(func(config *AnvilConfig) error {
		groups = make(map[string][]string)
		// Add built-in groups
		for name, tools := range config.Groups {
			groups[name] = tools
		}
		return nil
	})
	return groups, err
}

// BuiltInGroups returns the list of built-in group names
func BuiltInGroups() []string {
	return builtInGroups
}

// IsBuiltInGroup checks if a group name is a built-in group
func IsBuiltInGroup(groupName string) bool {
	for _, group := range builtInGroups {
		if group == groupName {
			return true
		}
	}
	return false
}

// AddCustomGroup adds a new custom group
func AddCustomGroup(name string, tools []string) error {
	return withConfigAndSave(func(config *AnvilConfig) error {
		ensureMap(&config.Groups)
		config.Groups[name] = tools
		return nil
	})
}

// UpdateGroupTools updates the tools list for an existing group
func UpdateGroupTools(groupName string, tools []string) error {
	return withConfigAndSave(func(config *AnvilConfig) error {
		// Check if the group exists
		if _, exists := config.Groups[groupName]; !exists {
			return fmt.Errorf("group '%s' does not exist", groupName)
		}
		// Update the group with new tools list
		config.Groups[groupName] = tools
		return nil
	})
}

// AddAppToGroup adds an app to a group, creating the group if it doesn't exist
func AddAppToGroup(groupName string, appName string) error {
	return AddAppsToGroup(groupName, []string{appName})
}

// AddAppsToGroup adds multiple apps to a group in a single operation
func AddAppsToGroup(groupName string, apps []string) error {
	return withConfigAndSave(func(config *AnvilConfig) error {
		// Initialize map if nil (more idiomatic/performant than reflection-based ensureMap)
		if config.Groups == nil {
			config.Groups = make(map[string][]string)
		}

		tools := config.Groups[groupName]

		// Use a set to track existing tools for O(1) lookups and deduplication
		// map[string]struct{} is idiomatic for sets (0 bytes per value)
		existingSet := make(map[string]struct{}, len(tools))
		for _, tool := range tools {
			existingSet[tool] = struct{}{}
		}

		for _, app := range apps {
			if _, exists := existingSet[app]; !exists {
				tools = append(tools, app)
				existingSet[app] = struct{}{}
			}
		}

		config.Groups[groupName] = tools
		return nil
	})
}
