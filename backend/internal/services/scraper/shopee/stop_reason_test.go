package shopee

import (
	"context"
	"encoding/json"
	"testing"
)

// A scrape stops for one of several reasons and they are not interchangeable.
//
// Before this, `Run` signalled every stop the same way — it returned
// (result, nil) — so a captcha, an operator cancelling, and a genuinely finished
// catalogue were indistinguishable to the caller. The job executor then wrote
// `completed` for all three. These tests pin the reason to the cause.

// lastPageReply builds a check_last_page envelope.
func lastPageReply(isLast bool) json.RawMessage {
	b, err := json.Marshal(map[string]any{
		"success": true,
		"data":    map[string]any{"is_last": isLast},
	})
	if err != nil {
		panic(err)
	}
	return b
}

// errNoNetworkPayload stands in for the extension having captured nothing over
// the network, which is what forces capturePage onto the DOM path.
type errNoNetworkPayload struct{}

func (errNoNetworkPayload) Error() string { return "no captured responses" }

// domReply builds an `extract` envelope in the shape the content script returns.
//
// containersMatched is the signal this work adds: it says whether any selector
// set found a product grid at all, which is the only way to tell a page that is
// genuinely empty from one whose markup has changed underneath the selectors.
func domReply(rows []map[string]any, containersMatched bool) json.RawMessage {
	if rows == nil {
		rows = []map[string]any{}
	}
	b, err := json.Marshal(map[string]any{
		"success":            true,
		"data":               rows,
		"containers_matched": containersMatched,
	})
	if err != nil {
		panic(err)
	}
	return b
}

// TestRun_LastPageReason pins the ordinary finish: the page said it was the last.
func TestRun_LastPageReason(t *testing.T) {
	f := newFakeSender()
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
}

// TestRun_CancelledReasonIsNotSuccess is the defect that started this work: a
// cancelled scrape returned exactly what a finished one did, so the job was
// written as completed and the operator was told their stop had produced a full
// catalogue.
func TestRun_CancelledReasonIsNotSuccess(t *testing.T) {
	f := newFakeSender()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res, err := NewScraper(f).Run(ctx, Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopCancelled {
		t.Errorf("expected reason=%s, got %q", StopCancelled, res.Reason)
	}
}

// TestRun_MaxPagesReason distinguishes "we ran out of budget" from "there was
// nothing more". Coverage is incomplete in the first case and complete in the
// second, and an operator deciding whether to re-run needs to know which.
func TestRun_MaxPagesReason(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"A","price":1000000,"historical_sold":1,"image":"i1"}}`),
		searchBody(`{"item_basic":{"itemid":2,"shopid":9,"name":"B","price":2000000,"historical_sold":2,"image":"i2"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{
		lastPageReply(false),
		lastPageReply(false),
	}

	res, err := NewScraper(f).Run(context.Background(), Config{
		Mode: "search", Query: "kaos", MaxPages: 2,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopMaxPages {
		t.Errorf("expected reason=%s, got %q", StopMaxPages, res.Reason)
	}
}

// TestRun_MaxProductsReason is the other budget stop.
func TestRun_MaxProductsReason(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"A","price":1000000,"historical_sold":1,"image":"i1"}},{"item_basic":{"itemid":2,"shopid":9,"name":"B","price":2000000,"historical_sold":2,"image":"i2"}}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(false)}

	res, err := NewScraper(f).Run(context.Background(), Config{
		Mode: "search", Query: "kaos", MaxProducts: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopMaxProducts {
		t.Errorf("expected reason=%s, got %q", StopMaxProducts, res.Reason)
	}
	if len(res.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(res.Products))
	}
}

// TestRun_EmptyPagesReason separates an exhausted result set from a broken one.
// Containers were found, they just held nothing, so the selectors are fine.
func TestRun_EmptyPagesReason(t *testing.T) {
	f := newFakeSender()
	// No network payload and a DOM reply that matched containers but produced no
	// rows: a genuinely empty page, repeated until the run gives up.
	f.errs["observe_network"] = errNoNetworkPayload{}
	empty := domReply(nil, true)
	f.responses["extract"] = []json.RawMessage{empty, empty, empty}
	f.responses["check_last_page"] = []json.RawMessage{
		lastPageReply(false), lastPageReply(false), lastPageReply(false),
	}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopEmptyPages {
		t.Errorf("expected reason=%s, got %q", StopEmptyPages, res.Reason)
	}
}

// TestRun_SelectorsBrokenReason is the markup-change case. Every selector set
// matched zero containers, so there is no product grid on the page at all — that
// is a broken scraper, not an empty shop, and reporting it as an empty success
// would hide a Shopee redesign until someone noticed the catalogue was gone.
func TestRun_SelectorsBrokenReason(t *testing.T) {
	f := newFakeSender()
	f.errs["observe_network"] = errNoNetworkPayload{}
	f.responses["extract"] = []json.RawMessage{domReply(nil, false)}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason != StopSelectorsBroken {
		t.Errorf("expected reason=%s, got %q", StopSelectorsBroken, res.Reason)
	}
	if res.BrokenPage != 1 {
		t.Errorf("expected BrokenPage=1, got %d", res.BrokenPage)
	}
}

// TestRun_SelectorsBrokenRequiresNetworkToBeEmptyToo guards against the
// misdiagnosis this reason invites. capturePage prefers the network path, so a
// page that yielded products over the network is healthy no matter what the DOM
// extractor would have said — its selectors were never even consulted.
func TestRun_SelectorsBrokenRequiresNetworkToBeEmptyToo(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(`{"item_basic":{"itemid":1,"shopid":9,"name":"Kaos","price":1000000,"historical_sold":3,"image":"img1"}}`),
	}
	// Scripted but never reached: the network path returns first.
	f.responses["extract"] = []json.RawMessage{domReply(nil, false)}
	f.responses["check_last_page"] = []json.RawMessage{lastPageReply(true)}

	res, err := NewScraper(f).Run(context.Background(), Config{Mode: "search", Query: "kaos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Reason == StopSelectorsBroken {
		t.Error("a page captured over the network must not be reported as broken selectors")
	}
	if len(res.Products) != 1 {
		t.Errorf("expected 1 product, got %d", len(res.Products))
	}
}

// TestStopReason_IsSnakeCase pins the wire format. These values reach the job
// summary JSON and then the dashboard, where the repo's convention is snake_case.
func TestStopReason_IsSnakeCase(t *testing.T) {
	want := map[StopReason]string{
		StopLastPage:        "last_page",
		StopMaxPages:        "max_pages",
		StopMaxProducts:     "max_products",
		StopCancelled:       "cancelled",
		StopBlocked:         "blocked",
		StopSelectorsBroken: "selectors_broken",
		StopEmptyPages:      "empty_pages",
	}
	for reason, text := range want {
		if string(reason) != text {
			t.Errorf("expected %q, got %q", text, string(reason))
		}
	}
}
