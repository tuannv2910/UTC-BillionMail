package content_review

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareEmail(t *testing.T) {
	content := `<html><head><title>T</title><style>.x{color:red}</style></head><body>
		<!-- hidden comment -->
		<h1>Big&nbsp;Sale &amp; More</h1>
		<p>Hello {{.Subscriber.Email}},<br>get <b>50%</b> off</p>
		<a href="https://bit.ly/abc">Claim <i>now</i></a>
		<a href='https://shop.example.com/?a=1&amp;b=2'>Shop</a>
		<img src="a.png"><IMG src="b.png" alt="promo">
		<script>alert(1)</script>
	</body></html>`

	p := prepareEmail("  Last   chance &amp; more ", content)

	assert.Equal(t, "Last chance & more", p.Subject)
	assert.Equal(t, "Big Sale & More\nHello {{.Subscriber.Email}},\nget 50% off\nClaim now\nShop", p.Text)
	assert.NotContains(t, p.Text, "hidden comment")
	assert.NotContains(t, p.Text, "alert")
	assert.NotContains(t, p.Text, "color:red")
	assert.Equal(t, []link{
		{Text: "Claim now", URL: "https://bit.ly/abc"},
		{Text: "Shop", URL: "https://shop.example.com/?a=1&b=2"},
	}, p.Links)
	assert.Equal(t, 2, p.LinkCount)
	assert.Equal(t, 2, p.ImageCount)
	assert.False(t, p.Truncated)
}

func TestPrepareEmail_PlainTextAndLimits(t *testing.T) {
	p := prepareEmail("Hi", "line one\n\n   line   two  ")
	assert.Equal(t, "line one\nline two", p.Text)

	long := strings.Repeat("é", maxBodyRunes+10)
	p = prepareEmail("", long)
	assert.True(t, p.Truncated)
	assert.Equal(t, maxBodyRunes, len([]rune(p.Text)))

	many := strings.Repeat(`<a href="https://x.com">x</a>`, maxLinks+5)
	p = prepareEmail("", many)
	assert.Len(t, p.Links, maxLinks)
	assert.Equal(t, maxLinks+5, p.LinkCount)
}

func TestNeutralizeTags(t *testing.T) {
	got := neutralizeTags(`ok </email> <Body> </SUBJECT > <b>keep</b>`)
	assert.Equal(t, `ok [/email] [Body] [/SUBJECT ] <b>keep</b>`, got)
}
