package mail_service

import (
	"net"
	"strings"
	"testing"
	"time"
)

func TestFormatFromHeader(t *testing.T) {
	tests := []struct {
		name, display, email, want string
	}{
		{"display name kept", "Vua Nem", "vuanem@marketing.vuanem.com", `"Vua Nem" <vuanem@marketing.vuanem.com>`},
		{"empty falls back to local part", "", "news@example.com", `"news" <news@example.com>`},
		{"comma is quoted", "Acme, Inc", "a@example.com", `"Acme, Inc" <a@example.com>`},
		{"non-ascii is encoded", "Vua Nệm", "a@example.com", "=?utf-8?q?Vua_N=E1=BB=87m?= <a@example.com>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FormatFromHeader(tt.display, tt.email); got != tt.want {
				t.Errorf("FormatFromHeader() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetRealNameIsUsedInFrom(t *testing.T) {
	m := NewMessage("s", "c")
	m.SetRealName("  Vua Nem ")
	if got := m.Headers["RealName"]; got != "Vua Nem" {
		t.Fatalf("RealName header = %q", got)
	}
}

func TestIsImplicitTLSPort(t *testing.T) {
	for _, p := range []string{"465", "2465", " 465 "} {
		if !IsImplicitTLSPort(p) {
			t.Errorf("%q should be implicit TLS", p)
		}
	}
	for _, p := range []string{"25", "587", "2525", ""} {
		if IsImplicitTLSPort(p) {
			t.Errorf("%q should not be implicit TLS", p)
		}
	}
}

// A server that never sends a greeting (e.g. an implicit-TLS port reached in
// plaintext) must fail fast instead of hanging (#379).
func TestConnectPlainTimesOutWithoutGreeting(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Skip("cannot listen:", err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			defer c.Close() // stay silent
		}
	}()

	old := smtpHandshakeTimeout
	smtpHandshakeTimeout = 300 * time.Millisecond
	defer func() { smtpHandshakeTimeout = old }()

	host, port, _ := net.SplitHostPort(ln.Addr().String())
	e := &EmailSender{Host: host, Port: port, Email: "a@example.com"}

	start := time.Now()
	err = e.Connect()
	if err == nil {
		t.Fatal("expected an error")
	}
	if time.Since(start) > 5*time.Second {
		t.Fatalf("connect took too long: %s", time.Since(start))
	}
	if !strings.Contains(err.Error(), "implicit TLS") {
		t.Errorf("error should hint at implicit TLS, got %v", err)
	}
}
