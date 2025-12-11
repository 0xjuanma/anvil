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

package initcmd

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/palantir"
)

// displayInitCompletion displays completion message and next steps.
func displayInitCompletion(warnings []string) error {
	o := palantir.GetGlobalOutputHandler()
	o.PrintHeader("Initialization Complete!")
	o.PrintInfo("Anvil has been successfully initialized and is ready to use.")
	o.PrintInfo("Configuration files have been created in: %s", config.AnvilConfigPath())

	if len(warnings) > 0 {
		fmt.Println("")
		o.PrintInfo("Recommended next steps to complete your setup:")
		for _, warning := range warnings {
			o.PrintInfo("  • %s", warning)
		}
		fmt.Println("")
		o.PrintInfo("These steps are optional but recommended for the best experience.")
	}

	fmt.Println("")
	o.PrintInfo("You can now use:")
	o.PrintInfo("  • 'anvil install [group]' to install development tool groups")
	o.PrintInfo("  • 'anvil install [app]' to install any individual application")
	o.PrintInfo("  • Edit %s/%s to customize your configuration", config.AnvilConfigDirectory(), constants.ANVIL_CONFIG_FILE)

	o.PrintWarning("Configuration Management Setup Required:")
	o.PrintInfo("  • Edit the 'github.config_repo' field in %s to enable config pull/push", constants.ANVIL_CONFIG_FILE)
	o.PrintInfo("  • Example: 'github.config_repo: username/dotfiles'")
	o.PrintInfo("  • Set GITHUB_TOKEN environment variable for authentication")
	o.PrintInfo("  • Run 'anvil doctor' once added to validate configuration")

	if groups, err := config.AvailableGroups(); err == nil {
		builtInGroups := config.BuiltInGroups()
		fmt.Println("")
		o.PrintInfo("Available groups: %s", strings.Join(builtInGroups, ", "))
		if len(groups) > len(builtInGroups) {
			o.PrintInfo("Custom groups: %d defined", len(groups)-len(builtInGroups))
		}
	} else {
		o.PrintInfo("Available groups: dev, essentials")
	}
	o.PrintInfo("Example: 'anvil install essentials' or 'anvil install firefox'")

	return nil
}
