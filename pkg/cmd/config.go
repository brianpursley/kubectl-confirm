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
	"encoding/json"

	"github.com/spf13/cobra"

	"github.com/brianpursley/kubectl-confirm/internal/util"
)

func (o *confirmOptions) printConfig(cmd *cobra.Command) error {
	util.PrintSectionTitle(cmd, "Config")
	defer cmd.Println()

	// Get the effective config
	stdout := bytes.Buffer{}
	err := util.ExecRun(util.GetKubectlPath(), []string{"config", "view", "-o=json"}, cmd.InOrStdin(), &stdout, cmd.ErrOrStderr())
	if err != nil {
		return err
	}
	var config map[string]interface{}
	err = json.Unmarshal(stdout.Bytes(), &config)
	if err != nil {
		return err
	}

	context := o.context
	if len(context) == 0 {
		context = config["current-context"].(string)
	}
	cmd.Printf("%-11s %s\n", "Context:", context)

	contextConfig := findContextInConfig(config, context)

	cluster := o.cluster
	if len(cluster) == 0 && contextConfig != nil && contextConfig["cluster"] != nil {
		cluster = contextConfig["cluster"].(string)
	}
	cmd.Printf("%-11s %s\n", "Cluster:", cluster)

	user := o.user
	if len(user) == 0 && contextConfig != nil && contextConfig["user"] != nil {
		user = contextConfig["user"].(string)
	}
	cmd.Printf("%-11s %s\n", "User:", user)

	namespace := o.namespace
	if len(namespace) == 0 && contextConfig != nil && contextConfig["namespace"] != nil {
		namespace = contextConfig["namespace"].(string)
	}
	if len(namespace) == 0 {
		namespace = "default"
	}
	cmd.Printf("%-11s %s\n", "Namespace:", namespace)

	return nil
}

func findContextInConfig(config map[string]interface{}, context string) map[string]interface{} {
	for _, c := range config["contexts"].([]interface{}) {
		cc := c.(map[string]interface{})
		if cc["name"] == context {
			return cc["context"].(map[string]interface{})
		}
	}
	return nil
}
