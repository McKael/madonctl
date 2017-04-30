// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

var mediaOpts struct {
	filePath string
}

// mediaCmd represents the media command
var mediaCmd = &cobra.Command{
	Use:     "media --file FILENAME",
	Aliases: []string{"upload"},
	Short:   "Upload a media attachment",
	//Long: `TBW...`,
	RunE: mediaRunE,
}

func init() {
	RootCmd.AddCommand(mediaCmd)

	mediaCmd.Flags().StringVar(&mediaOpts.filePath, "file", "", "Path of the media file")
	mediaCmd.MarkFlagRequired("file")
}

func mediaRunE(cmd *cobra.Command, args []string) error {
	opt := mediaOpts

	if opt.filePath == "" {
		return errors.New("no media file name provided")
	}

	if err := madonInit(true); err != nil {
		return err
	}

	attachment, err := gClient.UploadMedia(opt.filePath)
	if err != nil {
		errPrint("Error: %s", err.Error())
		return nil
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(attachment, nil, "")
}

// uploadFile uploads a media file and returns the attachment ID
func uploadFile(filePath string) (int64, error) {
	attachment, err := gClient.UploadMedia(filePath)
	if err != nil {
		return 0, err
	}
	if attachment == nil {
		return 0, nil
	}
	return attachment.ID, nil
}
