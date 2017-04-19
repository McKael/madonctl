// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"fmt"
	"io"
)

// ResourcePrinter is an interface used to print objects.
type ResourcePrinter interface {
	// PrintObj receives a runtime object, formats it and prints it to a writer.
	PrintObj(interface{}, io.Writer, string) error
}

// NewPrinter returns a ResourcePrinter for the specified kind of output.
// It returns nil if the output is not supported.
func NewPrinter(output, option string) (ResourcePrinter, error) {
	switch output {
	case "", "plain":
		return NewPrinterPlain(option)
	case "json":
		return NewPrinterJSON(option)
	case "yaml":
		return NewPrinterYAML(option)
	case "template":
		return NewPrinterTemplate(option)
	}
	return nil, fmt.Errorf("unhandled output format")
}
