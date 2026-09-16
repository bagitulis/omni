package jobs_test

import (
	"fmt"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
)

// Dry run for the status guard.
//
// The guard changes the behaviour of every status write in the codebase, and
// those writes run against live tenant schemas. This rehearses the change on a
// throwaway database and prints what WOULD be refused, so the blast radius is
// inspected before the code reaches a real one. It asserts as well as reports:
// a dry run that only prints can drift from what the guard actually does.

// TestDryRun_TransitionBlastRadius prints the full transition matrix and pins the
// two outcomes that matter operationally.
func TestDryRun_TransitionBlastRadius(t *testing.T) {
	targets := []models.JobStatus{
		models.JobStatusRunning,
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
		models.JobStatusBlocked,
		models.JobStatusPending,
	}

	for _, to := range targets {
		t.Log("\n" + jobs.DescribeTransitionPlan(to))
	}

	// Every terminal status must refuse every move. This is the invariant the
	// whole guard exists to hold.
	for _, from := range []models.JobStatus{
		models.JobStatusCompleted, models.JobStatusFailed, models.JobStatusCancelled,
	} {
		for _, to := range targets {
			for _, p := range jobs.PlanTransitions(to) {
				if p.From == from && p.Allowed {
					t.Errorf("terminal status %s must refuse a move to %s", from, to)
				}
			}
		}
	}
}

// TestDryRun_CountsAffectedJobsWithoutWriting rehearses the guard against a
// populated database and reports how many rows each write would touch, without
// performing any of the writes. This is the "how many entities would be
// affected" half of a dry run.
func TestDryRun_CountsAffectedJobsWithoutWriting(t *testing.T) {
	db := setupJobsTestDB(t)
	qm := jobs.NewQueueManager(db, "test-tenant")

	// One job in each status, mirroring a tenant queue mid-flight.
	seeded := map[models.JobStatus]string{}
	for _, status := range []models.JobStatus{
		models.JobStatusPending,
		models.JobStatusRunning,
		models.JobStatusBlocked,
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
	} {
		job := newJobInStatus(t, db, qm, status)
		seeded[status] = job.ID
	}

	// Rehearse: for each target status, count the seeded jobs the guard would
	// accept. No write is issued — PlanTransitions is pure.
	for _, to := range []models.JobStatus{
		models.JobStatusCompleted, models.JobStatusBlocked, models.JobStatusPending,
	} {
		affected := 0
		var refused []string
		for status := range seeded {
			allowed := false
			for _, p := range jobs.PlanTransitions(to) {
				if p.From == status && p.Allowed {
					allowed = true
				}
			}
			if allowed {
				affected++
			} else {
				refused = append(refused, string(status))
			}
		}
		t.Log(fmt.Sprintf("dry run: setting status=%s would affect %d of %d seeded jobs; refused from %v",
			to, affected, len(seeded), refused))
	}

	// Evidence of what did NOT happen: every seeded job still holds the status it
	// was seeded with, because the rehearsal issued no writes.
	for status, id := range seeded {
		got, err := qm.GetJob(id)
		if err != nil {
			t.Fatalf("GetJob(%s): %v", id, err)
		}
		if got.Status != status {
			t.Errorf("dry run mutated a job: %s is now %s, expected %s", id, got.Status, status)
		}
	}
}

// TestDryRun_MatchesRealGuard closes the gap a dry run can hide: the plan is only
// useful if it predicts what the live write actually does. Each prediction is
// executed for real here, on the throwaway database, and compared.
func TestDryRun_MatchesRealGuard(t *testing.T) {
	targets := []models.JobStatus{
		models.JobStatusRunning,
		models.JobStatusCompleted,
		models.JobStatusBlocked,
		models.JobStatusCancelled,
		models.JobStatusPending,
		models.JobStatusFailed,
	}
	sources := []models.JobStatus{
		models.JobStatusPending,
		models.JobStatusRunning,
		models.JobStatusBlocked,
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
	}

	for _, to := range targets {
		predicted := map[models.JobStatus]bool{}
		for _, p := range jobs.PlanTransitions(to) {
			predicted[p.From] = p.Allowed
		}

		for _, from := range sources {
			db := setupJobsTestDB(t)
			qm := jobs.NewQueueManager(db, "test-tenant")
			job := newJobInStatus(t, db, qm, from)

			err := qm.UpdateStatus(job.ID, to, "")
			actuallyAllowed := err == nil

			if actuallyAllowed != predicted[from] {
				t.Errorf("%s -> %s: dry run predicted allowed=%v, real write gave allowed=%v (err=%v)",
					from, to, predicted[from], actuallyAllowed, err)
			}
		}
	}
}
