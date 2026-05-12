package services

import (
	"context"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupUserTestDB creates an in-memory SQLite database for user management tests
func setupUserTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	// Only migrate User table - AuditLogEntry has map[string]interface{} which SQLite doesn't support
	err = db.AutoMigrate(&models.User{})
	require.NoError(t, err)

	return db
}

// createUserManagementService creates a UserManagementService for testing
// Note: auditRepo is nil to avoid SQLite compatibility issues with Details field
func createUserManagementService(t *testing.T, db *gorm.DB) *UserManagementService {
	t.Helper()
	userRepo := repositories.NewUserRepository(db)
	// Pass nil for auditRepo to skip audit logging in tests
	jwtSvc := utils.NewJWTService("test-secret-key")
	authSvc := NewAuthService(userRepo, nil, jwtSvc)
	return NewUserManagementService(userRepo, nil, authSvc)
}

func TestUserManagementService_CreateUser(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	tests := []struct {
		name       string
		req        *CreateUserRequest
		wantErr    bool
		errContain string
	}{
		{
			name: "valid user",
			req: &CreateUserRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "SecurePassword123!",
				Role:     models.RoleAdmin,
				TenantID: "test-tenant",
			},
			wantErr: false,
		},
		{
			name: "invalid role",
			req: &CreateUserRequest{
				Username: "testuser2",
				Email:    "test2@example.com",
				Password: "SecurePassword123!",
				Role:     "invalid_role",
				TenantID: "test-tenant",
			},
			wantErr:    true,
			errContain: "invalid role",
		},
		{
			name: "owner role",
			req: &CreateUserRequest{
				Username: "owneruser",
				Email:    "owner@example.com",
				Password: "SecurePassword123!",
				Role:     models.RoleOwner,
				TenantID: "test-tenant",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := svc.CreateUser(ctx, tt.req, "admin-user")

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContain != "" {
					assert.Contains(t, err.Error(), tt.errContain)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, user.ID)
				assert.Equal(t, tt.req.Username, user.Username)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.Equal(t, tt.req.Role, user.Role)
				assert.NotEmpty(t, user.Password) // Password is hashed
			}
		})
	}
}

func TestUserManagementService_CreateUser_DuplicateUsername(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create first user
	req1 := &CreateUserRequest{
		Username: "duplicate",
		Email:    "first@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	_, err := svc.CreateUser(ctx, req1, "admin")
	require.NoError(t, err)

	// Try to create second user with same username
	req2 := &CreateUserRequest{
		Username: "duplicate",
		Email:    "second@example.com",
		Password: "AnotherPassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	_, err = svc.CreateUser(ctx, req2, "admin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "username already exists")
}

func TestUserManagementService_CreateUser_DuplicateEmail(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create first user
	req1 := &CreateUserRequest{
		Username: "user1",
		Email:    "duplicate@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	_, err := svc.CreateUser(ctx, req1, "admin")
	require.NoError(t, err)

	// Try to create second user with same email
	req2 := &CreateUserRequest{
		Username: "user2",
		Email:    "duplicate@example.com",
		Password: "AnotherPassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	_, err = svc.CreateUser(ctx, req2, "admin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "email already exists")
}

func TestUserManagementService_GetUser(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "getuser",
		Email:    "getuser@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)

	// Get the user
	user, err := svc.GetUser(ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, user.ID)
	assert.Equal(t, created.Username, user.Username)
}

func TestUserManagementService_GetUserByUsername(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "findbyname",
		Email:    "findbyname@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)

	// Get by username
	user, err := svc.GetUserByUsername(ctx, "findbyname")
	require.NoError(t, err)
	assert.Equal(t, created.ID, user.ID)
}

func TestUserManagementService_UpdateUser(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "updateuser",
		Email:    "updateuser@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)

	// Update the user
	updateReq := &UpdateUserRequest{
		Username: "updatedname",
		Email:    "updated@example.com",
		Role:     models.RoleAdmin,
	}
	updated, err := svc.UpdateUser(ctx, created.ID, updateReq, "admin")
	require.NoError(t, err)
	assert.Equal(t, "updatedname", updated.Username)
	assert.Equal(t, "updated@example.com", updated.Email)
	assert.Equal(t, models.RoleAdmin, updated.Role)
}

func TestUserManagementService_UpdateUser_InvalidRole(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "roletest",
		Email:    "roletest@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)

	// Try to update with invalid role
	updateReq := &UpdateUserRequest{
		Role: "super_admin",
	}
	_, err = svc.UpdateUser(ctx, created.ID, updateReq, "admin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid role")
}

func TestUserManagementService_UpdateUser_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	updateReq := &UpdateUserRequest{
		Username: "newname",
	}
	_, err := svc.UpdateUser(ctx, "nonexistent-id", updateReq, "admin")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserManagementService_DeleteUser(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "deleteuser",
		Email:    "deleteuser@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)

	// Delete the user
	err = svc.DeleteUser(ctx, created.ID, "admin", "test-tenant")
	require.NoError(t, err)

	// Verify user is deleted
	user, err := svc.GetUser(ctx, created.ID)
	require.NoError(t, err)
	assert.Nil(t, user)
}

func TestUserManagementService_DeleteUser_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	err := svc.DeleteUser(ctx, "nonexistent-id", "admin", "test-tenant")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestUserManagementService_ListUsers(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create multiple users
	for i := 1; i <= 3; i++ {
		req := &CreateUserRequest{
			Username: "listuser" + string(rune('0'+i)),
			Email:    "listuser" + string(rune('0'+i)) + "@example.com",
			Password: "SecurePassword123!",
			Role:     models.RoleUser,
			TenantID: "test-tenant",
		}
		_, err := svc.CreateUser(ctx, req, "admin")
		require.NoError(t, err)
	}

	// List users
	result, err := svc.ListUsers(ctx, 1, 100)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Users), 3)
}

func TestUserManagementService_UnlockUser(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	userRepo := repositories.NewUserRepository(db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "lockeduser",
		Email:    "locked@example.com",
		Password: "SecurePassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)

	// Lock the user
	err = userRepo.LockAccount(ctx, created.ID, 30*time.Minute)
	require.NoError(t, err)

	// Verify user is locked
	user, _ := svc.GetUser(ctx, created.ID)
	assert.True(t, user.IsLocked())

	// Unlock the user
	err = svc.UnlockUser(ctx, created.ID, "admin", "test-tenant")
	require.NoError(t, err)

	// Verify user is unlocked
	user, _ = svc.GetUser(ctx, created.ID)
	assert.False(t, user.IsLocked())
}

func TestUserManagementService_ResetPassword(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	// Create a user
	req := &CreateUserRequest{
		Username: "resetpwd",
		Email:    "resetpwd@example.com",
		Password: "OldPassword123!",
		Role:     models.RoleUser,
		TenantID: "test-tenant",
	}
	created, err := svc.CreateUser(ctx, req, "admin")
	require.NoError(t, err)
	oldPasswordHash := created.Password

	// Reset password
	err = svc.ResetPassword(ctx, created.ID, "NewPassword456!", "admin", "test-tenant")
	require.NoError(t, err)

	// Verify password changed
	user, _ := svc.GetUser(ctx, created.ID)
	assert.NotEqual(t, oldPasswordHash, user.Password)
}

func TestUserManagementService_ResetPassword_NotFound(t *testing.T) {
	db := setupUserTestDB(t)
	svc := createUserManagementService(t, db)
	ctx := context.Background()

	err := svc.ResetPassword(ctx, "nonexistent-id", "NewPassword456!", "admin", "test-tenant")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestNewUserManagementService(t *testing.T) {
	db := setupUserTestDB(t)
	userRepo := repositories.NewUserRepository(db)
	jwtSvc := utils.NewJWTService("test-secret")
	authSvc := NewAuthService(userRepo, nil, jwtSvc)

	svc := NewUserManagementService(userRepo, nil, authSvc)

	assert.NotNil(t, svc)
	assert.Equal(t, userRepo, svc.userRepo)
	assert.Nil(t, svc.auditRepo)
	assert.Equal(t, authSvc, svc.authSvc)
}

func TestCreateUserRequest_Structure(t *testing.T) {
	req := &CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
		Role:     "admin",
		TenantID: "tenant-123",
	}

	assert.Equal(t, "testuser", req.Username)
	assert.Equal(t, "test@example.com", req.Email)
	assert.Equal(t, "password123", req.Password)
	assert.Equal(t, "admin", req.Role)
	assert.Equal(t, "tenant-123", req.TenantID)
}

func TestUpdateUserRequest_Structure(t *testing.T) {
	req := &UpdateUserRequest{
		Username: "newusername",
		Email:    "new@example.com",
		Role:     "owner",
	}

	assert.Equal(t, "newusername", req.Username)
	assert.Equal(t, "new@example.com", req.Email)
	assert.Equal(t, "owner", req.Role)
}
