// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/McKael/madon/v3"
)

var accountUpdateFlags, accountMuteFlags, accountFollowFlags *flag.FlagSet

var accountsOpts struct {
	accountID             int64
	accountUID            string
	unset                 bool     // TODO remove eventually?
	limit, keep           uint     // Limit the results
	sinceID, maxID        int64    // Query boundaries
	all                   bool     // Try to fetch all results
	onlyMedia, onlyPinned bool     // For acccount statuses
	excludeReplies        bool     // For acccount statuses
	remoteUID             string   // For account follow
	reblogs               bool     // For account follow
	acceptFR, rejectFR    bool     // For account follow_requests
	list                  bool     // For account follow_requests/reports
	accountIDs            string   // For account relationships
	statusIDs             string   // For account reports
	comment               string   // For account reports
	displayName, note     string   // For account update
	profileFields         []string // For account update
	avatar, header        string   // For account update
	defaultLanguage       string   // For account update
	defaultPrivacy        string   // For account update
	defaultSensitive      bool     // For account update
	locked, bot           bool     // For account update
	muteNotifications     bool     // For account mute
	following             bool     // For account search
}

func init() {
	RootCmd.AddCommand(accountsCmd)

	// Subcommands
	accountsCmd.AddCommand(accountSubcommands...)

	// Global flags
	accountsCmd.PersistentFlags().Int64VarP(&accountsOpts.accountID, "account-id", "a", 0, "Account ID number")
	accountsCmd.PersistentFlags().StringVarP(&accountsOpts.accountUID, "user-id", "u", "", "Account user ID")
	accountsCmd.PersistentFlags().UintVarP(&accountsOpts.limit, "limit", "l", 0, "Limit number of API results")
	accountsCmd.PersistentFlags().UintVarP(&accountsOpts.keep, "keep", "k", 0, "Limit number of results")
	accountsCmd.PersistentFlags().Int64Var(&accountsOpts.sinceID, "since-id", 0, "Request IDs greater than a value")
	accountsCmd.PersistentFlags().Int64Var(&accountsOpts.maxID, "max-id", 0, "Request IDs less (or equal) than a value")
	accountsCmd.PersistentFlags().BoolVar(&accountsOpts.all, "all", false, "Fetch all results")

	// Subcommand flags
	accountStatusesSubcommand.Flags().BoolVar(&accountsOpts.onlyPinned, "pinned", false, "Only statuses that have been pinned")
	accountStatusesSubcommand.Flags().BoolVar(&accountsOpts.onlyMedia, "only-media", false, "Only statuses with media attachments")
	accountStatusesSubcommand.Flags().BoolVar(&accountsOpts.excludeReplies, "exclude-replies", false, "Exclude replies to other statuses")

	accountFollowRequestsSubcommand.Flags().BoolVar(&accountsOpts.list, "list", false, "List pending follow requests")
	accountFollowRequestsSubcommand.Flags().BoolVar(&accountsOpts.acceptFR, "accept", false, "Accept the follow request from the account ID")
	accountFollowRequestsSubcommand.Flags().BoolVar(&accountsOpts.rejectFR, "reject", false, "Reject the follow request from the account ID")

	accountBlockSubcommand.Flags().BoolVarP(&accountsOpts.unset, "unset", "", false, "Unblock the account (deprecated)")

	accountMuteSubcommand.Flags().BoolVarP(&accountsOpts.unset, "unset", "", false, "Unmute the account (deprecated)")
	accountMuteSubcommand.Flags().BoolVarP(&accountsOpts.muteNotifications, "notifications", "", true, "Mute the notifications")
	accountFollowSubcommand.Flags().BoolVarP(&accountsOpts.unset, "unset", "", false, "Unfollow the account (deprecated)")
	accountFollowSubcommand.Flags().BoolVarP(&accountsOpts.reblogs, "show-reblogs", "", true, "Follow account's boosts")
	accountFollowSubcommand.Flags().StringVarP(&accountsOpts.remoteUID, "remote", "r", "", "Follow remote account (user@domain)")

	accountRelationshipsSubcommand.Flags().StringVar(&accountsOpts.accountIDs, "account-ids", "", "Comma-separated list of account IDs")

	accountReportsSubcommand.Flags().StringVar(&accountsOpts.statusIDs, "status-ids", "", "Comma-separated list of status IDs")
	accountReportsSubcommand.Flags().StringVar(&accountsOpts.comment, "comment", "", "Report comment")
	accountReportsSubcommand.Flags().BoolVar(&accountsOpts.list, "list", false, "List current user reports")

	accountSearchSubcommand.Flags().BoolVar(&accountsOpts.following, "following", false, "Restrict search to accounts you are following")

	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.displayName, "display-name", "", "User display name")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.note, "note", "", "User note (a.k.a. bio)")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.avatar, "avatar", "", "User avatar image")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.header, "header", "", "User header image")
	accountUpdateSubcommand.Flags().StringArrayVar(&accountsOpts.profileFields, "profile-field", nil, "Profile metadata field (NAME=VALUE)")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.defaultLanguage, "default-language", "", "Default toots language (iso 639 code)")
	accountUpdateSubcommand.Flags().StringVar(&accountsOpts.defaultPrivacy, "default-privacy", "", "Default toot privacy (public, unlisted, private)")
	accountUpdateSubcommand.Flags().BoolVar(&accountsOpts.defaultSensitive, "default-sensitive", false, "Mark medias as sensitive by default")
	accountUpdateSubcommand.Flags().BoolVar(&accountsOpts.locked, "locked", false, "Following account requires approval")
	accountUpdateSubcommand.Flags().BoolVar(&accountsOpts.bot, "bot", false, "Set as service (automated) account")

	// Deprecated flags
	accountBlockSubcommand.Flags().MarkDeprecated("unset", "please use unblock instead")
	accountMuteSubcommand.Flags().MarkDeprecated("unset", "please use unmute instead")
	accountFollowSubcommand.Flags().MarkDeprecated("unset", "please use unfollow instead")

	// Those variables will be used to check if the options were
	// explicitly set or not
	accountUpdateFlags = accountUpdateSubcommand.Flags()
	accountMuteFlags = accountMuteSubcommand.Flags()
	accountFollowFlags = accountFollowSubcommand.Flags()
}

// accountsCmd represents the accounts command
// This command does nothing without a subcommand
var accountsCmd = &cobra.Command{
	Use:     "account [--account-id ID] subcommand",
	Aliases: []string{"accounts"},
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
		Example: `  madonctl account show   # Display your own account

  madonctl account show --account-id 1234
  madonctl account show --user-id Gargron@mastodon.social
  madonctl account show --user-id https://mastodon.social/@Gargron

  madonctl account show 1234
  madonctl account show Gargron@mastodon.social
  madonctl account show https://mastodon.social/@Gargron
`,
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
	accountSearchSubcommand,
	accountStatusesSubcommand,
	accountFollowRequestsSubcommand,
	accountFollowSubcommand,
	accountUnfollowSubcommand,
	accountBlockSubcommand,
	accountUnblockSubcommand,
	accountMuteSubcommand,
	accountUnmuteSubcommand,
	accountPinSubcommand,
	accountUnpinSubcommand,
	accountRelationshipsSubcommand,
	accountReportsSubcommand,
	accountUpdateSubcommand,
	accountListEndorsementsSubcommand,
}

var accountSearchSubcommand = &cobra.Command{
	Use:   "search TEXT",
	Short: "Search for user accounts",
	Long: `Search for user accounts.

This command will lookup an account remotely if the search term is in the
@domain format and not yet known to the server.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountStatusesSubcommand = &cobra.Command{
	Use:     "statuses",
	Aliases: []string{"st"},
	Short:   "Display the account statuses",
	Example: `  madonctl account statuses
  madonctl account statuses 404                         # local account numeric ID
  madonctl account statuses @McKael                     # local account
  madonctl account statuses Gargron@mastodon.social     # remote (known account)
  madonctl account statuses https://mastodon.social/@Gargron  # any account URL
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountFollowRequestsSubcommand = &cobra.Command{
	Use:     "follow-requests",
	Aliases: []string{"follow-request", "fr"},
	Short:   "List, accept or deny a follow request",
	Example: `  madonctl account follow-requests --list
  madonctl account follow-requests --account-id X --accept
  madonctl account follow-requests --account-id Y --reject`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}
var accountFollowSubcommand = &cobra.Command{
	Use:   "follow",
	Short: "Follow an account",
	Example: `# Argument type can be set explicitly:
  madonctl account follow --account-id 1234
  madonctl account follow --remote Gargron@mastodon.social

# Or argument type can be guessed:
  madonctl account follow 4800
  madonctl account follow Gargron@mastodon.social --show-reblogs=false
  madonctl account follow https://mastodon.social/@Gargron
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountUnfollowSubcommand = &cobra.Command{
	Use:   "unfollow",
	Short: "Stop following an account",
	Example: `  madonctl account unfollow --account-id 1234

Same usage as madonctl follow.
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountBlockSubcommand = &cobra.Command{
	Use:   "block",
	Short: "Block the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountUnblockSubcommand = &cobra.Command{
	Use:   "unblock",
	Short: "Unblock the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountMuteSubcommand = &cobra.Command{
	Use:   "mute",
	Short: "Mute the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountUnmuteSubcommand = &cobra.Command{
	Use:   "unmute",
	Short: "Unmute the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountPinSubcommand = &cobra.Command{
	Use:     "pin",
	Short:   "Endorse (pin) the account",
	Aliases: []string{"endorse"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountUnpinSubcommand = &cobra.Command{
	Use:     "unpin",
	Short:   "Cancel endorsement of an account",
	Aliases: []string{"disavow"},
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountListEndorsementsSubcommand = &cobra.Command{
	Use:     "pinned",
	Short:   `Display the list of pinned (endorsed) accounts`,
	Aliases: []string{"list-endorsements", "get-endorsements"},
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
	Example: `  madonctl account reports --list
  madonctl account reports --account-id ACCOUNT --status-ids ID... --comment TEXT`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

var accountUpdateSubcommand = &cobra.Command{
	Use:   "update",
	Short: "Update connected user account",
	Long: `Update connected user account

All flags are optional (set to an empty string if you want to delete a field).
The options --avatar and --header should be paths to image files.

Please note the avatar and header images cannot be removed, they can only be
replaced.`,
	Example: `  madonctl account update --display-name "Mr President"
  madonctl account update --note "I like madonctl"
  madonctl account update --avatar happyface.png`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return accountSubcommandsRunE(cmd.Name(), args)
	},
}

// accountSubcommandsRunE is a generic function for status subcommands
func accountSubcommandsRunE(subcmd string, args []string) error {
	opt := accountsOpts

	if len(args) > 1 {
		return errors.New("too many arguments")
	}

	userInArg := false

	if len(args) == 1 {
		if len(args[0]) > 0 {
			userInArg = true
		} else {
			return errors.New("invalid argument (empty)")
		}
	}

	// Check account is provided in only one way
	aCounter := 0
	if opt.accountID > 0 {
		aCounter++
	}
	if opt.accountUID != "" {
		aCounter++
	}
	if opt.remoteUID != "" {
		aCounter++
	}
	if userInArg {
		aCounter++
	}

	if aCounter > 1 {
		return errors.New("too many account identifiers provided")
	}

	if userInArg {
		// Is the argument an account ID?
		if n, err := strconv.ParseInt(args[0], 10, 64); err == nil {
			opt.accountID = n
		} else if strings.HasPrefix(args[0], "https://") || strings.HasPrefix(args[0], "http://") {
			// That is not a remote UID scheme
			opt.accountUID = args[0]
		} else if subcmd == "follow" {
			// For the follow API, got to be a remote UID...
			opt.remoteUID = args[0]
			// ... unless it's local (i.e. no '@' in the identifier)...
			fid := strings.TrimLeft(args[0], "@")
			if !strings.ContainsRune(fid, '@') {
				opt.accountUID = args[0]
				opt.remoteUID = ""
			}
		} else {
			// Fall back to account UID
			opt.accountUID = args[0]
		}
	}

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
	case "favourites", "blocks", "mutes", "pinned":
		// Those subcommands can not use an account ID
		if opt.accountID > 0 {
			return errors.New("useless account ID")
		}
	case "follow", "unfollow":
		// We need an account ID or a remote UID
		if opt.accountID < 1 && opt.remoteUID == "" {
			return errors.New("missing account ID or URI")
		}
		if opt.accountID > 0 && opt.remoteUID != "" {
			return errors.New("cannot use both account ID and URI")
		}
		if (opt.unset || subcmd == "unfollow") && opt.accountID < 1 {
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
		accountList, err = gClient.SearchAccounts(strings.Join(args, " "), opt.following, limOpts)
		obj = accountList
	case "followers":
		var accountList []madon.Account
		accountList, err = gClient.GetAccountFollowers(opt.accountID, limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
		}
		obj = accountList
	case "following":
		var accountList []madon.Account
		accountList, err = gClient.GetAccountFollowing(opt.accountID, limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
		}
		obj = accountList
	case "statuses":
		var statusList []madon.Status
		statusList, err = gClient.GetAccountStatuses(opt.accountID, opt.onlyPinned, opt.onlyMedia, opt.excludeReplies, limOpts)
		if opt.keep > 0 && len(statusList) > int(opt.keep) {
			statusList = statusList[:opt.keep]
		}
		obj = statusList
	case "follow", "unfollow":
		var relationship *madon.Relationship
		if opt.unset || subcmd == "unfollow" {
			relationship, err = gClient.UnfollowAccount(opt.accountID)
			obj = relationship
			break
		}
		if opt.accountID <= 0 {
			if opt.remoteUID != "" {
				// Remote account
				var account *madon.Account
				account, err = gClient.FollowRemoteAccount(opt.remoteUID)
				obj = account
				break
			}
			return errors.New("error: no usable parameter")
		}

		// Locally-known account
		var followReblogs *bool
		if accountFollowFlags.Lookup("show-reblogs").Changed {
			// Set followReblogs as it's been explicitly requested
			followReblogs = &opt.reblogs
		}
		relationship, err = gClient.FollowAccount(opt.accountID, followReblogs)
		obj = relationship
	case "follow-requests":
		if opt.list {
			var followRequests []madon.Account
			followRequests, err = gClient.GetAccountFollowRequests(limOpts)
			if opt.accountID > 0 { // Display a specific request
				var fRequest *madon.Account
				for _, fr := range followRequests {
					if fr.ID == opt.accountID {
						fRequest = &fr
						break
					}
				}
				if fRequest != nil {
					followRequests = []madon.Account{*fRequest}
				} else {
					followRequests = []madon.Account{}
				}
			} else {
				if opt.keep > 0 && len(followRequests) > int(opt.keep) {
					followRequests = followRequests[:opt.keep]
				}
			}
			obj = followRequests
		} else {
			err = gClient.FollowRequestAuthorize(opt.accountID, !opt.rejectFR)
		}
	case "block", "unblock":
		var relationship *madon.Relationship
		if opt.unset || subcmd == "unblock" {
			relationship, err = gClient.UnblockAccount(opt.accountID)
		} else {
			relationship, err = gClient.BlockAccount(opt.accountID)
		}
		obj = relationship
	case "mute", "unmute":
		var relationship *madon.Relationship
		if opt.unset || subcmd == "unmute" {
			relationship, err = gClient.UnmuteAccount(opt.accountID)
		} else {
			var muteNotif *bool
			if accountMuteFlags.Lookup("notifications").Changed {
				muteNotif = &opt.muteNotifications
			}
			relationship, err = gClient.MuteAccount(opt.accountID, muteNotif)
		}
		obj = relationship
	case "pin", "unpin":
		var relationship *madon.Relationship
		if subcmd == "unpin" {
			relationship, err = gClient.UnpinAccount(opt.accountID)
		} else {
			relationship, err = gClient.PinAccount(opt.accountID)
		}
		obj = relationship
	case "favourites":
		var statusList []madon.Status
		statusList, err = gClient.GetFavourites(limOpts)
		if opt.keep > 0 && len(statusList) > int(opt.keep) {
			statusList = statusList[:opt.keep]
		}
		obj = statusList
	case "blocks":
		var accountList []madon.Account
		accountList, err = gClient.GetBlockedAccounts(limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
		}
		obj = accountList
	case "mutes":
		var accountList []madon.Account
		accountList, err = gClient.GetMutedAccounts(limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
		}
		obj = accountList
	case "pinned":
		var accountList []madon.Account
		accountList, err = gClient.GetEndorsements(limOpts)
		if opt.keep > 0 && len(accountList) > int(opt.keep) {
			accountList = accountList[:opt.keep]
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
			if opt.keep > 0 && len(reports) > int(opt.keep) {
				reports = reports[:opt.keep]
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
		var updateParams madon.UpdateAccountParams
		var source *madon.SourceParams
		change := false

		if accountUpdateFlags.Lookup("display-name").Changed {
			updateParams.DisplayName = &opt.displayName
			change = true
		}
		if accountUpdateFlags.Lookup("note").Changed {
			updateParams.Note = &opt.note
			change = true
		}
		if accountUpdateFlags.Lookup("avatar").Changed {
			updateParams.AvatarImagePath = &opt.avatar
			change = true
		}
		if accountUpdateFlags.Lookup("header").Changed {
			updateParams.HeaderImagePath = &opt.header
			change = true
		}
		if accountUpdateFlags.Lookup("locked").Changed {
			updateParams.Locked = &opt.locked
			change = true
		}
		if accountUpdateFlags.Lookup("bot").Changed {
			updateParams.Bot = &opt.bot
			change = true
		}
		if accountUpdateFlags.Lookup("default-language").Changed {
			if source == nil {
				source = &madon.SourceParams{}
			}
			source.Language = &opt.defaultLanguage
			change = true
		}
		if accountUpdateFlags.Lookup("default-privacy").Changed {
			if source == nil {
				source = &madon.SourceParams{}
			}
			source.Privacy = &opt.defaultPrivacy
			change = true
		}
		if accountUpdateFlags.Lookup("default-sensitive").Changed {
			if source == nil {
				source = &madon.SourceParams{}
			}
			source.Sensitive = &opt.defaultSensitive
			change = true
		}
		if accountUpdateFlags.Lookup("profile-field").Changed {
			var fa = []madon.Field{}
			for _, f := range opt.profileFields {
				kv := strings.SplitN(f, "=", 2)
				if len(kv) != 2 {
					return errors.New("cannot parse field")
				}
				fa = append(fa, madon.Field{Name: kv[0], Value: kv[1]})
			}
			updateParams.FieldsAttributes = &fa
			change = true
		}

		if !change { // We want at least one update
			return errors.New("missing parameters")
		}

		updateParams.Source = source

		var account *madon.Account
		account, err = gClient.UpdateAccount(updateParams)
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

		accList, err := gClient.SearchAccounts(user, false, &madon.LimitParams{Limit: 2})
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
