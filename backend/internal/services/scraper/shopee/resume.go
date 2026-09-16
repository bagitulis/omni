package shopee

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Resuming a scrape that stopped on a human-clearable obstacle.
//
// A blocked run holds a cursor: the page it never captured. Resuming re-queues
// the same job from that page instead of starting over, so an operator who
// solved a captcha on page 40 does not pay for pages 1-39 again.

// ErrJobNotFound reports that the job does not exist in the tenant's schema.
//
// A distinct error so the handler can answer 404 without matching on message
// text, and so a typo'd id is never reported as a successful resume.
var ErrJobNotFound = errors.New("job not found")

// ResumeOutcome describes what a resume attempt did.
//
// Resumed is false for a job that was not blocked. That is a no-op, not a
// failure: the dashboard polls, and two operators pressing Resume must not
// produce an error the second time.
type ResumeOutcome struct {
	JobID     string
	Status    string
	StartPage int
	Resumed   bool
}

// Resume moves a blocked job back into the queue, starting at its recorded page.
func (s *ScrapeService) Resume(ctx context.Context, tenantID, jobID string) (ResumeOutcome, error) {
	if tenantID == "" {
		// Fail closed: with no tenant there is no schema holding the job, and
		// guessing one would re-queue another tenant's work.
		return ResumeOutcome{}, fmt.Errorf("scrape: tenant_id is required")
	}
	if jobID == "" {
		return ResumeOutcome{}, fmt.Errorf("scrape: job_id is required")
	}
	if s == nil || s.tenantDB == nil {
		return ResumeOutcome{}, fmt.Errorf("scrape: resume is unavailable: no tenant database resolver")
	}

	db, err := s.tenantDB(tenantID)
	if err != nil {
		return ResumeOutcome{}, fmt.Errorf("scrape: tenant database: %w", err)
	}

	queue := jobs.NewQueueManager(db, "")
	job, err := queue.GetJob(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResumeOutcome{}, ErrJobNotFound
		}
		return ResumeOutcome{}, fmt.Errorf("scrape: load job %s: %w", jobID, err)
	}

	if job.Status != models.JobStatusBlocked {
		return ResumeOutcome{JobID: job.ID, Status: string(job.Status)}, nil
	}

	payload, startPage, err := resumePayload(job)
	if err != nil {
		return ResumeOutcome{}, err
	}

	// Both writes go in one transaction: without it, a failing status update
	// after a successful payload update would strand the job as `blocked` with a
	// cursor already shifted forward, and a second resume would skip the page
	// the operator just cleared. The status update is filtered on `blocked` too,
	// so a concurrent cancel between the SELECT and the UPDATE is respected.
	var affected int64
	err = db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&models.Job{}).
			Where("id = ? AND status = ?", job.ID, models.JobStatusBlocked).
			Updates(map[string]any{
				"data":       payload,
				"status":     models.JobStatusPending,
				"updated_at": time.Now(),
			})
		if res.Error != nil {
			return res.Error
		}
		affected = res.RowsAffected
		return nil
	})
	if err != nil {
		return ResumeOutcome{}, fmt.Errorf("scrape: requeue job %s: %w", jobID, err)
	}
	if affected == 0 {
		// Something moved the job out of blocked between the read and the write
		// — a cancel, or another resume. Report the current state as a no-op
		// rather than forcing it back into the queue.
		return currentStateOf(queue, jobID)
	}

	log.Info().
		Str("tenant_id", tenantID).
		Str("job_id", jobID).
		Int("start_page", startPage).
		Msg("Resumed blocked scrape job")

	return ResumeOutcome{
		JobID:     jobID,
		Status:    string(models.JobStatusPending),
		StartPage: startPage,
		Resumed:   true,
	}, nil
}

// resumePayload rebuilds a job's payload with the resume cursor applied.
func resumePayload(job *models.Job) (string, int, error) {
	var data ScrapeJobData
	if err := json.Unmarshal([]byte(job.Data), &data); err != nil {
		// Rewriting a payload that cannot be parsed would replace the job's data
		// with a guess, so the job stays blocked instead.
		return "", 0, fmt.Errorf("scrape: parse job payload for %s: %w", job.ID, err)
	}

	data.StartPage = resumeCursorFrom(job.ResultData)
	if err := validateStartPage(&data); err != nil {
		return "", 0, err
	}

	raw, err := json.Marshal(data)
	if err != nil {
		return "", 0, fmt.Errorf("scrape: build resume payload for %s: %w", job.ID, err)
	}
	return string(raw), data.StartPage, nil
}

// currentStateOf reports a job's status as a no-op outcome.
func currentStateOf(queue *jobs.QueueManager, jobID string) (ResumeOutcome, error) {
	job, err := queue.GetJob(jobID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ResumeOutcome{}, ErrJobNotFound
		}
		return ResumeOutcome{}, fmt.Errorf("scrape: reload job %s: %w", jobID, err)
	}
	return ResumeOutcome{JobID: job.ID, Status: string(job.Status)}, nil
}

// resumeCursorFrom reads the resume page out of a stored job summary.
//
// Anything unreadable — no summary, no cursor, a cursor the scraper could not
// honour — falls back to page 1. A job blocked before it recorded a cursor is
// still worth resuming; refusing would strand it as blocked with cancellation as
// its only exit.
func resumeCursorFrom(resultData string) int {
	var summary struct {
		ResumeFromPage int `json:"resume_from_page"`
	}
	if err := json.Unmarshal([]byte(resultData), &summary); err != nil {
		return 1
	}
	return normaliseStartPage(summary.ResumeFromPage)
}
