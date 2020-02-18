/*
Copyright © 2020 NAME HERE <EMAIL ADDRESS>

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
	"github.com/spf13/cobra"
	"os"
)

var (
	newRegistry  string
	newNamespace string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "relocate_chart",
	Short: "A tool for relocating images in a helm chart",
	Long: `Rewrites the default authored repository and namespace for images in a given helm chart tarball.

The chart and its sub-charts must be compliant with the image referencing convention.
See https://github.com/helm/helm/issues/7154 for more details.`,
	Example: `relocate_chart /path/to/chart.tgz --registry new.registry.local --namespace new-namespace

This above command will set the image paths in the chart.tgz to new.registry.local/new-namespace/`,
	RunE: relocate,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("a path to the chart tarball is required")
		}
		return ensurePathToAFile(args[0])
	},
}

func ensurePathToAFile(path string) error {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileInfo.IsDir() {
		return fmt.Errorf("%s is a directory. Expecting a helm chart tarball", path)
	}
	return nil
}

func relocate(_ *cobra.Command, args []string) error {
	return relocateChart(args[0])
}


// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVarP(&newRegistry, "registry", "r", "", "New registry to use in the chart")
	rootCmd.Flags().StringVarP(&newNamespace, "namespace", "n", "", "New namespace to use in the chart")
}