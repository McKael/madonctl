// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// VERSION of the madonctl application
var VERSION = "0.2.0"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display " + AppName + " version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("This is %s version %s\n", AppName, VERSION)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
