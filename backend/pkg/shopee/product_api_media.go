package shopee

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// UploadImageRequest represents image upload request
type UploadImageRequest struct {
	Image []byte `json:"-"` // Raw image bytes
}

// UploadImageResponse represents image upload response
type UploadImageResponse struct {
	Error    string `json:"error"`
	Message  string `json:"message"`
	Response struct {
		ImageInfo struct {
			ImageID      string   `json:"image_id"`
			ImageURLList []string `json:"image_url_list,omitempty"`
		} `json:"image_info"`
	} `json:"response"`
}

// UploadImage uploads image to Shopee using multipart/form-data
func (c *Client) UploadImage(imageBytes []byte) (*UploadImageResponse, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("image", "product_image.jpg")
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(imageBytes); err != nil {
		return nil, fmt.Errorf("write image data: %w", err)
	}
	writer.Close()

	reqURL, urlErr := c.buildURL("/api/v2/media_space/upload_image", nil)
	if urlErr != nil {
		return nil, urlErr
	}

	req, err := http.NewRequest("POST", reqURL, body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var rawResp map[string]interface{}
	if err := json.Unmarshal(respBody, &rawResp); err != nil {
		return nil, fmt.Errorf("decode response: %w (body: %s)", err, string(respBody[:min(200, len(respBody))]))
	}

	result := &UploadImageResponse{}
	if errStr, ok := rawResp["error"].(string); ok {
		result.Error = errStr
	}
	if msgStr, ok := rawResp["message"].(string); ok {
		result.Message = msgStr
	}

	if respData, ok := rawResp["response"].(map[string]interface{}); ok {
		if imageInfo, ok := respData["image_info"].(map[string]interface{}); ok {
			if imageID, ok := imageInfo["image_id"].(string); ok {
				result.Response.ImageInfo.ImageID = imageID
			}
		}
	}

	if result.Error != "" {
		return result, fmt.Errorf("shopee API error: %s - %s", result.Error, result.Message)
	}

	return result, nil
}

// UploadImageFromURL downloads image from URL and uploads to Shopee CDN
func (c *Client) UploadImageFromURL(imageURL string) (string, error) {
	resp, err := c.httpClient.Get(imageURL)
	if err != nil {
		return "", fmt.Errorf("download image failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download image failed with status %d", resp.StatusCode)
	}

	imageBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read image bytes failed: %w", err)
	}

	uploadResp, err := c.UploadImage(imageBytes)
	if err != nil {
		return "", err
	}

	return uploadResp.Response.ImageInfo.ImageID, nil
}
