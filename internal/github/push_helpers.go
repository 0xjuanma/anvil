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

package github

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// generateTimestampedBranchName generates a branch name with current date and time
func generateTimestampedBranchName(prefix string) string {
	now := time.Now()
	dateStr := now.Format("02012006") // DDMMYYYY
	timeStr := now.Format("1504")     // HHMM (24h format)
	return fmt.Sprintf("%s-%s-%s", prefix, dateStr, timeStr)
}

// getRepositoryURL returns the GitHub repository URL for display
func (gc *GitHubClient) getRepositoryURL() string {
	if strings.Contains(gc.RepoURL, "://") {
		return gc.RepoURL
	}
	return fmt.Sprintf("https://github.com/%s", gc.RepoURL)
}

// isSingleSmallFile determines if we should include the full diff output
func (gc *GitHubClient) isSingleSmallFile(statOutput string) bool {
	// Only get full diff for single files with reasonable size
	return strings.Contains(statOutput, "1 file changed") &&
		strings.Count(statOutput, "+")+strings.Count(statOutput, "-") <= 50
}

// extractFileCount parses the file count from Git's stat output
func (gc *GitHubClient) extractFileCount(statOutput string) int {
	if strings.TrimSpace(statOutput) == "" {
		return 0
	}

	// Parse "1 file changed" or "2 files changed"
	if strings.Contains(statOutput, "1 file changed") {
		return 1
	}

	// Use regex to extract number from "X files changed"
	re := regexp.MustCompile(`(\d+) files changed`)
	matches := re.FindStringSubmatch(statOutput)
	if len(matches) >= 2 {
		if count, err := strconv.Atoi(matches[1]); err == nil {
			return count
		}
	}

	return 0
}
