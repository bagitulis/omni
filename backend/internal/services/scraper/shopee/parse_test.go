package shopee

import (
	"strings"
	"testing"
)

// Price arithmetic is integer-only on purpose (float64 cannot represent values
// like 15000000 exactly), so it is table-tested across the boundary cases.

func TestFormatShopeePrice(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want string
		why  string
	}{
		{"empty", "", "", "absent price stays absent, not zero"},
		{"zero", "0", "0", "a free item is 0, not blank"},
		{"whole unit", "100000", "1", "100000 units = 1.00 -> trailing zeros trimmed"},
		{"whole large", "15000000", "150", "no float rounding at scale"},
		{"fractional", "150500", "1.505", "fraction preserved"},
		{"one unit", "1", "0.00001", "smallest representable price"},
		{"pad boundary", "99999", "0.99999", "just below one whole unit"},
		{"exact hundred", "10000000", "100", "trailing zeros removed from fraction"},
		{"null price marker", "-1", "", "Shopee uses a negative value for a null price"},
		{
			name: "formatted input is rejected, not passed through",
			raw:  "Rp15.000",
			want: "",
			why: "this function scales a JSON number literal; an already-formatted " +
				"string belongs to the DOM path and storing it here would put " +
				"uncomparable text in a price column",
		},
		{"whitespace", "  200000  ", "2", "surrounding space tolerated"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := formatShopeePrice(tc.raw); got != tc.want {
				t.Errorf("formatShopeePrice(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

func TestBuildProductURL(t *testing.T) {
	cases := []struct {
		name           string
		base, shop, id string
		want           string
	}{
		{"canonical", "https://shopee.co.id", "111", "222", "https://shopee.co.id/product/111/222"},
		{"trailing slash trimmed", "https://shopee.co.id/", "111", "222", "https://shopee.co.id/product/111/222"},
		{"default base when empty", "", "111", "222", "https://shopee.co.id/product/111/222"},
		{"missing shop yields empty", "https://shopee.co.id", "", "222", ""},
		{"missing item yields empty", "https://shopee.co.id", "111", "", ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := BuildProductURL(tc.base, tc.shop, tc.id); got != tc.want {
				t.Errorf("BuildProductURL(%q,%q,%q) = %q, want %q", tc.base, tc.shop, tc.id, got, tc.want)
			}
		})
	}
}

func TestExtractShopIDs(t *testing.T) {
	cases := []struct {
		name     string
		link     string
		wantShop string
		wantItem string
	}{
		{
			name:     "slug form with -i. segment",
			link:     "https://shopee.co.id/Some-Product-Name-i.111.222",
			wantShop: "111",
			wantItem: "222",
		},
		{
			name:     "canonical form",
			link:     "https://shopee.co.id/product/111/222",
			wantShop: "111",
			wantItem: "222",
		},
		{
			name:     "canonical form with trailing slash",
			link:     "https://shopee.co.id/product/111/222/",
			wantShop: "111",
			wantItem: "222",
		},
		{
			name:     "slug with query string",
			link:     "https://shopee.co.id/Nama-Produk-i.333.444?sp_atk=abc",
			wantShop: "333",
			wantItem: "444",
		},
		{
			name:     "unparseable link yields empty ids",
			link:     "https://shopee.co.id/just-a-page",
			wantShop: "",
			wantItem: "",
		},
		{
			name:     "empty link",
			link:     "",
			wantShop: "",
			wantItem: "",
		},
		{
			name:     "non-numeric product path is not mistaken for ids",
			link:     "https://shopee.co.id/product/abc/def",
			wantShop: "",
			wantItem: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			shop, item := ExtractShopIDs(tc.link)
			if shop != tc.wantShop || item != tc.wantItem {
				t.Errorf("ExtractShopIDs(%q) = (%q, %q), want (%q, %q)",
					tc.link, shop, item, tc.wantShop, tc.wantItem)
			}
		})
	}
}

func TestParseSearchResponse_HappyPath(t *testing.T) {
	body := []byte(`{
      "items": [
        {"item_basic": {"itemid": 222, "shopid": 111, "name": "Test Product",
                        "price": 15000000, "sold": 42, "image": "abc123",
                        "historical_sold": 100}}
      ]
    }`)

	products, err := ParseSearchResponse(body, "https://shopee.co.id")
	if err != nil {
		t.Fatalf("ParseSearchResponse error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}

	p := products[0]
	if p.ProductName != "Test Product" {
		t.Errorf("ProductName = %q, want Test Product", p.ProductName)
	}
	if p.Price != "150" {
		t.Errorf("Price = %q, want 150", p.Price)
	}
	if p.Sold != "42" {
		t.Errorf("Sold = %q, want 42", p.Sold)
	}
	if p.ShopeeItemID != "222" || p.ShopID != "111" {
		t.Errorf("ids = (%q, %q), want (111, 222)", p.ShopID, p.ShopeeItemID)
	}
	if p.Link != "https://shopee.co.id/product/111/222" {
		t.Errorf("Link = %q", p.Link)
	}
}

func TestParseSearchResponse_FallsBackToHistoricalSold(t *testing.T) {
	// A listing with no recent sales reports sold=0 but a historical count; the
	// historical value is the more useful number, so it is preferred over a
	// bare zero.
	body := []byte(`{"items":[{"item_basic":{"itemid":1,"shopid":2,"name":"X","sold":0,"historical_sold":77}}]}`)

	products, err := ParseSearchResponse(body, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}
	if products[0].Sold != "77" {
		t.Errorf("Sold = %q, want 77 (historical)", products[0].Sold)
	}
}

func TestParseSearchResponse_EmptyItemsIsNotAnError(t *testing.T) {
	// Reaching the end of results is normal; treating it as an error would abort
	// a scrape that has simply finished.
	products, err := ParseSearchResponse([]byte(`{"items":[]}`), "")
	if err != nil {
		t.Fatalf("empty items must not be an error, got: %v", err)
	}
	if len(products) != 0 {
		t.Errorf("got %d products, want 0", len(products))
	}
}

func TestParseSearchResponse_ShopeeErrorSurfaced(t *testing.T) {
	// Shopee reports errors inside a 200 body. Surfacing them gives an
	// actionable message instead of a silent empty result.
	body := []byte(`{"error":90309999,"error_msg":"rate limited"}`)

	_, err := ParseSearchResponse(body, "")
	if err == nil {
		t.Fatal("an in-body error must be surfaced, not treated as an empty result")
	}
}

func TestParseSearchResponse_MalformedBody(t *testing.T) {
	if _, err := ParseSearchResponse([]byte(`{not json`), ""); err == nil {
		t.Error("malformed JSON must return an error")
	}
	if _, err := ParseSearchResponse(nil, ""); err == nil {
		t.Error("an empty body must return an error")
	}
}

func TestParseSearchResponse_SkipsEmptyEntries(t *testing.T) {
	// Three junk shapes, all of which must be dropped:
	//   - all-zero ids (Shopee returns these for malformed entries)
	//   - zero ids but a valid-looking name
	//   - a valid id but no name
	// Storing a junk row with a fake link is worse than storing nothing, because
	// links are the dedup key.
	body := []byte(`{"items":[
      {"item_basic":{"itemid":0,"shopid":0,"name":""}},
      {"item_basic":{"itemid":0,"shopid":0,"name":"Zero IDs With A Name"}},
      {"item_basic":{"itemid":999,"shopid":888,"name":""}},
      {"item_basic":{"itemid":222,"shopid":111,"name":"Real"}}
    ]}`)

	products, err := ParseSearchResponse(body, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1 (junk entries skipped): %+v", len(products), products)
	}
	if products[0].ProductName != "Real" {
		t.Errorf("kept the wrong entry: %+v", products[0])
	}
	if products[0].Link != "https://shopee.co.id/product/111/222" {
		t.Errorf("Link = %q, want the canonical product URL", products[0].Link)
	}
}

// A zero id must never reach a link: every malformed entry would otherwise
// collapse onto the same fake "product/0/0" URL and destroy the dedup key.
func TestParseSearchResponse_NoZeroIDLinks(t *testing.T) {
	body := []byte(`{"items":[
      {"item_basic":{"itemid":0,"shopid":0,"name":"A"}},
      {"item_basic":{"itemid":5,"shopid":0,"name":"B"}},
      {"item_basic":{"itemid":0,"shopid":7,"name":"C"}}
    ]}`)

	products, err := ParseSearchResponse(body, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	for _, p := range products {
		if strings.Contains(p.Link, "/product/0/") || strings.HasSuffix(p.Link, "/0") {
			t.Errorf("product with a junk link was kept: %+v", p)
		}
	}
}

func TestParseSearchResponse_StringEncodedNumbers(t *testing.T) {
	// Some copies of the payload encode numbers as strings. json.Number handles
	// both, so the parse must not depend on which form arrives.
	body := []byte(`{"items":[{"item_basic":{"itemid":"222","shopid":"111","name":"S","price":"15000000"}}]}`)

	products, err := ParseSearchResponse(body, "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1", len(products))
	}
	if products[0].ShopeeItemID != "222" {
		t.Errorf("ShopeeItemID = %q, want 222", products[0].ShopeeItemID)
	}
	if products[0].Price != "150" {
		t.Errorf("Price = %q, want 150", products[0].Price)
	}
}
