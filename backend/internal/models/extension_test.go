package models

import (
	"testing"
	"time"
)

// RED: these tests define the contract for the extensions feature models
// before the structs exist. They must fail first, then pass.

func TestExtensionModel_TableNameAndColumns(t *testing.T) {
	e := Extension{}
	if got := e.TableName(); got != "extensions" {
		t.Errorf("Extension.TableName() = %q, want %q", got, "extensions")
	}
}

func TestExtensionModel_FieldMapping(t *testing.T) {
	now := time.Now()
	e := Extension{
		ID:               1,
		ExtensionID:      "ext-uuid-1",
		UserID:           nil,
		Hostname:         "browser-extension",
		BrowserInfo:      "Chrome/122",
		ExtensionVersion: "1.0.0",
		ProtocolVersion:  "1",
		Capabilities:     []string{"tab_discovery", "shopee_scrape"},
		Status:           "connected",
		LastSeen:         &now,
		PairedAt:         now,
	}

	if e.ExtensionID != "ext-uuid-1" {
		t.Error("ExtensionID not retained")
	}
	if len(e.Capabilities) != 2 {
		t.Errorf("Capabilities len = %d, want 2", len(e.Capabilities))
	}
	if e.Status != "connected" {
		t.Errorf("Status = %q, want connected", e.Status)
	}
}

func TestPairingCodeModel_TableName(t *testing.T) {
	p := PairingCode{}
	if got := p.TableName(); got != "pairing_codes" {
		t.Errorf("PairingCode.TableName() = %q, want %q", got, "pairing_codes")
	}
}

func TestScrapedProductModel_TableNameAndColumns(t *testing.T) {
	s := ScrapedProduct{}
	if got := s.TableName(); got != "scraped_products" {
		t.Errorf("ScrapedProduct.TableName() = %q, want %q", got, "scraped_products")
	}

	p := ScrapedProduct{
		JobID:        "job-1",
		ProductName:  "Test Product",
		Price:        "15000",
		Sold:         "100",
		Link:         "https://shopee.co.id/product-i.111.222",
		ImageURL:     "https://cf.shopee.co.id/file/abc",
		ShopeeItemID: "222",
		PageNumber:   1,
	}
	if p.ShopeeItemID != "222" {
		t.Errorf("ShopeeItemID = %q, want 222", p.ShopeeItemID)
	}
	if p.PageNumber != 1 {
		t.Errorf("PageNumber = %d, want 1", p.PageNumber)
	}
}

// ScrapedProduct with no tenant_id column — schema isolation handles tenancy.
// A tenant_id column would be a regression per models/job.go convention.
func TestScrapedProductModel_NoTenantIDColumn(t *testing.T) {
	s := ScrapedProduct{}
	if hasField(s, "TenantID") {
		t.Error("ScrapedProduct must NOT have a TenantID field (schema isolation handles tenancy)")
	}
}

func TestExtensionModel_NoTenantIDColumn(t *testing.T) {
	e := Extension{}
	if hasField(e, "TenantID") {
		t.Error("Extension must NOT have a TenantID field (schema isolation handles tenancy)")
	}
}
