package jobs_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
)

// The executor's terminal-status decision.
//
// executeJob had exactly two outcomes: nil error → completed, any error →
// failed. That is why a blocked scrape lands as `failed` in the database, and
// why the resume cursor the handler put in its return string is discarded — the
// summary is written by CompleteJobWithResult, which the failure branch never
// reaches. These tests pin the invariants that let a handler carry more than a
// boolean up to the executor.

// TestExecutor_HandlerCompletedSentinelSkipsTerminalWrite pins the mechanism the
// handler needs to say "I recorded the outcome myself; do not overwrite it".
//
// Without it the executor's own CompleteJobWithResult/FailJob call clobbers the
// blocked status the handler just wrote. This is invariant number one for
// blocked jobs to reach the dashboard at all.
func TestExecutor_HandlerCompletedSentinelSkipsTerminalWrite(t *testing.T) {
	if jobs.ErrHandlerCompleted == nil {
		t.Fatal("ErrHandlerCompleted must be an exported sentinel for handlers to signal a self-recorded outcome")
	}
	if !errors.Is(fmt.Errorf("wrapped: %w", jobs.ErrHandlerCompleted), jobs.ErrHandlerCompleted) {
		t.Error("ErrHandlerCompleted must be recognisable through fmt.Errorf wrapping, or callers cannot annotate it")
	}
	msg := jobs.ErrHandlerCompleted.Error()
	if !strings.Contains(strings.ToLower(msg), "handler") {
		t.Errorf("sentinel message %q should name what it means", msg)
	}
}

// TestExecutor_ClassifyHandlerResultBranchesOnSentinel is the pure-logic test for
// the decision the executor runs. Kept small on purpose: it does not need a DB,
// only the intent it will apply.
func TestExecutor_ClassifyHandlerResultBranchesOnSentinel(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want jobs.HandlerOutcome
	}{
		{"nil is completion", nil, jobs.HandlerOutcomeCompleted},
		{"plain error is failure", errors.New("boom"), jobs.HandlerOutcomeFailed},
		{"sentinel is self-recorded", jobs.ErrHandlerCompleted, jobs.HandlerOutcomeSelfRecorded},
		{"wrapped sentinel is self-recorded", fmt.Errorf("resume cursor at 4: %w", jobs.ErrHandlerCompleted), jobs.HandlerOutcomeSelfRecorded},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := jobs.ClassifyHandlerResult(tc.err); got != tc.want {
				t.Errorf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestExecutor_SelfRecordedNeverTouchesJob is the property that closes the loop:
// a handler that already wrote its own terminal state must not be second-guessed.
func TestExecutor_SelfRecordedNeverTouchesJob(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "")

	job, err := qm.AddJob(models.CreateJobRequest{Type: models.JobTypeShopeeScrape, Data: `{}`})
	if err != nil {
		t.Fatalf("AddJob: %v", err)
	}
	if err := qm.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		t.Fatalf("running: %v", err)
	}
	if err := qm.UpdateStatus(job.ID, models.JobStatusBlocked, "captcha"); err != nil {
		t.Fatalf("blocked: %v", err)
	}

	// The handler says: I wrote the outcome, do not touch it.
	jobs.ApplyHandlerOutcome(qm, job.ID, "ignored summary", jobs.HandlerOutcomeSelfRecorded)

	got, err := qm.GetJob(job.ID)
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if got.Status != models.JobStatusBlocked {
		t.Errorf("expected status to stay blocked, got %s — the executor overwrote the handler's own write", got.Status)
	}
}

// TestExecutor_CompletedWritesResultData keeps the ordinary path honest: the
// self-recorded branch must not accidentally silence the normal one.
//
// The stored value MUST be parseable JSON that reveals the summary's fields
// directly, not a JSON string containing more JSON. The dashboard's
// `parseScrapeSummary` does one JSON.parse and expects an object; a double-
// encoded value parses to a string and drops on the floor as null.
func TestExecutor_CompletedWritesResultData(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "")

	job, _ := qm.AddJob(models.CreateJobRequest{Type: models.JobTypeShopeeScrape, Data: `{}`})
	_ = qm.UpdateStatus(job.ID, models.JobStatusRunning, "")

	jobs.ApplyHandlerOutcome(qm, job.ID, `{"pages":3,"reason":"last_page"}`, jobs.HandlerOutcomeCompleted)

	got, _ := qm.GetJob(job.ID)
	if got.Status != models.JobStatusCompleted {
		t.Errorf("expected completed, got %s", got.Status)
	}
	// Parse once, as the dashboard does, and verify the fields are directly
	// accessible. A double-encoded value would parse to a string here.
	var summary map[string]any
	if err := json.Unmarshal([]byte(got.ResultData), &summary); err != nil {
		t.Fatalf("result_data is not parseable JSON: %v (raw=%q)", err, got.ResultData)
	}
	if pages, ok := summary["pages"].(float64); !ok || pages != 3 {
		t.Errorf("summary.pages missing after one JSON.parse — result_data is probably double-encoded: %#v", summary)
	}
}

// TestExecutor_FailedWritesErrorMessage pins the failure branch.
func TestExecutor_FailedWritesErrorMessage(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "")

	job, _ := qm.AddJob(models.CreateJobRequest{Type: models.JobTypeShopeeScrape, Data: `{}`})
	_ = qm.UpdateStatus(job.ID, models.JobStatusRunning, "")

	jobs.ApplyHandlerOutcome(qm, job.ID, "boom", jobs.HandlerOutcomeFailed)

	got, _ := qm.GetJob(job.ID)
	if got.Status != models.JobStatusFailed {
		t.Errorf("expected failed, got %s", got.Status)
	}
	if got.ErrorMessage != "boom" {
		t.Errorf("expected error_message=boom, got %q", got.ErrorMessage)
	}
}

// Placate the linter about unused ctx.
var _ = context.Background
