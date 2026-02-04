package image_test

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services/image"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.AutoMigrate(&models.Image{}); err != nil {
		t.Fatalf("Failed to migrate database: %v", err)
	}

	return db
}

func TestIncrementRef(t *testing.T) {
	db := setupTestDB(t)
	service := image.NewReferenceService(db)
	ctx := context.Background()

	img := models.Image{
		Filename:  "test.jpg",
		RefCount:  0,
		LocalPath: "path/to/file",
		TenantID:  "tenant1",
	}
	db.Create(&img)

	err := service.IncrementRef(ctx, img.ID)
	assert.NoError(t, err)

	var updatedImg models.Image
	db.First(&updatedImg, img.ID)
	assert.Equal(t, 1, updatedImg.RefCount)
}

func TestDecrementRef(t *testing.T) {
	db := setupTestDB(t)
	service := image.NewReferenceService(db)
	ctx := context.Background()

	img := models.Image{
		Filename:  "test.jpg",
		RefCount:  2,
		LocalPath: "path/to/file",
		TenantID:  "tenant1",
	}
	db.Create(&img)

	// Decrement from 2 to 1
	err := service.DecrementRef(ctx, img.ID)
	assert.NoError(t, err)

	var updatedImg models.Image
	db.First(&updatedImg, img.ID)
	assert.Equal(t, 1, updatedImg.RefCount)
	assert.Nil(t, updatedImg.DeletedAt)

	// Decrement from 1 to 0 (should soft delete)
	err = service.DecrementRef(ctx, img.ID)
	assert.NoError(t, err)

	db.First(&updatedImg, img.ID)
	assert.Equal(t, 0, updatedImg.RefCount)
	assert.NotNil(t, updatedImg.DeletedAt)
}

func TestBulkDecrement(t *testing.T) {
	db := setupTestDB(t)
	service := image.NewReferenceService(db)
	ctx := context.Background()

	img1 := models.Image{Filename: "1.jpg", RefCount: 1, LocalPath: "p1", TenantID: "t1"}
	img2 := models.Image{Filename: "2.jpg", RefCount: 2, LocalPath: "p2", TenantID: "t1"}
	db.Create(&img1)
	db.Create(&img2)

	err := service.BulkDecrement(ctx, []uint{img1.ID, img2.ID})
	assert.NoError(t, err)

	var u1, u2 models.Image
	db.First(&u1, img1.ID)
	db.First(&u2, img2.ID)

	assert.Equal(t, 0, u1.RefCount)
	assert.NotNil(t, u1.DeletedAt)

	assert.Equal(t, 1, u2.RefCount)
	assert.Nil(t, u2.DeletedAt)
}

func TestGetRefCount(t *testing.T) {
	db := setupTestDB(t)
	service := image.NewReferenceService(db)
	ctx := context.Background()

	img := models.Image{Filename: "test.jpg", RefCount: 5, LocalPath: "p", TenantID: "t"}
	db.Create(&img)

	count, err := service.GetRefCount(ctx, img.ID)
	assert.NoError(t, err)
	assert.Equal(t, 5, count)
}
