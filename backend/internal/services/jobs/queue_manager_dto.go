// Package jobs provides job queue DTOs
package jobs

import (
	"time"

	"github.com/omni/backend/internal/models"
)

// HistoryItemResponse is the camelCase response format matching Node.js
// This ensures frontend compatibility
type HistoryItemResponse struct {
	ID           int        `json:"id"`
	JobID        string     `json:"jobId"`
	JobType      string     `json:"jobType"`
	Status       string     `json:"status"`
	ErrorMessage string     `json:"errorMessage,omitempty"`
	DurationMs   int        `json:"durationMs"`
	StartedAt    *time.Time `json:"startedAt,omitempty"`
	CompletedAt  *time.Time `json:"completedAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

// PaginatedHistoryResult for paginated history response with camelCase
type PaginatedHistoryResult struct {
	Data       []HistoryItemResponse `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"pageSize"`
	TotalPages int                   `json:"totalPages"`
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
