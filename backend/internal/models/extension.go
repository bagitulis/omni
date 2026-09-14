package models

import "time"

// Extension pairing/connection status values.
const (
	// ExtensionStatusConnected means the extension has an active WebSocket.
	ExtensionStatusConnected = "connected"
	// ExtensionStatusDisconnected means the extension is paired but offline.
	ExtensionStatusDisconnected = "disconnected"
)

// Capabilities advertised by an extension during registration.
const (
	// CapabilityTabDiscovery allows the server to list tabs on the host browser.
	CapabilityTabDiscovery = "tab_discovery"
	// CapabilityShopeeScrape allows Shopee scrape job dispatch.
	CapabilityShopeeScrape = "shopee_scrape"
)

// Extension represents a paired Chrome extension install.
//
// Matches PostgreSQL table: tenant_{name}.extensions
// Note: No tenant_id column - schema isolation handles multi-tenancy,
// matching the models.Job convention. A pairing token binds exactly one
// extension to exactly one tenant; the tenant is derived server-side and
// is never supplied by the client.
type Extension struct {
	ID               int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	ExtensionID      string     `gorm:"column:extension_id;uniqueIndex;not null" json:"extension_id"`
	UserID           *int64     `gorm:"column:user_id" json:"user_id,omitempty"`
	Hostname         string     `gorm:"column:hostname" json:"hostname,omitempty"`
	BrowserInfo      string     `gorm:"column:browser_info" json:"browser_info,omitempty"`
	ChromeVersion    string     `gorm:"column:chrome_version" json:"chrome_version,omitempty"`
	ExtensionVersion string     `gorm:"column:extension_version" json:"extension_version,omitempty"`
	ProtocolVersion  string     `gorm:"column:protocol_version" json:"protocol_version,omitempty"`
	Capabilities     StringList `gorm:"column:capabilities;type:jsonb" json:"capabilities"`
	Status           string     `gorm:"column:status;not null;default:disconnected" json:"status"`
	TokenHash        string     `gorm:"column:token_hash;index" json:"-"`
	LastSeen         *time.Time `gorm:"column:last_seen" json:"last_seen,omitempty"`
	PairedAt         time.Time  `gorm:"column:paired_at" json:"paired_at"`
	CreatedAt        time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt        time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name.
func (Extension) TableName() string {
	return GetTableName("Extension")
}

// PairingCode is a short-lived, single-use code that binds a browser install to
// the tenant of the authenticated user who generated it.
//
// Matches PostgreSQL table: tenant_{name}.pairing_codes
// Note: No tenant_id column - schema isolation handles multi-tenancy.
type PairingCode struct {
	ID         int64      `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Code       string     `gorm:"column:code;type:varchar(64);uniqueIndex;not null" json:"code"`
	UserID     *int64     `gorm:"column:user_id" json:"user_id,omitempty"`
	Consumed   bool       `gorm:"column:consumed;not null;default:false" json:"consumed"`
	ConsumedAt *time.Time `gorm:"column:consumed_at" json:"consumed_at,omitempty"`
	ExpiresAt  time.Time  `gorm:"column:expires_at;not null;index" json:"expires_at"`
	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name.
func (PairingCode) TableName() string {
	return GetTableName("PairingCode")
}

// IsExpired reports whether the code is past its expiry.
//
// The comparison is inclusive: a code whose deadline is exactly now counts as
// expired. Using After() alone leaves a one-instant window where a code remains
// redeemable at its stated deadline.
func (p PairingCode) IsExpired(now time.Time) bool {
	return !now.Before(p.ExpiresAt)
}

// IsUsable reports whether the code can still be redeemed: not consumed and
// not expired.
func (p PairingCode) IsUsable(now time.Time) bool {
	return !p.Consumed && !p.IsExpired(now)
}

// ScrapedProduct is a single product row collected by a Shopee scrape job.
//
// Matches PostgreSQL table: tenant_{name}.scraped_products
// Note: No tenant_id column - schema isolation handles multi-tenancy.
type ScrapedProduct struct {
	ID           int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID        string    `gorm:"column:job_id;index;not null" json:"job_id"`
	Platform     string    `gorm:"column:platform;not null;default:shopee" json:"platform"`
	ScrapeMode   string    `gorm:"column:scrape_mode" json:"scrape_mode,omitempty"`
	Query        string    `gorm:"column:query" json:"query,omitempty"`
	ShopID       string    `gorm:"column:shop_id" json:"shop_id,omitempty"`
	ProductName  string    `gorm:"column:product_name" json:"product_name"`
	Price        string    `gorm:"column:price" json:"price"`
	Sold         string    `gorm:"column:sold" json:"sold"`
	Link         string    `gorm:"column:link;uniqueIndex:idx_scraped_products_job_link,priority:2" json:"link"`
	ImageURL     string    `gorm:"column:image_url" json:"image_url"`
	ShopeeItemID string    `gorm:"column:shopee_item_id;index" json:"shopee_item_id,omitempty"`
	PageNumber   int       `gorm:"column:page_number;default:1" json:"page_number"`
	Source       string    `gorm:"column:source" json:"source,omitempty"` // network | dom
	CreatedAt    time.Time `gorm:"column:created_at" json:"created_at"`
}

// TableName returns the table name.
func (ScrapedProduct) TableName() string {
	return GetTableName("ScrapedProduct")
}

// Scrape source values, recorded so operators can tell which capture path
// produced a row (network-first is primary, DOM is the fallback).
const (
	ScrapeSourceNetwork = "network"
	ScrapeSourceDOM     = "dom"
)

// Scrape mode values for Shopee jobs.
const (
	ScrapeModeSearch  = "search"
	ScrapeModeShop    = "shop"
	ScrapeModeProduct = "product"
)
