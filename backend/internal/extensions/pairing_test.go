package extensions

import (
	"errors"
	"fmt"
	"github.com/omni/backend/internal/repositories"
	"strings"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
)

// The pairing token is the only credential a paired browser holds, and it is
// what binds that browser to a tenant. These tests are pure unit tests (no
// database) so the security-critical logic is verifiable everywhere.

func TestGeneratePairingCode_FormatAndEntropy(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 200; i++ {
		code, err := GeneratePairingCode()
		if err != nil {
			t.Fatalf("GeneratePairingCode error: %v", err)
		}
		if len(code) != PairingCodeLength {
			t.Fatalf("code length = %d, want %d", len(code), PairingCodeLength)
		}
		// A code is typed by hand, so it must avoid characters that are easy to
		// confuse when transcribing (0/O, 1/I/L).
		for _, r := range code {
			if strings.ContainsRune(ambiguousCodeChars, r) {
				t.Fatalf("code %q contains ambiguous character %q", code, r)
			}
			if !strings.ContainsRune(pairingCodeAlphabet, r) {
				t.Fatalf("code %q contains out-of-alphabet character %q", code, r)
			}
		}
		if seen[code] {
			t.Fatalf("duplicate pairing code generated: %q", code)
		}
		seen[code] = true
	}
}

func TestHashAndVerifyToken(t *testing.T) {
	token, hash, err := GeneratePairingToken()
	if err != nil {
		t.Fatalf("GeneratePairingToken error: %v", err)
	}
	if token == "" || hash == "" {
		t.Fatal("token and hash must both be non-empty")
	}
	if token == hash {
		t.Fatal("stored hash must not equal the token itself")
	}
	if !VerifyPairingToken(token, hash) {
		t.Error("a freshly generated token must verify against its hash")
	}
	if VerifyPairingToken(token+"x", hash) {
		t.Error("a modified token must not verify")
	}
	if VerifyPairingToken("", hash) {
		t.Error("an empty token must never verify")
	}
	if VerifyPairingToken(token, "") {
		t.Error("an empty hash must never verify")
	}
	// The token must not be recoverable from the hash in stored form.
	if strings.Contains(hash, token) {
		t.Error("hash must not embed the plaintext token")
	}
}

func TestVerifyPairingToken_ConstantTimeShape(t *testing.T) {
	// Two different wrong tokens must both fail; this also guards against panic
	// on mismatched lengths.
	_, hash, err := GeneratePairingToken()
	if err != nil {
		t.Fatalf("GeneratePairingToken error: %v", err)
	}
	if VerifyPairingToken("short", hash) {
		t.Error("short token must not verify")
	}
	if VerifyPairingToken(strings.Repeat("a", 200), hash) {
		t.Error("long token must not verify")
	}
}

func TestPairingCodeExpiry(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	exp := PairingCodeExpiry(now)
	if got := exp.Sub(now); got != PairingCodeTTL {
		t.Errorf("expiry window = %v, want %v", got, PairingCodeTTL)
	}
}

func TestValidateExtensionID(t *testing.T) {
	valid := []string{
		"abcdefghijklmnopabcdefghijklmnop", // 32-char Chrome ID shape
		"ext-uuid-1234",
		"a1b2c3",
	}
	for _, id := range valid {
		if err := ValidateExtensionID(id); err != nil {
			t.Errorf("ValidateExtensionID(%q) = %v, want nil", id, err)
		}
	}

	invalid := []string{
		"",                       // empty
		"has space",              // whitespace
		"has/slash",              // path separator
		"has;injection",          // delimiter
		"has'quote",              // quote
		"has\"double",            // double quote
		"has<angle>",             // markup
		"has\nnewline",           // control character
		strings.Repeat("a", 129), // over length cap
	}
	for _, id := range invalid {
		if err := ValidateExtensionID(id); err == nil {
			t.Errorf("ValidateExtensionID(%q) = nil, want error", id)
		}
	}
}

func TestNormalizeCapabilities(t *testing.T) {
	// Unknown capabilities are dropped rather than stored: a compromised or
	// modified extension must not be able to enlarge its own privileges by
	// registering capabilities the server does not implement.
	got := NormalizeCapabilities([]string{
		models.CapabilityShopeeScrape,
		"evil_capability",
		models.CapabilityTabDiscovery,
		"",
	})

	if len(got) != 2 {
		t.Fatalf("NormalizeCapabilities returned %v, want exactly the 2 known capabilities", got)
	}
	if !got.Has(models.CapabilityShopeeScrape) || !got.Has(models.CapabilityTabDiscovery) {
		t.Errorf("known capabilities were dropped: %v", got)
	}
	for _, c := range got {
		if c == "evil_capability" {
			t.Error("unknown capability must be filtered out")
		}
	}
}

func TestNormalizeCapabilities_Deduplicates(t *testing.T) {
	got := NormalizeCapabilities([]string{
		models.CapabilityShopeeScrape,
		models.CapabilityShopeeScrape,
	})
	if len(got) != 1 {
		t.Errorf("duplicates must collapse to one entry, got %v", got)
	}
}

func TestNormalizeCapabilities_EmptyAndNil(t *testing.T) {
	if got := NormalizeCapabilities(nil); got != nil {
		t.Errorf("nil input must yield nil, got %v", got)
	}
	if got := NormalizeCapabilities([]string{}); len(got) != 0 {
		t.Errorf("empty input must yield empty, got %v", got)
	}
	// All-unknown input must yield empty, never a partial pass-through.
	if got := NormalizeCapabilities([]string{"nope", "also-nope"}); len(got) != 0 {
		t.Errorf("all-unknown input must yield empty, got %v", got)
	}
}

func TestErrPairingUnusableIsSingleSentinel(t *testing.T) {
	// A caller must not be able to distinguish "unknown code" from "expired" or
	// "already used" — doing so would let an attacker probe for valid codes.
	//
	// The sentinel is produced by the repository's consume query, so assert the
	// package-level value is a real error and that a caller can match it with
	// errors.Is through wrapping (which is how handlers will classify it).
	sentinel := repositories.ErrPairingCodeUnusable
	if sentinel == nil {
		t.Fatal("ErrPairingCodeUnusable must be defined")
	}

	wrapped := fmt.Errorf("pairing confirm: %w", sentinel)
	if !errors.Is(wrapped, repositories.ErrPairingCodeUnusable) {
		t.Error("wrapped unusable-code errors must still match the sentinel")
	}

	// A different failure must NOT match, or callers would treat e.g. a
	// database outage as an invalid code and mislead the operator.
	other := errors.New("connection refused")
	if errors.Is(other, repositories.ErrPairingCodeUnusable) {
		t.Error("unrelated errors must not match the pairing sentinel")
	}
}
