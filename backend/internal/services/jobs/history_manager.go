package jobs

import (
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// HistoryManager manages job execution history
// Note: Multi-tenancy is handled by schema isolation - no tenant_id column in database
type HistoryManager struct {
	db *gorm.DB
}

// NewHistoryManager creates a new history manager
func NewHistoryManager(db *gorm.DB, _ string) *HistoryManager {
	// tenantID parameter kept for backward compatibility but not used
	// Schema isolation handles multi-tenancy (search_path already set)
	return &HistoryManager{db: db}
}

// RecordExecution records job execution in history
func (m *HistoryManager) RecordExecution(job *models.Job, status models.JobStatus, errMsg string, durationMs int) error {
	history := &models.JobHistory{
		JobID:        job.ID,
		JobType:      job.Type,
		Status:       status,
		ErrorMessage: errMsg,
		DurationMs:   durationMs,
		StartedAt:    job.StartedAt,
		CompletedAt:  job.CompletedAt,
	}

	return m.db.Create(history).Error
}

// GetHistory retrieves job execution history
func (m *HistoryManager) GetHistory(filter HistoryFilter) ([]models.JobHistory, int64, error) {
	query := m.db.Model(&models.JobHistory{})

	if filter.JobID != "" {
		query = query.Where("job_id = ?", filter.JobID)
	}
	if filter.JobType != "" {
		query = query.Where("job_type = ?", filter.JobType)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if !filter.StartDate.IsZero() {
		query = query.Where("created_at >= ?", filter.StartDate)
	}
	if !filter.EndDate.IsZero() {
		query = query.Where("created_at <= ?", filter.EndDate)
	}

	var total int64
	query.Count(&total)

	// Pagination
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	offset := (filter.Page - 1) * filter.PageSize

	var history []models.JobHistory
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&history).Error

	return history, total, err
}

// HistoryFilter represents history query filters
type HistoryFilter struct {
	JobID     string
	JobType   string
	Status    models.JobStatus
	StartDate time.Time
	EndDate   time.Time
	Page      int
	PageSize  int
}

// GetStatistics returns execution statistics
func (m *HistoryManager) GetStatistics(days int) (*ExecutionStats, error) {
	since := time.Now().AddDate(0, 0, -days)

	var stats ExecutionStats

	// Total executions
	m.db.Model(&models.JobHistory{}).
		Where("created_at >= ?", since).
		Count(&stats.TotalExecutions)

	// Success count
	m.db.Model(&models.JobHistory{}).
		Where("created_at >= ? AND status = ?", since, models.JobStatusCompleted).
		Count(&stats.SuccessCount)

	// Failure count
	m.db.Model(&models.JobHistory{}).
		Where("created_at >= ? AND status = ?", since, models.JobStatusFailed).
		Count(&stats.FailureCount)

	// Average duration
	var avgDuration float64
	m.db.Model(&models.JobHistory{}).
		Where("created_at >= ? AND status = ?", since, models.JobStatusCompleted).
		Select("AVG(duration_ms)").
		Scan(&avgDuration)
	stats.AvgDurationMs = int64(avgDuration)

	// Success rate
	if stats.TotalExecutions > 0 {
		stats.SuccessRate = float64(stats.SuccessCount) / float64(stats.TotalExecutions) * 100
	}

	return &stats, nil
}

// ExecutionStats represents execution statistics
type ExecutionStats struct {
	TotalExecutions int64   `json:"total_executions"`
	SuccessCount    int64   `json:"success_count"`
	FailureCount    int64   `json:"failure_count"`
	SuccessRate     float64 `json:"success_rate"`
	AvgDurationMs   int64   `json:"avg_duration_ms"`
}

// Cleanup removes old history records
func (m *HistoryManager) Cleanup(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := m.db.Where("created_at < ?", cutoff).
		Delete(&models.JobHistory{})

	return result.RowsAffected, result.Error
}
