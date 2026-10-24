package cmd

import "strings"

// safePathPublicPrefixes are routes that must stay reachable without the admin
// SafePath session: they are opened by mailbox users, mail recipients or
// external services that never visit the admin entrance.
//   - /roundcube              webmail login for mailbox users (#364, #381, #387)
//   - /.well-known/acme-challenge/  Let's Encrypt HTTP-01 validation (#360, #364)
//   - /pmta/                  open/click tracking links in sent emails (#338, #361)
//   - /landing/video          public video landing page linked from emails
//   - /inbox-placement/oauth/callback  Google OAuth redirect
var safePathPublicPrefixes = []string{
	"/roundcube/",
	"/.well-known/acme-challenge/",
	"/pmta/",
}

var safePathPublicExact = map[string]struct{}{
	"/roundcube":                      {},
	"/landing/video":                  {},
	"/inbox-placement/oauth/callback": {},
}

// isSafePathPublicRoute reports whether path bypasses the SafePath check.
// Admin routes (including /api/*) are never matched here.
func isSafePathPublicRoute(path string) bool {
	if _, ok := safePathPublicExact[path]; ok {
		return true
	}
	for _, prefix := range safePathPublicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
}
