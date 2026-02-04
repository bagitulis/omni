package image

import (
	"bytes"
	"fmt"
	"image"
	_ "image/gif"  // Support GIF decoding
	_ "image/jpeg" // Support JPEG decoding
	"image/png"
	"path/filepath"
	"strings"

	"golang.org/x/image/draw"
)

// ThumbnailSizes defines the 3 standard thumbnail sizes
var ThumbnailSizes = map[string]int{
	"small":  150,
	"medium": 400,
	"large":  800,
}

// Thumbnails holds paths to generated thumbnail files
type Thumbnails struct {
	Small  string `json:"small"`
	Medium string `json:"medium"`
	Large  string `json:"large"`
}

// ThumbnailService handles thumbnail generation
type ThumbnailService struct {
	webpService    *WebPService
	storageService *StorageService
}

// NewThumbnailService creates a new thumbnail service
func NewThumbnailService(webp *WebPService, storage *StorageService) *ThumbnailService {
	return &ThumbnailService{
		webpService:    webp,
		storageService: storage,
	}
}

// GenerateThumbnails creates 3 thumbnails from image data
// Returns Thumbnails struct with paths to saved files
// Maintains aspect ratio (fit within bounds)
func (s *ThumbnailService) GenerateThumbnails(
	imageData []byte,
	tenantID string,
	category string,
	baseFilename string,
) (Thumbnails, error) {
	if tenantID == "" {
		return Thumbnails{}, fmt.Errorf("tenant_id is required")
	}

	img, _, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return Thumbnails{}, fmt.Errorf("failed to decode image: %w", err)
	}

	// Remove extension from baseFilename if present
	ext := filepath.Ext(baseFilename)
	nameWithoutExt := strings.TrimSuffix(baseFilename, ext)

	result := Thumbnails{}

	// Helper to process one size
	processSize := func(sizeName string, maxDim int) (string, error) {
		resized := s.resizeImage(img, maxDim)

		// Encode to PNG (intermediate)
		var buf bytes.Buffer
		if err := png.Encode(&buf, resized); err != nil {
			return "", fmt.Errorf("failed to encode intermediate png: %w", err)
		}

		// Convert to WebP
		finalData, err := s.webpService.ConvertToWebP(buf.Bytes())
		if err != nil {
			// Fallback to intermediate PNG if WebP conversion fails
			finalData = buf.Bytes()
		}

		// Determine extension
		finalExt := ".png"
		if IsWebP(finalData) {
			finalExt = ".webp"
		}

		filename := fmt.Sprintf("%s_%s%s", nameWithoutExt, sizeName, finalExt)

		path, err := s.storageService.SaveImageWithName(tenantID, category, filename, finalData)
		if err != nil {
			return "", err
		}
		return path, nil
	}

	// Generate Small
	path, err := processSize("small", ThumbnailSizes["small"])
	if err != nil {
		return Thumbnails{}, fmt.Errorf("failed to generate small thumbnail: %w", err)
	}
	result.Small = path

	// Generate Medium
	path, err = processSize("medium", ThumbnailSizes["medium"])
	if err != nil {
		return Thumbnails{}, fmt.Errorf("failed to generate medium thumbnail: %w", err)
	}
	result.Medium = path

	// Generate Large
	path, err = processSize("large", ThumbnailSizes["large"])
	if err != nil {
		return Thumbnails{}, fmt.Errorf("failed to generate large thumbnail: %w", err)
	}
	result.Large = path

	return result, nil
}

// resizeImage resizes image to fit within maxDimension while preserving aspect ratio
func (s *ThumbnailService) resizeImage(img image.Image, maxDimension int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is already smaller than maxDimension, return original
	if width <= maxDimension && height <= maxDimension {
		return img
	}

	var newWidth, newHeight int
	aspectRatio := float64(width) / float64(height)

	if width > height {
		newWidth = maxDimension
		newHeight = int(float64(maxDimension) / aspectRatio)
	} else {
		newHeight = maxDimension
		newWidth = int(float64(maxDimension) * aspectRatio)
	}

	// Ensure at least 1 pixel
	if newWidth < 1 {
		newWidth = 1
	}
	if newHeight < 1 {
		newHeight = 1
	}

	dst := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))

	// Use CatmullRom for high quality resizing
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)

	return dst
}
