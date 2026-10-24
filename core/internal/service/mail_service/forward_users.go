package mail_service

import (
	"net/mail"
	"regexp"
	"strings"
)

// forwardSeparators splits forward targets entered one per line (real or
// JSON-escaped "\n"), or separated by commas, semicolons or whitespace.
var forwardSeparators = regexp.MustCompile(`(\\r)?\\n|[\r\n,;\s]+`)

// ProcessForwardUsers splits a forward target list into trimmed, de-duplicated
// addresses, ignoring blank entries (#344).
func ProcessForwardUsers(forwardUsers string) []string {
	parts := forwardSeparators.Split(forwardUsers, -1)
	result := make([]string, 0, len(parts))
	seen := make(map[string]bool, len(parts))
	for _, addr := range parts {
		addr = strings.TrimSpace(addr)
		key := strings.ToLower(addr)
		if addr == "" || seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, addr)
	}
	return result
}

// InvalidForwardUsers returns the entries that are not plain email addresses.
func InvalidForwardUsers(addrs []string) []string {
	var invalid []string
	for _, addr := range addrs {
		parsed, err := mail.ParseAddress(addr)
		if err != nil || parsed.Address != addr || !strings.Contains(addr[strings.LastIndex(addr, "@")+1:], ".") {
			invalid = append(invalid, addr)
		}
	}
	return invalid
}
