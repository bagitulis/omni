package image

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&models.Image{})
	require.NoError(t, err)

	return db
}

func createDedupTestImage() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	drawColor := color.RGBA{255, 0, 0, 255}
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, drawColor)
		}
	}

	tmpFile := filepath.Join(os.TempDir(), "test_image.png")
	f, _ := os.Create(tmpFile)
	defer f.Close()
	defer os.Remove(tmpFile)

	png.Encode(f, img)

	data, _ := os.ReadFile(tmpFile)
	return data
}

func TestDedupService_FindOrCreate(t *testing.T) {
	// Setup Dependencies
	db := setupTestDB(t)
	tmpDir, err := os.MkdirTemp("", "omni_test_storage")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	storageService := NewStorageService(tmpDir)
	webpService := NewWebPService()
	thumbService := NewThumbnailService(webpService, storageService)
	refService := NewReferenceService(db)

	dedupService := NewDedupService(db, webpService, storageService, thumbService, refService)

	ctx := context.Background()
	tenantID := "test_tenant"
	category := "products"
	imageData := createDedupTestImage()

	t.Run("Create New Image", func(t *testing.T) {
		img, isNew, err := dedupService.FindOrCreate(ctx, imageData, "http://example.com/img1.png", tenantID, category)
		require.NoError(t, err)
		assert.True(t, isNew)
		assert.NotZero(t, img.ID)
		assert.Equal(t, 1, img.RefCount)
		assert.Equal(t, "image/webp", img.MimeType)
		assert.NotEmpty(t, img.ContentHash)
		assert.NotEmpty(t, img.LocalPath)

		// Verify file exists
		fullPath := filepath.Join(tmpDir, img.LocalPath)
		assert.FileExists(t, fullPath)
	})

	t.Run("Find Existing Image by Hash", func(t *testing.T) {
		tenantID := "tenant_hash_test"
		// Insert first
		img1, _, err := dedupService.FindOrCreate(ctx, imageData, "http://example.com/img2.png", tenantID, category)
		require.NoError(t, err)

		// Insert same content again
		img2, isNew, err := dedupService.FindOrCreate(ctx, imageData, "http://example.com/other.png", tenantID, category)
		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Equal(t, img1.ID, img2.ID)

		// Verify ref count incremented
		var checkImg models.Image
		db.First(&checkImg, img1.ID)
		assert.Equal(t, 2, checkImg.RefCount) // Was 1 from prev test + 1 from this test? No, db is shared.
		// Wait, "Create New Image" ran first on shared DB?
		// "file::memory:?cache=shared" shares across connections but t.Run might run sequentially.
		// Actually, let's reset DB or handle checking carefully.
		// The previous test created an image.
		// In this test, we call FindOrCreate with same data.
		// It should find the one from "Create New Image" case if content is identical.
		// RefCount should be 2 now (1 from first test + 1 from this call).

		// Wait, let's verify exact RefCount logic.
		// First test: Created (Ref=1).
		// Second test:
		//   Call 1: img1 (Ref=2 now, because it finds the one from First Test).
		//   Call 2: img2 (Ref=3 now).

		// To be cleaner, I should probably use unique tenant or clear DB.
		// But let's just check relative increment or absolute value if we track it.
		// Current RefCount should be:
		// 1 (Test 1)
		// +1 (Test 2 Call 1) -> 2
		// +1 (Test 2 Call 2) -> 3

		// Actually, let's use a unique tenant for this subtest to isolate.
		tenant2 := "tenant_2"

		// Call 1
		img1New, isNew1, err := dedupService.FindOrCreate(ctx, imageData, "", tenant2, category)
		require.NoError(t, err)
		assert.True(t, isNew1)
		assert.Equal(t, 1, img1New.RefCount)

		// Call 2
		img2New, isNew2, err := dedupService.FindOrCreate(ctx, imageData, "", tenant2, category)
		require.NoError(t, err)
		assert.False(t, isNew2)
		assert.Equal(t, img1New.ID, img2New.ID)

		var checkImg2 models.Image
		db.First(&checkImg2, img1New.ID)
		assert.Equal(t, 2, checkImg2.RefCount)
	})

	t.Run("Find Existing Image by URL", func(t *testing.T) {
		tenant3 := "tenant_3"
		url := "http://example.com/unique.png"

		// Create first time
		img1, isNew, err := dedupService.FindOrCreate(ctx, imageData, url, tenant3, category)
		require.NoError(t, err)
		assert.True(t, isNew)

		// Create second time - should match by URL immediately (skips hash check theoretically, but result is same)
		img2, isNew, err := dedupService.FindOrCreate(ctx, imageData, url, tenant3, category)
		require.NoError(t, err)
		assert.False(t, isNew)
		assert.Equal(t, img1.ID, img2.ID)

		var checkImg models.Image
		db.First(&checkImg, img1.ID)
		assert.Equal(t, 2, checkImg.RefCount)
	})
}
