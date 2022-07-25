/*
Copyright 2022 Brian Pursley.

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

package util

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

// PrintSectionTitle prints a title with formatting
func PrintSectionTitle(cmd *cobra.Command, title string) {
	cmd.Printf("========== %s %s\n", title, strings.Repeat("=", 17-len(title)))
}

// GetKubectlPath returns the path that should be used to execute kubectl. You can set the
// KUBECTL_PATH environment variable to override the path.
func GetKubectlPath() string {
	if kubectlPath, found := os.LookupEnv("KUBECTL_PATH"); found {
		return kubectlPath
	}
	return "kubectl"
}

// HasOutputFlag returns try if osArgs contains -o or --output
func HasOutputFlag() bool {
	for _, a := range os.Args {
		if strings.HasPrefix(a, "-o") || strings.HasPrefix(a, "--output") {
			return true
		}
	}
	return false
}

// ExecRun runs the specified executable with args and the specified stdin, stdout, and stderr
var ExecRun = func(name string, args []string, stdin io.Reader, stdout, stderr io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	return cmd.Run()
}

// IsNonRegularFile returns true if the file is not a regular file
var IsNonRegularFile = func(name string) bool {
	fi, err := os.Stat(name)
	return err == nil && !fi.IsDir() && !fi.Mode().IsRegular()
}

// Exit is a wrapper around os.Exit to help with mocking
var Exit = func(code int) {
	os.Exit(code)
}
