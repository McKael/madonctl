// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon/v3"
)

var notificationsOpts struct {
	list, clear, dismiss bool
	notifID              madon.ActivityID
	types                string
	excludeTypes         string
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
  madonctl accounts notifications --list --exclude-types mention,reblog
  madonctl accounts notifications --list --notification-types mentions
  madonctl accounts notifications --list --notification-types favourites
  madonctl accounts notifications --list --notification-types follows,reblogs`,
	Long: `Manage notifications

This commands let you list, display and dismiss notifications.

Please note that --notifications-types filters the notifications locally,
while --exclude-types is supported by the API and should be more efficient.`,
	RunE: notificationRunE,
}

func init() {
	accountsCmd.AddCommand(notificationsCmd)

	notificationsCmd.Flags().BoolVar(&notificationsOpts.list, "list", false, "List all current notifications")
	notificationsCmd.Flags().BoolVar(&notificationsOpts.clear, "clear", false, "Clear all current notifications")
	notificationsCmd.Flags().BoolVar(&notificationsOpts.dismiss, "dismiss", false, "Delete a notification")
	notificationsCmd.Flags().StringVar(&notificationsOpts.notifID, "notification-id", "", "Get a notification")
	notificationsCmd.Flags().StringVar(&notificationsOpts.types, "notification-types", "", "Filter notifications (mention, favourite, reblog, follow)")
	notificationsCmd.Flags().StringVar(&notificationsOpts.excludeTypes, "exclude-types", "", "Exclude notifications types (mention, favourite, reblog, follow)")
}

func notificationRunE(cmd *cobra.Command, args []string) error {
	opt := notificationsOpts

	if !opt.list && !opt.clear && opt.notifID == "" {
		return errors.New("missing parameters")
	}

	if err := madonInit(true); err != nil {
		return err
	}

	var limOpts *madon.LimitParams
	if accountsOpts.all || accountsOpts.limit > 0 || accountsOpts.sinceID != "" || accountsOpts.maxID != "" {
		limOpts = new(madon.LimitParams)
		limOpts.All = accountsOpts.all
	}

	if accountsOpts.limit > 0 {
		limOpts.Limit = int(accountsOpts.limit)
	}
	if accountsOpts.maxID != "" {
		limOpts.MaxID = accountsOpts.maxID
	}
	if accountsOpts.sinceID != "" {
		limOpts.SinceID = accountsOpts.sinceID
	}

	var filterMap *map[string]bool
	if opt.types != "" {
		var err error
		filterMap, err = buildFilterMap(opt.types)
		if err != nil {
			return errors.Wrap(err, "bad notification filter")
		}
	}

	var xTypes []string
	if xt, err := splitNotificationTypes(opt.excludeTypes); err == nil {
		xTypes = xt
	} else {
		return errors.Wrap(err, "invalid exclude-types argument")
	}

	var obj interface{}
	var err error

	if opt.list {
		var notifications []madon.Notification
		notifications, err = gClient.GetNotifications(xTypes, limOpts)

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
	} else if opt.notifID != "" {
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

func splitNotificationTypes(types string) ([]string, error) {
	var typeList []string
	if types == "" {
		return typeList, nil
	}
	for _, f := range strings.Split(types, ",") {
		switch f {
		case "mention", "mentions":
			f = "mention"
		case "favourite", "favourites", "favorite", "favorites", "fave", "faves":
			f = "favourite"
		case "reblog", "reblogs", "retoot", "retoots":
			f = "reblog"
		case "follow", "follows":
			f = "follow"
		default:
			return nil, errors.Errorf("unknown notification type: '%s'", f)
		}
		typeList = append(typeList, f)
	}
	return typeList, nil
}

func buildFilterMap(types string) (*map[string]bool, error) {
	filterMap := make(map[string]bool)
	if types == "" {
		return &filterMap, nil
	}
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
	return &filterMap, nil
}
