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
// Shopee renders these as "1,2rb terjual" (Indonesian for "1.2k sold"). The unit
// word is expanded because storing "1,2rb" would make sorting by sold count
// useless.
//
// Two properties matter for correctness:
//
//   - The unit must be the marker IMMEDIATELY after the number. Searching the
//     whole string for "rb" matched any word containing those letters
//     ("garbage", "arb"), inflating a count by 1000x.
//   - The count need not be the first thing in the string. Requiring a leading
//     digit returned "" for "Terjual 1,2rb", silently losing the value.
func CleanDOMSold(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	// Find the start of the numeric run. A leading separator is part of the
	// number (",5rb" means 0.5rb = 500), so it must not be skipped: starting at
	// the first digit would read ",5rb" as "5rb" and inflate it 10x.
	start := -1
	for i, r := range trimmed {
		if (r >= '0' && r <= '9') || r == ',' || r == '.' {
			start = i
			break
		}
	}
	if start < 0 {
		// A word-only value ("terjual") carries no count.
		return ""
	}

	number, rest := splitLeadingNumber(trimmed[start:])
	if number == "" {
		return ""
	}

	multiplier := unitMultiplier(rest)
	return expandAbbreviated(number, multiplier)
}

// splitLeadingNumber splits a numeric run from the text that follows it.
//
// A separator between digits is kept as a decimal point; two separators in a row
// end the run, since that is grouping rather than a number. The returned unit
// text is what immediately follows the digits, which is what determines the
// multiplier.
func splitLeadingNumber(s string) (number, rest string) {
	var num strings.Builder
	sawSeparator := false

	for i, r := range s {
		switch {
		case r >= '0' && r <= '9':
			num.WriteRune(r)
			sawSeparator = false
		case r == ',' || r == '.':
			if sawSeparator {
				// Two separators in a row: the number ended before the second.
				return strings.TrimSuffix(num.String(), "."), s[i:]
			}
			// A leading separator with no integer part is 0.x, so make the
			// absent integer part explicit. Dropping it turned ",5rb" (500)
			// into "5rb" (5000), a 10x overstatement.
			if num.Len() == 0 {
				num.WriteRune('0')
			}
			num.WriteRune('.')
			sawSeparator = true
		default:
			return strings.TrimSuffix(num.String(), "."), s[i:]
		}
	}

	return strings.TrimSuffix(num.String(), "."), ""
}

// unitMultiplier maps the unit marker following a count to its multiplier.
//
// Only the leading token is examined, and only as a prefix, so "rb" must START
// the unit text. That is what stops "garbage" from being read as a thousands
// marker while still accepting "rb terjual" and "rbterjual".
func unitMultiplier(rest string) int64 {
	// Take the first token: everything up to the first non-letter.
	token := strings.ToLower(strings.TrimLeft(rest, " \t\u00a0"))
	end := 0
	for i, r := range token {
		if (r >= 'a' && r <= 'z') || r == '+' {
			end = i + len(string(r))
			continue
		}
		break
	}
	token = token[:end]

	switch {
	case strings.HasPrefix(token, "rb"):
		return 1000
	case strings.HasPrefix(token, "jt"), strings.HasPrefix(token, "mio"):
		return 1000000
	default:
		return 1
	}
}

// expandAbbreviated turns a possibly fractional count into a whole number.
//
// Parsed with strconv rather than a hand-rolled accumulator, and rounded to
// the nearest whole count: a sold count is a user-visible integer, so "1,2rb"
// becomes 1200 rather than a truncated 1000 or a fractional value.
func expandAbbreviated(digits string, multiplier int64) string {
	if multiplier == 1 {
		// No unit word: drop the grouping separators to yield an integer, then
		// normalise leading zeros so "05" and "5" compare equal. Without this
		// the stored value is not comparable across rows.
		return normalizeDigits(strings.ReplaceAll(digits, ".", ""))
	}

	value, err := strconv.ParseFloat(digits, 64)
	if err != nil {
		// Unparseable but numeric-looking: keep the digits rather than lose data.
		return normalizeDigits(strings.ReplaceAll(digits, ".", ""))
	}

	scaled := value * float64(multiplier)
	if scaled >= float64(1<<62) {
		// Guard against an absurd value overflowing the conversion below.
		return normalizeDigits(strings.ReplaceAll(digits, ".", ""))
	}
	return strconv.FormatInt(int64(math.Round(scaled)), 10)
}

// normalizeDigits strips leading zeros, keeping a lone "0" for zero itself.
func normalizeDigits(s string) string {
	s = strings.TrimLeft(s, "0")
	if s == "" {
		return "0"
	}
	return s
}
