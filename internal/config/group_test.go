package config

import (
	"testing"
)

func TestAddAppToGroup_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		initialGroups map[string][]string
		groupName     string
		appName       string
		expectedTools []string
		expectError   bool
	}{
		{
			name:          "Add app to new group",
			initialGroups: map[string][]string{},
			groupName:     "new-group",
			appName:       "app1",
			expectedTools: []string{"app1"},
			expectError:   false,
		},
		{
			name: "Add app to existing group",
			initialGroups: map[string][]string{
				"existing-group": {"app1"},
			},
			groupName:     "existing-group",
			appName:       "app2",
			expectedTools: []string{"app1", "app2"},
			expectError:   false,
		},
		{
			name: "Add duplicate app to existing group",
			initialGroups: map[string][]string{
				"existing-group": {"app1"},
			},
			groupName:     "existing-group",
			appName:       "app1",
			expectedTools: []string{"app1"},
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupTestConfig(t)
			defer cleanup()

			// Setup initial state
			if len(tt.initialGroups) > 0 {
				for gName, tools := range tt.initialGroups {
					if err := AddCustomGroup(gName, tools); err != nil {
						t.Fatalf("Failed to setup initial group %s: %v", gName, err)
					}
				}
			}

			// Execute
			err := AddAppToGroup(tt.groupName, tt.appName)

			// Assert Error
			if (err != nil) != tt.expectError {
				t.Errorf("AddAppToGroup() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Assert State
			config, err := LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			tools, exists := config.Groups[tt.groupName]
			if !exists {
				t.Errorf("Group %s not created", tt.groupName)
				return
			}

			if len(tools) != len(tt.expectedTools) {
				t.Errorf("Expected %d tools, got %d", len(tt.expectedTools), len(tools))
			}

			// Verify contents (order matters for append, but sets are unordered in logic - slice append order is preserved)
			for i, tool := range tools {
				if tool != tt.expectedTools[i] {
					t.Errorf("Expected tool at index %d to be %s, got %s", i, tt.expectedTools[i], tool)
				}
			}
		})
	}
}

func TestAddAppsToGroup_TableDriven(t *testing.T) {
	tests := []struct {
		name          string
		initialGroups map[string][]string
		groupName     string
		appsToAdd     []string
		expectedTools []string
		expectError   bool
	}{
		{
			name:          "Add multiple apps to new group",
			initialGroups: map[string][]string{},
			groupName:     "batch-group",
			appsToAdd:     []string{"app1", "app2"},
			expectedTools: []string{"app1", "app2"},
			expectError:   false,
		},
		{
			name: "Add multiple apps to existing group",
			initialGroups: map[string][]string{
				"existing-batch": {"app1"},
			},
			groupName:     "existing-batch",
			appsToAdd:     []string{"app2", "app3"},
			expectedTools: []string{"app1", "app2", "app3"},
			expectError:   false,
		},
		{
			name: "Add mixed new and duplicate apps",
			initialGroups: map[string][]string{
				"mixed-batch": {"app1", "app2"},
			},
			groupName:     "mixed-batch",
			appsToAdd:     []string{"app2", "app3", "app1"}, // app2 and app1 are dupes
			expectedTools: []string{"app1", "app2", "app3"},
			expectError:   false,
		},
		{
			name:          "Add empty list",
			initialGroups: map[string][]string{},
			groupName:     "empty-batch",
			appsToAdd:     []string{},
			expectedTools: []string{}, // Or nil/empty
			expectError:   false,
		},
		{
			name:          "Add empty string app",
			initialGroups: map[string][]string{},
			groupName:     "empty-string-group",
			appsToAdd:     []string{""},
			expectedTools: []string{""},
			expectError:   false,
		},
		{
			name: "Complex mixed order",
			initialGroups: map[string][]string{
				"mixed-order": {"a", "b"},
			},
			groupName:     "mixed-order",
			appsToAdd:     []string{"b", "c", "a", "d"},
			expectedTools: []string{"a", "b", "c", "d"}, // order preserved for existing (a,b) then new appended (c,d). b and a skipped.
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cleanup := setupTestConfig(t)
			defer cleanup()

			// Setup initial state
			if len(tt.initialGroups) > 0 {
				for gName, tools := range tt.initialGroups {
					if err := AddCustomGroup(gName, tools); err != nil {
						t.Fatalf("Failed to setup initial group %s: %v", gName, err)
					}
				}
			}

			// Execute
			err := AddAppsToGroup(tt.groupName, tt.appsToAdd)

			// Assert Error
			if (err != nil) != tt.expectError {
				t.Errorf("AddAppsToGroup() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Assert State
			config, err := LoadConfig()
			if err != nil {
				t.Fatalf("Failed to load config: %v", err)
			}

			tools, exists := config.Groups[tt.groupName]
			if !exists {
				// Special case: if adding empty list to new group, it creates the entry with empty/nil list
				// Check if we expected that
				if len(tt.expectedTools) == 0 && len(tt.appsToAdd) == 0 {
					// It's acceptable if key exists or not depending on implementation details of empty append
					// Our implementation creates the key: config.Groups[groupName] = tools
					t.Errorf("Group %s not created (even empty list adds key)", tt.groupName)
				}
				return
			}

			if len(tools) != len(tt.expectedTools) {
				t.Errorf("Expected %d tools, got %d: %v vs %v", len(tt.expectedTools), len(tools), tt.expectedTools, tools)
			}

			// Verify contents. Since map iteration order in `AddAppsToGroup` (for `existingSet`) is random BUT
			// `tools` comes from `config.Groups` which is a slice, order is preserved for existing items.
			// New items are appended in order of `apps` input loop.
			// So order is deterministic: [existing..., new...]
			for i, tool := range tools {
				if tool != tt.expectedTools[i] {
					t.Errorf("Expected tool at index %d to be %s, got %s", i, tt.expectedTools[i], tool)
				}
			}
		})
	}
}
