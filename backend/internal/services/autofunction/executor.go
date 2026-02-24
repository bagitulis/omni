package autofunction

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// FunctionHandler is a function that handles auto function execution
type FunctionHandler func(ctx context.Context, tenantID string, config *models.AutoFunctionConfig) (string, error)

// Executor executes auto functions
type Executor struct {
	db       *gorm.DB
	handlers map[string]FunctionHandler
	mu       sync.RWMutex
	timeout  time.Duration
}

// NewExecutor creates a new auto function executor
func NewExecutor(db *gorm.DB) *Executor {
	return &Executor{
		db:       db,
		handlers: make(map[string]FunctionHandler),
		timeout:  10 * time.Minute,
	}
}

// RegisterHandler registers a handler for a function name
func (e *Executor) RegisterHandler(functionName string, handler FunctionHandler) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.handlers[functionName] = handler
}

// GetHandler returns a handler for a function name (nil if not found)
func (e *Executor) GetHandler(functionName string) FunctionHandler {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.handlers[functionName]
}

// Execute executes an auto function by config ID
func (e *Executor) Execute(tenantID string, configID uint) {
	e.ExecuteWithDB(e.db, tenantID, configID)
}

// ExecuteWithDB executes an auto function by config ID using the provided DB
func (e *Executor) ExecuteWithDB(db *gorm.DB, tenantID string, configID uint) {
	var cfg models.AutoFunctionConfig
	if err := db.Where("id = ?", configID).First(&cfg).Error; err != nil {
		log.Printf("Auto function config not found: %d", configID)
		return
	}
	e.executeConfig(db, tenantID, &cfg)
}

// ExecuteManual manually executes an auto function
func (e *Executor) ExecuteManual(tenantID string, configID uint) error {
	var cfg models.AutoFunctionConfig
	if err := e.db.Where("id = ?", configID).First(&cfg).Error; err != nil {
		return err
	}
	go e.executeConfig(e.db, tenantID, &cfg)
	return nil
}

// ExecuteByName manually executes an auto function by name
func (e *Executor) ExecuteByName(tenantID string, name string) error {
	return e.ExecuteByNameWithDB(e.db, tenantID, name)
}

// ExecuteByNameWithDB manually executes an auto function by name using the provided DB
// This is used when the caller has a tenant-scoped DB (e.g., from HTTP handler context)
func (e *Executor) ExecuteByNameWithDB(db *gorm.DB, tenantID string, name string) error {
	var cfg models.AutoFunctionConfig
	if err := db.Where("name = ?", name).First(&cfg).Error; err != nil {
		return err
	}
	go e.executeConfig(db, tenantID, &cfg)
	return nil
}

// executeConfig executes a single config
func (e *Executor) executeConfig(db *gorm.DB, tenantID string, cfg *models.AutoFunctionConfig) {
	startTime := time.Now()

	e.mu.RLock()
	handler, ok := e.handlers[cfg.Name]
	e.mu.RUnlock()

	if !ok {
		e.recordHistory(db, cfg.Name, "failed", "no handler for function: "+cfg.Name, startTime)
		return
	}

	// Execute with timeout
	ctx, cancel := context.WithTimeout(context.Background(), e.timeout)
	defer cancel()

	_, err := handler(ctx, tenantID, cfg)

	// Update last run time and schedule next
	now := time.Now()
	nextExecution := now.Add(time.Duration(cfg.IntervalMinutes) * time.Minute)
	db.Model(cfg).Updates(map[string]interface{}{
		"last_executed":            now,
		"next_scheduled_execution": nextExecution,
	})

	// Record history
	status := "success"
	errMsg := ""
	if err != nil {
		status = "failed"
		errMsg = err.Error()
		log.Printf("Auto function %s failed: %v", cfg.Name, err)
	}

	e.recordHistory(db, cfg.Name, status, errMsg, startTime)
}

// recordHistory records execution in history
func (e *Executor) recordHistory(db *gorm.DB, functionName, status, errMsg string, startTime time.Time) {
	history := &models.AutoFunctionHistory{
		FunctionName: functionName,
		Status:       status,
		ErrorMessage: errMsg,
		DurationMs:   int(time.Since(startTime).Milliseconds()),
		ExecutedAt:   startTime,
	}

	if err := db.Create(history).Error; err != nil {
		log.Printf("Failed to record auto function history: %v", err)
	}
}

// GetHistory retrieves execution history for a function name
func (e *Executor) GetHistory(name string, limit int) ([]models.AutoFunctionHistory, error) {
	var history []models.AutoFunctionHistory
	err := e.db.Where("function_name = ?", name).
		Order("executed_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// GetAllHistory retrieves all execution history
func (e *Executor) GetAllHistory(tenantID string, limit int) ([]models.AutoFunctionHistory, error) {
	var history []models.AutoFunctionHistory
	err := e.db.Order("executed_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

// SetTimeout sets execution timeout
func (e *Executor) SetTimeout(d time.Duration) {
	e.timeout = d
}
