package shopee

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// Parsing of Shopee's own search/shop API responses.
//
// Network capture is the PRIMARY path: Shopee returns structured JSON, which is
// far more robust than DOM selectors and does not depend on class names that
// change without notice. DOM extraction is the fallback (see product_dom.go);
// this file handles the structured case.
//
// The payload shape below is taken from Shopee's public search_items endpoint.
// It is undocumented and can change, so every field is read defensively: a
// missing field yields an empty value rather than an error, and only a
// completely unparseable body is a failure.

// searchResponse is the envelope of a Shopee search response.
type searchResponse struct {
	Items    []searchItem `json:"items"`
	Error    *int         `json:"error"`
	ErrorMsg string       `json:"error_msg"`
}

// searchItem is one product card inside a search response.
type searchItem struct {
	ItemBasic itemBasic `json:"item_basic"`
}

// itemBasic holds the fields we collect.
//
// Numeric fields arrive as ints in copies of the payload but as strings in
// others, so the ones we persist as text are declared as strings and converted
// explicitly.
type itemBasic struct {
	ItemID         json.Number `json:"itemid"`
	ShopID         json.Number `json:"shopid"`
	Name           string      `json:"name"`
	Price          json.Number `json:"price"`
	PriceMin       json.Number `json:"price_min"`
	Sold           json.Number `json:"sold"`
	HistoricalSold json.Number `json:"historical_sold"`
	Image          string      `json:"image"`
	ShopName       string      `json:"shop_name"`
	Stock          json.Number `json:"stock"`
	CTime          json.Number `json:"ctime"`
}

// ParsedProduct is one product extracted from a network payload.
type ParsedProduct struct {
	ProductName  string
	Price        string
	Sold         string
	Link         string
	ImageURL     string
	ShopeeItemID string
	ShopID       string

	// Page is the 1-based page the product was captured from. Recorded so a
	// partial run can be resumed and so operators can see where coverage stopped.
	Page int

	// Source records which capture path produced this row: ScrapeSourceNetwork or
	// ScrapeSourceDOM. Exposed so it is visible when the fallback is doing the
	// work in production, which would otherwise go unnoticed.
	Source string
}

// ParseSearchResponse extracts products from a Shopee search_items body.
//
// Returns an empty slice (not an error) when the body is valid JSON but contains
// no items: an empty result page is a normal end-of-results signal, and treating
// it as an error would abort a scrape that has simply reached the end.
//
// Items are decoded individually. Decoding the whole array at once means one
// malformed entry — a non-numeric itemid, a boolean where a number belongs —
// fails the entire unmarshal and discards every valid product on the page. That
// turns a single bad card into a lost page.
func ParseSearchResponse(body []byte, baseURL string) ([]ParsedProduct, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	// Decode the envelope, keeping items raw so each can be parsed alone.
	var envelope struct {
		Items    []json.RawMessage `json:"items"`
		Error    *int              `json:"error"`
		ErrorMsg string            `json:"error_msg"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		// A body that is not even the expected envelope shape is a real error:
		// the caller may be looking at an HTML error page or a changed API.
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	// Shopee reports its own errors inside a 200 body; surfacing them here gives
	// an actionable message instead of an empty result.
	if envelope.Error != nil && *envelope.Error != 0 {
		return nil, fmt.Errorf("shopee returned error %d: %s", *envelope.Error, envelope.ErrorMsg)
	}

	out := make([]ParsedProduct, 0, len(envelope.Items))
	skipped := 0
	for _, rawItem := range envelope.Items {
		var item searchItem
		if err := json.Unmarshal(rawItem, &item); err != nil {
			// One malformed card must not cost the whole page. Counted so a
			// systemic change (all items malformed) is visible rather than
			// looking like a genuinely empty result.
			skipped++
			continue
		}
		p := parseItemBasic(item.ItemBasic, baseURL)
		if !isUsableProduct(p) {
			continue
		}
		out = append(out, p)
	}

	if len(out) == 0 && skipped > 0 {
		// Every item failed to parse: that is a shape change, not an empty page,
		// and reporting it as empty would silently stop the scrape.
		return nil, fmt.Errorf("all %d items failed to parse; the response shape may have changed", skipped)
	}

	return out, nil
}

// isUsableProduct reports whether a parsed entry is worth persisting.
//
// A row is dropped when it has no name, no usable item id, or no link. The id
// checks matter: Shopee returns all-zero ids for malformed entries, and those
// would otherwise build a junk "https://.../product/0/0" link. Storing a fake
// URL is worse than storing nothing, because links are the dedup key — every
// junk row would share that one URL and collapse into a single misleading row.
func isUsableProduct(p ParsedProduct) bool {
	if p.ProductName == "" || p.Link == "" {
		return false
	}
	if p.ShopeeItemID == "" || p.ShopeeItemID == "0" {
		return false
	}
	return p.ShopID != "" && p.ShopID != "0"
}

// parseItemBasic converts one API item into our representation.
func parseItemBasic(b itemBasic, baseURL string) ParsedProduct {
	itemID := b.ItemID.String()
	shopID := b.ShopID.String()

	// Shopee's price field is in units of 1/100000 of the currency. Converting
	// here keeps the persisted value comparable with the DOM path, which reads a
	// formatted price string.
	price := formatShopeePrice(b.Price.String())

	sold := b.Sold.String()
	if sold == "" || sold == "0" {
		sold = b.HistoricalSold.String()
	}

	return ParsedProduct{
		ProductName:  strings.TrimSpace(b.Name),
		Price:        price,
		Sold:         sold,
		Link:         BuildProductURL(baseURL, shopID, itemID),
		ImageURL:     b.Image,
		ShopeeItemID: itemID,
		ShopID:       shopID,
	}
}

// productURLRe matches the -i.{shopID}.{itemID} segment of a Shopee product URL.
var productURLRe = regexp.MustCompile(`-i\.(\d+)\.(\d+)`)

// BuildProductURL constructs a canonical product URL.
//
// The canonical /product/{shopID}/{itemID} form is used rather than the
// descriptive slug, so the same product produces an identical link regardless of
// which capture path found it. That matters because links are the dedup key.
func BuildProductURL(baseURL, shopID, itemID string) string {
	if shopID == "" || itemID == "" {
		return ""
	}
	if baseURL == "" {
		baseURL = "https://shopee.co.id"
	}
	baseURL = strings.TrimRight(baseURL, "/")
	return fmt.Sprintf("%s/product/%s/%s", baseURL, shopID, itemID)
}

// ExtractShopIDs pulls shop and item IDs out of a product URL.
//
// Used by the DOM fallback, where the only identifier available is the link
// itself. Returns empty strings when the URL does not match, which callers treat
// as "keep the product but without an ID" rather than dropping it.
func ExtractShopIDs(link string) (shopID, itemID string) {
	if link == "" {
		return "", ""
	}
	m := productURLRe.FindStringSubmatch(link)
	if len(m) < 3 {
		// The canonical /product/{shop}/{item} form has no -i. segment.
		parts := strings.Split(strings.Trim(link, "/"), "/")
		for i := 0; i+1 < len(parts); i++ {
			if parts[i] == "product" {
				if isNumeric(parts[i+1]) && i+2 < len(parts) && isNumeric(parts[i+2]) {
					return parts[i+1], parts[i+2]
				}
			}
		}
		return "", ""
	}
	return m[1], m[2]
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// isPlainInteger reports whether s is a run of ASCII digits only: no sign, no
// decimal point, no exponent, no whitespace.
//
// Used before scaling a price. A JSON number may legally be written as "1e5" or
// "1500000.00000"; passing either through the integer path would store the
// literal text as a price ("1e5"), which is not a price at all.
func isPlainInteger(s string) bool {
	if s == "" {
		return false
	}
	return isNumeric(s)
}

// formatShopeePrice converts Shopee's integer price (1/100000 units) to a
// decimal string.
//
// Done with integer arithmetic rather than float, because float64 cannot exactly
// represent values like 15000000 and would introduce rounding into a field users
// compare across rows.
//
// Only a plain digit run is scaled. Anything else — a sign prefix, a decimal
// point, exponent notation — is not the integer this function expects and is
// rejected rather than stored verbatim: writing "1e5" into a price column
// produces a value that looks like data but cannot be compared or summed.
func formatShopeePrice(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	if !isPlainInteger(raw) {
		// A negative value is Shopee's sentinel for a null price.
		if strings.HasPrefix(raw, "-") {
			return ""
		}
		// A decimal literal (e.g. "1500000.00000") still denotes a real amount,
		// so scale its integer part rather than discarding the price.
		if whole, _, found := strings.Cut(raw, "."); found && isPlainInteger(whole) {
			return formatShopeePrice(whole)
		}
		// Exponent form or any other shape: not a price we can trust. Returning
		// empty is honest; storing the literal would look like data.
		return ""
	}

	// Normalise leading zeros BEFORE scaling. Padding first would change the
	// value: "0100" padded to six digits becomes "000100", which reads as
	// 0.001 rather than 100/100000 = 0.001 ... precisely the kind of silent
	// magnitude error this function exists to avoid. Stripping the zeros first
	// makes the digit count meaningful.
	raw = strings.TrimLeft(raw, "0")
	if raw == "" {
		// The value was all zeros: a zero price.
		return "0"
	}

	// Pad to at least 6 digits so the fractional part always exists.
	padded := raw
	for len(padded) < 6 {
		padded = "0" + padded
	}
	whole := padded[:len(padded)-5]
	frac := padded[len(padded)-5:]

	// Drop trailing zeros in the fraction; if nothing is left, the price was a
	// whole number.
	frac = strings.TrimRight(frac, "0")
	if frac == "" {
		return whole
	}
	return whole + "." + frac
}
