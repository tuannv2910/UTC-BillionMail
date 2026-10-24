package mail_service

import (
	"context"
	"regexp"
	"strings"

	"billionmail-core/internal/service/public"

	"github.com/gogf/gf/v2/frame/g"
)

// Send-only domains (#303): a domain whose mail is hosted elsewhere (its MX
// points to another server) must not be treated as local by Postfix, otherwise
// mail from local users to that domain is rejected with "User unknown in
// virtual mailbox table" instead of being delivered via MX. Excluding the
// domain from virtual_mailbox_domains makes Postfix route it like any other
// external domain, while sending from it keeps working.

const virtualDomainsMapsFile = "../conf/postfix/sql/pgsql_virtual_domains_maps.cf"

const receiveFilterSQL = "COALESCE(receive_enabled, 1) = 1"

var virtualDomainsQueryLine = regexp.MustCompile(`(?m)^(\s*query\s*=\s*SELECT\s+domain\s+FROM\s+domain\s+WHERE\s+domain\s*=\s*'%s'\s+AND\s+active\s*=\s*1)[ \t]*(\r?)$`)

// addReceiveFilter returns the cf content with the receive filter added to the
// virtual domain query, and whether it changed. Unknown/customised queries are
// left untouched.
func addReceiveFilter(content string) (string, bool) {
	if strings.Contains(content, "receive_enabled") {
		return content, false
	}
	if !virtualDomainsQueryLine.MatchString(content) {
		return content, false
	}
	return virtualDomainsQueryLine.ReplaceAllString(content, "${1} AND "+receiveFilterSQL+"${2}"), true
}

// EnsureVirtualDomainsReceiveFilter patches the deployed Postfix lookup so that
// domains with receive_enabled = 0 are not local. It runs after the database
// migration has added the column; the caller reloads Postfix.
func EnsureVirtualDomainsReceiveFilter(ctx context.Context) bool {
	path := public.AbsPath(virtualDomainsMapsFile)
	if !public.FileExists(path) {
		return false
	}
	content, err := public.ReadFile(path)
	if err != nil {
		g.Log().Warningf(ctx, "Failed to read %s: %v", path, err)
		return false
	}
	updated, changed := addReceiveFilter(content)
	if !changed {
		return false
	}
	if _, err = public.WriteFile(path, updated); err != nil {
		g.Log().Warningf(ctx, "Failed to update %s: %v", path, err)
		return false
	}
	g.Log().Infof(ctx, "Added receive_enabled filter to %s", path)
	return true
}
