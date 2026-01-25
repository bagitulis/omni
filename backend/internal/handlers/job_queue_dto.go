package handlers

import (
	"fmt"
	"sort"
	"time"

	"github.com/omni/backend/internal/models"
)

// =============================================================================
// Job Queue DTOs (Data Transfer Objects)
// =============================================================================

// UnifiedHistoryItem represents a unified history item from both tables
type UnifiedHistoryItem struct {
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

// =============================================================================
// Helper Functions
// =============================================================================

// mergeHistories merges job_history and auto_functions_history into a unified list
func mergeHistories(jobHistory []models.JobHistory, autoFuncHistory []models.AutoFunctionHistory) []UnifiedHistoryItem {
	combined := make([]UnifiedHistoryItem, 0, len(jobHistory)+len(autoFuncHistory))

	// Add job history
	for _, h := range jobHistory {
		combined = append(combined, UnifiedHistoryItem{
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

	// Add auto function history (convert to unified format)
	for _, h := range autoFuncHistory {
		combined = append(combined, UnifiedHistoryItem{
			ID:           int(h.ID),
			JobID:        fmt.Sprintf("auto_%d", h.ID),
			JobType:      h.FunctionName,
			Status:       h.Status,
			ErrorMessage: h.ErrorMessage,
			DurationMs:   h.DurationMs,
			StartedAt:    &h.ExecutedAt,
			CompletedAt:  nil,
			CreatedAt:    h.ExecutedAt,
		})
	}

	// Sort by created_at DESC
	sort.Slice(combined, func(i, j int) bool {
		return combined[i].CreatedAt.After(combined[j].CreatedAt)
	})

	// Limit to 10 most recent
	if len(combined) > 10 {
		combined = combined[:10]
	}

	return combined
}
