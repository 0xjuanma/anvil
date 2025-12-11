/*
Copyright © 2022 Juanma Roca juanmaxroca@gmail.com

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

// Package config provides subcommands for managing configuration files and assets.
package config

import (
	importcmd "github.com/0xjuanma/anvil/cmd/config/import"
	"github.com/0xjuanma/anvil/cmd/config/pull"
	"github.com/0xjuanma/anvil/cmd/config/push"
	"github.com/0xjuanma/anvil/cmd/config/show"
	"github.com/0xjuanma/anvil/cmd/config/sync"
	"github.com/0xjuanma/anvil/internal/constants"
	"github.com/spf13/cobra"
)

var ConfigCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage configuration files and assets",
	Long:  constants.CONFIG_COMMAND_LONG_DESCRIPTION,
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

func init() {
	// Add pull, push, show, sync, and import as sub-commands of config
	ConfigCmd.AddCommand(pull.PullCmd)
	ConfigCmd.AddCommand(push.PushCmd)
	ConfigCmd.AddCommand(show.ShowCmd)
	ConfigCmd.AddCommand(sync.SyncCmd)
	ConfigCmd.AddCommand(importcmd.ImportCmd)
}
