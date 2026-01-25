package products

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
)

// ImageUploader handles image upload to platforms
type ImageUploader struct {
	httpClient *http.Client
	maxSize    int64 // Max size in bytes
	maxWidth   int
	maxHeight  int
}

// NewImageUploader creates a new image uploader
func NewImageUploader() *ImageUploader {
	return &ImageUploader{
		httpClient: &http.Client{},
		maxSize:    2 * 1024 * 1024, // 2MB default
		maxWidth:   2000,
		maxHeight:  2000,
	}
}

// ImageUploadResult represents the result of image upload
type ImageUploadResult struct {
	Success  bool   `json:"success"`
	ImageID  string `json:"image_id,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
	Error    string `json:"error,omitempty"`
}

// DownloadImage downloads an image from URL
func (u *ImageUploader) DownloadImage(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := u.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	// Limit read to maxSize
	data, err := io.ReadAll(io.LimitReader(resp.Body, u.maxSize+1))
	if err != nil {
		return nil, fmt.Errorf("read image: %w", err)
	}

	if int64(len(data)) > u.maxSize {
		return nil, fmt.Errorf("image too large: %d bytes (max %d)", len(data), u.maxSize)
	}

	return data, nil
}

// ValidateImage checks if image meets platform requirements
func (u *ImageUploader) ValidateImage(data []byte) error {
	// Check size
	if int64(len(data)) > u.maxSize {
		return fmt.Errorf("image too large: %d bytes (max %d)", len(data), u.maxSize)
	}

	// Decode to check dimensions
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("invalid image format: %w", err)
	}

	if cfg.Width > u.maxWidth || cfg.Height > u.maxHeight {
		return fmt.Errorf("image dimensions %dx%d exceed max %dx%d", cfg.Width, cfg.Height, u.maxWidth, u.maxHeight)
	}

	return nil
}

// UploadToShopee uploads image to Shopee
func (u *ImageUploader) UploadToShopee(ctx context.Context, api PlatformAPI, imageURL string) (*ImageUploadResult, error) {
	imageID, err := api.UploadImage(ctx, imageURL)
	if err != nil {
		return &ImageUploadResult{Success: false, Error: err.Error()}, nil
	}
	return &ImageUploadResult{Success: true, ImageID: imageID}, nil
}

// UploadToLazada uploads image to Lazada
func (u *ImageUploader) UploadToLazada(ctx context.Context, api PlatformAPI, imageURL string) (*ImageUploadResult, error) {
	resultURL, err := api.UploadImage(ctx, imageURL)
	if err != nil {
		return &ImageUploadResult{Success: false, Error: err.Error()}, nil
	}
	return &ImageUploadResult{Success: true, ImageURL: resultURL}, nil
}

// UploadToTiktok uploads image to TikTok Shop
func (u *ImageUploader) UploadToTiktok(ctx context.Context, api PlatformAPI, imageURL string) (*ImageUploadResult, error) {
	imageID, err := api.UploadImage(ctx, imageURL)
	if err != nil {
		return &ImageUploadResult{Success: false, Error: err.Error()}, nil
	}
	return &ImageUploadResult{Success: true, ImageID: imageID}, nil
}

// BatchUpload uploads multiple images
func (u *ImageUploader) BatchUpload(ctx context.Context, api PlatformAPI, platform string, imageURLs []string) []ImageUploadResult {
	results := make([]ImageUploadResult, len(imageURLs))

	for i, url := range imageURLs {
		var result *ImageUploadResult
		var err error

		switch platform {
		case "shopee":
			result, err = u.UploadToShopee(ctx, api, url)
		case "lazada":
			result, err = u.UploadToLazada(ctx, api, url)
		case "tiktok":
			result, err = u.UploadToTiktok(ctx, api, url)
		default:
			results[i] = ImageUploadResult{Success: false, Error: "unsupported platform"}
			continue
		}

		if err != nil {
			results[i] = ImageUploadResult{Success: false, Error: err.Error()}
		} else {
			results[i] = *result
		}
	}

	return results
}
