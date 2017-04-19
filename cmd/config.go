// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/McKael/madonctl/printer"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display configuration",
}

func init() {
	RootCmd.AddCommand(configCmd)

	// Subcommands
	configCmd.AddCommand(configSubcommands...)
}

var configSubcommands = []*cobra.Command{
	&cobra.Command{
		Use:   "dump",
		Short: "Dump the configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configDump()
		},
	},
	&cobra.Command{
		Use:     "whoami",
		Aliases: []string{"token"},
		Short:   "Display user token",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configDisplayToken()
		},
	},
}

const configurationTemplate = `---
instance: '{{.InstanceURL}}'
app_id: '{{.ID}}'
app_secret: '{{.Secret}}'

{{if .UserToken}}token: {{.UserToken.access_token}}{{else}}#token: ''{{end}}
#login: ''
#password: ''
safe_mode: true
...
`

func configDump() error {
	if viper.GetBool("safe_mode") {
		errPrint("Cannot dump: disabled by configuration (safe_mode)")
		return nil
	}

	if err := madonInitClient(); err != nil {
		return err
	}
	// Try to sign in, but don't mind if it fails
	if err := madonLogin(); err != nil {
		errPrint("Info: could not log in: %s", err)
	}

	var p printer.ResourcePrinter
	var err error

	if getOutputFormat() == "plain" {
		cfile := viper.ConfigFileUsed()
		if cfile == "" {
			cfile = defaultConfigFile
		}
		errPrint("You can copy the following lines into a configuration file.")
		errPrint("E.g. %s -i INSTANCE -L USERNAME -P PASS config dump > %s\n", AppName, cfile)
		p, err = printer.NewPrinterTemplate(configurationTemplate)
	} else {
		p, err = getPrinter()
	}
	if err != nil {
		return err
	}
	return p.PrintObj(gClient, nil, "")
}

func configDisplayToken() error {
	if viper.GetBool("safe_mode") {
		errPrint("Cannot dump: disabled by configuration (safe_mode)")
		return nil
	}

	if err := madonInit(true); err != nil {
		return err
	}

	p, err := getPrinter()
	if err != nil {
		return err
	}
	return p.PrintObj(gClient.UserToken, nil, "")
}
