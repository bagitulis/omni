package tiktok

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"time"
)

// UploadImageResponse represents TikTok image upload response
type UploadImageResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		URI    string `json:"uri"`
		URL    string `json:"url"`
		Width  int    `json:"width"`
		Height int    `json:"height"`
	} `json:"data"`
}

// UploadImage uploads an image to TikTok CDN
// Steps:
// 1. Download image from source URL
// 2. Upload to TikTok via multipart/form-data
// 3. Return TikTok-hosted URI
//
// useCase: MAIN_IMAGE, ATTRIBUTE_IMAGE, DESCRIPTION_IMAGE, CERTIFICATION_IMAGE, SIZE_CHART_IMAGE
func (c *Client) UploadImage(imageURL, useCase string) (*UploadImageResponse, error) {
	// Step 1: Download image from source URL
	imageData, contentType, err := c.downloadImage(imageURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}

	// Step 2: Upload to TikTok
	return c.uploadImageBytes(imageData, contentType, useCase)
}

// downloadImage downloads image from URL and returns bytes with content type
func (c *Client) downloadImage(imageURL string) ([]byte, string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	req, err := http.NewRequest("GET", imageURL, nil)
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; TikTokImageUploader/1.0)")

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, "", fmt.Errorf("failed to download image: HTTP %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	contentType := resp.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	return data, contentType, nil
}

// uploadImageBytes uploads image bytes to TikTok
func (c *Client) uploadImageBytes(imageData []byte, contentType, useCase string) (*UploadImageResponse, error) {
	apiPath := "/product/202309/images/upload"
	timestamp := time.Now().Unix()

	// Build params for signature
	// NOTE: Image upload does NOT require shop_cipher
	params := map[string]string{
		"app_key":   c.appKey,
		"timestamp": fmt.Sprintf("%d", timestamp),
	}
	if c.accessToken != "" {
		params["access_token"] = c.accessToken
	}
	// shop_cipher is NOT added for image upload endpoint

	// Generate signature (no body for multipart)
	params["sign"] = c.generateSign(apiPath, params)

	// Build URL with params
	u, _ := url.Parse(BaseURL + apiPath)
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Determine file extension
	ext := ".jpg"
	switch contentType {
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	case "image/gif":
		ext = ".gif"
	}
	filename := fmt.Sprintf("image_%d%s", timestamp, ext)

	// Create multipart form
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// Add image file
	part, err := writer.CreateFormFile("data", filepath.Base(filename))
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(imageData); err != nil {
		return nil, err
	}

	// Add use_case field
	if useCase != "" {
		if err := writer.WriteField("use_case", useCase); err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	// Create request
	req, err := http.NewRequest("POST", u.String(), &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	if c.accessToken != "" {
		req.Header.Set("x-tts-access-token", c.accessToken)
	}

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result UploadImageResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w, body: %s", err, string(respBody))
	}

	if result.Code != 0 {
		return &result, fmt.Errorf("tiktok API error: code=%d, message=%s", result.Code, result.Message)
	}

	return &result, nil
}

// UploadImageFromBytes uploads image from bytes directly (if you already have image data)
func (c *Client) UploadImageFromBytes(imageData []byte, contentType, useCase string) (*UploadImageResponse, error) {
	return c.uploadImageBytes(imageData, contentType, useCase)
}
