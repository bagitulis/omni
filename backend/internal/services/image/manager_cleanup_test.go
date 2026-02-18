package image

import (
	"context"
	"os"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestManager_CleanupOrphans_DeletesUnlinkedImages(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.Image{},
		&models.MasterProductImage{},
		&models.ShopeeProductImage{},
		&models.TiktokProductImage{},
		&models.LazadaProductImage{},
	)
	require.NoError(t, err)

	tmpDir, err := os.MkdirTemp("", "image-manager-cleanup")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	mgr := NewManager(db, tmpDir)
	ctx := context.Background()
	tenantID := "tenant_cleanup_test"

	referenced := &models.Image{TenantID: tenantID, Filename: "referenced.webp", LocalPath: "tenant/images/referenced", RefCount: 2}
	unlinkedPositiveRef := &models.Image{TenantID: tenantID, Filename: "unlinked-positive.webp", LocalPath: "tenant/images/unlinked-positive", RefCount: 3}
	zeroRef := &models.Image{TenantID: tenantID, Filename: "zero-ref.webp", LocalPath: "tenant/images/zero-ref", RefCount: 0}

	require.NoError(t, db.Create(referenced).Error)
	require.NoError(t, db.Create(unlinkedPositiveRef).Error)
	require.NoError(t, db.Create(zeroRef).Error)
	require.NoError(t, db.Create(&models.MasterProductImage{ProductID: 100, ImageID: referenced.ID}).Error)

	deletedCount, err := mgr.CleanupOrphans(ctx, tenantID)
	require.NoError(t, err)
	assert.Equal(t, 2, deletedCount)

	var count int64
	require.NoError(t, db.Model(&models.Image{}).Where("id = ?", referenced.ID).Count(&count).Error)
	assert.Equal(t, int64(1), count)

	require.NoError(t, db.Model(&models.Image{}).Where("id = ?", unlinkedPositiveRef.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count)

	require.NoError(t, db.Model(&models.Image{}).Where("id = ?", zeroRef.ID).Count(&count).Error)
	assert.Equal(t, int64(0), count)
}
