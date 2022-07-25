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
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/brianpursley/kubectl-confirm/internal/util"
)

func TestDiff(t *testing.T) {
	testCases := []struct {
		name                      string
		options                   confirmOptions
		osArgs                    []string
		expectKubectlRun          bool
		expectedKubectlDryRunArgs []string
		fakeKubectlDiffStdout     string
		expectedStdout            string
	}{
		{
			name: "non regular files",
			options: confirmOptions{
				hasAnyNonRegularFiles: true,
			},
			expectKubectlRun: false,
			expectedStdout: `========== Diff =============
*** Skipped because one or more non-regular files were specified ***

`,
		},
		{
			name:                      "single file no changes",
			osArgs:                    []string{"kubectl-confirm", "apply", "--filename", "foo.yaml"},
			expectKubectlRun:          true,
			expectedKubectlDryRunArgs: []string{"apply", "--filename", "foo.yaml", "--dry-run=server", "--output=yaml"},
			fakeKubectlDiffStdout:     "",
			expectedStdout: `========== Diff =============
no changes detected

`,
		},
		{
			name:                      "single file with changes",
			osArgs:                    []string{"kubectl-confirm", "apply", "--filename", "foo.yaml"},
			expectKubectlRun:          true,
			expectedKubectlDryRunArgs: []string{"apply", "--filename", "foo.yaml", "--dry-run=server", "--output=yaml"},
			fakeKubectlDiffStdout:     "fake diff output\n",
			expectedStdout: `========== Diff =============
fake diff output

`,
		},
		{
			name:                      "multiple files",
			osArgs:                    []string{"kubectl-confirm", "apply", "--filename", "foo.yaml", "--filename", "bar.yaml"},
			expectKubectlRun:          true,
			expectedKubectlDryRunArgs: []string{"apply", "--filename", "foo.yaml", "--filename", "bar.yaml", "--dry-run=server", "--output=yaml"},
			fakeKubectlDiffStdout:     "fake diff output\n",
			expectedStdout: `========== Diff =============
fake diff output

`,
		},
		{
			name:                      "recursive",
			osArgs:                    []string{"kubectl-confirm", "apply", "--filename", "foo/", "--recursive"},
			expectKubectlRun:          true,
			expectedKubectlDryRunArgs: []string{"apply", "--filename", "foo/", "--recursive", "--dry-run=server", "--output=yaml"},
			fakeKubectlDiffStdout:     "fake diff output\n",
			expectedStdout: `========== Diff =============
fake diff output

`,
		},
		{
			name:                      "kustomize",
			osArgs:                    []string{"kubectl-confirm", "apply", "--kustomize", "foo/"},
			expectKubectlRun:          true,
			expectedKubectlDryRunArgs: []string{"apply", "--kustomize", "foo/", "--dry-run=server", "--output=yaml"},
			fakeKubectlDiffStdout:     "fake diff output\n",
			expectedStdout: `========== Diff =============
fake diff output

`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var fakeError error
			if tc.fakeKubectlDiffStdout != "" {
				fakeError = fmt.Errorf("exit status 1")
			}
			fakeExecRunner := util.NewFakeExecRunner()
			fakeExecRunner.SetupRun("fake dry run output", "", nil)          // Dry run
			fakeExecRunner.SetupRun(tc.fakeKubectlDiffStdout, "", fakeError) // Diff

			cmd, _, stdout, stderr := util.NewTestCommand()

			os.Args = tc.osArgs

			err := tc.options.diff(cmd)
			if err != nil {
				t.Fatalf("diff failed: %v", err)
			}

			if tc.expectKubectlRun {
				if fakeExecRunner.RunNames[0] != "kubectl" {
					t.Fatalf("expected kubectl to be run, but it was not")
				}
				if !reflect.DeepEqual(fakeExecRunner.RunArgs[0], tc.expectedKubectlDryRunArgs) {
					t.Fatalf("wrong kubectl args.\nexpected: %v\ngot: %v\n", tc.expectedKubectlDryRunArgs, fakeExecRunner.RunArgs[0])
				}

				if fakeExecRunner.RunNames[1] != "kubectl" {
					t.Fatalf("expected kubectl to be run, but it was not")
				}
				if fakeExecRunner.RunArgs[1][0] != "diff" {
					t.Fatalf("expected kubectl diff to be called, but it was not")
				}
			}

			if stdout.String() != tc.expectedStdout {
				t.Fatalf("wrong stdout\nexpected:\n%s\ngot:\n%s\n", tc.expectedStdout, stdout.String())
			}

			if stderr.Len() > 0 {
				t.Fatalf("unexpected stderr:\n%s", stderr.String())
			}
		})
	}
}
