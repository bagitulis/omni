// Package jobs provides multi-tenant job executor for background processing
package jobs

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services"
	"gorm.io/gorm"
)

const (
	// MaxGlobalConcurrency is the maximum number of jobs running simultaneously across all tenants
	MaxGlobalConcurrency = 20
	// MaxPerTenantConcurrency is the maximum number of jobs running simultaneously for a single tenant
	MaxPerTenantConcurrency = 3
)

// MultiTenantExecutor executes jobs across all tenant schemas
type MultiTenantExecutor struct {
	serverCtx      context.Context
	systemDB       *gorm.DB
	basePath       string
	handlers       map[string]JobHandler
	pollInterval   time.Duration
	jobTimeout     time.Duration
	stopCh         chan struct{}
	wg             sync.WaitGroup
	mu             sync.RWMutex
	running        bool
	notifyCallback NotifyCallback
	// Concurrency controls
	globalSem   chan struct{}            // Global semaphore (buffered channel)
	tenantSem   map[string]chan struct{} // Per-tenant semaphores
	tenantSemMu sync.Mutex              // Protects tenantSem map
}

// NewMultiTenantExecutor creates a new multi-tenant job executor
func NewMultiTenantExecutor(ctx context.Context, systemDB *gorm.DB, basePath string) *MultiTenantExecutor {
	return &MultiTenantExecutor{
		serverCtx:    ctx,
		systemDB:     systemDB,
		basePath:     basePath,
		handlers:     make(map[string]JobHandler),
		pollInterval: 5 * time.Second,
		jobTimeout:   30 * time.Minute, // Long timeout for escrow sync (1000+ orders)
		stopCh:       make(chan struct{}),
		globalSem:    make(chan struct{}, MaxGlobalConcurrency),
		tenantSem:    make(map[string]chan struct{}),
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

	// Recover zombie jobs across all tenants on startup
	go e.recoverAllZombieJobs()

	e.wg.Add(2)
	go e.pollLoop()
	go e.maintenanceLoop()

	log.Info().Msgf("[MultiTenantExecutor] Started with poll interval %v, maintenance interval 24h", e.pollInterval)
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
		case <-e.serverCtx.Done():
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
// Refactored to be non-blocking (Tenant Parallelism)
func (e *MultiTenantExecutor) processTenantJobs(tenantID string) {
	tenantDB, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		log.Info().Msgf("[MultiTenantExecutor] Error getting tenant DB for %s: %v", tenantID, err)
		return
	}

	qm := NewQueueManager(tenantDB, tenantID)

	// Use atomic ClaimNextJob to prevent race conditions
	job, err := qm.ClaimNextJob()
	if err != nil {
		return // No jobs or error
	}

	log.Info().Msgf("[MultiTenantExecutor] Processing job %s (type: %s) for tenant %s", job.ID, job.Type, tenantID)

	// Try to acquire per-tenant semaphore (non-blocking)
	tenantSem := e.getTenantSem(tenantID)
	select {
	case tenantSem <- struct{}{}:
		// Got per-tenant slot
	default:
		// Tenant at max concurrency, skip this poll cycle
		log.Debug().Str("tenant_id", tenantID).Msg("[MultiTenantExecutor] Tenant at max concurrency, skipping")
		return
	}

	// Try to acquire global semaphore (non-blocking)
	select {
	case e.globalSem <- struct{}{}:
		// Got global slot
	default:
		// Global limit reached, release tenant slot
		<-tenantSem
		log.Debug().Msg("[MultiTenantExecutor] Global concurrency limit reached, skipping")
		return
	}

	// Execute with panic recovery
	e.wg.Add(1)
	go func() {
		defer e.wg.Done()
		defer func() { <-e.globalSem }()
		defer func() { <-tenantSem }()
		defer func() {
			if r := recover(); r != nil {
				log.Error().Interface("panic", r).Str("job_id", job.ID).Str("tenant_id", tenantID).Msg("[MultiTenantExecutor] Job panicked")
				qm := NewQueueManager(tenantDB, tenantID)
				qm.FailJob(job.ID, fmt.Sprintf("job panicked: %v", r))
			}
		}()
		e.executeJob(tenantDB, tenantID, job)
	}()
}

// getTenantSem returns the per-tenant semaphore, creating it if needed
func (e *MultiTenantExecutor) getTenantSem(tenantID string) chan struct{} {
	e.tenantSemMu.Lock()
	defer e.tenantSemMu.Unlock()
	sem, ok := e.tenantSem[tenantID]
	if !ok {
		sem = make(chan struct{}, MaxPerTenantConcurrency)
		e.tenantSem[tenantID] = sem
	}
	return sem
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
	ctx, cancel := context.WithTimeout(e.serverCtx, e.jobTimeout)
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
		switch ClassifyHandlerResult(res.err) {
		case HandlerOutcomeSelfRecorded:
			// The handler already wrote a terminal state that a generic
			// completion/failure would overwrite (a blocked scrape holds a
			// resume cursor). Leave the row alone; only push the notification.
			log.Info().Str("job_id", job.ID).Msg("[MultiTenantExecutor] Job handler recorded its own outcome")
			e.pushNotificationToDB(tenantDB, tenantID, job, true, res.result)
		case HandlerOutcomeFailed:
			log.Info().Msgf("[MultiTenantExecutor] Job %s failed: %v", job.ID, res.err)
			qm.FailJob(job.ID, res.err.Error())
			e.pushNotificationToDB(tenantDB, tenantID, job, false, res.err.Error())
		default:
			log.Info().Msgf("[MultiTenantExecutor] Job %s completed successfully", job.ID)
			qm.CompleteJobWithResult(job.ID, res.result)
			e.pushNotificationToDB(tenantDB, tenantID, job, true, res.result)
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

// SetNotifyCallback sets the notification callback for job completion
func (e *MultiTenantExecutor) SetNotifyCallback(cb NotifyCallback) {
	e.notifyCallback = cb
}

// pushNotificationToDB pushes a notification after job completion
func (e *MultiTenantExecutor) pushNotificationToDB(db *gorm.DB, tenantID string, job *models.Job, success bool, detail string) {
	// First push to database for persistence and SSE broadcasting
	repo := repositories.NewNotificationRepository(db)
	svc := services.NewNotificationService(repo).WithTenant(tenantID)
	err := svc.PushJobResult(context.Background(), job, success, detail)
	if err != nil {
		log.Error().Err(err).Msg("[MultiTenantExecutor] Failed to push notification to DB")
	}

	// Then trigger local callback if any
	if e.notifyCallback != nil {
		e.notifyCallback(job.Type, success, detail)
	}
}

// recoverAllZombieJobs recovers zombie jobs across all tenants on startup
func (e *MultiTenantExecutor) recoverAllZombieJobs() {
	tenants, err := e.getTenantList()
	if err != nil {
		log.Error().Err(err).Msg("[MultiTenantExecutor] Failed to get tenant list for zombie recovery")
		return
	}

	for _, tenantID := range tenants {
		tenantDB, err := config.GetTenantDBByID(tenantID)
		if err != nil {
			continue
		}
		qm := NewQueueManager(tenantDB, tenantID)
		recovered, err := qm.RecoverZombieJobs()
		if err == nil && recovered > 0 {
			log.Info().Msgf("[MultiTenantExecutor] Recovered %d zombie jobs for tenant %s", recovered, tenantID)
		}
	}
}

// maintenanceLoop runs daily maintenance tasks (e.g., notification cleanup)
func (e *MultiTenantExecutor) maintenanceLoop() {
	defer e.wg.Done()

	// Run every 24 hours
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Run once immediately on start
	e.runMaintenance()

	for {
		select {
		case <-e.stopCh:
			return
		case <-ticker.C:
			e.runMaintenance()
		}
	}
}

// runMaintenance executes maintenance tasks for all tenants
func (e *MultiTenantExecutor) runMaintenance() {
	tenants, err := e.getTenantList()
	if err != nil {
		log.Error().Err(err).Msg("[MultiTenantExecutor] Maintenance: Failed to get tenant list")
		return
	}

	log.Info().Msgf("[MultiTenantExecutor] Starting daily maintenance for %d tenants", len(tenants))

	for _, tenantID := range tenants {
		tenantDB, err := config.GetTenantDBByID(tenantID)
		if err != nil {
			continue
		}

		repo := repositories.NewNotificationRepository(tenantDB)
		svc := services.NewNotificationService(repo)
		settings, err := svc.GetSettings(context.Background())
		if err == nil && settings.RetentionDays > 0 {
			deleted := svc.CleanupOlderThan(context.Background(), settings.RetentionDays)
			if deleted > 0 {
				log.Info().Msgf("[MultiTenantExecutor] Maintenance: Cleaned up %d notifications for tenant %s", deleted, tenantID)
			}
		}
	}
}
