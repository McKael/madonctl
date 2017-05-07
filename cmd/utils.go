// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/McKael/madonctl/printer"
)

func checkOutputFormat(cmd *cobra.Command, args []string) error {
	of := viper.GetString("output")
	switch of {
	case "", "plain", "json", "yaml", "template":
		return nil // Accepted
	}
	return errors.Errorf("output format '%s' not supported", of)
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

	// Initialize color mode
	switch viper.GetString("color") {
	case "on", "yes", "force":
		printer.ColorMode = 1
	case "off", "no":
		printer.ColorMode = 2
	}

	if of == "template" {
		opt = outputTemplate
		if outputTemplateFile != "" {
			tmpl, err := readTemplate(outputTemplateFile, viper.GetString("template_directory"))
			if err != nil {
				return nil, err
			}
			opt = string(tmpl)
		}
	}
	return printer.NewPrinter(of, opt)
}

func readTemplate(name, templateDir string) ([]byte, error) {
	if strings.HasPrefix(name, "/") || strings.HasPrefix(name, "./") || strings.HasPrefix(name, "../") {
		return ioutil.ReadFile(name)
	}

	if templateDir != "" {
		// If the template file can be found in the template directory,
		// use this file.
		fullName := filepath.Join(templateDir, name)
		if fileExists(fullName) {
			name = fullName
		}
	}

	return ioutil.ReadFile(name)
}

func fileExists(filename string) bool {
	if _, err := os.Stat(filename); err != nil {
		return false
	}
	return true
}

func errPrint(format string, a ...interface{}) (n int, err error) {
	return fmt.Fprintf(os.Stderr, format+"\n", a...)
}
