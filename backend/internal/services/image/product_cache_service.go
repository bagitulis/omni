package image

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"
)

const (
	defaultImageCacheTimeout = 15 * time.Second
	maxImageSizeBytes        = 10 * 1024 * 1024
)

// ProductImageCacheService downloads and stores remote product images locally.
// Returns a local URL under /uploads for frontend consumption.
type ProductImageCacheService struct {
	storage *StorageService
	webp    *WebPService
	client  *http.Client
}

// NewProductImageCacheService creates a new cache service using UPLOAD_PATH.
func NewProductImageCacheService() *ProductImageCacheService {
	basePath := getUploadBasePath()
	return &ProductImageCacheService{
		storage: NewStorageService(basePath),
		webp:    NewWebPService(),
		client: &http.Client{
			Timeout: defaultImageCacheTimeout,
		},
	}
}

// CacheRemoteImage downloads an image URL, converts to WebP when possible,
// stores it in /uploads/<tenant>/products, and returns the public URL.
// If the URL is already local (/uploads/...), it returns it as-is.
func (s *ProductImageCacheService) CacheRemoteImage(
	ctx context.Context,
	tenantID string,
	imageURL string,
	filenamePrefix string,
	allowedHostSuffixes []string,
) (string, error) {
	if imageURL == "" {
		return "", nil
	}
	if strings.HasPrefix(imageURL, "/uploads/") {
		return imageURL, nil
	}

	parsed, err := url.Parse(imageURL)
	if err != nil {
		return "", fmt.Errorf("invalid image URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("unsupported scheme: %s", parsed.Scheme)
	}
	if len(allowedHostSuffixes) > 0 && !hostAllowed(parsed.Host, allowedHostSuffixes) {
		return "", fmt.Errorf("host not allowed: %s", parsed.Host)
	}

	data, contentType, err := s.fetchImage(ctx, imageURL)
	if err != nil {
		return "", err
	}

	data, ext := s.ensureWebP(data, contentType, parsed.Path)
	contentHash := hashBytes(data)
	filename := buildCacheFilename(filenamePrefix, contentHash, ext)

	relativePath, err := s.storage.SaveImageWithName(tenantID, "products", filename, data)
	if err != nil {
		return "", err
	}

	return "/uploads/" + filepath.ToSlash(relativePath), nil
}

func (s *ProductImageCacheService) fetchImage(ctx context.Context, imageURL string) ([]byte, string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", "omni-image-cache/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("fetch image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	if resp.ContentLength > maxImageSizeBytes {
		return nil, "", fmt.Errorf("image too large: %d bytes", resp.ContentLength)
	}

	reader := io.LimitReader(resp.Body, maxImageSizeBytes+1)
	data, err := io.ReadAll(reader)
	if err != nil {
		return nil, "", fmt.Errorf("read image: %w", err)
	}
	if int64(len(data)) > maxImageSizeBytes {
		return nil, "", fmt.Errorf("image too large: %d bytes", len(data))
	}

	return data, resp.Header.Get("Content-Type"), nil
}

func (s *ProductImageCacheService) ensureWebP(data []byte, contentType, urlPath string) ([]byte, string) {
	if IsWebP(data) {
		return data, ".webp"
	}

	converted, _ := s.webp.ConvertToWebP(data)
	if IsWebP(converted) {
		return converted, ".webp"
	}

	return data, extFromContentType(contentType, urlPath)
}

func extFromContentType(contentType, urlPath string) string {
	contentType = strings.ToLower(contentType)
	if strings.Contains(contentType, "image/webp") {
		return ".webp"
	}
	if strings.Contains(contentType, "image/jpeg") {
		return ".jpg"
	}
	if strings.Contains(contentType, "image/png") {
		return ".png"
	}
	if strings.Contains(contentType, "image/gif") {
		return ".gif"
	}

	if ext := strings.ToLower(filepath.Ext(urlPath)); ext != "" {
		return ext
	}
	return ".jpg"
}

func buildCacheFilename(prefix, contentHash, ext string) string {
	if prefix == "" {
		prefix = "asset"
	}
	return fmt.Sprintf("%s_%s%s", sanitizeFilename(prefix), contentHash, ext)
}

func hashBytes(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hostAllowed(host string, allowed []string) bool {
	host = strings.ToLower(host)
	for _, suffix := range allowed {
		suffix = strings.ToLower(strings.TrimSpace(suffix))
		if suffix == "" {
			continue
		}
		if host == suffix || strings.HasSuffix(host, "."+suffix) {
			return true
		}
	}
	return false
}
