/*
Package cmd ...

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
	"errors"
	"fmt"
	"os"
	"sort"

	"github.com/AlecAivazis/survey/v2"
	"github.com/jglista/sqlizer/internal/config"
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage sqlizer config",
	Long:  `config contains commands for managing saved database connections`,
}

var configAddCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new database connection",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		var name string
		if err := survey.AskOne(&survey.Input{Message: "Connection name:"}, &name, survey.WithValidator(survey.Required)); err != nil {
			return err
		}

		existing, listErr := config.List(path)
		if listErr != nil {
			return listErr
		}
		if _, exists := existing[name]; exists {
			var overwrite bool
			if err := survey.AskOne(&survey.Confirm{
				Message: fmt.Sprintf("Connection %q already exists. Overwrite?", name),
				Default: false,
			}, &overwrite); err != nil {
				return err
			}
			if !overwrite {
				fmt.Fprintln(os.Stderr, "Aborted.")
				return nil
			}
		}

		var host, port, user, pass string
		qs := []*survey.Question{
			{Name: "host", Prompt: &survey.Input{Message: "Host:"}, Validate: survey.Required},
			{Name: "port", Prompt: &survey.Input{Message: "Port:", Default: "1433"}, Validate: survey.Required},
			{Name: "user", Prompt: &survey.Input{Message: "User:"}, Validate: survey.Required},
			{Name: "pass", Prompt: &survey.Password{Message: "Password:"}},
		}
		answers := struct {
			Host string
			Port string
			User string
			Pass string
		}{}
		if err := survey.Ask(qs, &answers); err != nil {
			return err
		}
		host, port, user, pass = answers.Host, answers.Port, answers.User, answers.Pass

		conn := config.Connection{Host: host, Port: port, User: user, Pass: pass}
		if err := config.Overwrite(path, name, conn); err != nil {
			return err
		}
		fmt.Printf("Connection %q saved.\n", name)

		var makeDefault bool
		if err := survey.AskOne(&survey.Confirm{
			Message: fmt.Sprintf("Set %q as the default connection?", name),
			Default: false,
		}, &makeDefault); err != nil {
			return err
		}
		if makeDefault {
			if err := config.SetDefault(path, name); err != nil {
				return err
			}
			fmt.Printf("%q set as default.\n", name)
		}
		return nil
	},
}

var configListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all saved connections",
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		connections, err := config.List(path)
		if err != nil {
			return err
		}
		defaultName, err := config.GetDefault(path)
		if err != nil {
			return err
		}

		if len(connections) == 0 {
			fmt.Println("No connections saved. Run `sqlizer config add` to add one.")
			return nil
		}

		names := make([]string, 0, len(connections))
		for n := range connections {
			names = append(names, n)
		}
		sort.Strings(names)

		for _, n := range names {
			conn := connections[n]
			marker := "  "
			if n == defaultName {
				marker = "* "
			}
			fmt.Printf("%s%s  %s:%s\n", marker, n, conn.Host, conn.Port)
		}
		return nil
	},
}

var configRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove a saved connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		defaultName, err := config.GetDefault(path)
		if err != nil {
			return err
		}

		if err := config.Remove(path, name); err != nil {
			return err
		}
		fmt.Printf("Connection %q removed.\n", name)

		if name == defaultName {
			fmt.Fprintf(os.Stderr, "Warning: %q was the default connection. Run `sqlizer config set-default <name>` to set a new default.\n", name)
		}
		return nil
	},
}

var configSetDefaultCmd = &cobra.Command{
	Use:   "set-default <name>",
	Short: "Set the default connection",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		name := args[0]
		path, err := config.ConfigPath()
		if err != nil {
			return err
		}

		if err := config.SetDefault(path, name); err != nil {
			if errors.Is(err, config.ErrNotFound) {
				return fmt.Errorf("connection %q not found; run `sqlizer config list` to see available connections", name)
			}
			return err
		}
		fmt.Printf("%q set as default.\n", name)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configAddCmd)
	configCmd.AddCommand(configListCmd)
	configCmd.AddCommand(configRemoveCmd)
	configCmd.AddCommand(configSetDefaultCmd)
}
