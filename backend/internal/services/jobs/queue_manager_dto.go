// Package jobs provides job queue DTOs
package jobs

import (
	"time"

	"github.com/omni/backend/internal/models"
)

// HistoryItemResponse is the snake_case response format matching API standards
// This ensures frontend compatibility
type HistoryItemResponse struct {
	ID           int        `json:"id"`
	JobID        string     `json:"job_id"`
	JobType      string     `json:"job_type"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"error_message,omitempty"`
	DurationMs   int        `json:"duration_ms"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// PaginatedHistoryResult for paginated history response with snake_case
type PaginatedHistoryResult struct {
	Data       []HistoryItemResponse `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// QueueStatus represents queue status
type QueueStatus struct {
	CurrentJob     *models.Job  `json:"current_job"`
	PendingQueue   []models.Job `json:"pending_queue"`
	TotalPending   int64        `json:"total_pending"`
	TotalCompleted int64        `json:"total_completed"`
}

// PaginatedFilter for paginated history
type PaginatedFilter struct {
	Page      int
	PageSize  int
	Status    string
	JobType   string
	Search    string
	SortBy    string
	SortOrder string
}
