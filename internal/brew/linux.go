package brew

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

import (
	"fmt"
	"time"

	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/0xjuanma/anvil/internal/system"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/palantir"
)

// InstallBrewLinux installs Homebrew on Linux systems
func InstallBrewLinux() error {
	if IsBrewInstalled() {
		return nil
	}

	palantir.GetGlobalOutputHandler().PrintInfo("Installing Homebrew on Linux (this may take a few minutes)")
	palantir.GetGlobalOutputHandler().PrintInfo("You may be prompted for your password to complete the installation")
	fmt.Println()

	spinner := charm.NewDotsSpinner("Preparing Homebrew installation for Linux")
	spinner.Start()
	time.Sleep(constants.SpinnerDelay)
	spinner.Stop()

	fmt.Print("\r\033[K→ Enter password when prompted: ")

	// Use Linux-specific Homebrew installation script
	installScript := `echo | /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"`

	spinner = charm.NewDotsSpinner("Installing Homebrew on Linux")
	spinner.Start()
	err := system.RunInteractiveCommand("/bin/bash", "-c", installScript)
	spinner.Stop()
	fmt.Println()

	if err != nil {
		palantir.GetGlobalOutputHandler().PrintError("Homebrew installation failed")
		return fmt.Errorf("failed to install Homebrew on Linux: %w", err)
	}

	spinner = charm.NewDotsSpinner("Verifying Homebrew installation")
	spinner.Start()

	if !IsBrewInstalledAtPath() {
		spinner.Error("Homebrew installation verification failed")
		return fmt.Errorf("Homebrew installation completed but brew command not accessible")
	}

	spinner.Success("Homebrew installed successfully on Linux")
	return nil
}
