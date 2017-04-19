// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"github.com/spf13/cobra"
)

// timelinesCmd represents the timelines command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Display current instance information",
	RunE:  instanceRunE,
}

func init() {
	RootCmd.AddCommand(instanceCmd)
}

func instanceRunE(cmd *cobra.Command, args []string) error {
	if err := madonInit(false); err != nil {
		return err
	}

	i, err := gClient.GetCurrentInstance()
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(i, nil, "")
}
