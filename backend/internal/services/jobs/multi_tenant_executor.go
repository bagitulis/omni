// Package jobs provides multi-tenant job executor for background processing
package jobs

import (
	"context"
	"github.com/rs/zerolog/log"
	"sync"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// MultiTenantExecutor executes jobs across all tenant schemas
type MultiTenantExecutor struct {
	systemDB     *gorm.DB
	basePath     string
	handlers     map[string]JobHandler
	pollInterval time.Duration
	jobTimeout   time.Duration
	stopCh       chan struct{}
	wg           sync.WaitGroup
	mu           sync.RWMutex
	running      bool
}

// NewMultiTenantExecutor creates a new multi-tenant job executor
func NewMultiTenantExecutor(systemDB *gorm.DB, basePath string) *MultiTenantExecutor {
	return &MultiTenantExecutor{
		systemDB:     systemDB,
		basePath:     basePath,
		handlers:     make(map[string]JobHandler),
		pollInterval: 5 * time.Second,
		jobTimeout:   10 * time.Minute, // Long timeout for escrow sync
		stopCh:       make(chan struct{}),
	}
}

// RegisterHandler registers a handler for a job type
func (e *MultiTenantExecutor) RegisterHandler(jobType string, handler JobHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[jobType] = handler
	log.Info().Msgf("[MultiTenantExecutor] Registered handler for job type: %s", jobType)
}

// Start starts the executor
func (e *MultiTenantExecutor) Start() {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return
	}
	e.running = true
	e.mu.Unlock()

	e.wg.Add(1)
	go e.pollLoop()

	log.Info().Msgf("[MultiTenantExecutor] Started with poll interval %v, job timeout %v", e.pollInterval, e.jobTimeout)
}

// Stop stops the executor gracefully
func (e *MultiTenantExecutor) Stop() {
	e.mu.Lock()
	if !e.running {
		e.mu.Unlock()
		return
	}
	e.running = false
	e.mu.Unlock()

	close(e.stopCh)
	e.wg.Wait()
	log.Info().Msg("[MultiTenantExecutor] Stopped")
}

// pollLoop continuously polls for pending jobs across all tenants
func (e *MultiTenantExecutor) pollLoop() {
	defer e.wg.Done()

	ticker := time.NewTicker(e.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.processAllTenants()
		}
	}
}

// processAllTenants iterates all tenants and processes pending jobs
func (e *MultiTenantExecutor) processAllTenants() {
	tenants, err := e.getTenantList()
	if err != nil {
		log.Info().Msgf("[MultiTenantExecutor] Error getting tenant list: %v", err)
		return
	}

	for _, tenantID := range tenants {
		select {
		case <-e.stopCh:
			return
		default:
			e.processTenantJobs(tenantID)
		}
	}
}

// getTenantList retrieves list of all tenant IDs from system schema
func (e *MultiTenantExecutor) getTenantList() ([]string, error) {
	var tenants []struct {
		TenantID string `gorm:"column:tenant_id"`
	}

	// Query system.tenants table - use tenant_id (string) NOT id (UUID)
	err := e.systemDB.Table("system.tenants").
		Select("tenant_id").
		Where("is_active = ?", true).
		Find(&tenants).Error

	if err != nil {
		return nil, err
	}

	ids := make([]string, len(tenants))
	for i, t := range tenants {
		ids[i] = t.TenantID
	}
	return ids, nil
}

// processTenantJobs processes pending jobs for a specific tenant
func (e *MultiTenantExecutor) processTenantJobs(tenantID string) {
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		log.Info().Msgf("[MultiTenantExecutor] Error getting tenant DB for %s: %v", tenantID, err)
		return
	}

	qm := NewQueueManager(tenantDB, tenantID)
	jobs, err := qm.GetPendingJobs(1) // Process one job at a time per tenant
	if err != nil || len(jobs) == 0 {
		return
	}

	job := jobs[0]
	log.Info().Msgf("[MultiTenantExecutor] Processing job %s (type: %s) for tenant %s", job.ID, job.Type, tenantID)
	e.executeJob(tenantDB, tenantID, &job)
}

// executeJob executes a single job
func (e *MultiTenantExecutor) executeJob(tenantDB *gorm.DB, tenantID string, job *models.Job) {
	e.mu.RLock()
	handler, ok := e.handlers[job.Type]
	e.mu.RUnlock()

	if !ok {
		log.Info().Msgf("[MultiTenantExecutor] No handler for job type: %s", job.Type)
		qm := NewQueueManager(tenantDB, tenantID)
		qm.UpdateStatus(job.ID, models.JobStatusFailed, "no handler for job type: "+job.Type)
		return
	}

	qm := NewQueueManager(tenantDB, tenantID)

	// Mark as running
	if err := qm.UpdateStatus(job.ID, models.JobStatusRunning, ""); err != nil {
		log.Info().Msgf("[MultiTenantExecutor] Failed to update job status: %v", err)
		return
	}

	// Create context with timeout and job ID
	ctx, cancel := context.WithTimeout(context.Background(), e.jobTimeout)
	defer cancel()

	// Add job ID to context for progress updates
	ctx = context.WithValue(ctx, models.ContextKeyJobID, job.ID)
	ctx = context.WithValue(ctx, models.ContextKeyTenantID, tenantID)

	// Execute in goroutine to allow timeout handling
	resultCh := make(chan executeResult, 1)
	go func() {
		result, err := handler(ctx, job.Data)
		resultCh <- executeResult{result: result, err: err}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msgf("[MultiTenantExecutor] Job %s timed out after %v", job.ID, e.jobTimeout)
		qm.FailJob(job.ID, "job timed out after "+e.jobTimeout.String())
	case res := <-resultCh:
		if res.err != nil {
			log.Info().Msgf("[MultiTenantExecutor] Job %s failed: %v", job.ID, res.err)
			qm.FailJob(job.ID, res.err.Error())
		} else {
			log.Info().Msgf("[MultiTenantExecutor] Job %s completed successfully", job.ID)
			qm.CompleteJobWithResult(job.ID, res.result)
		}
	}
}

// SetPollInterval sets the polling interval
func (e *MultiTenantExecutor) SetPollInterval(d time.Duration) {
	e.pollInterval = d
}

// SetJobTimeout sets the job execution timeout
func (e *MultiTenantExecutor) SetJobTimeout(d time.Duration) {
	e.jobTimeout = d
}
