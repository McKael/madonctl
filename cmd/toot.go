// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/McKael/madon"
)

// toot is a kind of alias for status post

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

	// Flag completion
	annotation := make(map[string][]string)
	annotation[cobra.BashCompCustom] = []string{"__madonctl_visibility"}

	tootAliasCmd.Flags().Lookup("visibility").Annotations = annotation
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
  echo "Hello from #madonctl" | madonctl toot --visibility unlisted --stdin`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := madonInit(true); err != nil {
			return err
		}
		return statusSubcommandRunE("post", args)
	},
}

func toot(tootText string) (*madon.Status, error) {
	opt := statusOpts

	switch opt.visibility {
	case "", "direct", "private", "unlisted", "public":
		// OK
	default:
		return nil, errors.New("invalid visibility argument value")
	}

	if opt.inReplyToID < 0 {
		return nil, errors.New("invalid in-reply-to argument value")
	}

	ids, err := splitIDs(opt.mediaIDs)
	if err != nil {
		return nil, errors.New("cannot parse media IDs")
	}

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

	if tootText == "" && len(ids) == 0 && opt.spoiler == "" {
		return nil, errors.New("toot is empty")
	}

	return gClient.PostStatus(tootText, opt.inReplyToID, ids, opt.sensitive, opt.spoiler, opt.visibility)
}
