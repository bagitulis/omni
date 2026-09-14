package shopee

import (
	"encoding/json"
	"math"
	"strconv"
	"strings"
)

// DOM fallback parsing.
//
// Network capture is the primary path. This exists because Shopee's internal
// endpoints are undocumented and change without notice, and because a page that
// blocks or rewrites the JSON API still renders product cards. Without a
// fallback, a single API change would break scraping entirely.
//
// The content script does the DOM traversal (it has the live document); this
// file parses what it returns. Keeping the parse here means it is unit-testable
// without a browser.

// DOMProduct is one card as reported by the content script.
//
// Field names mirror the JSON the content script emits.
type DOMProduct struct {
	Name     string `json:"name"`
	Price    string `json:"price"`
	Sold     string `json:"sold"`
	Link     string `json:"link"`
	ImageURL string `json:"image_url"`
}

// ParseDOMProducts converts the content script's extraction into ParsedProduct.
//
// Prices from the DOM are already formatted ("Rp15.000"), so they are cleaned
// rather than rescaled — the two paths must agree on the stored representation,
// which is why cleaning is centralised here.
func ParseDOMProducts(raw []byte, baseURL string) ([]ParsedProduct, error) {
	if len(raw) == 0 {
		return nil, nil
	}

	// The content script may return either a bare array or an envelope with
	// data; accept both so a wrapper change does not silently yield nothing.
	items, err := decodeDOMItems(raw)
	if err != nil {
		return nil, err
	}

	out := make([]ParsedProduct, 0, len(items))
	for _, it := range items {
		link := normalizeDOMLink(it.Link, baseURL)
		shopID, itemID := ExtractShopIDs(link)

		p := ParsedProduct{
			ProductName:  strings.TrimSpace(it.Name),
			Price:        CleanDOMPrice(it.Price),
			Sold:         CleanDOMSold(it.Sold),
			Link:         link,
			ImageURL:     strings.TrimSpace(it.ImageURL),
			ShopeeItemID: itemID,
			ShopID:       shopID,
		}
		if !isUsableProduct(p) {
			continue
		}
		out = append(out, p)
	}
	return out, nil
}

// decodeDOMItems accepts either a bare array or a {data: [...]} envelope.
func decodeDOMItems(raw []byte) ([]DOMProduct, error) {
	var direct []DOMProduct
	if err := json.Unmarshal(raw, &direct); err == nil {
		return direct, nil
	}

	var envelope struct {
		Data []DOMProduct `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

// normalizeDOMLink makes a card's href absolute.
//
// The content script usually absolutises already, but a relative href would
// otherwise be stored and then fail to match the canonical form used by the
// network path — breaking dedup between the two capture paths.
func normalizeDOMLink(link, baseURL string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	if strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://") {
		return link
	}
	if baseURL == "" {
		baseURL = "https://shopee.co.id"
	}
	if !strings.HasPrefix(link, "/") {
		link = "/" + link
	}
	return strings.TrimRight(baseURL, "/") + link
}

// CleanDOMPrice normalises a formatted price string for storage.
//
// The digits are kept with their separators removed, matching the network path's
// plain numeric form. Currency symbols and spacing are dropped so the two paths
// produce comparable values; without this, "Rp15.000" and "150" would look like
// different products' prices for the same amount.
func CleanDOMPrice(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	var b strings.Builder
	for _, r := range raw {
		switch {
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '.' || r == ',':
			// Separators are dropped: Shopee mixes "." and "," between locales,
			// so preserving either would be inconsistent.
			continue
		case r == '-' || r == '~':
			// A range or a negative marker means the value is not a single
			// price; keep the leading digits only.
			if b.Len() > 0 {
				return b.String()
			}
		}
	}

	digits := b.String()
	if digits == "" {
		return ""
	}
	// Strip leading zeros so "015000" and "15000" compare equal.
	trimmed := strings.TrimLeft(digits, "0")
	if trimmed == "" {
		return "0"
	}
	return trimmed
}

// CleanDOMSold extracts the numeric part of a sold-count string.
//
// Shopee renders these as "1,2rb terjual" (Indonesian for "1.2k sold"), so the
// leading number is kept and the unit word dropped. Abbreviations are expanded
// because storing "1,2rb" would make sorting by sold count useless.
func CleanDOMSold(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	lower := strings.ToLower(raw)

	// Collect the leading numeric run. A separator between digits is a decimal
	// point; one that ends the run is not.
	var num strings.Builder
	sawSeparator := false
	for _, r := range lower {
		switch {
		case r >= '0' && r <= '9':
			num.WriteRune(r)
			sawSeparator = false
		case r == ',' || r == '.':
			// A leading separator is not part of a number.
			if num.Len() == 0 {
				continue
			}
			// Two separators in a row ends the numeric run.
			if sawSeparator {
				goto parsed
			}
			num.WriteRune('.')
			sawSeparator = true
		default:
			goto parsed
		}
	}

parsed:
	digits := strings.TrimSuffix(num.String(), ".")
	if digits == "" {
		// A word-only value ("terjual") carries no count.
		return ""
	}

	// Determine the multiplier from the unit that follows the number.
	multiplier := int64(1)
	switch {
	case strings.Contains(lower, "rb"):
		multiplier = 1000
	case strings.Contains(lower, "jt"), strings.Contains(lower, "mio"):
		multiplier = 1000000
	}

	return expandAbbreviated(digits, multiplier)
}

// expandAbbreviated turns a possibly fractional count into a whole number.
//
// Parsed with strconv rather than a hand-rolled accumulator, and rounded to
// the nearest whole count: a sold count is a user-visible integer, so "1,2rb"
// becomes 1200 rather than a truncated 1000 or a fractional value.
func expandAbbreviated(digits string, multiplier int64) string {
	if multiplier == 1 {
		// No unit word: drop the decimal separator to yield an integer.
		return strings.ReplaceAll(digits, ".", "")
	}

	value, err := strconv.ParseFloat(digits, 64)
	if err != nil {
		// Unparseable but numeric-looking: keep the digits rather than lose data.
		return strings.ReplaceAll(digits, ".", "")
	}

	scaled := value * float64(multiplier)
	if scaled >= float64(1<<62) {
		// Guard against an absurd value overflowing the conversion below.
		return strings.ReplaceAll(digits, ".", "")
	}
	return strconv.FormatInt(int64(math.Round(scaled)), 10)
}
