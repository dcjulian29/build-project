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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dcjulian29/go-toolbox/color"
	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
)

// PerformAction executes the build action identified by the action string. It
// supports the following build systems and actions:
//   - "ansible"    – runs ansible-lint on the current directory (requires ansible.cfg)
//   - "archive"    – compresses the current directory into a .7z archive using 7z
//   - "bash"       – executes build.sh using bash (Linux/macOS only)
//   - "bat"        – executes build.bat using cmd.exe (Windows only)
//   - "cake"       – runs a Cake build script via the dotnet Cake.Tool, installing
//     or restoring it if necessary
//   - "cmd"        – executes build.cmd using cmd.exe (Windows only)
//   - "docker"     – runs "docker build ." (requires a dockerfile)
//   - "go"         – runs go mod tidy, go vet, and go build (requires go.mod)
//   - "goreleaser" – runs goreleaser release --snapshot --clean
//   - "powershell" – executes build.ps1 using Windows PowerShell (Windows only)
//   - "pwsh"       – executes build.ps1 using PowerShell Core (cross-platform)
//   - "sh"         – executes build.sh using sh (Linux/macOS only)
//
// Returns an error if the action fails, or an error indicating that nothing
// was found to build or that the build system is unknown.
func PerformAction(action string) error {

	var err error = nil

	switch action {
	case "ansible":
		if filesystem.FileExists("ansible.cfg") {
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
			if filesystem.FileExists("build.sh") {
				err = execute.ExternalProgram("bash", "build.sh")
			} else {
				err = errors.New("build.sh file does not exists")
			}
		} else {
			err = errors.New("this type of build system requires Linux or MacOS")
		}
	case "bat":
		if runtime.GOOS == "windows" {
			if filesystem.FileExists("build.bat") {
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
			if !filesystem.FileExists("dotnet-tools.json") {
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
			if filesystem.FileExists("build.cmd") {
				err = execute.ExternalProgram("cmd.exe", "/C", "build.cmd")
			} else {
				err = errors.New("build.cmd file does not exists")
			}
		} else {
			err = errors.New("this type of build system requires Windows")
		}
	case "docker":
		if filesystem.FileExists("dockerfile") {
			err = execute.ExternalProgram("docker", "build", ".")
		} else {
			err = errors.New("dockerfile file does not exist")
		}
	case "go":
		if filesystem.FileExists("go.mod") {
			err = execute.ExternalProgram("go", "mod", "tidy")

			if err == nil {
				err = execute.ExternalProgram("go", "vet")
			}

			if err == nil {
				err = execute.ExternalProgram("go", "build", "-a", "-v", ".")
			}
		} else {
			err = errors.New("go.mod file does not exists")
		}
	case "goreleaser":
		if filesystem.FileExists(".goreleaser.yml") || filesystem.FileExists(".goreleaser.yaml") {
			err = execute.ExternalProgram("goreleaser", "release", "--snapshot", "--clean")
		} else {
			err = errors.New(".goreleaser.yml file does not exists")
		}
	case "powershell":
		if filesystem.FileExists("build.ps1") {
			if runtime.GOOS == "windows" {
				err = execute.ExternalProgram("powershell", "-f", "build.ps1")
			} else {
				err = errors.New("this type of build system requires Windows")
			}
		} else {
			err = errors.New("build.ps1 file does not exists")
		}
	case "pwsh":
		if filesystem.FileExists("build.ps1") {
			err = execute.ExternalProgram("pwsh", "-f", "build.ps1")
		} else {
			err = errors.New("build.ps1 file does not exists")
		}
	case "sh":
		if runtime.GOOS != "windows" {
			if filesystem.FileExists("build.sh") {
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

	return err
}
