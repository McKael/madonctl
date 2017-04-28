// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"errors"

	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var notificationsOpts struct {
	list, clear, dismiss bool
	notifID              int
}

// notificationsCmd represents the notifications subcommand
var notificationsCmd = &cobra.Command{
	Use:     "notifications", // XXX
	Aliases: []string{"notification", "notif"},
	Short:   "Manage notifications",
	Example: `  madonctl accounts notifications --list
  madonctl accounts notifications --list --clear
  madonctl accounts notifications --dismiss --notification-id N
  madonctl accounts notifications --notification-id N`,
	//Long:    `TBW...`,
	RunE: notificationRunE,
}

func init() {
	accountsCmd.AddCommand(notificationsCmd)

	notificationsCmd.Flags().BoolVar(&notificationsOpts.list, "list", false, "List all current notifications")
	notificationsCmd.Flags().BoolVar(&notificationsOpts.clear, "clear", false, "Clear all current notifications")
	notificationsCmd.Flags().BoolVar(&notificationsOpts.dismiss, "dismiss", false, "Delete a notification")
	notificationsCmd.Flags().IntVar(&notificationsOpts.notifID, "notification-id", 0, "Get a notification")
}

func notificationRunE(cmd *cobra.Command, args []string) error {
	opt := notificationsOpts

	if !opt.list && !opt.clear && opt.notifID < 1 {
		return errors.New("missing parameters")
	}

	if err := madonInit(true); err != nil {
		return err
	}

	var obj interface{}
	var err error

	if opt.list {
		var notifications []madon.Notification
		notifications, err = gClient.GetNotifications()
		if accountsOpts.limit > 0 {
			notifications = notifications[:accountsOpts.limit]
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
		return nil
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
