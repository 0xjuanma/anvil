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

// Common SSH key names in order of preference for auto-detection
var CommonSSHKeyNames = []string{
	"id_ed25519",
	"id_ed25519_personal",
	"id_rsa",
	"id_rsa_personal",
	"id_ecdsa",
}

// DefaultSSHKeyName is the default SSH key name used when no keys are found
const DefaultSSHKeyName = "id_ed25519"

// CommonSSHKeyFiles lists common SSH key file names for validation
var CommonSSHKeyFiles = []string{
	"id_rsa",
	"id_ed25519",
	"id_ecdsa",
}
