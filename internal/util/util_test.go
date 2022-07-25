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
	"os"
	"testing"
)

func TestGetKubectlPath(t *testing.T) {
	if path := GetKubectlPath(); path != "kubectl" {
		t.Fatalf("expected getKubectlPath to return \"kubectl\", but it was %q", path)
	}

	_ = os.Setenv("KUBECTL_PATH", "foo")
	defer os.Unsetenv("KUBECTL_PATH")
	if path := GetKubectlPath(); path != "foo" {
		t.Fatalf("expected getKubectlPath to return \"foo\", but it was %q", path)
	}
}
