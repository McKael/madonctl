// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"fmt"
	"io"
	"os"

	"github.com/ghodss/yaml"
)

// YAMLPrinter represents a YAML printer
type YAMLPrinter struct {
}

// NewPrinterYAML returns a YAML ResourcePrinter
func NewPrinterYAML(option string) (*YAMLPrinter, error) {
	return &YAMLPrinter{}, nil
}

// PrintObj sends the object as text to the writer
// If the writer w is nil, standard output will be used.
// For YAMLPrinter, the option parameter is currently not used.
func (p *YAMLPrinter) PrintObj(obj interface{}, w io.Writer, option string) error {
	if w == nil {
		w = os.Stdout
	}

	//yamlEncoder := yaml.NewEncoder(w)
	//return yamlEncoder.Encode(obj)

	output, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}
	_, err = fmt.Fprint(w, string(output))
	return err
}
