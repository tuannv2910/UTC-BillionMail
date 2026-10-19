package blacklist

import (
	"fmt"
	"strings"
)

const (
	maxEmailLength  = 320
	maxLocalLength  = 64
	maxDomainLength = 253
	maxLabelLength  = 63
)

// NormalizeEmail trims and lower-cases an address and returns it with its domain.
// Gmail dots and "+tag" suffixes are kept: stripping them would block different mailboxes.
func NormalizeEmail(raw string) (email, domain string, err error) {
	email = strings.ToLower(strings.TrimSpace(raw))
	if email == "" || len(email) > maxEmailLength {
		return "", "", fmt.Errorf("%w: %q", ErrInvalidEmail, raw)
	}

	at := strings.LastIndexByte(email, '@')
	if at <= 0 || at > maxLocalLength || strings.ContainsAny(email, " \t\r\n<>,;") || strings.Count(email, "@") != 1 {
		return "", "", fmt.Errorf("%w: %q", ErrInvalidEmail, raw)
	}

	domain, err = NormalizeDomain(email[at+1:])
	if err != nil {
		return "", "", fmt.Errorf("%w: %q", ErrInvalidEmail, raw)
	}
	return email[:at+1] + domain, domain, nil
}

// NormalizeDomain trims, lower-cases and strips a leading "@" or trailing "." from a domain.
func NormalizeDomain(raw string) (string, error) {
	domain := strings.ToLower(strings.TrimSpace(raw))
	domain = strings.TrimSuffix(strings.TrimPrefix(domain, "@"), ".")
	if domain == "" || len(domain) > maxDomainLength || !strings.Contains(domain, ".") {
		return "", fmt.Errorf("%w: %q", ErrInvalidDomain, raw)
	}

	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > maxLabelLength || label[0] == '-' || label[len(label)-1] == '-' {
			return "", fmt.Errorf("%w: %q", ErrInvalidDomain, raw)
		}
		for _, c := range label {
			if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '-') {
				return "", fmt.Errorf("%w: %q", ErrInvalidDomain, raw)
			}
		}
	}
	return domain, nil
}

// normalizeValue normalizes raw according to scope and returns the value and its domain.
func normalizeValue(scope Scope, raw string) (value, domain string, err error) {
	switch scope {
	case ScopeEmail:
		return NormalizeEmail(raw)
	case ScopeDomain:
		domain, err = NormalizeDomain(raw)
		return domain, domain, err
	default:
		return "", "", ErrInvalidScope
	}
}
