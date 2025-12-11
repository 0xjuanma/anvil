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

// Common spinner messages
const (
	SpinnerSyncingAnvilSettings = "Syncing anvil settings"
	SpinnerValidatingTools      = "Validating and installing required tools"
	SpinnerCreatingDirectories  = "Creating necessary directories"
	SpinnerCheckingEnvironment  = "Checking local environment configurations"
	SpinnerRunningDiscovery     = "Running discovery logic"
	SpinnerValidatingRepository = "Validating repository access and branch configuration"
	SpinnerCloningRepository    = "Cloning or updating repository"
	SpinnerPullingChanges       = "Pulling latest changes"
	SpinnerSettingUpAuth        = "Setting up authentication..."
	SpinnerAnalyzingChanges      = "Analyzing changes..."
	SpinnerLoadingConfig         = "Loading anvil configuration..."
)

// Common status messages
const (
	StatusInstalled           = "Installed"
	StatusFailed              = "Failed"
	StatusInstalling          = "Installing..."
	StatusPending             = "Pending"
	StatusConfigurationSynced = "configuration synced successfully"
	StatusSettingsSynced      = "settings synced successfully"
	StatusUpToDate            = "Configuration up-to-date (no changes)"
	StatusPushedSuccessfully  = "Configuration pushed successfully"
)

// Common error message templates
const (
	ErrConfigNotPulled      = "config not pulled yet"
	ErrAppConfigNotDefined  = "app config path not defined"
	ErrAppConfigNotConfigured = "app config path not configured in settings"
	ErrNoConfigsSection     = "no configs section found in %s"
	ErrGitHubRepoNotSet     = "GitHub repository not configured. Please set 'github.config_repo' in your %s"
)
