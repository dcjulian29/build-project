package internal

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

import (
	"fmt"

	"github.com/dcjulian29/go-toolbox/color"
	"github.com/dcjulian29/go-toolbox/filesystem"
)

// DetectAction inspects well-known build system indicator files in the current
// working directory to determine the appropriate build action.
// After file-based detection, it delegates to AutoDetectBuildScript to further
// refine the action based on available shell scripts and the
// current operating system.
func DetectAction() string {
	fmt.Println(color.Info("Detecting build system..."))

	var action string

	if filesystem.FileExists("ansible.cfg") {
		action = "ansible"
	}

	if filesystem.FileExists("dockerfile") {
		action = "docker"
	}

	if filesystem.FileExists("go.mod") {
		action = "go"
	}

	if filesystem.FileExists(".goreleaser.yml") || filesystem.FileExists(".goreleaser.yaml") {
		action = "goreleaser"
	}

	if filesystem.FileExists("build.cake") {
		action = "cake"
	}

	return DetectBuildScript(action)
}
