package mail_boxes

import "testing"

func TestGenerateCompliantPassword(t *testing.T) {
	for i := 0; i < 2000; i++ {
		p := GenerateCompliantPassword(8)
		if len(p) != 8 {
			t.Fatalf("unexpected length %d", len(p))
		}
		if !IsCompliantPassword(p) {
			t.Fatalf("password %q does not meet the policy", p)
		}
	}
	if got := GenerateCompliantPassword(4); len(got) != 8 {
		t.Errorf("short lengths should be raised to 8, got %d", len(got))
	}
}

func TestIsCompliantPassword(t *testing.T) {
	cases := map[string]bool{
		"Abcdefg1": true,
		"abcdefg1": false,
		"ABCDEFG1": false,
		"Abcdefgh": false,
		"Ab1":      false,
		"123456":   false,
	}
	for pwd, want := range cases {
		if got := IsCompliantPassword(pwd); got != want {
			t.Errorf("IsCompliantPassword(%q) = %v, want %v", pwd, got, want)
		}
	}
}

func TestValidateMailboxAddress(t *testing.T) {
	tests := []struct {
		user, local, domain string
		want                bool
	}{
		{"john@example.com", "john", "example.com", true},
		{"John.Doe@Example.com", "john.doe", "example.com", true},
		{"", "", "", false},
		{"@example.com", "", "example.com", false},
		{"john@", "john", "", false},
		{"john@example.com", "jane", "example.com", false},
		{"jo hn@example.com", "jo hn", "example.com", false},
		{"john@localhost", "john", "localhost", false},
	}
	for _, tt := range tests {
		if got := ValidateMailboxAddress(tt.user, tt.local, tt.domain); got != tt.want {
			t.Errorf("ValidateMailboxAddress(%q,%q,%q) = %v, want %v", tt.user, tt.local, tt.domain, got, tt.want)
		}
	}
}
