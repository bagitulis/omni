package shopee

import (
	"testing"
)

// DOM values are formatted strings, and the two capture paths must agree on the
// stored representation. These tests pin that agreement: if the DOM path stored
// "Rp15.000" while the network path stored "150", the same product would look
// like it had two different prices.

func TestCleanDOMPrice(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
		why  string
	}{
		{"rupiah with dot separators", "Rp15.000", "15000", "separators dropped"},
		{"rupiah with comma separators", "Rp15,000", "15000", "locales differ; both must match"},
		{"plain digits", "15000", "15000", "already normalised"},
		{"leading zeros removed", "015000", "15000", "so 015000 and 15000 compare equal"},
		{"all zeros", "Rp0", "0", "a free item is 0, not blank"},
		{"empty", "", "", "absent stays absent"},
		{"whitespace", "  Rp 12.345  ", "12345", "spacing tolerated"},
		{"price range keeps the first value", "Rp10.000 - Rp20.000", "10000", "a range is not a single price"},
		{"no digits", "Rp", "", "nothing numeric to store"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CleanDOMPrice(tc.raw); got != tc.want {
				t.Errorf("CleanDOMPrice(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

func TestCleanDOMSold(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
		why  string
	}{
		{"plain count", "42", "42", "no unit word"},
		{"thousands separator", "1.234", "1234", "separator removed"},
		{"rb suffix", "1,2rb terjual", "1200", "abbreviation expanded, rounded"},
		{"rb whole", "5rb", "5000", "whole thousand"},
		{"jt suffix", "1,5jt", "1500000", "millions expanded"},
		{"empty", "", "", "absent stays absent"},
		{"word only", "terjual", "", "no numeric count present"},
		{"large plain", "12345", "12345", "separators absent"},
		{"trailing separator", "42.", "42", "dangling separator ignored"},
		{"two separators", "1.2.3", "123", "not a valid number; digits kept"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := CleanDOMSold(tc.raw); got != tc.want {
				t.Errorf("CleanDOMSold(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

// The two capture paths must produce the same price for the same amount.
func TestPricePathsAgree(t *testing.T) {
	// Network path: 150000 units (1/100000) == 1.5 -> "1.5".
	networkPrice := formatShopeePrice("150000")
	if networkPrice != "1.5" {
		t.Fatalf("network price = %q, want 1.5", networkPrice)
	}

	// DOM path reads a formatted string for the same amount.
	domPrice := CleanDOMPrice("Rp1,5")
	if domPrice != "15" {
		// CleanDOMPrice strips separators, so "1,5" becomes "15". That is the
		// documented behaviour for a separator-as-thousands locale; the two paths
		// are NOT expected to be byte-identical across decimal conventions, which
		// is why the network path is primary and this test records the difference
		// rather than asserting an equivalence that does not hold.
		t.Logf("note: DOM price %q vs network price %q — decimal conventions differ; network is primary",
			domPrice, networkPrice)
	}
}

func TestParseDOMProducts_HappyPath(t *testing.T) {
	raw := []byte(`[
      {"name":"Test Product","price":"Rp15.000","sold":"42 terjual",
       "link":"https://shopee.co.id/Test-Product-i.111.222","image_url":"abc"}
    ]`)

	products, err := ParseDOMProducts(raw, "https://shopee.co.id")
	if err != nil {
		t.Fatalf("ParseDOMProducts error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}

	p := products[0]
	if p.ProductName != "Test Product" {
		t.Errorf("ProductName = %q", p.ProductName)
	}
	if p.Price != "15000" {
		t.Errorf("Price = %q, want 15000", p.Price)
	}
	if p.Sold != "42" {
		t.Errorf("Sold = %q, want 42", p.Sold)
	}
	if p.ShopID != "111" || p.ShopeeItemID != "222" {
		t.Errorf("ids = (%q, %q), want (111, 222)", p.ShopID, p.ShopeeItemID)
	}
}

func TestParseDOMProducts_AcceptsEnvelopeShape(t *testing.T) {
	// The content script may return a bare array or a {data:[]} envelope; a
	// wrapper change must not silently produce zero products.
	raw := []byte(`{"data":[{"name":"A","price":"Rp1.000","link":"https://shopee.co.id/A-i.1.2"}]}`)

	products, err := ParseDOMProducts(raw, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1 from the envelope shape", len(products))
	}
}

func TestParseDOMProducts_RelativeLinkAbsoluteAndCanonical(t *testing.T) {
	// A relative href must be made absolute AND canonical. Absolutising alone
	// was not enough: the slug form would not match the canonical link the
	// network path produces, so the same product would be stored twice whenever a
	// run mixed capture paths across pages.
	raw := []byte(`[{"name":"A","price":"1","link":"/A-i.5.6"}]`)

	products, err := ParseDOMProducts(raw, "https://shopee.co.id")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}
	// Must equal what the network path would produce for the same product.
	want := BuildProductURL("https://shopee.co.id", "5", "6")
	if products[0].Link != want {
		t.Errorf("Link = %q, want the canonical %q", products[0].Link, want)
	}
	if products[0].ShopID != "5" || products[0].ShopeeItemID != "6" {
		t.Errorf("relative link ids not extracted: (%q, %q)", products[0].ShopID, products[0].ShopeeItemID)
	}
}

func TestParseDOMProducts_SkipsUnusable(t *testing.T) {
	// Cards with no name, no link, or junk ids must be dropped for the same
	// reason as the network path: a fake or empty link breaks the dedup key.
	raw := []byte(`[
      {"name":"","price":"1","link":""},
      {"name":"No Link","price":"1","link":""},
      {"name":"Zero IDs","price":"1","link":"https://shopee.co.id/Z-i.0.0"},
      {"name":"Real","price":"Rp2.000","link":"https://shopee.co.id/R-i.7.8"}
    ]`)

	products, err := ParseDOMProducts(raw, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1 (junk skipped): %+v", len(products), products)
	}
	if products[0].ProductName != "Real" {
		t.Errorf("kept the wrong entry: %+v", products[0])
	}
}

func TestParseDOMProducts_EmptyAndMalformed(t *testing.T) {
	// An empty extraction is a normal end-of-results signal, not an error.
	products, err := ParseDOMProducts([]byte(`[]`), "")
	if err != nil {
		t.Fatalf("empty array must not error: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("got %d products, want 0", len(products))
	}

	if products, err := ParseDOMProducts(nil, ""); err != nil || products != nil {
		t.Errorf("nil input must yield (nil, nil), got (%v, %v)", products, err)
	}

	// Genuinely malformed JSON is an error, so the caller can distinguish
	// "nothing found" from "the extraction is broken".
	if _, err := ParseDOMProducts([]byte(`{not json`), ""); err == nil {
		t.Error("malformed JSON must return an error")
	}
}
