package autofunction

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
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
		log.Println("⚠️ Multi-tenant auto-function scheduler already running")
		return
	}
	s.running = true
	s.stopCh = make(chan struct{})
	s.mu.Unlock()

	log.Println("🚀 Multi-tenant auto-function scheduler started (check every 60s)")
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
	log.Println("🛑 Multi-tenant auto-function scheduler stopped")
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

	for {
		select {
		case <-s.stopCh:
			return
		case <-ticker.C:
			s.checkAllTenants()
		}
	}
}

// checkAllTenants iterates all tenants and checks their auto functions
func (s *MultiTenantScheduler) checkAllTenants() {
	tenantIDs, err := s.getRealTenantIDs()
	if err != nil {
		log.Printf("❌ Failed to get tenant IDs: %v", err)
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
		log.Printf("❌ Failed to get DB for tenant %s: %v", tenantID, err)
		return
	}

	// Ensure default auto functions exist for this tenant
	s.ensureDefaultAutoFunctions(tenantDB, tenantID)

	// Find all enabled auto functions that are due
	var configs []models.AutoFunctionConfig
	err = tenantDB.Where("enabled = ?", true).Find(&configs).Error
	if err != nil {
		log.Printf("❌ Failed to get auto functions for tenant %s: %v", tenantID, err)
		return
	}

	now := time.Now()
	for _, cfg := range configs {
		if s.shouldExecute(&cfg, now) {
			log.Printf("⏰ [%s] %s trigger condition met, executing...", tenantID, cfg.Name)
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

// executeAutoFunction executes an auto function and updates its state
func (s *MultiTenantScheduler) executeAutoFunction(tenantID string, tenantDB *gorm.DB, cfg *models.AutoFunctionConfig, now time.Time) {
	startTime := time.Now()

	// Get handler
	handler := s.executor.GetHandler(cfg.Name)
	if handler == nil {
		log.Printf("❌ [%s] No handler registered for function: %s", tenantID, cfg.Name)
		s.recordHistory(tenantDB, cfg.Name, "failed", "no handler registered: "+cfg.Name, startTime)
		return
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	result, err := handler(ctx, tenantID, cfg)

	// Calculate next execution time
	nextExecution := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)

	// Update config in database
	updates := map[string]interface{}{
		"last_executed":            now,
		"next_scheduled_execution": nextExecution,
	}
	if err := tenantDB.Model(cfg).Updates(updates).Error; err != nil {
		log.Printf("❌ [%s] Failed to update config: %v", tenantID, err)
	}

	// Record history
	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
		log.Printf("❌ [%s] Auto function %s failed: %v", tenantID, cfg.Name, err)
	} else {
		log.Printf("✅ [%s] Auto function %s completed: %s", tenantID, cfg.Name, result)
	}

	s.recordHistory(tenantDB, cfg.Name, status, errMsg, startTime)
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
		log.Printf("❌ Failed to record auto function history: %v", err)
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
	{Name: "sync_products_inventory", IntervalMinutes: 120, StartTime: "08:00", EndTime: "22:00"}, // Every 2 hours during business hours
}

// ensureDefaultAutoFunctions ensures default auto functions exist for a tenant.
// Uses per-name check so new defaults are seeded even when other configs exist.
func (s *MultiTenantScheduler) ensureDefaultAutoFunctions(tenantDB *gorm.DB, tenantID string) {
	now := time.Now()

	for _, def := range defaultAutoFunctions {
		var count int64
		tenantDB.Model(&models.AutoFunctionConfig{}).Where("name = ?", def.Name).Count(&count)
		if count > 0 {
			continue
		}

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

		if err := tenantDB.Create(cfg).Error; err != nil {
			log.Printf("❌ [%s] Failed to seed %s: %v", tenantID, def.Name, err)
		} else {
			log.Printf("✅ [%s] Seeded auto function: %s (interval: %dm)", tenantID, def.Name, def.IntervalMinutes)
		}
	}
}
