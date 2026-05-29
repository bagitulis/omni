package autofunction

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/jobs"
	"github.com/omni/backend/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// MultiTenantScheduler manages auto-function scheduling across all tenants
// Similar to Node.js AutoFunctionScheduler which iterates getRealTenantIds()
type MultiTenantScheduler struct {
	systemDB   *gorm.DB
	executor   *Executor
	basePath   string
	running    bool
	stopCh     chan struct{}
	mu         sync.RWMutex
	checkEvery time.Duration
}

// NewMultiTenantScheduler creates a scheduler that works across all tenants
func NewMultiTenantScheduler(systemDB *gorm.DB, executor *Executor, basePath string) *MultiTenantScheduler {
	return &MultiTenantScheduler{
		systemDB:   systemDB,
		executor:   executor,
		basePath:   basePath,
		checkEvery: 60 * time.Second, // Check every 60 seconds like Node.js
	}
}

// Start starts the multi-tenant scheduler
func (s *MultiTenantScheduler) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		log.Info().Msg("⚠️ Multi-tenant auto-function scheduler already running")
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	log.Info().Msg("🚀 Multi-tenant auto-function scheduler started (check every 60s)")
	go s.runLoop()
}

// Stop stops the scheduler
func (s *MultiTenantScheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopCh)
	s.running = false
	log.Info().Msg("🛑 Multi-tenant auto-function scheduler stopped")
}

// IsRunning returns whether scheduler is running
func (s *MultiTenantScheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// runLoop is the main scheduler loop
func (s *MultiTenantScheduler) runLoop() {
	ticker := time.NewTicker(s.checkEvery)
	defer ticker.Stop()

	// Run immediately on start
	s.checkAllTenants()

	tickCount := 0
	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			tickCount++
			s.checkAllTenants()

			// Cleanup old completed/failed jobs every ~6 hours (360 ticks at 60s each)
			if tickCount%360 == 0 {
				s.cleanupOldJobs()
			}

			// Watchdog: detect and timeout stuck jobs every ~15 minutes (15 ticks at 60s each)
			if tickCount%15 == 0 {
				s.timeoutStuckJobs()
			}
		}
	}
}

// checkAllTenants iterates all tenants and checks their auto functions
func (s *MultiTenantScheduler) checkAllTenants() {
	tenantIDs, err := s.getRealTenantIDs()
	if err != nil {
		log.Info().Msgf("❌ Failed to get tenant IDs: %v", err)
		return
	}

	for _, tenantID := range tenantIDs {
		s.checkTenantAutoFunctions(tenantID)
	}
}

// getRealTenantIDs gets all real tenant IDs from PostgreSQL schemas
// Similar to Node.js getRealTenantIds() - excludes system, default
func (s *MultiTenantScheduler) getRealTenantIDs() ([]string, error) {
	var schemas []string
	err := s.systemDB.Raw(`
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'tenant_%'
		ORDER BY schema_name
	`).Scan(&schemas).Error
	if err != nil {
		return nil, err
	}

	tenantIDs := make([]string, 0, len(schemas))
	for _, schema := range schemas {
		// Extract tenant ID: tenant_xxx -> xxx
		if len(schema) > 7 { // "tenant_" = 7 chars
			tenantIDs = append(tenantIDs, schema[7:])
		}
	}
	return tenantIDs, nil
}

// checkTenantAutoFunctions checks and executes due auto functions for a tenant
func (s *MultiTenantScheduler) checkTenantAutoFunctions(tenantID string) {
	// Get tenant-specific DB connection
	tenantDB, err := config.GetTenantDBWithContext(tenantID, s.basePath)
	if err != nil {
		log.Info().Msgf("❌ Failed to get DB for tenant %s: %v", tenantID, err)
		return
	}

	// Ensure default auto functions exist for this tenant
	s.ensureDefaultAutoFunctions(tenantDB, tenantID)

	// Find all enabled auto functions that are due
	var configs []models.AutoFunctionConfig
	err = tenantDB.Where("enabled = ?", true).Limit(100).Find(&configs).Error
	if err != nil {
		log.Info().Msgf("❌ Failed to get auto functions for tenant %s: %v", tenantID, err)
		return
	}

	now := time.Now()
	for _, cfg := range configs {
		if s.shouldExecute(&cfg, now) {
			log.Info().Msgf("⏰ [%s] %s trigger condition met, executing...", tenantID, cfg.Name)
			s.executeAutoFunction(tenantID, tenantDB, &cfg, now)
		}
	}
}

// shouldExecute checks if a config should be executed now
// Similar to Node.js shouldExecute()
func (s *MultiTenantScheduler) shouldExecute(cfg *models.AutoFunctionConfig, now time.Time) bool {
	// If no next scheduled execution, it should run
	if cfg.NextScheduledExecution == nil {
		return true
	}

	// Check if within time window (if configured)
	// Time window uses WIB (UTC+7) timezone for Indonesia
	if cfg.StartTime != nil && cfg.EndTime != nil {
		currentTime := utils.ToWIB(now).Format("15:04")
		if currentTime < *cfg.StartTime || currentTime > *cfg.EndTime {
			return false
		}
	}

	// Check if execution time has passed
	// Convert nextExec to WIB for comparison
	nextExec := utils.ToWIB(*cfg.NextScheduledExecution)
	nowWIB := utils.ToWIB(now)
	return nowWIB.After(nextExec) || nowWIB.Equal(nextExec)
}

// executeAutoFunction executes an auto function and updates its state.
// Enqueues to job queue for visibility in Current Job / Queue tabs.
func (s *MultiTenantScheduler) executeAutoFunction(tenantID string, tenantDB *gorm.DB, cfg *models.AutoFunctionConfig, now time.Time) {
	startTime := time.Now()

	// Guard: skip if already running (prevents race condition with concurrent ticks)
	if cfg.IsRunning {
		log.Info().Msgf("\u26a0\ufe0f [%s] Function %s already running, skipping", tenantID, cfg.Name)
		return
	}

	// Enqueue to job queue for visibility in Current Job / Queue tabs
	qm := jobs.NewQueueManager(tenantDB, tenantID)
	jobID, enqueueErr := qm.EnqueueJob("auto_function", map[string]interface{}{
		"function_name": cfg.Name,
		"trigger":       "scheduled",
	}, "normal")
	if enqueueErr != nil {
		log.Info().Msgf("\u26a0\ufe0f [%s] Failed to enqueue job for %s: %v (continuing anyway)", tenantID, cfg.Name, enqueueErr)
	}

	// Mark as running (visible in Script Monitor)
	tenantDB.Model(cfg).Updates(map[string]interface{}{
		"is_running":      true,
		"run_started_at":  startTime,
	})

	// Track whether execution succeeded (for progress_data cleanup)
	execSuccess := false

	// Panic recovery: prevent a single handler panic from killing the entire scheduler.
	// Also ensures is_running is cleared on exit (even on panic).
	// Only clear progress_data on SUCCESS — preserve it on failure/panic for resume.
	defer func() {
		if r := recover(); r != nil {
			log.Error().Msgf("\U0001f525 [%s] PANIC in %s: %v", tenantID, cfg.Name, r)
			s.recordHistory(tenantDB, cfg.Name, "failed", fmt.Sprintf("panic: %v", r), startTime)
			if jobID != "" {
				_ = qm.FailJob(jobID, fmt.Sprintf("panic: %v", r))
			}
		}

		updates := map[string]interface{}{
			"is_running":    false,
			"run_started_at": nil,
		}
		if execSuccess {
			updates["progress_data"] = ""
		}
		tenantDB.Model(cfg).Updates(updates)
	}()

	// Get handler
	handler := s.executor.GetHandler(cfg.Name)
	if handler == nil {
		log.Info().Msgf("\u274c [%s] No handler registered for function: %s", tenantID, cfg.Name)
		s.recordHistory(tenantDB, cfg.Name, "failed", "no handler registered: "+cfg.Name, startTime)
		if jobID != "" {
			_ = qm.FailJob(jobID, "no handler registered: "+cfg.Name)
		}
		return
	}

	// Mark job as running in queue
	if jobID != "" {
		_ = qm.UpdateStatus(jobID, models.JobStatusRunning, "")
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := handler(ctx, tenantID, cfg)

	// Calculate next execution time
	nextExecution := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)

	// Update config in database
	cfgUpdates := map[string]interface{}{
		"last_executed":            now,
		"next_scheduled_execution": nextExecution,
	}
	if err := tenantDB.Model(cfg).Updates(cfgUpdates).Error; err != nil {
		log.Info().Msgf("\u274c [%s] Failed to update config: %v", tenantID, err)
	}

	// Record history: only write to auto_functions_history when no job queue entry exists.
	// When jobID is set, the job queue entry (job_history) is the single source of truth,
	// preventing duplicate entries in the merged History tab.
	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
		log.Info().Msgf("\u274c [%s] Auto function %s failed: %v", tenantID, cfg.Name, err)
		if jobID != "" {
			_ = qm.FailJob(jobID, errMsg)
		}
	} else {
		execSuccess = true
		log.Info().Msgf("\u2705 [%s] Auto function %s completed: %s", tenantID, cfg.Name, result)
		if jobID != "" {
			_ = qm.CompleteJobWithResult(jobID, result)
		}
	}

	// Only record to auto_functions_history if job queue entry was NOT created
	// (avoids duplicate entries in merged history view)
	if jobID == "" {
		s.recordHistory(tenantDB, cfg.Name, status, errMsg, startTime)
	}
}

// recordHistory records execution history in tenant's database
func (s *MultiTenantScheduler) recordHistory(tenantDB *gorm.DB, functionName, status, errMsg string, startTime time.Time) {
	history := &models.AutoFunctionHistory{
		FunctionName: functionName,
		Status:       status,
		ErrorMessage: errMsg,
		DurationMs:   int(time.Since(startTime).Milliseconds()),
		ExecutedAt:   startTime,
	}

	if err := tenantDB.Create(history).Error; err != nil {
		log.Info().Msgf("❌ Failed to record auto function history: %v", err)
	}
}

// cleanupOldJobs removes completed/failed jobs older than 7 days from all tenants.
// Prevents unbounded growth of the jobs table.
func (s *MultiTenantScheduler) cleanupOldJobs() {
	tenantIDs, err := s.getRealTenantIDs()
	if err != nil {
		return
	}

	for _, tenantID := range tenantIDs {
		tenantDB, err := config.GetTenantDB(tenantID, s.basePath)
		if err != nil {
			continue
		}
		qm := jobs.NewQueueManager(tenantDB, tenantID)
		if _, err := qm.CleanupOldJobs(7 * 24 * time.Hour); err != nil {
			log.Info().Msgf("\u26a0\ufe0f [%s] Job cleanup failed: %v", tenantID, err)
		}
	}
}

// timeoutStuckJobs detects auto-functions stuck in is_running=true for >15 minutes
// and clears them. Also detects stuck job queue entries via CheckAndTimeoutStuckJobs.
func (s *MultiTenantScheduler) timeoutStuckJobs() {
	tenantIDs, err := s.getRealTenantIDs()
	if err != nil {
		return
	}

	cutoff := time.Now().Add(-15 * time.Minute)
	for _, tenantID := range tenantIDs {
		tenantDB, err := config.GetTenantDB(tenantID, s.basePath)
		if err != nil {
			continue
		}

		// Clear stuck auto-function is_running flags
		result := tenantDB.Model(&models.AutoFunctionConfig{}).
			Where("is_running = ? AND run_started_at < ?", true, cutoff).
			Updates(map[string]interface{}{
				"is_running":    false,
				"run_started_at": nil,
			})
		if result.RowsAffected > 0 {
			log.Info().Msgf("\u26a0\ufe0f [%s] Watchdog: cleared %d stuck auto-functions", tenantID, result.RowsAffected)
		}

		// Clear stuck job queue entries
		qm := jobs.NewQueueManager(tenantDB, tenantID)
		stuckJobs := qm.CheckAndTimeoutStuckJobs(15)
		if len(stuckJobs) > 0 {
			log.Info().Msgf("\u26a0\ufe0f [%s] Watchdog: timed out %d stuck jobs", tenantID, len(stuckJobs))
		}
	}
}

// defaultAutoFunctions defines the default auto functions to seed
var defaultAutoFunctions = []struct {
	Name            string
	IntervalMinutes int
	StartTime       string
	EndTime         string
}{
	{Name: "locked_today", IntervalMinutes: 1440, StartTime: "22:00", EndTime: "23:59"},           // Daily at 22:00-23:59 WIB
	{Name: "sync_from_sheets", IntervalMinutes: 30, StartTime: "08:00", EndTime: "22:00"},         // Every 30 min during business hours
	{Name: "auto_update_token", IntervalMinutes: 180, StartTime: "00:00", EndTime: "23:59"},       // Every 3 hours
	{Name: "sync_products", IntervalMinutes: 120, StartTime: "08:00", EndTime: "22:00"},           // Every 2 hours during business hours
	{Name: "sync_products_inventory", IntervalMinutes: 60, StartTime: "08:00", EndTime: "22:00"}, // Every 60 min during business hours
	{Name: "retry_failed_syncs", IntervalMinutes: 30, StartTime: "08:00", EndTime: "22:00"},      // Every 30 min during business hours
	{Name: "price_drift_detection", IntervalMinutes: 360, StartTime: "08:00", EndTime: "22:00"},  // Every 6 hours during business hours
}

// ensureDefaultAutoFunctions ensures default auto functions exist for a tenant.
// Uses ON CONFLICT DO NOTHING to atomically skip existing rows, preventing duplicates
// even under race conditions where Count returns 0 for concurrent callers.
func (s *MultiTenantScheduler) ensureDefaultAutoFunctions(tenantDB *gorm.DB, tenantID string) {
	now := time.Now()

	// Ensure unique index exists (idempotent — fails silently if already present)
	tenantDB.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_auto_functions_config_name ON auto_functions_config (name)`)

	for _, def := range defaultAutoFunctions {
		startTime := def.StartTime
		endTime := def.EndTime
		nextExec := now.Add(time.Duration(def.IntervalMinutes) * time.Minute)

		cfg := &models.AutoFunctionConfig{
			Name:                   def.Name,
			Enabled:                true,
			IntervalMinutes:        def.IntervalMinutes,
			StartTime:              &startTime,
			EndTime:                &endTime,
			NextScheduledExecution: &nextExec,
			CreatedAt:              now,
			UpdatedAt:              now,
		}

		if err := tenantDB.Clauses(clause.OnConflict{DoNothing: true}).Create(cfg).Error; err != nil {
			log.Info().Msgf("❌ [%s] Failed to seed %s: %v", tenantID, def.Name, err)
		}
	}
}
