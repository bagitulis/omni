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
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"

	"github.com/nfnt/resize"
	"gorm.io/gorm"
)

const (
	maxDownloadSize   = 10 * 1024 * 1024 // 10MB
	downloadTimeout   = 30 * time.Second
	thumbMaxWidth     = 80
	thumbMaxHeight    = 80
	thumbQuality      = 75
	mediumMaxWidth    = 300
	mediumMaxHeight   = 300
	mediumQuality     = 80
	originalMaxWidth  = 1200
	originalMaxHeight = 1200
	originalQuality   = 85
)

// ImagePaths represents the three thumbnail paths for an image
type ImagePaths struct {
	Thumb    string `json:"thumb"`
	Medium   string `json:"medium"`
	Original string `json:"original"`
}

// Manager handles unified image management with content-based deduplication
type Manager interface {
	// CacheImage downloads remote URL, dedupes by content hash, creates thumbnails
	// Returns the Image record (new or existing)
	CacheImage(ctx context.Context, tenantID, remoteURL string) (*models.Image, error)

	// GetPaths returns the 3 thumbnail paths for an image
	GetPaths(img *models.Image) ImagePaths

	// AddRef increments reference count
	AddRef(ctx context.Context, imageID uint) error

	// RemoveRef decrements reference count (does not delete, just marks for cleanup)
	RemoveRef(ctx context.Context, imageID uint) error

	// CleanupOrphans deletes images with ref_count <= 0
	CleanupOrphans(ctx context.Context, tenantID string) (int, error)

	// GetByHash finds image by content hash (for dedup check)
	GetByHash(ctx context.Context, tenantID, hash string) (*models.Image, error)

	// GetByURL finds image by original URL (for fast lookup)
	GetByURL(ctx context.Context, tenantID, url string) (*models.Image, error)
}

type manager struct {
	db         *gorm.DB
	uploadPath string
	webp       *WebPService
	client     *http.Client
	mu         sync.Mutex // Protects concurrent cache operations
}

// NewManager creates a new image manager
func NewManager(db *gorm.DB, uploadPath string) Manager {
	if uploadPath == "" {
		uploadPath = getUploadBasePath()
	}

	return &manager{
		db:         db,
		uploadPath: uploadPath,
		webp:       NewWebPService(),
		client: &http.Client{
			Timeout: downloadTimeout,
		},
	}
}

// CacheImage downloads remote URL, dedupes by content hash, creates thumbnails
func (m *manager) CacheImage(ctx context.Context, tenantID, remoteURL string) (*models.Image, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}
	if remoteURL == "" {
		return nil, fmt.Errorf("remote_url is required")
	}

	// Step 1: Normalize URL
	normalizedURL := normalizeURL(remoteURL)

	// Step 2: Check DB by original_url
	existing, err := m.GetByURL(ctx, tenantID, normalizedURL)
	if err == nil && existing != nil {
		// Found existing image, increment ref count
		if err := m.AddRef(ctx, existing.ID); err != nil {
			return nil, fmt.Errorf("failed to increment ref count: %w", err)
		}
		return existing, nil
	}

	// Step 3: Download image
	imageData, err := m.downloadImage(ctx, normalizedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Step 4: Calculate content hash
	contentHash := calculateHash(imageData)

	// Step 5: Check DB by content_hash (use lock to prevent race condition)
	m.mu.Lock()
	defer m.mu.Unlock()

	existing, err = m.GetByHash(ctx, tenantID, contentHash)
	if err == nil && existing != nil {
		// Found duplicate, update original_url if different and increment ref count
		if existing.OriginalURL != normalizedURL {
			if err := m.db.WithContext(ctx).Model(existing).Update("original_url", normalizedURL).Error; err != nil {
				return nil, fmt.Errorf("failed to update original_url: %w", err)
			}
		}
		if err := m.AddRef(ctx, existing.ID); err != nil {
			return nil, fmt.Errorf("failed to increment ref count: %w", err)
		}
		return existing, nil
	}

	// Step 6: Create directory structure
	basePath := filepath.Join(m.uploadPath, tenantID, "images", contentHash)
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Step 7: Resize and save thumbnails
	width, height, err := m.createThumbnails(imageData, basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create thumbnails: %w", err)
	}

	// Step 8: Insert DB record with ref_count=1
	localPath := filepath.Join(tenantID, "images", contentHash)
	img := &models.Image{
		TenantID:    tenantID,
		Filename:    contentHash + ".webp", // Use content hash as filename for deduplication
		ContentHash: contentHash,
		OriginalURL: normalizedURL,
		LocalPath:   filepath.ToSlash(localPath),
		MimeType:    "image/webp",
		Category:    "gallery",
		Width:       width,
		Height:      height,
		FileSize:    int64(len(imageData)),
		RefCount:    1,
	}

	if err := m.db.WithContext(ctx).Create(img).Error; err != nil {
		return nil, fmt.Errorf("failed to create image record: %w", err)
	}

	return img, nil
}

// GetPaths returns the 3 thumbnail paths for an image
func (m *manager) GetPaths(img *models.Image) ImagePaths {
	if img == nil {
		return ImagePaths{}
	}

	basePath := "/" + img.LocalPath
	return ImagePaths{
		Thumb:    basePath + "/thumb.webp",
		Medium:   basePath + "/medium.webp",
		Original: basePath + "/original.webp",
	}
}

// AddRef increments reference count
func (m *manager) AddRef(ctx context.Context, imageID uint) error {
	if imageID == 0 {
		return fmt.Errorf("invalid image_id")
	}

	result := m.db.WithContext(ctx).Model(&models.Image{}).
		Where("id = ?", imageID).
		UpdateColumn("ref_count", gorm.Expr("ref_count + 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to increment ref_count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("image not found: %d", imageID)
	}

	return nil
}

// RemoveRef decrements reference count
func (m *manager) RemoveRef(ctx context.Context, imageID uint) error {
	if imageID == 0 {
		return fmt.Errorf("invalid image_id")
	}

	result := m.db.WithContext(ctx).Model(&models.Image{}).
		Where("id = ?", imageID).
		UpdateColumn("ref_count", gorm.Expr("ref_count - 1"))

	if result.Error != nil {
		return fmt.Errorf("failed to decrement ref_count: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("image not found: %d", imageID)
	}

	return nil
}

// CleanupOrphans deletes images with ref_count <= 0
func (m *manager) CleanupOrphans(ctx context.Context, tenantID string) (int, error) {
	if tenantID == "" {
		return 0, fmt.Errorf("tenant_id is required")
	}

	var orphans []models.Image
	if err := m.db.WithContext(ctx).
		Where("tenant_id = ? AND ref_count <= 0", tenantID).
		Find(&orphans).Error; err != nil {
		return 0, fmt.Errorf("failed to find orphan images: %w", err)
	}

	if len(orphans) == 0 {
		return 0, nil
	}

	// Delete files and DB records in transaction
	err := m.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, img := range orphans {
			// Delete physical files
			basePath := filepath.Join(m.uploadPath, img.LocalPath)
			if err := os.RemoveAll(basePath); err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("failed to delete files for image %d: %w", img.ID, err)
			}

			// Delete DB record
			if err := tx.Delete(&img).Error; err != nil {
				return fmt.Errorf("failed to delete image record %d: %w", img.ID, err)
			}
		}
		return nil
	})

	if err != nil {
		return 0, err
	}

	return len(orphans), nil
}

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
