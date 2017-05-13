// Copyright (c) 2015 Shawn Goertzen
// Copyright (c) 2017 Mikael Berthe
//
// This code mostly comes from github.com/sgoertzen/html2text,
// with some specific but intrusive changes for Mastodon HTML messages.
// For example, links are not displayed for hashtags and mentions,
// and links alone are displayed for the other cases.
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package html2text

import (
	"bytes"
	"errors"
	"golang.org/x/net/html"
	"strings"
)

var breakers = map[string]bool{
	"br":  true,
	"div": true,
	"tr":  true,
	"li":  true,
	"p":   true,
}

// Textify turns an HTML body into a text string
func Textify(body string) (string, error) {
	r := strings.NewReader(body)
	doc, err := html.Parse(r)
	if err != nil {
		return "", errors.New("unable to parse the html")
	}
	var buffer bytes.Buffer
	process(doc, &buffer, "")

	s := strings.TrimSpace(buffer.String())
	return s, nil
}

func process(n *html.Node, b *bytes.Buffer, class string) {
	processChildren := true

	if n.Type == html.ElementNode && n.Data == "head" {
		return
	} else if n.Type == html.ElementNode && n.Data == "a" && n.FirstChild != nil {
		anchor(n, b, class)
		processChildren = false
	} else if n.Type == html.TextNode {
		// Clean up data
		cleanData := strings.Replace(strings.Trim(n.Data, " \t"), "\u00a0", " ", -1)

		// Heuristics to add a whitespace character...
		var prevSpace, nextSpace bool // hint if previous/next char is a space
		var last byte
		bl := b.Len()
		if bl > 0 {
			last = b.Bytes()[bl-1]
			if last == ' ' {
				prevSpace = true
			}
		}
		if len(cleanData) > 0 && cleanData[0] == ' ' {
			nextSpace = true
		}
		if prevSpace && nextSpace {
			b.WriteString(cleanData[1:]) // Trim 1 space
		} else {
			if bl > 0 && last != '\n' && last != '@' && last != '#' && !prevSpace && !nextSpace {
				b.WriteString(" ")
			}
			b.WriteString(cleanData)
		}
	}

	if processChildren {
		var class string
		if n.Type == html.ElementNode && n.Data == "span" {
			for _, attr := range n.Attr {
				if attr.Key == "class" {
					class = attr.Val
					break
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			process(c, b, class)
		}
	}

	if b.Len() > 0 {
		bl := b.Len()
		last := b.Bytes()[bl-1]
		if last != '\n' && n.Type == html.ElementNode && breakers[n.Data] {
			// Remove previous space
			for last == ' ' {
				bl--
				b.Truncate(bl)
				if bl > 0 {
					last = b.Bytes()[bl-1]
				} else {
					last = '\x00'
				}
			}
			b.WriteString("\n")
		}
	}
}

func anchor(n *html.Node, b *bytes.Buffer, class string) {
	bl := b.Len()
	var last byte
	if bl > 0 {
		last = b.Bytes()[bl-1]
	}

	// Add heading space if needed
	if last != ' ' && last != '\n' && last != '#' && last != '@' {
		b.WriteString(" ")
	}

	var tmpbuf bytes.Buffer
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		process(c, &tmpbuf, class)
	}

	if class == "tag" || class == "h-card" || last == '@' {
		b.Write(tmpbuf.Bytes())
		return
	}

	s := tmpbuf.String()
	if strings.HasPrefix(s, "#") || strings.HasPrefix(s, "@") {
		b.WriteString(s) // Tag or mention: display content
		return
	}

	// Display href link
	for _, attr := range n.Attr {
		if attr.Key == "href" {
			link := n.Attr[0].Val
			b.WriteString(link)
			break
		}
	}
}
