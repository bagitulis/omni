package shopee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// Dry run for the blocked-write path.
//
// A blocked scrape writes its own status and result_data on the tenant DB, then
// signals the executor with ErrHandlerCompleted so the executor does not
// overwrite the row. This rehearses the whole flow against a throwaway
// database, prints the blast radius (which rows would change, and to what),
// and asserts no other row was touched.

func dryRunSeed(t *testing.T) (*gorm.DB, string, string) {
	t.Helper()
	path := fmt.Sprintf("/tmp/dry_run_%s.db", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Job{}, &models.JobHistory{}, &models.ScrapedProduct{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	qm := jobs.NewQueueManager(db, "")

	target, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
	})
	if err != nil {
		t.Fatalf("AddJob target: %v", err)
	}
	if err := qm.UpdateStatus(target.ID, models.JobStatusRunning, ""); err != nil {
		t.Fatalf("running: %v", err)
	}

	bystander, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"different","extension_id":"ext-2"}`,
	})
	if err != nil {
		t.Fatalf("AddJob bystander: %v", err)
	}

	return db, target.ID, bystander.ID
}

// TestDryRun_BlockedWriteBlastRadius rehearses the blocked write and reports
// exactly which rows would move, so the effect is visible before the real
// deployment sees a captcha.
func TestDryRun_BlockedWriteBlastRadius(t *testing.T) {
	db, targetID, bystanderID := dryRunSeed(t)

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

	// Record the state of the whole jobs table before, so any unintended write
	// shows up.
	var before []models.Job
	if err := db.Order("id").Find(&before).Error; err != nil {
		t.Fatalf("snapshot before: %v", err)
	}

	summary, err := svc.runJobWithID(context.Background(), "t1", `{"mode":"search","query":"kaos","extension_id":"ext-1"}`, targetID)
	if !errors.Is(err, jobs.ErrHandlerCompleted) {
		t.Fatalf("expected ErrHandlerCompleted, got %v", err)
	}

	var after []models.Job
	if err := db.Order("id").Find(&after).Error; err != nil {
		t.Fatalf("snapshot after: %v", err)
	}

	changed := diffJobs(before, after)
	t.Log("\n=== dry-run blast radius (blocked write) ===")
	for _, c := range changed {
		t.Log(c)
	}
	t.Log("=== end blast radius ===")

	if len(changed) != 1 {
		t.Errorf("expected exactly 1 row to move, got %d", len(changed))
	}
	if len(changed) == 1 && changed[0] != fmt.Sprintf("job %s: running -> blocked", targetID) {
		t.Errorf("unexpected change: %s", changed[0])
	}

	// Bystander untouched — evidence of what did NOT happen.
	var bs models.Job
	if err := db.First(&bs, "id = ?", bystanderID).Error; err != nil {
		t.Fatalf("reload bystander: %v", err)
	}
	if bs.Status != models.JobStatusPending {
		t.Errorf("bystander was disturbed: status = %s", bs.Status)
	}

	if summary == "" {
		t.Error("summary must be returned even when the row is self-written")
	}
}

// diffJobs describes which rows in a moved from `before` to `after`.
func diffJobs(before, after []models.Job) []string {
	byID := make(map[string]models.Job, len(before))
	for _, j := range before {
		byID[j.ID] = j
	}
	var changes []string
	for _, a := range after {
		b := byID[a.ID]
		if b.Status != a.Status {
			changes = append(changes, fmt.Sprintf("job %s: %s -> %s", a.ID, b.Status, a.Status))
		}
	}
	return changes
}
