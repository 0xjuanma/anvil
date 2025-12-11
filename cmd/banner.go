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

package cmd

import (
	"fmt"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/version"
)

// showWelcomeBanner displays the enhanced welcome banner with quick start
// information when Anvil is run without arguments.
func showWelcomeBanner() {
	// Main banner
	bannerContent := fmt.Sprintf("%s\n🔥 One CLI to rule them all 🔥\n\tversion: %s\n\n", constants.AnvilLogo, version.Version())
	fmt.Println(charm.RenderBox("", bannerContent, "#FF6B9D", true))

	quickStart := `
  anvil init [--discover]				  Initialize your environment and discover installed apps
  anvil install essentials/[group-name]    Install specific group
  anvil doctor							 Check system health and list available checks
  anvil config show [app-name]			 Show your anvil settings or app settings
  anvil config push [app-name]			 Push your app configurations to GitHub
  anvil config pull [app-name]			 Pull your app configurations from GitHub
  anvil config sync [app-name]			 Sync your app configurations to your local machine
  anvil clean							  Clean your anvil environment
  anvil update							 Update your anvil installation
  anvil --version/-v					   Show the version of anvil
`
	fmt.Println(charm.RenderBox("Quick Start", quickStart, "#00D9FF", false))

	// Footer
	fmt.Println()
	fmt.Println("  Documentation: anvil --help")
}

// showVersionInfo displays the version information with branding when
// the --version flag is used.
func showVersionInfo() {
	fmt.Println(charm.RenderBox("ANVIL CLI", version.Version(), "#FF6B9D", true))
}
