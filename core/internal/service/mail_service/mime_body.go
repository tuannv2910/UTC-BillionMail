package mail_service

import (
	"bytes"
	"fmt"
	"html"
	"mime/quotedprintable"
	"regexp"
	"strings"

	"github.com/gogf/gf/util/grand"
)

// HTML-only messages are penalised by spam filters (MIME_HTML_ONLY) and are
// unreadable in text-only clients. HTML bodies are therefore sent as
// multipart/alternative with an auto-generated text/plain part (#331).

var (
	reHTMLDropBlocks = regexp.MustCompile(`(?is)<(head|style|script|title)[^>]*>.*?</(head|style|script|title)>`)
	reHTMLComments   = regexp.MustCompile(`(?s)<!--.*?-->`)
	reHTMLHiddenImg  = regexp.MustCompile(`(?i)<img[^>]*>`)
	reHTMLAnchor     = regexp.MustCompile(`(?is)<a\s[^>]*href\s*=\s*["']([^"']+)["'][^>]*>(.*?)</a>`)
	reHTMLBreak      = regexp.MustCompile(`(?i)<br\s*/?>`)
	reHTMLBlockEnd   = regexp.MustCompile(`(?i)</(p|div|h[1-6]|tr|table|ul|ol|blockquote|section|article|header|footer)>`)
	reHTMLListItem   = regexp.MustCompile(`(?i)<li[^>]*>`)
	reHTMLCell       = regexp.MustCompile(`(?i)</t[dh]>`)
	reHTMLTag        = regexp.MustCompile(`(?s)<[^>]+>`)
	reSpaces         = regexp.MustCompile(`[ \t\f\v\x{00a0}]+`)
	reBlankLines     = regexp.MustCompile(`\n{3,}`)
)

// HTMLToText converts an HTML email body into a readable plain-text version.
func HTMLToText(htmlBody string) string {
	s := reHTMLComments.ReplaceAllString(htmlBody, "")
	s = reHTMLDropBlocks.ReplaceAllString(s, "")
	s = reHTMLHiddenImg.ReplaceAllString(s, "")
	s = reHTMLAnchor.ReplaceAllStringFunc(s, func(m string) string {
		parts := reHTMLAnchor.FindStringSubmatch(m)
		href := strings.TrimSpace(html.UnescapeString(parts[1]))
		text := strings.TrimSpace(reHTMLTag.ReplaceAllString(parts[2], ""))
		if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(strings.ToLower(href), "javascript:") {
			return text
		}
		if text == "" || html.UnescapeString(text) == href {
			return href
		}
		return fmt.Sprintf("%s (%s)", text, href)
	})
	s = reHTMLBreak.ReplaceAllString(s, "\n")
	s = reHTMLListItem.ReplaceAllString(s, "\n- ")
	s = reHTMLCell.ReplaceAllString(s, " ")
	s = reHTMLBlockEnd.ReplaceAllString(s, "\n\n")
	s = reHTMLTag.ReplaceAllString(s, "")
	s = html.UnescapeString(s)
	s = strings.ReplaceAll(s, "\r\n", "\n")

	lines := strings.Split(s, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimSpace(reSpaces.ReplaceAllString(line, " "))
	}
	s = strings.Join(lines, "\n")
	s = reBlankLines.ReplaceAllString(s, "\n\n")
	return strings.TrimSpace(s)
}

func quotedPrintable(content string) string {
	buf := new(bytes.Buffer)
	wt := quotedprintable.NewWriter(buf)
	_, _ = wt.Write([]byte(content))
	_ = wt.Close()
	return buf.String()
}

// isHTMLContentType reports whether a Content-Type header is text/html.
func isHTMLContentType(contentType string) bool {
	return strings.HasPrefix(strings.ToLower(strings.TrimSpace(contentType)), "text/html")
}

// buildAlternativeBody returns the top-level Content-Type and the encoded
// multipart/alternative body (text/plain first, HTML last as RFC 2046 requires).
func buildAlternativeBody(htmlBody string) (contentType string, body string) {
	boundary := "bm-alt-" + grand.S(24)
	text := HTMLToText(htmlBody)

	var b strings.Builder
	b.WriteString("This is a multi-part message in MIME format.\r\n\r\n")
	b.WriteString("--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/plain; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	b.WriteString(quotedPrintable(text))
	b.WriteString("\r\n--" + boundary + "\r\n")
	b.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	b.WriteString("Content-Transfer-Encoding: quoted-printable\r\n\r\n")
	b.WriteString(quotedPrintable(htmlBody))
	b.WriteString("\r\n--" + boundary + "--\r\n")

	return fmt.Sprintf("multipart/alternative; boundary=\"%s\"", boundary), b.String()
}
