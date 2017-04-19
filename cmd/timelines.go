// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"github.com/spf13/cobra"
)

var timelineOpts struct {
	local bool
}

// timelineCmd represents the timelines command
var timelineCmd = &cobra.Command{
	Use:     "timeline [home|public|:HASHTAG] [--local]",
	Aliases: []string{"tl"},
	Short:   "Fetch a timeline",
	Long: `
The timeline command fetches a timeline (home, local or federated).
It can also get a hashtag-based timeline if the keyword or prefixed with
':' or '#'.`,
	Example: `  madonctl timeline
  madonctl timeline public --local
  madonctl timeline :mastodon`,
	RunE:      timelineRunE,
	ValidArgs: []string{"home", "public"},
}

func init() {
	RootCmd.AddCommand(timelineCmd)

	timelineCmd.Flags().BoolVar(&timelineOpts.local, "local", false, "Posts from the local instance")
}

func timelineRunE(cmd *cobra.Command, args []string) error {
	opt := timelineOpts

	tl := "home"
	if len(args) > 0 {
		tl = args[0]
	}

	// The home timeline is the only one requiring to be logged in
	if err := madonInit(tl == "home"); err != nil {
		return err
	}

	sl, err := gClient.GetTimelines(tl, opt.local)
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(sl, nil, "")
}
