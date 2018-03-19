// Copyright © 2017-2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"
	"github.com/spf13/viper"

	"github.com/McKael/madon"
)

// toot is a kind of alias for status post

var tootAliasFlags *flag.FlagSet

func init() {
	RootCmd.AddCommand(tootAliasCmd)

	tootAliasCmd.Flags().BoolVar(&statusOpts.sensitive, "sensitive", false, "Mark post as sensitive (NSFW)")
	tootAliasCmd.Flags().StringVar(&statusOpts.visibility, "visibility", "", "Visibility (direct|private|unlisted|public)")
	tootAliasCmd.Flags().StringVar(&statusOpts.spoiler, "spoiler", "", "Spoiler warning (CW)")
	tootAliasCmd.Flags().StringVar(&statusOpts.mediaIDs, "media-ids", "", "Comma-separated list of media IDs")
	tootAliasCmd.Flags().StringVarP(&statusOpts.mediaFilePath, "file", "f", "", "Media attachment file name")
	tootAliasCmd.Flags().StringVar(&statusOpts.textFilePath, "text-file", "", "Text file name (message content)")
	tootAliasCmd.Flags().Int64VarP(&statusOpts.inReplyToID, "in-reply-to", "r", 0, "Status ID to reply to")
	tootAliasCmd.Flags().BoolVar(&statusOpts.stdin, "stdin", false, "Read message content from standard input")
	tootAliasCmd.Flags().BoolVar(&statusOpts.addMentions, "add-mentions", false, "Add mentions when replying")
	tootAliasCmd.Flags().BoolVar(&statusOpts.sameVisibility, "same-visibility", false, "Use same visibility as original message (for replies)")

	// Flag completion
	annotation := make(map[string][]string)
	annotation[cobra.BashCompCustom] = []string{"__madonctl_visibility"}

	tootAliasCmd.Flags().Lookup("visibility").Annotations = annotation

	// This one will be used to check if the options were explicitly set or not
	tootAliasFlags = tootAliasCmd.Flags()
}

var tootAliasCmd = &cobra.Command{
	Use:     "toot",
	Aliases: []string{"post", "pouet"},
	Short:   "Post a message (toot)",
	Example: `  madonctl toot message
  madonctl toot --spoiler Warning "Hello, World"
  madonctl status post --media-ids ID1,ID2 "Here are the photos"
  madonctl post --sensitive --file image.jpg Image
  madonctl toot --text-file message.txt
  madonctl toot --in-reply-to STATUSID "@user response"
  madonctl toot --in-reply-to STATUSID --add-mentions "response"
  echo "Hello from #madonctl" | madonctl toot --visibility unlisted --stdin

The default visibility can be set in the configuration file with the option
'default_visibility' (or with an environmnent variable).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := madonInit(true); err != nil {
			return err
		}
		return statusSubcommandRunE("post", args)
	},
}

func toot(tootText string) (*madon.Status, error) {
	opt := statusOpts

	// Get default visibility from configuration
	if opt.visibility == "" {
		if v := viper.GetString("default_visibility"); v != "" {
			opt.visibility = v
		}
	}

	switch opt.visibility {
	case "", "direct", "private", "unlisted", "public":
		// OK
	default:
		return nil, errors.Errorf("invalid visibility argument value '%s'", opt.visibility)
	}

	if opt.inReplyToID < 0 {
		return nil, errors.New("invalid in-reply-to argument value")
	}

	ids, err := splitIDs(opt.mediaIDs)
	if err != nil {
		return nil, errors.New("cannot parse media IDs")
	}

	if tootText == "" && len(ids) == 0 && opt.spoiler == "" && opt.mediaFilePath == "" {
		return nil, errors.New("toot is empty")
	}

	if opt.inReplyToID > 0 {
		var initialStatus *madon.Status
		var preserveVis bool
		if opt.sameVisibility &&
			!tootAliasFlags.Lookup("visibility").Changed &&
			!statusPostFlags.Lookup("visibility").Changed {
			// Preserve visibility unless the --visibility flag
			// has been used in the command line.
			preserveVis = true
		}
		if preserveVis || opt.addMentions {
			// Fetch original status message
			initialStatus, err = gClient.GetStatus(opt.inReplyToID)
			if err != nil {
				return nil, errors.Wrap(err, "cannot get original message")
			}
		}
		if preserveVis {
			v := initialStatus.Visibility
			// We do not set public visibility unless explicitly requested
			if v == "public" {
				opt.visibility = "unlisted"
			} else {
				opt.visibility = v
			}
		}
		if opt.addMentions {
			mentions, err := mentionsList(initialStatus)
			if err != nil {
				return nil, err
			}
			tootText = mentions + tootText
		}
	}

	// Uploading media file last
	if opt.mediaFilePath != "" {
		if len(ids) > 3 {
			return nil, errors.New("too many media attachments")
		}

		fileMediaID, err := uploadFile(opt.mediaFilePath)
		if err != nil {
			return nil, errors.Wrap(err, "cannot attach media file")
		}
		if fileMediaID > 0 {
			ids = append(ids, fileMediaID)
		}
	}

	return gClient.PostStatus(tootText, opt.inReplyToID, ids, opt.sensitive, opt.spoiler, opt.visibility)
}

func mentionsList(s *madon.Status) (string, error) {
	a, err := gClient.GetCurrentAccount()
	if err != nil {
		return "", errors.Wrap(err, "cannot check account details")
	}

	var mentions []string
	// Add the sender if she is not the connected user
	if s.Account.Acct != a.Acct {
		mentions = append(mentions, "@"+s.Account.Acct)
	}
	for _, m := range s.Mentions {
		if m.Acct != a.Acct {
			mentions = append(mentions, "@"+m.Acct)
		}
	}
	mentionsStr := strings.Join(mentions, " ")
	if len(mentionsStr) > 0 {
		return mentionsStr + " ", nil
	}
	return "", nil
}
