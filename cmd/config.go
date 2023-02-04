// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"os"

	"github.com/mattn/go-isatty"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/McKael/madonctl/printer"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Display configuration",
	Long: `Display configuration

Display current configuration.  You can use this command to generate an
initial configuration file (see the examples below).

This command is disabled if the safe_mode setting is set to true in the
configuration file.`,
	Example: `  madonctl config dump -i INSTANCE -L USERNAME -P PASS > config.yaml
  madonctl whoami
  madonctl whoami --template '{{.access_token}}'`,
}

func init() {
	RootCmd.AddCommand(configCmd)

	// Subcommands
	configCmd.AddCommand(configSubcommands...)
}

var configSubcommands = []*cobra.Command{
	&cobra.Command{
		Use:     "dump",
		Short:   "Dump the configuration",
		Example: `  madonctl config dump -i INSTANCE -L USERNAME -P PASS > config.yaml`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return configDump(false)
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

#default_visibility: unlisted

#template_directory: ''
#default_output: theme
#default_theme: ansi
#color: auto
#verbose: false
...
`

func configDump(oauth2 bool) error {
	if !oauth2 {
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
	}

	var p printer.ResourcePrinter
	var err error

	if getOutputFormat() == "plain" {
		cfile := viper.ConfigFileUsed()
		if cfile == "" {
			cfile = defaultConfigFile
		}
		if isatty.IsTerminal(os.Stdout.Fd()) {
			errPrint("You can copy the following lines into a configuration file.")
			errPrint("E.g. %s -i INSTANCE -L USERNAME -P PASS config dump > %s", AppName, cfile)
			errPrint(" or  %s -i INSTANCE oauth2 > %s\n", AppName, cfile)
		}
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
	return p.printObj(gClient.UserToken)
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
