package cmd

import "testing"

func TestIsSafePathPublicRoute(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"/roundcube", true},
		{"/roundcube/", true},
		{"/roundcube/skins/elastic/styles.css", true},
		{"/.well-known/acme-challenge/abc123", true},
		{"/pmta/eyJhYmMiOiJ4In0.sig", true},
		{"/landing/video", true},
		{"/inbox-placement/oauth/callback", true},

		{"/", false},
		{"/overview", false},
		{"/api/domains/list", false},
		{"/roundcubeadmin", false},
		{"/pmta", false},
		{"/.well-known/other", false},
		{"/rspamd/", false},
	}
	for _, tt := range tests {
		if got := isSafePathPublicRoute(tt.path); got != tt.want {
			t.Errorf("isSafePathPublicRoute(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
