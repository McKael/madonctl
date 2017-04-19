// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

var searchOpts struct {
	resolve bool
}

// searchCmd represents the search command
var searchCmd = &cobra.Command{
	Use:   "search [--resolve] STRING",
	Short: "Search for contents (accounts or statuses)",
	//Long: `TBW...`,
	RunE: searchRunE,
}

func init() {
	RootCmd.AddCommand(searchCmd)

	searchCmd.Flags().BoolVar(&searchOpts.resolve, "resolve", false, "Resolve non-local accounts")
}

func searchRunE(cmd *cobra.Command, args []string) error {
	opt := searchOpts

	if len(args) == 0 {
		return errors.New("no search string provided")
	}

	if err := madonInit(true); err != nil {
		return err
	}

	results, err := gClient.Search(strings.Join(args, " "), opt.resolve)
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(results, nil, "")
}
