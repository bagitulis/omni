package shopee

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/models"
)

// The job layer's conversion and URL selection are pure, so they are tested
// without a browser or a database.

func TestScrapeJobData_NoTenantField(t *testing.T) {
	// The job lives in the tenant's own schema, so there is no tenant id to
	// carry. A field here would imply the tenant could be supplied by the
	// payload, which is the cross-tenant hazard the schema design removes.
	raw, err := json.Marshal(ScrapeJobData{Mode: "search", Query: "x", ExtensionID: "e"})
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var fields map[string]any
	if err := json.Unmarshal(raw, &fields); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, present := fields["tenant_id"]; present {
		t.Fatal("ScrapeJobData must not carry a tenant_id")
	}
}

func TestScrapeJobData_RoundTrip(t *testing.T) {
	original := ScrapeJobData{
		Mode:         "shop",
		ShopURL:      "https://shopee.co.id/testshop",
		ShopID:       "12345",
		MaxPages:     3,
		MaxProducts:  100,
		ExtensionID:  "ext-1",
		ResultsJobID: "job-abc",
	}

	raw, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var decoded ScrapeJobData
	if err := json.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if decoded.Mode != original.Mode || decoded.ShopURL != original.ShopURL {
		t.Errorf("round trip lost data: %+v", decoded)
	}
	if decoded.JobID() != "job-abc" {
		t.Errorf("JobID() = %q, want job-abc", decoded.JobID())
	}
}

func TestEntryURLFor(t *testing.T) {
	cases := []struct {
		name string
		data ScrapeJobData
		want string
	}{
		{
			name: "search builds a keyword URL",
			data: ScrapeJobData{Mode: "search", Query: "keyboard"},
			want: "https://shopee.co.id/search?keyword=keyboard",
		},
		{
			name: "search encodes spaces as plus",
			data: ScrapeJobData{Mode: "search", Query: "mechanical keyboard"},
			want: "https://shopee.co.id/search?keyword=mechanical+keyboard",
		},
		{
			name: "search honours a custom base URL",
			data: ScrapeJobData{Mode: "search", Query: "x", BaseURL: "https://shopee.co.id/"},
			want: "https://shopee.co.id/search?keyword=x",
		},
		{
			name: "shop uses the shop URL",
			data: ScrapeJobData{Mode: "shop", ShopURL: "https://shopee.co.id/testshop"},
			want: "https://shopee.co.id/testshop",
		},
		{
			name: "product uses the product URL",
			data: ScrapeJobData{Mode: "product", ProductURL: "https://shopee.co.id/product/1/2"},
			want: "https://shopee.co.id/product/1/2",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := entryURLFor(&tc.data); got != tc.want {
				t.Errorf("entryURLFor() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestToScrapedProducts(t *testing.T) {
	parsed := []ParsedProduct{
		{
			ProductName:  "A",
			Price:        "15000",
			Sold:         "42",
			Link:         "https://shopee.co.id/product/10/20",
			ImageURL:     "img",
			ShopeeItemID: "20",
			ShopID:       "10",
			Page:         2,
			Source:       models.ScrapeSourceNetwork,
		},
	}
	data := &ScrapeJobData{Mode: "search", Query: "keyboard", ShopID: "fallback-shop"}

	rows := toScrapedProducts(parsed, "job-1", data)
	if len(rows) != 1 {
		t.Fatalf("got %d rows, want 1", len(rows))
	}

	r := rows[0]
	if r.JobID != "job-1" {
		t.Errorf("JobID = %q, want job-1", r.JobID)
	}
	if r.Platform != "shopee" {
		t.Errorf("Platform = %q, want shopee", r.Platform)
	}
	if r.ScrapeMode != "search" || r.Query != "keyboard" {
		t.Errorf("mode/query not carried: %+v", r)
	}
	// The product's own shop id must win over the job-level fallback.
	if r.ShopID != "10" {
		t.Errorf("ShopID = %q, want the product's own 10", r.ShopID)
	}
	if r.PageNumber != 2 {
		t.Errorf("PageNumber = %d, want 2", r.PageNumber)
	}
	if r.Source != models.ScrapeSourceNetwork {
		t.Errorf("Source = %q, want network", r.Source)
	}
}

func TestToScrapedProducts_FallsBackToJobShopID(t *testing.T) {
	// When the parser could not extract a shop id (for example a shop-scrape
	// listing), the job-level id is the useful fallback.
	parsed := []ParsedProduct{{ProductName: "A", Link: "https://shopee.co.id/product/1/2", ShopeeItemID: "2"}}
	data := &ScrapeJobData{Mode: "shop", ShopID: "jobshop"}

	rows := toScrapedProducts(parsed, "job-1", data)
	if rows[0].ShopID != "jobshop" {
		t.Errorf("ShopID = %q, want the job fallback jobshop", rows[0].ShopID)
	}
}

func TestToScrapedProducts_PreservesSourcePerRow(t *testing.T) {
	// Provenance is per row, so a mixed run is represented honestly rather than
	// labelling everything with whichever path ran last.
	parsed := []ParsedProduct{
		{ProductName: "A", Link: "https://shopee.co.id/product/1/2", ShopeeItemID: "2", Source: models.ScrapeSourceDOM},
		{ProductName: "B", Link: "https://shopee.co.id/product/1/3", ShopeeItemID: "3", Source: models.ScrapeSourceNetwork},
	}

	rows := toScrapedProducts(parsed, "job-1", &ScrapeJobData{Mode: "search"})
	if rows[0].Source != models.ScrapeSourceDOM || rows[1].Source != models.ScrapeSourceNetwork {
		t.Errorf("per-row sources not preserved: %q, %q", rows[0].Source, rows[1].Source)
	}
}

func TestMarkSource(t *testing.T) {
	products := []ParsedProduct{{ProductName: "A"}, {ProductName: "B"}}
	markSource(products, models.ScrapeSourceDOM)

	for i, p := range products {
		if p.Source != models.ScrapeSourceDOM {
			t.Errorf("products[%d].Source = %q, want dom", i, p.Source)
		}
	}
}

func TestTrimSlashAndFirstNonEmpty(t *testing.T) {
	if got := trimSlash("https://x.com///"); got != "https://x.com" {
		t.Errorf("trimSlash = %q", got)
	}
	if got := trimSlash(""); got != "" {
		t.Errorf("trimSlash(\"\") = %q, want empty", got)
	}
	if got := firstNonEmpty("", "", "c", "d"); got != "c" {
		t.Errorf("firstNonEmpty = %q, want c", got)
	}
	if got := firstNonEmpty("", ""); got != "" {
		t.Errorf("firstNonEmpty with all empty = %q, want empty", got)
	}
}
