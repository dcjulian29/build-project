/*
Copyright © 2026 Julian Easterling

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

package internal

import (
	"os/exec"
)

// IsShellAvailable checks whether a given shell interpreter is available and
// executable on the current system. It attempts to run a no-op exit command using
// the specified shell (supported values: "bash", "sh", "pwsh", "cmd",
// "powershell"). Returns true if the shell exits successfully, or false if the
// shell is unsupported or the command fails.
func IsShellAvailable(shell string) bool {
	var cmd *exec.Cmd

	switch shell {
	case "bash", "sh":
		cmd = exec.Command(shell, "-c", "exit 0")
	case "cmd":
		cmd = exec.Command(shell, "/C", "exit 0")
	case "powershell", "pwsh":
		cmd = exec.Command(shell, "-Command", "exit 0")
	default:
		return false
	}

	return cmd.Run() == nil
}
