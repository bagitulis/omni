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

// normalizeDOMLink makes a card's href absolute AND canonical.
//
// Canonicalising matters as much as absolutising: links are the dedup key and
// the database's unique key, and the two capture paths see the same product
// differently. The DOM path gets a slug with tracking parameters
// (".../Name-i.111.222?sp_atk=abc") while the network path builds
// ".../product/111/222". Left alone, one product would be stored twice whenever a
// run mixes capture paths across pages.
//
// When the shop and item ids cannot be recovered the link is only absolutised:
// keeping an imperfect link is better than dropping the product.
func normalizeDOMLink(link, baseURL string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}

	absolute := link
	if !strings.HasPrefix(absolute, "http://") && !strings.HasPrefix(absolute, "https://") {
		if baseURL == "" {
			baseURL = "https://shopee.co.id"
		}
		if !strings.HasPrefix(absolute, "/") {
			absolute = "/" + absolute
		}
		absolute = strings.TrimRight(baseURL, "/") + absolute
	}

	// Collapse to the canonical form when both ids are present.
	if shopID, itemID := ExtractShopIDs(absolute); shopID != "" && itemID != "" {
		origin := baseURL
		if origin == "" {
			origin = originOf(absolute)
		}
		if canonical := BuildProductURL(origin, shopID, itemID); canonical != "" {
			return canonical
		}
	}

	// Otherwise at least drop the fragment, which never identifies a product.
	if idx := strings.IndexByte(absolute, '#'); idx >= 0 {
		absolute = absolute[:idx]
	}
	return absolute
}

// originOf returns the scheme and host of an absolute URL, or "" when it cannot
// be determined.
func originOf(rawURL string) string {
	schemeEnd := strings.Index(rawURL, "://")
	if schemeEnd < 0 {
		return ""
	}
	rest := rawURL[schemeEnd+3:]
	if slash := strings.IndexByte(rest, '/'); slash >= 0 {
		rest = rest[:slash]
	}
	if rest == "" {
		return ""
	}
	return rawURL[:schemeEnd+3] + rest
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

	// A negative amount is not a price we can use, and dropping the sign would
	// invent a positive one: "-15.000" must be rejected, not stored as "15000".
	// A minus AFTER digits is a range separator ("15.000 - 25.000"), where the
	// leading value is the one to keep.
	if idx := strings.IndexByte(raw, '-'); idx >= 0 && countDigitsBefore(raw, '-') == 0 {
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
		case r == '~':
			// A range marker: the leading value is the price to keep.
			if b.Len() > 0 {
				return normalizeDigits(b.String())
			}
		case r == '-':
			// A range separator. Everything after it is the high bound, which is
			// not the price to store — returning here is what stops
			// "Rp10.000 - Rp20.000" from concatenating into "1000020000".
			if b.Len() > 0 {
				return normalizeDigits(b.String())
			}
		}
	}

	digits := b.String()
	if digits == "" {
		return ""
	}
	return normalizeDigits(digits)
}

// countDigitsBefore counts digits appearing before the first occurrence of sep.
func countDigitsBefore(s string, sep rune) int {
	n := 0
	for _, r := range s {
		if r == sep {
			return n
		}
		if r >= '0' && r <= '9' {
			n++
		}
	}
	return n
}

// soldLabelHints are words that mark a sold count. When one is present, the
// number nearest to it is the count — not merely the first number in the string.
var soldLabelHints = []string{"terjual", "sold", "terkirim", "dibeli"}

// CleanDOMSold extracts the numeric part of a sold-count string.
//
// Shopee renders these as "1,2rb terjual" (Indonesian for "1.2k sold"). The unit
// word is expanded because storing "1,2rb" would make sorting by sold count
// useless.
//
// Three properties matter for correctness:
//
//   - The unit must be the marker IMMEDIATELY after the number. Searching the
//     whole string for "rb" matched any word containing those letters
//     ("garbage", "arb"), inflating a count by 1000x.
//   - When a sold indicator is present ("terjual"), the number NEAREST to it is
//     the count. Taking the first number read a discount as the count:
//     "Promo 20% terjual 5" became 20 instead of 5.
//   - A separator with no integer part before it is a decimal only when it sits
//     directly against digits; otherwise it is punctuation. "Baru, terjual 12"
//     must yield 12, not a spurious 0.
func CleanDOMSold(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}

	region := trimmed
	// Prefer the segment after the sold label, so an earlier number (a discount,
	// a rating) is not mistaken for the count.
	if idx := indexOfSoldLabel(trimmed); idx >= 0 {
		region = trimmed[idx:]
	}

	start := numericRunStart(region)
	if start < 0 {
		// No number in the sold region; fall back to anywhere in the string so a
		// layout we did not anticipate still yields a count.
		region = trimmed
		start = numericRunStart(region)
	}
	if start < 0 {
		return ""
	}

	number, rest := splitLeadingNumber(region[start:])
	if number == "" {
		return ""
	}

	multiplier := unitMultiplier(rest)
	return expandAbbreviated(number, multiplier)
}

// indexOfSoldLabel returns the offset just past the first sold indicator, or -1.
func indexOfSoldLabel(s string) int {
	lower := strings.ToLower(s)
	best := -1
	for _, hint := range soldLabelHints {
		if i := strings.Index(lower, hint); i >= 0 {
			end := i + len(hint)
			if best == -1 || end < best {
				best = end
			}
		}
	}
	return best
}

// numericRunStart returns the index of the first character that can begin a
// number, or -1.
//
// A separator only counts as the start of a number when a digit follows it
// directly. Otherwise it is punctuation: treating the comma in "Baru, terjual 12"
// as a decimal point produced a spurious "0".
func numericRunStart(s string) int {
	runes := []rune(s)
	for i, r := range runes {
		if r >= '0' && r <= '9' {
			return len(string(runes[:i]))
		}
		if r == ',' || r == '.' {
			if i+1 < len(runes) && runes[i+1] >= '0' && runes[i+1] <= '9' {
				return len(string(runes[:i]))
			}
		}
	}
	return -1
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
