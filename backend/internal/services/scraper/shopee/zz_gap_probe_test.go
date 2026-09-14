package shopee

import (
	"testing"
)

// Probes for gaps reported by an exhaustive path audit. Each asserts the
// CORRECT behaviour, so a failure is the evidence that the gap is real.

// GAP A: the multiplier is detected by substring search over the WHOLE string,
// so any incidental "rb" multiplies the count by 1000. Confirmed trace:
// "5 terjual (arb)" contains "rb" inside "arb".
func TestGap_RbSubstringFalsePositive(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		why  string
	}{
		{"5 terjual (arb)", "5", "'arb' contains 'rb' but is not a thousand marker"},
		{"5 terjual garbage", "5", "'garbage' contains 'rb'"},
		{"12 terjual", "12", "no unit word at all"},
		{"1,2rb terjual", "1200", "genuine rb marker"},
		{"5rb", "5000", "genuine rb marker, no space"},
	}

	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			got := CleanDOMSold(tc.raw)
			if got != tc.want {
				t.Errorf("CleanDOMSold(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

// GAP B: the count must be the LEADING number. A word-first rendering yields ""
// even though the count is present, which silently loses the value.
func TestGap_WordFirstSoldLosesCount(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		why  string
	}{
		{"Terjual 1,2rb", "1200", "the count is present, just not first"},
		{"terjual 500", "500", "the count is present, just not first"},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			if got := CleanDOMSold(tc.raw); got != tc.want {
				t.Errorf("CleanDOMSold(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

// GAP C (FIXED): a leading decimal separator with no integer part (",5rb"
// meaning 0.5rb = 500) was dropped, inflating the count 10x.
func TestGap_LeadingSeparatorInflates(t *testing.T) {
	if got := CleanDOMSold(",5rb terjual"); got != "500" {
		t.Errorf("CleanDOMSold(\",5rb terjual\") = %q, want 500 "+
			"(0.5 rb = 500; dropping the leading separator makes it 5 rb = 5000)", got)
	}
	// The same value written with an explicit zero must agree.
	if got := CleanDOMSold("0,5rb terjual"); got != "500" {
		t.Errorf("CleanDOMSold(\"0,5rb terjual\") = %q, want 500", got)
	}
}

// GAP D (FIXED): sold counts were not normalised for leading zeros, unlike
// prices, so "05" and "5" were stored differently and could not be compared.
func TestGap_SoldLeadingZerosNotStripped(t *testing.T) {
	if got := CleanDOMSold("05"); got != "5" {
		t.Errorf("CleanDOMSold(\"05\") = %q, want 5 (leading zeros must be normalised "+
			"like CleanDOMPrice does, or the stored value is not comparable)", got)
	}
}

// GAP E (FIXED): the network price path passed non-plain-number literals through
// unchanged, storing things like "1e5" as a price. Exponent and sign forms are
// now rejected rather than stored as text, and a decimal literal is scaled from
// its integer part.
//
// Note: 100 units of 1/100000 is 0.001, which is why "0100" -> "0.001". That is
// the documented scale, not a bug.
func TestGap_NonPlainPriceLiteralsNotStoredVerbatim(t *testing.T) {
	cases := []struct {
		raw  string
		want string
		why  string
	}{
		{"1e5", "", "exponent form is not a plain integer; storing \"1e5\" looks like data"},
		{"+1500000", "", "a sign prefix is not a price"},
		{"-1", "", "negative is Shopee's null-price sentinel"},
		{"1500000.00000", "15", "a decimal literal scales from its integer part"},
		{"15000000", "150", "plain integer path unchanged"},
		{"0100", "0.001", "100 units of 1/100000"},
		{"0", "0", "zero price"},
	}
	for _, tc := range cases {
		t.Run(tc.raw, func(t *testing.T) {
			if got := formatShopeePrice(tc.raw); got != tc.want {
				t.Errorf("formatShopeePrice(%q) = %q, want %q (%s)", tc.raw, got, tc.want, tc.why)
			}
		})
	}
}

// GAP F: one malformed item poisons the entire response, discarding every valid
// item alongside it.
func TestGap_OneBadItemDiscardsWholePage(t *testing.T) {
	// A single non-numeric itemid makes json.Number reject the body.
	body := []byte(`{"items":[
      {"item_basic":{"itemid":111,"shopid":1,"name":"Good One"}},
      {"item_basic":{"itemid":"not-a-number","shopid":2,"name":"Bad"}},
      {"item_basic":{"itemid":222,"shopid":3,"name":"Good Two"}}
    ]}`)

	products, err := ParseSearchResponse(body, "")
	if err != nil {
		t.Fatalf("one malformed item must not discard the whole page: %v", err)
	}
	if len(products) != 2 {
		t.Errorf("got %d products, want the 2 valid ones", len(products))
	}
}

// GAP G: unwrapEnvelope treats an explicit `"data": null` as a payload, so the
// original body (which may hold the real data) is discarded.
func TestGap_NullDataEnvelopeDiscardsSiblings(t *testing.T) {
	raw := []byte(`{"data":null,"responses":[{"status":200,"url":"u","body":"{}"}]}`)

	unwrapped := string(unwrapEnvelope(raw))
	if unwrapped == "null" {
		t.Errorf("unwrapEnvelope returned %q, discarding the sibling `responses` field: "+
			"an explicit null data must fall back to the original body", unwrapped)
	}
}

// GAP H: product mode ignores pagination, so every page navigates to the SAME
// URL and the loop re-captures it until maxPages.
func TestGap_ProductModeRepeatsSameURL(t *testing.T) {
	s := NewScraper(newFakeSender())
	cfg := Config{Mode: "product", ProductURL: "https://shopee.co.id/product/1/2"}

	first := s.pageURL(cfg, 1)
	second := s.pageURL(cfg, 2)
	if first == second {
		t.Errorf("pageURL returned the same URL for page 1 and page 2 (%q): product mode "+
			"will re-capture the same page until maxPages, doing N times the work", first)
	}
}

// GAP I: a missing shop/product URL is only detected after a browser tab has
// already been opened, so the failure point is late and wasteful.
func TestGap_ModeURLsNotValidatedUpfront(t *testing.T) {
	s := NewScraper(newFakeSender())

	// shop mode with no shop URL must be rejected before any tab is opened.
	if _, err := s.Run(t.Context(), Config{Mode: "shop"}); err == nil {
		t.Error("shop mode with no shop_url must be rejected")
	}
	if _, err := s.Run(t.Context(), Config{Mode: "product"}); err == nil {
		t.Error("product mode with no product_url must be rejected")
	}
}
