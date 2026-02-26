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
package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dcjulian29/go-toolbox/color"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/io"
)

func main() {
	action := ""

	if len(os.Args) > 1 {
		if strings.HasPrefix(os.Args[1], "-") {
			action = os.Args[1]
			action = strings.ReplaceAll(action, "-", "")
		} else {
			action = autoDetectAction()
		}
	} else {
		action = autoDetectAction()
	}

	if err := preformAction(action); err != nil {
		fmt.Println(color.Red(err.Error()))
		os.Exit(1)
	}
}

func autoDetectAction() string {
	fmt.Println(color.Info("Detecting build system..."))

	action := ""

	if io.FileExists("ansible.cfg") {
		action = "ansible"
	}

	if io.FileExists("dockerfile") {
		action = "docker"
	}

	if io.FileExists("go.mod") {
		action = "go"
	}

	if io.FileExists(".goreleaser.yml") || io.FileExists(".goreleaser.yaml") {
		action = "goreleaser"
	}

	if io.FileExists("build.cake") {
		action = "cake"
	}

	// Check for platform-specific build scripts
	if runtime.GOOS == "windows" {
		if io.FileExists("build.bat") && isShellAvailable("cmd") {
			action = "bat"
		}

		if io.FileExists("build.cmd") && isShellAvailable("cmd") {
			action = "cmd"
		}

		if io.FileExists("build.ps1") && isShellAvailable("powershell") {
			action = "powershell"
		}
	} else {
		if io.FileExists("build.sh") && isShellAvailable("sh") {
			action = "sh"
		}

		if io.FileExists("build.sh") && isShellAvailable("bash") {
			action = "bash"
		}
	}

	if io.FileExists("build.ps1") && isShellAvailable("pwsh") {
		action = "pwsh"
	}

	return action
}

func isShellAvailable(shell string) bool {
	var cmd *exec.Cmd

	switch shell {
	case "bash", "sh", "pwsh":
		cmd = exec.Command(shell, "-c", "exit 0")
	case "cmd":
		cmd = exec.Command(shell, "/C", "exit 0")
	case "powershell":
		cmd = exec.Command(shell, "-Command", "exit 0")
	default:
		return false
	}

	return cmd.Run() == nil
}

func preformAction(action string) error {

	var err error = nil

	switch action {
	case "ansible":
		if io.FileExists("ansible.cfg") {
			err = execute.ExternalProgram("ansible-lint", ".")
		} else {
			err = errors.New("ansible.cfg file does not exists")
		}
	case "archive":
		pwd, _ := os.Getwd()
		name := filepath.Base(pwd)

		dst := fmt.Sprintf("../%s.7z", name)

		fmt.Printf("Archiving '%s'...\n", pwd)

		err = execute.ExternalProgram("7z", "a", "-t7z", "-mx9", "-y", "-r", dst, ".")
	case "bash":
		if runtime.GOOS != "windows" {
			if io.FileExists("build.sh") {
				err = execute.ExternalProgram("bash", "build.sh")
			} else {
				err = errors.New("build.sh file does not exists")
			}
		} else {
			err = errors.New("this type of build system requires Linux or MacOS")
		}
	case "bat":
		if runtime.GOOS == "windows" {
			if io.FileExists("build.bat") {
				err = execute.ExternalProgram("cmd.exe", "/C", "build.bat")
			} else {
				err = errors.New("build.bat file does not exists")
			}
		} else {
			err = errors.New("this type of build system requires Windows")
		}
	case "cake":
		tools, err := execute.ExternalProgramCapture("dotnet", "tool", "list")
		if err != nil {
			err = errors.New("dotnet SDK is not present")
		}

		if !strings.Contains(tools, "cake.tool") {
			if !io.FileExists("dotnet-tools.json") {
				err = execute.ExternalProgram("dotnet", "new", "tool-manifest")
			}

			if err == nil {
				fmt.Println(color.Info("Installing Cake.Tool"))
				err = execute.ExternalProgram("dotnet", "tool", "install", "Cake.Tool")

				if err != nil {
					err = errors.New("Cake.Tool is not present and could not be installed")
				}
			}
		}

		arguments := []string{"cake"}

		if len(os.Args) > 1 {
			if os.Args[1] == "cake" {
				arguments = append(arguments, "--target="+os.Args[2])
			} else {
				arguments = append(arguments, "--target="+os.Args[1])
			}

			msg := fmt.Sprintf("Using target '%s'", strings.Split(arguments[len(arguments)-1], "=")[1])
			fmt.Println(color.Info(msg))
		}

		err = execute.ExternalProgram("dotnet", arguments...)

		if err != nil && strings.Contains(err.Error(), "Could not execute because the specified command or file was not found") {
			err = execute.ExternalProgram("dotnet", "tool", "restore")
			if err != nil {
				err = errors.New("Cake.Tool is not present and could not be restored")
			} else {
				err = execute.ExternalProgram("dotnet", arguments...)
			}
		}
	case "cmd":
		if runtime.GOOS == "windows" {
			if io.FileExists("build.cmd") {
				err = execute.ExternalProgram("cmd.exe", "/C", "build.cmd")
			} else {
				err = errors.New("build.cmd file does not exists")
			}
		} else {
			err = errors.New("this type of build system requires Windows")
		}
	case "docker":
		if io.FileExists("dockerfile") {
			err = execute.ExternalProgram("docker", "build", ".")
		} else {
			err = errors.New("dockerfile file does not exist")
		}
	case "go":
		if io.FileExists("go.mod") {
			err = execute.ExternalProgram("go", "mod", "tidy")

			if err == nil {
				err = execute.ExternalProgram("go", "vet")
			}

			if err == nil {
				err = execute.ExternalProgram("go", "build", "-a", ".")
			}
		} else {
			err = errors.New("go.mod file does not exists")
		}
	case "goreleaser":
		if io.FileExists(".goreleaser.yml") || io.FileExists(".goreleaser.yaml") {
			err = execute.ExternalProgram("goreleaser", "release", "--snapshot", "--clean")
		} else {
			err = errors.New(".goreleaser.yml file does not exists")
		}
	case "powershell":
		if io.FileExists("build.ps1") {
			if runtime.GOOS == "windows" {
				err = execute.ExternalProgram("powershell", "-f", "build.ps1")
			} else {
				err = errors.New("this type of build system requires Windows")
			}
		} else {
			err = errors.New("build.ps1 file does not exists")
		}
	case "pwsh":
		if io.FileExists("build.ps1") {
			err = execute.ExternalProgram("pwsh", "-f", "build.ps1")
		} else {
			err = errors.New("build.ps1 file does not exists")
		}
	case "sh":
		if runtime.GOOS != "windows" {
			if io.FileExists("build.sh") {
				err = execute.ExternalProgram("sh", "build.sh")
			} else {
				err = errors.New("build.sh file does not exists")
			}
		} else {
			err = errors.New("this type of build system requires Linux or MacOS")
		}
	case "":
		err = errors.New("nothing found to build in this directory")
	default:
		err = errors.New("unknown build system specified")
	}

	if err != nil {
		return errors.New(color.Fatal(err.Error()))
	}

	return nil
}
