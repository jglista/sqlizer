/*
Package cmd

Copyright © 2021 Joe Glista <josephsglista@gmail.com>

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

	"github.com/AlecAivazis/survey/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage sqlizer config",
	Long:  `config contains commands for quickly interacting with and changing your config`,
	Run:   func(cmd *cobra.Command, args []string) {},
}

type initAnswers struct {
	Host         string
	Port         int
	DatabaseName string
	User         string
	Password     string
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "initializes sqlizer config",
	Long:  "",
	Run: func(cmd *cobra.Command, args []string) {
		aws := initAnswers{}
		qs := []*survey.Question{
			{
				Name: "Host",
				Prompt: &survey.Input{
					Message: "Enter the host name for your SQL Server instance",
				},
			},
			{
				Name: "Port",
				Prompt: &survey.Input{
					Message: "Enter the port number for your SQL Server instance",
				},
			},
			{
				Name: "User",
				Prompt: &survey.Input{
					Message: "Target database user?",
				},
			},
			{
				Name: "Password",
				Prompt: &survey.Input{
					Message: "Target database pass?",
				},
			},
		}
		survey.Ask(qs, &aws, nil)

		viper.Set("server.host", aws.Host)
		viper.Set("server.port", aws.Port)
		viper.Set("server.user", aws.User)
		viper.Set("server.pass", aws.Password)
		err := viper.SafeWriteConfig()
		if err != nil {
			fmt.Fprintf(os.Stderr, "While writing config: %s", err.Error())
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// configCmd.PersistentFlags().String("foo", "", "A help for foo")
	configCmd.AddCommand(initCmd)

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// configCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
