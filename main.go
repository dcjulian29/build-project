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

var (
	isBash       bool
	isCore       bool
	isDos        bool
	isPowershell bool
)

func main() {
	determineTerminal()

	action := ""

	if len(os.Args) > 1 {
		action = os.Args[1]
		action = strings.ReplaceAll(action, "-", "")
	} else {
		action = autoDetect()
	}

	err := preformAction(action)

	if err != nil {
		fmt.Println(fmt.Errorf("error: %s", color.RedString(err.Error())))
		os.Exit(1)
	}
}

func autoDetect() string {
	action := ""

	if fileExists("ansible.cfg") {
		action = "ansible"
	}

	if fileExists("dockerfile") {
		action = "docker"
	}

	if fileExists("go.mod") {
		action = "go"
	}

	if fileExists(".goreleaser.yml") || fileExists(".goreleaser.yaml") {
		action = "goreleaser"
	}

	if fileExists("build.cake") {
		action = "cake"
	}

	if fileExists("build.sh") {
		if isBash {
			action = "sh"
		}
	} else {
		if fileExists("build.bat") {
			if isDos || isPowershell {
				action = "bat"
			}
		}

		if fileExists("build.cmd") {
			if isDos || isPowershell {
				action = "cmd"
			}
		}
	}

	if fileExists("build.ps1") {
		if isPowershell {
			action = "powershell"
		}
	}

	return action
}

func determineTerminal() {
	line := "echo $SHELL"
	cmd := exec.Command(line, "")
	cmd.Stdin = os.Stdin
	out, _ := cmd.CombinedOutput()

	terminal := string(out)

	if len(terminal) > 0 || terminal != "$SHELL" {
		if strings.Contains(terminal, "bash") {
			isBash = true
		}
	} else {
		line := "(dir 2>&1 *`|echo CMD);&<# rem #>echo ($PSVersionTable).PSEdition"
		cmd := exec.Command(line, "")
		cmd.Stdin = os.Stdin
		out, _ := cmd.CombinedOutput()

		terminal := string(out)

		switch terminal {
		case "CMD":
			isDos = true

		case "Core":
			isCore = true
			isPowershell = true

		case "Desktop":
			isPowershell = true
		}
	}
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}

	return !info.IsDir()
}

func preformAction(action string) error {

	var err error = nil

	switch action {
	case "archive":
		archive()
	case "cake":
		buildCake()
	case "powershell":
		err = buildPowershell()
	case "bat":
		err = buildDos(true)
	case "cmd":
		err = buildDos(false)
	case "sh":
		err = buildBash()
	case "goreleaser":
		buildGoReleaser()
	case "go":
		buildGo()
	case "ansible":
		buildAnsible()
	case "docker":
		buildDocker()
	case "":
		return fmt.Errorf("%s", color.RedString("nothing found to build in this directory"))
	default:
		return fmt.Errorf("%s", color.RedString("unknown build system specified"))
	}

	if err != nil {
		return fmt.Errorf("%s", color.RedString(err.Error()))
	}

	return nil
}

func run(binary string, params []string) error {
	cmd := exec.Command(binary, params...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin

	return cmd.Run()
}
