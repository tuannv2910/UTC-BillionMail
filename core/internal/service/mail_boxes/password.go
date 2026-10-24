package mail_boxes

import (
	"crypto/rand"
	"math/big"
	"regexp"
	"strings"
)

const (
	passwordLower  = "abcdefghijklmnopqrstuvwxyz"
	passwordUpper  = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	passwordDigits = "0123456789"
)

var mailboxLocalPartPattern = regexp.MustCompile(`^[A-Za-z0-9._+\-]+$`)

func randomIndex(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		return 0
	}
	return int(v.Int64())
}

// GenerateCompliantPassword returns a random password that always contains at
// least one lowercase letter, one uppercase letter and one digit, matching the
// rule enforced by the UI (#350). Minimum length is 8.
func GenerateCompliantPassword(length int) string {
	if length < 8 {
		length = 8
	}
	all := passwordLower + passwordUpper + passwordDigits
	pwd := make([]byte, length)
	pwd[0] = passwordLower[randomIndex(len(passwordLower))]
	pwd[1] = passwordUpper[randomIndex(len(passwordUpper))]
	pwd[2] = passwordDigits[randomIndex(len(passwordDigits))]
	for i := 3; i < length; i++ {
		pwd[i] = all[randomIndex(len(all))]
	}
	// Fisher–Yates shuffle so the required classes are not always first
	for i := length - 1; i > 0; i-- {
		j := randomIndex(i + 1)
		pwd[i], pwd[j] = pwd[j], pwd[i]
	}
	return string(pwd)
}

// IsCompliantPassword reports whether pwd satisfies the mailbox password rule.
func IsCompliantPassword(pwd string) bool {
	return len(pwd) >= 8 &&
		strings.ContainsAny(pwd, passwordLower) &&
		strings.ContainsAny(pwd, passwordUpper) &&
		strings.ContainsAny(pwd, passwordDigits)
}

// ValidateMailboxAddress checks that a mailbox identity is a usable address so
// imports cannot create rows with an empty or malformed username (#378).
func ValidateMailboxAddress(username, localPart, domain string) bool {
	username = strings.TrimSpace(username)
	localPart = strings.TrimSpace(localPart)
	domain = strings.TrimSpace(domain)
	if username == "" || localPart == "" || domain == "" {
		return false
	}
	if !strings.EqualFold(username, localPart+"@"+domain) {
		return false
	}
	if !mailboxLocalPartPattern.MatchString(localPart) {
		return false
	}
	return strings.Contains(domain, ".") && !strings.ContainsAny(domain, " @/")
}
