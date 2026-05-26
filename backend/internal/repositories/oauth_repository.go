package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

type OAuthStateCreateParams struct {
	TenantID    string
	Platform    string
	AttemptID   string
	Intent      string
	StoreID     string
	UserID      string
	SessionID   string
	CSRFNonce   string
	State       string
	RedirectURL string
	ExpiresAt   time.Time
}

// OAuthRepository handles OAuth data access
type OAuthRepository struct {
	db *gorm.DB
}

// NewOAuthRepository creates a new OAuth repository
func NewOAuthRepository(db *gorm.DB) *OAuthRepository {
	return &OAuthRepository{db: db}
}

func (r *OAuthRepository) DB() *gorm.DB {
	return r.db
}

// CreateState creates a new OAuth state
func (r *OAuthRepository) CreateState(ctx context.Context, tenantID, platform, redirectURL string) (*models.OAuthState, error) {
	state := &models.OAuthState{
		ID:          uuid.New().String(),
		TenantID:    tenantID,
		Platform:    platform,
		State:       uuid.New().String(),
		RedirectURL: redirectURL,
		ExpiresAt:   time.Now().Add(10 * time.Minute), // 10 minute expiry
		CreatedAt:   time.Now(),
	}
	err := r.db.WithContext(ctx).Create(state).Error
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (r *OAuthRepository) CreateBoundState(ctx context.Context, params OAuthStateCreateParams) (*models.OAuthState, error) {
	if params.TenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if params.Platform == "" {
		return nil, fmt.Errorf("platform is required")
	}
	if params.AttemptID == "" {
		return nil, fmt.Errorf("attempt_id is required")
	}
	if params.State == "" {
		return nil, fmt.Errorf("state is required")
	}
	state := &models.OAuthState{
		ID:          uuid.New().String(),
		TenantID:    params.TenantID,
		Platform:    params.Platform,
		AttemptID:   params.AttemptID,
		Intent:      params.Intent,
		StoreID:     params.StoreID,
		UserID:      params.UserID,
		SessionID:   params.SessionID,
		CSRFNonce:   params.CSRFNonce,
		State:       params.State,
		RedirectURL: params.RedirectURL,
		ExpiresAt:   params.ExpiresAt,
		CreatedAt:   time.Now(),
	}
	if state.Intent == "" {
		state.Intent = "connect"
	}
	if err := r.db.WithContext(ctx).Create(state).Error; err != nil {
		return nil, err
	}
	return state, nil
}

func (r *OAuthRepository) ConsumeState(ctx context.Context, stateValue string) (*models.OAuthState, string, error) {
	if stateValue == "" {
		return nil, "invalid_state", nil
	}
	var oauthState models.OAuthState
	err := r.db.WithContext(ctx).Where("state = ?", stateValue).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, "invalid_state", nil
		}
		return nil, "failed", err
	}
	if time.Now().After(oauthState.ExpiresAt) {
		return &oauthState, "expired", nil
	}
	if err := r.db.WithContext(ctx).Where("id = ?", oauthState.ID).Delete(&models.OAuthState{}).Error; err != nil {
		return nil, "failed", err
	}
	return &oauthState, "completed", nil
}

// FindStateByState finds OAuth state by state value
func (r *OAuthRepository) FindStateByState(ctx context.Context, state string) (*models.OAuthState, error) {
	var oauthState models.OAuthState
	err := r.db.WithContext(ctx).Where("state = ?", state).First(&oauthState).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &oauthState, nil
}

// DeleteState deletes an OAuth state
func (r *OAuthRepository) DeleteState(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.OAuthState{}).Error
}

// DeleteExpiredStates deletes all expired OAuth states
func (r *OAuthRepository) DeleteExpiredStates(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&models.OAuthState{}).Error
}

// CreateLog creates a new OAuth log entry
func (r *OAuthRepository) CreateLog(ctx context.Context, log *models.OAuthLog) error {
	if log.ID == "" {
		log.ID = uuid.New().String()
	}
	log.CreatedAt = time.Now()
	return r.db.WithContext(ctx).Create(log).Error
}

// UpdateLogStatus updates OAuth log status
func (r *OAuthRepository) UpdateLogStatus(ctx context.Context, id, status, errorMsg string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"status":       status,
		"processed_at": now,
	}
	if errorMsg != "" {
		updates["error_msg"] = errorMsg
	}
	return r.db.WithContext(ctx).Model(&models.OAuthLog{}).
		Where("id = ?", id).Updates(updates).Error
}

// FindLogsByTenant finds OAuth logs by tenant
func (r *OAuthRepository) FindLogsByTenant(ctx context.Context, tenantID string, limit int) ([]models.OAuthLog, error) {
	var logs []models.OAuthLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}

// FindLogsByPlatform finds OAuth logs by platform
func (r *OAuthRepository) FindLogsByPlatform(ctx context.Context, tenantID, platform string, limit int) ([]models.OAuthLog, error) {
	var logs []models.OAuthLog
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs).Error
	return logs, err
}
