// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

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
	&cobra.Command{
		Use: "themes",
		//Aliases: []string{},
		Short: "Display available themes",
		RunE: func(cmd *cobra.Command, args []string) error {
			return configDisplayThemes()
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
	// Try to sign in if a login was provided
	if viper.GetString("token") != "" || viper.GetString("login") != "" {
		if err := madonLogin(); err != nil {
			errPrint("Error: could not log in: %v", err)
			os.Exit(-1)
		}
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
		pOptions := printer.Options{"template": configurationTemplate}
		p, err = printer.NewPrinterTemplate(pOptions)
	} else {
		p, err = getPrinter()
	}
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
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
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.PrintObj(gClient.UserToken, nil, "")
}

// configDisplayThemes lists the available themes
// It is intended for shell completion.
func configDisplayThemes() error {
	var p printer.ResourcePrinter

	themes, err := getThemes()
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}

	if getOutputFormat() == "plain" {
		pOptions := printer.Options{"template": `{{printf "%s\n" .}}`}
		p, err = printer.NewPrinterTemplate(pOptions)
	} else {
		p, err = getPrinter()
	}
	if err != nil {
		errPrint("Error: %v", err)
		os.Exit(1)
	}
	return p.PrintObj(themes, nil, "")
}
