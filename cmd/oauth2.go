// Copyright © 2017-2018 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	//"github.com/McKael/madonctl/printer"
)

var oauth2Cmd = &cobra.Command{
	Use:   "oauth2",
	Short: "OAuth2 authentication/authorization",
	Example: `  madonctl oauth2                 # Interactive OAuth2 login
  madonctl oauth2 get-url         # Display OAuth2 auhtorization URL
  madonctl oauth2 code CODE       # Enter OAuth2 code

  madonctl oauth2 > config.yaml   # Redirect to configuration file`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return oAuth2Interactive(args)
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize application; do not log in yet
		return madonInit(false)
	},
}

func init() {
	RootCmd.AddCommand(oauth2Cmd)

	// Subcommands
	oauth2Cmd.AddCommand(oauth2Subcommands...)
}

var oauth2Subcommands = []*cobra.Command{
	&cobra.Command{
		Use:   "get-url",
		Short: "Get OAuth2 URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			return oAuth2GetURL()
		},
	},
	&cobra.Command{
		Use:   "code CODE",
		Short: "Log in with OAuth2 code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return oAuth2ExchangeCode(args)
		},
	},
}

func oAuth2GetURL() error {
	// (gClient != nil thanks to PreRun)

	url, err := gClient.LoginOAuth2("", scopes)
	if err != nil {
		return errors.Wrap(err, "OAuth2 authentication failed")
	}

	fmt.Printf("%s\n", url)
	return nil
}

func oAuth2ExchangeCode(args []string) error {
	// (gClient != nil thanks to PreRun)

	if len(args) != 1 {
		return errors.New("wrong usage: code needs 1 argument")
	}

	code := args[0]

	if code == "" {
		return errors.New("no code entered")
	}

	// The code has been set; proceed with token exchange
	_, err := gClient.LoginOAuth2(code, scopes)
	if err != nil {
		return err
	}

	if gClient.UserToken != nil {
		errPrint("Login successful.\n")
		configDump(true)
	}
	return nil
}

// oAuth2Interactive is the default behaviour
func oAuth2Interactive(args []string) error {
	// (gClient != nil thanks to PreRun)

	url, err := gClient.LoginOAuth2("", scopes)
	if err != nil {
		return errors.Wrap(err, "OAuth2 authentication failed")
	}

	fmt.Fprintf(os.Stderr, "Visit the URL for the auth dialog:\n%s\n", url)
	fmt.Fprintf(os.Stderr, "Enter code: ")
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		return err
	}

	if code == "" {
		return errors.New("no code entered")
	}

	// The code has been set; proceed with token exchange
	return oAuth2ExchangeCode([]string{code})
}
