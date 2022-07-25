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
	"github.com/spf13/cobra"
	"os"

	"github.com/brianpursley/kubectl-confirm/internal/util"
)

func (o *confirmOptions) dryRun(cmd *cobra.Command) error {
	util.PrintSectionTitle(cmd, "Dry Run")
	defer cmd.Println()

	if o.hasAnyNonRegularFiles {
		cmd.Println("*** Skipped because one or more non-regular files were specified ***")
		return nil
	}

	stdout := bytes.Buffer{}
	stderr := bytes.Buffer{}

	err := util.ExecRun(util.GetKubectlPath(), append(os.Args[1:], "--dry-run=server"), cmd.InOrStdin(), &stdout, &stderr)
	if err != nil {
		return fmt.Errorf("%s", stderr.String())
	}

	cmd.Print(stdout.String())
	return nil
}
