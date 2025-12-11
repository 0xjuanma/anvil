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

package show

import (
	"fmt"
	"strings"

	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/errors"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/tools"
	"github.com/0xjuanma/anvil/internal/utils"
	"github.com/0xjuanma/palantir"
)

// showAnvilSettingsSection displays specific sections of the anvil settings.
func showAnvilSettingsSection(showGroups, showConfigs, showSources, showGit, showGitHub bool) error {
	o := palantir.GetGlobalOutputHandler()

	// Stage 1: Locate settings file
	configPath := config.AnvilConfigPath()

	// Check settings file
	err := checkSettingsFileExists(o, configPath)
	if err != nil {
		return err
	}

	// Stage 2: Load configuration
	anvilConfig, err := config.LoadConfig()
	if err != nil {
		return errors.NewFileSystemError(constants.OpShow, "load-config", err)
	}

	// Stage 3: Display requested section
	switch {
	case showGroups:
		if err := showGroupsSection(); err != nil {
			return err
		}
	case showConfigs:
		if err := showConfigsSection(anvilConfig); err != nil {
			return err
		}
	case showSources:
		if err := showSourcesSection(anvilConfig); err != nil {
			return err
		}
	case showGit:
		if err := showGitSection(anvilConfig); err != nil {
			return err
		}
	case showGitHub:
		if err := showGitHubSection(anvilConfig); err != nil {
			return err
		}
	}

	return nil
}

// showGroupsSection displays the groups section using shared rendering functions.
func showGroupsSection() error {
	appData, err := tools.LoadAndPrepareAppData()
	if err != nil {
		return err
	}

	// Use the shared tree view renderer
	content := utils.RenderTreeView(appData)

	fmt.Println(charm.RenderBox("Groups", content, "#E0C867", false))

	return nil
}

// showConfigsSection displays the configs section.
func showConfigsSection(anvilConfig *config.AnvilConfig) error {
	var boxContent strings.Builder

	if len(anvilConfig.Configs) == 0 {
		boxContent.WriteString("  No configured source directories found.\n")
		boxContent.WriteString("  Use 'anvil config push <app-name> <path>' to configure source directories.\n")
	} else {
		for appName, path := range anvilConfig.Configs {
			boxContent.WriteString(fmt.Sprintf("    %s: %s\n", utils.ColorAppName(appName), path))
		}
	}

	fmt.Println(charm.RenderBox("Config Sources", boxContent.String(), "#E0C867", false))
	fmt.Println()
	fmt.Println("  💡 Use 'anvil config push <app-name>' to push source directories")
	fmt.Println("  💡 Use 'anvil config pull <app-name>' to pull configurations")
	fmt.Println()

	return nil
}

// showSourcesSection displays the installation sources section.
func showSourcesSection(anvilConfig *config.AnvilConfig) error {
	var boxContent strings.Builder

	if len(anvilConfig.Sources) == 0 {
		boxContent.WriteString("  No installation sources configured.\n")
		boxContent.WriteString("  Add sources to your settings.yaml to configure installation URLs or commands.\n")
	} else {
		for appName, url := range anvilConfig.Sources {
			boxContent.WriteString(fmt.Sprintf("    %s: %s\n", utils.ColorAppName(appName), url))
		}
	}

	fmt.Println(charm.RenderBox("Installation Sources", boxContent.String(), "#E0C867", false))
	fmt.Println()
	fmt.Println("  💡 Sources map app names to their installation URLs or commands")
	fmt.Println()

	return nil
}

// showGitSection displays the git configuration section.
func showGitSection(anvilConfig *config.AnvilConfig) error {
	var boxContent strings.Builder

	boxContent.WriteString(fmt.Sprintf("    Username: %s\n", utils.BoldText(anvilConfig.Git.Username, "")))
	boxContent.WriteString(fmt.Sprintf("    Email: %s\n", utils.BoldText(anvilConfig.Git.Email, "")))
	if anvilConfig.Git.SSHKeyPath != "" {
		boxContent.WriteString(fmt.Sprintf("    SSH Key Path: %s\n", utils.BoldText(anvilConfig.Git.SSHKeyPath, "")))
	}

	fmt.Println(charm.RenderBox("Git Configuration", boxContent.String(), "#CC78EB", false))
	fmt.Println()
	fmt.Println("  💡 Git configuration is auto-populated from your local git settings")
	fmt.Println()

	return nil
}

// showGitHubSection displays the GitHub configuration section.
func showGitHubSection(anvilConfig *config.AnvilConfig) error {
	var boxContent strings.Builder

	boxContent.WriteString(fmt.Sprintf("    Repository: %s\n", utils.BoldText(anvilConfig.GitHub.ConfigRepo, "")))
	boxContent.WriteString(fmt.Sprintf("    Branch: %s\n", utils.BoldText(anvilConfig.GitHub.Branch, "")))
	boxContent.WriteString(fmt.Sprintf("    Local Path: %s\n", utils.BoldText(anvilConfig.GitHub.LocalPath, "")))
	if anvilConfig.GitHub.Token != "" {
		boxContent.WriteString(fmt.Sprintf("    Token: %s\n", utils.BoldText(anvilConfig.GitHub.Token, "")))
	}
	if anvilConfig.GitHub.TokenEnvVar != "" {
		boxContent.WriteString(fmt.Sprintf("    Token Environment Variable: %s\n", utils.BoldText(anvilConfig.GitHub.TokenEnvVar, "")))
	}

	fmt.Println(charm.RenderBox("GitHub Configuration", boxContent.String(), "#CC78EB", false))

	return nil
}
