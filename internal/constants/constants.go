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

package constants

// Command operation constants
const (
	OpInit    = "init"
	OpInstall = "install"
	OpConfig  = "config"
	OpImport  = "import"
	OpPull    = "pull"
	OpPush    = "push"
	OpShow    = "show"
	OpSync    = "sync"
	OpDoctor  = "doctor"
	OpClean   = "clean"
	OpUpdate  = "update"
)

// System command constants
const (
	BrewCommand = "brew"
	GitCommand  = "git"
	CurlCommand = "curl"
)

// Brew subcommand constants
const (
	BrewInstall = "install"
	BrewList    = "list"
	BrewInfo    = "info"
	BrewUpdate  = "update"
	BrewUpgrade = "upgrade"
	BrewSearch  = "search"
)

// Git subcommand constants
const (
	GitConfig    = "config"
	GitGlobal    = "--global"
	GitUserName  = "user.name"
	GitUserEmail = "user.email"
)

// Directory constants
const (
	SSHDir     = ".ssh"
	OhMyZshDir = ".oh-my-zsh"
)

// macOS package names (Homebrew formulae and casks)
const (
	PkgGit       = "git"
	PkgZsh       = "zsh"
	PkgIterm2    = "iterm2"
	PkgVSCode    = "visual-studio-code"
	PkgSlack     = "slack"
	PkgChrome    = "google-chrome"
	Pkg1Password = "1password"
	PkgDocker    = "docker"
	PkgKubectl   = "kubectl"
)

// Environment variables
const (
	EnvEditor = "EDITOR"
	EnvShell  = "SHELL"
	EnvTerm   = "TERM"
	EnvHome   = "HOME"
	EnvPath   = "PATH"
)

// Oh-my-zsh installation
const (
	OhMyZshInstallURL = "https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh"
	OhMyZshInstallCmd = `sh -c "$(curl -fsSL https://raw.github.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" "" --unattended`
)

// ASCII Art Logo
const (
	AnvilLogo = ` █████╗ ███╗   ██╗██╗   ██╗██╗██╗     
██╔══██╗████╗  ██║██║   ██║██║██║     
███████║██╔██╗ ██║██║   ██║██║██║     
██╔══██║██║╚██╗██║╚██╗ ██╔╝██║██║     
██║  ██║██║ ╚████║ ╚████╔╝ ██║███████╗
╚═╝  ╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚═╝╚══════╝`
)

// Anvil config constants
const (
	ANVIL             = "anvil"
	ANVIL_CONFIG_FILE = "settings.yaml"
	ANVIL_CONFIG_DIR  = ".anvil"
	DOTFILES_DIR      = "dotfiles"
)

// Common directory permissions
const (
	DirPerm  = 0755
	FilePerm = 0644
)
