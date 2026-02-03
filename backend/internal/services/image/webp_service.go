// Package image provides image processing services
package image

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // Register PNG decoder
	"os"
	"strconv"

	"github.com/chai2010/webp"  // WebP encoder/decoder with CGO support
	_ "golang.org/x/image/webp" // Fallback WebP decoder
)

// WebPService handles image format conversion
// Supports converting images.
// NOTE: WebP encoding is enabled using CGO/libwebp.
// ConvertToWebP will convert images to WebP format.
type WebPService struct {
	quality int
}

// NewWebPService creates a new image conversion service
// Quality defaults to 80, can be overridden via WEBP_QUALITY env var
func NewWebPService() *WebPService {
	quality := 80
	if envQuality := os.Getenv("WEBP_QUALITY"); envQuality != "" {
		if q, err := strconv.Atoi(envQuality); err == nil && q > 0 && q <= 100 {
			quality = q
		}
	}
	return &WebPService{quality: quality}
}

// NewWebPServiceWithQuality creates service with specific quality
func NewWebPServiceWithQuality(quality int) *WebPService {
	if quality <= 0 || quality > 100 {
		quality = 80
	}
	return &WebPService{quality: quality}
}

// ConvertToJPEG converts any supported image to JPEG format
// Returns original data if conversion fails (graceful fallback)
func (s *WebPService) ConvertToJPEG(imageData []byte) ([]byte, error) {
	return s.ConvertToJPEGWithQuality(imageData, s.quality)
}

// ConvertToJPEGWithQuality converts image with specific quality setting
func (s *WebPService) ConvertToJPEGWithQuality(imageData []byte, quality int) ([]byte, error) {
	if quality <= 0 || quality > 100 {
		quality = s.quality
	}

	// Decode the image (supports JPEG, PNG, WebP via registered decoders)
	reader := bytes.NewReader(imageData)
	img, format, err := image.Decode(reader)
	if err != nil {
		// Return original on decode failure
		return imageData, nil
	}

	// If already JPEG and within quality tolerance, return original
	if format == "jpeg" {
		return imageData, nil
	}

	// Encode to JPEG
	var buf bytes.Buffer
	opts := &jpeg.Options{Quality: quality}
	if err := jpeg.Encode(&buf, img, opts); err != nil {
		// Return original on encode failure
		return imageData, nil
	}

	return buf.Bytes(), nil
}

// ConvertToWebP converts any supported image to WebP format
func (s *WebPService) ConvertToWebP(imageData []byte) ([]byte, error) {
	return s.ConvertToWebPWithQuality(imageData, float32(s.quality))
}

// ConvertToWebPWithQuality converts image to WebP with specific quality
func (s *WebPService) ConvertToWebPWithQuality(imageData []byte, quality float32) ([]byte, error) {
	// Decode source image
	reader := bytes.NewReader(imageData)
	img, _, err := image.Decode(reader)
	if err != nil {
		// Fallback to original if decode fails
		return imageData, nil
	}

	// Encode to WebP
	var buf bytes.Buffer
	if err := webp.Encode(&buf, img, &webp.Options{Quality: quality}); err != nil {
		// Fallback to original if encode fails
		return imageData, nil
	}

	return buf.Bytes(), nil
}

// DecodeWebP decodes a WebP image to Go image.Image
func DecodeWebP(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	img, _, err := image.Decode(reader)
	return img, err
}

// GetQuality returns the current quality setting
func (s *WebPService) GetQuality() int {
	return s.quality
}

// IsWebP checks if the data is in WebP format
func IsWebP(data []byte) bool {
	if len(data) < 12 {
		return false
	}
	return string(data[0:4]) == "RIFF" && string(data[8:12]) == "WEBP"
}

// IsJPEG checks if the data is in JPEG format
func IsJPEG(data []byte) bool {
	if len(data) < 2 {
		return false
	}
	return data[0] == 0xFF && data[1] == 0xD8
}

// IsPNG checks if the data is in PNG format
func IsPNG(data []byte) bool {
	if len(data) < 8 {
		return false
	}
	return string(data[0:8]) == "\x89PNG\r\n\x1a\n"
}

// GetImageFormat detects image format from data
func GetImageFormat(data []byte) string {
	if IsWebP(data) {
		return "webp"
	}
	if IsJPEG(data) {
		return "jpeg"
	}
	if IsPNG(data) {
		return "png"
	}
	return "unknown"
}

// GetImageDimensions returns width and height of an image
func GetImageDimensions(data []byte) (width, height int, err error) {
	reader := bytes.NewReader(data)
	config, _, err := image.DecodeConfig(reader)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to decode image config: %w", err)
	}
	return config.Width, config.Height, nil
}
