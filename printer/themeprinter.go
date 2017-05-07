// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/m0t0k1ch1/gomif"
	"github.com/pkg/errors"

	"github.com/McKael/madon"
)

const themeDirName = "themes"

// ThemePrinter represents a Theme printer
type ThemePrinter struct {
	name        string
	templateDir string
}

// NewPrinterTheme returns a Theme ResourcePrinter
// For ThemePrinter, the options parameter contains the name of the theme
// and the template base directory (themes are assumed to be in the "themes"
// subdirectory).
func NewPrinterTheme(options Options) (*ThemePrinter, error) {
	name, ok := options["name"]
	if !ok || name == "" {
		return nil, fmt.Errorf("no theme name")
	}
	if strings.IndexRune(name, '/') >= 0 {
		// Should we allow that?  (subthemes, etc.)
		return nil, fmt.Errorf("invalid theme name")
	}
	return &ThemePrinter{
		name:        name,
		templateDir: options["template_directory"],
	}, nil
}

// PrintObj sends the object as text to the writer
// If the writer w is nil, standard output will be used.
func (p *ThemePrinter) PrintObj(obj interface{}, w io.Writer, tmpl string) error {
	if w == nil {
		w = os.Stdout
	}

	if p.name == "" {
		return fmt.Errorf("no theme name") // Should not happen
	}

	var objType string

	switch obj.(type) {
	case []madon.Account, madon.Account, *madon.Account:
		objType = "account"
	case []madon.Application, madon.Application, *madon.Application:
		objType = "application"
	case []madon.Attachment, madon.Attachment, *madon.Attachment:
		objType = "attachment"
	case []madon.Card, madon.Card, *madon.Card:
		objType = "card"
	case []madon.Client, madon.Client, *madon.Client:
		objType = "client"
	case []madon.Context, madon.Context, *madon.Context:
		objType = "context"
	case []madon.Instance, madon.Instance, *madon.Instance:
		objType = "instance"
	case []madon.Mention, madon.Mention, *madon.Mention:
		objType = "mention"
	case []madon.Notification, madon.Notification, *madon.Notification:
		objType = "notification"
	case []madon.Relationship, madon.Relationship, *madon.Relationship:
		objType = "relationship"
	case []madon.Report, madon.Report, *madon.Report:
		objType = "report"
	case []madon.Results, madon.Results, *madon.Results:
		objType = "results"
	case []madon.Status, madon.Status, *madon.Status:
		objType = "status"
	case []madon.StreamEvent, madon.StreamEvent, *madon.StreamEvent:
		objType = "streamEvent"
	case []madon.Tag, madon.Tag, *madon.Tag:
		objType = "tag"

	case []*gomif.InstanceStatus, *gomif.InstanceStatus:
		objType = "instancestatus"
	}

	var rp *ResourcePrinter

	if objType != "" {
		// Check template exists
		templatePath := filepath.Join(p.templateDir, themeDirName, p.name, objType) + ".tmpl"
		if _, err := os.Stat(templatePath); err != nil {
			fmt.Fprintf(os.Stderr, "Warning: theme template not found, falling back to plaintext printer\n") // XXX
		} else {
			t, err := ioutil.ReadFile(templatePath)
			if err != nil {
				return errors.Wrap(err, "cannot read template")
			}
			np, err := NewPrinter("template", Options{"template": string(t)})
			if err != nil {
				return errors.Wrap(err, "cannot create template printer")
			}
			rp = &np
		}
	}

	if rp != nil {
		return (*rp).PrintObj(obj, w, "")
	}

	// No resource printer; let's fall back to plain printer
	// XXX Maybe we should just fail?
	plainP, err := NewPrinter("plain", nil)
	if err != nil {
		return errors.Wrap(err, "cannot create plaintext printer")
	}
	return plainP.PrintObj(obj, w, "")
}
