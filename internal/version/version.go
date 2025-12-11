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

// Package version provides version management for the Anvil CLI.
// The version is typically set at build time and can be retrieved at runtime.
package version

// appVersion holds the application version set at build time
var appVersion = "dev"

// SetVersion sets the application version
func SetVersion(v string) {
	appVersion = v
}

// Version returns the current application version
func Version() string {
	return appVersion
}
