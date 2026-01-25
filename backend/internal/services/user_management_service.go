package services

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// UserManagementService handles user management operations
type UserManagementService struct {
	userRepo  *repositories.UserRepository
	auditRepo *repositories.AuditRepository
	authSvc   *AuthService
}

// NewUserManagementService creates a new user management service
func NewUserManagementService(userRepo *repositories.UserRepository, auditRepo *repositories.AuditRepository, authSvc *AuthService) *UserManagementService {
	return &UserManagementService{
		userRepo:  userRepo,
		auditRepo: auditRepo,
		authSvc:   authSvc,
	}
}

// CreateUserRequest represents create user request
type CreateUserRequest struct {
	Username string
	Email    string
	Password string
	Role     string
	TenantID string // For audit logging only
}

// UpdateUserRequest represents update user request
type UpdateUserRequest struct {
	Username string
	Email    string
	Role     string
}

// CreateUser creates a new user
func (s *UserManagementService) CreateUser(ctx context.Context, req *CreateUserRequest, createdBy string) (*models.User, error) {
	// Validate role
	if !models.IsValidRole(req.Role) {
		return nil, errors.New("invalid role")
	}

	// Check if username exists
	exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("username already exists")
	}

	// Check if email exists
	if req.Email != "" {
		exists, err = s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already exists")
		}
	}

	// Hash password
	passwordHash, err := s.authSvc.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  passwordHash,
		Role:      req.Role,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Audit log
	s.logUserAction(ctx, models.AuditActionUserCreated, createdBy, user.ID, req.TenantID)

	return user, nil
}

// UpdateUser updates an existing user
func (s *UserManagementService) UpdateUser(ctx context.Context, userID string, req *UpdateUserRequest, updatedBy string) (*models.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Validate role if provided
	if req.Role != "" && !models.IsValidRole(req.Role) {
		return nil, errors.New("invalid role")
	}

	// Check username uniqueness
	if req.Username != "" && req.Username != user.Username {
		exists, err := s.userRepo.ExistsByUsername(ctx, req.Username)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("username already exists")
		}
		user.Username = req.Username
	}

	// Check email uniqueness
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmail(ctx, req.Email)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("email already exists")
		}
		user.Email = req.Email
	}

	if req.Role != "" {
		user.Role = req.Role
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	// Audit log - tenantID comes from request context, not user
	s.logUserAction(ctx, models.AuditActionUserUpdated, updatedBy, user.ID, "")

	return user, nil
}

// DeleteUser deletes a user
func (s *UserManagementService) DeleteUser(ctx context.Context, userID, deletedBy, tenantID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return err
	}

	// Audit log
	s.logUserAction(ctx, models.AuditActionUserDeleted, deletedBy, userID, tenantID)

	return nil
}

// GetUser gets a user by ID
func (s *UserManagementService) GetUser(ctx context.Context, userID string) (*models.User, error) {
	return s.userRepo.FindByID(ctx, userID)
}

// GetUserByUsername gets a user by username
func (s *UserManagementService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return s.userRepo.FindByUsername(ctx, username)
}

// ListUsers lists all users
func (s *UserManagementService) ListUsers(ctx context.Context) ([]models.User, error) {
	return s.userRepo.FindAll(ctx)
}

// UnlockUser unlocks a locked user account
func (s *UserManagementService) UnlockUser(ctx context.Context, userID, unlockedBy, tenantID string) error {
	if err := s.userRepo.ResetFailedAttempts(ctx, userID); err != nil {
		return err
	}

	s.logUserAction(ctx, "user_unlocked", unlockedBy, userID, tenantID)
	return nil
}

// ResetPassword resets user password (admin action)
func (s *UserManagementService) ResetPassword(ctx context.Context, userID, newPassword, resetBy, tenantID string) error {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	hash, err := s.authSvc.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = hash
	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	s.logUserAction(ctx, models.AuditActionPasswordReset, resetBy, userID, tenantID)
	return nil
}

// logUserAction logs user management actions
func (s *UserManagementService) logUserAction(ctx context.Context, action, userID, targetUserID, tenantID string) {
	if s.auditRepo == nil {
		return
	}
	s.auditRepo.Create(ctx, &models.AuditLogEntry{
		TenantID:     tenantID,
		Action:       action,
		UserID:       userID,
		TargetUserID: targetUserID,
		Status:       "success",
	})
}
