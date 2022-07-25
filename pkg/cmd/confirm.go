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
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/brianpursley/kubectl-confirm/internal/util"
	"github.com/brianpursley/kubectl-confirm/internal/version"
)

// Commands that have --dry-run flag
var dryRunCommands = map[string]bool{
	"annotate":  true,
	"apply":     true,
	"autoscale": true,
	"cordon":    true,
	"create":    true,
	"delete":    true,
	"drain":     true,
	"expose":    true,
	"label":     true,
	"patch":     true,
	"replace":   true,
	"run":       true,
	"scale":     true,
	"set":       true,
	"taint":     true,
	"uncordon":  true,
}

// Commands that have both --dry-run and --output flags
var diffCommands = map[string]bool{
	"annotate":  true,
	"apply":     true,
	"autoscale": true,
	"create":    true,
	"expose":    true,
	"label":     true,
	"patch":     true,
	"replace":   true,
	"run":       true,
	"scale":     true,
	"set":       true,
	"taint":     true,
}

type confirmOptions struct {
	cluster   string
	context   string
	namespace string
	user      string

	filenames []string
	kustomize string

	hasAnyNonRegularFiles bool
}

const shortHelpText string = `
The Kubectl Confirm plugin prints information and prompts you to confirm whether to continue before running a Kubectl command. 
`

const longHelpText = shortHelpText + `
The plugin will show the following information:

  * Configuration (context, cluster, user, and namespace)
  * Dry run output (if available for the kubectl command)
  * Diff output (if available for the kubectl command)

After the information is displayed, you will be asked to confirm whether to proceed.

Upon confirmation, the Kubectl command will be executed. 

All arguments and flags will be passed through to Kubectl.
`

// NewConfirmCommand returns a cobra command for the confirm command
func NewConfirmCommand() *cobra.Command {
	options := confirmOptions{}

	var executableName = filepath.Base(os.Args[0])
	if strings.HasPrefix(executableName, "kubectl-") {
		executableName = "kubectl"
	}
	usage := fmt.Sprintf("%s confirm [command] [flags] [options]", executableName)

	cmd := cobra.Command{
		SilenceUsage: true,
		Short:        shortHelpText,
		Long:         longHelpText,
		Use:          usage,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return options.run(cmd, args)
		},
	}

	cmd.Flags().StringVar(&options.cluster, "cluster", "", "")
	_ = cmd.Flags().MarkHidden("cluster")
	cmd.Flags().StringVar(&options.context, "context", "", "")
	_ = cmd.Flags().MarkHidden("context")
	cmd.Flags().StringVarP(&options.namespace, "namespace", "n", "", "")
	_ = cmd.Flags().MarkHidden("namespace")
	cmd.Flags().StringVar(&options.user, "user", "", "")
	_ = cmd.Flags().MarkHidden("user")

	cmd.Flags().StringArrayVarP(&options.filenames, "filename", "f", []string{}, "")
	_ = cmd.Flags().MarkHidden("filename")
	cmd.Flags().StringVarP(&options.kustomize, "kustomize", "k", "", "")
	_ = cmd.Flags().MarkHidden("kustomize")

	return &cmd
}

func (o *confirmOptions) run(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("expected at least one argument")
	}

	commandName := args[0]

	// Help
	if commandName == "help" {
		return cmd.Help()
	}

	// Version
	if commandName == "version" {
		if !util.HasOutputFlag() {
			cmd.Printf("Kubectl Confirm Plugin Version: %s\n\n", version.String())
		}
		return util.ExecRun(util.GetKubectlPath(), os.Args[1:], cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
	}

	// Check for non-regular files (ie. process substitution). In this case, dry run and diff cannot be performed, or
	// else they will consume the file stream and the real execution will fail. This check sets a flag on the options
	// indicating that one or more non-regular files were detected.
	o.checkForNonRegularFiles()

	// Config
	if err := o.printConfig(cmd); err != nil {
		return err
	}

	// Dry Run
	if dryRunCommands[commandName] {
		err := o.dryRun(cmd)
		if err != nil {
			return err
		}
	}

	// Diff
	if diffCommands[commandName] {
		if err := o.diff(cmd); err != nil {
			return err
		}
	}

	// Prompt
	util.PrintSectionTitle(cmd, "Confirm")
	cmd.Printf("The following command will be executed:\n%s %s\n\n", util.GetKubectlPath(), strings.Join(os.Args[1:], " "))
	cmd.Printf("Enter 'yes' to continue: ")
	var response string
	_, _ = fmt.Fscanln(cmd.InOrStdin(), &response)
	cmd.Println()
	if response != "yes" {
		cmd.PrintErr("Command aborted.\n")
		util.Exit(1)
		return nil
	}

	// Execute the real command
	return util.ExecRun(util.GetKubectlPath(), os.Args[1:], cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
}

func (o *confirmOptions) checkForNonRegularFiles() {
	o.hasAnyNonRegularFiles = false
	if len(o.kustomize) > 0 && util.IsNonRegularFile(o.kustomize) {
		o.hasAnyNonRegularFiles = true
		return
	}
	for _, f := range o.filenames {
		if util.IsNonRegularFile(f) {
			o.hasAnyNonRegularFiles = true
			return
		}
	}
}
