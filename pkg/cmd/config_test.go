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
	"reflect"
	"testing"

	"github.com/brianpursley/kubectl-confirm/internal/util"
)

func TestPrintConfig(t *testing.T) {
	fakeStdout := `
{
	"current-context": "foo",
	"contexts": [
		{
			"name": "foo",
			"context": {
                "cluster": "foo-cluster",
                "user": "foo-user",
                "namespace": "foo-namespace"
			}
		},
		{
			"name": "bar",
			"context": {
                "cluster": "bar-cluster",
                "user": "bar-user",
                "namespace": "bar-namespace"
			}
		},
		{
			"name": "baz",
			"context": {
                "cluster": "baz-cluster",
                "user": "baz-user"
			}
		}
	]
}
`
	testCases := []struct {
		name           string
		options        confirmOptions
		expectedStdout string
	}{
		{
			name:    "no options set should use current context from config",
			options: confirmOptions{},
			expectedStdout: `========== Config ===========
Context:    foo
Cluster:    foo-cluster
User:       foo-user
Namespace:  foo-namespace

`,
		},
		{
			name:    "context set in options",
			options: confirmOptions{context: "bar"},
			expectedStdout: `========== Config ===========
Context:    bar
Cluster:    bar-cluster
User:       bar-user
Namespace:  bar-namespace

`,
		},
		{
			name:    "default namespace",
			options: confirmOptions{context: "baz"},
			expectedStdout: `========== Config ===========
Context:    baz
Cluster:    baz-cluster
User:       baz-user
Namespace:  default

`,
		},
		{
			name: "override cluster, namespace, and user using options",
			options: confirmOptions{
				context:   "bar",
				cluster:   "override-cluster",
				namespace: "override-namespace",
				user:      "override-user",
			},
			expectedStdout: `========== Config ===========
Context:    bar
Cluster:    override-cluster
User:       override-user
Namespace:  override-namespace

`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fakeExecRunner := util.NewFakeExecRunner()
			fakeExecRunner.SetupRun(fakeStdout, "", nil)

			cmd, _, stdout, stderr := util.NewTestCommand()

			err := tc.options.printConfig(cmd)
			if err != nil {
				t.Fatalf("printConfig failed: %v", err)
			}

			if fakeExecRunner.LastRunName() != "kubectl" {
				t.Fatalf("expected kubectl to be run, but it was not")
			}

			expectedKubectlArgs := []string{"config", "view", "-o=json"}
			if !reflect.DeepEqual(fakeExecRunner.LastRunArgs(), expectedKubectlArgs) {
				t.Fatalf("wrong kubectl args.\nexpected: %v\ngot: %v\n", expectedKubectlArgs, fakeExecRunner.LastRunArgs())
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
