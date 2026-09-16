package shopee

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// A blocked page must stop the scrape and say why.
//
// Shopee interposes a captcha or a login wall by serving a page with neither a
// product grid nor pagination. Before the blocker check, that page satisfied
// "no next control, therefore last page", so the run ended as a clean success
// over an empty catalogue and the operator had nothing to act on.

// blockedReply builds a check_blocked envelope in the shape the content script
// returns.
func blockedReply(kind, url string) json.RawMessage {
	b, err := json.Marshal(map[string]any{
		"success": true,
		"data":    map[string]any{"blocked": kind != "", "kind": kind, "url": url},
	})
	if err != nil {
		panic(err)
	}
	return b
}

// TestRun_BlockedStopsTheRun pins the core behaviour: a captcha ends the run as
// blocked, not as a completion.
func TestRun_BlockedStopsTheRun(t *testing.T) {
	f := newFakeSender()
	f.responses["check_blocked"] = []json.RawMessage{
		blockedReply("captcha", "https://shopee.co.id/verify/traffic"),
	}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopBlocked {
		t.Fatalf("expected reason=%s, got %q", StopBlocked, res.Reason)
	}
	if res.Blocker == nil {
		t.Fatal("a blocked run must carry a blocker")
	}
	if res.Blocker.Kind != BlockerCaptcha {
		t.Errorf("blocker kind = %q, want %s", res.Blocker.Kind, BlockerCaptcha)
	}
	if res.Blocker.URL != "https://shopee.co.id/verify/traffic" {
		t.Errorf("blocker URL = %q", res.Blocker.URL)
	}
	if res.Blocker.Page != 1 {
		t.Errorf("blocker page = %d, want 1", res.Blocker.Page)
	}
}

// TestRun_BlockedKeepsEarlierPages is what makes a block resumable rather than a
// total loss: pages captured before the wall went up are real data.
func TestRun_BlockedKeepsEarlierPages(t *testing.T) {
	f := newFakeSender()
	f.responses["check_blocked"] = []json.RawMessage{
		blockedReply("", ""),
		blockedReply("login_required", "https://shopee.co.id/buyer/login"),
	}
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"A","price":1000000,"historical_sold":1,"image":"i1"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(false)}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopBlocked {
		t.Fatalf("expected reason=%s, got %q", StopBlocked, res.Reason)
	}
	if len(res.Products) != 1 {
		t.Errorf("products captured before the block must be kept, got %d", len(res.Products))
	}
	if res.Blocker.Page != 2 {
		t.Errorf("blocker page = %d, want 2 (the page that was never captured)", res.Blocker.Page)
	}
	if res.Pages != 1 {
		t.Errorf("Pages = %d, want 1 (only page 1 was captured)", res.Pages)
	}
}

// TestRun_BlockedResumeCursorIsTheBlockedPage: the blocked page produced nothing,
// so a resume has to start there, not after it.
func TestRun_BlockedResumeCursorIsTheBlockedPage(t *testing.T) {
	f := newFakeSender()
	f.responses["check_blocked"] = []json.RawMessage{
		blockedReply("captcha", "https://shopee.co.id/verify/traffic"),
	}

	res, _ := NewScraper(f).Run(context.Background(), Config{
		Mode: "search", Query: "kaos", StartPage: 4,
	})
	if res.Blocker == nil {
		t.Fatal("expected a blocker")
	}
	if res.Blocker.Page != 4 {
		t.Errorf("resume cursor = %d, want 4", res.Blocker.Page)
	}
}

// TestRun_UnblockedPageProceeds keeps the check from becoming a brake on healthy
// runs.
func TestRun_UnblockedPageProceeds(t *testing.T) {
	f := newFakeSender()
	f.responses["check_blocked"] = []json.RawMessage{blockedReply("", "")}
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"Kaos","price":1000000,"historical_sold":3,"image":"img1"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(true)}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopLastPage {
		t.Errorf("expected reason=%s, got %q", StopLastPage, res.Reason)
	}
	if len(res.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(res.Products))
	}
}

// TestRun_BlockerCheckFailureDoesNotStopTheRun: an extension too old to know the
// command, or a transient messaging failure, must not be read as a block. Doing
// so would stop every scrape the moment the check itself broke.
func TestRun_BlockerCheckFailureIsNotABlock(t *testing.T) {
	f := newFakeSender()
	f.errs["check_blocked"] = errNoNetworkPayload{}
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"Kaos","price":1000000,"historical_sold":3,"image":"img1"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(true)}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason == StopBlocked {
		t.Error("a failed blocker check must not be treated as a block")
	}
	if len(res.Products) != 1 {
		t.Errorf("expected the run to proceed and capture 1 product, got %d", len(res.Products))
	}
}

// TestRun_UnknownBlockerKindIsStillABlock: the extension may learn a new kind
// before the backend does, and "I do not recognise this wall" is still a wall.
func TestRun_UnknownBlockerKindIsStillABlock(t *testing.T) {
	f := newFakeSender()
	f.responses["check_blocked"] = []json.RawMessage{
		blockedReply("some_new_wall", "https://shopee.co.id/verify/whatever"),
	}

	res, _ := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if res.Reason != StopBlocked {
		t.Fatalf("expected reason=%s, got %q", StopBlocked, res.Reason)
	}
	if res.Blocker.Kind != BlockerKind("some_new_wall") {
		t.Errorf("the reported kind must be preserved, got %q", res.Blocker.Kind)
	}
}

// TestRun_BlockedBeforeExtraction checks the ordering. The blocker check runs
// after navigation but BEFORE extraction, so a captcha page is never scraped for
// products and cannot be misreported as broken selectors.
func TestRun_BlockedIsCheckedBeforeExtraction(t *testing.T) {
	f := newFakeSender()
	f.responses["check_blocked"] = []json.RawMessage{
		blockedReply("captcha", "https://shopee.co.id/verify/traffic"),
	}

	NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})

	if f.called("extract") {
		t.Error("a blocked page must not be extracted")
	}
	blockedAt, extractAt := -1, -1
	for i, c := range f.calls {
		if c == "check_blocked" && blockedAt == -1 {
			blockedAt = i
		}
		if c == "observe_network" && extractAt == -1 {
			extractAt = i
		}
	}
	if blockedAt == -1 {
		t.Fatal("check_blocked was never called")
	}
	if extractAt != -1 && extractAt < blockedAt {
		t.Error("the blocker check must run before capture")
	}
}

// TestScrapeOutcomeError_BlockedNamesPage completes the operator-facing message:
// the kind, the URL, and the page they need to return to.
func TestScrapeOutcomeError_BlockedNamesPage(t *testing.T) {
	err := scrapeOutcomeError(&Result{
		Reason:  StopBlocked,
		Blocker: &Blocker{Kind: BlockerCaptcha, URL: "https://shopee.co.id/verify/traffic", Page: 4},
	})
	if err == nil {
		t.Fatal("expected an error")
	}
	if !strings.Contains(err.Error(), "page 4") {
		t.Errorf("error must name the blocked page, got %q", err.Error())
	}
}
