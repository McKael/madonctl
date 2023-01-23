// Copyright © 2017-2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon/v2"
)

var domainBlocksOpts struct {
	show, block, unblock bool

	limit          uint              // Limit the results
	sinceID, maxID madon.ActivityID // Query boundaries
	all            bool              // Try to fetch all results
}

// timelinesCmd represents the timelines command
var domainBlocksCmd = &cobra.Command{
	Use:     "domain-blocks --show|--block|--unblock [DOMAINNAME]",
	Aliases: []string{"domain-block"},
	Short:   "Display, add or remove user-blocked domains",
	RunE:    domainBlocksRunE,
	Example: `  madonctl domain-blocks --show
  madonctl domain-blocks --block   example.com
  madonctl domain-blocks --unblock example.com`,
}

func init() {
	RootCmd.AddCommand(domainBlocksCmd)

	domainBlocksCmd.Flags().BoolVar(&domainBlocksOpts.show, "show", false, "List current user-blocked domains")
	domainBlocksCmd.Flags().BoolVar(&domainBlocksOpts.block, "block", false, "Block domain")
	domainBlocksCmd.Flags().BoolVar(&domainBlocksOpts.unblock, "unblock", false, "Unblock domain")

	domainBlocksCmd.Flags().UintVarP(&domainBlocksOpts.limit, "limit", "l", 0, "Limit number of results")
	domainBlocksCmd.Flags().StringVar(&domainBlocksOpts.sinceID, "since-id", "", "Request IDs greater than a value")
	domainBlocksCmd.Flags().StringVar(&domainBlocksOpts.maxID, "max-id", "", "Request IDs less (or equal) than a value")
	domainBlocksCmd.Flags().BoolVar(&domainBlocksOpts.all, "all", false, "Fetch all results")
}

func domainBlocksRunE(cmd *cobra.Command, args []string) error {
	opt := domainBlocksOpts
	var domName madon.DomainName

	// Check flags
	if opt.block && opt.unblock {
		return errors.New("cannot use both --block and --unblock")
	}

	if opt.block || opt.unblock {
		if opt.show {
			return errors.New("cannot use both --[un]block and --show")
		}
		if len(args) != 1 {
			return errors.New("missing domain name")
		}
		domName = madon.DomainName(args[0])
	}

	if !opt.show && !opt.block && !opt.unblock {
		return errors.New("missing flag: please provide --show, --block or --unblock")
	}

	// Set up LimitParams
	var limOpts *madon.LimitParams
	if opt.all || opt.limit > 0 || opt.sinceID != "" || opt.maxID != "" {
		limOpts = new(madon.LimitParams)
		limOpts.All = opt.all
	}
	if opt.limit > 0 {
		limOpts.Limit = int(opt.limit)
	}
	if opt.maxID != "" {
		limOpts.MaxID = opt.maxID
	}
	if opt.sinceID != "" {
		limOpts.SinceID = opt.sinceID
	}

	// Log in
	if err := madonInit(true); err != nil {
		return err
	}

	var obj interface{}
	var err error

	switch {
	case opt.show:
		var domainList []madon.DomainName
		domainList, err = gClient.GetBlockedDomains(limOpts)
		obj = domainList
	case opt.block:
		err = gClient.BlockDomain(domName)
	case opt.unblock:
		err = gClient.UnblockDomain(domName)
	default:
		return errors.New("domainBlocksCmd: internal error")
	}

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
