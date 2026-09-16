package shopee

import (
	"context"
	"encoding/json"
	"fmt"
)

// Turning a scrape's stop reason into a job outcome.
//
// Split from job.go so the mapping — the thing that decides whether an operator
// is told "done", "blocked", or "broken" — is readable on its own, and so job.go
// stays within the file-size guideline as the resume plumbing lands.

// scrapeOutcome is what the job layer should record for a finished run.
type scrapeOutcome int

const (
	// outcomeFailed is the zero value deliberately: a reason this code does not
	// recognise must never be read as success. Unknown reasons come from version
	// skew between the scraper and this mapping, and the safe reading of "I do
	// not know how that run ended" is "not finished".
	outcomeFailed scrapeOutcome = iota
	outcomeCompleted
	outcomeCancelled
	outcomeBlocked
)

// outcomeForReason maps a stop reason to the job outcome it deserves.
func outcomeForReason(reason StopReason) scrapeOutcome {
	switch reason {
	case StopLastPage, StopMaxPages, StopMaxProducts:
		// The run ended on a rule it was given, having seen everything it was
		// allowed to see.
		return outcomeCompleted
	case StopCancelled:
		return outcomeCancelled
	case StopBlocked:
		return outcomeBlocked
	case StopSelectorsBroken, StopEmptyPages, StopPageError:
		return outcomeFailed
	default:
		return outcomeFailed
	}
}

// scrapeOutcomeError converts a non-successful run into the error the job
// executor records, or nil when the run genuinely completed.
//
// A cancelled run wraps context.Canceled so callers can recognise it with
// errors.Is rather than by matching on message text.
func scrapeOutcomeError(res *Result) error {
	switch outcomeForReason(res.Reason) {
	case outcomeCompleted:
		return nil
	case outcomeCancelled:
		return fmt.Errorf("scrape: stopped after %d page(s): %w", res.Pages, context.Canceled)
	case outcomeBlocked:
		if res.Blocker == nil {
			return fmt.Errorf("scrape: blocked after %d page(s)", res.Pages)
		}
		return fmt.Errorf("scrape: blocked by %s at %s on page %d",
			res.Blocker.Kind, res.Blocker.URL, res.Blocker.Page)
	default:
		if res.Reason == StopSelectorsBroken {
			// Name the page: whoever fixes the selectors has to open it, and a
			// bare "selectors broken" sends them hunting through logs first.
			return fmt.Errorf(
				"scrape: %s — no product grid found on page %d; every selector set missed",
				StopSelectorsBroken, res.BrokenPage)
		}
		return fmt.Errorf("scrape: stopped with reason %q after %d page(s)", res.Reason, res.Pages)
	}
}

// buildSummary renders the job's result payload.
//
// This is what the dashboard reads, so every key is snake_case per the repo's
// JSON convention, and the blocker is present only when there actually was one.
func buildSummary(res *Result, data *ScrapeJobData, jobID string, persisted int) string {
	summary := map[string]any{
		"job_id":       jobID,
		"products":     persisted,
		"pages":        res.Pages,
		"source":       res.Source,
		"extension_id": data.ExtensionID,
		"mode":         data.Mode,
		"reason":       string(res.Reason),
	}

	if res.Blocker != nil {
		summary["blocker"] = res.Blocker
		// The blocked page was never captured, so it — not the page after it —
		// is where a resume has to start.
		summary["resume_from_page"] = res.Blocker.Page
	}

	// Marshal cannot fail for this map: every value is a string, an int, or a
	// struct of those. Ignoring the error keeps the caller's signature honest
	// rather than inventing a failure mode that cannot occur.
	raw, _ := json.Marshal(summary)
	return string(raw)
}

// normaliseStartPage clamps a resume cursor to a real page number.
func normaliseStartPage(startPage int) int {
	if startPage <= 0 {
		return 1
	}
	return startPage
}

// validateStartPage rejects a resume cursor the mode cannot honour.
//
// A product detail page has no second page — pageURL returns "" for anything
// past the first — so resuming one at page 3 would navigate nowhere and report an
// empty success.
func validateStartPage(data *ScrapeJobData) error {
	if data.Mode == "product" && data.StartPage > 1 {
		return fmt.Errorf("scrape: start_page %d is not valid for product mode", data.StartPage)
	}
	return nil
}
