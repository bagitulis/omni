// Package lazada provides image API functions for Lazada platform
package lazada

import (
	"fmt"
)

// =============================================================================
// Image Migration Types
// =============================================================================

// ImageMigrateResponse represents the response from POST /image/migrate
type ImageMigrateResponse struct {
	Code    string            `json:"code"`
	Message string            `json:"message,omitempty"`
	Data    *ImageMigrateData `json:"data,omitempty"`
}

// ImageMigrateData contains the migrated image data
type ImageMigrateData struct {
	Image  *ImageInfo  `json:"image,omitempty"`
	Images *ImagesList `json:"images,omitempty"`
}

// ImagesList represents the images array in response
type ImagesList struct {
	Image []ImageInfo `json:"image"`
}

// ImageInfo represents image information
type ImageInfo struct {
	URL string `json:"url"`
}

// =============================================================================
// Image API Methods
// =============================================================================

// MigrateImage migrates an external image URL to Lazada CDN
// API: POST /image/migrate
// Reference: backend-node/src/services/lazadaProductCreateService.ts - migrateImage
// Per Lazada SDK: payload must be in params (query string), not body
func (c *Client) MigrateImage(imageURL string) (string, error) {
	if imageURL == "" {
		return "", fmt.Errorf("image URL is required")
	}

	// Lazada expects XML payload for image migration
	xmlPayload := fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<Request>
  <Image>
    <Url>%s</Url>
  </Image>
</Request>`, imageURL)

	params := map[string]string{
		"payload": xmlPayload,
	}

	var resp ImageMigrateResponse
	err := c.doRequest("POST", "/image/migrate", params, &resp)
	if err != nil {
		// Log but don't fail - return original URL as fallback
		return imageURL, nil
	}

	if resp.Code != "0" && resp.Code != "" {
		// Return original URL as fallback
		return imageURL, nil
	}

	// Try to extract the CDN URL
	if resp.Data != nil {
		if resp.Data.Image != nil && resp.Data.Image.URL != "" {
			return resp.Data.Image.URL, nil
		}
		if resp.Data.Images != nil && len(resp.Data.Images.Image) > 0 {
			return resp.Data.Images.Image[0].URL, nil
		}
	}

	// Fallback to original URL
	return imageURL, nil
}

// MigrateImages migrates multiple images to Lazada CDN
func (c *Client) MigrateImages(imageURLs []string) ([]string, error) {
	result := make([]string, 0, len(imageURLs))

	for _, url := range imageURLs {
		if url == "" {
			continue
		}
		migratedURL, err := c.MigrateImage(url)
		if err != nil {
			// Use original URL on error
			result = append(result, url)
		} else {
			result = append(result, migratedURL)
		}
	}

	return result, nil
}
