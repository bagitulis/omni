package shopee

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// End-to-end coverage of the scrape handler's blocked path.
//
// The generic executor has exactly two outcomes: nil → completed, error →
// failed. Before this test forced it, a blocked run took the error path — the
// row landed as `failed`, the summary containing the resume cursor was
// discarded, and the dashboard had nothing to show the operator.

func setupScrapeDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := "/tmp/scrape_blocked_" + uuid.New().String() + ".db"
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Job{}, &models.JobHistory{}, &models.ScrapedProduct{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

// scriptedHubSender lets RunJob run against a canned scrape sequence.
//
// The real HubSender talks to the browser extension. This one satisfies the
// tab-lifecycle calls RunJob makes before it hands off to Scraper, so the whole
// pipeline can be exercised in-process.
type scriptedHubSender struct {
	replies map[string][]json.RawMessage
}

func (s *scriptedHubSender) Send(_ context.Context, action string, _ any) (json.RawMessage, error) {
	if q, ok := s.replies[action]; ok && len(q) > 0 {
		r := q[0]
		s.replies[action] = q[1:]
		return r, nil
	}
	return json.RawMessage(`{"success":true}`), nil
}

func (*scriptedHubSender) OpenTab(_ context.Context, _ string) (int64, error) { return 1, nil }
func (*scriptedHubSender) CloseTab(_ context.Context)                         {}

// TestRunJob_BlockedIsWrittenAsBlockedNotFailed is the whole point of the
// change: a captcha must land as `blocked` in the database, with the resume
// cursor stored where the resume endpoint can find it later.
func TestRunJob_BlockedIsWrittenAsBlockedNotFailed(t *testing.T) {
	db := setupScrapeDB(t)
	qm := jobs.NewQueueManager(db, "")

	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
	})
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	if err := qm.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		t.Fatalf("running: %v", err)
	}

	svc := &ScrapeService{
		tenantDB: func(string) (*gorm.DB, error) { return db, nil },
		now:      time.Now,
	}
	svc.senderForTest = func(_ string) senderLike {
		return &scriptedHubSender{
			replies: map[string][]json.RawMessage{
				"check_blocked": {mustJSON(t, map[string]any{
					"success": true,
					"data":    map[string]any{"blocked": true, "kind": "captcha", "url": "https://shopee.co.id/verify/traffic"},
				})},
			},
		}
	}

	_, err = svc.runJobWithID(context.Background(), "t1", job.Data, job.ID)

	if !errors.Is(err, jobs.ErrHandlerCompleted) {
		t.Fatalf("a blocked run must signal ErrHandlerCompleted so the executor does not overwrite the row, got %v", err)
	}

	got, gerr := qm.GetJob(job.ID)
	if gerr != nil {
		t.Fatalf("GetJob: %v", gerr)
	}
	if got.Status != models.JobStatusBlocked {
		t.Errorf("status = %s, want blocked", got.Status)
	}
	if got.ResultData == "" {
		t.Fatal("result_data must carry the summary the resume endpoint reads")
	}

	// result_data must be directly parseable — a double-encoded string reads as
	// null on the dashboard and hides the blocker forever.
	var summary map[string]any
	if err := json.Unmarshal([]byte(got.ResultData), &summary); err != nil {
		t.Fatalf("result_data is not parseable JSON: %v (raw=%q)", err, got.ResultData)
	}
	if summary["reason"] != "blocked" {
		t.Errorf("summary.reason = %v, want blocked", summary["reason"])
	}
	if summary["resume_from_page"] == nil {
		t.Error("summary must carry resume_from_page for Resume to know where to restart")
	}
	blocker, _ := summary["blocker"].(map[string]any)
	if blocker["kind"] != "captcha" || blocker["url"] == "" {
		t.Errorf("summary.blocker missing kind/url: %#v", blocker)
	}
}

// TestRunJob_CompletedTakesTheOrdinaryPath keeps the change honest: only the
// blocked case is diverted; a healthy scrape still returns nil for the executor.
func TestRunJob_CompletedTakesTheOrdinaryPath(t *testing.T) {
	db := setupScrapeDB(t)
	qm := jobs.NewQueueManager(db, "")

	job, _ := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
	})
	_ = qm.UpdateStatus(job.ID, models.JobStatusRunning, "")

	svc := &ScrapeService{
		tenantDB: func(string) (*gorm.DB, error) { return db, nil },
		now:      time.Now,
	}
	svc.senderForTest = func(_ string) senderLike {
		return &scriptedHubSender{
			replies: map[string][]json.RawMessage{
				"observe_network": {searchBodyForTest(t)},
				"check_last_page": {mustJSON(t, map[string]any{"success": true, "data": map[string]any{"is_last": true, "reason": "next_disabled"}})},
			},
		}
	}

	summary, err := svc.runJobWithID(context.Background(), "t1", job.Data, job.ID)
	if err != nil {
		t.Fatalf("completed run must return nil, got %v", err)
	}
	if !strings.Contains(summary, `"reason":"last_page"`) {
		t.Errorf("summary missing reason=last_page: %q", summary)
	}

	// A completed run is written by the executor via CompleteJobWithResult, so
	// this test does not assert on the row itself — the executor test in the
	// jobs package covers that. What matters here is the return signature.
}

// mustJSON is a small helper to build canned replies.
func mustJSON(t *testing.T, v any) json.RawMessage {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustJSON: %v", err)
	}
	return b
}

// searchBodyForTest returns an observe_network reply carrying one product.
func searchBodyForTest(t *testing.T) json.RawMessage {
	t.Helper()
	inner := `{"items":[{"item_basic":{"itemid":1,"shopid":9,"name":"Kaos","price":1000000,"historical_sold":3,"image":"i1"}}]}`
	return mustJSON(t, map[string]any{
		"success": true,
		"data": map[string]any{
			"responses": []map[string]any{{
				"status": 200,
				"url":    "https://shopee.co.id/api/v4/search/search_items",
				"body":   inner,
			}},
		},
	})
}
