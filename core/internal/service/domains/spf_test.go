package domains

import "testing"

func TestFormatSPFValue(t *testing.T) {
	if got := FormatSPFValue("147.93.139.179"); got != "v=spf1 +a +mx +ip4:147.93.139.179 -all" {
		t.Errorf("ipv4: got %q", got)
	}
	if got := FormatSPFValue("2001:db8::1"); got != "v=spf1 +a +mx +ip6:2001:db8::1 -all" {
		t.Errorf("ipv6: got %q", got)
	}
}

func TestGetSPFRecordWithIP_UsesDedicatedIP(t *testing.T) {
	rec, err := GetSPFRecordWithIP("example.com", false, "147.93.139.179")
	if err != nil {
		t.Fatal(err)
	}
	if rec.Value != "v=spf1 +a +mx +ip4:147.93.139.179 -all" || rec.Type != "TXT" || rec.Host != "@" {
		t.Errorf("unexpected record %+v", rec)
	}
}
