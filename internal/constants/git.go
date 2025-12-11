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

// Git environment variables for non-interactive mode
const (
	GitTerminalPrompt = "GIT_TERMINAL_PROMPT=0"
	GitAskPass        = "GIT_ASKPASS=/bin/false"
	SSHAskPass        = "SSH_ASKPASS=/bin/false"
	GitSSHCommand     = "GIT_SSH_COMMAND=ssh -o BatchMode=yes -o StrictHostKeyChecking=no"
)

// GitNonInteractiveEnvVars returns environment variables for non-interactive git operations
func GitNonInteractiveEnvVars() []string {
	return []string{
		GitTerminalPrompt,
		GitAskPass,
		SSHAskPass,
		GitSSHCommand,
	}
}
