package image

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/nfnt/resize"
	"github.com/omni/backend/internal/models"

	"gorm.io/gorm"
)

// GetByHash finds image by content hash
func (m *manager) GetByHash(ctx context.Context, tenantID, hash string) (*models.Image, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if hash == "" {
		return nil, fmt.Errorf("hash is required")
	}

	var img models.Image
	err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND content_hash = ?", tenantID, hash).
		First(&img).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query image by hash: %w", err)
	}

	return &img, nil
}

// GetByURL finds image by original URL
func (m *manager) GetByURL(ctx context.Context, tenantID, url string) (*models.Image, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if url == "" {
		return nil, fmt.Errorf("url is required")
	}

	var img models.Image
	err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND original_url = ?", tenantID, url).
		First(&img).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to query image by url: %w", err)
	}

	return &img, nil
}

// downloadImage downloads an image from a remote URL
func (m *manager) downloadImage(ctx context.Context, imageURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("User-Agent", "omni-image-manager/1.0")

	resp, err := m.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// Check content length
	if resp.ContentLength > maxDownloadSize {
		return nil, fmt.Errorf("image too large: %d bytes", resp.ContentLength)
	}

	// Read with size limit
	reader := io.LimitReader(resp.Body, maxDownloadSize+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if int64(len(data)) > maxDownloadSize {
		return nil, fmt.Errorf("image exceeds size limit: %d bytes", len(data))
	}

	return data, nil
}

// createThumbnails creates three sizes of thumbnails and returns original dimensions
func (m *manager) createThumbnails(imageData []byte, basePath string) (width, height int, err error) {
	// Decode original image
	reader := bytes.NewReader(imageData)
	img, _, err := image.Decode(reader)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()
	width = bounds.Dx()
	height = bounds.Dy()

	// Create original (max 1200x1200)
	resizedOriginal := resizeImage(img, originalMaxWidth, originalMaxHeight)
	if err := m.saveWebP(resizedOriginal, filepath.Join(basePath, "original.webp"), originalQuality); err != nil {
		return 0, 0, fmt.Errorf("failed to save original: %w", err)
	}

	// Create medium (max 300x300)
	resizedMedium := resizeImage(img, mediumMaxWidth, mediumMaxHeight)
	if err := m.saveWebP(resizedMedium, filepath.Join(basePath, "medium.webp"), mediumQuality); err != nil {
		return 0, 0, fmt.Errorf("failed to save medium: %w", err)
	}

	// Create thumb (max 80x80)
	resizedThumb := resizeImage(img, thumbMaxWidth, thumbMaxHeight)
	if err := m.saveWebP(resizedThumb, filepath.Join(basePath, "thumb.webp"), thumbQuality); err != nil {
		return 0, 0, fmt.Errorf("failed to save thumb: %w", err)
	}

	return width, height, nil
}

// resizeImage resizes an image to fit within maxWidth x maxHeight while maintaining aspect ratio
func resizeImage(img image.Image, maxWidth, maxHeight int) image.Image {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is smaller than max dimensions, return as-is
	if width <= maxWidth && height <= maxHeight {
		return img
	}

	// Calculate new dimensions maintaining aspect ratio
	var newWidth, newHeight uint
	if width > height {
		newWidth = uint(maxWidth)
		newHeight = 0 // resize will calculate this
	} else {
		newWidth = 0
		newHeight = uint(maxHeight)
	}

	return resize.Resize(newWidth, newHeight, img, resize.Lanczos3)
}

// saveWebP saves an image as WebP format
func (m *manager) saveWebP(img image.Image, filePath string, quality int) error {
	// Encode image to JPEG first (intermediate format)
	tmpData, err := encodeImageToJPEG(img, quality)
	if err != nil {
		return fmt.Errorf("failed to encode image: %w", err)
	}

	// Convert to WebP using the existing service
	webpService := NewWebPServiceWithQuality(quality)
	webpData, err := webpService.ConvertToWebP(tmpData)
	if err != nil {
		return fmt.Errorf("failed to convert to webp: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filePath, webpData, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// encodeImageToJPEG encodes an image.Image to JPEG bytes
func encodeImageToJPEG(img image.Image, quality int) ([]byte, error) {
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		return nil, fmt.Errorf("failed to encode jpeg: %w", err)
	}
	return buf.Bytes(), nil
}

// normalizeURL normalizes a URL (trim spaces, ensure https if possible)
func normalizeURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)

	// Parse URL
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}

	// Upgrade http to https for common CDN domains
	if parsed.Scheme == "http" {
		commonCDNs := []string{"shopee", "lazada", "tiktok", "cloudinary", "amazonaws"}
		for _, cdn := range commonCDNs {
			if strings.Contains(strings.ToLower(parsed.Host), cdn) {
				parsed.Scheme = "https"
				break
			}
		}
	}

	return parsed.String()
}

// calculateHash calculates SHA256 hash of data
func calculateHash(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
