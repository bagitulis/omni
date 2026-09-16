package shopee

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/models"
)

// Capturing one page.
//
// Split from scraper.go, which owns the page loop and its stopping rules. This
// file owns the question the loop asks each iteration: what is on this page, and
// how did we get it — over Shopee's own API, or by reading the rendered DOM.

// captureOutcome carries what a page capture learned beyond the products
// themselves.
type captureOutcome struct {
	// selectorsMissed is true only when the network path produced nothing AND
	// the DOM path found no product grid at all. Both halves matter: the network
	// path wins when it works, so a DOM verdict on a page it never examined says
	// nothing about the selectors.
	selectorsMissed bool

	// blocker is set when Shopee interposed a challenge on this page. The page
	// was not captured, so its number is also the resume cursor.
	blocker *Blocker
}

// capturePage captures one page, preferring network over DOM.
//
// Network-first because Shopee's own responses are structured JSON; the DOM path
// exists only so that a blocked or changed API does not stop collection
// entirely. The source is recorded so it is visible when the fallback is doing
// the work in production.
func (s *Scraper) capturePage(ctx context.Context, cfg Config, baseURL string, page int) ([]ParsedProduct, captureOutcome, error) {
	// Arm the observer BEFORE navigating so the page's own requests are seen.
	// The observer hooks fetch/XHR when it loads; anything already in flight is
	// missed, which is why this is not done after the navigation.
	if _, err := s.sender.Send(ctx, "install_observer", map[string]any{}); err != nil {
		// Non-fatal: the DOM path can still work.
		_ = err
	}

	if err := s.navigate(ctx, cfg, page); err != nil {
		return nil, captureOutcome{}, err
	}

	// Check for a challenge before extracting anything. A captcha page has no
	// product grid, so extracting it first would report broken selectors and send
	// someone to fix markup that never changed.
	if blocker := s.checkBlocked(ctx, page); blocker != nil {
		return nil, captureOutcome{blocker: blocker}, nil
	}

	if err := s.scroll(ctx); err != nil {
		// Scrolling failure only means lazy content may be missing; extraction
		// still runs rather than discarding the page.
		_ = err
	}

	// Network first.
	if products, err := s.fromNetwork(ctx, baseURL); err == nil && len(products) > 0 {
		markSource(products, models.ScrapeSourceNetwork)
		return products, captureOutcome{}, nil
	}

	// DOM fallback.
	products, containersMatched, err := s.fromDOM(ctx, baseURL)
	if err != nil {
		return nil, captureOutcome{}, err
	}
	markSource(products, models.ScrapeSourceDOM)
	return products, captureOutcome{selectorsMissed: !containersMatched}, nil
}

// checkBlocked asks the page whether a human has to clear something first.
//
// A failed check is NOT a block. An extension too old to know the command, or a
// transient messaging failure, would otherwise stop every scrape the moment the
// check itself broke — turning a diagnostic into an outage.
func (s *Scraper) checkBlocked(ctx context.Context, page int) *Blocker {
	raw, err := s.sender.Send(ctx, "check_blocked", map[string]any{})
	if err != nil {
		return nil
	}

	var out struct {
		Blocked bool   `json:"blocked"`
		Kind    string `json:"kind"`
		URL     string `json:"url"`
	}
	if json.Unmarshal(unwrapEnvelope(raw), &out) != nil {
		return nil
	}
	if !out.Blocked {
		return nil
	}

	// The kind is passed through rather than validated against the known set:
	// the extension may learn a new wall before the backend does, and discarding
	// an unrecognised kind would turn a real block into a silent success.
	return &Blocker{Kind: BlockerKind(out.Kind), URL: out.URL, Page: page}
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

func (s *Scraper) scroll(ctx context.Context) error {
	_, err := s.sender.Send(ctx, "smart_scroll", map[string]any{
		"step": 800, "waitTime": 900, "maxSteps": 25,
	})
	return err
}

func (s *Scraper) nextPage(ctx context.Context) error {
	_, err := s.sender.Send(ctx, "click_next", map[string]any{})
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

// fromDOM asks the content script to extract product cards.
//
// The second return value reports whether any selector set matched a product
// grid. It is read from the full envelope rather than through unwrapEnvelope,
// which returns only the "data" member and would discard the flag: that function
// has other callers and a deliberate null-data invariant, so it is left alone.
//
// An envelope without the field at all is treated as "matched", so an older
// extension build degrades to the previous behaviour instead of having every
// page reported as broken selectors.
func (s *Scraper) fromDOM(ctx context.Context, baseURL string) ([]ParsedProduct, bool, error) {
	raw, err := s.sender.Send(ctx, "extract", map[string]any{})
	if err != nil {
		return nil, true, fmt.Errorf("scraper: DOM extraction: %w", err)
	}

	containersMatched := true
	var meta struct {
		ContainersMatched *bool `json:"containers_matched"`
	}
	if json.Unmarshal(raw, &meta) == nil && meta.ContainersMatched != nil {
		containersMatched = *meta.ContainersMatched
	}

	products, err := ParseDOMProducts(unwrapEnvelope(raw), baseURL)
	if err != nil {
		return nil, containersMatched, fmt.Errorf("scraper: parse DOM products: %w", err)
	}
	if len(products) > maxProductsPerPage {
		products = products[:maxProductsPerPage]
	}
	return products, containersMatched, nil
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
