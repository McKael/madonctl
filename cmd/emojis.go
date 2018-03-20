// Copyright © 2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var emojiOpts struct {
	// Used for several subcommands to limit the number of results
	limit, keep uint
	//sinceID, maxID int64
	all bool
}

//emojiCmd represents the emoji command
var emojiCmd = &cobra.Command{
	Use:     "emojis",
	Aliases: []string{"emoji"},
	Short:   "Display server emojis",
	RunE:    emojiGetRunE, // Defaults to list
}

func init() {
	RootCmd.AddCommand(emojiCmd)

	// Subcommands
	emojiCmd.AddCommand(emojiSubcommands...)

	emojiGetCustomSubcommand.Flags().UintVarP(&emojiOpts.limit, "limit", "l", 0, "Limit number of API results")
	emojiGetCustomSubcommand.Flags().UintVarP(&emojiOpts.keep, "keep", "k", 0, "Limit number of results")
	emojiGetCustomSubcommand.Flags().BoolVar(&emojiOpts.all, "all", false, "Fetch all results")
}

var emojiSubcommands = []*cobra.Command{
	emojiGetCustomSubcommand,
}

var emojiGetCustomSubcommand = &cobra.Command{
	Use:     "list",
	Short:   "Display the custom emojis (default subcommand)",
	Long:    `Display the list of custom emojis of the instance.`,
	Aliases: []string{"get", "display", "show"},
	RunE:    emojiGetRunE,
}

func emojiGetRunE(cmd *cobra.Command, args []string) error {
	opt := emojiOpts

	// Set up LimitParams
	var limOpts *madon.LimitParams
	if opt.all || opt.limit > 0 {
		limOpts = new(madon.LimitParams)
		limOpts.All = opt.all
	}
	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}

	// We don't have to log in
	if err := madonInit(false); err != nil {
		return err
	}

	var obj interface{}
	var err error

	var emojiList []madon.Emoji
	emojiList, err = gClient.GetCustomEmojis(limOpts)

	if opt.keep > 0 && len(emojiList) > int(opt.keep) {
		emojiList = emojiList[:opt.keep]
	}

	obj = emojiList

	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	if obj == nil {
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.printObj(obj)
}
