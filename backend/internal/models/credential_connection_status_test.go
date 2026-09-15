package models

import (
	"encoding/json"
	"testing"
	"time"
)

// Phase 9 / Bug B / Bug D — CredentialConnection expiry-aware helpers.
//
// The API currently reports status="connected" for any row with the raw DB
// column set to "connected", regardless of whether the tokens have expired.
// The UI has no way to distinguish "connected + token healthy" from
// "connected + token dead". Fix: expose EffectiveStatus + IsAccessTokenExpired
// + IsRefreshTokenExpired helpers, and surface both expiry timestamps + the
// derived status in the masked response.

// TestCredentialConnection_IsAccessTokenExpired covers the boundary and
// zero-value cases.
func TestCredentialConnection_IsAccessTokenExpired(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	cases := []struct {
		name        string
		tokenExpiry int64
		want        bool
	}{
		{"future_1_hour", nowMs + 3600_000, false},
		{"past_1_hour", nowMs - 3600_000, true},
		{"zero_never_set", 0, true},                                 // no expiry = treat as expired (safer default)
		{"exactly_now", nowMs - 1, true},                            // just past
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &CredentialConnection{TokenExpiry: tc.tokenExpiry}
			if got := c.IsAccessTokenExpired(); got != tc.want {
				t.Errorf("IsAccessTokenExpired(TokenExpiry=%d) = %v, want %v", tc.tokenExpiry, got, tc.want)
			}
		})
	}
}

// TestCredentialConnection_IsRefreshTokenExpired.
func TestCredentialConnection_IsRefreshTokenExpired(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	cases := []struct {
		name          string
		refreshExpiry int64
		want          bool
	}{
		{"future", nowMs + 30*86400_000, false},
		{"past", nowMs - 1000, true},
		{"zero", 0, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := &CredentialConnection{RefreshExpiry: tc.refreshExpiry}
			if got := c.IsRefreshTokenExpired(); got != tc.want {
				t.Errorf("IsRefreshTokenExpired(RefreshExpiry=%d) = %v, want %v", tc.refreshExpiry, got, tc.want)
			}
		})
	}
}

// TestCredentialConnection_EffectiveStatus is the policy the UI needs:
//   - disabled row → "disconnected"
//   - refresh token expired → "expired" (needs re-OAuth, no recovery path)
//   - access token expired but refresh alive → "refresh_required"
//   - status=connected + healthy tokens → "connected"
//   - other explicit statuses pass through
func TestCredentialConnection_EffectiveStatus(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	past := nowMs - 1000
	future := nowMs + 3600_000
	longFuture := nowMs + 30*86400_000
	disabledAt := time.Now().Add(-1 * time.Hour)

	cases := []struct {
		name string
		conn CredentialConnection
		want string
	}{
		{
			name: "healthy_connected",
			conn: CredentialConnection{Status: "connected", TokenExpiry: future, RefreshExpiry: longFuture},
			want: "connected",
		},
		{
			name: "disabled_row",
			conn: CredentialConnection{Status: "connected", TokenExpiry: future, RefreshExpiry: longFuture, DisabledAt: &disabledAt},
			want: "disconnected",
		},
		{
			name: "refresh_expired_hard_dead",
			conn: CredentialConnection{Status: "connected", TokenExpiry: past, RefreshExpiry: past},
			want: "expired",
		},
		{
			name: "access_expired_refresh_alive",
			conn: CredentialConnection{Status: "connected", TokenExpiry: past, RefreshExpiry: longFuture},
			want: "refresh_required",
		},
		{
			name: "explicit_status_passes_through",
			conn: CredentialConnection{Status: "action_required", TokenExpiry: future, RefreshExpiry: longFuture},
			want: "action_required",
		},
		{
			name: "no_expiries_recorded_still_reports_expired",
			conn: CredentialConnection{Status: "connected", TokenExpiry: 0, RefreshExpiry: 0},
			want: "expired",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.conn.EffectiveStatus(); got != tc.want {
				t.Errorf("EffectiveStatus = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestCredentialConnectionMaskedResponse_ExposesRefreshExpiryAndEffectiveStatus
// — Bug D: the masked response must expose refresh_expiry + effective_status
// so the UI can render the correct badge without a second round-trip.
func TestCredentialConnectionMaskedResponse_ExposesRefreshExpiryAndEffectiveStatus(t *testing.T) {
	nowMs := time.Now().UnixMilli()
	c := &CredentialConnection{
		StoreIdentifier: "530635055",
		Platform:        "shopee",
		Status:          "connected",
		TokenExpiry:     nowMs - 1000,       // expired
		RefreshExpiry:   nowMs - 500,        // also expired
	}
	got := c.ToMaskedResponse()
	if got.RefreshExpiry != c.RefreshExpiry {
		t.Errorf("RefreshExpiry = %d, want %d", got.RefreshExpiry, c.RefreshExpiry)
	}
	if got.EffectiveStatus != "expired" {
		t.Errorf("EffectiveStatus = %q, want expired (both tokens dead)", got.EffectiveStatus)
	}
	// Assert JSON key names use snake_case per AGENTS.md.
	body, _ := json.Marshal(got)
	str := string(body)
	if !contains(str, `"refresh_expiry":`) {
		t.Errorf("masked response missing refresh_expiry key; body=%s", str)
	}
	if !contains(str, `"effective_status":`) {
		t.Errorf("masked response missing effective_status key; body=%s", str)
	}
}
