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

package cmd

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/brianpursley/kubectl-confirm/internal/util"
	"github.com/brianpursley/kubectl-confirm/internal/version"
)

func TestCheckForNonRegularFiles(t *testing.T) {
	util.IsNonRegularFile = func(name string) bool {
		return name == "63"
	}
	testCases := []struct {
		name                          string
		options                       confirmOptions
		expectedHasAnyNonRegularFiles bool
	}{
		{
			name: "no non regular files",
			options: confirmOptions{
				kustomize: "example/",
				filenames: []string{"foo.yaml", "bar.yaml"},
			},
		},
		{
			name: "kustomize is non regular file",
			options: confirmOptions{
				kustomize: "63",
				filenames: []string{"foo.yaml", "bar.yaml"},
			},
			expectedHasAnyNonRegularFiles: true,
		},
		{
			name: "filename is non regular file",
			options: confirmOptions{
				kustomize: "example/",
				filenames: []string{"foo.yaml", "63"},
			},
			expectedHasAnyNonRegularFiles: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.options.checkForNonRegularFiles()

			if tc.options.hasAnyNonRegularFiles != tc.expectedHasAnyNonRegularFiles {
				t.Fatalf("wrong hasAnyNonRegularFiles. expected: %v, got: %v", tc.expectedHasAnyNonRegularFiles, tc.options.hasAnyNonRegularFiles)
			}
		})
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		name                string
		options             confirmOptions
		kubectlPath         string
		fakeOsArgs          []string
		fakeArgs            []string
		response            string
		expectedStdout      string
		unexpectedStdout    string
		expectedStderr      string
		expectKubectl       bool
		expectedKubectlArgs []string
		expectedExitCode    int
	}{
		{
			name:             "help should show help and exit",
			options:          confirmOptions{},
			fakeArgs:         []string{"help"},
			fakeOsArgs:       []string{"confirm", "help"},
			expectKubectl:    false,
			expectedExitCode: 0,
		},
		{
			name:                "version should show version",
			options:             confirmOptions{},
			fakeArgs:            []string{"version"},
			fakeOsArgs:          []string{"confirm", "version"},
			expectKubectl:       true,
			expectedStdout:      version.String(),
			expectedKubectlArgs: []string{"version"},
			expectedExitCode:    0,
		},
		{
			name:                "version should not show plugin version if output flag is specified",
			options:             confirmOptions{},
			fakeArgs:            []string{"version"},
			fakeOsArgs:          []string{"confirm", "version", "-o", "yaml"},
			expectKubectl:       true,
			unexpectedStdout:    "Kubectl Confirm Plugin Version: ",
			expectedKubectlArgs: []string{"version", "-o", "yaml"},
			expectedExitCode:    0,
		},
		// TODO: Should show dry run when command is dry-runnable
		// TODO: Should skip dry run when command is not dry-runnable
		{
			name:          "should show diff when command is diff-able",
			options:       confirmOptions{},
			fakeArgs:      []string{"apply"},
			fakeOsArgs:    []string{"confirm", "apply", "-f", "foo.yaml"},
			response:      "yes\n",
			expectKubectl: true,
			expectedStdout: `========== Confirm ==========
The following command will be executed:
kubectl apply -f foo.yaml

Enter 'yes' to continue: `,
			expectedKubectlArgs: []string{"apply", "-f", "foo.yaml"},
			expectedExitCode:    0,
		},
		{
			name:          "should use KUBECTL_PATH environment variable",
			options:       confirmOptions{},
			kubectlPath:   "override-kubectl-path",
			fakeArgs:      []string{"apply"},
			fakeOsArgs:    []string{"confirm", "apply", "-f", "foo.yaml"},
			response:      "yes\n",
			expectKubectl: true,
			expectedStdout: `========== Confirm ==========
The following command will be executed:
override-kubectl-path apply -f foo.yaml

Enter 'yes' to continue: `,
			expectedKubectlArgs: []string{"apply", "-f", "foo.yaml"},
			expectedExitCode:    0,
		},
		{
			name:          "should skip diff when command is not not diff-able",
			options:       confirmOptions{},
			fakeArgs:      []string{"delete"},
			fakeOsArgs:    []string{"confirm", "delete", "-f", "foo.yaml"},
			response:      "yes\n",
			expectKubectl: true,
			expectedStdout: `========== Confirm ==========
The following command will be executed:
kubectl delete -f foo.yaml

Enter 'yes' to continue: `,
			unexpectedStdout:    "========== Diff =============",
			expectedKubectlArgs: []string{"delete", "-f", "foo.yaml"},
			expectedExitCode:    0,
		},
		{
			name:          "should abort if response is not yes",
			options:       confirmOptions{},
			fakeArgs:      []string{"delete"},
			fakeOsArgs:    []string{"confirm", "delete", "-f", "foo.yaml"},
			response:      "no\n",
			expectKubectl: true,
			expectedStdout: `========== Confirm ==========
The following command will be executed:
kubectl delete -f foo.yaml

Enter 'yes' to continue: `,
			expectedStderr:      "Command aborted.",
			expectedKubectlArgs: []string{"delete", "-f", "foo.yaml", "--dry-run=server"},
			expectedExitCode:    1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cmd, stdin, stdout, stderr := util.NewTestCommand()
			stdin.Write(bytes.NewBufferString(tc.response).Bytes())

			var actualExitCode int
			util.Exit = func(code int) {
				actualExitCode = code
			}

			commandName := tc.fakeArgs[0]

			fakeExecRunner := util.NewFakeExecRunner()
			fakeExecRunner.SetupRun(`{"current-context": "foo", "contexts": [{"name": "foo", "context": {}}]}`, "", nil)
			if dryRunCommands[commandName] {
				fakeExecRunner.SetupRun("fake dry run output", "", nil)
			}
			if diffCommands[commandName] {
				fakeExecRunner.SetupRun("fake diff dry run output", "", nil)
				fakeExecRunner.SetupRun("fake diff output", "", nil)
			}
			if tc.response == "yes\n" {
				fakeExecRunner.SetupRun("fake real command output", "", nil)
			}

			os.Args = tc.fakeOsArgs

			if len(tc.kubectlPath) > 0 {
				_ = os.Setenv("KUBECTL_PATH", tc.kubectlPath)
				defer os.Unsetenv("KUBECTL_PATH")
			}

			err := tc.options.run(cmd, tc.fakeArgs)
			if err != nil {
				t.Fatalf("Run failed: %v", err)
			}

			if tc.expectKubectl {
				expectedLastRunName := "kubectl"
				if len(tc.kubectlPath) > 0 {
					expectedLastRunName = tc.kubectlPath
				}
				if fakeExecRunner.LastRunName() != expectedLastRunName {
					t.Fatalf("expected %q to be run, but it was %q", expectedLastRunName, fakeExecRunner.LastRunName())
				}

				if !reflect.DeepEqual(fakeExecRunner.LastRunArgs(), tc.expectedKubectlArgs) {
					t.Fatalf("wrong kubectl args.\nexpected: %v\ngot: %v\n", tc.expectedKubectlArgs, fakeExecRunner.LastRunArgs())
				}
			} else {
				if fakeExecRunner.RunCount() > 0 {
					t.Fatalf("unexpected run %q with args = %v", fakeExecRunner.LastRunName(), fakeExecRunner.LastRunArgs())
				}
			}

			if !strings.Contains(stdout.String(), tc.expectedStdout) {
				t.Fatalf("expected stdout to contain %q, but it did not", tc.expectedStdout)
			}

			if tc.unexpectedStdout != "" && strings.Contains(stdout.String(), tc.unexpectedStdout) {
				t.Fatalf("expected stdout not to contain %q, but it did", tc.unexpectedStdout)
			}

			if !strings.Contains(stderr.String(), tc.expectedStderr) {
				t.Fatalf("expected stderr to contain %q, but it did not", tc.expectedStderr)
			}

			if actualExitCode != tc.expectedExitCode {
				t.Fatalf("wrong exit code. expected: %d, got %d", tc.expectedExitCode, actualExitCode)
			}
		})
	}
}
