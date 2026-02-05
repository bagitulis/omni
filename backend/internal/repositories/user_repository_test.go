package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.User{})
	repo := NewUserRepository(db)
	ctx := context.Background()

	// Helper to create test user
	createTestUser := func(t *testing.T, username, email string) *models.User {
		user := &models.User{
			Username: username,
			Email:    email,
			Password: "hashed_password",
			Role:     "owner",
		}
		err := repo.Create(ctx, user)
		require.NoError(t, err)
		return user
	}

	t.Run("Create", func(t *testing.T) {
		tests := []struct {
			name    string
			user    *models.User
			wantErr bool
		}{
			{
				name: "creates user with auto-generated ID",
				user: &models.User{
					Username: "create_test1",
					Email:    "create_test1@example.com",
					Password: "password123",
					Role:     "owner",
				},
				wantErr: false,
			},
			{
				name: "creates user with custom ID",
				user: &models.User{
					ID:       "custom-id-123",
					Username: "create_test2",
					Email:    "create_test2@example.com",
					Password: "password123",
					Role:     "admin",
				},
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				err := repo.Create(ctx, tt.user)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.user.ID)
				assert.False(t, tt.user.CreatedAt.IsZero())
				assert.False(t, tt.user.UpdatedAt.IsZero())
			})
		}
	})

	t.Run("FindByID", func(t *testing.T) {
		user := createTestUser(t, "findbyid_user", "findbyid@example.com")

		tests := []struct {
			name    string
			id      string
			wantNil bool
			wantErr bool
		}{
			{
				name:    "finds existing user",
				id:      user.ID,
				wantNil: false,
				wantErr: false,
			},
			{
				name:    "returns nil for non-existent user",
				id:      "non-existent-id",
				wantNil: true,
				wantErr: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				found, err := repo.FindByID(ctx, tt.id)
				if tt.wantErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, found)
				} else {
					require.NotNil(t, found)
					assert.Equal(t, user.ID, found.ID)
					assert.Equal(t, user.Username, found.Username)
				}
			})
		}
	})

	t.Run("FindByEmail", func(t *testing.T) {
		user := createTestUser(t, "findbyemail_user", "findbyemail@example.com")

		found, err := repo.FindByEmail(ctx, user.Email)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)

		notFound, err := repo.FindByEmail(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("FindByUsername", func(t *testing.T) {
		user := createTestUser(t, "findbyusername_user", "findbyusername@example.com")

		found, err := repo.FindByUsername(ctx, user.Username)
		assert.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, user.ID, found.ID)

		notFound, err := repo.FindByUsername(ctx, "nonexistent_user")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("FindByUsernameOrEmail", func(t *testing.T) {
		user := createTestUser(t, "find_either_user", "find_either@example.com")

		// Find by username
		byUsername, err := repo.FindByUsernameOrEmail(ctx, user.Username)
		assert.NoError(t, err)
		require.NotNil(t, byUsername)
		assert.Equal(t, user.ID, byUsername.ID)

		// Find by email
		byEmail, err := repo.FindByUsernameOrEmail(ctx, user.Email)
		assert.NoError(t, err)
		require.NotNil(t, byEmail)
		assert.Equal(t, user.ID, byEmail.ID)

		// Not found
		notFound, err := repo.FindByUsernameOrEmail(ctx, "nonexistent")
		assert.NoError(t, err)
		assert.Nil(t, notFound)
	})

	t.Run("FindAll", func(t *testing.T) {
		// Create users
		createTestUser(t, "findall_user1", "findall1@example.com")
		createTestUser(t, "findall_user2", "findall2@example.com")

		users, err := repo.FindAll(ctx)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(users), 2)
	})

	t.Run("Update", func(t *testing.T) {
		user := createTestUser(t, "update_user", "update@example.com")
		originalUpdatedAt := user.UpdatedAt

		// Wait a bit to ensure timestamp changes
		time.Sleep(10 * time.Millisecond)

		user.Username = "updated_username"
		err := repo.Update(ctx, user)
		assert.NoError(t, err)
		assert.True(t, user.UpdatedAt.After(originalUpdatedAt))

		// Verify update persisted
		found, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, "updated_username", found.Username)
	})

	t.Run("Delete", func(t *testing.T) {
		user := createTestUser(t, "delete_user", "delete@example.com")

		err := repo.Delete(ctx, user.ID)
		assert.NoError(t, err)

		// Verify deleted
		found, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Nil(t, found)
	})

	t.Run("IncrementFailedAttempts", func(t *testing.T) {
		user := createTestUser(t, "failattempt_user", "failattempt@example.com")

		err := repo.IncrementFailedAttempts(ctx, user.ID)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, found.FailedLoginAttempts)
		assert.NotNil(t, found.LastFailedLogin)

		// Increment again
		err = repo.IncrementFailedAttempts(ctx, user.ID)
		assert.NoError(t, err)

		found, err = repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 2, found.FailedLoginAttempts)
	})

	t.Run("ResetFailedAttempts", func(t *testing.T) {
		user := createTestUser(t, "resetfail_user", "resetfail@example.com")

		// First increment
		_ = repo.IncrementFailedAttempts(ctx, user.ID)
		_ = repo.IncrementFailedAttempts(ctx, user.ID)

		// Reset
		err := repo.ResetFailedAttempts(ctx, user.ID)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.Equal(t, 0, found.FailedLoginAttempts)
	})

	t.Run("LockAccount", func(t *testing.T) {
		user := createTestUser(t, "lock_user", "lock@example.com")

		err := repo.LockAccount(ctx, user.ID, 30*time.Minute)
		assert.NoError(t, err)

		found, err := repo.FindByID(ctx, user.ID)
		assert.NoError(t, err)
		assert.NotNil(t, found.AccountLockedUntil)
		assert.True(t, found.AccountLockedUntil.After(time.Now()))
		assert.True(t, found.IsLocked())
	})

	t.Run("ExistsByUsername", func(t *testing.T) {
		user := createTestUser(t, "exists_user", "exists@example.com")

		exists, err := repo.ExistsByUsername(ctx, user.Username)
		assert.NoError(t, err)
		assert.True(t, exists)

		notExists, err := repo.ExistsByUsername(ctx, "nonexistent_username")
		assert.NoError(t, err)
		assert.False(t, notExists)
	})

	t.Run("ExistsByEmail", func(t *testing.T) {
		user := createTestUser(t, "existsemail_user", "existsemail@example.com")

		exists, err := repo.ExistsByEmail(ctx, user.Email)
		assert.NoError(t, err)
		assert.True(t, exists)

		notExists, err := repo.ExistsByEmail(ctx, "nonexistent@example.com")
		assert.NoError(t, err)
		assert.False(t, notExists)
	})
}
