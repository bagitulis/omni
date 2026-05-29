package jobs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// QueueManager manages job queue operations
// Note: Multi-tenancy is handled by schema isolation - no tenant_id column in database
type QueueManager struct {
	db *gorm.DB
}

// NewQueueManager creates a new queue manager
func NewQueueManager(db *gorm.DB, _ string) *QueueManager {
	// tenantID parameter kept for backward compatibility but not used
	// Schema isolation handles multi-tenancy (search_path already set)
	return &QueueManager{db: db}
}

// AddJob adds a new job to the queue
func (m *QueueManager) AddJob(req models.CreateJobRequest) (*models.Job, error) {
	jobID := req.ID
	if jobID == "" {
		jobID = uuid.New().String()
	}

	dataJSON := req.Data
	if dataJSON == "" {
		dataJSON = "{}"
	}

	priority := req.Priority
	if priority == "" {
		priority = "normal"
	}

	job := &models.Job{
		ID:       jobID,
		Type:     req.Type,
		Data:     dataJSON,
		Priority: priority,
		Status:   models.JobStatusPending,
	}

	if err := m.db.Create(job).Error; err != nil {
		return nil, err
	}

	return job, nil
}

// GetPendingJobs retrieves pending jobs ordered by priority
func (m *QueueManager) GetPendingJobs(limit int) ([]models.Job, error) {
	var jobs []models.Job

	err := m.db.Where("status = ?", models.JobStatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&jobs).Error

	return jobs, err
}

// ClaimNextJob atomically claims the next pending job using SELECT FOR UPDATE SKIP LOCKED.
// This prevents multiple workers from picking the same job.
func (m *QueueManager) ClaimNextJob() (*models.Job, error) {
	var job models.Job
	now := time.Now()

	err := m.db.Transaction(func(tx *gorm.DB) error {
		// Atomic: select + lock + skip already locked rows
		if err := tx.Raw(`
			SELECT * FROM jobs
			WHERE status = ?
			ORDER BY
				CASE priority WHEN 'high' THEN 0 WHEN 'normal' THEN 1 ELSE 2 END,
				created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		`, models.JobStatusPending).Scan(&job).Error; err != nil {
			return err
		}

		if job.ID == "" {
			return gorm.ErrRecordNotFound
		}

		// Atomically set to running
		return tx.Model(&models.Job{}).Where("id = ?", job.ID).Updates(map[string]interface{}{
			"status":     models.JobStatusRunning,
			"started_at": &now,
			"updated_at": now,
		}).Error
	})

	if err != nil {
		return nil, err
	}

	job.Status = models.JobStatusRunning
	job.StartedAt = &now
	return &job, nil
}

// RecoverZombieJobs resets 'running' jobs back to 'failed' on startup.
// These are jobs that were interrupted by a server crash.
func (m *QueueManager) RecoverZombieJobs() (int64, error) {
	result := m.db.Model(&models.Job{}).
		Where("status = ?", models.JobStatusRunning).
		Updates(map[string]interface{}{
			"status":        models.JobStatusFailed,
			"error_message": "recovered: server restarted while job was running",
			"completed_at":  time.Now(),
			"updated_at":    time.Now(),
		})
	return result.RowsAffected, result.Error
}

// UpdateStatus updates job status
func (m *QueueManager) UpdateStatus(jobID string, status models.JobStatus, errMsg string) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if errMsg != "" {
		updates["error_message"] = errMsg
	}

	switch status {
	case models.JobStatusRunning:
		now := time.Now()
		updates["started_at"] = &now
	case models.JobStatusCompleted, models.JobStatusFailed:
		now := time.Now()
		updates["completed_at"] = &now
	}

	return m.db.Model(&models.Job{}).Where("id = ?", jobID).Updates(updates).Error
}

// GetJob retrieves a job by ID
func (m *QueueManager) GetJob(jobID string) (*models.Job, error) {
	var job models.Job
	err := m.db.Where("id = ?", jobID).First(&job).Error
	if err != nil {
		return nil, err
	}
	return &job, nil
}

// ListJobs lists jobs with filters
func (m *QueueManager) ListJobs(filter models.JobFilter) ([]models.Job, int64, error) {
	query := m.db.Model(&models.Job{})

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.Type != "" {
		query = query.Where("type = ?", filter.Type)
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

	var jobs []models.Job
	err := query.Order("created_at DESC").Offset(offset).Limit(filter.PageSize).Find(&jobs).Error

	return jobs, total, err
}

// CancelJob cancels a pending job
func (m *QueueManager) CancelJob(jobID string) error {
	return m.db.Model(&models.Job{}).
		Where("id = ? AND status = ?", jobID, models.JobStatusPending).
		Update("status", models.JobStatusCancelled).Error
}

// CleanupOldJobs removes completed jobs older than specified duration
func (m *QueueManager) CleanupOldJobs(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan)
	result := m.db.Where("status IN ? AND completed_at < ?",
		[]models.JobStatus{models.JobStatusCompleted, models.JobStatusFailed, models.JobStatusCancelled},
		cutoff).
		Delete(&models.Job{})

	return result.RowsAffected, result.Error
}

// GetJobStats returns job statistics
func (m *QueueManager) GetJobStats() (map[string]interface{}, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var counts []StatusCount
	err := m.db.Model(&models.Job{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&counts).Error

	if err != nil {
		return nil, err
	}

	stats := make(map[string]interface{})
	for _, c := range counts {
		stats[c.Status] = c.Count
	}

	return stats, nil
}

// SerializePayload converts payload to JSON string
func SerializePayload(payload interface{}) (string, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// EnqueueJob enqueues a job with type, data, and priority
func (m *QueueManager) EnqueueJob(jobType string, data map[string]interface{}, priority string) (string, error) {
	dataJSON, err := SerializePayload(data)
	if err != nil {
		return "", err
	}

	job := &models.Job{
		ID:       uuid.New().String(),
		Type:     jobType,
		Data:     dataJSON,
		Priority: priority,
		Status:   models.JobStatusPending,
	}

	if err := m.db.Create(job).Error; err != nil {
		return "", err
	}

	return job.ID, nil
}

// UpdateProgress updates job progress information
func (m *QueueManager) UpdateProgress(jobID string, percent, processed, total int, message string) error {
	return m.db.Model(&models.Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"progress_percent": percent,
		"progress_message": message,
		"processed_items":  processed,
		"total_items":      total,
		"updated_at":       time.Now(),
	}).Error
}

// CompleteJobWithResult marks job as completed and stores result data
func (m *QueueManager) CompleteJobWithResult(jobID string, resultData interface{}) error {
	resultJSON, err := SerializePayload(resultData)
	if err != nil {
		return err
	}

	now := time.Now()
	return m.db.Model(&models.Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":           models.JobStatusCompleted,
		"progress_percent": 100,
		"result_data":      resultJSON,
		"completed_at":     &now,
		"updated_at":       now,
	}).Error
}

// FailJob marks job as failed with error message
func (m *QueueManager) FailJob(jobID string, errMsg string) error {
	now := time.Now()
	return m.db.Model(&models.Job{}).Where("id = ?", jobID).Updates(map[string]interface{}{
		"status":        models.JobStatusFailed,
		"error_message": errMsg,
		"completed_at":  &now,
		"updated_at":    now,
	}).Error
}

// GetRunningJobs returns all currently running jobs
func (m *QueueManager) GetRunningJobs() ([]models.Job, error) {
	var jobs []models.Job
	err := m.db.Where("status = ?", models.JobStatusRunning).
		Order("started_at ASC").
		Find(&jobs).Error
	return jobs, err
}

// GetJobsByType returns jobs of a specific type with status filter
func (m *QueueManager) GetJobsByType(jobType string, status models.JobStatus) ([]models.Job, error) {
	var jobs []models.Job
	query := m.db.Where("type = ?", jobType)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("created_at DESC").Limit(1000).Find(&jobs).Error
	return jobs, err
}
