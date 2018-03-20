// Copyright © 2017-2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	flag "github.com/spf13/pflag"

	"github.com/McKael/madon"
)

var mediaFlags *flag.FlagSet

var mediaOpts struct {
	mediaID     int64
	filePath    string
	description string
	focus       string
}

// mediaCmd represents the media command
var mediaCmd = &cobra.Command{
	Use:     "media --file FILENAME",
	Aliases: []string{"upload"},
	Short:   "Upload or update a media attachment",
	Long: `Upload or update a media attachment

This command can be used to upload media that will be attached to
a status later.

A media description or focal point (focus) can be updated
as long as it is not yet attached to a status, with the '--update MEDIA_ID'
option.`,
	Example: `  madonctl upload --file FILENAME
  madonctl media --file FILENAME --description "My screenshot"
  madonctl media --update 3217821 --focus "0.5,-0.7"
  madonctl media --update 2468123 --description "Winter Snow"`,
	RunE: mediaRunE,
}

func init() {
	RootCmd.AddCommand(mediaCmd)

	mediaCmd.Flags().StringVar(&mediaOpts.filePath, "file", "", "Path of the media file")
	mediaCmd.Flags().Int64Var(&mediaOpts.mediaID, "update", 0, "Media to update (ID)")

	mediaCmd.Flags().StringVar(&mediaOpts.description, "description", "", "Plain text description")
	mediaCmd.Flags().StringVar(&mediaOpts.focus, "focus", "", "Focal point")

	// This will be used to check if the options were explicitly set or not
	mediaFlags = mediaCmd.Flags()
}

func mediaRunE(cmd *cobra.Command, args []string) error {
	opt := mediaOpts

	if opt.filePath == "" {
		if opt.mediaID < 1 {
			return errors.New("no media file name provided")
		}
	} else if opt.mediaID > 0 {
		return errors.New("cannot use both --file and --update")
	}

	if err := madonInit(true); err != nil {
		return err
	}

	var attachment *madon.Attachment
	var err error

	if opt.filePath != "" {
		attachment, err = gClient.UploadMedia(opt.filePath, opt.description, opt.focus)
	} else {
		// Update
		var desc, foc *string
		if mediaFlags.Lookup("description").Changed {
			desc = &opt.description
		}
		if mediaFlags.Lookup("focus").Changed {
			foc = &opt.focus
		}
		attachment, err = gClient.UpdateMedia(opt.mediaID, desc, foc)
	}
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}

	p, err := getPrinter()
	if err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(1)
	}
	return p.printObj(attachment)
}

// uploadFile uploads a media file and returns the attachment ID
func uploadFile(filePath string) (int64, error) {
	attachment, err := gClient.UploadMedia(filePath, "", "")
	if err != nil {
		return 0, err
	}
	if attachment == nil {
		return 0, nil
	}
	return attachment.ID, nil
}
