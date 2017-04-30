// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var timelineOpts struct {
	local          bool
	limit          uint
	sinceID, maxID int64
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
	timelineCmd.Flags().UintVarP(&timelineOpts.limit, "limit", "l", 0, "Limit number of results")
	timelineCmd.PersistentFlags().Int64Var(&timelineOpts.sinceID, "since-id", 0, "Request IDs greater than a value")
	timelineCmd.PersistentFlags().Int64Var(&timelineOpts.maxID, "max-id", 0, "Request IDs less (or equal) than a value")
}

func timelineRunE(cmd *cobra.Command, args []string) error {
	opt := timelineOpts
	var limOpts *madon.LimitParams

	if opt.limit > 0 || opt.sinceID > 0 || opt.maxID > 0 {
		limOpts = new(madon.LimitParams)
	}

	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}
	if opt.maxID > 0 {
		limOpts.MaxID = opt.maxID
	}
	if opt.sinceID > 0 {
		limOpts.SinceID = opt.sinceID
	}

	tl := "home"
	if len(args) > 0 {
		tl = args[0]
	}

	// The home timeline is the only one requiring to be logged in
	if err := madonInit(tl == "home"); err != nil {
		return err
	}

	sl, err := gClient.GetTimelines(tl, opt.local, limOpts)
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	if opt.limit > 0 && len(sl) > int(opt.limit) {
		sl = sl[:opt.limit]
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(sl, nil, "")
}
