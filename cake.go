/*
Copyright © 2024 Julian Easterling <julian@julianscorner.com>

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
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/fatih/color"
)

func buildCake() error {
	cmd := exec.Command("dotnet", "tool", "list")
	tools, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("%s", color.RedString("dotnet SDK is not present!s"))
	}

	if !strings.Contains(string(tools), "cake.tool") {
		if !fileExists(".config/dotnet-tools.json") {
			cmd := exec.Command("dotnet", "new", "tool-manifest")
			_, err := cmd.CombinedOutput()

			if err != nil {
				return fmt.Errorf("%s", color.RedString("Installing Cake.Tool: %s", err))
			}

			err = run("dotnet", []string{"tool", "install", "Cake.Tool"})

			if err != nil {
				return fmt.Errorf("%s", color.RedString("Cake.Tool is not present and could not be installed!"))
			}
		}
	}

	var params []string

	if len(os.Args) > 0 {
		if os.Args[0] == "cake" {
			if !strings.Contains("-", os.Args[1]) {
				params = []string{"--target=" + os.Args[1]}
			} else {
				params = os.Args[1:]
			}
		} else {
			params = os.Args
		}
	}

	params = append([]string{"cake"}, params...)
	return run("dotnet", params)
}
