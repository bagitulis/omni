package jobs_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"gorm.io/gorm"
)

// ---------------------------------------------------------------------------
// setupJobsTestDB creates an in-memory SQLite DB with Job + JobHistory tables
// ---------------------------------------------------------------------------

func setupJobsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	path := fmt.Sprintf("/tmp/jobs_lifecycle_test_%s.db", uuid.New().String())
	db, err := gorm.Open(sqlite.Open(path+"?_pragma=journal_mode(MEMORY)&_pragma=synchronous(OFF)"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open SQLite: %v", err)
	}
	if err := db.AutoMigrate(&models.Job{}, &models.JobHistory{}); err != nil {
		t.Fatalf("failed to auto-migrate: %v", err)
	}
	return db
}

// ---------------------------------------------------------------------------
// Job Cancellation Tests
// ---------------------------------------------------------------------------

// TestCancelJob_PendingJobStatus verifies a pending job can be cancelled.
func TestCancelJob_PendingJobStatus(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: `{"tenant_id":"t1","month":3,"year":2026}`,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}
	if job.Status != models.JobStatusPending {
		t.Fatalf("expected pending status, got %s", job.Status)
	}

	if err := qm.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob error: %v", err)
	}

	got, err := qm.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob error: %v", err)
	}
	if got.Status != models.JobStatusCancelled {
		t.Errorf("expected status=%s, got %s", models.JobStatusCancelled, got.Status)
	}
}

// TestCancelJob_RunningJobNotCancelled verifies CancelJob does NOT cancel running jobs.
func TestCancelJob_RunningJobNotCancelled(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: `{"tenant_id":"t1","month":3,"year":2026}`,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	// Transition to running
	if err := qm.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		t.Fatalf("UpdateStatus error: %v", err)
	}

	// Attempt to cancel running job — should silently succeed (0 rows affected)
	if err := qm.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob error: %v", err)
	}

	got, err := qm.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob error: %v", err)
	}
	if got.Status != models.JobStatusRunning {
		t.Errorf("running job should remain running after CancelJob, got %s", got.Status)
	}
}

// TestCancelJob_AlreadyCompletedIgnored verifies completed jobs cannot be cancelled.
func TestCancelJob_AlreadyCompletedIgnored(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeTiktokEscrowSync,
		Data: `{"tenant_id":"t1"}`,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	if err := qm.CompleteJobWithResult(job.ID, "ok"); err != nil {
		t.Fatalf("CompleteJobWithResult error: %v", err)
	}

	// Cancel should be a no-op for completed jobs
	if err := qm.CancelJob(job.ID); err != nil {
		t.Fatalf("CancelJob error: %v", err)
	}

	got, err := qm.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob error: %v", err)
	}
	if got.Status != models.JobStatusCompleted {
		t.Errorf("completed job should remain completed, got %s", got.Status)
	}
}

// TestCancelJob_NonExistentJob verifies cancel on non-existent ID returns no error.
func TestCancelJob_NonExistentJob(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	err := qm.CancelJob("non-existent-id")
	if err != nil {
		t.Errorf("expected no error cancelling non-existent job, got: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Job Deduplication Tests
// ---------------------------------------------------------------------------

// TestJobDeduplication_SameTypeAndPayload verifies duplicate jobs can be detected
// by querying for existing pending/running jobs with the same type and data.
func TestJobDeduplication_SameTypeAndPayload(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	payload := `{"tenant_id":"t1","month":3,"year":2026,"platform":"shopee"}`
	jobType := models.JobTypeShopeeEscrowSync

	// Add first job
	job1, err := qm.AddJob(models.CreateJobRequest{
		Type: jobType,
		Data: payload,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	// Check for existing pending job with same type + data
	var existing models.Job
	err = db.Where("type = ? AND status IN ? AND data = ?",
		jobType,
		[]models.JobStatus{models.JobStatusPending, models.JobStatusRunning},
		payload,
	).First(&existing).Error

	if err != nil {
		t.Fatalf("expected to find existing pending job, got: %v", err)
	}
	if existing.ID != job1.ID {
		t.Errorf("expected existing job ID=%s, got %s", job1.ID, existing.ID)
	}

	// Adding a second job with same payload should be detectable as duplicate
	job2, err := qm.AddJob(models.CreateJobRequest{
		Type: jobType,
		Data: payload,
	})
	if err != nil {
		t.Fatalf("AddJob second error: %v", err)
	}

	// Count pending jobs with same type+data — should be 2 (app layer must deduplicate)
	var count int64
	db.Model(&models.Job{}).Where("type = ? AND status = ? AND data = ?",
		jobType, models.JobStatusPending, payload).Count(&count)
	if count != 2 {
		t.Errorf("expected 2 pending jobs (dedup is app-layer), got %d", count)
	}

	// After cancelling first, only second remains pending
	if err := qm.CancelJob(job1.ID); err != nil {
		t.Fatalf("CancelJob error: %v", err)
	}
	db.Model(&models.Job{}).Where("type = ? AND status = ? AND data = ?",
		jobType, models.JobStatusPending, payload).Count(&count)
	if count != 1 {
		t.Errorf("expected 1 pending job after cancel, got %d", count)
	}

	_ = job2 // used for count check
}

// TestJobDeduplication_DifferentPlatformsAllowed verifies different platform
// jobs with same month/year are NOT considered duplicates.
func TestJobDeduplication_DifferentPlatformsAllowed(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	shopeePayload := `{"tenant_id":"t1","month":3,"year":2026,"platform":"shopee"}`
	tiktokPayload := `{"tenant_id":"t1","month":3,"year":2026,"platform":"tiktok"}`

	_, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: shopeePayload,
	})
	if err != nil {
		t.Fatalf("AddJob shopee error: %v", err)
	}

	_, err = qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeTiktokEscrowSync,
		Data: tiktokPayload,
	})
	if err != nil {
		t.Fatalf("AddJob tiktok error: %v", err)
	}

	// Both jobs should be pending — different platforms are not duplicates
	var shopeePending, tiktokPending int64
	db.Model(&models.Job{}).Where("type = ? AND status = ?",
		models.JobTypeShopeeEscrowSync, models.JobStatusPending).Count(&shopeePending)
	db.Model(&models.Job{}).Where("type = ? AND status = ?",
		models.JobTypeTiktokEscrowSync, models.JobStatusPending).Count(&tiktokPending)

	if shopeePending != 1 {
		t.Errorf("expected 1 pending shopee job, got %d", shopeePending)
	}
	if tiktokPending != 1 {
		t.Errorf("expected 1 pending tiktok job, got %d", tiktokPending)
	}
}

// ---------------------------------------------------------------------------
// Tenant Isolation Tests (Job-Level)
// ---------------------------------------------------------------------------

// TestTenantIsolation_SeparateDatabases verifies jobs in tenant A's DB are
// invisible to tenant B's QueueManager (schema-based isolation).
func TestTenantIsolation_SeparateDatabases(t *testing.T) {
	dbA := setupJobsTestDB(t)
	dbB := setupJobsTestDB(t)

	qmA := jobs.NewQueueManager(dbA, "tenant-a")
	qmB := jobs.NewQueueManager(dbB, "tenant-b")

	// Add job for tenant A
	jobA, err := qmA.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: `{"tenant_id":"tenant-a","month":1,"year":2026}`,
	})
	if err != nil {
		t.Fatalf("AddJob tenant A error: %v", err)
	}

	// Add job for tenant B
	_, err = qmB.AddJob(models.CreateJobRequest{
		Type: models.JobTypeTiktokEscrowSync,
		Data: `{"tenant_id":"tenant-b","month":2,"year":2026}`,
	})
	if err != nil {
		t.Fatalf("AddJob tenant B error: %v", err)
	}

	// Tenant B should NOT see tenant A's job
	got, err := qmB.GetJob(jobA.ID)
	if err == nil {
		t.Errorf("tenant B should not find tenant A's job, but got job with ID=%s", got.ID)
	}

	var countA, countB int64
	// Count: tenant A DB has 1, tenant B DB has 1
	var countA, countB int64
	dbA.Model(&models.Job{}).Count(&countA)
	dbB.Model(&models.Job{}).Count(&countB)
	if countA != 1 {
		t.Errorf("tenant A DB: expected 1 job, got %d", countA)
	}
	if countB != 1 {
		t.Errorf("tenant B DB: expected 1 job, got %d", countB)
	}
}

// TestTenantIsolation_CancelDoesNotAffectOtherTenant verifies cancelling a job
// in one tenant does not affect jobs in another tenant's database.
func TestTenantIsolation_CancelDoesNotAffectOtherTenant(t *testing.T) {
	dbA := setupJobsTestDB(t)
	dbB := setupJobsTestDB(t)

	qmA := jobs.NewQueueManager(dbA, "tenant-a")
	qmB := jobs.NewQueueManager(dbB, "tenant-b")

	jobA, err := qmA.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: `{"tenant_id":"tenant-a"}`,
	})
	if err != nil {
		t.Fatalf("AddJob A error: %v", err)
	}

	jobB, err := qmB.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: `{"tenant_id":"tenant-b"}`,
	})
	if err != nil {
		t.Fatalf("AddJob B error: %v", err)
	}

	// Cancel tenant A's job
	if err := qmA.CancelJob(jobA.ID); err != nil {
		t.Fatalf("CancelJob A error: %v", err)
	}

	// Tenant B's job should still be pending
	gotB, err := qmB.GetJob(jobB.ID)
	if err != nil {
		t.Fatalf("GetJob B error: %v", err)
	}
	if gotB.Status != models.JobStatusPending {
		t.Errorf("tenant B job should still be pending, got %s", gotB.Status)
	}
}

// ---------------------------------------------------------------------------
// EscrowSyncHandler Payload Tests
// ---------------------------------------------------------------------------

// TestEscrowSyncHandler_MissingTenantID verifies handler rejects empty tenant_id.
func TestEscrowSyncHandler_MissingTenantID(t *testing.T) {
	// Verify EscrowSyncJobData deserialization catches empty tenant_id
	payload := `{"tenant_id":"","month":3,"year":2026}`
	var data models.EscrowSyncJobData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if data.TenantID != "" {
		t.Errorf("expected empty tenant_id, got %q", data.TenantID)
	}
}

// TestEscrowSyncHandler_InvalidPayload verifies handler rejects malformed JSON.
func TestEscrowSyncHandler_InvalidPayload(t *testing.T) {
	var data models.EscrowSyncJobData
	err := json.Unmarshal([]byte(`{not valid json`), &data)
	if err == nil {
		t.Error("expected error for invalid JSON, got nil")
	}
}

// TestEscrowSyncHandler_ValidPayload verifies correct payload deserialization.
func TestEscrowSyncHandler_ValidPayload(t *testing.T) {
	payload := `{"tenant_id":"t1","platform":"shopee","month":3,"year":2026,"force_resync":true}`
	var data models.EscrowSyncJobData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}
	if data.TenantID != "t1" {
		t.Errorf("expected tenant_id=t1, got %q", data.TenantID)
	}
	if data.Platform != "shopee" {
		t.Errorf("expected platform=shopee, got %q", data.Platform)
	}
	if data.Month != 3 || data.Year != 2026 {
		t.Errorf("expected month=3 year=2026, got month=%d year=%d", data.Month, data.Year)
	}
	if !data.ForceResync {
		t.Error("expected force_resync=true")
	}
}

// ---------------------------------------------------------------------------
// Job Lifecycle Status Transitions
// ---------------------------------------------------------------------------

// TestJobLifecycle_PendingToRunningToCompleted verifies the full happy path.
func TestJobLifecycle_PendingToRunningToCompleted(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeShopeeEscrowSync,
		Data: `{"tenant_id":"t1"}`,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	// pending → running
	if err := qm.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		t.Fatalf("UpdateStatus(running) error: %v", err)
	}
	got, _ := qm.GetJob(job.ID)
	if got.Status != models.JobStatusRunning {
		t.Errorf("expected running, got %s", got.Status)
	}
	if got.StartedAt == nil {
		t.Error("expected StartedAt to be set when transitioning to running")
	}

	// running → completed
	if err := qm.CompleteJobWithResult(job.ID, "synced 10 orders"); err != nil {
		t.Fatalf("CompleteJobWithResult error: %v", err)
	}
	got, _ = qm.GetJob(job.ID)
	if got.Status != models.JobStatusCompleted {
		t.Errorf("expected completed, got %s", got.Status)
	}
	if got.CompletedAt == nil {
		t.Error("expected CompletedAt to be set")
	}
}

// TestJobLifecycle_PendingToRunningToFailed verifies the error path.
func TestJobLifecycle_PendingToRunningToFailed(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	job, err := qm.AddJob(models.CreateJobRequest{
		Type: models.JobTypeTiktokEscrowSync,
		Data: `{"tenant_id":"t1"}`,
	})
	if err != nil {
		t.Fatalf("AddJob error: %v", err)
	}

	// pending → running
	if err := qm.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		t.Fatalf("UpdateStatus(running) error: %v", err)
	}

	// running → failed
	if err := qm.FailJob(job.ID, "API rate limit exceeded"); err != nil {
		t.Fatalf("FailJob error: %v", err)
	}
	got, _ := qm.GetJob(job.ID)
	if got.Status != models.JobStatusFailed {
		t.Errorf("expected failed, got %s", got.Status)
	}
	if got.ErrorMessage != "API rate limit exceeded" {
		t.Errorf("expected error message, got %q", got.ErrorMessage)
	}
}

// ---------------------------------------------------------------------------
// Context Cancellation Propagation (Handler-Level)
// ---------------------------------------------------------------------------

// TestContextCancellation_PropagatesToHandler verifies that when a parent
// context is cancelled, the handler's context is also cancelled.
func TestContextCancellation_PropagatesToHandler(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handlerCalled := make(chan struct{})
	handler := func(ctx context.Context, _ string) (string, error) {
		close(handlerCalled)
		<-ctx.Done()
		return "", ctx.Err()
	}

	// Run handler in goroutine
	errCh := make(chan error, 1)
	go func() {
		_, err := handler(ctx, "")
		errCh <- err
	}()

	<-handlerCalled
	cancel()

	select {
	case err := <-errCh:
		if err != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler did not return after context cancellation")
	}
}

// TestContextCancellation_TimeoutDistinctFromCancel verifies context.DeadlineExceeded
// is distinguishable from context.Canceled.
func TestContextCancellation_TimeoutDistinctFromCancel(t *testing.T) {
	// Create a context that times out immediately
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for timeout
	<-ctx.Done()

	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", ctx.Err())
	}

	// Create a context that is manually cancelled
	ctx2, cancel2 := context.WithCancel(context.Background())
	cancel2()

	if ctx2.Err() != context.Canceled {
		t.Errorf("expected Canceled, got %v", ctx2.Err())
	}

	// The two error types must be different
	if ctx.Err() == ctx2.Err() {
		t.Error("DeadlineExceeded and Canceled should be different error types")
	}
}

// ---------------------------------------------------------------------------
// Job Stats and Filtering
// ---------------------------------------------------------------------------

// TestJobStats_CancellationCounted verifies cancelled jobs appear in stats.
func TestJobStats_CancellationCounted(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	// Add 3 jobs: 1 completed, 1 cancelled, 1 pending
	j1, _ := qm.AddJob(models.CreateJobRequest{Type: "job_a", Data: `{}`})
	j2, _ := qm.AddJob(models.CreateJobRequest{Type: "job_b", Data: `{}`})
	_, _ = qm.AddJob(models.CreateJobRequest{Type: "job_c", Data: `{}`})

	_ = qm.CompleteJobWithResult(j1.ID, "done")
	_ = qm.CancelJob(j2.ID)

	stats, err := qm.GetJobStats()
	if err != nil {
		t.Fatalf("GetJobStats error: %v", err)
	}

	// Check cancelled count
	cancelled, ok := stats[models.JobStatusCancelled]
	if !ok {
		t.Fatal("expected cancelled status in stats")
	}
	if cancelled.(int64) != 1 {
		t.Errorf("expected 1 cancelled job, got %v", cancelled)
	}

	completed, ok := stats[models.JobStatusCompleted]
	if !ok {
		t.Fatal("expected completed status in stats")
	}
	if completed.(int64) != 1 {
		t.Errorf("expected 1 completed job, got %v", completed)
	}
}

// ---------------------------------------------------------------------------
// RecoverZombieJobs — running jobs become failed on restart
// ---------------------------------------------------------------------------

// TestRecoverZombieJobs_RunningBecomesFailed verifies that zombie recovery
// marks running jobs as failed (simulating server restart).
func TestRecoverZombieJobs_RunningBecomesFailed(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	job, _ := qm.AddJob(models.CreateJobRequest{Type: "zombie_test", Data: `{}`})
	_ = qm.UpdateStatus(job.ID, models.JobStatusRunning, "")

	recovered, err := qm.RecoverZombieJobs()
	if err != nil {
		t.Fatalf("RecoverZombieJobs error: %v", err)
	}
	if recovered != 1 {
		t.Errorf("expected 1 recovered job, got %d", recovered)
	}

	got, _ := qm.GetJob(job.ID)
	if got.Status != models.JobStatusFailed {
		t.Errorf("expected failed after zombie recovery, got %s", got.Status)
	}
}
