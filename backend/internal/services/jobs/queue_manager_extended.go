package jobs

import (
	"fmt"
	"sort"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// GetCurrentRunningJob returns the currently running job
// Note: Uses silent logger to avoid "record not found" log spam when no jobs are running
func (m *QueueManager) GetCurrentRunningJob() *models.Job {
	var job models.Job
	// Use silent session to suppress "record not found" logs for expected empty results
	err := m.db.Session(&gorm.Session{Logger: logger.Default.LogMode(logger.Silent)}).
		Where("status = ?", models.JobStatusRunning).First(&job).Error
	if err != nil {
		return nil
	}
	return &job
}

// GetJobHistoryByID retrieves job history by job ID
func (m *QueueManager) GetJobHistoryByID(jobID string, limit int) []models.JobHistory {
	var history []models.JobHistory
	m.db.Where("job_id = ?", jobID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history)
	return history
}

// GetPendingQueue returns pending jobs in queue
func (m *QueueManager) GetPendingQueue(limit int) []models.Job {
	var jobs []models.Job
	m.db.Where("status = ?", models.JobStatusPending).
		Order("priority DESC, created_at ASC").
		Limit(limit).
		Find(&jobs)
	return jobs
}

// GetQueueStatus returns full queue status
func (m *QueueManager) GetQueueStatus() QueueStatus {
	currentJob := m.GetCurrentRunningJob()
	pendingQueue := m.GetPendingQueue(10)

	var totalPending, totalCompleted int64
	m.db.Model(&models.Job{}).Where("status = ?", models.JobStatusPending).Count(&totalPending)
	m.db.Model(&models.Job{}).Where("status = ?", models.JobStatusCompleted).Count(&totalCompleted)

	return QueueStatus{
		CurrentJob:     currentJob,
		PendingQueue:   pendingQueue,
		TotalPending:   totalPending,
		TotalCompleted: totalCompleted,
	}
}

// GetRecentHistory returns recent job history
func (m *QueueManager) GetRecentHistory(limit int) []models.JobHistory {
	var history []models.JobHistory
	m.db.Order("created_at DESC").
		Limit(limit).
		Find(&history)
	return history
}

// GetHistoryPaginated returns paginated job history INCLUDING auto_functions_history
// Returns data in camelCase format matching Node.js backend for frontend compatibility
func (m *QueueManager) GetHistoryPaginated(filter PaginatedFilter) (*PaginatedHistoryResult, error) {
	// Get job_history items
	var jobHistoryItems []models.JobHistory
	jobQuery := m.db.Model(&models.JobHistory{})

	if filter.Status != "" && filter.Status != "all" {
		jobQuery = jobQuery.Where("status = ?", filter.Status)
	}
	if filter.JobType != "" && filter.JobType != "all" {
		jobQuery = jobQuery.Where("job_type = ?", filter.JobType)
	}
	if filter.Search != "" {
		jobQuery = jobQuery.Where("job_id ILIKE ? OR job_type ILIKE ?", "%"+filter.Search+"%", "%"+filter.Search+"%")
	}

	if err := jobQuery.Order("created_at DESC").Limit(1000).Find(&jobHistoryItems).Error; err != nil {
		log.Error().Err(err).Msg("Failed to query job_history")
	}

	// Get auto_functions_history items
	var autoFuncHistoryItems []models.AutoFunctionHistory
	autoQuery := m.db.Model(&models.AutoFunctionHistory{})

	if filter.Status != "" && filter.Status != "all" {
		autoQuery = autoQuery.Where("status = ?", filter.Status)
	}
	if filter.JobType != "" && filter.JobType != "all" {
		autoQuery = autoQuery.Where("function_name = ?", filter.JobType)
	}
	if filter.Search != "" {
		autoQuery = autoQuery.Where("function_name ILIKE ?", "%"+filter.Search+"%")
	}

	if err := autoQuery.Order("executed_at DESC").Limit(1000).Find(&autoFuncHistoryItems).Error; err != nil {
		log.Error().Err(err).Msg("Failed to query auto_functions_history")
	}

	// Merge and convert to snake_case response format
	allItems := make([]HistoryItemResponse, 0, len(jobHistoryItems)+len(autoFuncHistoryItems))

	// Convert job_history to response format
	for _, h := range jobHistoryItems {
		allItems = append(allItems, HistoryItemResponse{
			ID:           h.ID,
			JobID:        h.JobID,
			JobType:      h.JobType,
			Status:       string(h.Status),
			ErrorMessage: h.ErrorMessage,
			DurationMs:   h.DurationMs,
			StartedAt:    h.StartedAt,
			CompletedAt:  h.CompletedAt,
			CreatedAt:    h.CreatedAt,
		})
	}

	// Convert auto_functions_history to response format
	for _, h := range autoFuncHistoryItems {
		completedAt := h.ExecutedAt.Add(time.Duration(h.DurationMs) * time.Millisecond)
		allItems = append(allItems, HistoryItemResponse{
			ID:           int(h.ID),
			JobID:        fmt.Sprintf("auto_%d", h.ID),
			JobType:      h.FunctionName,
			Status:       h.Status,
			ErrorMessage: h.ErrorMessage,
			DurationMs:   h.DurationMs,
			StartedAt:    &h.ExecutedAt,
			CompletedAt:  &completedAt,
			CreatedAt:    h.ExecutedAt,
		})
	}

	// Sort by CreatedAt
	sort.Slice(allItems, func(i, j int) bool {
		if filter.SortOrder == "asc" {
			return allItems[i].CreatedAt.Before(allItems[j].CreatedAt)
		}
		return allItems[i].CreatedAt.After(allItems[j].CreatedAt)
	})

	// Calculate pagination
	total := int64(len(allItems))
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}

	offset := (filter.Page - 1) * filter.PageSize
	end := offset + filter.PageSize
	if end > len(allItems) {
		end = len(allItems)
	}
	if offset > len(allItems) {
		offset = len(allItems)
	}

	paginatedItems := allItems[offset:end]

	totalPages := int(total) / filter.PageSize
	if int(total)%filter.PageSize > 0 {
		totalPages++
	}

	return &PaginatedHistoryResult{
		Data:       paginatedItems,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetDistinctJobTypes returns distinct job types from both job_history and auto_functions_history
func (m *QueueManager) GetDistinctJobTypes() []string {
	typeSet := make(map[string]bool)

	// Get from job_history
	var jobTypes []string
	m.db.Model(&models.JobHistory{}).
		Distinct("job_type").
		Pluck("job_type", &jobTypes)
	for _, t := range jobTypes {
		if t != "" {
			typeSet[t] = true
		}
	}

	// Get from auto_functions_history
	var funcNames []string
	m.db.Model(&models.AutoFunctionHistory{}).
		Distinct("function_name").
		Pluck("function_name", &funcNames)
	for _, n := range funcNames {
		if n != "" {
			typeSet[n] = true
		}
	}

	// Convert set to slice
	result := make([]string, 0, len(typeSet))
	for t := range typeSet {
		result = append(result, t)
	}
	sort.Strings(result)
	return result
}

// ClearJobHistory clears all job history including auto_functions_history
func (m *QueueManager) ClearJobHistory() int64 {
	result := m.db.Delete(&models.JobHistory{}, "1=1")
	autoResult := m.db.Delete(&models.AutoFunctionHistory{}, "1=1")
	return result.RowsAffected + autoResult.RowsAffected
}

// CancelJobByStringID cancels a job by string ID
func (m *QueueManager) CancelJobByStringID(jobID string, force bool) bool {
	query := m.db.Model(&models.Job{}).Where("id = ?", jobID)

	if !force {
		query = query.Where("status = ?", models.JobStatusPending)
	}

	result := query.Update("status", models.JobStatusCancelled)
	return result.RowsAffected > 0
}

// CheckAndTimeoutStuckJobs checks for stuck jobs and marks them as timed out
func (m *QueueManager) CheckAndTimeoutStuckJobs(timeoutMinutes int) []models.Job {
	cutoff := time.Now().Add(-time.Duration(timeoutMinutes) * time.Minute)

	var stuckJobs []models.Job
	m.db.Where("status = ? AND started_at < ?",
		models.JobStatusRunning, cutoff).
		Find(&stuckJobs)

	if len(stuckJobs) > 0 {
		var ids []string
		for _, j := range stuckJobs {
			ids = append(ids, j.ID)
		}
		m.db.Model(&models.Job{}).Where("id IN ?", ids).
			Updates(map[string]interface{}{
				"status":        models.JobStatusFailed,
				"error_message": "Job timed out",
			})
	}

	return stuckJobs
}

// AddToHistory adds a job record to history
func (m *QueueManager) AddToHistory(job *models.Job, durationMs int) error {
	history := &models.JobHistory{
		JobID:        job.ID,
		JobType:      job.Type,
		Status:       job.Status,
		ErrorMessage: job.ErrorMessage,
		DurationMs:   durationMs,
		StartedAt:    job.StartedAt,
		CompletedAt:  job.CompletedAt,
	}
	return m.db.Create(history).Error
}
