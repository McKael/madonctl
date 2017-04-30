// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
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
	"text/template"

	"github.com/m0t0k1ch1/gomif"

	"github.com/McKael/madon"
)

// TemplatePrinter represents a Template printer
type TemplatePrinter struct {
	rawTemplate string
	template    *template.Template
}

// NewPrinterTemplate returns a Template ResourcePrinter
// For TemplatePrinter, the option parameter contains the template string.
func NewPrinterTemplate(option string) (*TemplatePrinter, error) {
	tmpl := option
	t, err := template.New("output").Funcs(template.FuncMap{
		"fromhtml": html2string,
		"fromunix": unix2string,
	}).Parse(tmpl)
	if err != nil {
		return nil, err
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
	case []madon.Account, []madon.Application, []madon.Attachment, []madon.Card,
		[]madon.Client, []madon.Context, []madon.Instance, []madon.Mention,
		[]madon.Notification, []madon.Relationship, []madon.Report,
		[]madon.Results, []madon.Status, []madon.StreamEvent, []madon.Tag,
		[]*gomif.InstanceStatus:
		return p.templateForeach(ot, w)
	}

	return p.templatePrintSingleObj(obj, w)
}

func (p *TemplatePrinter) templatePrintSingleObj(obj interface{}, w io.Writer) error {
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
