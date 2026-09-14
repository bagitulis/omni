package extensions

import (
	"time"
)

// TokenBinding ties one pairing token to exactly one tenant.
//
// This is the security boundary of the feature: it is the single place where a
// tenant is decided for an extension, and it is constructed from server-side
// values only (the authenticated user's tenant at pairing time). There is
// deliberately no way to retarget a binding, and no representation for a token
// that spans two tenants.
type TokenBinding struct {
	// TokenHash is the SHA-256 hash of the pairing token. The plaintext token is
	// never stored.
	TokenHash string

	// TenantID is the tenant resolved server-side when the code was redeemed.
	TenantID string

	// UserID is the user who generated the pairing code, when known.
	UserID string

	// ExpiresAt is when the binding stops being valid. A zero value means the
	// binding does not expire.
	ExpiresAt time.Time
}

// Matches reports whether other refers to the same binding.
//
// Both the token hash and the tenant must agree. Comparing the token alone would
// be sufficient in practice, but requiring the tenant too means a bug that
// swapped a tenant on lookup would surface as a mismatch instead of silently
// granting access to the wrong schema.
func (b TokenBinding) Matches(other TokenBinding) bool {
	if b.TokenHash == "" || other.TokenHash == "" {
		return false
	}
	if b.TokenHash != other.TokenHash {
		return false
	}
	return b.TenantID == other.TenantID
}

// Valid reports whether the binding can be trusted at time now.
//
// Fails closed: a missing token hash or empty tenant is invalid, mirroring the
// platform's "no default tenant" rule.
func (b TokenBinding) Valid(now time.Time) bool {
	if b.TokenHash == "" || b.TenantID == "" {
		return false
	}
	if b.ExpiresAt.IsZero() {
		return true
	}
	return now.Before(b.ExpiresAt)
}

// PairRequest is the extension→server request that redeems a pairing code.
//
// Note the absence of a TenantID field: the tenant comes from the pairing code,
// which was itself created by an authenticated user. Adding a tenant field here
// would let a client choose its own tenant, so isolation_test.go asserts it
// stays absent.
type PairRequest struct {
	// Code is the short human-typed pairing code.
	Code string `json:"code"`

	// ExtensionID is the browser's own generated identifier.
	ExtensionID string `json:"extension_id"`

	// Hostname, BrowserInfo, ChromeVersion, ExtensionVersion and ProtocolVersion
	// are descriptive metadata recorded for operator visibility. None of them
	// are trusted for authorisation.
	Hostname         string `json:"hostname"`
	BrowserInfo      string `json:"browser_info"`
	ChromeVersion    string `json:"chrome_version"`
	ExtensionVersion string `json:"extension_version"`
	ProtocolVersion  string `json:"protocol_version"`

	// Capabilities is the extension's advertised feature list. It is filtered
	// against the server's allow-list on registration, so an extension cannot
	// widen its own privileges.
	Capabilities []string `json:"capabilities"`
}

// ConnectRequest is the WebSocket handshake payload.
//
// As with PairRequest there is no TenantID: the tenant is derived from the
// token server-side.
type ConnectRequest struct {
	Token       string `json:"token"`
	ExtensionID string `json:"extension_id"`
}

// ScrapeRequest asks for a Shopee scrape job.
//
// Still no TenantID. The job is written into the schema of the tenant bound to
// the pairing token that carried the request.
type ScrapeRequest struct {
	Mode        string `json:"mode"`        // search | shop | product
	Query       string `json:"query"`       // search mode
	ShopURL     string `json:"shop_url"`    // shop mode
	ProductURL  string `json:"product_url"` // product mode
	MaxPages    int    `json:"max_pages"`
	MaxProducts int    `json:"max_products"`
	SortBy      string `json:"sort_by"`
}
