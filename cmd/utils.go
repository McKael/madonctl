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
	case "", "plain", "json", "yaml", "template", "theme":
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
	if of == "plain" {
		// If the format is plain and there is a template option,
		// set the format to "template".  Same for "theme".
		if outputTemplate != "" || outputTemplateFile != "" {
			of = "template"
		} else if outputTheme != "" {
			of = "theme"
		}
	}
	return of
}

type mcPrinter struct {
	printer.ResourcePrinter
	command string
}

type mcResourcePrinter interface {
	printer.ResourcePrinter
	printObj(interface{}) error
	setCommand(string)
}

// getPrinter returns a resource printer for the requested output format.
func getPrinter() (mcResourcePrinter, error) {
	opt := make(printer.Options)
	of := getOutputFormat()

	// Initialize color mode
	switch viper.GetString("color") {
	case "on", "yes", "force":
		opt["color_mode"] = "on"
	case "off", "no":
		opt["color_mode"] = "off"
	default:
		opt["color_mode"] = "auto"
	}

	if of == "theme" {
		opt["name"] = outputTheme
		opt["template_directory"] = viper.GetString("template_directory")
	} else if of == "template" {
		opt["template"] = outputTemplate
		if outputTemplateFile != "" {
			tmpl, err := readTemplate(outputTemplateFile, viper.GetString("template_directory"))
			if err != nil {
				return nil, err
			}
			opt["template"] = string(tmpl)
		}
	}
	var mcrp mcPrinter
	p, err := printer.NewPrinter(of, opt)
	if err != nil {
		return mcrp, err
	}
	mcrp.ResourcePrinter = p
	return mcrp, nil
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

func getThemes() ([]string, error) {
	templDir := viper.GetString("template_directory")
	if templDir == "" {
		return nil, errors.New("template_directory not defined")
	}
	files, err := ioutil.ReadDir(filepath.Join(templDir, "themes"))
	if err != nil {
		return nil, errors.Wrap(err, "cannot read theme directory")
	}
	var tl []string
	for _, f := range files {
		if f.IsDir() {
			tl = append(tl, f.Name())
		}
	}
	return tl, nil
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

func (mcp mcPrinter) printObj(obj interface{}) error {
	return mcp.PrintObj(obj, nil, "")
}

func (mcp mcPrinter) setCommand(cmd string) {
	mcp.command = cmd
}
