// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var notificationsOpts struct {
	list, clear, dismiss bool
	notifID              int64
	types                string
}

// notificationsCmd represents the notifications subcommand
var notificationsCmd = &cobra.Command{
	Use:     "notifications", // XXX
	Aliases: []string{"notification", "notif"},
	Short:   "Manage notifications",
	Example: `  madonctl accounts notifications --list
  madonctl accounts notifications --list --clear
  madonctl accounts notifications --dismiss --notification-id N
  madonctl accounts notifications --notification-id N
  madonctl accounts notifications --list --notification-types mentions
  madonctl accounts notifications --list --notification-types favourites
  madonctl accounts notifications --list --notification-types follows,reblogs`,
	//Long:    `TBW...`,
	RunE: notificationRunE,
}

func init() {
	accountsCmd.AddCommand(notificationsCmd)

	notificationsCmd.Flags().BoolVar(&notificationsOpts.list, "list", false, "List all current notifications")
	notificationsCmd.Flags().BoolVar(&notificationsOpts.clear, "clear", false, "Clear all current notifications")
	notificationsCmd.Flags().BoolVar(&notificationsOpts.dismiss, "dismiss", false, "Delete a notification")
	notificationsCmd.Flags().Int64Var(&notificationsOpts.notifID, "notification-id", 0, "Get a notification")
	notificationsCmd.Flags().StringVar(&notificationsOpts.types, "notification-types", "", "Filter notifications (mentions, favourites, reblogs, follows)")
}

func notificationRunE(cmd *cobra.Command, args []string) error {
	opt := notificationsOpts

	if !opt.list && !opt.clear && opt.notifID < 1 {
		return errors.New("missing parameters")
	}

	if err := madonInit(true); err != nil {
		return err
	}

	var limOpts *madon.LimitParams
	if accountsOpts.all || accountsOpts.limit > 0 || accountsOpts.sinceID > 0 || accountsOpts.maxID > 0 {
		limOpts = new(madon.LimitParams)
		limOpts.All = accountsOpts.all
	}

	if accountsOpts.limit > 0 {
		limOpts.Limit = int(accountsOpts.limit)
	}
	if accountsOpts.maxID > 0 {
		limOpts.MaxID = int64(accountsOpts.maxID)
	}
	if accountsOpts.sinceID > 0 {
		limOpts.SinceID = int64(accountsOpts.sinceID)
	}

	var filterMap *map[string]bool
	if opt.types != "" {
		var err error
		filterMap, err = buildFilterMap(opt.types)
		if err != nil {
			return nil
		}
	}

	var obj interface{}
	var err error

	if opt.list {
		var notifications []madon.Notification
		notifications, err = gClient.GetNotifications(limOpts)

		// Filter notifications
		if filterMap != nil && len(*filterMap) > 0 {
			if verbose {
				errPrint("Filtering notifications")
			}
			var newNotifications []madon.Notification
			for _, notif := range notifications {
				if (*filterMap)[notif.Type] {
					newNotifications = append(newNotifications, notif)
				}
			}
			notifications = newNotifications
		}

		if accountsOpts.keep > 0 && len(notifications) > int(accountsOpts.keep) {
			notifications = notifications[:accountsOpts.keep]
		}
		obj = notifications
	} else if opt.notifID > 0 {
		if opt.dismiss {
			err = gClient.DismissNotification(opt.notifID)
		} else {
			obj, err = gClient.GetNotification(opt.notifID)
		}
	}

	if err == nil && opt.clear {
		err = gClient.ClearNotifications()
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
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	return p.printObj(obj)
}

func buildFilterMap(types string) (*map[string]bool, error) {
	filterMap := make(map[string]bool)
	if types != "" {
		for _, f := range strings.Split(types, ",") {
			switch f {
			case "mention", "mentions":
				filterMap["mention"] = true
			case "favourite", "favourites", "favorite", "favorites", "fave", "faves":
				filterMap["favourite"] = true
			case "reblog", "reblogs", "retoot", "retoots":
				filterMap["reblog"] = true
			case "follow", "follows":
				filterMap["follow"] = true
			default:
				return nil, errors.Errorf("unknown notification type: '%s'", f)
			}
		}
	}
	return &filterMap, nil
}
