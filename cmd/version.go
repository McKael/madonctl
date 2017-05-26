// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/McKael/madon"
	"github.com/McKael/madonctl/printer"
)

// madonctlVersion contains the version of the madonctl tool
// and the version of the madon library it is linked with.
type madonctlVersion struct {
	AppName      string `json:"application_name"`
	Version      string `json:"version"`
	MadonVersion string `json:"madon_version"`
}

// VERSION of the madonctl application
var VERSION = "1.1.1-dev"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Display " + AppName + " version",
	RunE: func(cmd *cobra.Command, args []string) error {
		const versionTemplate = `This is {{.application_name}} ` +
			`version {{.version}} ` +
			`(using madon library version {{.madon_version}}).{{"\n"}}`
		var v = madonctlVersion{
			AppName:      AppName,
			Version:      VERSION,
			MadonVersion: madon.MadonVersion,
		}
		var p printer.ResourcePrinter
		var err error
		of := getOutputFormat()
		if of == "template" {
			p, err = getPrinter()
		} else { // Default
			pOptions := printer.Options{"template": versionTemplate}
			p, err = printer.NewPrinterTemplate(pOptions)
		}
		if err != nil {
			errPrint("Error: %s", err.Error())
			os.Exit(1)
		}
		return p.PrintObj(v, nil, "")
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
