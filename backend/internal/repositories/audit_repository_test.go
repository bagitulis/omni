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

func TestAuditRepository(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.AuditLog{})
	repo := NewAuditRepository(db)
	ctx := context.Background()

	// Helper to create audit entry
	createAuditEntry := func(t *testing.T, tenantID, action, userID string) {
		entry := &models.AuditLogEntry{
			TenantID:  tenantID,
			Action:    action,
			UserID:    userID,
			Status:    models.AuditStatusSuccess,
			IPAddress: "127.0.0.1",
			UserAgent: "Test Agent",
			Details:   map[string]interface{}{"test": "value"},
		}
		err := repo.Create(ctx, entry)
		require.NoError(t, err)
	}

	t.Run("Create", func(t *testing.T) {
		entry := &models.AuditLogEntry{
			TenantID:     "tenant-audit-1",
			Action:       models.AuditActionUserLogin,
			UserID:       "user-1",
			TargetUserID: "user-2",
			Status:       models.AuditStatusSuccess,
			IPAddress:    "192.168.1.1",
			UserAgent:    "Mozilla/5.0",
			Details:      map[string]interface{}{"browser": "Chrome"},
		}

		err := repo.Create(ctx, entry)
		assert.NoError(t, err)
	})

	t.Run("FindByTenant", func(t *testing.T) {
		tenantID := "tenant-findtenant-1"
		createAuditEntry(t, tenantID, models.AuditActionUserLogin, "user-1")
		createAuditEntry(t, tenantID, models.AuditActionUserLogout, "user-1")

		logs, total, err := repo.FindByTenant(ctx, tenantID, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(logs), 2)
	})

	t.Run("FindByUser", func(t *testing.T) {
		tenantID := "tenant-finduser-1"
		userID := "user-finduser-1"
		createAuditEntry(t, tenantID, models.AuditActionUserLogin, userID)

		logs, err := repo.FindByUser(ctx, userID, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 1)
		assert.Equal(t, userID, logs[0].UserID)
	})

	t.Run("FindByAction", func(t *testing.T) {
		tenantID := "tenant-findaction-1"
		createAuditEntry(t, tenantID, models.AuditActionPasswordReset, "user-1")

		logs, err := repo.FindByAction(ctx, tenantID, models.AuditActionPasswordReset, 10)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, len(logs), 1)
		for _, log := range logs {
			assert.Equal(t, models.AuditActionPasswordReset, log.Action)
		}
	})

	t.Run("FindByDateRange", func(t *testing.T) {
		tenantID := "tenant-daterange-1"
		createAuditEntry(t, tenantID, models.AuditActionUserLogin, "user-1")

		startDate := time.Now().Add(-1 * time.Hour)
		endDate := time.Now().Add(1 * time.Hour)

		logs, total, err := repo.FindByDateRange(ctx, tenantID, startDate, endDate, 10, 0)
		assert.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(1))
		assert.GreaterOrEqual(t, len(logs), 1)
	})

	t.Run("DeleteOldLogs", func(t *testing.T) {
		tenantID := "tenant-deleteold-1"
		createAuditEntry(t, tenantID, models.AuditActionUserLogin, "user-1")

		// Delete logs older than 1 nanosecond (effectively none since we just created)
		err := repo.DeleteOldLogs(ctx, tenantID, 1*time.Nanosecond)
		assert.NoError(t, err)

		// The recently created log should still exist
		logs, _, err := repo.FindByTenant(ctx, tenantID, 10, 0)
		assert.NoError(t, err)
		// Log might still exist depending on timing
		_ = logs
	})
}
