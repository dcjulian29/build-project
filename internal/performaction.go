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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dcjulian29/go-toolbox/execute"
	"github.com/dcjulian29/go-toolbox/filesystem"
	"github.com/dcjulian29/go-toolbox/textformat"
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
//   - "dotnet"     - runs dotnet build to compile
//   - "go"         – runs go mod tidy, go vet, and go build (requires go.mod)
//   - "goreleaser" – runs goreleaser release --snapshot --clean
//   - "hugo"       - runs "hugo build" (requires a hugo.toml file)
//   - "powershell" – executes build.ps1 using Windows PowerShell (Windows only)
//   - "pwsh"       – executes build.ps1 using PowerShell Core (cross-platform)
//   - "sh"         – executes build.sh using sh (Linux/macOS only)
//
// Returns an error if the action fails, or an error indicating that nothing
// was found to build or that the build system is unknown.
func PerformAction(action string) error {
	switch action {
	case "ansible":
		if filesystem.FileExist("ansible.cfg") {
			return execute.ExternalProgram("ansible-lint", ".")
		}

		return errors.New("ansible.cfg file does not exists")

	case "archive":
		pwd, _ := os.Getwd()
		name := filepath.Base(pwd)

		dst := fmt.Sprintf("../%s.7z", name)

		fmt.Printf("Archiving '%s'...\n", pwd)

		return execute.ExternalProgram("7z", "a", "-t7z", "-mx9", "-y", "-r", dst, ".")

	case "bash":
		if runtime.GOOS != "windows" {
			if filesystem.FileExist("build.sh") {
				return execute.ExternalProgram("bash", "build.sh")
			}

			return errors.New("build.sh file does not exists")
		}

		return errors.New("this type of build system requires Linux or MacOS")

	case "bat":
		if runtime.GOOS == "windows" {
			if filesystem.FileExist("build.bat") {
				return execute.ExternalProgram("cmd.exe", "/C", "build.bat")
			}

			return errors.New("build.bat file does not exists")
		}

		return errors.New("this type of build system requires Windows")

	case "cake":
		tools, err := execute.ExternalProgramCapture("dotnet", "tool", "list")
		if err != nil {
			return errors.New("dotnet SDK is not present")
		}

		if !strings.Contains(tools, "cake.tool") {
			if !filesystem.FileExist("dotnet-tools.json") {
				err = execute.ExternalProgram("dotnet", "new", "tool-manifest")
			}

			if err == nil {
				fmt.Println(textformat.Info("Installing Cake.Tool"))
				err = execute.ExternalProgram("dotnet", "tool", "install", "Cake.Tool")

				if err != nil {
					return errors.New("Cake.Tool is not present and could not be installed")
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
			fmt.Println(textformat.Info(msg))
		}

		err = execute.ExternalProgram("dotnet", arguments...)

		if err != nil && strings.Contains(err.Error(), "Could not execute because the specified command or file was not found") {
			if err := execute.ExternalProgram("dotnet", "tool", "restore"); err != nil {
				return errors.New("Cake.Tool is not present and could not be restored")
			}

			return execute.ExternalProgram("dotnet", arguments...)

		}

		return nil

	case "cmd":
		if runtime.GOOS == "windows" {
			if filesystem.FileExist("build.cmd") {
				return execute.ExternalProgram("cmd.exe", "/C", "build.cmd")
			}

			return errors.New("build.cmd file does not exists")
		}

		return errors.New("this type of build system requires Windows")

	case "docker":
		if filesystem.FileExist("dockerfile") {
			return execute.ExternalProgram("docker", "build", ".")
		}

		return errors.New("dockerfile file does not exist")

	case "dotnet":
		if !IsShellAvailable("dotnet") {
			return errors.New("dotnet CLI is not available; install the .NET SDK from https://dot.net")
		}

		s, _ := filepath.Glob("*.sln")
		c, _ := filepath.Glob("*.csproj")
		f, _ := filepath.Glob("*.fsproj")

		var target string

		switch {
		case len(s) == 1:
			target = s[0]
		case len(c) == 1:
			target = c[0]
		case len(f) == 1:
			target = f[0]
		}

		if len(target) > 0 {
			return execute.ExternalProgram("dotnet", "build", target)
		}

		return errors.New("no dotnet project file was found")

	case "go":
		if filesystem.FileExist("go.mod") {
			if err := execute.ExternalProgram("go", "mod", "tidy"); err != nil {
				return err
			}

			if err := execute.ExternalProgram("go", "vet", "./..."); err != nil {
				return err
			}

			if err := execute.ExternalProgram("go", "build", "-a", "-v", "./..."); err != nil {
				return err
			}

			if err := execute.ExternalProgram("go", "test", "-v", "./..."); err != nil {
				return err
			}
		}

		return errors.New("go.mod file does not exists")

	case "goreleaser":
		if filesystem.FileExist(".goreleaser.yml") || filesystem.FileExist(".goreleaser.yaml") {
			return execute.ExternalProgram("goreleaser", "release", "--snapshot", "--clean")
		}

		return errors.New(".goreleaser.yml file does not exists")

	case "hugo":
		if filesystem.FileExist("hugo.toml") {
			return execute.ExternalProgram("hugo", "build")
		}

		return errors.New("hugo.toml file does not exist")

	case "powershell":
		if filesystem.FileExist("build.ps1") {
			if runtime.GOOS == "windows" {
				return execute.ExternalProgram("powershell", "-f", "build.ps1")
			}

			return errors.New("this type of build system requires Windows")
		}

		return errors.New("build.ps1 file does not exists")

	case "pwsh":
		if filesystem.FileExist("build.ps1") {
			return execute.ExternalProgram("pwsh", "-f", "build.ps1")
		}

		return errors.New("build.ps1 file does not exists")

	case "sh":
		if runtime.GOOS != "windows" {
			if filesystem.FileExist("build.sh") {
				return execute.ExternalProgram("sh", "build.sh")
			}

			return errors.New("build.sh file does not exists")
		}

		return errors.New("this type of build system requires Linux or MacOS")

	case "":
		return errors.New("nothing found to build in this directory")

	default:
		return errors.New("unknown build system specified")
	}
}
