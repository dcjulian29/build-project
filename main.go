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

// build-project is a command-line tool that automatically detects and executes
// the appropriate build system for the project in the current working directory.
// It looks for well-known files (e.g., go.mod, dockerfile, *.sln, ansible.cfg,
// .goreleaser.yml) and build scripts (e.g., build.sh, build.ps1, build.bat)
// to determine the correct build action, then runs the corresponding toolchain.
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/dcjulian29/build-project/internal"
	"github.com/dcjulian29/go-toolbox/textformat"
)

// main is the entry point for the build-project CLI. It parses command-line
// flags, detects the appropriate build action for the current directory,
// and delegates execution to the internal package.
func main() {
	var action string

	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			action = os.Args[1]
			action = strings.TrimLeft(action, "-")
		} else {
			action = internal.DetectAction()
		}
	} else {
		action = internal.DetectAction()
	}

	if err := internal.PerformAction(action); err != nil {
		fmt.Println(textformat.Red(err.Error()))
		os.Exit(1)
	}
}
