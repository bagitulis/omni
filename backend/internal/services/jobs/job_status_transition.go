package jobs

import (
	"fmt"

	"github.com/omni/backend/internal/models"
)

// Job status transitions.
//
// Status is written from several places — the multi-tenant executor, the
// single-tenant executor, the cancel endpoints, the zombie sweep, the stuck-job
// sweep — and none of them used to check what the job's status already was. That
// was survivable while every status was terminal-or-running. It stops being
// survivable with `blocked`: a blocked scrape holds the resume cursor an operator
// is about to act on, and an unguarded write would discard it while reporting
// success.

// allowedTransitions maps a current status to the statuses it may move to.
//
// Terminal statuses map to nothing: once a job is completed, failed, or
// cancelled, its outcome is a fact and later writes are bugs, not updates.
var allowedTransitions = map[models.JobStatus][]models.JobStatus{
	models.JobStatusPending: {
		models.JobStatusRunning,
		models.JobStatusCancelled,
		models.JobStatusFailed,
	},
	models.JobStatusRunning: {
		// running -> running is permitted because ClaimNextJob already sets
		// running when it claims the row, and executeJob sets it again. Refusing
		// the repeat would break every job at the moment it starts.
		models.JobStatusRunning,
		models.JobStatusCompleted,
		models.JobStatusFailed,
		models.JobStatusCancelled,
		models.JobStatusBlocked,
	},
	models.JobStatusBlocked: {
		// pending is the resume path: the operator cleared the obstacle and the
		// job goes back in the queue with its start page.
		models.JobStatusPending,
		models.JobStatusCancelled,
		models.JobStatusFailed,
	},
	models.JobStatusCompleted: {},
	models.JobStatusFailed:    {},
	models.JobStatusCancelled: {},
}

// transitionAllowed reports whether from -> to is a legal move.
func transitionAllowed(from, to models.JobStatus) bool {
	for _, allowed := range allowedTransitions[from] {
		if allowed == to {
			return true
		}
	}
	return false
}

// legalSources returns every status that may move to the target status.
//
// Used to build a WHERE clause so the check and the write are one statement.
// Reading the status first and updating second would leave a window in which a
// concurrent sweep changes it, which is exactly the class of bug this guards.
func legalSources(to models.JobStatus) []models.JobStatus {
	var sources []models.JobStatus
	for from := range allowedTransitions {
		if transitionAllowed(from, to) {
			sources = append(sources, from)
		}
	}
	return sources
}

// errIllegalTransition describes a refused move.
//
// It names the current status because the caller almost always wants to know what
// beat it to the write — a timeout sweep, a cancel, or another worker.
func errIllegalTransition(jobID string, current, to models.JobStatus) error {
	return fmt.Errorf("job %s: illegal status transition %s -> %s", jobID, current, to)
}
