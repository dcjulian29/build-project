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
	"fmt"
	"runtime"
)

func buildDos(batchFile bool) error {
	if runtime.GOOS == "windows" {
		if batchFile {
			if fileExists("build.bat") {
				return run("cmd.exe", []string{"/C", "build.bat"})
			} else {
				return fmt.Errorf("%s", "build.bat file does not exists")
			}
		} else {
			if fileExists("build.cmd") {
				return run("cmd.exe", []string{"/C", "build.cmd"})
			} else {
				return fmt.Errorf("%s", "build.cmd file does not exists")
			}
		}
	} else {
		return fmt.Errorf("%s", "this type of build system requires Windows")
	}
}
