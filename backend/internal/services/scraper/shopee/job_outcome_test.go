package shopee

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

// A scrape's stop reason decides the job's fate.
//
// The original defect: `run` inspected only the error from scraper.Run, so every
// stop that returned a nil error — a captcha, an operator cancelling, a Shopee
// redesign — was handed to the executor as a success and written as `completed`.
// These tests pin each reason to the outcome it deserves.

// TestOutcomeForReason_TerminalMapping pins the reason -> outcome table. The
// mapping is the whole point of the change, so it is asserted directly rather
// than only through the paths that happen to reach it.
func TestOutcomeForReason_TerminalMapping(t *testing.T) {
	cases := []struct {
		reason StopReason
		want   scrapeOutcome
	}{
		{StopLastPage, outcomeCompleted},
		{StopMaxPages, outcomeCompleted},
		{StopMaxProducts, outcomeCompleted},
		{StopCancelled, outcomeCancelled},
		{StopBlocked, outcomeBlocked},
		{StopSelectorsBroken, outcomeFailed},
		{StopEmptyPages, outcomeFailed},
	}
	for _, tc := range cases {
		t.Run(string(tc.reason), func(t *testing.T) {
			if got := outcomeForReason(tc.reason); got != tc.want {
				t.Errorf("outcomeForReason(%s) = %v, want %v", tc.reason, got, tc.want)
			}
		})
	}
}

// TestOutcomeForReason_UnknownIsNotSuccess is the fail-closed rule. A reason this
// code has never heard of — an older or newer scraper, a value someone forgot to
// map — must not be read as a completed catalogue.
func TestOutcomeForReason_UnknownIsNotSuccess(t *testing.T) {
	if got := outcomeForReason(StopReason("something_new")); got == outcomeCompleted {
		t.Error("an unmapped stop reason must not map to completed")
	}
}

// TestOutcomeForReason_EmptyIsNotSuccess covers the zero value, which is what a
// Result built by older code or by a test fixture carries.
func TestOutcomeForReason_EmptyIsNotSuccess(t *testing.T) {
	if got := outcomeForReason(StopReason("")); got == outcomeCompleted {
		t.Error("an empty stop reason must not map to completed")
	}
}

// TestScrapeOutcomeError_CarriesReason checks the failure path names what broke.
// "scrape failed" with no page number sends whoever is on call to read logs; the
// message has to say which page and why.
func TestScrapeOutcomeError_CarriesReason(t *testing.T) {
	res := &Result{Reason: StopSelectorsBroken, BrokenPage: 4, Pages: 3}

	err := scrapeOutcomeError(res)
	if err == nil {
		t.Fatal("selectors_broken must produce an error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "selectors_broken") {
		t.Errorf("error must name the reason, got %q", msg)
	}
	if !strings.Contains(msg, "4") {
		t.Errorf("error must name the broken page, got %q", msg)
	}
}

// TestScrapeOutcomeError_BlockedNamesBlocker gives the operator the two facts
// they need to act: what kind of wall it is, and where.
func TestScrapeOutcomeError_BlockedNamesBlocker(t *testing.T) {
	res := &Result{
		Reason:  StopBlocked,
		Blocker: &Blocker{Kind: BlockerCaptcha, URL: "https://shopee.co.id/verify/traffic", Page: 4},
	}

	err := scrapeOutcomeError(res)
	if err == nil {
		t.Fatal("blocked must produce an error describing the blocker")
	}
	msg := err.Error()
	if !strings.Contains(msg, "captcha") {
		t.Errorf("error must name the blocker kind, got %q", msg)
	}
	if !strings.Contains(msg, "verify/traffic") {
		t.Errorf("error must name the blocked URL, got %q", msg)
	}
}

// TestScrapeOutcomeError_CompletedIsNil keeps the happy path quiet.
func TestScrapeOutcomeError_CompletedIsNil(t *testing.T) {
	if err := scrapeOutcomeError(&Result{Reason: StopLastPage}); err != nil {
		t.Errorf("a completed run must not produce an error, got %v", err)
	}
}

// TestScrapeOutcomeError_CancelledIsContextErr lets the executor recognise a
// cancellation as a cancellation rather than as an opaque failure.
func TestScrapeOutcomeError_CancelledIsContextErr(t *testing.T) {
	err := scrapeOutcomeError(&Result{Reason: StopCancelled, Pages: 2})
	if err == nil {
		t.Fatal("a cancelled run must not be reported as a success")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("a cancelled run must wrap context.Canceled, got %v", err)
	}
}

// TestBuildSummary_IncludesReason pins the persisted contract. The summary is
// what the dashboard reads, so the reason has to survive into it, in snake_case
// like every other field there.
func TestBuildSummary_IncludesReason(t *testing.T) {
	res := &Result{Reason: StopMaxPages, Pages: 3, Source: "network"}
	data := &ScrapeJobData{Mode: "search", ExtensionID: "ext-1"}

	raw := buildSummary(res, data, "job-1", 12)

	var out map[string]any
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("summary is not valid JSON: %v", err)
	}
	if out["reason"] != string(StopMaxPages) {
		t.Errorf("summary reason = %v, want %s", out["reason"], StopMaxPages)
	}
	if out["products"] != float64(12) {
		t.Errorf("summary products = %v, want 12", out["products"])
	}
	if out["pages"] != float64(3) {
		t.Errorf("summary pages = %v, want 3", out["pages"])
	}
}

// TestBuildSummary_IncludesBlockerAndResumeCursor is what makes a blocked job
// resumable: without the cursor the operator solves a captcha for a run that has
// no idea where to restart.
func TestBuildSummary_IncludesBlockerAndResumeCursor(t *testing.T) {
	res := &Result{
		Reason:  StopBlocked,
		Pages:   3,
		Blocker: &Blocker{Kind: BlockerCaptcha, URL: "https://shopee.co.id/verify/traffic", Page: 4},
	}
	data := &ScrapeJobData{Mode: "search", ExtensionID: "ext-1"}

	var out map[string]any
	if err := json.Unmarshal([]byte(buildSummary(res, data, "job-1", 40)), &out); err != nil {
		t.Fatalf("summary is not valid JSON: %v", err)
	}

	blocker, ok := out["blocker"].(map[string]any)
	if !ok {
		t.Fatalf("summary must carry a blocker object, got %v", out["blocker"])
	}
	if blocker["kind"] != string(BlockerCaptcha) {
		t.Errorf("blocker kind = %v, want %s", blocker["kind"], BlockerCaptcha)
	}
	if blocker["blocked_page"] != float64(4) {
		t.Errorf("blocked_page = %v, want 4", blocker["blocked_page"])
	}
	if out["resume_from_page"] != float64(4) {
		t.Errorf("resume_from_page = %v, want 4 (the page that was never captured)", out["resume_from_page"])
	}
}

// TestBuildSummary_OmitsBlockerWhenNotBlocked keeps the payload honest: a key
// that is always present teaches readers to ignore it.
func TestBuildSummary_OmitsBlockerWhenNotBlocked(t *testing.T) {
	var out map[string]any
	raw := buildSummary(&Result{Reason: StopLastPage, Pages: 1}, &ScrapeJobData{Mode: "search"}, "job-1", 5)
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		t.Fatalf("summary is not valid JSON: %v", err)
	}
	if _, present := out["blocker"]; present {
		t.Error("a run that was not blocked must not carry a blocker")
	}
}

// TestStartPage_RejectedForProductMode: a product detail page has no page 2, so
// there is nothing to resume to and a start page above 1 is a caller bug.
func TestStartPage_RejectedForProductMode(t *testing.T) {
	data := &ScrapeJobData{
		Mode:        "product",
		ProductURL:  "https://shopee.co.id/product/1/2",
		ExtensionID: "ext-1",
		StartPage:   3,
	}
	if err := validateStartPage(data); err == nil {
		t.Error("start_page > 1 must be rejected for product mode")
	}
}

// TestStartPage_AllowedForSearchAndShop is the resume path proper.
func TestStartPage_AllowedForSearchAndShop(t *testing.T) {
	for _, mode := range []string{"search", "shop"} {
		data := &ScrapeJobData{Mode: mode, StartPage: 4, ExtensionID: "e"}
		if err := validateStartPage(data); err != nil {
			t.Errorf("start_page 4 must be allowed for %s mode, got %v", mode, err)
		}
	}
}

// TestStartPage_NormalisesNonPositive keeps a missing or nonsense cursor from
// silently skipping the first page.
func TestStartPage_NormalisesNonPositive(t *testing.T) {
	for _, in := range []int{0, -5} {
		if got := normaliseStartPage(in); got != 1 {
			t.Errorf("normaliseStartPage(%d) = %d, want 1", in, got)
		}
	}
	if got := normaliseStartPage(7); got != 7 {
		t.Errorf("normaliseStartPage(7) = %d, want 7", got)
	}
}

// TestRun_StartPageKeepsFullPageBudget pins the semantics that the loop's shape
// makes easy to get wrong. maxPages is an absolute bound in the loop, so a naive
// resume at page 5 would silently shrink the run from 50 pages to 46. MaxPages
// means "pages captured in this run", on a resume as much as on a first run.
func TestRun_StartPageKeepsFullPageBudget(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"A","price":1000000,"historical_sold":1,"image":"i1"}}`),
		searchBody(`{"item_basic":{"itemid":2,"shopid":9,"name":"B","price":2000000,"historical_sold":2,"image":"i2"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(false), lastPageReply(false)}

	res, err := NewScraper(f).Run(context.Background(), Config{
		Mode: "search", Query: "kaos", MaxPages: 2, StartPage: 5,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Pages != 2 {
		t.Errorf("Pages = %d, want 2 (pages captured in this run, not the absolute page number)", res.Pages)
	}
	if res.Reason != StopMaxPages {
		t.Errorf("reason = %q, want %s", res.Reason, StopMaxPages)
	}
}

// TestRun_StartPageNavigatesToResumePage checks the run actually resumes where it
// said it would, rather than quietly re-scraping from page 1.
func TestRun_StartPageNavigatesToResumePage(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"A","price":1000000,"historical_sold":1,"image":"i1"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(true)}

	if _, err := NewScraper(f).Run(context.Background(), Config{
		Mode: "search", Query: "kaos", StartPage: 5, MaxPages: 1,
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	gotos := f.payloads["goto"]
	if len(gotos) == 0 {
		t.Fatal("expected a navigation")
	}
	url, _ := gotos[0]["url"].(string)
	// Shopee's page parameter is 0-indexed, so page 5 is page=4.
	if !strings.Contains(url, "page=4") {
		t.Errorf("first navigation = %q, want the resume page (page=4)", url)
	}
}
