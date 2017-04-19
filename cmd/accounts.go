// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

var accountsOpts struct {
	accountID                 int
	unset                     bool // TODO remove eventually?
	limit                     int
	onlyMedia, excludeReplies bool   // For acccount statuses
	remoteUID                 string // For account follow
	acceptFR, rejectFR        bool   // For account follow_requests
	list                      bool   // For account follow_requests/reports
	accountIDs                string // For account relationships
	statusIDs                 string // For account reports
	comment                   string // For account reports
	show                      bool   // For account reports
}

func init() {
	RootCmd.AddCommand(accountsCmd)

	// Subcommands
	accountsCmd.AddCommand(accountSubcommands...)

	// Global flags
	accountsCmd.PersistentFlags().IntVarP(&accountsOpts.accountID, "account-id", "a", 0, "Account ID number")
	//accountsCmd.PersistentFlags().IntVarP(&accountsOpts.limit, "limit", "l", 0, "Limit number of results")

	// Subcommand flags
	accountStatusesSubcommand.Flags().BoolVar(&accountsOpts.onlyMedia, "only-media", false, "Only statuses with media attachments")
	accountStatusesSubcommand.Flags().BoolVar(&accountsOpts.excludeReplies, "exclude-replies", false, "Exclude replies to other statuses")

	accountFollowRequestsSubcommand.Flags().BoolVar(&accountsOpts.list, "list", false, "List pending follow requests")
	accountFollowRequestsSubcommand.Flags().BoolVar(&accountsOpts.acceptFR, "accept", false, "Accept the follow request from the account ID")
	accountFollowRequestsSubcommand.Flags().BoolVar(&accountsOpts.rejectFR, "reject", false, "Reject the follow request from the account ID")

	accountBlockSubcommand.Flags().BoolVarP(&accountsOpts.unset, "unset", "", false, "Unblock the account")
	accountMuteSubcommand.Flags().BoolVarP(&accountsOpts.unset, "unset", "", false, "Unmute the account")
	accountFollowSubcommand.Flags().BoolVarP(&accountsOpts.unset, "unset", "", false, "Unfollow the account")
	accountFollowSubcommand.Flags().StringVarP(&accountsOpts.remoteUID, "remote", "r", "", "Follow remote account (user@domain)")

	accountRelationshipsSubcommand.Flags().StringVar(&accountsOpts.accountIDs, "account-ids", "", "Comma-separated list of account IDs")

	accountReportsSubcommand.Flags().StringVar(&accountsOpts.statusIDs, "status-ids", "", "Comma-separated list of status IDs")
	accountReportsSubcommand.Flags().StringVar(&accountsOpts.comment, "comment", "", "Report comment")
	accountReportsSubcommand.Flags().BoolVar(&accountsOpts.list, "list", false, "List current user reports")
}

// accountsCmd represents the accounts command
// This command does nothing without a subcommand
var accountsCmd = &cobra.Command{
	Use:     "accounts [--account-id ID] subcommand",
	Aliases: []string{"account"},
	Short:   "Account-related functions",
	//Long:    `TBW...`, // TODO
	//PersistentPreRunE: func(cmd *cobra.Command, args []string) error {},
}

// Note: Some account subcommands are not defined in this file.
var accountSubcommands = []*cobra.Command{
	&cobra.Command{
		Use: "show",
		Long: `Displays the details about the requested account.
If no account ID is specified, the current user account is used.`,
		Aliases: []string{"display"},
		Short:   "Display the account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "followers",
		Short: "Display the accounts following the specified account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "following",
		Short: "Display the accounts followed by the specified account",
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "favourites",
		Aliases: []string{"favorites", "favourited", "favorited"},
		Short:   "Display the user's favourites",
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "blocks",
		Aliases: []string{"blocked"},
		Short:   "Display the user's blocked accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:     "mutes",
		Aliases: []string{"muted"},
		Short:   "Display the user's muted accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	&cobra.Command{
		Use:   "search TEXT",
		Short: "Search for user accounts",
		Long: `Search for user accounts.
The server will lookup an account remotely if the search term is in the
username@domain format and not yet in the database.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return accountSubcommandsRunE(cmd.Name(), args)
		},
	},
	accountStatusesSubcommand,
	accountFollowRequestsSubcommand,
	accountFollowSubcommand,
	accountBlockSubcommand,
	accountMuteSubcommand,
	accountRelationshipsSubcommand,
	accountReportsSubcommand,
}

var accountStatusesSubcommand = &cobra.Command{
	Use:     "statuses",
	Aliases: []string{"st"},
	Short:   "Display the account statuses",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountFollowRequestsSubcommand = &cobra.Command{
	Use:     "follow-requests",
	Aliases: []string{"follow-request", "fr"},
	Example: `  madonctl accounts follow-requests --list
  madonctl accounts follow-requests --account-id X --accept
  madonctl accounts follow-requests --account-id Y --reject`,
	Short: "List, accept or deny a follow request",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE("follow-requests", args)
	},
}
var accountFollowSubcommand = &cobra.Command{
	Use:   "follow",
	Short: "Follow or unfollow the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE("follow", args)
	},
}

var accountBlockSubcommand = &cobra.Command{
	Use:   "block",
	Short: "Block or unblock the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE("block", args)
	},
}

var accountMuteSubcommand = &cobra.Command{
	Use:   "mute",
	Short: "Mute or unmute the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE("mute", args)
	},
}

var accountRelationshipsSubcommand = &cobra.Command{
	Use:   "relationships --account-ids ACC1,ACC2...",
	Short: "List relationships with the accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE("relationships", args)
	},
}

var accountReportsSubcommand = &cobra.Command{
	Use:   "reports",
	Short: "List reports or report a user account",
	Example: `  madonctl accounts reports --list
  madonctl accounts reports --account-id ACCOUNT --status-ids ID... --comment TEXT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE("reports", args)
	},
}

// accountSubcommandsRunE is a generic function for status subcommands
func accountSubcommandsRunE(subcmd string, args []string) error {
	opt := accountsOpts

	switch subcmd {
	case "show", "search":
		// These subcommands do not require an account ID
	case "favourites", "blocks", "mutes":
		// Those subcommands can not use an account ID
		if opt.accountID > 0 {
			return errors.New("useless account ID")
		}
	case "follow":
		if opt.accountID < 1 && opt.remoteUID == "" {
			return errors.New("missing account ID or URI")
		}
		if opt.accountID > 0 && opt.remoteUID != "" {
			return errors.New("cannot use both account ID and URI")
		}
		if opt.unset && opt.accountID < 1 {
			return errors.New("unfollowing requires an account ID")
		}
	case "follow-requests":
		if opt.list {
			if opt.acceptFR || opt.rejectFR {
				return errors.New("incompatible options")
			}
		} else {
			if !opt.acceptFR && !opt.rejectFR { // No flag
				return errors.New("missing parameter (--list, --accept or --reject)")
			}
			// This is a FR reply
			if opt.acceptFR && opt.rejectFR {
				return errors.New("incompatible options")
			}
			if opt.accountID < 1 {
				return errors.New("missing account ID")
			}
		}
	case "relationships":
		if opt.accountID < 1 && len(opt.accountIDs) == 0 {
			return errors.New("missing account IDs")
		}
		if opt.accountID > 0 && len(opt.accountIDs) > 0 {
			return errors.New("incompatible options")
		}
	case "reports":
		if opt.list {
			break // No argument needed
		}
		if opt.accountID < 1 || len(opt.statusIDs) == 0 || opt.comment == "" {
			return errors.New("missing parameter")
		}
	default:
		// The other subcommands here require an account ID
		if opt.accountID < 1 {
			return errors.New("missing account ID")
		}
	}

	// All account subcommands need to have signed in
	if err := madonInit(true); err != nil {
		return err
	}

	var obj interface{}
	var err error

	switch subcmd {
	case "show":
		var account *madon.Account
		if opt.accountID > 0 {
			account, err = gClient.GetAccount(opt.accountID)
		} else {
			account, err = gClient.GetCurrentAccount()
		}
		obj = account
	case "search":
		var accountList []madon.Account
		limit := 0 // TODO use a global flag
		accountList, err = gClient.SearchAccounts(strings.Join(args, " "), limit)
		obj = accountList
	case "followers":
		var accountList []madon.Account
		accountList, err = gClient.GetAccountFollowers(opt.accountID)
		obj = accountList
	case "following":
		var accountList []madon.Account
		accountList, err = gClient.GetAccountFollowing(opt.accountID)
		obj = accountList
	case "statuses":
		var statusList []madon.Status
		statusList, err = gClient.GetAccountStatuses(opt.accountID, opt.onlyMedia, opt.excludeReplies)
		obj = statusList
	case "follow":
		if opt.unset {
			err = gClient.UnfollowAccount(opt.accountID)
		} else {
			if opt.accountID > 0 {
				err = gClient.FollowAccount(opt.accountID)
			} else {
				var account *madon.Account
				account, err = gClient.FollowRemoteAccount(opt.remoteUID)
				obj = account
			}
		}
	case "follow-requests":
		if opt.list {
			var followRequests []madon.Account
			followRequests, err = gClient.GetAccountFollowRequests()
			obj = followRequests
		} else {
			err = gClient.FollowRequestAuthorize(opt.accountID, !opt.rejectFR)
		}
	case "block":
		if opt.unset {
			err = gClient.UnblockAccount(opt.accountID)
		} else {
			err = gClient.BlockAccount(opt.accountID)
		}
	case "mute":
		if opt.unset {
			err = gClient.UnmuteAccount(opt.accountID)
		} else {
			err = gClient.MuteAccount(opt.accountID)
		}
	case "favourites":
		var statusList []madon.Status
		statusList, err = gClient.GetFavourites()
		obj = statusList
	case "blocks":
		var accountList []madon.Account
		accountList, err = gClient.GetBlockedAccounts()
		obj = accountList
	case "mutes":
		var accountList []madon.Account
		accountList, err = gClient.GetMutedAccounts()
		obj = accountList
	case "relationships":
		var ids []int
		ids, err = splitIDs(opt.accountIDs)
		if err != nil {
			return errors.New("cannot parse account IDs")
		}
		if opt.accountID > 0 { // Allow --account-id
			ids = []int{opt.accountID}
		}
		if len(ids) < 1 {
			return errors.New("missing account IDs")
		}
		var relationships []madon.Relationship
		relationships, err = gClient.GetAccountRelationships(ids)
		obj = relationships
	case "reports":
		if opt.list {
			var reports []madon.Report
			reports, err = gClient.GetReports()
			obj = reports
			break
		}
		// Send a report
		var ids []int
		ids, err = splitIDs(opt.statusIDs)
		if err != nil {
			return errors.New("cannot parse status IDs")
		}
		if len(ids) < 1 {
			return errors.New("missing status IDs")
		}
		var report *madon.Report
		report, err = gClient.ReportUser(opt.accountID, ids, opt.comment)
		obj = report
	default:
		return errors.New("accountSubcommand: internal error")
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
