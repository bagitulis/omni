package services

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNotificationService_PushWithMetadata(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.Notification{})
	repo := repositories.NewNotificationRepository(db)
	svc := NewNotificationService(repo)

	// Create BulkOperationMetadata
	metadata := &models.BulkOperationMetadata{
		OperationType: "stock_sync",
		Total:         10,
		Succeeded:     8,
		Failed:        2,
		Platforms: map[string]models.PlatformSyncStats{
			"shopee": {Succeeded: 5, Failed: 1},
			"lazada": {Succeeded: 3, Failed: 1},
		},
		FailedItems: []models.FailedItemDetail{
			{SKU: "SKU001", Platform: "shopee", Error: "out of stock"},
		},
	}
	metadataJSON, err := json.Marshal(metadata)
	require.NoError(t, err)

	// Push with metadata
	notif, err := svc.Push("success", "inventory", "Bulk Stock Sync", "Test message", "", string(metadataJSON))
	require.NoError(t, err)
	require.NotNil(t, notif)
	assert.NotZero(t, notif.ID)

	// Fetch from DB and verify metadata persisted
	fetched, err := svc.GetByID(notif.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)

	var fetchedMetadata models.BulkOperationMetadata
	err = json.Unmarshal([]byte(fetched.Metadata), &fetchedMetadata)
	require.NoError(t, err)
	assert.Equal(t, "stock_sync", fetchedMetadata.OperationType)
	assert.Equal(t, 10, fetchedMetadata.Total)
	assert.Equal(t, 8, fetchedMetadata.Succeeded)
	assert.Equal(t, 2, fetchedMetadata.Failed)
	assert.Len(t, fetchedMetadata.FailedItems, 1)
	assert.Equal(t, "SKU001", fetchedMetadata.FailedItems[0].SKU)
}

func TestNotificationService_PushWithoutMetadata(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.Notification{})
	repo := repositories.NewNotificationRepository(db)
	svc := NewNotificationService(repo)

	// Push without metadata (legacy path — variadic handles empty)
	notif, err := svc.Push("info", "system", "Test Title", "Test message", "")
	require.NoError(t, err)
	require.NotNil(t, notif)
	assert.Empty(t, notif.Metadata)

	// Fetch from DB and verify metadata field is indeed empty
	fetched, err := svc.GetByID(notif.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Empty(t, fetched.Metadata)
}

func TestNotificationService_PushNullMetadataSafety(t *testing.T) {
	db := testutils.SetupTestPostgresWithModels(t, &models.Notification{})
	repo := repositories.NewNotificationRepository(db)
	svc := NewNotificationService(repo)

	// Push with empty metadata string — should not panic and metadata should be empty
	notif, err := svc.Push("warning", "inventory", "Test", "Test message", "", "")
	require.NoError(t, err)
	require.NotNil(t, notif)
	assert.Empty(t, notif.Metadata)
}
