// Copyright © 2017-2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// timelinesCmd represents the timelines command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Display current instance information",
	Long: `Display instance information

This command display the instance information returned by the server.
`,
	RunE: instanceRunE,
}

func init() {
	RootCmd.AddCommand(instanceCmd)
}

func instanceRunE(cmd *cobra.Command, args []string) error {
	if err := madonInit(false); err != nil {
		return err
	}

	// Get current instance data through the API
	i, err := gClient.GetCurrentInstance()
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	return p.printObj(i)
}
