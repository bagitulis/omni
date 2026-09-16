package shopee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// Resuming a blocked scrape.
//
// A blocked job is the only non-terminal status a human has to act on, so these
// tests pin the two things that make it worth having: the cursor survives the
// round trip, and a resume never invents work for a job that is not blocked.

func setupResumeTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&models.Job{}); err != nil {
		t.Fatalf("auto-migrate: %v", err)
	}
	return db
}

// seedJob inserts a job directly in the requested status, bypassing the queue's
// transition guard so a test can start from a state the guard would refuse to
// reach in one step.
func seedJob(t *testing.T, db *gorm.DB, status models.JobStatus, data, resultData string) string {
	t.Helper()
	job := models.Job{
		ID:         uuid.New().String(),
		Type:       models.JobTypeShopeeScrape,
		Status:     status,
		Priority:   "normal",
		Data:       data,
		ResultData: resultData,
	}
	if err := db.Create(&job).Error; err != nil {
		t.Fatalf("seed job: %v", err)
	}
	return job.ID
}

func resumeServiceFor(db *gorm.DB) *ScrapeService {
	return NewScrapeService(func(tenantID string) (*gorm.DB, error) {
		if tenantID != "tenant-a" {
			return nil, fmt.Errorf("unknown tenant %q", tenantID)
		}
		return db, nil
	}, nil)
}

func payloadOf(t *testing.T, db *gorm.DB, jobID string) ScrapeJobData {
	t.Helper()
	var job models.Job
	if err := db.Where("id = ?", jobID).First(&job).Error; err != nil {
		t.Fatalf("read back job: %v", err)
	}
	var data ScrapeJobData
	if err := json.Unmarshal([]byte(job.Data), &data); err != nil {
		t.Fatalf("job payload is not valid JSON: %v", err)
	}
	return data
}

func statusOfJob(t *testing.T, db *gorm.DB, jobID string) models.JobStatus {
	t.Helper()
	var job models.Job
	if err := db.Where("id = ?", jobID).First(&job).Error; err != nil {
		t.Fatalf("read back job: %v", err)
	}
	return job.Status
}

// TestResume_BlockedJobIsRequeuedFromCursor is the whole point of the feature:
// the operator cleared the captcha, and the pages that already succeeded are not
// scraped again.
func TestResume_BlockedJobIsRequeuedFromCursor(t *testing.T) {
	db := setupResumeTestDB(t)
	jobID := seedJob(t, db, models.JobStatusBlocked,
		`{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
		`{"job_id":"j","reason":"blocked","resume_from_page":4,"blocker":{"kind":"captcha","url":"https://shopee.co.id/verify/traffic","blocked_page":4}}`)

	out, err := resumeServiceFor(db).Resume(context.Background(), "tenant-a", jobID)
	if err != nil {
		t.Fatalf("Resume error: %v", err)
	}

	if !out.Resumed {
		t.Error("a blocked job must report Resumed=true")
	}
	if out.Status != string(models.JobStatusPending) {
		t.Errorf("Status = %q, want pending", out.Status)
	}
	if out.StartPage != 4 {
		t.Errorf("StartPage = %d, want 4 (the page that was never captured)", out.StartPage)
	}
	if got := statusOfJob(t, db, jobID); got != models.JobStatusPending {
		t.Errorf("persisted status = %q, want pending", got)
	}
	if got := payloadOf(t, db, jobID).StartPage; got != 4 {
		t.Errorf("persisted start_page = %d, want 4; without it the resumed run restarts from page 1", got)
	}
}

// TestResume_NotBlockedIsANoOp: a resume on a running or finished job is not an
// error. The dashboard polls, and two operators clicking Resume must not produce
// a failure the second time.
func TestResume_NotBlockedIsANoOp(t *testing.T) {
	for _, status := range []models.JobStatus{
		models.JobStatusPending,
		models.JobStatusRunning,
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
	} {
		t.Run(string(status), func(t *testing.T) {
			db := setupResumeTestDB(t)
			jobID := seedJob(t, db, status,
				`{"mode":"search","query":"kaos","extension_id":"ext-1"}`,
				`{"resume_from_page":4}`)

			out, err := resumeServiceFor(db).Resume(context.Background(), "tenant-a", jobID)
			if err != nil {
				t.Fatalf("a resume on a %s job must be a no-op, got error: %v", status, err)
			}
			if out.Resumed {
				t.Errorf("Resumed = true for status %s, want false", status)
			}
			if out.Status != string(status) {
				t.Errorf("Status = %q, want the current status %q", out.Status, status)
			}
			if got := statusOfJob(t, db, jobID); got != status {
				t.Errorf("persisted status changed to %q; a no-op must not touch the job", got)
			}
			if got := payloadOf(t, db, jobID).StartPage; got != 0 {
				t.Errorf("payload start_page = %d; a no-op must not rewrite the payload", got)
			}
		})
	}
}

// TestResume_UnknownJobIsNotFound keeps a typo'd id from being reported as a
// successful resume.
func TestResume_UnknownJobIsNotFound(t *testing.T) {
	db := setupResumeTestDB(t)

	_, err := resumeServiceFor(db).Resume(context.Background(), "tenant-a", "no-such-job")
	if !errors.Is(err, ErrJobNotFound) {
		t.Fatalf("want ErrJobNotFound, got %v", err)
	}
}

// TestResume_MissingTenantFailsClosed enforces the no-default-tenant rule at the
// service boundary, not only at the handler.
func TestResume_MissingTenantFailsClosed(t *testing.T) {
	called := false
	svc := NewScrapeService(func(string) (*gorm.DB, error) {
		called = true
		return nil, nil
	}, nil)

	_, err := svc.Resume(context.Background(), "", "job-1")
	if err == nil {
		t.Fatal("an empty tenant must be an error, never a default")
	}
	if called {
		t.Error("the tenant database must not be resolved for an empty tenant")
	}
}

// TestResume_BlockedWithoutCursorRestartsAtPageOne: a job blocked before it
// recorded a cursor is still resumable. Refusing would strand it as blocked with
// no way out but cancellation.
func TestResume_BlockedWithoutCursorRestartsAtPageOne(t *testing.T) {
	cases := map[string]string{
		"no result data":        "",
		"summary without key":   `{"job_id":"j","reason":"blocked","pages":0}`,
		"unparsable result":     `not json at all`,
		"cursor below page one": `{"resume_from_page":0}`,
	}

	for name, resultData := range cases {
		t.Run(name, func(t *testing.T) {
			db := setupResumeTestDB(t)
			jobID := seedJob(t, db, models.JobStatusBlocked,
				`{"mode":"search","query":"kaos","extension_id":"ext-1"}`, resultData)

			out, err := resumeServiceFor(db).Resume(context.Background(), "tenant-a", jobID)
			if err != nil {
				t.Fatalf("Resume error: %v", err)
			}
			if !out.Resumed || out.StartPage != 1 {
				t.Errorf("Resumed=%v StartPage=%d, want a resume from page 1", out.Resumed, out.StartPage)
			}
			if got := statusOfJob(t, db, jobID); got != models.JobStatusPending {
				t.Errorf("persisted status = %q, want pending", got)
			}
		})
	}
}

// TestResume_ProductModeCursorIsRejected: a product detail page has no page 2, so
// queueing such a resume would produce a run that navigates nowhere and reports
// an empty success.
func TestResume_ProductModeCursorIsRejected(t *testing.T) {
	db := setupResumeTestDB(t)
	jobID := seedJob(t, db, models.JobStatusBlocked,
		`{"mode":"product","product_url":"https://shopee.co.id/product/1/2","extension_id":"ext-1"}`,
		`{"resume_from_page":3}`)

	if _, err := resumeServiceFor(db).Resume(context.Background(), "tenant-a", jobID); err == nil {
		t.Fatal("a product-mode resume past page 1 must be rejected")
	}
	if got := statusOfJob(t, db, jobID); got != models.JobStatusBlocked {
		t.Errorf("status = %q after a refused resume, want it to stay blocked", got)
	}
}

// TestResume_UnknownTenantSurfacesTheError: a tenant whose schema cannot be
// resolved must not silently become a no-op resume.
func TestResume_UnknownTenantSurfacesTheError(t *testing.T) {
	db := setupResumeTestDB(t)

	if _, err := resumeServiceFor(db).Resume(context.Background(), "tenant-zzz", "job-1"); err == nil {
		t.Fatal("an unresolvable tenant must be an error")
	}
}

// TestResume_CorruptPayloadIsRejected: rewriting a payload the service cannot
// parse would replace the job's data with a guess.
func TestResume_CorruptPayloadIsRejected(t *testing.T) {
	db := setupResumeTestDB(t)
	jobID := seedJob(t, db, models.JobStatusBlocked, `{not json`, `{"resume_from_page":2}`)

	if _, err := resumeServiceFor(db).Resume(context.Background(), "tenant-a", jobID); err == nil {
		t.Fatal("an unparsable job payload must be an error, not a rewrite")
	}
	if got := statusOfJob(t, db, jobID); got != models.JobStatusBlocked {
		t.Errorf("status = %q after a refused resume, want it to stay blocked", got)
	}
}

func TestResumeCursorFrom(t *testing.T) {
	cases := []struct {
		name       string
		resultData string
		want       int
	}{
		{"cursor present", `{"resume_from_page":7}`, 7},
		{"cursor absent", `{"pages":2}`, 1},
		{"empty", "", 1},
		{"not json", `<html>`, 1},
		{"zero clamps to one", `{"resume_from_page":0}`, 1},
		{"negative clamps to one", `{"resume_from_page":-3}`, 1},
		{"wrong type", `{"resume_from_page":"four"}`, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := resumeCursorFrom(tc.resultData); got != tc.want {
				t.Errorf("resumeCursorFrom(%q) = %d, want %d", tc.resultData, got, tc.want)
			}
		})
	}
}
