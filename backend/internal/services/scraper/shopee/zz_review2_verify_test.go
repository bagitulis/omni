package shopee

import (
	"context"
	"encoding/json"
	"testing"
)

// Verification of the second external review's claims. Each test reproduces the
// claim; a failure is the evidence.

// CLAIM 1: fromNetwork keeps products from an OLDER response when a NEWER
// response for the same page is validly empty. The loop walks responses
// newest-first and appends without stopping at the first usable page, so a stale
// page's rows can be attributed to the current page.
func TestClaimR1_StalePageMustNotBleedIntoCurrentPage(t *testing.T) {
	newerEmpty := `{"items":[]}`
	olderFull := `{"items":[{"item_basic":{"itemid":1,"shopid":10,"name":"Old","price":100000}}]}`

	reply := map[string]any{
		"success": true,
		"data": map[string]any{
			"responses": []map[string]any{
				{"status": 200, "url": "u?page=0", "body": olderFull},
				{"status": 200, "url": "u?page=1", "body": newerEmpty},
			},
		},
	}
	raw, _ := json.Marshal(reply)

	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{raw}

	s := NewScraper(f)
	products, err := s.fromNetwork(context.Background(), "")
	if err != nil {
		t.Fatalf("fromNetwork error: %v", err)
	}
	// The newest response describes the CURRENT page and says it is empty, so the
	// older page's rows must not be reported as this page's results.
	if len(products) != 0 {
		t.Errorf("got %d products from a stale response (%+v); the newest response for "+
			"this page was validly empty, so its rows would be stamped with the wrong "+
			"page number and would stop the empty-page counter from advancing",
			len(products), products)
	}
}

// CLAIM 2 (FIXED): a fractional item id was accepted and produced a fabricated
// URL like ".../product/2/1.5", which then persisted and deduped as if it named
// a real product.
func TestClaimR2_FractionalItemIDRejected(t *testing.T) {
	body := []byte(`{"items":[{"item_basic":{"itemid":1.5,"shopid":2,"name":"X","price":100000}}]}`)

	products, err := ParseSearchResponse(body, "")
	if len(products) != 0 {
		t.Fatalf("a fractional item id was accepted as a real product: %+v "+
			"(it would persist the fabricated URL %q)", products, products[0].Link)
	}
	// Bad identifiers are surfaced rather than silently swallowed: a page of them
	// is a real problem the operator should see.
	if err == nil {
		t.Error("a page whose only item has a malformed identifier must be reported, " +
			"not returned as a genuinely empty page")
	}
}

// CLAIM 3: a canonical product link carrying a query string loses its ids, so the
// row is dropped even though the product is perfectly identifiable.
func TestClaimR3_CanonicalLinkWithQueryKeepsProduct(t *testing.T) {
	raw := []byte(`[{"name":"X","price":"Rp15.000","sold":"1",
      "link":"/product/111/222?sp_atk=abc","image_url":""}]`)

	products, err := ParseDOMProducts(raw, "https://shopee.co.id")
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if len(products) != 1 {
		t.Fatalf("got %d products, want 1: a canonical link with a query string is "+
			"still a valid product", len(products))
	}
	if products[0].ShopID != "111" || products[0].ShopeeItemID != "222" {
		t.Errorf("ids = (%q, %q), want (111, 222)",
			products[0].ShopID, products[0].ShopeeItemID)
	}
}

// CLAIM 5: a negative DOM price is turned into a positive one.
func TestClaimR5_NegativePriceNotTurnedPositive(t *testing.T) {
	cases := []string{"-Rp15.000", "-15.000"}
	for _, raw := range cases {
		if got := CleanDOMPrice(raw); got != "" {
			t.Errorf("CleanDOMPrice(%q) = %q, want empty: a negative price is not a "+
				"price, and returning a positive number silently invents one", raw, got)
		}
	}
}

// CLAIM 6: a number that appears BEFORE the sold count is picked up in its place.
func TestClaimR6_CountBeforeSoldLabelNotMisread(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		why  string
	}{
		{"Baru, terjual 12", "12", "the comma after 'Baru' is punctuation, not a decimal point"},
		{"Promo 20% terjual 5", "5", "a discount percentage is not the sold count"},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			if got := CleanDOMSold(tc.raw); got != tc.want {
				t.Errorf("CleanDOMSold(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

// CLAIM 7: an item with the wrong STRUCTURE yields an unusable row but does not
// count as skipped, so the "all items malformed" guard never fires and the page
// looks genuinely empty.
func TestClaimR7_WrongStructureDetectedAsShapeChange(t *testing.T) {
	// itemid/shopid/name at the wrong level: valid JSON, wrong shape.
	body := []byte(`{"items":[{"itemid":1,"shopid":2,"name":"X","price":100000}]}`)

	products, err := ParseSearchResponse(body, "")
	if err == nil && len(products) == 0 {
		t.Error("an item with the wrong structure produced an empty result with no error, " +
			"so a shape change looks like a genuinely empty page and three such pages " +
			"end the scrape early")
	}
}
