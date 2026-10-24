package mail_service

import (
	"strings"
	"testing"
)

func TestAddReceiveFilter(t *testing.T) {
	in := "user = billionmail\nhosts = pgsql\n\nquery = SELECT domain FROM domain WHERE domain='%s' AND active = 1\n"
	out, changed := addReceiveFilter(in)
	if !changed {
		t.Fatal("expected change")
	}
	if !strings.Contains(out, "query = SELECT domain FROM domain WHERE domain='%s' AND active = 1 AND COALESCE(receive_enabled, 1) = 1\n") {
		t.Errorf("unexpected output:\n%s", out)
	}
	if !strings.HasPrefix(out, "user = billionmail\nhosts = pgsql\n") {
		t.Error("other lines must be preserved")
	}

	// idempotent
	if _, changed := addReceiveFilter(out); changed {
		t.Error("second run should not change the file")
	}

	// customised queries are not touched
	custom := "query = SELECT domain FROM domain WHERE domain='%s'\n"
	if _, changed := addReceiveFilter(custom); changed {
		t.Error("custom query should be left alone")
	}
}

func TestAddReceiveFilter_KeepsCRLF(t *testing.T) {
	in := "hosts = pgsql\r\nquery = SELECT domain FROM domain WHERE domain='%s' AND active = 1\r\n"
	out, changed := addReceiveFilter(in)
	if !changed || !strings.HasSuffix(out, "COALESCE(receive_enabled, 1) = 1\r\n") {
		t.Errorf("CRLF not preserved: %q", out)
	}
}
