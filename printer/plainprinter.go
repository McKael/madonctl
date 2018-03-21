// Copyright © 2017-2018 Mikael Berthe <mikael@lilotux.net>
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

	"github.com/McKael/madon"
	"github.com/McKael/madonctl/printer/html2text"
)

// PlainPrinter is the default "plain text" printer
type PlainPrinter struct {
	Indent      string
	NoSubtitles bool
}

// NewPrinterPlain returns a plaintext ResourcePrinter
// For PlainPrinter, the option parameter contains the indent prefix.
func NewPrinterPlain(options Options) (*PlainPrinter, error) {
	indentInc := "  "
	if i, ok := options["indent"]; ok {
		indentInc = i
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
		[]madon.Emoji, []madon.Instance, []madon.InstancePeer,
		[]madon.List, []madon.Mention, []madon.Notification,
		[]madon.Relationship, []madon.Report, []madon.Results,
		[]madon.Status, []madon.StreamEvent, []madon.Tag,
		[]madon.DomainName:
		return p.plainForeach(o, w, initialIndent)
	case *madon.DomainName:
		return p.plainPrintDomainName(o, w, initialIndent)
	case madon.DomainName:
		return p.plainPrintDomainName(&o, w, initialIndent)
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
	case *madon.Emoji:
		return p.plainPrintEmoji(o, w, initialIndent)
	case madon.Emoji:
		return p.plainPrintEmoji(&o, w, initialIndent)
	case *madon.Instance:
		return p.plainPrintInstance(o, w, initialIndent)
	case madon.Instance:
		return p.plainPrintInstance(&o, w, initialIndent)
	case *madon.InstancePeer:
		return p.plainPrintInstancePeer(o, w, initialIndent)
	case madon.InstancePeer:
		return p.plainPrintInstancePeer(&o, w, initialIndent)
	case *madon.List:
		return p.plainPrintList(o, w, initialIndent)
	case madon.List:
		return p.plainPrintList(&o, w, initialIndent)
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
	t, err := html2text.Textify(h)
	if err == nil {
		return t
	}
	return h // Failed: return initial string
}

// unix2time convert a UNIX timestamp to a time.Time
func unix2time(ts interface{}) (time.Time, error) {
	switch t := ts.(type) {
	case int64:
		return time.Unix(t, 0), nil
	case int:
		return time.Unix(int64(t), 0), nil
	case float64:
		return time.Unix(int64(t), 0), nil
	}
	return time.Time{}, fmt.Errorf("invalid timestamp type")
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

func (p *PlainPrinter) plainPrintDomainName(d *madon.DomainName, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Domain Name", "%s", string(*d))
	return nil
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
	if a.Moved != nil {
		m := a.Moved
		indentedPrint(w, indent+p.Indent, true, false, "Moved to account ID", "%d (%s)", m.ID, m.Username)
		indentedPrint(w, indent+p.Indent, false, false, "New user ID", "%s", m.Acct)
		indentedPrint(w, indent+p.Indent, false, false, "New display name", "%s", m.DisplayName)
	}
	return nil
}

func (p *PlainPrinter) plainPrintAttachment(a *madon.Attachment, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Attachment ID", "%d", a.ID)
	indentedPrint(w, indent, false, false, "Type", "%s", a.Type)
	indentedPrint(w, indent, false, false, "Local URL", "%s", a.URL)
	if a.RemoteURL != nil {
		indentedPrint(w, indent, false, true, "Remote URL", "%s", *a.RemoteURL)
	}
	indentedPrint(w, indent, false, true, "Preview URL", "%s", a.PreviewURL)
	if a.TextURL != nil {
		indentedPrint(w, indent, false, true, "Text URL", "%s", *a.TextURL)
	}
	if a.Description != nil {
		indentedPrint(w, indent, false, true, "Description", "%s", *a.Description)
	}
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
	indentedPrint(w, indent, true, false, "Context", "%d relative(s)", len(c.Ancestors)+len(c.Descendants))
	if len(c.Ancestors) > 0 {
		indentedPrint(w, indent, false, false, "Ancestors", "")
		p.PrintObj(c.Ancestors, w, indent+p.Indent)
	}
	if len(c.Descendants) > 0 {
		indentedPrint(w, indent, false, false, "Descendants", "")
		p.PrintObj(c.Descendants, w, indent+p.Indent)
	}
	return nil
}

func (p *PlainPrinter) plainPrintEmoji(e *madon.Emoji, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Emoji shortcode", "%s", e.ShortCode)
	indentedPrint(w, indent, false, false, "URL", "%s", e.URL)
	return nil
}

func (p *PlainPrinter) plainPrintInstance(i *madon.Instance, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Instance title", "%s", i.Title)
	indentedPrint(w, indent, false, true, "Description", "%s", html2string(i.Description))
	indentedPrint(w, indent, false, true, "URL", "%s", i.URI)
	indentedPrint(w, indent, false, true, "Email", "%s", i.Email)
	indentedPrint(w, indent, false, true, "Version", "%s", i.Version)
	if i.ContactAccount != nil {
		c := i.ContactAccount
		indentedPrint(w, indent+p.Indent, true, false, "Contact account ID", "%d (%s)", c.ID, c.Username)
		indentedPrint(w, indent+p.Indent, false, false, "Contact user ID", "%s", c.Acct)
		indentedPrint(w, indent+p.Indent, false, false, "Contact display name", "%s", c.DisplayName)
	}
	return nil
}

func (p *PlainPrinter) plainPrintInstancePeer(i *madon.InstancePeer, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "Peer", "%s", *i)
	return nil
}

func (p *PlainPrinter) plainPrintList(l *madon.List, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "List ID", "%d", l.ID)
	indentedPrint(w, indent, false, false, "Title", "%s", l.Title)
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
	indentedPrint(w, indent, true, false, "Account ID", "%d", r.ID)
	indentedPrint(w, indent, false, false, "Following", "%v", r.Following)
	//indentedPrint(w, indent, false, false, "Showing reblogs", "%v", r.ShowingReblogs)
	indentedPrint(w, indent, false, false, "Followed-by", "%v", r.FollowedBy)
	indentedPrint(w, indent, false, false, "Blocking", "%v", r.Blocking)
	indentedPrint(w, indent, false, false, "Muting", "%v", r.Muting)
	indentedPrint(w, indent, false, false, "Muting notifications", "%v", r.MutingNotifications)
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
	if s.Account != nil {
		author := s.Account.Acct
		if s.Account.DisplayName != "" {
			author += " (" + s.Account.DisplayName + ")"
		}
		indentedPrint(w, indent, false, false, "From", "%s", author)
	}
	if s.Pinned {
		indentedPrint(w, indent, false, false, "Pinned", "%v", s.Pinned)
	}
	if s.Visibility == "private" {
		indentedPrint(w, indent, false, false, "Private", "true")
	}
	indentedPrint(w, indent, false, false, "Timestamp", "%v", s.CreatedAt.Local())

	if s.Reblog != nil {
		if s.Reblog.Account != nil {
			indentedPrint(w, indent, false, false, "Reblogged from", "%s", s.Reblog.Account.Username)
		}
		return p.plainPrintStatus(s.Reblog, w, indent+p.Indent)
	}

	if s.Sensitive {
		indentedPrint(w, indent, false, false, "Sensitive (NSFW)", "%v", s.Sensitive)
	}

	indentedPrint(w, indent, false, false, "Contents", "%s", html2string(s.Content))
	if s.InReplyToID != nil && *s.InReplyToID > 0 {
		indentedPrint(w, indent, false, false, "In-Reply-To", "%d", *s.InReplyToID)
	}
	if s.Reblogged {
		indentedPrint(w, indent, false, false, "Reblogged", "%v", s.Reblogged)
	}
	indentedPrint(w, indent, false, false, "URL", "%s", s.URL)
	// Display minimum details of attachments
	//return p.PrintObj(s.MediaAttachments, w, indent+p.Indent)
	for _, a := range s.MediaAttachments {
		indentedPrint(w, indent+p.Indent, true, false, "Attachment ID", "%d", a.ID)
		if a.TextURL != nil && *a.TextURL != "" {
			indentedPrint(w, indent+p.Indent, true, false, "Text URL", "%s", *a.TextURL)
		} else if a.URL != "" {
			indentedPrint(w, indent+p.Indent, false, false, "URL", "%s", a.URL)
		} else if a.RemoteURL != nil {
			indentedPrint(w, indent+p.Indent, false, false, "Remote URL", "%s", *a.RemoteURL)
		}
		if a.Description != nil && *a.Description != "" {
			indentedPrint(w, indent+p.Indent, false, true, "Description", "%s", a.Description)
		}
	}
	return nil
}

func (p *PlainPrinter) plainPrintUserToken(s *madon.UserToken, w io.Writer, indent string) error {
	indentedPrint(w, indent, true, false, "User token", "%s", s.AccessToken)
	indentedPrint(w, indent, false, true, "Type", "%s", s.TokenType)
	if s.CreatedAt != 0 {
		indentedPrint(w, indent, false, true, "Timestamp", "%v", time.Unix(s.CreatedAt, 0))
	}
	indentedPrint(w, indent, false, true, "Scope", "%s", s.Scope)
	return nil
}
