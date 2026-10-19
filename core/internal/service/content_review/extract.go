package content_review

import (
	"html"
	"regexp"
	"strings"
	"unicode/utf8"
)

const (
	maxBodyRunes   = 12000 // keeps the prompt within a predictable token budget
	maxLinks       = 50
	maxAnchorRunes = 200
)

var (
	reHidden  = regexp.MustCompile(`(?is)<!--.*?-->|<script\b.*?</script\s*>|<style\b.*?</style\s*>|<head\b.*?</head\s*>`)
	reBreak   = regexp.MustCompile(`(?i)<br\s*/?>|</(?:p|div|li|tr|td|th|h[1-6]|table|blockquote|section|article|header|footer)\s*>`)
	reTag     = regexp.MustCompile(`(?s)<[^>]*>`)
	reLink    = regexp.MustCompile(`(?is)<a\b[^>]*?\bhref\s*=\s*["']([^"']*)["'][^>]*>(.*?)</a\s*>`)
	reImage   = regexp.MustCompile(`(?i)<img\b`)
	reOurTags = regexp.MustCompile(`(?i)</?\s*(?:email|subject|body|links|stats)\b[^>]*>`)
)

// link is a hyperlink found in the email.
type link struct {
	Text string
	URL  string
}

// preparedEmail is the email reduced to what the model needs to read.
type preparedEmail struct {
	Subject    string
	Text       string // visible text, one block per line
	Links      []link
	LinkCount  int // total links, including those beyond maxLinks
	ImageCount int
	Truncated  bool // Text was cut at maxBodyRunes
}

// prepareEmail converts HTML (or plain text) into visible text plus link and image stats.
func prepareEmail(subject, content string) preparedEmail {
	p := preparedEmail{
		Subject:    collapseSpaces(html.UnescapeString(subject)),
		ImageCount: len(reImage.FindAllStringIndex(content, -1)),
	}

	visible := reHidden.ReplaceAllString(content, " ")
	for _, m := range reLink.FindAllStringSubmatch(visible, -1) {
		p.LinkCount++
		if len(p.Links) < maxLinks {
			p.Links = append(p.Links, link{
				Text: truncateRunes(htmlToText(m[2]), maxAnchorRunes),
				URL:  strings.TrimSpace(html.UnescapeString(m[1])),
			})
		}
	}

	p.Text = htmlToText(visible)
	if utf8.RuneCountInString(p.Text) > maxBodyRunes {
		p.Text = truncateRunes(p.Text, maxBodyRunes)
		p.Truncated = true
	}
	return p
}

// htmlToText strips tags and entities, keeping block boundaries as line breaks.
func htmlToText(s string) string {
	s = reBreak.ReplaceAllString(s, "\n")
	s = reTag.ReplaceAllString(s, " ")
	s = html.UnescapeString(s)

	lines := strings.Split(s, "\n")
	out := lines[:0]
	for _, line := range lines {
		if line = collapseSpaces(line); line != "" {
			out = append(out, line)
		}
	}
	return strings.Join(out, "\n")
}

func collapseSpaces(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// neutralizeTags stops email text from closing or opening the prompt's own delimiter tags.
func neutralizeTags(s string) string {
	return reOurTags.ReplaceAllStringFunc(s, func(tag string) string {
		return strings.NewReplacer("<", "[", ">", "]").Replace(tag)
	})
}
