package repositories

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// RefreshSessionRepository handles refresh session database operations
type RefreshSessionRepository struct {
	db *gorm.DB
}

// NewRefreshSessionRepository creates a new repository instance
func NewRefreshSessionRepository(db *gorm.DB) *RefreshSessionRepository {
	return &RefreshSessionRepository{db: db}
}

// HashToken creates SHA-256 hash of refresh token
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// Create creates a new refresh session
func (r *RefreshSessionRepository) Create(ctx context.Context, session *models.RefreshSession) error {
	if session.ID == "" {
		session.ID = uuid.New().String()
	}
	session.CreatedAt = time.Now()
	session.LastUsedAt = time.Now()
	return r.db.WithContext(ctx).Create(session).Error
}

// FindByTokenHash finds session by hashed token
func (r *RefreshSessionRepository) FindByTokenHash(ctx context.Context, tokenHash string) (*models.RefreshSession, error) {
	var session models.RefreshSession
	err := r.db.WithContext(ctx).
		Where("token_hash = ?", tokenHash).
		First(&session).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &session, nil
}

// FindByUserID finds all sessions for a user
func (r *RefreshSessionRepository) FindByUserID(ctx context.Context, userID string) ([]models.RefreshSession, error) {
	var sessions []models.RefreshSession
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// Rotate marks old session as replaced and creates new one (atomic)
func (r *RefreshSessionRepository) Rotate(ctx context.Context, oldTokenHash, newTokenHash string, newSession *models.RefreshSession) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Mark old session as replaced
		if err := tx.Model(&models.RefreshSession{}).
			Where("token_hash = ?", oldTokenHash).
			Updates(map[string]interface{}{
				"replaced_by_hash": newTokenHash,
				"last_used_at":     time.Now(),
			}).Error; err != nil {
			return err
		}

		// Create new session
		newSession.ID = uuid.New().String()
		newSession.TokenHash = newTokenHash
		newSession.CreatedAt = time.Now()
		newSession.LastUsedAt = time.Now()
		return tx.Create(newSession).Error
	})
}

// RevokeByTokenHash revokes a specific session
func (r *RefreshSessionRepository) RevokeByTokenHash(ctx context.Context, tokenHash string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("token_hash = ? AND revoked_at IS NULL", tokenHash).
		Update("revoked_at", now).Error
}

// RevokeAllForUser revokes all sessions for a user (logout all devices)
func (r *RefreshSessionRepository) RevokeAllForUser(ctx context.Context, userID string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL", userID).
		Update("revoked_at", now).Error
}

// RevokeSessionChain revokes entire chain if reuse detected (security incident)
func (r *RefreshSessionRepository) RevokeSessionChain(ctx context.Context, userID string) error {
	// When token reuse is detected, revoke ALL sessions for that user
	// This is a security measure to prevent stolen token abuse
	return r.RevokeAllForUser(ctx, userID)
}

// UpdateTenantID updates the tenant_id of a refresh session (for tenant switching)
func (r *RefreshSessionRepository) UpdateTenantID(ctx context.Context, userID, newTenantID string) error {
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ? AND replaced_by_hash IS NULL",
			userID, time.Now()).
		Update("tenant_id", newTenantID).Error
}

// UpdateLastUsed updates the last_used_at timestamp
func (r *RefreshSessionRepository) UpdateLastUsed(ctx context.Context, tokenHash string) error {
	return r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("token_hash = ?", tokenHash).
		Update("last_used_at", time.Now()).Error
}

// CleanupExpired removes expired sessions older than retention period
func (r *RefreshSessionRepository) CleanupExpired(ctx context.Context, retentionDays int) (int64, error) {
	threshold := time.Now().AddDate(0, 0, -retentionDays)
	result := r.db.WithContext(ctx).
		Where("expires_at < ? OR (revoked_at IS NOT NULL AND revoked_at < ?)", threshold, threshold).
		Delete(&models.RefreshSession{})
	return result.RowsAffected, result.Error
}

// GetActiveSessions returns active sessions for user (for device management UI)
func (r *RefreshSessionRepository) GetActiveSessions(ctx context.Context, userID string) ([]models.RefreshSession, error) {
	var sessions []models.RefreshSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ? AND replaced_by_hash IS NULL",
			userID, time.Now()).
		Order("last_used_at DESC").
		Find(&sessions).Error
	return sessions, err
}

// CountActiveSessions counts active sessions for a user
func (r *RefreshSessionRepository) CountActiveSessions(ctx context.Context, userID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.RefreshSession{}).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ? AND replaced_by_hash IS NULL",
			userID, time.Now()).
		Count(&count).Error
	return count, err
}
