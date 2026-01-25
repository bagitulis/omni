// Package products provides product cloning image operations
package products

import (
	"context"
	"fmt"
	"io"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// =============================================================================
// Image Processing Functions
// =============================================================================

func (s *CloneService) processImages(ctx context.Context, targetPlatform string, images []string) ([]string, error) {
	if len(images) == 0 {
		return images, nil
	}

	// For TikTok, skip image processing here - it's handled in createTiktokProduct
	// TikTok requires downloading source image and uploading via multipart/form-data
	// which is done using tiktokPkg.UploadImage in createTiktokProduct
	if targetPlatform == "tiktok" {
		return images, nil
	}

	// For Lazada, skip image processing here - it's handled in createLazadaProduct
	// Lazada uses /image/migrate API which takes URL directly (no need to download/upload)
	// which is done using lazadaPkg.MigrateImage in createLazadaProduct
	if targetPlatform == "lazada" {
		return images, nil
	}

	// For Shopee, skip image processing here - it's handled in createShopeeProduct
	// Shopee uses UploadImageFromURL which downloads and uploads in one step
	// This is handled in uploadShopeeImages within createShopeeProduct
	if targetPlatform == "shopee" {
		return images, nil
	}

	processedURLs := make([]string, 0, len(images))

	for _, imgURL := range images {
		newURL, err := s.uploadImageToPlatform(ctx, targetPlatform, imgURL)
		if err != nil {
			continue
		}
		processedURLs = append(processedURLs, newURL)
	}

	if len(processedURLs) == 0 {
		return nil, fmt.Errorf("failed to process any images")
	}

	return processedURLs, nil
}

func (s *CloneService) uploadImageToPlatform(_ context.Context, platform, imageURL string) (string, error) {
	resp, err := s.httpClient.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	switch platform {
	case "shopee":
		return s.uploadToShopee(imageData)
	case "lazada":
		return s.uploadToLazada(imageData)
	case "tiktok":
		return s.uploadToTiktok(imageData)
	default:
		return imageURL, nil
	}
}

func (s *CloneService) uploadToShopee(imageData []byte) (string, error) {
	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		return "", err
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)

	result, err := client.UploadImage(imageData)
	if err != nil {
		return "", err
	}

	// Return first URL from the list
	if len(result.Response.ImageInfo.ImageURLList) > 0 {
		return result.Response.ImageInfo.ImageURLList[0], nil
	}
	return "", nil
}

func (s *CloneService) uploadToLazada(_ []byte) (string, error) {
	return "", fmt.Errorf("lazada image upload not implemented")
}

func (s *CloneService) uploadToTiktok(_ []byte) (string, error) {
	return "", fmt.Errorf("tiktok image upload not implemented")
}
