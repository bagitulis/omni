package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// UserRepository handles user data access
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsername finds a user by username
func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindByUsernameOrEmail finds a user by username or email
func (r *UserRepository) FindByUsernameOrEmail(ctx context.Context, usernameOrEmail string) (*models.User, error) {
	var user models.User
	err := r.db.WithContext(ctx).
		Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

// FindAll finds all users
func (r *UserRepository) FindAll(ctx context.Context) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&users).Error
	return users, err
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Create(user).Error
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *models.User) error {
	user.UpdatedAt = time.Now()
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user by ID
func (r *UserRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.User{}).Error
}

// IncrementFailedAttempts increments failed login attempts
func (r *UserRepository) IncrementFailedAttempts(ctx context.Context, id string) error {
	now := time.Now()
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failed_login_attempts": gorm.Expr("failed_login_attempts + 1"),
			"last_failed_login":     now,
			"updated_at":            now,
		}).Error
}

// ResetFailedAttempts resets failed login attempts
func (r *UserRepository) ResetFailedAttempts(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"failed_login_attempts": 0,
			"account_locked_until":  nil,
			"updated_at":            time.Now(),
		}).Error
}

// LockAccount locks the user account
func (r *UserRepository) LockAccount(ctx context.Context, id string, duration time.Duration) error {
	lockUntil := time.Now().Add(duration)
	return r.db.WithContext(ctx).Model(&models.User{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"account_locked_until": lockUntil,
			"updated_at":           time.Now(),
		}).Error
}

// ExistsByUsername checks if username exists
func (r *UserRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// ExistsByEmail checks if email exists
func (r *UserRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).
		Where("email = ?", email).Count(&count).Error
	return count > 0, err
}

// Count returns the total number of users
func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Count(&count).Error
	return count, err
}

// FindPaginated returns users with pagination
func (r *UserRepository) FindPaginated(ctx context.Context, offset, limit int) ([]models.User, error) {
	var users []models.User
	err := r.db.WithContext(ctx).Order("created_at DESC").Offset(offset).Limit(limit).Find(&users).Error
	return users, err
}

// CountByRole returns the number of users with a specific role
func (r *UserRepository) CountByRole(ctx context.Context, role string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.User{}).Where("role = ?", role).Count(&count).Error
	return count, err
}
