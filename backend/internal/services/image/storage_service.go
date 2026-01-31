// Package image provides image processing and storage services
package image

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ImageInfo represents metadata about a stored image
type ImageInfo struct {
	Filename string    `json:"filename"`
	Path     string    `json:"path"`     // relative path from basePath
	Size     int64     `json:"size"`     // file size in bytes
	ModTime  time.Time `json:"mod_time"` // last modification time
}

// StorageService handles local image file storage
type StorageService struct {
	basePath string
}

// NewStorageService creates a new storage service
// basePath is the root directory for image storage (e.g., "backend/uploads/images")
func NewStorageService(basePath string) *StorageService {
	return &StorageService{basePath: basePath}
}

// SaveImage saves image data to local filesystem
// Returns the relative path from basePath
func (s *StorageService) SaveImage(tenantID, category, originalFilename string, data []byte) (string, error) {
	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}

	// Validate category
	if category != "products" && category != "gallery" {
		category = "gallery" // default to gallery
	}

	// Generate unique filename with timestamp
	ext := filepath.Ext(originalFilename)
	baseName := strings.TrimSuffix(originalFilename, ext)
	// Sanitize filename
	baseName = sanitizeFilename(baseName)
	timestamp := time.Now().UnixNano()
	filename := fmt.Sprintf("%d_%s%s", timestamp, baseName, ext)

	// Build full path
	dirPath := filepath.Join(s.basePath, tenantID, category)
	fullPath := filepath.Join(dirPath, filename)

	// Create directory if not exists
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	// Return relative path from basePath
	relativePath := filepath.Join(tenantID, category, filename)
	return relativePath, nil
}

// GetImage reads image data from local filesystem
// path is relative to basePath
func (s *StorageService) GetImage(tenantID, path string) ([]byte, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// Security check: ensure path starts with tenant ID
	if !strings.HasPrefix(path, tenantID+string(filepath.Separator)) && !strings.HasPrefix(path, tenantID+"/") {
		return nil, fmt.Errorf("access denied: path must belong to tenant")
	}

	fullPath := filepath.Join(s.basePath, path)
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	return data, nil
}

// ListImages returns list of images in a category for a tenant
func (s *StorageService) ListImages(tenantID, category string) ([]ImageInfo, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("tenant_id is required")
	}

	dirPath := filepath.Join(s.basePath, tenantID, category)

	// Check if directory exists
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return []ImageInfo{}, nil // Return empty list, not error
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var images []ImageInfo
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Filter for image files
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".webp" && ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".gif" {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		relativePath := filepath.Join(tenantID, category, entry.Name())
		images = append(images, ImageInfo{
			Filename: entry.Name(),
			Path:     relativePath,
			Size:     info.Size(),
			ModTime:  info.ModTime(),
		})
	}

	return images, nil
}

// DeleteImage removes an image file
// path is relative to basePath
func (s *StorageService) DeleteImage(tenantID, path string) error {
	if tenantID == "" {
		return fmt.Errorf("tenant_id is required")
	}

	// Security check: ensure path starts with tenant ID
	if !strings.HasPrefix(path, tenantID+string(filepath.Separator)) && !strings.HasPrefix(path, tenantID+"/") {
		return fmt.Errorf("access denied: path must belong to tenant")
	}

	fullPath := filepath.Join(s.basePath, path)

	// Check if file exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return fmt.Errorf("file not found: %s", path)
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetBasePath returns the base storage path
func (s *StorageService) GetBasePath() string {
	return s.basePath
}

// sanitizeFilename removes or replaces unsafe characters from filename
func sanitizeFilename(name string) string {
	// Replace unsafe characters with underscore
	replacer := strings.NewReplacer(
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
		" ", "_",
	)
	return replacer.Replace(name)
}
