package mail_service

import (
	"io"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"strings"
	"testing"
)

func TestHTMLToText(t *testing.T) {
	in := `<html><head><style>p{color:red}</style><title>x</title></head><body>
<!-- comment --><h1>Xin ch&agrave;o</h1><p>Hello <b>Nguyễn</b>,<br>welcome!</p>
<ul><li>One</li><li>Two</li></ul>
<p><a href="https://example.com/a?x=1&amp;y=2">Read more</a> · <a href="#top">top</a></p>
<img src="https://track/pixel" style="display:none" alt="" /></body></html>`
	got := HTMLToText(in)

	for _, want := range []string{
		"Xin chào",
		"Hello Nguyễn,\nwelcome!",
		"- One",
		"- Two",
		"Read more (https://example.com/a?x=1&y=2)",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
	for _, unwanted := range []string{"color:red", "comment", "<", "track/pixel", "#top"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("unexpected %q in:\n%s", unwanted, got)
		}
	}
}

func TestBuildAlternativeBody(t *testing.T) {
	html := "<p>Chào bạn <a href=\"https://x.com\">link</a></p>"
	ct, body := buildAlternativeBody(html)

	mediaType, params, err := mime.ParseMediaType(ct)
	if err != nil || mediaType != "multipart/alternative" {
		t.Fatalf("bad content type %q: %v", ct, err)
	}

	r := multipart.NewReader(strings.NewReader(body), params["boundary"])
	var types, contents []string
	for {
		p, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		b, _ := io.ReadAll(quotedprintable.NewReader(p))
		types = append(types, p.Header.Get("Content-Type"))
		contents = append(contents, string(b))
	}

	if len(types) != 2 || !strings.HasPrefix(types[0], "text/plain") || !strings.HasPrefix(types[1], "text/html") {
		t.Fatalf("unexpected parts %v", types)
	}
	if contents[0] != "Chào bạn link (https://x.com)" {
		t.Errorf("plain part = %q", contents[0])
	}
	if contents[1] != html {
		t.Errorf("html part = %q", contents[1])
	}
}

func TestIsHTMLContentType(t *testing.T) {
	if !isHTMLContentType("text/html; charset=utf-8") || isHTMLContentType("text/plain") {
		t.Error("isHTMLContentType misclassified")
	}
}
