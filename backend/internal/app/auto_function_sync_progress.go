package app

import (
	"encoding/json"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// syncProgress tracks which platforms have been synced in the current run.
// Stored as JSON in AutoFunctionConfig.ProgressData for resume after crash.
type syncProgress struct {
	StartedAt      time.Time `json:"started_at"`
	ShopeeComplete bool      `json:"shopee_complete"`
	TiktokComplete bool      `json:"tiktok_complete"`
	LazadaComplete bool      `json:"lazada_complete"`
	ShopeeCount    int       `json:"shopee_count"`
	TiktokCount    int       `json:"tiktok_count"`
	LazadaCount    int       `json:"lazada_count"`
}

// loadSyncProgress reads progress data from the auto-function config.
// Returns empty progress if no data or parse error.
func loadSyncProgress(cfg *models.AutoFunctionConfig) syncProgress {
	if cfg == nil || cfg.ProgressData == "" {
		return syncProgress{}
	}

	var p syncProgress
	if err := json.Unmarshal([]byte(cfg.ProgressData), &p); err != nil {
		return syncProgress{}
	}
	return p
}

// saveSyncProgress persists progress data to the auto-function config.
// Called after each platform sync completes so we can resume on crash.
func saveSyncProgress(db *gorm.DB, cfg *models.AutoFunctionConfig, p syncProgress) {
	data, err := json.Marshal(p)
	if err != nil {
		log.Warn().Err(err).Msg("[SyncProgress] Failed to marshal progress")
		return
	}

	db.Model(cfg).Update("progress_data", string(data))
}

// clearSyncProgress removes progress data after successful completion.
func clearSyncProgress(db *gorm.DB, cfg *models.AutoFunctionConfig) {
	db.Model(cfg).Update("progress_data", "")
}
