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
	"runtime"

	"github.com/dcjulian29/go-toolbox/filesystem"
)

// DetectBuildScript refines the build action by checking for platform-specific
// build scripts (e.g., build.bat, build.cmd, build.ps1 on Windows; build.sh on
// Linux/macOS) and verifying that the corresponding shell interpreter is available.
// On all platforms, it also checks for a build.ps1 file executable via PowerShell
// Core (pwsh). The function returns the most appropriate action string based on
// the detected scripts and available shells.
func DetectBuildScript(action string) string {
	switch runtime.GOOS {
	case "windows":
		if filesystem.FileExists("build.bat") && IsShellAvailable("cmd") {
			action = "bat"
		}

		if filesystem.FileExists("build.cmd") && IsShellAvailable("cmd") {
			action = "cmd"
		}

		if filesystem.FileExists("build.ps1") && IsShellAvailable("powershell") {
			action = "powershell"
		}
	case "linux", "darwin":
		if filesystem.FileExists("build.sh") && IsShellAvailable("sh") {
			action = "sh"
		}

		if filesystem.FileExists("build.sh") && IsShellAvailable("bash") {
			action = "bash"
		}
	}

	if filesystem.FileExists("build.ps1") && IsShellAvailable("pwsh") {
		action = "pwsh"
	}

	return action
}
