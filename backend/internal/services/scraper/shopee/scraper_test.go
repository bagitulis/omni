package shopee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
)

// A scripted CommandSender lets the page loop be tested without a browser,
// including failure shapes that are hard to provoke against a real extension.

type fakeSender struct {
	// responses maps an action to a queue of replies. Each Send pops one.
	responses map[string][]json.RawMessage
	// errs maps an action to an error to return, checked before responses.
	errs map[string]error
	// calls records the actions in order, so routing can be asserted.
	calls []string
	// payloads records the payload for each action.
	payloads map[string][]map[string]any
	// failAfter makes an action fail from the Nth call onwards (1-based).
	failAfter map[string]int
	counts    map[string]int
}

func newFakeSender() *fakeSender {
	return &fakeSender{
		responses: make(map[string][]json.RawMessage),
		errs:      make(map[string]error),
		payloads:  make(map[string][]map[string]any),
		failAfter: make(map[string]int),
		counts:    make(map[string]int),
	}
}

func (f *fakeSender) Send(_ context.Context, action string, payload any) (json.RawMessage, error) {
	f.calls = append(f.calls, action)
	f.counts[action]++

	if p, ok := payload.(map[string]any); ok {
		f.payloads[action] = append(f.payloads[action], p)
	}

	if limit, ok := f.failAfter[action]; ok && f.counts[action] >= limit {
		return nil, fmt.Errorf("scripted failure for %s", action)
	}
	if err, ok := f.errs[action]; ok {
		return nil, err
	}

	queue := f.responses[action]
	if len(queue) == 0 {
		// Default success so tests only script what they care about.
		return json.RawMessage(`{"success":true}`), nil
	}
	reply := queue[0]
	f.responses[action] = queue[1:]
	return reply, nil
}

func (f *fakeSender) called(action string) bool {
	for _, c := range f.calls {
		if c == action {
			return true
		}
	}
	return false
}

// searchBody builds an observe_network reply in the shape the content script
// actually returns: an envelope whose data.responses[].body is a JSON *string*.
//
// The body is marshalled rather than concatenated, because the inner JSON must
// be escaped as a string — hand-writing it produces unparseable JSON, which is
// exactly the trap this helper exists to avoid.
func searchBody(itemsJSON string) json.RawMessage {
	inner := `{"items":[` + itemsJSON + `]}`

	reply := map[string]any{
		"success": true,
		"data": map[string]any{
			"responses": []map[string]any{{
				"status": 200,
				"url":    "https://shopee.co.id/api/v4/search/search_items",
				"body":   inner,
			}},
		},
	}
	out, err := json.Marshal(reply)
	if err != nil {
		panic("searchBody: marshal failed: " + err.Error())
	}
	return out
}

// emptyNetworkBody builds a reply with no captured responses.
func emptyNetworkBody() json.RawMessage {
	out, _ := json.Marshal(map[string]any{
		"success": true,
		"data":    map[string]any{"responses": []any{}},
	})
	return out
}

func itemJSON(id, shop, name string) string {
	return fmt.Sprintf(`{"item_basic":{"itemid":%s,"shopid":%s,"name":"%s","price":15000000}}`, id, shop, name)
}

func TestScraper_SearchHappyPath(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(itemJSON("1", "10", "A") + "," + itemJSON("2", "10", "B")),
	}
	f.responses["check_last_page"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":{"is_last":true}}`),
	}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "keyboard"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if len(res.Products) != 2 {
		t.Fatalf("got %d products, want 2", len(res.Products))
	}
	if res.Pages != 1 {
		t.Errorf("Pages = %d, want 1", res.Pages)
	}
}

func TestScraper_NetworkPreferredOverDOM(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{searchBody(itemJSON("1", "10", "FromNetwork"))}
	f.responses["extract"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":[{"name":"FromDOM","price":"1","link":"https://shopee.co.id/D-i.10.99"}]}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":{"is_last":true}}`)}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) != 1 || res.Products[0].ProductName != "FromNetwork" {
		t.Fatalf("network path must win when it yields results, got %+v", res.Products)
	}
	// The DOM path must not even be consulted when network succeeded.
	if f.called("extract") {
		t.Error("DOM extraction must not run when network capture succeeded")
	}
}

func TestScraper_FallsBackToDOMWhenNetworkEmpty(t *testing.T) {
	f := newFakeSender()
	// Network returns nothing usable.
	f.responses["observe_network"] = []json.RawMessage{
		emptyNetworkBody(),
	}
	f.responses["extract"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":[{"name":"FromDOM","price":"Rp1.000","link":"https://shopee.co.id/D-i.10.99"}]}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":{"is_last":true}}`)}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) != 1 || res.Products[0].ProductName != "FromDOM" {
		t.Fatalf("DOM fallback must be used when network yields nothing, got %+v", res.Products)
	}
	if !f.called("extract") {
		t.Error("DOM extraction should have run as the fallback")
	}
}

func TestScraper_ArmObserverBeforeNavigate(t *testing.T) {
	// The observer hooks fetch/XHR when it loads, so arming it after navigation
	// would miss the page's own requests entirely.
	f := newFakeSender()
	f.responses["check_last_page"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":{"is_last":true}}`)}

	s := NewScraper(f)
	_, _ = s.Run(context.Background(), Config{Mode: "search", Query: "x"})

	observerIdx, gotoIdx := -1, -1
	for i, c := range f.calls {
		if c == "install_observer" && observerIdx == -1 {
			observerIdx = i
		}
		if c == "goto" && gotoIdx == -1 {
			gotoIdx = i
		}
	}
	if observerIdx == -1 {
		t.Fatal("install_observer was never called")
	}
	if gotoIdx == -1 {
		t.Fatal("goto was never called")
	}
	if observerIdx > gotoIdx {
		t.Errorf("install_observer (index %d) must precede goto (index %d)", observerIdx, gotoIdx)
	}
}

func TestScraper_DeduplicatesAcrossPages(t *testing.T) {
	f := newFakeSender()
	// Page 1 and page 2 both contain the same item, plus one new item.
	pageBody := searchBody(itemJSON("1", "10", "Dup") + "," + itemJSON("2", "10", "New"))
	f.responses["observe_network"] = []json.RawMessage{
		pageBody,
		searchBody(itemJSON("1", "10", "Dup")),
	}
	f.responses["check_last_page"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":{"is_last":false}}`),
		json.RawMessage(`{"success":true,"data":{"is_last":true}}`),
	}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x", MaxPages: 5})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) != 2 {
		t.Fatalf("got %d products, want 2 (the duplicate must be collapsed): %+v", len(res.Products), res.Products)
	}
}

func TestScraper_StopsAfterConsecutiveEmptyPages(t *testing.T) {
	f := newFakeSender()
	// Never returns products; check_last_page never reports last, so only the
	// empty-page counter can stop the loop.
	emptyBody := emptyNetworkBody()
	for i := 0; i < 10; i++ {
		f.responses["observe_network"] = append(f.responses["observe_network"], emptyBody)
	}
	f.responses["extract"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":[]}`)}
	f.responses["check_last_page"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":{"is_last":false}}`),
	}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x", MaxPages: 100})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) != 0 {
		t.Errorf("expected no products, got %d", len(res.Products))
	}
	// Must stop well before MaxPages, or a dead selector would hammer Shopee.
	if res.Pages > emptyPagesBeforeStop+1 {
		t.Errorf("Pages = %d; must stop after ~%d consecutive empty pages", res.Pages, emptyPagesBeforeStop)
	}
}

func TestScraper_MaxProductsStopsEarly(t *testing.T) {
	f := newFakeSender()
	items := []string{}
	for i := 1; i <= 50; i++ {
		items = append(items, itemJSON(fmt.Sprint(i), "10", fmt.Sprintf("P%d", i)))
	}
	f.responses["observe_network"] = []json.RawMessage{searchBody(strings.Join(items, ","))}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x", MaxProducts: 5})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) != 5 {
		t.Errorf("got %d products, want 5 (MaxProducts must be honoured)", len(res.Products))
	}
}

func TestScraper_CancelledContextReturnsPartialResults(t *testing.T) {
	// A cancelled context is a normal operator stop, so work already collected
	// must be returned rather than discarded.
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{searchBody(itemJSON("1", "10", "A"))}
	f.responses["check_last_page"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":{"is_last":false}}`),
		json.RawMessage(`{"success":true,"data":{"is_last":false}}`),
	}

	ctx, cancel := context.WithCancel(context.Background())
	s := NewScraper(f)
	// Cancel once the first page has been captured.
	var res *Result
	var err error
	cfg := Config{
		Mode:     "search",
		Query:    "x",
		MaxPages: 10,
		OnPage: func(page, captured int) {
			if page == 1 {
				cancel()
			}
		},
	}
	res, err = s.Run(ctx, cfg)
	if err != nil {
		t.Fatalf("a cancelled run must not error: %v", err)
	}
	if len(res.Products) != 1 {
		t.Errorf("got %d products, want the 1 collected before cancellation", len(res.Products))
	}
}

func TestScraper_LaterPageFailureKeepsEarlierResults(t *testing.T) {
	// A failure on page 3 must not discard pages 1 and 2.
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{
		searchBody(itemJSON("1", "10", "A")),
	}
	f.responses["check_last_page"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":{"is_last":false}}`),
	}
	f.failAfter["observe_network"] = 2 // second call fails

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x", MaxPages: 5})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) != 1 {
		t.Errorf("got %d products, want the 1 from the successful page", len(res.Products))
	}
}

func TestScraper_FirstPageFailureIsAnError(t *testing.T) {
	// With nothing collected yet, a failure must surface rather than return a
	// silent empty success.
	f := newFakeSender()
	f.errs["goto"] = errors.New("extension unreachable")

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x"})
	if err == nil {
		t.Fatal("a failure before any results must be an error")
	}
	if res != nil {
		t.Errorf("expected nil result on hard failure, got %+v", res)
	}
}

func TestScraper_PageURLConstruction(t *testing.T) {
	cases := []struct {
		name string
		cfg  Config
		page int
		want string
	}{
		{
			name: "search page 1 has no page param",
			cfg:  Config{Mode: "search", Query: "keyboard"},
			page: 1,
			want: "https://shopee.co.id/search?keyword=keyboard",
		},
		{
			name: "search page 2 is 0-indexed",
			cfg:  Config{Mode: "search", Query: "keyboard"},
			page: 2,
			want: "https://shopee.co.id/search?keyword=keyboard&page=1",
		},
		{
			name: "space encodes as plus",
			cfg:  Config{Mode: "search", Query: "mechanical keyboard"},
			page: 1,
			want: "https://shopee.co.id/search?keyword=mechanical+keyboard",
		},
		{
			name: "non-ascii is percent-encoded",
			cfg:  Config{Mode: "search", Query: "kemeja"},
			page: 1,
			want: "https://shopee.co.id/search?keyword=kemeja",
		},
		{
			name: "shop url page 2",
			cfg:  Config{Mode: "shop", ShopURL: "https://shopee.co.id/testshop"},
			page: 2,
			want: "https://shopee.co.id/testshop?page=1",
		},
		{
			name: "shop url with existing query uses ampersand",
			cfg:  Config{Mode: "shop", ShopURL: "https://shopee.co.id/testshop?sortBy=sales"},
			page: 2,
			want: "https://shopee.co.id/testshop?sortBy=sales&page=1",
		},
	}

	s := NewScraper(newFakeSender())
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := s.pageURL(tc.cfg, tc.page); got != tc.want {
				t.Errorf("pageURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestScraper_RequiresQueryForSearch(t *testing.T) {
	s := NewScraper(newFakeSender())
	for _, q := range []string{"", "   "} {
		if _, err := s.Run(context.Background(), Config{Mode: "search", Query: q}); err == nil {
			t.Errorf("search with query %q must be rejected", q)
		}
	}
}

func TestScraper_RejectsUnsupportedMode(t *testing.T) {
	s := NewScraper(newFakeSender())
	if _, err := s.Run(context.Background(), Config{Mode: "telepathy"}); err == nil {
		t.Error("an unsupported mode must be rejected")
	}
}

func TestScraper_NilSenderIsAnError(t *testing.T) {
	s := NewScraper(nil)
	if _, err := s.Run(context.Background(), Config{Mode: "search", Query: "x"}); err == nil {
		t.Error("a scraper with no sender must error rather than panic")
	}
}

func TestScraper_ObserverFailureStillAllowsDOM(t *testing.T) {
	// If the MAIN-world observer cannot be installed (for example a page that
	// blocks injection), the DOM path must still collect data.
	f := newFakeSender()
	f.errs["install_observer"] = errors.New("injection blocked")
	f.responses["observe_network"] = []json.RawMessage{emptyNetworkBody()}
	f.responses["extract"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":[{"name":"ViaDOM","price":"1","link":"https://shopee.co.id/V-i.1.2"}]}`),
	}
	f.responses["check_last_page"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":{"is_last":true}}`)}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x"})
	if err != nil {
		t.Fatalf("observer failure must not abort the run: %v", err)
	}
	if len(res.Products) != 1 {
		t.Fatalf("got %d products, want 1 via DOM fallback", len(res.Products))
	}
}

func TestScraper_ProductsPerPageIsCapped(t *testing.T) {
	// A broken selector that matches every element must not flood the database.
	f := newFakeSender()
	items := []string{}
	for i := 1; i <= maxProductsPerPage+50; i++ {
		items = append(items, itemJSON(fmt.Sprint(i), "10", fmt.Sprintf("P%d", i)))
	}
	f.responses["observe_network"] = []json.RawMessage{searchBody(strings.Join(items, ","))}
	f.responses["check_last_page"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":{"is_last":true}}`)}

	s := NewScraper(f)
	res, err := s.Run(context.Background(), Config{Mode: "search", Query: "x"})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(res.Products) > maxProductsPerPage {
		t.Errorf("got %d products, want at most %d per page", len(res.Products), maxProductsPerPage)
	}
}

func TestScraper_ProgressCallbackReportsRunningTotal(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{searchBody(itemJSON("1", "10", "A"))}
	f.responses["check_last_page"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":{"is_last":true}}`)}

	var pages, captured []int
	s := NewScraper(f)
	_, err := s.Run(context.Background(), Config{
		Mode:  "search",
		Query: "x",
		OnPage: func(page, count int) {
			pages = append(pages, page)
			captured = append(captured, count)
		},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(pages) != 1 || pages[0] != 1 {
		t.Errorf("pages reported = %v, want [1]", pages)
	}
	if len(captured) != 1 || captured[0] != 1 {
		t.Errorf("captured reported = %v, want [1]", captured)
	}
}
