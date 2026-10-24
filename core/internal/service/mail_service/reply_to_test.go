package mail_service

import "testing"

func TestFormatReplyTo(t *testing.T) {
	cases := map[string]string{
		"":                              "",
		"support@example.com":           "support@example.com",
		" Support <help@example.com>":   `"Support" <help@example.com>`,
		"not-an-email":                  "",
		"a@localhost":                   "",
		"x@example.com\r\nBcc: y@z.com": "",
	}
	for in, want := range cases {
		if got := FormatReplyTo(in); got != want {
			t.Errorf("FormatReplyTo(%q) = %q, want %q", in, got, want)
		}
	}
}
