package extensions

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/omni/backend/internal/models"
)

// Pairing constants.
const (
	// PairingCodeLength is how many characters a human types to pair. Eight
	// characters from a 32-symbol alphabet is 40 bits, which is ample for a
	// code that is single-use and expires in minutes.
	PairingCodeLength = 8

	// PairingCodeTTL is deliberately short: a pairing code is a bearer
	// credential while it lives.
	PairingCodeTTL = 5 * time.Minute

	// maxExtensionIDLen caps an accepted extension ID.
	maxExtensionIDLen = 128
)

// pairingCodeAlphabet is the symbol set for pairing codes. It excludes visually
// ambiguous characters (0/O and 1/I/L) so a code can be read aloud or retyped
// without transcription errors. That is a real source of failed pairings, not a
// cosmetic concern.
//
// This is built explicitly rather than relying on base32's alphabet: base32
// includes "L", which is one of the lookalikes we must exclude.
const pairingCodeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

// ambiguousCodeChars are characters that MUST NOT appear in a code.
const ambiguousCodeChars = "01OIL"

// knownCapabilities is the allow-list of capabilities the server implements.
//
// An allow-list (rather than a deny-list) means a modified or hostile extension
// cannot widen its own privileges by registering something new: unknown values
// are dropped at registration.
var knownCapabilities = map[string]bool{
	models.CapabilityTabDiscovery: true,
	models.CapabilityShopeeScrape: true,
}

// GeneratePairingCode returns a fresh, human-typable pairing code.
//
// Symbols are drawn with crypto/rand. Values in the incomplete final block are
// rejected rather than folded with modulo, because 256 is not a multiple of the
// alphabet size and modulo would make the earliest symbols slightly more likely.
func GeneratePairingCode() (string, error) {
	const alphabetLen = len(pairingCodeAlphabet)
	const cutoff = 256 - (256 % alphabetLen)

	out := make([]byte, 0, PairingCodeLength)
	var buf [1]byte

	for len(out) < PairingCodeLength {
		if _, err := rand.Read(buf[:]); err != nil {
			return "", fmt.Errorf("generate pairing code: %w", err)
		}
		if int(buf[0]) >= cutoff {
			continue // unbiased rejection
		}
		out = append(out, pairingCodeAlphabet[int(buf[0])%alphabetLen])
	}

	return string(out), nil
}

// GeneratePairingToken returns a new bearer token and the hash to persist.
//
// The plaintext is returned to the caller exactly once and never stored; only
// the hash is written to the database, so a database disclosure does not yield
// usable extension credentials.
func GeneratePairingToken() (token string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", "", fmt.Errorf("generate pairing token: %w", err)
	}
	token = hex.EncodeToString(buf)
	return token, HashPairingToken(token), nil
}

// HashPairingToken hashes a token for storage.
//
// SHA-256 (not bcrypt) is appropriate here: the input is 256 bits of CSPRNG
// output, so there is no low-entropy guess space for a slow KDF to protect
// against, and pairing must stay cheap on every WebSocket connect.
func HashPairingToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// VerifyPairingToken reports whether token matches the stored hash.
//
// Compared with hmac.Equal so the check is constant-time: a byte-by-byte
// comparison leaks, through timing, how long a shared prefix is, which is
// enough to reconstruct a token one byte at a time.
func VerifyPairingToken(token, hash string) bool {
	if token == "" || hash == "" {
		return false
	}
	expected := HashPairingToken(token)
	return hmac.Equal([]byte(expected), []byte(hash))
}

// PairingCodeExpiry returns the expiry instant for a code issued at now.
func PairingCodeExpiry(now time.Time) time.Time {
	return now.Add(PairingCodeTTL)
}

// ValidateExtensionID rejects an extension identifier that is empty, over-long,
// or contains characters that could be abused in a path, query, log line, or
// intercepted query. The identifier arrives from the browser, so it is untrusted
// input even though it is only ever a lookup key.
func ValidateExtensionID(id string) error {
	if id == "" {
		return fmt.Errorf("extension_id is required")
	}
	if len(id) > maxExtensionIDLen {
		return fmt.Errorf("extension_id exceeds %d characters", maxExtensionIDLen)
	}
	for _, r := range id {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '-' || r == '_':
		default:
			return fmt.Errorf("extension_id contains an invalid character: %q", r)
		}
	}
	return nil
}

// NormalizeCapabilities filters an advertised capability list against the
// server's allow-list, dropping unknown values and duplicates.
//
// Returns nil for empty input so the JSONB column stores NULL rather than "[]"
// — matching how the rest of the codebase treats an absent list.
func NormalizeCapabilities(in []string) models.StringList {
	if len(in) == 0 {
		return nil
	}
	out := make(models.StringList, 0, len(in))
	seen := make(map[string]bool, len(in))
	for _, c := range in {
		if !knownCapabilities[c] || seen[c] {
			continue
		}
		seen[c] = true
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

// HasCapability reports whether a normalized capability list grants cap.
func HasCapability(caps models.StringList, cap string) bool {
	return caps.Has(cap)
}

// sanitizeForLog strips characters that could forge a log line if an untrusted
// identifier is ever logged.
func sanitizeForLog(s string) string {
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}
