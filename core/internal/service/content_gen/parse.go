package content_gen

import (
	"regexp"
	"strings"
)

var (
	// Labels may come wrapped in markdown, e.g. "**SUBJECT:** Hello" or "## SUBJECT: Hello".
	reSubject   = regexp.MustCompile(`(?im)^[\s*#]*SUBJECT[\s*]*:[\s*]*(.+?)\s*$`)
	rePreheader = regexp.MustCompile(`(?im)^[\s*#]*PREHEADER[\s*]*:[\s*]*(.+?)\s*$`)
	reFenced    = regexp.MustCompile("(?s)```(?:html)?\\s*\\n(.*?)\\n?```")
	reDocument  = regexp.MustCompile(`(?is)(<!DOCTYPE html.*</html>|<html\b.*</html>)`)

	reScript       = regexp.MustCompile(`(?is)<script\b.*?</script\s*>|<script\b[^>]*/?>`)
	reTag          = regexp.MustCompile(`(?s)<[a-zA-Z][^>]*>`)
	reEventHandler = regexp.MustCompile(`(?i)\s+on[a-z]+\s*=\s*("[^"]*"|'[^']*'|[^\s>]+)`)
	reJSURL        = regexp.MustCompile(`(?i)(href|src)\s*=\s*(["'])\s*javascript:[^"']*(["'])`)
)

// parseReply extracts the subject, preheader and HTML document from the model's reply.
func parseReply(reply string) (subject, preheader, html string, ok bool) {
	if m := reSubject.FindStringSubmatch(reply); m != nil {
		subject = strings.Trim(m[1], `"'`)
	}
	if m := rePreheader.FindStringSubmatch(reply); m != nil {
		preheader = strings.Trim(m[1], `"'`)
	}

	if m := reFenced.FindStringSubmatch(reply); m != nil && strings.Contains(strings.ToLower(m[1]), "</html>") {
		html = strings.TrimSpace(m[1])
	} else if m := reDocument.FindString(reply); m != "" {
		html = strings.TrimSpace(m)
	}
	return subject, preheader, html, html != ""
}

// sanitizeHTML removes scripts, inline event handlers and javascript: URLs. Email clients strip
// them anyway, but the editor preview renders the HTML in the admin's browser.
func sanitizeHTML(html string) string {
	html = reScript.ReplaceAllString(html, "")
	return reTag.ReplaceAllStringFunc(html, func(tag string) string {
		tag = reEventHandler.ReplaceAllString(tag, "")
		return reJSURL.ReplaceAllString(tag, `$1=$2#$3`)
	})
}
