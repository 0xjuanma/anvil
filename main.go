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
package main

import (
	"os"

	"github.com/0xjuanma/anvil/cmd"
	"github.com/0xjuanma/anvil/internal/terminal/charm"
	"github.com/0xjuanma/anvil/internal/version"
)

// appVersion is set at build time via ldflags
var appVersion = "dev-local"

func main() {
	// Initialize enhanced Charm output
	charm.InitCharmOutput()

	// Set application version
	version.SetVersion(appVersion)

	// Execute the CLI
	if err := cmd.Execute(); err != nil {
		// Error is already formatted with context by Execute()
		os.Exit(1)
	}
}
