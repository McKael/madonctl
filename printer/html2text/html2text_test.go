package html2text

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTextify(t *testing.T) {
	expected := "body\nbody2"
	r, e := Textify("<html><body><b>body</b><br/>body2</body></html>")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyDiv(t *testing.T) {
	expected := "first\nsecond"
	r, e := Textify("<div>first</div>second")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

/*
func TestTextifyLink(t *testing.T) {
	expected := "somelink (link: someurl)"
	r, e := Textify("<a href=\"someurl\">somelink</a>")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}
*/

func TestTextifyDontDuplicateLink(t *testing.T) {
	expected := "www.awesome.com"
	r, e := Textify("<a href=\"www.awesome.com\">www.awesome.com</a>")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifySpaces(t *testing.T) {
	expected := "hello"
	r, e := Textify("<div> hello  </div>")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

/*  I don't think we want that for Mastodon...
func TestTextifySpacesMultiple(t *testing.T) {
	expected := "hello goodbye"
	r, e := Textify("<span> hello  </span><span>   goodbye   </span>")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}
*/

func TestTextifyNonBreakingSpace(t *testing.T) {
	expected := "a   a"
	r, e := Textify("a &nbsp; a")
	assert.Equal(t, expected, r)
	assert.Nil(t, e)
}

func TestTextifyLimitedNewLines(t *testing.T) {
	expected := "abc\nxyz"
	r, e := Textify("abc <br/> <br/> <br/> <br/>xyz")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyTable(t *testing.T) {
	expected := `Join by phone
1-877-668-4490 Call-in toll-free number (US/Canada)
1-408-792-6300 Call-in toll number (US/Canada)
Access code: 111 111 111
https://akqa.webex.com/akqa/globalcallin.php?serviceType=MC&ED=299778282&tollFree=1 | http://www.webex.com/pdf/tollfree_restrictions.pdf`

	test := `<table width="747" style="width:448.2pt;"> <col width="747" style="width:448.2pt;"> <tbody> <tr> <td><font face="Arial" color="#666666"><b>Join by phone</b></font></td> </tr> <tr> <td><font face="Arial" size="3" color="#666666"><span style="font-size:11.5pt;"><b>1-877-668-4490</b> Call-in toll-free number (US/Canada)</span></font></td> </tr> <tr> <td><font face="Arial" size="3" color="#666666"><span style="font-size:11.5pt;"><b>1-408-792-6300</b> Call-in toll number (US/Canada)</span></font></td> </tr> <tr> <td><font face="Arial" size="3" color="#666666"><span style="font-size:11.5pt;">Access code: 111 111 111</span></font></td> </tr> <tr> <td><a href="https://akqa.webex.com/akqa/globalcallin.php?serviceType=MC&amp;ED=299778282&amp;tollFree=1"><font face="Arial" size="2" color="#00AFF9"><span style="font-size:10pt;"><u>Global call-in numbers</u></span></font></a><font face="Arial" size="3" color="#666666"><span style="font-size:11.5pt;"> &nbsp;|&nbsp; </span></font><a href="http://www.webex.com/pdf/tollfree_restrictions.pdf"><font face="Arial" size="2" color="#00AFF9"><span style="font-size:10pt;"><u>Toll-free calling restrictions</u></span></font></a></td> </tr> </tbody> </table>`

	r, e := Textify(test)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyComment(t *testing.T) {
	expected := "this should appear"
	r, e := Textify("<!-- this should not appear -->this should appear")
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyCommentInHead(t *testing.T) {
	expected := "qwerty"

	body := `<html> <head> <meta http-equiv="Content-Type" content="text/html; charset=utf-8"> <meta name="Generator" content="Microsoft Exchange Server"> <!-- converted from rtf --><style><!-- .EmailQuote { margin-left: 1pt; padding-left: 4pt; border-left: #800000 2px solid; } --></style> </head> <body>qwerty</body> </html>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyLists(t *testing.T) {
	expected := "a\nb\n1\n2"

	body := `<ol><li>a</li><li>b</li></ol><ul><li>1</li><li>2</li></ul>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonSample1(t *testing.T) {
	expected := "@magi hello \\U0001F607 @TEST"

	body := `<p><span class=\"h-card\"><a href=\"https://example.com/@magi\">@<span>magi</span></a></span> hello \U0001F607 <span class=\"h-card\"><a href=\"https://example.com/@TEST\">@<span>TEST</span></a></span></p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonSample2(t *testing.T) {
	expected := "@cadey It looks good at first glance\n\"case <-stop\"  Actually you don't listen to stop channel, you close it if you want to stop the listener."

	body := `<p><span class="h-card"><a href="https://www.example.com/@cadey" class="u-url mention">@<span>cadey</span></a></span> It looks good at first glance</p><p>&quot;case &lt;-stop&quot;  Actually you don&apos;t listen to stop channel, you close it if you want to stop the listener.</p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonSample3(t *testing.T) {
	expected := "From timeline: Materials research creates potential for improved computer chips and transistors #phys #physics ..."

	body := `From timeline: Materials research creates potential for improved computer chips and transistors #<span class="tag"><a href="https://social.oalm.gub.uy/tag/phys">phys</a></span> #<span class="tag"><a href="https://social.oalm.gub.uy/tag/physics">physics</a></span><p>...</p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonSample4(t *testing.T) {
	expected := "Vous reprendrez bien un peu de #Tolkein ?\n#Arte +7 propose un ensemble de 6 vidéos en plus du documentaire:\nhttp://www.arte.tv/fr/videos/RC-014610/tolkien/"

	body := `<p>Vous reprendrez bien un peu de <a href="https://framapiaf.org/tags/tolkein">#<span>Tolkein</span></a> ?<br><a href="https://framapiaf.org/tags/arte">#<span>Arte</span></a>+7 propose un ensemble de 6 vidéos en plus du documentaire:</p><p><a href="http://www.arte.tv/fr/videos/RC-014610/tolkien/"><span class="invisible">http://www.</span><span class="ellipsis">arte.tv/fr/videos/RC-014610/to</span><span class="invisible">lkien/</span></a></p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonMention(t *testing.T) {
	expected := "La tête à @Toto \\o/"

	body := `<p>La tête à <span class="h-card"><a href="https://example.com/@Toto">@<span>Toto</span></a></span> \o/</p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonMentionAndTag(t *testing.T) {
	expected := "@ACh Mais heu ! Moi aussi je fais du #TootRadio de gens morts il y a 5 siècles. Gesulado, Charpentier, Mireille Mathieu..."

	body := `<p><span class="h-card"><a href="https://mamot.fr/@ACh">@<span>ACh</span></a></span> Mais heu ! Moi aussi je fais du <a href="https://example.com/tags/tootradio">#<span>TootRadio</span></a> de gens morts il y a 5 siècles. Gesulado, Charpentier, Mireille Mathieu...</p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonLinkSpacing(t *testing.T) {
	expected := "\"Twitter\" https://twitter.com/holly/status/123456789012345678"

	body := `<p>"Twitter" <a href="https://twitter.com/holly/status/123456789012345678"><span class="invisible">https://</span><span class="ellipsis">twitter.com/holly/status/86266</span><span class="invisible">1234567890123</span></a></p>`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}

func TestTextifyMastodonMentionGNUSocial(t *testing.T) {
	expected := "@username Hello."

	body := `@<a href="https://example.com/user/12345">username</a> Hello.`

	r, e := Textify(body)
	assert.Nil(t, e)
	assert.Equal(t, expected, r)
}
