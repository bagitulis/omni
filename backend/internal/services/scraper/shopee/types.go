package shopee

// Scrape outcome types.
//
// Split out of scraper.go, which was already past the line ceiling before this
// work added to it. Keeping the vocabulary of an outcome — why a run stopped and
// what blocked it — in its own file also makes it obvious that these values are
// part of the contract with the job layer and the dashboard, not internal
// scraper bookkeeping.

// StopReason records why a scrape stopped.
//
// It exists because `break` is not an explanation. The page loop has six exits
// and they used to be indistinguishable to the caller, so a captcha, an operator
// pressing stop, and a genuinely exhausted catalogue all produced the same
// "success". The job layer maps these to job statuses, so the values are part of
// the persisted contract and are snake_case to match the repo's JSON convention.
type StopReason string

const (
	// StopLastPage — the page reported itself as the last one. Coverage complete.
	StopLastPage StopReason = "last_page"
	// StopMaxPages — the page budget ran out. Coverage incomplete by design.
	StopMaxPages StopReason = "max_pages"
	// StopMaxProducts — the product budget ran out. Coverage incomplete by design.
	StopMaxProducts StopReason = "max_products"
	// StopCancelled — the context was cancelled. Partial results are kept, but
	// this is never a completion.
	StopCancelled StopReason = "cancelled"
	// StopBlocked — Shopee interposed a captcha, login wall, or rate limit. A
	// human must clear it; the run can resume afterwards.
	StopBlocked StopReason = "blocked"
	// StopSelectorsBroken — no product grid was found on the page at all. The
	// markup changed; the scraper is broken and must fail loudly rather than
	// report an empty catalogue.
	StopSelectorsBroken StopReason = "selectors_broken"
	// StopEmptyPages — several consecutive pages held no products. The result set
	// is exhausted; the selectors are fine.
	StopEmptyPages StopReason = "empty_pages"
	// StopPageError — a page or a navigation failed after products had already
	// been collected. The run keeps what it has, but coverage is incomplete and
	// the caller must not read it as a finished catalogue.
	StopPageError StopReason = "page_error"
)

// BlockerKind classifies what stopped a blocked run.
type BlockerKind string

const (
	BlockerCaptcha       BlockerKind = "captcha"
	BlockerLoginRequired BlockerKind = "login_required"
	BlockerRateLimited   BlockerKind = "rate_limited"
	BlockerAntiBot       BlockerKind = "anti_bot"
)

// Blocker describes an obstacle a human has to clear.
//
// It carries no page text or HTML: the operator needs to know what kind of wall
// it is and where, and capturing the page body would drag Shopee's content into
// omni's database and logs for no operational gain.
type Blocker struct {
	Kind BlockerKind `json:"kind"`
	URL  string      `json:"url"`
	// Page is the 1-based page the block appeared on. It doubles as the resume
	// cursor: the run restarts here, because this page was never captured.
	Page int `json:"blocked_page"`
}
