package jobs

import (
	"errors"

	"github.com/omni/backend/internal/models"
)

// A job handler's outcome.
//
// The executor used to translate a handler return into exactly two states —
// completed or failed — which is what silently turned a blocked scrape into a
// failure. A handler that already recorded its own terminal state (a blocked
// scrape holds a resume cursor the generic completion path cannot express)
// signals it with ErrHandlerCompleted, and the executor leaves the row alone.

// ErrHandlerCompleted signals that the handler has already written the job's
// terminal state, so the executor must not overwrite it.
//
// Wrapping is supported: a caller may annotate the sentinel with context (the
// resume cursor, the blocker kind) and the executor still recognises it through
// errors.Is.
var ErrHandlerCompleted = errors.New("job handler recorded its own outcome")

// HandlerOutcome names what the executor should do after the handler returns.
type HandlerOutcome int

const (
	HandlerOutcomeCompleted HandlerOutcome = iota
	HandlerOutcomeFailed
	// HandlerOutcomeSelfRecorded is the "leave it alone" case: the handler wrote
	// its own status and any executor write would clobber it.
	HandlerOutcomeSelfRecorded
)

// ClassifyHandlerResult maps a handler's error return to the executor's action.
func ClassifyHandlerResult(err error) HandlerOutcome {
	if err == nil {
		return HandlerOutcomeCompleted
	}
	if errors.Is(err, ErrHandlerCompleted) {
		return HandlerOutcomeSelfRecorded
	}
	return HandlerOutcomeFailed
}

// ApplyHandlerOutcome writes the executor's terminal state for a handler run.
//
// Split out of executeJob so the branch can be tested against a real
// QueueManager without spinning the executor up: what matters is which write
// happens, not the goroutine it happens in.
func ApplyHandlerOutcome(qm *QueueManager, jobID, summary string, outcome HandlerOutcome) {
	switch outcome {
	case HandlerOutcomeSelfRecorded:
		// Deliberately nothing. The handler owns this row now.
		return
	case HandlerOutcomeCompleted:
		_ = qm.CompleteJobWithResult(jobID, summary)
	case HandlerOutcomeFailed:
		_ = qm.FailJob(jobID, summary)
	}
}

// Compile-time nudge: the outcome type is exhaustive on purpose. If a new value
// is added, the switch above will still compile, but tests that enumerate the
// set (executor_outcome_test.go) will catch it.
var _ models.JobStatus = models.JobStatusPending
