package mail_service

import (
	"net/mail"
	"strings"
)

// FormatReplyTo validates a Reply-To value ("a@x.com" or "Name <a@x.com>")
// and returns it RFC 5322 encoded, or "" when empty/invalid (#86, #357).
func FormatReplyTo(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	addr, err := mail.ParseAddress(value)
	if err != nil || !strings.Contains(addr.Address[strings.LastIndex(addr.Address, "@")+1:], ".") {
		return ""
	}
	if addr.Name == "" {
		return addr.Address
	}
	return addr.String()
}
