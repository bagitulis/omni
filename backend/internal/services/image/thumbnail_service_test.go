package image

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with some color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestGenerateThumbnails(t *testing.T) {
	// Setup
	tempDir, err := os.MkdirTemp("", "thumbnail_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	webpService := NewWebPService()
	storageService := NewStorageService(tempDir)
	thumbnailService := NewThumbnailService(webpService, storageService)

	// Test Case 1: Landscape Image (1600x800)
	t.Run("Landscape Image", func(t *testing.T) {
		imgData := createTestImage(1600, 800)
		tenantID := "test_tenant"
		category := "products"
		baseFilename := "landscape.png"

		thumbs, err := thumbnailService.GenerateThumbnails(imgData, tenantID, category, baseFilename)
		assert.NoError(t, err)

		// Verify Small (150)
		// Expected: 150w x 75h
		verifyThumbnail(t, storageService, tenantID, thumbs.Small, 150, 75)

		// Verify Medium (400)
		// Expected: 400w x 200h
		verifyThumbnail(t, storageService, tenantID, thumbs.Medium, 400, 200)

		// Verify Large (800)
		// Expected: 800w x 400h
		verifyThumbnail(t, storageService, tenantID, thumbs.Large, 800, 400)
	})

	// Test Case 2: Portrait Image (800x1600)
	t.Run("Portrait Image", func(t *testing.T) {
		imgData := createTestImage(800, 1600)
		tenantID := "test_tenant"
		category := "products"
		baseFilename := "portrait.png"

		thumbs, err := thumbnailService.GenerateThumbnails(imgData, tenantID, category, baseFilename)
		assert.NoError(t, err)

		// Verify Small (150)
		// Expected: 75w x 150h
		verifyThumbnail(t, storageService, tenantID, thumbs.Small, 75, 150)

		// Verify Medium (400)
		// Expected: 200w x 400h
		verifyThumbnail(t, storageService, tenantID, thumbs.Medium, 200, 400)

		// Verify Large (800)
		// Expected: 400w x 800h
		verifyThumbnail(t, storageService, tenantID, thumbs.Large, 400, 800)
	})

	// Test Case 3: Small Image (100x100) - Should not upscale
	t.Run("Small Image", func(t *testing.T) {
		imgData := createTestImage(100, 100)
		tenantID := "test_tenant"
		category := "products"
		baseFilename := "small.png"

		thumbs, err := thumbnailService.GenerateThumbnails(imgData, tenantID, category, baseFilename)
		assert.NoError(t, err)

		// All should be 100x100
		verifyThumbnail(t, storageService, tenantID, thumbs.Small, 100, 100)
		verifyThumbnail(t, storageService, tenantID, thumbs.Medium, 100, 100)
		verifyThumbnail(t, storageService, tenantID, thumbs.Large, 100, 100)
	})
}

func verifyThumbnail(t *testing.T, storage *StorageService, tenantID, path string, expectedW, expectedH int) {
	data, err := storage.GetImage(tenantID, path)
	assert.NoError(t, err, "Failed to get image")

	img, _, err := image.Decode(bytes.NewReader(data))
	assert.NoError(t, err, "Failed to decode generated thumbnail")

	bounds := img.Bounds()
	assert.Equal(t, expectedW, bounds.Dx(), "Width mismatch for %s", path)
	assert.Equal(t, expectedH, bounds.Dy(), "Height mismatch for %s", path)
}
