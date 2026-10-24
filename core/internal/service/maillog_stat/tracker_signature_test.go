package maillog_stat

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	// Keep unit tests independent from the database-backed secret.
	SetTrackerSecret([]byte("unit-test-tracker-secret-0123456789"))
}

func TestEncryptSigned_RoundTrip(t *testing.T) {
	token := EncryptSigned(map[string]interface{}{"type": "click", "url": "https://example.com"})
	require.Contains(t, token, tokenSignatureSeparator)

	var data map[string]interface{}
	signed, err := DecryptVerified(token, &data)
	require.NoError(t, err)
	assert.True(t, signed)
	assert.Equal(t, "https://example.com", data["url"])
}

func TestDecryptVerified_RejectsTamperedSignature(t *testing.T) {
	token := EncryptSigned(map[string]interface{}{"type": "open"})
	payload, sig := SplitToken(token)

	flipped := []byte(sig)
	if flipped[0] == 'A' {
		flipped[0] = 'B'
	} else {
		flipped[0] = 'A'
	}

	var data map[string]interface{}
	_, err := DecryptVerified(payload+"."+string(flipped), &data)
	assert.ErrorIs(t, err, ErrInvalidTokenSignature)
}

func TestDecryptVerified_ForgedTokenIsNotSigned(t *testing.T) {
	// An attacker can still produce the legacy self-keyed format, but it must
	// never be reported as signed.
	forged := Encrypt(map[string]interface{}{"type": "click", "url": "https://evil.example"})

	var data map[string]interface{}
	signed, err := DecryptVerified(forged, &data)
	require.NoError(t, err)
	assert.False(t, signed)

	// Re-using a signature from another payload must fail.
	other := EncryptSigned(map[string]interface{}{"type": "open"})
	_, sig := SplitToken(other)
	_, err = DecryptVerified(forged+"."+sig, &data)
	assert.ErrorIs(t, err, ErrInvalidTokenSignature)
}

func TestDecrypt_MalformedCiphertextDoesNotPanic(t *testing.T) {
	var data map[string]interface{}
	// 33 bytes: key/iv halves plus a single non-block-aligned byte
	bad := strings.Repeat("A", 44)
	assert.NotPanics(t, func() { _ = Decrypt(bad, &data) })
}

func TestIsSafeRedirectURL(t *testing.T) {
	assert.True(t, IsSafeRedirectURL("https://example.com/a?b=c"))
	assert.True(t, IsSafeRedirectURL("http://example.com"))
	assert.False(t, IsSafeRedirectURL("javascript:alert(1)"))
	assert.False(t, IsSafeRedirectURL("//evil.example"))
	assert.False(t, IsSafeRedirectURL("/relative"))
	assert.False(t, IsSafeRedirectURL("data:text/html,hi"))
}
