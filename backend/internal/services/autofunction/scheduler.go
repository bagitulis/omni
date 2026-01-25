package autofunction

import (
	"log"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// Scheduler manages scheduled auto functions using interval-based execution
// Note: Each tenant has its own database connection (schema isolation)
// so no tenant_id column is needed in queries
type Scheduler struct {
	db          *gorm.DB
	executor    *Executor
	tenantID    string // Current tenant ID for execution context
	configs     map[uint]*scheduledConfig
	mu          sync.RWMutex
	running     bool
	stopCh      chan struct{}
	checkTicker *time.Ticker
}

type scheduledConfig struct {
	config  *models.AutoFunctionConfig
	enabled bool
}

// NewScheduler creates a new scheduler
func NewScheduler(db *gorm.DB, executor *Executor) *Scheduler {
	return &Scheduler{
		db:       db,
		executor: executor,
		configs:  make(map[uint]*scheduledConfig),
	}
}

// SetTenantID sets the tenant ID for execution context
func (s *Scheduler) SetTenantID(tenantID string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tenantID = tenantID
}

// Start starts the scheduler
func (s *Scheduler) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil
	}

	// Load all enabled configs
	var configs []models.AutoFunctionConfig
	if err := s.db.Where("enabled = ?", true).Find(&configs).Error; err != nil {
		return err
	}

	// Store each config
	for i := range configs {
		cfg := &configs[i]
		s.configs[cfg.ID] = &scheduledConfig{
			config:  cfg,
			enabled: true,
		}
	}

	// Start check ticker (check every minute)
	s.stopCh = make(chan struct{})
	s.checkTicker = time.NewTicker(1 * time.Minute)
	s.running = true

	go s.runCheckLoop()

	log.Printf("Scheduler started with %d functions", len(configs))
	return nil
}

// runCheckLoop periodically checks for functions to execute
func (s *Scheduler) runCheckLoop() {
	for {
		select {
		case <-s.stopCh:
			return
		case <-s.checkTicker.C:
			s.checkAndExecuteDueFunctions()
		}
	}
}

// checkAndExecuteDueFunctions checks and executes any due functions
func (s *Scheduler) checkAndExecuteDueFunctions() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	now := time.Now()
	for _, sc := range s.configs {
		if !sc.enabled || sc.config.NextScheduledExecution == nil {
			continue
		}

		// Check if within time window
		if !s.isWithinTimeWindow(sc.config, now) {
			continue
		}

		// Check if execution is due
		if now.After(*sc.config.NextScheduledExecution) {
			go s.executor.Execute(s.tenantID, sc.config.ID)
		}
	}
}

// isWithinTimeWindow checks if current time is within start/end window
func (s *Scheduler) isWithinTimeWindow(cfg *models.AutoFunctionConfig, now time.Time) bool {
	if cfg.StartTime == nil || cfg.EndTime == nil {
		return true // No window restriction
	}

	currentTime := now.Format("15:04")
	return currentTime >= *cfg.StartTime && currentTime <= *cfg.EndTime
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return
	}

	close(s.stopCh)
	if s.checkTicker != nil {
		s.checkTicker.Stop()
	}
	s.running = false
	log.Println("Scheduler stopped")
}

// AddOrUpdate adds or updates a scheduled function
func (s *Scheduler) AddOrUpdate(cfg *models.AutoFunctionConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.configs[cfg.ID] = &scheduledConfig{
		config:  cfg,
		enabled: cfg.Enabled,
	}

	return nil
}

// Remove removes a scheduled function
func (s *Scheduler) Remove(configID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.configs, configID)
}

// GetNextRun returns the next run time for a config
func (s *Scheduler) GetNextRun(configID uint) *time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if sc, ok := s.configs[configID]; ok && sc.config != nil {
		return sc.config.NextScheduledExecution
	}
	return nil
}

// IsRunning returns whether scheduler is running
func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetScheduledCount returns count of scheduled functions
func (s *Scheduler) GetScheduledCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.configs)
}

// CancelScheduled cancels a scheduled execution
func (s *Scheduler) CancelScheduled(configID uint) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if sc, ok := s.configs[configID]; ok {
		sc.config.NextScheduledExecution = nil
	}
}

// EnableFunction enables a function by name
func (s *Scheduler) EnableFunction(tenantID, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sc := range s.configs {
		if sc.config.Name == name {
			sc.enabled = true
			sc.config.Enabled = true
			return true
		}
	}
	return false
}

// DisableFunction disables a function by name
func (s *Scheduler) DisableFunction(tenantID, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sc := range s.configs {
		if sc.config.Name == name {
			sc.enabled = false
			sc.config.Enabled = false
			sc.config.NextScheduledExecution = nil
			return true
		}
	}
	return false
}
