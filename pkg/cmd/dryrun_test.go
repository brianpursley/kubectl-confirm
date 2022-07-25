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

func TestDryRun(t *testing.T) {
	testCases := []struct {
		name                string
		options             confirmOptions
		expectKubectl       bool
		fakeOsArgs          []string
		expectedKubectlArgs []string
		fakeKubectlStdout   string
		fakeKubectlStderr   string
		fakeKubectlError    error
		expectedStdout      string
		expectedError       error
	}{
		{
			name: "non regular files",
			options: confirmOptions{
				hasAnyNonRegularFiles: true,
			},
			fakeOsArgs:    []string{"confirm", "foo", "bar", "--baz"},
			expectKubectl: false,
			expectedStdout: `========== Dry Run ==========
*** Skipped because one or more non-regular files were specified ***

`,
		},
		{
			name:                "successful dry run",
			options:             confirmOptions{},
			expectKubectl:       true,
			fakeOsArgs:          []string{"confirm", "foo", "bar", "--baz"},
			expectedKubectlArgs: []string{"foo", "bar", "--baz", "--dry-run=server"},
			fakeKubectlStdout:   "fake dry run output\n",
			expectedStdout: `========== Dry Run ==========
fake dry run output

`,
		},
		{
			name:                "error",
			options:             confirmOptions{},
			expectKubectl:       true,
			fakeOsArgs:          []string{"confirm", "foo", "bar", "--baz"},
			expectedKubectlArgs: []string{"foo", "bar", "--baz", "--dry-run=server"},
			fakeKubectlError:    fmt.Errorf("exit status 1"),
			fakeKubectlStderr:   "unknown flag: --dry-run",
			expectedStdout: `========== Dry Run ==========

`,
			expectedError: fmt.Errorf("unknown flag: --dry-run"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeExecRunner := util.NewFakeExecRunner()
			fakeExecRunner.SetupRun(tc.fakeKubectlStdout, tc.fakeKubectlStderr, tc.fakeKubectlError)

			os.Args = tc.fakeOsArgs
			cmd, _, stdout, stderr := util.NewTestCommand()

			err := tc.options.dryRun(cmd)
			if tc.expectedError == nil {
				if err != nil {
					t.Fatalf("dryRun failed: %v", err)
				}
			} else {
				if err == nil {
					t.Fatalf("expected an error, but no error was returned.\nExpected: %v\n", tc.expectedError)
				} else if err.Error() != tc.expectedError.Error() {
					t.Fatalf("wrong error returned.\nExpected: %v\nGot: %v\n", tc.expectedError, err.Error())
				}
			}

			if tc.expectKubectl {
				if fakeExecRunner.LastRunName() != "kubectl" {
					t.Fatalf("expected kubectl to be run, but it was not")
				}

				if !reflect.DeepEqual(fakeExecRunner.LastRunArgs(), tc.expectedKubectlArgs) {
					t.Fatalf("wrong kubectl args.\nexpected: %v\ngot: %v\n", tc.expectedKubectlArgs, fakeExecRunner.LastRunArgs())
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
