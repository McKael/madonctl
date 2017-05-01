// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package colors

import (
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

// Style values
const (
	Bold = iota + 1
	Faint
	Italic
	Underline
	BlinkSlow
	BlinkFast
	Inverse
	Hide
	CrossedOut
)

// Styles contains style names
var Styles = [...]string{
	"bold",
	"faint",
	"italic",
	"underline",
	"blink-slow",
	"blink-fast",
	"inverse",
	"hide",
	"crossed-out",
}

// Color values
const (
	Black = iota
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
)

// Colors contains color names
var Colors = [...]string{
	"black",
	"red",
	"green",
	"yellow",
	"blue",
	"magenta",
	"cyan",
	"white",
}

// ANSICode returns an ANSI escape sequence string for the requested output
func ANSICode(fg, bg, style int) string {
	var seq []string
	if fg >= 0 {
		seq = append(seq, strconv.Itoa(30+fg))
	}
	if bg >= 0 {
		seq = append(seq, strconv.Itoa(40+bg))
	}
	if style >= 0 {
		seq = append(seq, strconv.Itoa(1+style))
	}

	if len(seq) == 0 {
		return "\x1b[" + "0" + "m" // Reset sequence
	}

	return "\x1b[" + strings.Join(seq, ";") + "m"
}

// ANSICodeString returns an ANSI escape sequence string for the requested
// output.
// The description is a coma-separated list of foreground color, background
// color and style: [fg],[bg],[style]
// An empty description or "reset" returns the ANSI reset sequence.
func ANSICodeString(desc string) (string, error) {
	col := [2]int{-1, -1} // fg, bg
	style := -1

	if desc == "" || desc == "reset" {
		return ANSICode(col[0], col[1], style), nil
	}

	styles := strings.SplitN(desc, ",", 3)
	for i, s := range styles {
		if s == "" {
			continue
		}
		switch {
		case i < 2: // Color
			for n, c := range Colors {
				if c == s {
					col[i] = n
					break
				}
			}
			if col[i] == -1 {
				return ANSICode(-1, -1, -1),
					errors.Errorf("color name '%s' not found", s)
			}
		case i == 2: // Style
			for n, sn := range Styles {
				if sn == s {
					style = n
					break
				}
			}
			if style == -1 {
				return ANSICode(-1, -1, -1),
					errors.Errorf("style name '%s' not found", s)
			}
		}
	}

	return ANSICode(col[0], col[1], style), nil
}
