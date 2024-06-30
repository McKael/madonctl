// Copyright © 2017-2023 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"
	"text/template"
	"time"

	"github.com/kr/text"
	"github.com/mattn/go-isatty"

	"github.com/McKael/madon/v3"
	"github.com/McKael/madonctl/v3/printer/colors"
)

// disableColors can be set to true to disable the color template function
var disableColors bool

// TemplatePrinter represents a Template printer
type TemplatePrinter struct {
	rawTemplate string
	template    *template.Template
}

// NewPrinterTemplate returns a Template ResourcePrinter
// For TemplatePrinter, the options parameter contains the template string.
// The "color_mode" option defines the color behaviour: it can be
// "auto" (default), "on" (forced), "off" (disabled).
func NewPrinterTemplate(options Options) (*TemplatePrinter, error) {
	tmpl := options["template"]
	if tmpl == "" {
		return nil, fmt.Errorf("empty template")
	}
	t, err := template.New("output").Funcs(template.FuncMap{
		"fromhtml": html2string,
		"fromunix": unix2time,
		"tolocal":  dateToLocal,
		"color":    ansiColor,
		"trim":     strings.TrimSpace,
		"wrap":     wrap,
	}).Parse(tmpl)
	if err != nil {
		return nil, err
	}

	// Update disableColors.
	// In auto-mode, check if stdout is a TTY.
	colorMode := options["color_mode"]
	if colorMode == "off" || (colorMode != "on" && !isatty.IsTerminal(os.Stdout.Fd())) {
		disableColors = true
	}

	return &TemplatePrinter{
		rawTemplate: tmpl,
		template:    t,
	}, nil
}

// PrintObj sends the object as text to the writer
// If the writer w is nil, standard output will be used.
func (p *TemplatePrinter) PrintObj(obj interface{}, w io.Writer, tmpl string) error {
	if w == nil {
		w = os.Stdout
	}

	if p.template == nil {
		return fmt.Errorf("template not built")
	}

	switch ot := obj.(type) { // I wish I knew a better way...
	case []madon.Account, []madon.Application, []madon.Attachment,
		[]madon.Card, []madon.Client, []madon.Context, []madon.Emoji,
		[]madon.Instance, []madon.List, []madon.Mention,
		[]madon.Notification, []madon.Relationship, []madon.Report,
		[]madon.Results, []madon.Status, []madon.StreamEvent,
		[]madon.Tag, []string:
		return p.templateForeach(ot, w)
	}

	return p.templatePrintSingleObj(obj, w)
}

func (p *TemplatePrinter) templatePrintSingleObj(obj interface{}, w io.Writer) error {
	if s, ok := obj.(string); ok {
		// obj is a simple string
		if err := p.safeExecute(w, s); err != nil {
			return fmt.Errorf("error executing template %q: %v", p.rawTemplate, err)
		}
		return nil
	}

	// This code comes from Kubernetes.
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	out := map[string]interface{}{}
	if err := json.Unmarshal(data, &out); err != nil {
		return err
	}
	if err = p.safeExecute(w, out); err != nil {
		return fmt.Errorf("error executing template %q: %v", p.rawTemplate, err)
	}
	return nil
}

// safeExecute tries to execute the template, but catches panics and returns an error
// should the template engine panic.
// This code comes from Kubernetes.
func (p *TemplatePrinter) safeExecute(w io.Writer, obj interface{}) error {
	var panicErr error
	// Sorry for the double anonymous function. There's probably a clever way
	// to do this that has the defer'd func setting the value to be returned, but
	// that would be even less obvious.
	retErr := func() error {
		defer func() {
			if x := recover(); x != nil {
				panicErr = fmt.Errorf("caught panic: %+v", x)
			}
		}()
		return p.template.Execute(w, obj)
	}()
	if panicErr != nil {
		return panicErr
	}
	return retErr
}

func (p *TemplatePrinter) templateForeach(ol interface{}, w io.Writer) error {
	switch reflect.TypeOf(ol).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(ol)

		for i := 0; i < s.Len(); i++ {
			o := s.Index(i).Interface()
			if err := p.templatePrintSingleObj(o, w); err != nil {
				return err
			}
		}
	}
	return nil
}

func ansiColor(desc string) (string, error) {
	if disableColors {
		return "", nil
	}
	return colors.ANSICodeString(desc)
}

// Parse datetime string from RFC3339 (default format in templates because
// of implicit conversion to string) and return a local time.
func dateToLocal(s string) (time.Time, error) {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return t, err
	}
	return t.Local(), err
}

// Wrap text with indent prefix
func wrap(indent string, lineLength int, txt string) string {
	width := lineLength - len(indent)
	if width < 10 {
		width = 10
	}

	lines := strings.SplitAfter(txt, "\n")
	var out []string
	for _, l := range lines {
		l = strings.TrimLeft(l, " ")
		out = append(out, text.Indent(text.Wrap(l, width), indent))
	}
	return strings.Join(out, "\n")
}
