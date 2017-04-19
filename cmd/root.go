// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/McKael/madon"
	"github.com/McKael/madonctl/printer"
)

// AppName is the CLI application name
const AppName = "madonctl"
const defaultConfigFile = "$HOME/.config/" + AppName + "/" + AppName + ".yaml"

var cfgFile string
var safeMode bool
var instanceURL, appID, appSecret string
var login, password, token string
var gClient *madon.Client
var verbose bool
var outputFormat string
var outputTemplate, outputTemplateFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:               AppName,
	Short:             "A CLI utility for Mastodon API",
	PersistentPreRunE: checkOutputFormat,
	Long: `madonctl is a CLI tool for the Mastodon REST API.

You can use a configuration file to store common options.
For example, create ` + defaultConfigFile + ` with the following
contents:

	---
	instance: "INSTANCE"
	login: "USERNAME"
	password: "USERPASSWORD"
	...

The simplest way to generate a configuration file is to use the 'config dump'
command.

(Configuration files in JSON are also accepted.)

If you want shell auto-completion (for bash or zsh), you can generate the
completion scripts with "madonctl completion $SHELL".
For example if you use bash:

	madonctl completion bash > _bash_madonctl
	source _bash_madonctl

Now you should have tab completion for subcommands and flags.

Note: Most examples assume the user's credentials are set in the configuration
file.
`,
	Example: `  madonctl instance
  madonctl toot "Hello, World"
  madonctl toot --visibility direct "@McKael Hello, You"
  madonctl toot --visibility private --spoiler CW "The answer was 42"
  madonctl post --file image.jpg Selfie
  madonctl --instance INSTANCE --login USERNAME --password PASS timeline
  madonctl accounts notifications --list --clear
  madonctl accounts blocked
  madonctl accounts search Gargron
  madonctl search --resolve https://mastodon.social/@Gargron
  madonctl accounts follow --remote Gargron@mastodon.social
  madonctl accounts --account-id 399 statuses
  madonctl status --status-id 416671 show
  madonctl status --status-id 416671 favourite
  madonctl status --status-id 416671 boost
  madonctl accounts show
  madonctl accounts show -o yaml
  madonctl accounts --account-id 1 followers --template '{{.acct}}{{"\n"}}'
  madonctl config whoami
  madonctl timeline :mastodon`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		errPrint("Error: %s", err.Error())
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "",
		"config file (default is "+defaultConfigFile+")")
	RootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose mode")
	RootCmd.PersistentFlags().StringVarP(&instanceURL, "instance", "i", "", "Mastodon instance")
	RootCmd.PersistentFlags().StringVarP(&login, "login", "L", "", "Instance user login")
	RootCmd.PersistentFlags().StringVarP(&password, "password", "P", "", "Instance user password")
	RootCmd.PersistentFlags().StringVarP(&token, "token", "t", "", "User token")
	RootCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "plain",
		"Output format (plain|json|yaml|template)")
	RootCmd.PersistentFlags().StringVar(&outputTemplate, "template", "",
		"Go template (for output=template)")
	RootCmd.PersistentFlags().StringVar(&outputTemplateFile, "template-file", "",
		"Go template file (for output=template)")

	// Configuration file bindings
	viper.BindPFlag("output", RootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("verbose", RootCmd.PersistentFlags().Lookup("verbose"))
	// XXX viper.BindPFlag("apiKey", RootCmd.PersistentFlags().Lookup("api-key"))
	viper.BindPFlag("instance", RootCmd.PersistentFlags().Lookup("instance"))
	viper.BindPFlag("login", RootCmd.PersistentFlags().Lookup("login"))
	viper.BindPFlag("password", RootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("token", RootCmd.PersistentFlags().Lookup("token"))
}

func checkOutputFormat(cmd *cobra.Command, args []string) error {
	of := viper.GetString("output")
	switch of {
	case "", "plain", "json", "yaml", "template":
		return nil // Accepted
	}
	return fmt.Errorf("output format '%s' not supported", of)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName(AppName) // name of config file (without extension)
	viper.AddConfigPath("$HOME/.config/" + AppName)
	viper.AddConfigPath("$HOME/." + AppName)

	// Read in environment variables that match, with a prefix
	viper.SetEnvPrefix(AppName)
	viper.AutomaticEnv()

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); viper.GetBool("verbose") && err == nil {
		errPrint("Using config file: %s", viper.ConfigFileUsed())
	}
}

// getOutputFormat return the requested output format, defaulting to "plain".
func getOutputFormat() string {
	of := viper.GetString("output")
	if of == "" {
		of = "plain"
	}
	// Override format if a template is provided
	if of == "plain" && (outputTemplate != "" || outputTemplateFile != "") {
		// If the format is plain and there is a template option,
		// set the format to "template".
		of = "template"
	}
	return of
}

// getPrinter returns a resource printer for the requested output format.
func getPrinter() (printer.ResourcePrinter, error) {
	var opt string
	of := getOutputFormat()

	if of == "template" {
		opt = outputTemplate
		if outputTemplateFile != "" {
			tmpl, err := ioutil.ReadFile(outputTemplateFile)
			if err != nil {
				return nil, err
			}
			opt = string(tmpl)
		}
	}
	return printer.NewPrinter(of, opt)
}

func errPrint(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format+"\n", a...)
}
