package blacklist

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEmail(t *testing.T) {
	valid := []struct {
		in, email, domain string
	}{
		{"User@Example.COM", "user@example.com", "example.com"},
		{"  john.doe+promo@Gmail.com \n", "john.doe+promo@gmail.com", "gmail.com"},
		{"a@mail.example.co.uk.", "a@mail.example.co.uk", "mail.example.co.uk"},
		{"x@xn--bcher-kva.de", "x@xn--bcher-kva.de", "xn--bcher-kva.de"},
	}
	for _, tc := range valid {
		t.Run(tc.in, func(t *testing.T) {
			email, domain, err := NormalizeEmail(tc.in)
			require.NoError(t, err)
			assert.Equal(t, tc.email, email)
			assert.Equal(t, tc.domain, domain)
		})
	}

	invalid := []string{
		"", "   ", "plainaddress", "@example.com", "user@", "user@localhost",
		"a@b@example.com", "user name@example.com", "<user@example.com>", "user@exa_mple.com",
		"user@-example.com", "user@example..com",
	}
	for _, in := range invalid {
		t.Run("invalid "+in, func(t *testing.T) {
			_, _, err := NormalizeEmail(in)
			assert.ErrorIs(t, err, ErrInvalidEmail)
		})
	}
}

func TestNormalizeDomain(t *testing.T) {
	got, err := NormalizeDomain(" @Mailinator.COM. ")
	require.NoError(t, err)
	assert.Equal(t, "mailinator.com", got)

	for _, in := range []string{"", "@", "com", "exa mple.com", "example-.com"} {
		_, err := NormalizeDomain(in)
		assert.ErrorIs(t, err, ErrInvalidDomain, in)
	}
}

func TestReasonPolicy(t *testing.T) {
	for r := range reasonNames {
		assert.True(t, r.Valid(), r.String())
		assert.GreaterOrEqual(t, r.Policy().Threshold, 1, r.String())
	}
	assert.Zero(t, ReasonInvalidMailbox.Policy().TTL, "hard bounces block permanently")
	assert.Positive(t, ReasonMailboxFull.Policy().TTL, "mailbox full is temporary")
	assert.Positive(t, ReasonBlocked.Policy().TTL, "policy blocks are often about our IP, so temporary")
	assert.False(t, Reason(99).Valid())
	assert.Equal(t, "unknown", Reason(99).String())
}
