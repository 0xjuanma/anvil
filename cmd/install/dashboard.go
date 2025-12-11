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

package install

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
)

// toolStatus represents the status of a tool installation.
type toolStatus struct {
	name   string
	status string // pending, installing, done, failed
	emoji  string
}

const (
	toolStatusPending    = "pending"
	toolStatusInstalling = "installing"
	toolStatusDone       = "done"
	toolStatusFailed     = "failed"
)

// printInstallDashboard displays the current installation progress.
func printInstallDashboard(groupName string, statuses []toolStatus, current, total int) {
	var content strings.Builder
	content.WriteString("\n")

	// Show each tool with its status
	for i, status := range statuses {
		var statusText string
		switch status.status {
		case toolStatusDone:
			statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, constants.StatusInstalled)
		case toolStatusFailed:
			statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, constants.StatusFailed)
		case toolStatusInstalling:
			statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, constants.StatusInstalling)
		default:
			statusText = fmt.Sprintf("%-20s %s %-15s", status.name, status.emoji, constants.StatusPending)
		}

		content.WriteString(fmt.Sprintf("  [%d/%d] %s\n", i+1, total, statusText))
	}

	content.WriteString("\n")

	// Calculate progress
	percentage := (current * 100) / total
	barWidth := 30
	filled := (percentage * barWidth) / 100
	bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)

	content.WriteString(fmt.Sprintf("  Progress: %d%% %s\n", percentage, bar))

	// Clear previous output and print new dashboard
	fmt.Print("\033[2J\033[H") // Clear screen and move cursor to top
	fmt.Println(charm.RenderBox(fmt.Sprintf("Installing '%s' group (%d tools)", groupName, total), content.String(), "#00D9FF", false))
}
