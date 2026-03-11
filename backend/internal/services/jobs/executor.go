package jobs

import (
	"context"
	"encoding/json"
	"github.com/rs/zerolog/log"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"
)

// JobHandler is a function that processes a job
type JobHandler func(ctx context.Context, payload string) (string, error)

// Executor executes jobs from the queue
type Executor struct {
	queueManager *QueueManager
	handlers     map[string]JobHandler
	workerCount  int
	pollInterval time.Duration
	timeout      time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	running      bool
}

// NewExecutor creates a new job executor
func NewExecutor(qm *QueueManager, workerCount int) *Executor {
	return &Executor{
		queueManager: qm,
		handlers:     make(map[string]JobHandler),
		workerCount:  workerCount,
		pollInterval: 5 * time.Second,
		timeout:      5 * time.Minute,
		stopCh:       make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a job type
func (e *Executor) RegisterHandler(jobType string, handler JobHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[jobType] = handler
}

// Start starts the executor workers
func (e *Executor) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	for i := 0; i < e.workerCount; i++ {
		e.wg.Add(1)
		go e.worker(i)
	}

	log.Info().Msgf("Job executor started with %d workers", e.workerCount)
}

// Stop stops the executor gracefully
func (e *Executor) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	e.mu.Unlock()

	close(e.stopCh)
	e.wg.Wait()
	log.Info().Msg("Job executor stopped")
}

// worker processes jobs from the queue
func (e *Executor) worker(id int) {
	defer e.wg.Done()

	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.processNextJob()
		}
	}
}

// processNextJob processes the next available job
func (e *Executor) processNextJob() {
	jobs, err := e.queueManager.GetPendingJobs(1)
	if err != nil || len(jobs) == 0 {
		return
	}

	job := jobs[0]
	e.executeJob(&job)
}

// executeJob executes a single job
func (e *Executor) executeJob(job *models.Job) {
	e.mu.RLock()
	handler, ok := e.handlers[job.Type]
	e.mu.RUnlock()

	if !ok {
		e.queueManager.UpdateStatus(job.ID, models.JobStatusFailed, "no handler for job type: "+job.Type)
		return
	}

	// Mark as running
	if err := e.queueManager.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		log.Info().Msgf("Failed to update job status: %v", err)
		return
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	resultCh := make(chan executeResult, 1)
	go func() {
		result, err := handler(ctx, job.Data)
		resultCh <- executeResult{result: result, err: err}
	}()

	select {
	case <-ctx.Done():
		e.handleJobError(job, "job timed out")
	case res := <-resultCh:
		if res.err != nil {
			e.handleJobError(job, res.err.Error())
		} else {
			e.queueManager.UpdateStatus(job.ID, models.JobStatusCompleted, "")
			e.recordHistory(job, models.JobStatusCompleted, res.result, "")
		}
	}
}

type executeResult struct {
	result string
	err    error
}

// handleJobError handles job execution error
// Note: Retry logic removed - jobs table doesn't have retry_count/max_retries columns
func (e *Executor) handleJobError(job *models.Job, errMsg string) {
	e.queueManager.UpdateStatus(job.ID, models.JobStatusFailed, errMsg)
	e.recordHistory(job, models.JobStatusFailed, "", errMsg)
	log.Info().Msgf("Job %s failed: %s", job.ID, errMsg)
}

// recordHistory records job execution in history
func (e *Executor) recordHistory(job *models.Job, status models.JobStatus, result, errMsg string) {
	// This would be implemented via HistoryManager
	data, _ := json.Marshal(map[string]interface{}{
		"job_id": job.ID,
		"status": status,
		"result": result,
		"error":  errMsg,
	})
	_ = data // Log or store as needed
}

// SetPollInterval sets the polling interval
func (e *Executor) SetPollInterval(d time.Duration) {
	e.pollInterval = d
}

// SetTimeout sets the job execution timeout
func (e *Executor) SetTimeout(d time.Duration) {
	e.timeout = d
}
