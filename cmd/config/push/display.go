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

package push

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/github"
	"github.com/0xjuanma/palantir"
)

// showNewAppInfo displays information about new app additions.
func showNewAppInfo(appName, configPath string) {
	output := palantir.GetGlobalOutputHandler()
	fmt.Println("")
	output.PrintHeader("New App Addition")
	output.PrintInfo("App: %s", appName)
	output.PrintInfo("Local path: %s", configPath)
	fmt.Println("")
	output.PrintInfo("This app will be added to the repository for the first time.")
	output.PrintInfo("All configuration files will be committed to a new branch.")
}

// handleAppLocationError provides helpful error messages for app location resolution failures.
func handleAppLocationError(appName string, err error) error {
	if strings.Contains(err.Error(), "not found in configs or temp directory") {
		o := palantir.GetGlobalOutputHandler()
		o.PrintError("App '%s' is not known to anvil\n", appName)
		o.PrintInfo("💡 To push app configurations:\n")
		o.PrintInfo("1. Configure the app's local config path in %s:\n", constants.ANVIL_CONFIG_FILE)
		o.PrintInfo("configs:")
		o.PrintInfo("  %s: /path/to/your/%s/configs\n", appName, appName)
		o.PrintInfo("2. Or pull the app's configs first to discover it:")
		o.PrintInfo("   anvil config pull %s\n", appName)
		o.PrintInfo("3. Then configure the local path in %s\n", constants.ANVIL_CONFIG_FILE)
		o.PrintInfo("4. For completely new apps, ensure the local path exists and contains config files")
		return fmt.Errorf("app not configured")
	}

	return fmt.Errorf("failed to resolve app location: %w", err)
}

// showSecurityWarning displays a security warning about private repositories.
func showSecurityWarning(privateRepo string) {
	// 🚨 SECURITY WARNING: Remind users about private repository requirement
	o := palantir.GetGlobalOutputHandler()
	o.PrintWarning("SECURITY REMINDER: Configuration files may contain sensitive data")
	o.PrintInfo("   • Personal file paths and system information")
	o.PrintInfo("   • Private development environment details")
	o.PrintInfo("🛡️ Anvil ENFORCES private repositories for security")
	o.PrintInfo("   • Repository '%s' must be PRIVATE", privateRepo)
	o.PrintInfo("   • Public repositories will be BLOCKED\n")
}

// displaySuccessMessage displays a success message after the push operation.
func displaySuccessMessage(appName string, result *github.PushConfigResult, diffSummary *github.DiffSummary, anvilConfig *config.AnvilConfig) {
	// Display full success message for actual push
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Push Complete!")
	o.PrintSuccess(fmt.Sprintf("%s configuration push completed successfully!\n", appName))
	o.PrintInfo("Push Summary:")
	o.PrintInfo("  • Branch created: %s", result.BranchName)
	o.PrintInfo("  • Commit message: %s", result.CommitMessage)
	o.PrintInfo("  • Files committed: \n\n%s", diffSummary.GitStatOutput)
	o.PrintInfo("🔗 Repository: %s", result.RepositoryURL)
	o.PrintSuccess("You can now create a Pull Request on GitHub to merge these changes!")
	o.PrintInfo("Direct link: %s/compare/%s...%s", result.RepositoryURL, anvilConfig.GitHub.Branch, result.BranchName)
}

// showDiffOutput displays diff information using Git's native output.
func showDiffOutput(diffSummary *github.DiffSummary) {
	o := palantir.GetGlobalOutputHandler()
	if diffSummary.TotalFiles == 0 {
		o.PrintInfo("No changes detected")
		return
	}

	o.PrintHeader("Changes to be pushed:")

	// Show Git's native stat output directly
	if diffSummary.GitStatOutput != "" {
		fmt.Println("")
		o.PrintInfo(diffSummary.GitStatOutput)
	}

	// For single small files, show full diff
	if diffSummary.TotalFiles == 1 && diffSummary.FullDiff != "" {
		lines := strings.Split(diffSummary.FullDiff, "\n")
		if len(lines) <= 50 {
			o.PrintInfo("📄 Full diff:\n")
			o.PrintInfo(diffSummary.FullDiff)
		} else {
			o.PrintInfo("📄 Diff preview (first 50 lines):\n")
			o.PrintInfo(strings.Join(lines[:50], "\n"))
			o.PrintInfo("\n... [diff truncated] ...")
		}
	}
	fmt.Println("")
}
