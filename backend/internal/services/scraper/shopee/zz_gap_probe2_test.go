package shopee

import (
	"context"
	"encoding/json"
	"testing"
)

// Probes for the remaining flagged paths. As with the earlier probe files, each
// asserts the CORRECT behaviour so a failure is the evidence.

// GAP J: the two capture paths produce DIFFERENT link strings for the same
// product, because the DOM path only absolutises the href and keeps the slug and
// query string while the network path builds a canonical URL.
//
// Links are the dedup key AND the DB unique key, so a mismatch means the same
// product is stored twice when a run mixes capture paths across pages.
func TestGap_DOMAndNetworkLinksAgree(t *testing.T) {
	// The same product, as each path sees it.
	networkLink := BuildProductURL("https://shopee.co.id", "111", "222")

	domLinks := []string{
		"https://shopee.co.id/Some-Product-Name-i.111.222",
		"https://shopee.co.id/Some-Product-Name-i.111.222?sp_atk=abc",
		"https://shopee.co.id/product/111/222",
	}

	for _, raw := range domLinks {
		products, err := ParseDOMProducts(
			[]byte(`[{"name":"P","price":"1","link":"`+raw+`"}]`), "")
		if err != nil {
			t.Fatalf("parse %q: %v", raw, err)
		}
		if len(products) != 1 {
			t.Fatalf("parse %q produced %d products", raw, len(products))
		}
		if products[0].Link != networkLink {
			t.Errorf("DOM link %q stored as %q, want the canonical %q: the same product "+
				"from two capture paths would not dedup",
				raw, products[0].Link, networkLink)
		}
	}
}

// GAP K: progress is not reported for a page that yields nothing, so a caller
// watching progress sees a stall during the empty pages before a stop rather
// than seeing the pages being visited.
func TestGap_OnPageReportsEmptyPages(t *testing.T) {
	f := newFakeSender()
	f.responses["observe_network"] = []json.RawMessage{emptyNetworkBody()}
	f.responses["extract"] = []json.RawMessage{json.RawMessage(`{"success":true,"data":[]}`)}
	f.responses["check_last_page"] = []json.RawMessage{
		json.RawMessage(`{"success":true,"data":{"is_last":false}}`),
	}

	var visited []int
	s := NewScraper(f)
	_, err := s.Run(context.Background(), Config{
		Mode:     "search",
		Query:    "x",
		MaxPages: 5,
		OnPage: func(page, captured int) {
			visited = append(visited, page)
		},
	})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(visited) == 0 {
		t.Error("no progress was reported for pages that yielded no products, so a " +
			"caller cannot tell a slow page from a stalled scrape")
	}
}

// GAP L: a derived job id must be unique even when two runs start in the same
// clock tick, since the id is the results correlation key.
func TestGap_DerivedJobIDIsUnique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id := deriveJobID(nil)
		if id == "" {
			t.Fatal("derived job id must not be empty")
		}
		if seen[id] {
			t.Fatalf("duplicate derived job id %q: two runs would share one results group", id)
		}
		seen[id] = true
	}
}
