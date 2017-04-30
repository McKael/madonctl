// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/m0t0k1ch1/gomif"
	"github.com/spf13/cobra"

	"github.com/McKael/madonctl/printer"
)

var instanceOpts struct {
	stats  bool
	server string
}

// timelinesCmd represents the timelines command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Display current instance information",
	RunE:  instanceRunE,
	Example: `  madonctl instance
  madonctl instance -i mastodon.social
  madonctl instance --stats
  madonctl instance --stats --server mastodon.social --template '{{.Users}}'`,
}

func init() {
	RootCmd.AddCommand(instanceCmd)

	instanceCmd.Flags().BoolVar(&instanceOpts.stats, "stats", false, "Display server statistics (from instances.mastodon.xyz)")
	instanceCmd.Flags().StringVar(&instanceOpts.server, "server", "", "Display statistics for a specific server (for --stats)")
}

func instanceRunE(cmd *cobra.Command, args []string) error {
	opt := instanceOpts

	p, err := getPrinter()
	if err != nil {
		return err
	}

	if opt.stats {
		// Get instance statistics using gomif
		if opt.server == "" {
			if err := madonInit(false); err != nil {
				return err
			}
			opt.server = strings.TrimLeft(gClient.InstanceURL, "https://")
		}
		is, err := instanceFetchStatus(opt.server)
		if err != nil {
			errPrint("Error: %s", err.Error())
			os.Exit(1)
		}
		if is == nil {
			return nil
		}
		istats := &printer.InstanceStatistics{
			InstanceName:   opt.server,
			InstanceStatus: *is,
		}
		return p.PrintObj(istats, nil, "")
	}

	// Get current instance data through the API
	if err := madonInit(false); err != nil {
		return err
	}
	i, err := gClient.GetCurrentInstance()
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	return p.PrintObj(i, nil, "")
}

func instanceFetchStatus(server string) (*gomif.InstanceStatus, error) {
	if server == "" {
		return nil, errors.New("no instance server name")
	}

	client := gomif.NewClient()

	return client.FetchLastInstanceStatus(
		context.Background(),
		server,
		3600, // span (sec)
	)
}
