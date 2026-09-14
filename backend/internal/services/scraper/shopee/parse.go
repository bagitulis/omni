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
}

// ParseSearchResponse extracts products from a Shopee search_items body.
//
// Returns an empty slice (not an error) when the body is valid JSON but contains
// no items: an empty result page is a normal end-of-results signal, and treating
// it as an error would abort a scrape that has simply reached the end.
func ParseSearchResponse(body []byte, baseURL string) ([]ParsedProduct, error) {
	if len(body) == 0 {
		return nil, fmt.Errorf("empty response body")
	}

	var resp searchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("parse search response: %w", err)
	}

	// Shopee reports its own errors inside a 200 body; surfacing them here gives
	// an actionable message instead of an empty result.
	if resp.Error != nil && *resp.Error != 0 {
		return nil, fmt.Errorf("shopee returned error %d: %s", *resp.Error, resp.ErrorMsg)
	}

	out := make([]ParsedProduct, 0, len(resp.Items))
	for _, it := range resp.Items {
		p := parseItemBasic(it.ItemBasic, baseURL)
		if !isUsableProduct(p) {
			continue
		}
		out = append(out, p)
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

// formatShopeePrice converts Shopee's integer price (1/100000 units) to a
// decimal string.
//
// Done with integer arithmetic rather than float, because float64 cannot exactly
// represent values like 15000000 and would introduce rounding into a field users
// compare across rows.
func formatShopeePrice(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if !isNumeric(raw) {
		// Already formatted (for example "Rp15.000"), or negative for a null price.
		if strings.HasPrefix(raw, "-") {
			return ""
		}
		return raw
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
