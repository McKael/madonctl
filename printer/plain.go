// Copyright © 2017 Mikael Berthe <mikael@lilotux.net>
//
// Licensed under the MIT license.
// Please see the LICENSE file is this directory.

package printer

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"time"

	"github.com/jaytaylor/html2text"
	"github.com/m0t0k1ch1/gomif"

	"github.com/McKael/madon"
)

// PlainPrinter is the default "plain text" printer
type PlainPrinter struct {
	Indent      string
	NoSubtitles bool
}

// NewPrinterPlain returns a plaintext ResourcePrinter
// For PlainPrinter, the option parameter contains the indent prefix.
func NewPrinterPlain(option string) (*PlainPrinter, error) {
	indentInc := "  "
	if option != "" {
		indentInc = option
	}
	return &PlainPrinter{Indent: indentInc}, nil
}

// PrintObj sends the object as text to the writer
// If the writer w is nil, standard output will be used.
// For PlainPrinter, the option parameter contains the initial indent.
func (p *PlainPrinter) PrintObj(obj interface{}, w io.Writer, initialIndent string) error {
	if w == nil {
		w = os.Stdout
	}
	switch o := obj.(type) {
	case []madon.Account, []madon.Attachment, []madon.Card, []madon.Context,
		[]madon.Instance, []madon.Mention, []madon.Notification,
		[]madon.Relationship, []madon.Report, []madon.Results,
		[]madon.Status, []madon.StreamEvent, []madon.Tag,
		[]*gomif.InstanceStatus:
		return p.plainForeach(o, w, initialIndent)
	case *madon.Account:
		return p.plainPrintAccount(o, w, initialIndent)
	case madon.Account:
		return p.plainPrintAccount(&o, w, initialIndent)
	case *madon.Attachment:
		return p.plainPrintAttachment(o, w, initialIndent)
	case madon.Attachment:
		return p.plainPrintAttachment(&o, w, initialIndent)
	case *madon.Card:
		return p.plainPrintCard(o, w, initialIndent)
	case madon.Card:
		return p.plainPrintCard(&o, w, initialIndent)
	case *madon.Context:
		return p.plainPrintContext(o, w, initialIndent)
	case madon.Context:
		return p.plainPrintContext(&o, w, initialIndent)
	case *madon.Instance:
		return p.plainPrintInstance(o, w, initialIndent)
	case madon.Instance:
		return p.plainPrintInstance(&o, w, initialIndent)
	case *madon.Notification:
		return p.plainPrintNotification(o, w, initialIndent)
	case madon.Notification:
		return p.plainPrintNotification(&o, w, initialIndent)
	case *madon.Relationship:
		return p.plainPrintRelationship(o, w, initialIndent)
	case madon.Relationship:
		return p.plainPrintRelationship(&o, w, initialIndent)
	case *madon.Report:
		return p.plainPrintReport(o, w, initialIndent)
	case madon.Report:
		return p.plainPrintReport(&o, w, initialIndent)
	case *madon.Results:
		return p.plainPrintResults(o, w, initialIndent)
	case madon.Results:
		return p.plainPrintResults(&o, w, initialIndent)
	case *madon.Status:
		return p.plainPrintStatus(o, w, initialIndent)
	case madon.Status:
		return p.plainPrintStatus(&o, w, initialIndent)
	case *madon.UserToken:
		return p.plainPrintUserToken(o, w, initialIndent)
	case madon.UserToken:
		return p.plainPrintUserToken(&o, w, initialIndent)
	case *gomif.InstanceStatus:
		return p.plainPrintInstanceStatistics(o, w, initialIndent)
	case gomif.InstanceStatus:
		return p.plainPrintInstanceStatistics(&o, w, initialIndent)
	}
	// TODO: Mention
	// TODO: StreamEvent
	// TODO: Tag

	return fmt.Errorf("PlainPrinter not yet implemented for %T (try json or yaml...)", obj)
}

func (p *PlainPrinter) plainForeach(ol interface{}, w io.Writer, ii string) error {
	switch reflect.TypeOf(ol).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(ol)

		for i := 0; i < s.Len(); i++ {
			o := s.Index(i).Interface()
			if err := p.PrintObj(o, w, ii); err != nil {
				return err
			}
		}
	}
	return nil
}

func html2string(h string) string {
	t, err := html2text.FromString(h)
	if err == nil {
		return t
	}
	return h // Failed: return initial string
}

func indentedPrint(w io.Writer, indent string, title, skipIfEmpty bool, label string, format string, args ...interface{}) {
	prefix := indent
	if title {
		prefix += "- "
	} else {
		prefix += "  "
	}
	value := fmt.Sprintf(format, args...)
	if !title && skipIfEmpty && len(value) == 0 {
		return
	}
	fmt.Fprintf(w, "%s%s: %s\n", prefix, label, value)
}

func (p *PlainPrinter) plainPrintAccount(a *madon.Account, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Account ID", "%d (%s)", a.ID, a.Username)
	indentedPrint(w, indent, false, false, "User ID", "%s", a.Acct)
	indentedPrint(w, indent, false, false, "Display name", "%s", a.DisplayName)
	indentedPrint(w, indent, false, false, "Creation date", "%v", a.CreatedAt.Local())
	indentedPrint(w, indent, false, false, "URL", "%s", a.URL)
	indentedPrint(w, indent, false, false, "Statuses count", "%d", a.StatusesCount)
	indentedPrint(w, indent, false, false, "Followers count", "%d", a.FollowersCount)
	indentedPrint(w, indent, false, false, "Following count", "%d", a.FollowingCount)
	if a.Locked {
		indentedPrint(w, indent, false, false, "Locked", "%v", a.Locked)
	}
	indentedPrint(w, indent, false, true, "User note", "%s", html2string(a.Note)) // XXX too long?
	return nil
}

func (p *PlainPrinter) plainPrintAttachment(a *madon.Attachment, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Attachment ID", "%d", a.ID)
	indentedPrint(w, indent, false, false, "Type", "%s", a.Type)
	indentedPrint(w, indent, false, false, "Local URL", "%s", a.URL)
	indentedPrint(w, indent, false, true, "Remote URL", "%s", a.RemoteURL)
	indentedPrint(w, indent, false, true, "Preview URL", "%s", a.PreviewURL)
	indentedPrint(w, indent, false, true, "Text URL", "%s", a.PreviewURL)
	return nil
}

func (p *PlainPrinter) plainPrintCard(c *madon.Card, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Card title", "%s", c.Title)
	indentedPrint(w, indent, false, true, "Description", "%s", c.Description)
	indentedPrint(w, indent, false, true, "URL", "%s", c.URL)
	indentedPrint(w, indent, false, true, "Image", "%s", c.Image)
	return nil
}

func (p *PlainPrinter) plainPrintContext(c *madon.Context, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Context", "%d relative(s)", len(c.Ancestors)+len(c.Descendents))
	if len(c.Ancestors) > 0 {
		indentedPrint(w, indent, false, false, "Ancestors", "")
		p.PrintObj(c.Ancestors, w, indent+p.Indent)
	}
	if len(c.Descendents) > 0 {
		indentedPrint(w, indent, false, false, "Descendents", "")
		p.PrintObj(c.Descendents, w, indent+p.Indent)
	}
	return nil
}

func (p *PlainPrinter) plainPrintInstance(i *madon.Instance, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Instance title", "%s", i.Title)
	indentedPrint(w, indent, false, true, "Description", "%s", html2string(i.Description))
	indentedPrint(w, indent, false, true, "URL", "%s", i.URI)
	indentedPrint(w, indent, false, true, "Email", "%s", i.Email)
	indentedPrint(w, indent, false, true, "Version", "%s", i.Version)
	return nil
}

func (p *PlainPrinter) plainPrintNotification(n *madon.Notification, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Notification ID", "%d", n.ID)
	indentedPrint(w, indent, false, false, "Type", "%s", n.Type)
	indentedPrint(w, indent, false, false, "Timestamp", "%v", n.CreatedAt.Local())
	if n.Account != nil {
		indentedPrint(w, indent+p.Indent, true, false, "Account", "(%d) @%s - %s",
			n.Account.ID, n.Account.Acct, n.Account.DisplayName)
	}
	if n.Status != nil {
		p.plainPrintStatus(n.Status, w, indent+p.Indent)
	}
	return nil
}

func (p *PlainPrinter) plainPrintRelationship(r *madon.Relationship, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "ID", "%d", r.ID)
	indentedPrint(w, indent, false, false, "Following", "%v", r.Following)
	indentedPrint(w, indent, false, false, "Followed-by", "%v", r.FollowedBy)
	indentedPrint(w, indent, false, false, "Blocking", "%v", r.Blocking)
	indentedPrint(w, indent, false, false, "Muting", "%v", r.Muting)
	indentedPrint(w, indent, false, false, "Requested", "%v", r.Requested)
	return nil
}

func (p *PlainPrinter) plainPrintReport(r *madon.Report, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Report ID", "%d", r.ID)
	indentedPrint(w, indent, false, false, "Action taken", "%s", r.ActionTaken)
	return nil
}

func (p *PlainPrinter) plainPrintResults(r *madon.Results, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Results", "%d account(s), %d status(es), %d hashtag(s)",
		len(r.Accounts), len(r.Statuses), len(r.Hashtags))
	if len(r.Accounts) > 0 {
		indentedPrint(w, indent, false, false, "Accounts", "")
		p.PrintObj(r.Accounts, w, indent+p.Indent)
	}
	if len(r.Statuses) > 0 {
		indentedPrint(w, indent, false, false, "Statuses", "")
		p.PrintObj(r.Statuses, w, indent+p.Indent)
	}
	if len(r.Hashtags) > 0 {
		indentedPrint(w, indent, false, false, "Hashtags", "")
		for _, tag := range r.Hashtags {
			indentedPrint(w, indent+p.Indent, true, false, "Tag", "%s", tag)
		}
	}
	return nil
}

func (p *PlainPrinter) plainPrintStatus(s *madon.Status, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Status ID", "%d", s.ID)
	indentedPrint(w, indent, false, false, "From", "%s", s.Account.Acct)
	indentedPrint(w, indent, false, false, "Timestamp", "%v", s.CreatedAt.Local())

	if s.Reblog != nil {
		indentedPrint(w, indent, false, false, "Reblogged from", "%s", s.Reblog.Account.Username)
		return p.plainPrintStatus(s.Reblog, w, indent+p.Indent)
	}

	indentedPrint(w, indent, false, false, "Contents", "%s", html2string(s.Content))
	if s.InReplyToID > 0 {
		indentedPrint(w, indent, false, false, "In-Reply-To", "%d", s.InReplyToID)
	}
	if s.Reblogged {
		indentedPrint(w, indent, false, false, "Reblogged", "%v", s.Reblogged)
	}
	indentedPrint(w, indent, false, false, "URL", "%s", s.URL)
	return nil
}

func (p *PlainPrinter) plainPrintUserToken(s *madon.UserToken, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "User token", "%s", s.AccessToken)
	indentedPrint(w, indent, false, true, "Type", "%s", s.TokenType)
	if s.CreatedAt != 0 {
		indentedPrint(w, indent, false, true, "Timestamp", "%v", time.Unix(int64(s.CreatedAt), 0))
	}
	indentedPrint(w, indent, false, true, "Scope", "%s", s.Scope)
	return nil
}

func (p *PlainPrinter) plainPrintInstanceStatistics(is *gomif.InstanceStatus, w io.Writer, indent string) error {
	if is == nil {
		return nil
	}
	indentedPrint(w, indent, true, false, "Instance", "%s", is.InstanceName)
	indentedPrint(w, indent, false, false, "Users", "%d", is.Users)
	indentedPrint(w, indent, false, false, "Statuses", "%d", is.Statuses)
	indentedPrint(w, indent, false, false, "Open Registrations", "%v", is.OpenRegistrations)
	indentedPrint(w, indent, false, false, "Up", "%v", is.Up)
	indentedPrint(w, indent, false, false, "Date", "%s", time.Unix(is.Date, 0))
	return nil
}
