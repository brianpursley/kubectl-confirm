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
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/brianpursley/kubectl-confirm/internal/util"
)

func (o *confirmOptions) diff(cmd *cobra.Command) error {
	util.PrintSectionTitle(cmd, "Diff")
	defer cmd.Println()

	if o.hasAnyNonRegularFiles {
		cmd.Println("*** Skipped because one or more non-regular files were specified ***")
		return nil
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	err := util.ExecRun(util.GetKubectlPath(), append(os.Args[1:], "--dry-run=server", "--output=yaml"), cmd.InOrStdin(), &stdout, &stderr)
	if err != nil {
		return fmt.Errorf("%s", stderr.String())
	}

	f, err := os.CreateTemp("", "kubectl-confirm-")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(stdout.Bytes()); err != nil {
		return err
	}

	// Run kubectl diff
	err = util.ExecRun(util.GetKubectlPath(), []string{"diff", "--filename", f.Name()}, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
	if err == nil {
		cmd.Println("no changes detected")
	}
	return nil
}
