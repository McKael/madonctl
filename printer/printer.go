// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"fmt"
	"io"
)

// Options contains options used when creating a ResourcePrinter
type Options map[string]string

// ResourcePrinter is an interface used to print objects.
type ResourcePrinter interface {
	// PrintObj receives a runtime object, formats it and prints it to a writer.
	PrintObj(interface{}, io.Writer, string) error
}

// NewPrinter returns a ResourcePrinter for the specified kind of output.
// It returns nil if the output is not supported.
func NewPrinter(output string, options Options) (ResourcePrinter, error) {
	switch output {
	case "", "plain":
		return NewPrinterPlain(options)
	case "json":
		return NewPrinterJSON(options)
	case "yaml":
		return NewPrinterYAML(options)
	case "template":
		return NewPrinterTemplate(options)
	case "theme":
		return NewPrinterTheme(options)
	}
	return nil, fmt.Errorf("unhandled output format")
}
