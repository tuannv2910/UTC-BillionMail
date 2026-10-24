package maillog_stat

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"sync"

	"github.com/gogf/gf/v2/frame/g"
)

// Tracking tokens used to carry their own AES key and IV, so anyone could mint
// a valid token and use /pmta/* to forge opens/clicks or as an open redirect
// (#390). Tokens are now "<encrypted payload>.<HMAC-SHA256 tag>" where the tag
// is computed with a server-held secret stored in bm_options.

const (
	trackerSecretOptionName = "pmta_tracker_secret"
	// tokenSignatureSeparator never appears in base64.URLEncoding output.
	tokenSignatureSeparator = "."
	// Truncated HMAC length in bytes (128 bits).
	tokenSignatureBytes = 16
)

var (
	ErrInvalidTokenSignature = errors.New("invalid tracking token signature")

	trackerSecretMu sync.RWMutex
	trackerSecret   []byte
)

// SetTrackerSecret overrides the signing secret (used by tests).
func SetTrackerSecret(secret []byte) {
	trackerSecretMu.Lock()
	defer trackerSecretMu.Unlock()
	trackerSecret = append([]byte(nil), secret...)
}

func getTrackerSecret() []byte {
	trackerSecretMu.RLock()
	s := trackerSecret
	trackerSecretMu.RUnlock()
	if len(s) > 0 {
		return s
	}

	trackerSecretMu.Lock()
	defer trackerSecretMu.Unlock()
	if len(trackerSecret) == 0 {
		trackerSecret = loadOrCreateTrackerSecret(context.Background())
	}
	return trackerSecret
}

func randomSecret() []byte {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return b
}

// loadOrCreateTrackerSecret reads the secret from bm_options, creating it on
// first use. If the database is unavailable a process-local secret is used so
// tokens are still unforgeable (they just won't survive a restart).
func loadOrCreateTrackerSecret(ctx context.Context) (secret []byte) {
	defer func() {
		if r := recover(); r != nil {
			g.Log().Warningf(ctx, "Tracker secret: database unavailable, using in-memory secret: %v", r)
			secret = randomSecret()
		}
	}()

	val, err := g.DB().Model("bm_options").Ctx(ctx).Where("name", trackerSecretOptionName).Value("value")
	if err == nil && val != nil && val.String() != "" {
		if b, decErr := hex.DecodeString(val.String()); decErr == nil && len(b) >= 16 {
			return b
		}
	}

	newSecret := randomSecret()
	_, err = g.DB().Model("bm_options").Ctx(ctx).Insert(g.Map{
		"name":  trackerSecretOptionName,
		"value": hex.EncodeToString(newSecret),
	})
	if err != nil {
		// Another process may have inserted it concurrently — prefer the stored value.
		val, err2 := g.DB().Model("bm_options").Ctx(ctx).Where("name", trackerSecretOptionName).Value("value")
		if err2 == nil && val != nil {
			if b, decErr := hex.DecodeString(val.String()); decErr == nil && len(b) >= 16 {
				return b
			}
		}
		g.Log().Warningf(ctx, "Tracker secret: failed to persist secret, using in-memory secret: %v", err)
	}
	return newSecret
}

func tokenSignature(payload string) string {
	mac := hmac.New(sha256.New, getTrackerSecret())
	mac.Write([]byte(payload))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil)[:tokenSignatureBytes])
}

// SignToken appends an HMAC tag to an encrypted payload.
func SignToken(payload string) string {
	return payload + tokenSignatureSeparator + tokenSignature(payload)
}

// SplitToken separates the payload from its signature. signature is "" for
// legacy unsigned tokens.
func SplitToken(token string) (payload, signature string) {
	if i := strings.LastIndex(token, tokenSignatureSeparator); i >= 0 {
		return token[:i], token[i+1:]
	}
	return token, ""
}

// EncryptSigned encrypts data and signs the resulting token.
func EncryptSigned(data interface{}) string {
	return SignToken(Encrypt(data))
}

// DecryptVerified decrypts a tracking token. signed reports whether the token
// carried a valid server signature; a token with a wrong signature is rejected
// with ErrInvalidTokenSignature. Unsigned (legacy) tokens decrypt with
// signed == false and must not be trusted.
func DecryptVerified(token string, result interface{}) (signed bool, err error) {
	payload, sig := SplitToken(token)
	if sig != "" {
		expected := tokenSignature(payload)
		if !hmac.Equal([]byte(sig), []byte(expected)) {
			return false, ErrInvalidTokenSignature
		}
		signed = true
	}
	if err = Decrypt(payload, result); err != nil {
		return false, err
	}
	return signed, nil
}
