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
	"github.com/0xjuanma/anvil/internal/config"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/palantir"
)

// checkToolConfiguration checks if a tool is properly configured.
func checkToolConfiguration(toolName string) error {
	switch toolName {
	case constants.PkgGit:
		return checkGitConfiguration()
	default:
		return nil
	}
}

// checkGitConfiguration checks if git is properly configured.
func checkGitConfiguration() error {
	cfg, err := config.LoadConfig()
	if err == nil && (cfg.Git.Username == "" || cfg.Git.Email == "") {
		o := palantir.GetGlobalOutputHandler()
		o.PrintInfo("Git installed successfully")
		o.PrintWarning("Consider configuring git with:")
		o.PrintInfo("  git config --global user.name 'Your Name'")
		o.PrintInfo("  git config --global user.email 'your.email@example.com'")
	}
	return nil
}
