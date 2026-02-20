package image

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"

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
