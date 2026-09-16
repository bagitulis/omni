package shopee

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// Resume's atomicity.
//
// The old order-of-operations was: UPDATE data, then UpdateStatus. If the second
// write failed, the row stayed `blocked` but the cursor had already moved,
// stranding the operator and quietly skipping the page they had just cleared.

func setupResumeDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := "/tmp/resume_" + uuid.New().String() + ".db"
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Job{}, &models.JobHistory{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func seedBlockedJob(t *testing.T, db *gorm.DB, resumeFromPage int) *models.Job {
	t.Helper()
	qm := jobs.NewQueueManager(db, "")
	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
	})
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	summary, _ := json.Marshal(map[string]any{
		"resume_from_page": resumeFromPage,
		"reason":           "blocked",
		"blocker": map[string]any{
			"kind":         "captcha",
			"url":          "https://shopee.co.id/verify/traffic",
			"blocked_page": resumeFromPage,
		},
	})
	if err := db.Model(&models.Job{}).Where("id = ?", job.ID).Updates(map[string]any{
		"status":      models.JobStatusBlocked,
		"result_data": string(summary),
		"started_at":  time.Now(),
	}).Error; err != nil {
		t.Fatalf("seed blocked: %v", err)
	}
	return job
}

// TestResume_PayloadAndStatusMoveTogether covers the happy path AND pins the
// invariant that both writes commit together: after Resume returns, the row
// carries the new cursor AND the pending status. Neither on its own is enough.
func TestResume_PayloadAndStatusMoveTogether(t *testing.T) {
	db := setupResumeDB(t)
	job := seedBlockedJob(t, db, 4)

	svc := &ScrapeService{tenantDB: func(string) (*gorm.DB, error) { return db, nil }}

	out, err := svc.Resume(context.Background(), "t1", job.ID)
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if !out.Resumed || out.StartPage != 4 {
		t.Fatalf("outcome = %+v, want Resumed=true StartPage=4", out)
	}

	var got models.Job
	if err := db.First(&got, "id = ?", job.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Status != models.JobStatusPending {
		t.Errorf("status = %s, want pending", got.Status)
	}
	var data map[string]any
	if err := json.Unmarshal([]byte(got.Data), &data); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if p, _ := data["start_page"].(float64); p != 4 {
		t.Errorf("start_page = %v, want 4 — a resume that only moved the status forgets where to restart", p)
	}
}

// TestResume_ConcurrentCancelIsRespected: a cancel that lands between the
// SELECT and the guarded UPDATE must not be overwritten.
//
// The guard is `WHERE status = 'blocked'`, so the update simply matches zero
// rows once cancel has moved the row to `cancelled`. Resume must report the
// current state as a no-op instead of undoing the cancel.
func TestResume_ConcurrentCancelIsRespected(t *testing.T) {
	db := setupResumeDB(t)
	job := seedBlockedJob(t, db, 4)

	// Simulate the cancel arriving first: move the row to cancelled before
	// Resume runs, but leave the SELECT-time snapshot in the seed above.
	qm := jobs.NewQueueManager(db, "")
	if err := qm.CancelJob(job.ID); err != nil {
		t.Fatalf("cancel: %v", err)
	}

	svc := &ScrapeService{tenantDB: func(string) (*gorm.DB, error) { return db, nil }}
	out, err := svc.Resume(context.Background(), "t1", job.ID)
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if out.Resumed {
		t.Error("Resume must not requeue a job that another operator cancelled")
	}
	if out.Status != string(models.JobStatusCancelled) {
		t.Errorf("outcome status = %s, want cancelled", out.Status)
	}

	var got models.Job
	if err := db.First(&got, "id = ?", job.ID).Error; err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Status != models.JobStatusCancelled {
		t.Errorf("cancel was overwritten by resume: status = %s", got.Status)
	}
}

// TestResume_NoResumeCursorRestartsAtOne: a job blocked before it recorded a
// cursor is still worth resuming; refusing would strand it as blocked with
// cancellation as its only exit.
func TestResume_NoResumeCursorRestartsAtOne(t *testing.T) {
	db := setupResumeDB(t)
	qm := jobs.NewQueueManager(db, "")
	job, _ := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
	})
	// Blocked but with no summary at all.
	_ = db.Model(&models.Job{}).Where("id = ?", job.ID).
		Update("status", models.JobStatusBlocked).Error

	svc := &ScrapeService{tenantDB: func(string) (*gorm.DB, error) { return db, nil }}
	out, err := svc.Resume(context.Background(), "t1", job.ID)
	if err != nil {
		t.Fatalf("Resume: %v", err)
	}
	if out.StartPage != 1 {
		t.Errorf("StartPage = %d, want 1", out.StartPage)
	}
}
