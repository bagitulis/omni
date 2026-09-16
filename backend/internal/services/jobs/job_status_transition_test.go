package jobs_test

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// Status transitions must be guarded.
//
// A `blocked` scrape carries a resume cursor. Every one of these tests exists
// because an unguarded write would destroy that cursor while reporting success:
// the operator would be told the job finished, and the captcha they were asked to
// solve would have no job left to resume.

// newJobInStatus creates a job and forces it into the requested status, bypassing
// the guard so a test can set up a state the guard itself would refuse to reach.
func newJobInStatus(t *testing.T, db *gorm.DB, qm *jobs.QueueManager, status models.JobStatus) *models.Job {
	t.Helper()
	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeScrape,
		Data: `{"mode":"search","query":"kaos","extension_id":"abc"}`,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}
	if status == models.JobStatusPending {
		return job
	}
	if err := db.Model(&models.Job{}).Where("id = ?", job.ID).
		Update("status", status).Error; err != nil {
		t.Fatalf("seed status %s: %v", status, err)
	}
	job.Status = status
	return job
}

func statusOf(t *testing.T, qm *jobs.QueueManager, jobID string) models.JobStatus {
	t.Helper()
	got, err := qm.GetJob(jobID)
	if err != nil {
		t.Fatalf("GetJob error: %v", err)
	}
	return got.Status
}

// TestUpdateStatus_RunningToBlockedIsAllowed pins the transition the scraper needs
// when Shopee serves a captcha mid-run.
func TestUpdateStatus_RunningToBlockedIsAllowed(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusRunning)

	if err := qm.UpdateStatus(job.ID, models.JobStatusBlocked, "captcha on page 4"); err != nil {
		t.Fatalf("running -> blocked must be allowed, got error: %v", err)
	}
	if got := statusOf(t, qm, job.ID); got != models.JobStatusBlocked {
		t.Errorf("expected status=%s, got %s", models.JobStatusBlocked, got)
	}
}

// TestUpdateStatus_BlockedDoesNotStampCompletedAt guards the resume cursor's
// meaning: a blocked job has not completed, so a completion timestamp would make
// it look finished to every query that filters on completed_at.
func TestUpdateStatus_BlockedDoesNotStampCompletedAt(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusRunning)

	if err := qm.UpdateStatus(job.ID, models.JobStatusBlocked, "captcha"); err != nil {
		t.Fatalf("UpdateStatus error: %v", err)
	}

	got, err := qm.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob error: %v", err)
	}
	if got.CompletedAt != nil {
		t.Errorf("blocked job must not have completed_at, got %v", *got.CompletedAt)
	}
}

// TestUpdateStatus_BlockedToPendingIsAllowed is the resume path itself.
func TestUpdateStatus_BlockedToPendingIsAllowed(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusBlocked)

	if err := qm.UpdateStatus(job.ID, models.JobStatusPending, ""); err != nil {
		t.Fatalf("blocked -> pending (resume) must be allowed, got error: %v", err)
	}
	if got := statusOf(t, qm, job.ID); got != models.JobStatusPending {
		t.Errorf("expected status=%s, got %s", models.JobStatusPending, got)
	}
}

// TestUpdateStatus_RejectsTerminalOverwrite is the core defect this guard exists
// for: without it, any caller can move a finished job back into flight.
func TestUpdateStatus_RejectsTerminalOverwrite(t *testing.T) {
	cases := []struct {
		name string
		from models.JobStatus
		to   models.JobStatus
	}{
		{"completed to running", models.JobStatusCompleted, models.JobStatusRunning},
		{"failed to completed", models.JobStatusFailed, models.JobStatusCompleted},
		{"cancelled to running", models.JobStatusCancelled, models.JobStatusRunning},
		{"completed to blocked", models.JobStatusCompleted, models.JobStatusBlocked},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			db := setupJobsTestDB(t)
			qm := jobs.NewQueueManager(db, "test-tenant")
			job := newJobInStatus(t, db, qm, tc.from)

			err := qm.UpdateStatus(job.ID, tc.to, "")
			if err == nil {
				t.Errorf("%s -> %s must be rejected, got nil error", tc.from, tc.to)
			}
			if got := statusOf(t, qm, job.ID); got != tc.from {
				t.Errorf("status must stay %s after a rejected transition, got %s", tc.from, got)
			}
		})
	}
}

// TestUpdateStatus_RejectsBlockedToCompleted stops a blocked scrape from being
// reported as a clean success — the exact lie this whole change removes.
func TestUpdateStatus_RejectsBlockedToCompleted(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusBlocked)

	if err := qm.UpdateStatus(job.ID, models.JobStatusCompleted, ""); err == nil {
		t.Error("blocked -> completed must be rejected, got nil error")
	}
	if got := statusOf(t, qm, job.ID); got != models.JobStatusBlocked {
		t.Errorf("expected status to stay %s, got %s", models.JobStatusBlocked, got)
	}
}

// TestCancelJob_CancelsBlockedJob lets an operator abandon a captcha they do not
// want to solve. Without it a blocked job can only be waited out.
func TestCancelJob_CancelsBlockedJob(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusBlocked)

	if err := qm.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob error: %v", err)
	}
	if got := statusOf(t, qm, job.ID); got != models.JobStatusCancelled {
		t.Errorf("expected status=%s, got %s", models.JobStatusCancelled, got)
	}
}

// TestCheckAndTimeoutStuckJobs_LeavesBlockedAlone pins the sweeper's scope. A
// blocked job is waiting on a human, so elapsed time is expected and must not be
// read as a hung worker.
func TestCheckAndTimeoutStuckJobs_LeavesBlockedAlone(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusBlocked)

	// Start it well before any plausible cutoff so only the status filter can
	// spare it.
	old := time.Now().Add(-90 * time.Minute)
	if err := db.Model(&models.Job{}).Where("id = ?", job.ID).
		Update("started_at", &old).Error; err != nil {
		t.Fatalf("seed started_at: %v", err)
	}

	qm.CheckAndTimeoutStuckJobs(30)

	if got := statusOf(t, qm, job.ID); got != models.JobStatusBlocked {
		t.Errorf("blocked job must survive the stuck-job sweep, got %s", got)
	}
}

// TestCompleteJobWithResult_RejectsBlockedJob closes the other route to the same
// lie: the executor's success path must not overwrite a blocked job either.
func TestCompleteJobWithResult_RejectsBlockedJob(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")
	job := newJobInStatus(t, db, qm, models.JobStatusBlocked)

	if err := qm.CompleteJobWithResult(job.ID, map[string]any{"products": 0}); err == nil {
		t.Error("CompleteJobWithResult on a blocked job must be rejected, got nil error")
	}
	if got := statusOf(t, qm, job.ID); got != models.JobStatusBlocked {
		t.Errorf("expected status to stay %s, got %s", models.JobStatusBlocked, got)
	}
}
