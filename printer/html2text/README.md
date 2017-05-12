# html2text

This is a copy of github.com/sgoertzen/html2text, heavily customized for
Mastodon's HTML messages.

html2text is an HTML to text converter written in Go.
This library will strip the html tags from the source and perform clean up on the text.
This includes things like adding new lines correctly and appending on urls from links.

For Mastodon tags, URLs are not displayed.
