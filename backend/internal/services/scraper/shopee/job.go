package shopee

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/extensions"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// Job type registered with the existing job executor.
//
// The constant lives in models alongside the other job types, so registration
// and dispatch cannot drift; this alias keeps the scraper package's call sites
// readable.
const JobTypeShopeeScrape = models.JobTypeShopeeScrape

// Handler returns a function matching the job executor's JobHandler signature.
//
// The executor supplies the tenant through the context
// (models.ContextKeyTenantID), not as an argument, so the tenant is read from
// there. This preserves the same guarantee as the synchronous path: the job
// cannot name a tenant it does not belong to — the executor resolved it from the
// job's own schema before calling.
func (s *ScrapeService) Handler() func(ctx context.Context, payload string) (string, error) {
	return func(ctx context.Context, payload string) (string, error) {
		tenantID, _ := ctx.Value(models.ContextKeyTenantID).(string)
		if tenantID == "" {
			// Fail closed rather than guessing: with no tenant there is no schema
			// to write into.
			return "", fmt.Errorf("scrape: tenant_id missing from job context")
		}
		// The executor puts the row's id in the context. Reusing it here keeps
		// the self-recorded blocked row and the executor's own bookkeeping on
		// the same row — without it, a resume would look up the derived id and
		// find nothing.
		jobID, _ := ctx.Value(models.ContextKeyJobID).(string)
		return s.runJobWithID(ctx, tenantID, payload, jobID)
	}
}

// ScrapeJobData is the payload stored on a queued job.
//
// Note the absence of a tenant_id: the job lives in the tenant's own schema, so
// the schema is the tenant. Carrying an id here would imply it could be changed,
// which is exactly the cross-tenant hazard the schema design removes.
type ScrapeJobData struct {
	Mode        string `json:"mode"`
	Query       string `json:"query,omitempty"`
	ShopURL     string `json:"shop_url,omitempty"`
	ShopID      string `json:"shop_id,omitempty"`
	ProductURL  string `json:"product_url,omitempty"`
	MaxPages    int    `json:"max_pages,omitempty"`
	MaxProducts int    `json:"max_products,omitempty"`
	BaseURL     string `json:"base_url,omitempty"`
	ExtensionID string `json:"extension_id"`

	// StartPage is the 1-based page a resumed run begins at. Zero means "start at
	// the beginning"; it is set from the resume cursor recorded when a run was
	// blocked, so the operator who cleared a captcha does not have to re-scrape
	// everything that already succeeded.
	StartPage int `json:"start_page,omitempty"`

	// ResultsJobID, when set, is the id scraped products are grouped under. It
	// lets the caller correlate rows with a job it already created.
	ResultsJobID string `json:"results_job_id,omitempty"`
}

// JobID returns the identifier scraped products should be grouped under.
func (d ScrapeJobData) JobID() string { return d.ResultsJobID }

// ScrapeService runs a scrape job end to end: open a tab, run the page loop, and
// persist the products.
type ScrapeService struct {
	// tenantDB resolves the schema-scoped database for a tenant. Passed in so
	// this package does not depend on the connection plumbing.
	tenantDB func(tenantID string) (*gorm.DB, error)

	// hub is used to build the command sender.
	hub *extensions.Hub

	// now is a test seam.
	now func() time.Time

	// senderForTest injects a scripted CommandSender so the whole pipeline can
	// be exercised without a real WebSocket. Production leaves it nil and the
	// service builds a HubSender from the hub above.
	senderForTest func(extensionID string) senderLike
}

// senderLike is the subset of CommandSender the service needs beyond the
// scraper's own interface: open_tab and close_tab.
type senderLike interface {
	CommandSender
	OpenTab(ctx context.Context, url string) (int64, error)
	CloseTab(ctx context.Context)
}

// NewScrapeService creates a ScrapeService.
func NewScrapeService(tenantDB func(tenantID string) (*gorm.DB, error), hub *extensions.Hub) *ScrapeService {
	return &ScrapeService{tenantDB: tenantDB, hub: hub, now: time.Now}
}

// RunJob executes a scrape job. It satisfies the jobs.JobHandler signature so it
// can be registered with the existing executor.
//
// payload is the serialised ScrapeJobData.
func (s *ScrapeService) RunJob(ctx context.Context, tenantID, payload string) (string, error) {
	return s.runJobWithID(ctx, tenantID, payload, "")
}

// runJobWithID is RunJob with the job id supplied by the caller.
//
// Split out so the executor's own job id can be reused when it is available:
// otherwise a blocked run's self-recorded row and its return-value summary would
// carry different ids, and the resume endpoint would look up the wrong one.
func (s *ScrapeService) runJobWithID(ctx context.Context, tenantID, payload, jobID string) (string, error) {
	if tenantID == "" {
		// Fail closed: with no tenant there is no schema to write into, and
		// guessing one would be a cross-tenant leak.
		return "", fmt.Errorf("scrape: tenant_id is required")
	}

	var data ScrapeJobData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return "", fmt.Errorf("scrape: parse job payload: %w", err)
	}
	if data.ExtensionID == "" {
		return "", fmt.Errorf("scrape: extension_id is required")
	}

	db, err := s.tenantDB(tenantID)
	if err != nil {
		return "", fmt.Errorf("scrape: tenant database: %w", err)
	}

	if jobID == "" {
		// The job id is the correlation key for scraped products. A caller that
		// wants the rows groupable supplies one; otherwise a unique id is
		// derived so rows are still attributable to this run.
		jobID = data.JobID()
	}
	if jobID == "" {
		jobID = deriveJobID(s.now)
	}

	return s.run(ctx, db, &data, jobID)
}

// deriveJobID produces a unique id for a run that did not supply one.
//
// A timestamp alone is not sufficient: two runs starting within the same clock
// tick would collide, and Windows timer granularity makes that reachable under
// load. A random suffix removes the possibility entirely, which matters because
// a collision would merge two runs' products under one group.
func deriveJobID(now func() time.Time) string {
	if now == nil {
		now = time.Now
	}
	return fmt.Sprintf("scrape_%d_%s", now().UnixNano(), uuid.New().String()[:8])
}

// run performs the scrape and persists the results.
//
// A run that stops for any reason other than "it saw everything it was allowed
// to see" reports an error, even though scraper.Run returned none. That is the
// whole correction: a captcha, an operator's stop, and a Shopee redesign used to
// arrive here as nil errors and were written to the job queue as completions.
func (s *ScrapeService) run(ctx context.Context, db *gorm.DB, data *ScrapeJobData, jobID string) (string, error) {
	if err := validateStartPage(data); err != nil {
		return "", err
	}

	sender := s.buildSender(data.ExtensionID)

	entryURL := entryURLFor(data)
	if _, err := sender.OpenTab(ctx, entryURL); err != nil {
		return "", fmt.Errorf("scrape: open tab: %w", err)
	}
	// Closed even on failure, so a failed scrape does not leave orphan tabs
	// accumulating in the operator's browser. WithoutCancel because the caller's
	// context is already cancelled on the failure path.
	defer sender.CloseTab(context.WithoutCancel(ctx))

	scraper := NewScraper(sender)
	repo := repositories.NewExtensionRepository(db)
	persisted := 0

	cfg := Config{
		Mode:        data.Mode,
		Query:       data.Query,
		ShopURL:     data.ShopURL,
		ProductURL:  data.ProductURL,
		MaxPages:    data.MaxPages,
		MaxProducts: data.MaxProducts,
		BaseURL:     data.BaseURL,
		StartPage:   data.StartPage,
	}

	result, err := scraper.Run(ctx, cfg)
	if err != nil {
		return "", fmt.Errorf("scrape: run: %w", err)
	}

	// Products are persisted before the outcome is judged: a blocked or cancelled
	// run still collected real rows up to the point it stopped, and discarding
	// them would make the operator re-scrape pages that already succeeded.
	if len(result.Products) > 0 {
		rows := toScrapedProducts(result.Products, jobID, data)
		if insErr := repo.InsertScrapedProducts(ctx, rows); insErr != nil {
			return "", fmt.Errorf("scrape: persist products: %w", insErr)
		}
		persisted = len(rows)
	}

	summary := buildSummary(result, data, jobID, persisted)

	// A cancelled context is authoritative even when the scraper thought it
	// finished cleanly: the executor's own timeout uses ctx.Done, so a run that
	// crossed the deadline on its last page must not be recorded as complete.
	if ctxErr := ctx.Err(); ctxErr != nil {
		return summary, fmt.Errorf("scrape: stopped after %d page(s): %w", result.Pages, ctxErr)
	}

	// A blocked run must land on the row as `blocked`, with the summary so the
	// resume endpoint can read the cursor. The generic executor knows only
	// completed/failed, so this path writes the row itself and returns
	// ErrHandlerCompleted to tell the executor it is done.
	if outcomeForReason(result.Reason) == outcomeBlocked {
		if writeErr := writeBlockedRow(ctx, db, jobID, summary, result); writeErr != nil {
			return summary, fmt.Errorf("scrape: record blocked outcome for %s: %w", jobID, writeErr)
		}
		return summary, blockedHandlerErr(result)
	}

	if outcomeErr := scrapeOutcomeError(result); outcomeErr != nil {
		// The summary is returned alongside the error so partial results and any
		// error-mode metadata are not lost with it.
		return summary, outcomeErr
	}

	return summary, nil
}

// buildSender returns the CommandSender, honouring the test seam if set.
func (s *ScrapeService) buildSender(extensionID string) senderLike {
	if s.senderForTest != nil {
		return s.senderForTest(extensionID)
	}
	return NewHubSender(s.hub, extensionID)
}

// writeBlockedRow persists the blocked status and the summary on the job row.
//
// The write is guarded so a concurrent cancel that fires between capture and
// persistence is not overwritten: only a row still `running` is moved to
// `blocked`. Any other current status is left alone — the operator already
// acted, and reinstating a challenge would be surprising.
func writeBlockedRow(ctx context.Context, db *gorm.DB, jobID, summary string, result *Result) error {
	updates := map[string]any{
		"status":      models.JobStatusBlocked,
		"result_data": summary,
		"updated_at":  time.Now(),
	}
	if result.Blocker != nil {
		updates["error_message"] = fmt.Sprintf("blocked by %s at %s on page %d",
			result.Blocker.Kind, result.Blocker.URL, result.Blocker.Page)
	}
	res := db.WithContext(ctx).Model(&models.Job{}).
		Where("id = ? AND status = ?", jobID, models.JobStatusRunning).
		Updates(updates)
	if res.Error != nil {
		return res.Error
	}
	// Zero rows means the job was cancelled or otherwise moved out of running
	// before the block was recorded. That is a legitimate outcome, not an error:
	// the operator's action wins, and the scrape simply stopped.
	return nil
}

// blockedHandlerErr wraps the sentinel with a message the executor can log.
func blockedHandlerErr(result *Result) error {
	kind := "unknown"
	page := 0
	if result.Blocker != nil {
		kind = string(result.Blocker.Kind)
		page = result.Blocker.Page
	}
	return fmt.Errorf("scrape: blocked (%s at page %d): %w", kind, page, jobs.ErrHandlerCompleted)
}

// entryURLFor picks the URL the scrape tab should open at.
func entryURLFor(data *ScrapeJobData) string {
	switch data.Mode {
	case "shop":
		return data.ShopURL
	case "product":
		return data.ProductURL
	default:
		base := data.BaseURL
		if base == "" {
			base = "https://shopee.co.id"
		}
		return fmt.Sprintf("%s/search?keyword=%s", trimSlash(base), urlEncode(data.Query))
	}
}

// toScrapedProducts converts parsed products into tenant-scoped rows.
func toScrapedProducts(products []ParsedProduct, jobID string, data *ScrapeJobData) []models.ScrapedProduct {
	out := make([]models.ScrapedProduct, 0, len(products))
	for _, p := range products {
		out = append(out, models.ScrapedProduct{
			JobID:        jobID,
			Platform:     "shopee",
			ScrapeMode:   data.Mode,
			Query:        data.Query,
			ShopID:       firstNonEmpty(p.ShopID, data.ShopID),
			ProductName:  p.ProductName,
			Price:        p.Price,
			Sold:         p.Sold,
			Link:         p.Link,
			ImageURL:     p.ImageURL,
			ShopeeItemID: p.ShopeeItemID,
			PageNumber:   p.Page,
			Source:       p.Source,
		})
	}
	return out
}

func trimSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
