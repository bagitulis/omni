package shopee

import (
	"context"
	"encoding/json"
	"fmt"
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

	// StartPage is the 1-based page to begin at. Zero or less means page 1.
	//
	// MaxPages counts pages captured in THIS run, so a resume gets a full budget
	// rather than the remainder of the original one: the loop's bound is derived
	// from StartPage, not fixed at MaxPages.
	StartPage int

	// OnPage is called after each page is captured, for progress reporting. It
	// must not block. page is 1-based; captured is the running total.
	OnPage func(page, captured int)
}

// Result is the outcome of a scrape run.
type Result struct {
	Products []ParsedProduct
	// Pages counts the pages captured in THIS run. On a resume it is not the
	// absolute page number: a run resuming at page 5 and capturing two pages
	// reports 2, so the figure always answers "how much did this run do".
	Pages int
	// Source records which capture path produced the data: "network", "dom", or
	// "" when nothing was captured. Exposed so an operator can tell whether the
	// preferred path is working in production.
	Source string

	// Reason records why the run stopped. The job layer maps it to a job status,
	// so an honest value here is what keeps a blocked or cancelled scrape from
	// being recorded as a completed one.
	Reason StopReason

	// Blocker is set only when Reason is StopBlocked.
	Blocker *Blocker

	// BrokenPage is the 1-based page on which every selector set missed. Set only
	// when Reason is StopSelectorsBroken, so the failure message can name the
	// page someone has to open to fix the selectors.
	BrokenPage int
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
		// A product detail page has no page 2, so a resume cursor past the first
		// page would navigate nowhere and report an empty success.
		if cfg.StartPage > 1 {
			return nil, fmt.Errorf("scraper: start_page %d is not valid for product mode", cfg.StartPage)
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

	// A resumed run starts at its cursor and still gets the full page budget.
	// The loop bound is derived rather than fixed at maxPages because maxPages
	// means "pages captured in this run": using it as an absolute last-page
	// number would silently shrink a resume from page 5 to 46 pages instead of 50.
	startPage := normaliseStartPage(cfg.StartPage)
	lastPage := startPage + maxPages - 1

	for page := startPage; page <= lastPage; page++ {
		if err := ctx.Err(); err != nil {
			// A cancelled context is a normal stop (the operator pressed stop),
			// so partial results are returned rather than discarded — but the
			// reason must say so, or the job layer records a completion.
			res.Reason = StopCancelled
			break
		}
		if len(res.Products) >= maxProducts {
			res.Reason = StopMaxProducts
			break
		}

		pageProducts, capture, err := s.capturePage(ctx, cfg, baseURL, page)
		if err != nil {
			if len(res.Products) > 0 {
				// Fail soft once we have something: a failure on a later page
				// must not throw away the pages already collected.
				res.Reason = StopPageError
				break
			}
			return nil, err
		}

		// A challenge page yielded nothing and is where a resume must restart, so
		// it is not counted among the pages this run captured.
		if capture.blocker != nil {
			res.Reason = StopBlocked
			res.Blocker = capture.blocker
			break
		}

		// Count pages captured by this run, not the absolute page number: on a
		// resume the latter would report progress the run never made.
		res.Pages = page - startPage + 1

		// No product grid anywhere on the page means the markup changed. This is
		// only trustworthy when the network path also came up empty: capturePage
		// prefers the network, so on a healthy page the DOM selectors are never
		// consulted and their verdict would be meaningless.
		if capture.selectorsMissed {
			res.Reason = StopSelectorsBroken
			res.BrokenPage = page
			break
		}

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
				res.Reason = StopEmptyPages
				break
			}
			// Try the next page rather than stopping on the first empty one.
			if page < lastPage && !s.isLastPage(ctx) {
				if s.pageURL(cfg, page+1) == "" {
					res.Reason = StopLastPage
					break
				}
				if nextErr := s.nextPage(ctx); nextErr != nil {
					res.Reason = StopPageError
					break
				}
				continue
			}
			res.Reason = StopLastPage
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
			res.Reason = StopMaxProducts
			break
		}
		// Stop once the page reported itself as last.
		if s.isLastPage(ctx) {
			res.Reason = StopLastPage
			break
		}
		if page < lastPage {
			// A mode with no further pages (product detail) yields no URL, so
			// there is nothing more to capture. Without this the loop would
			// re-navigate to the same page until maxPages.
			if s.pageURL(cfg, page+1) == "" {
				res.Reason = StopLastPage
				break
			}
			if nextErr := s.nextPage(ctx); nextErr != nil {
				res.Reason = StopPageError
				break
			}
			continue
		}
		// The loop is about to end on its own bound, which is the page budget
		// rather than the end of the catalogue.
		res.Reason = StopMaxPages
	}

	if res.Reason == "" {
		// Reached only when maxPages was exhausted without any other exit firing.
		res.Reason = StopMaxPages
	}

	return res, nil
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
