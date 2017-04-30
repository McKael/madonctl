// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"context"
	"os"
	"strings"

	"github.com/m0t0k1ch1/gomif"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var instanceOpts struct {
	stats      bool
	server     string
	start, end int64
}

// timelinesCmd represents the timelines command
var instanceCmd = &cobra.Command{
	Use:   "instance",
	Short: "Display current instance information",
	Long: `Display instance information

This command display the instance information returned by the server.

With '--stats', the instances.mastodon.xyz API is queried and instance
statistics will be returned (the instance server can be specified).
To get a range of statistics, both '--start' and '--end' should be provided
with UNIX timestamps (e.g. "date +%s").
`,
	RunE: instanceRunE,
	Example: `  madonctl instance
  madonctl instance -i mastodon.social
  madonctl instance --stats
  madonctl instance --stats --start 1493565000 --end 1493566000
  madonctl instance --stats --server mastodon.social --template '{{.Users}}'`,
}

func init() {
	RootCmd.AddCommand(instanceCmd)

	instanceCmd.Flags().BoolVar(&instanceOpts.stats, "stats", false, "Display server statistics (from instances.mastodon.xyz)")
	instanceCmd.Flags().StringVar(&instanceOpts.server, "server", "", "Display statistics for a specific server (for --stats)")
	instanceCmd.Flags().Int64Var(&instanceOpts.start, "start", 0, "Start timestamp (for --stats)")
	instanceCmd.Flags().Int64Var(&instanceOpts.end, "end", 0, "End timestamp (for --stats)")
}

func instanceRunE(cmd *cobra.Command, args []string) error {
	opt := instanceOpts

	if opt.stats {
		return instanceStats()
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

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(i, nil, "")
}

func instanceStats() error {
	opt := instanceOpts

	// Get instance statistics using gomif
	if opt.server == "" {
		if err := madonInit(false); err != nil {
			return err
		}
		opt.server = strings.TrimLeft(gClient.InstanceURL, "https://")
	}

	if opt.server == "" {
		return errors.New("no instance server name")
	}

	client := gomif.NewClient()
	var obj interface{}
	var err error

	if opt.start > 0 && opt.end > 0 {
		var isl []*gomif.InstanceStatus
		isl, err = client.FetchInstanceStatuses(
			context.Background(),
			opt.server, opt.start, opt.end,
		)
		obj = isl
	} else if opt.start > 0 || opt.end > 0 {
		return errors.New("invalid parameters: missing timestamp")
	} else {
		var is *gomif.InstanceStatus
		is, err = client.FetchLastInstanceStatus(
			context.Background(),
			opt.server,
			3600, // span (sec)
		)
		obj = is
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
		return err
	}
	return p.PrintObj(obj, nil, "")
}
