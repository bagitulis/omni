package shopee

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/omni/backend/internal/models"
	"strings"
)

// Scrape orchestration: drives the connected browser through a scrape and
// converts its replies into ParsedProducts.
//
// This file deliberately contains no database or WebSocket code. It depends on a
// narrow CommandSender interface so the page loop, capture strategy, and stopping
// rules can be unit-tested with a scripted fake — which is the only way to test
// them here, since the real transport needs a browser.

// CommandSender sends a command to the paired extension and returns its result.
//
// Declared as an interface rather than taking the hub directly so the page loop
// can be tested against a scripted sequence of replies, including the failure
// shapes (disconnect, timeout) that are hard to provoke against a real browser.
type CommandSender interface {
	// Send runs one command. payload is JSON. Returns the result payload, or an
	// error when the extension is unreachable or the command failed.
	Send(ctx context.Context, action string, payload any) (json.RawMessage, error)
}

// Config controls one scrape run.
type Config struct {
	Mode        string // search | shop | product
	Query       string // search mode
	ShopURL     string // shop mode
	ProductURL  string // product mode
	MaxPages    int
	MaxProducts int
	BaseURL     string

	// OnPage is called after each page is captured, for progress reporting. It
	// must not block. page is 1-based; captured is the running total.
	OnPage func(page, captured int)
}

// Result is the outcome of a scrape run.
type Result struct {
	Products []ParsedProduct
	Pages    int
	// Source records which capture path produced the data: "network", "dom", or
	// "" when nothing was captured. Exposed so an operator can tell whether the
	// preferred path is working in production.
	Source string
}

// Defaults and limits.
const (
	defaultMaxPages    = 50
	defaultMaxProducts = 5000
	// Pages with no results end the run. One empty page is often a slow network;
	// several in a row means the results are exhausted.
	emptyPagesBeforeStop = 3
	// A sanity cap on products from a single page, so a broken selector that
	// matches every element cannot flood the database.
	maxProductsPerPage = 500
)

// Scraper runs Shopee scrapes through a CommandSender.
type Scraper struct {
	sender CommandSender
}

// NewScraper creates a Scraper.
func NewScraper(sender CommandSender) *Scraper {
	return &Scraper{sender: sender}
}

// Run executes a scrape for the configured mode.
func (s *Scraper) Run(ctx context.Context, cfg Config) (*Result, error) {
	if s.sender == nil {
		return nil, fmt.Errorf("scraper: no command sender configured")
	}

	// Validate the mode's required input BEFORE doing any work. Catching a
	// missing URL here means the caller never opens a browser tab only to fail
	// on navigation, and the error names the actual problem.
	switch cfg.Mode {
	case "search", "":
		if strings.TrimSpace(cfg.Query) == "" {
			return nil, fmt.Errorf("scraper: query is required for search mode")
		}
	case "shop":
		if strings.TrimSpace(cfg.ShopURL) == "" {
			return nil, fmt.Errorf("scraper: shop_url is required for shop mode")
		}
	case "product":
		if strings.TrimSpace(cfg.ProductURL) == "" {
			return nil, fmt.Errorf("scraper: product_url is required for product mode")
		}
	default:
		return nil, fmt.Errorf("scraper: unsupported mode %q", cfg.Mode)
	}

	return s.runPaginated(ctx, cfg)
}

// runPaginated drives the page loop: navigate, capture, extract, repeat.
func (s *Scraper) runPaginated(ctx context.Context, cfg Config) (*Result, error) {
	maxPages := cfg.MaxPages
	if maxPages <= 0 {
		maxPages = defaultMaxPages
	}
	maxProducts := cfg.MaxProducts
	if maxProducts <= 0 {
		maxProducts = defaultMaxProducts
	}
	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://shopee.co.id"
	}

	res := &Result{}
	seen := make(map[string]bool)
	consecutiveEmpty := 0

	for page := 1; page <= maxPages; page++ {
		if err := ctx.Err(); err != nil {
			// A cancelled context is a normal stop (the operator pressed stop),
			// so partial results are returned rather than discarded.
			break
		}
		if len(res.Products) >= maxProducts {
			break
		}

		pageProducts, err := s.capturePage(ctx, cfg, baseURL, page)
		if err != nil {
			if len(res.Products) > 0 {
				// Fail soft once we have something: a failure on a later page
				// must not throw away the pages already collected.
				break
			}
			return nil, err
		}
		res.Pages = page

		if len(pageProducts) == 0 {
			consecutiveEmpty++
			// Report the page even though it yielded nothing. Without this a
			// caller watching progress sees nothing at all during the empty
			// pages that precede a stop, and cannot tell a slow page from a
			// stalled scrape.
			if cfg.OnPage != nil {
				cfg.OnPage(page, len(res.Products))
			}
			if consecutiveEmpty >= emptyPagesBeforeStop {
				break
			}
			// Try the next page rather than stopping on the first empty one.
			if page < maxPages && !s.isLastPage(ctx) {
				if s.pageURL(cfg, page+1) == "" {
					break
				}
				if nextErr := s.nextPage(ctx); nextErr != nil {
					break
				}
				continue
			}
			break
		}
		consecutiveEmpty = 0

		// Deduplicate within the run: Shopee repeats cards across pages, and the
		// same product appearing twice would inflate counts and waste rows.
		for _, p := range pageProducts {
			if seen[p.Link] {
				continue
			}
			seen[p.Link] = true
			// Stamp provenance here rather than in each capture path, so both
			// paths are attributed consistently and the page number cannot be
			// forgotten in one of them.
			p.Page = page
			res.Products = append(res.Products, p)
			res.Source = p.Source
			if len(res.Products) >= maxProducts {
				break
			}
		}

		if cfg.OnPage != nil {
			cfg.OnPage(page, len(res.Products))
		}

		if len(res.Products) >= maxProducts {
			break
		}
		// Stop once the page reported itself as last.
		if s.isLastPage(ctx) {
			break
		}
		if page < maxPages {
			// A mode with no further pages (product detail) yields no URL, so
			// there is nothing more to capture. Without this the loop would
			// re-navigate to the same page until maxPages.
			if s.pageURL(cfg, page+1) == "" {
				break
			}
			if nextErr := s.nextPage(ctx); nextErr != nil {
				break
			}
		}
	}

	return res, nil
}

// capturePage captures one page, preferring network over DOM.
//
// Network-first because Shopee's own responses are structured JSON; the DOM path
// exists only so that a blocked or changed API does not stop collection
// entirely. The source is recorded so it is visible when the fallback is doing
// the work in production.
func (s *Scraper) capturePage(ctx context.Context, cfg Config, baseURL string, page int) ([]ParsedProduct, error) {
	// Arm the observer BEFORE navigating so the page's own requests are seen.
	// The observer hooks fetch/XHR when it loads; anything already in flight is
	// missed, which is why this is not done after the navigation.
	if _, err := s.sender.Send(ctx, "install_observer", map[string]any{}); err != nil {
		// Non-fatal: the DOM path can still work.
		_ = err
	}

	if err := s.navigate(ctx, cfg, page); err != nil {
		return nil, err
	}

	if err := s.scroll(ctx); err != nil {
		// Scrolling failure only means lazy content may be missing; extraction
		// still runs rather than discarding the page.
		_ = err
	}

	// Network first.
	if products, err := s.fromNetwork(ctx, baseURL); err == nil && len(products) > 0 {
		markSource(products, models.ScrapeSourceNetwork)
		return products, nil
	}

	// DOM fallback.
	products, err := s.fromDOM(ctx, baseURL)
	if err != nil {
		return nil, err
	}
	markSource(products, models.ScrapeSourceDOM)
	return products, nil
}

// markSource stamps the capture path onto each product.
//
// Done centrally so a path cannot forget to label its output, which would make
// it impossible to tell from the data whether the fallback is carrying
// production traffic.
func markSource(products []ParsedProduct, source string) {
	for i := range products {
		products[i].Source = source
	}
}

// navigate opens the URL for the requested page.
func (s *Scraper) navigate(ctx context.Context, cfg Config, page int) error {
	url := s.pageURL(cfg, page)
	if url == "" {
		return fmt.Errorf("scraper: no URL for mode %q", cfg.Mode)
	}
	_, err := s.sender.Send(ctx, "goto", map[string]any{"url": url})
	if err != nil {
		return fmt.Errorf("scraper: navigate to %s: %w", url, err)
	}
	return nil
}

// pageURL builds the URL for a 1-based page number.
//
// Shopee paginates with a 0-indexed `page` query parameter, and its search URL
// uses `keyword`, so both are constructed here rather than in the transport.
//
// Product mode returns "" for any page after the first: a product detail page
// has no pagination, so "page 2 of a product" does not exist. Returning the same
// URL would make the loop navigate to and re-capture the identical page up to
// maxPages times.
func (s *Scraper) pageURL(cfg Config, page int) string {
	base := cfg.BaseURL
	if base == "" {
		base = "https://shopee.co.id"
	}
	base = strings.TrimRight(base, "/")

	if cfg.Mode == "product" {
		if page > 1 {
			return ""
		}
		return cfg.ProductURL
	}

	var url string
	switch cfg.Mode {
	case "shop":
		url = cfg.ShopURL
	default:
		url = fmt.Sprintf("%s/search?keyword=%s", base, urlEncode(cfg.Query))
	}
	if url == "" {
		return ""
	}
	if page <= 1 {
		return url
	}
	sep := "?"
	if strings.Contains(url, "?") {
		sep = "&"
	}
	return fmt.Sprintf("%s%spage=%d", url, sep, page-1)
}

func (s *Scraper) scroll(ctx context.Context) error {
	_, err := s.sender.Send(ctx, "smart_scroll", map[string]any{
		"step": 800, "waitTime": 900, "maxSteps": 25,
	})
	return err
}

func (s *Scraper) isLastPage(ctx context.Context) bool {
	raw, err := s.sender.Send(ctx, "check_last_page", map[string]any{})
	if err != nil {
		// Unknown means "not known to be last"; stopping early would truncate a
		// scrape on a transient failure.
		return false
	}
	var out struct {
		IsLast bool `json:"is_last"`
	}
	if json.Unmarshal(unwrapEnvelope(raw), &out) != nil {
		return false
	}
	return out.IsLast
}

func (s *Scraper) nextPage(ctx context.Context) error {
	_, err := s.sender.Send(ctx, "click_next", map[string]any{})
	return err
}

// fromNetwork reads captured API responses and parses the products out.
//
// The content script replies with an envelope — {"success":true,"data":{...}} —
// so the payload is unwrapped first. Every captured response is then tried
// newest-first, because one page can produce several matching calls and only one
// carries the listing.
func (s *Scraper) fromNetwork(ctx context.Context, baseURL string) ([]ParsedProduct, error) {
	raw, err := s.sender.Send(ctx, "observe_network", map[string]any{
		"filter": "search_items",
		"limit":  5,
	})
	if err != nil {
		return nil, err
	}

	// Unwrap the envelope. A bare object is accepted too, so the content
	// script's wrapper can change without breaking capture.
	payload := unwrapEnvelope(raw)

	var captured struct {
		Responses []struct {
			Status int    `json:"status"`
			URL    string `json:"url"`
			Body   string `json:"body"`
		} `json:"responses"`
	}
	if err := json.Unmarshal(payload, &captured); err != nil {
		return nil, err
	}

	var all []ParsedProduct
	for i := len(captured.Responses) - 1; i >= 0; i-- {
		resp := captured.Responses[i]
		if resp.Status != 200 || resp.Body == "" {
			continue
		}
		products, parseErr := ParseSearchResponse([]byte(resp.Body), baseURL)
		if parseErr != nil {
			continue // try the next captured response
		}

		// Use the FIRST response that parses, then stop. Walking on would bleed a
		// previous page's rows into the current one: a valid-but-empty response
		// for this page is a real answer, and appending an older page's products
		// would stamp them with the wrong page number and stop the empty-page
		// counter from advancing, so the scrape would never detect the end.
		all = append(all, products...)
		if len(all) >= maxProductsPerPage {
			all = all[:maxProductsPerPage]
		}
		return all, nil
	}
	return nil, nil
}

// unwrapEnvelope returns the "data" member of a content-script reply, or the
// input unchanged when there is no usable envelope.
//
// An explicit `"data": null` must NOT be treated as a payload: returning it
// discards any sibling fields, and unmarshalling the literal `null` into a
// struct silently succeeds with a zero value. A reply like
// {"data":null,"responses":[...]} would then lose its responses entirely and
// look like a genuinely empty result.
func unwrapEnvelope(raw json.RawMessage) json.RawMessage {
	var envelope struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return raw
	}
	if !isUsableJSONPayload(envelope.Data) {
		return raw
	}
	return envelope.Data
}

// isUsableJSONPayload reports whether a raw JSON value is present and non-null.
func isUsableJSONPayload(raw json.RawMessage) bool {
	if len(raw) == 0 {
		return false
	}
	trimmed := strings.TrimSpace(string(raw))
	return trimmed != "" && trimmed != "null"
}

// fromDOM asks the content script to extract product cards.
func (s *Scraper) fromDOM(ctx context.Context, baseURL string) ([]ParsedProduct, error) {
	raw, err := s.sender.Send(ctx, "extract", map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("scraper: DOM extraction: %w", err)
	}

	products, err := ParseDOMProducts(unwrapEnvelope(raw), baseURL)
	if err != nil {
		return nil, fmt.Errorf("scraper: parse DOM products: %w", err)
	}
	if len(products) > maxProductsPerPage {
		products = products[:maxProductsPerPage]
	}
	return products, nil
}

// urlEncode percent-encodes a query term.
//
// Hand-written rather than pulling in net/url so the encoding of a space is
// explicit: Shopee expects "+" in a query string, not "%20".
func urlEncode(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-', r == '_', r == '.', r == '~':
			b.WriteRune(r)
		case r == ' ':
			b.WriteRune('+')
		default:
			for _, by := range []byte(string(r)) {
				fmt.Fprintf(&b, "%%%02X", by)
			}
		}
	}
	return b.String()
}
