// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"encoding/json"
	"io"
	"os"
)

// JSONPrinter represents a JSON printer
type JSONPrinter struct {
}

// NewPrinterJSON returns a JSON ResourcePrinter
func NewPrinterJSON(options Options) (*JSONPrinter, error) {
	return &JSONPrinter{}, nil
}

// PrintObj sends the object as text to the writer
// If the writer w is nil, standard output will be used.
// For JSONPrinter, the option parameter is currently not used.
func (p *JSONPrinter) PrintObj(obj interface{}, w io.Writer, option string) error {
	if w == nil {
		w = os.Stdout
	}

	jsonEncoder := json.NewEncoder(w)
	//jsonEncoder.SetIndent("", "  ")
	return jsonEncoder.Encode(obj)
}
