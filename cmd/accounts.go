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
	flag "github.com/spf13/pflag"

	"github.com/McKael/madon"
)

var accountsOpts struct {
	accountID                 int64
	accountUID                string
	unset                     bool   // TODO remove eventually?
	limit                     uint   // Limit the results
	sinceID, maxID            int64  // Query boundaries
	all                       bool   // Try to fetch all results
	onlyMedia, excludeReplies bool   // For acccount statuses
	remoteUID                 string // For account follow
	acceptFR, rejectFR        bool   // For account follow_requests
	list                      bool   // For account follow_requests/reports
	accountIDs                string // For account relationships
	statusIDs                 string // For account reports
	comment                   string // For account reports
	show                      bool   // For account reports
	displayName, note         string // For account update
	avatar, header            string // For account update
}

var updateFlags *flag.FlagSet

func init() {
	RootCmd.AddCommand(accountsCmd)

	// Subcommands
	accountsCmd.AddCommand(accountSubcommands...)

	// Global flags
	accountsCmd.PersistentFlags().Int64VarP(&accountsOpts.accountID, "account-id", "a", 0, "Account ID number")
	accountsCmd.PersistentFlags().StringVarP(&accountsOpts.accountUID, "user-id", "u", "", "Account user ID")
	accountsCmd.PersistentFlags().UintVarP(&accountsOpts.limit, "limit", "l", 0, "Limit number of results")
	accountsCmd.PersistentFlags().Int64Var(&accountsOpts.sinceID, "since-id", 0, "Request IDs greater than a value")
	accountsCmd.PersistentFlags().Int64Var(&accountsOpts.maxID, "max-id", 0, "Request IDs less (or equal) than a value")
	accountsCmd.PersistentFlags().BoolVar(&accountsOpts.all, "all", false, "Fetch all results")

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

	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.displayName, "display-name", "", "User display name")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.note, "note", "", "User note (a.k.a. bio)")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.avatar, "avatar", "", "User avatar image")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.header, "header", "", "User header image")

	// This one will be used to check if the options were explicitely set or not
	updateFlags = accountUpdateSubcommand.Flags()
}

// accountsCmd represents the accounts command
// This command does nothing without a subcommand
var accountsCmd = &cobra.Command{
	Use:     "accounts [--account-id ID] subcommand",
	Aliases: []string{"account"},
	Short:   "Account-related functions",
	//Long:    `TBW...`, // TODO
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
	accountUpdateSubcommand,
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
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}
var accountFollowSubcommand = &cobra.Command{
	Use:   "follow",
	Short: "Follow or unfollow the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountBlockSubcommand = &cobra.Command{
	Use:   "block",
	Short: "Block or unblock the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountMuteSubcommand = &cobra.Command{
	Use:   "mute",
	Short: "Mute or unmute the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountRelationshipsSubcommand = &cobra.Command{
	Use:   "relationships --account-ids ACC1,ACC2...",
	Short: "List relationships with the accounts",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountReportsSubcommand = &cobra.Command{
	Use:   "reports",
	Short: "List reports or report a user account",
	Example: `  madonctl accounts reports --list
  madonctl accounts reports --account-id ACCOUNT --status-ids ID... --comment TEXT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountUpdateSubcommand = &cobra.Command{
	Use:   "update",
	Short: "Update connected user account",
	Long: `Update connected user account

All flags are optional (set to an empty string if you want to delete a field).
The flags --avatar and --header can be paths to image files or base64-encoded
images (see Mastodon API specifications for the details).

Please note the avatar and header images cannot be removed, they can only be
replaced.`,
	Example: `  madonctl accounts update --display-name "Mr President"
  madonctl accounts update --note "I like madonctl"
  madonctl accounts update --avatar happyface.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

// accountSubcommandsRunE is a generic function for status subcommands
func accountSubcommandsRunE(subcmd string, args []string) error {
	opt := accountsOpts

	if opt.accountUID != "" {
		if opt.accountID > 0 {
			return errors.New("cannot use both account ID and UID")
		}
		// Sign in early to look the user id up
		var err error
		if err = madonInit(true); err != nil {
			return err
		}
		opt.accountID, err = accountLookupUser(opt.accountUID)
		if err != nil || opt.accountID < 1 {
			if err != nil {
				errPrint("Cannot find user '%s': %v", opt.accountUID, err)
			} else {
				errPrint("Cannot find user '%s'", opt.accountUID)
			}
			os.Exit(1)
		}
	}

	switch subcmd {
	case "show", "search", "update":
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
	case "followers", "following", "statuses":
		// If the user's account ID is missing, get it
		if opt.accountID < 1 {
			// Sign in now to look the user id up
			if err := madonInit(true); err != nil {
				return err
			}
			account, err := gClient.GetCurrentAccount()
			if err != nil {
				return err
			}
			opt.accountID = account.ID
			if verbose {
				errPrint("User account ID: %d", opt.accountID)
			}
		}
	default:
		// The other subcommands here require an account ID
		if opt.accountID < 1 {
			return errors.New("missing account ID")
		}
	}

	var limOpts *madon.LimitParams
	if opt.all || opt.limit > 0 || opt.sinceID > 0 || opt.maxID > 0 {
		limOpts = new(madon.LimitParams)
		limOpts.All = opt.all
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
		accountList, err = gClient.SearchAccounts(strings.Join(args, " "), limOpts)
		obj = accountList
	case "followers":
		var accountList []madon.Account
		accountList, err = gClient.GetAccountFollowers(opt.accountID, limOpts)
		if opt.limit > 0 && len(accountList) > int(opt.limit) {
			accountList = accountList[:opt.limit]
		}
		obj = accountList
	case "following":
		var accountList []madon.Account
		accountList, err = gClient.GetAccountFollowing(opt.accountID, limOpts)
		if opt.limit > 0 && len(accountList) > int(opt.limit) {
			accountList = accountList[:opt.limit]
		}
		obj = accountList
	case "statuses":
		var statusList []madon.Status
		statusList, err = gClient.GetAccountStatuses(opt.accountID, opt.onlyMedia, opt.excludeReplies, limOpts)
		if opt.limit > 0 && len(statusList) > int(opt.limit) {
			statusList = statusList[:opt.limit]
		}
		obj = statusList
	case "follow":
		var relationship *madon.Relationship
		if opt.unset {
			relationship, err = gClient.UnfollowAccount(opt.accountID)
			obj = relationship
		} else {
			if opt.accountID > 0 {
				relationship, err = gClient.FollowAccount(opt.accountID)
				obj = relationship
			} else {
				var account *madon.Account
				account, err = gClient.FollowRemoteAccount(opt.remoteUID)
				obj = account
			}
		}
	case "follow-requests":
		if opt.list {
			var followRequests []madon.Account
			followRequests, err = gClient.GetAccountFollowRequests(limOpts)
			if opt.limit > 0 && len(followRequests) > int(opt.limit) {
				followRequests = followRequests[:opt.limit]
			}
			obj = followRequests
		} else {
			err = gClient.FollowRequestAuthorize(opt.accountID, !opt.rejectFR)
		}
	case "block":
		var relationship *madon.Relationship
		if opt.unset {
			relationship, err = gClient.UnblockAccount(opt.accountID)
		} else {
			relationship, err = gClient.BlockAccount(opt.accountID)
		}
		obj = relationship
	case "mute":
		var relationship *madon.Relationship
		if opt.unset {
			relationship, err = gClient.UnmuteAccount(opt.accountID)
		} else {
			relationship, err = gClient.MuteAccount(opt.accountID)
		}
		obj = relationship
	case "favourites":
		var statusList []madon.Status
		statusList, err = gClient.GetFavourites(limOpts)
		if opt.limit > 0 && len(statusList) > int(opt.limit) {
			statusList = statusList[:opt.limit]
		}
		obj = statusList
	case "blocks":
		var accountList []madon.Account
		accountList, err = gClient.GetBlockedAccounts(limOpts)
		if opt.limit > 0 && len(accountList) > int(opt.limit) {
			accountList = accountList[:opt.limit]
		}
		obj = accountList
	case "mutes":
		var accountList []madon.Account
		accountList, err = gClient.GetMutedAccounts(limOpts)
		if opt.limit > 0 && len(accountList) > int(opt.limit) {
			accountList = accountList[:opt.limit]
		}
		obj = accountList
	case "relationships":
		var ids []int64
		ids, err = splitIDs(opt.accountIDs)
		if err != nil {
			return errors.New("cannot parse account IDs")
		}
		if opt.accountID > 0 { // Allow --account-id
			ids = []int64{opt.accountID}
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
			reports, err = gClient.GetReports(limOpts)
			if opt.limit > 0 && len(reports) > int(opt.limit) {
				reports = reports[:opt.limit]
			}
			obj = reports
			break
		}
		// Send a report
		var ids []int64
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
	case "update":
		var dn, note, avatar, header *string
		change := false
		if updateFlags.Lookup("display-name").Changed {
			dn = &opt.displayName
			change = true
		}
		if updateFlags.Lookup("note").Changed {
			note = &opt.note
			change = true
		}
		if updateFlags.Lookup("avatar").Changed {
			avatar = &opt.avatar
			change = true
		}
		if updateFlags.Lookup("header").Changed {
			header = &opt.header
			change = true
		}

		if !change { // We want at least one update
			return errors.New("missing parameters")
		}

		var account *madon.Account
		account, err = gClient.UpdateAccount(dn, note, avatar, header)
		obj = account
	default:
		return errors.New("accountSubcommand: internal error")
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

// accountLookupUser tries to find a (single) user matching 'user'
// If the user is an HTTP URL, it will use the search API, else
// it will use the accounts/search API.
func accountLookupUser(user string) (int64, error) {
	var accID int64

	if strings.HasPrefix(user, "https://") || strings.HasPrefix(user, "http://") {
		res, err := gClient.Search(user, true)
		if err != nil {
			return 0, err
		}
		if res != nil {
			if len(res.Accounts) > 1 {
				return 0, errors.New("several results")
			}
			if len(res.Accounts) == 1 {
				accID = res.Accounts[0].ID
			}
		}
	} else {
		// Remove leading '@'
		user = strings.TrimLeft(user, "@")

		accList, err := gClient.SearchAccounts(user, &madon.LimitParams{Limit: 2})
		if err != nil {
			return 0, err
		}
		for _, u := range accList {
			if u.Acct == user {
				accID = u.ID
				break
			}
		}
	}

	if accID < 1 {
		return 0, errors.New("user not found")
	}
	if verbose {
		errPrint("User '%s' is account ID %d", user, user)
	}
	return accID, nil
}
