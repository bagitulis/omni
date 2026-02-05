package handlers

import (
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestMergeHistories tests the mergeHistories function
func TestMergeHistories_Empty(t *testing.T) {
	result := mergeHistories(nil, nil)
	assert.NotNil(t, result)
	assert.Len(t, result, 0)
}

// TestMergeHistories_OnlyJobHistory tests with only job history
func TestMergeHistories_OnlyJobHistory(t *testing.T) {
	now := time.Now()
	jobHistory := []models.JobHistory{
		{
			ID:           1,
			JobID:        "job-1",
			JobType:      "sync",
			Status:       models.JobStatusCompleted,
			ErrorMessage: "",
			DurationMs:   100,
			StartedAt:    &now,
			CreatedAt:    now,
		},
	}

	result := mergeHistories(jobHistory, nil)
	assert.Len(t, result, 1)
	assert.Equal(t, "job-1", result[0].JobID)
	assert.Equal(t, "sync", result[0].JobType)
}

// TestMergeHistories_OnlyAutoFuncHistory tests with only auto function history
func TestMergeHistories_OnlyAutoFuncHistory(t *testing.T) {
	now := time.Now()
	autoFuncHistory := []models.AutoFunctionHistory{
		{
			ID:           1,
			FunctionName: "auto-sync",
			Status:       "completed",
			ErrorMessage: "",
			DurationMs:   50,
			ExecutedAt:   now,
		},
	}

	result := mergeHistories(nil, autoFuncHistory)
	assert.Len(t, result, 1)
	assert.Equal(t, "auto_1", result[0].JobID)
	assert.Equal(t, "auto-sync", result[0].JobType)
}

// TestMergeHistories_Combined tests merging both histories
func TestMergeHistories_Combined(t *testing.T) {
	now := time.Now()
	earlier := now.Add(-time.Hour)

	jobHistory := []models.JobHistory{
		{
			ID:        1,
			JobID:     "job-1",
			JobType:   "sync",
			Status:    models.JobStatusCompleted,
			CreatedAt: earlier,
		},
	}

	autoFuncHistory := []models.AutoFunctionHistory{
		{
			ID:           1,
			FunctionName: "auto-sync",
			Status:       "completed",
			ExecutedAt:   now,
		},
	}

	result := mergeHistories(jobHistory, autoFuncHistory)
	assert.Len(t, result, 2)
	// Should be sorted by created_at DESC, so auto-func should be first
	assert.Equal(t, "auto_1", result[0].JobID)
	assert.Equal(t, "job-1", result[1].JobID)
}

// TestMergeHistories_Limit10 tests that result is limited to 10 items
func TestMergeHistories_Limit10(t *testing.T) {
	now := time.Now()
	jobHistory := make([]models.JobHistory, 15)
	for i := 0; i < 15; i++ {
		jobHistory[i] = models.JobHistory{
			ID:        i + 1,
			JobID:     "job-" + string(rune('a'+i)),
			JobType:   "sync",
			Status:    models.JobStatusCompleted,
			CreatedAt: now.Add(time.Duration(-i) * time.Minute),
		}
	}

	result := mergeHistories(jobHistory, nil)
	assert.Len(t, result, 10)
}

// TestUnifiedHistoryItem_Structure tests the unified history item structure
func TestUnifiedHistoryItem_Structure(t *testing.T) {
	item := UnifiedHistoryItem{
		ID:           1,
		JobID:        "test-job",
		JobType:      "sync",
		Status:       "completed",
		ErrorMessage: "",
		DurationMs:   100,
		CreatedAt:    time.Now(),
	}

	assert.Equal(t, 1, item.ID)
	assert.Equal(t, "test-job", item.JobID)
	assert.Equal(t, "sync", item.JobType)
	assert.Equal(t, "completed", item.Status)
	assert.Equal(t, 100, item.DurationMs)
}
