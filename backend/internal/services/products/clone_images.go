// Package products provides product cloning image operations
package products

import "context"

// =============================================================================
// Image Processing Functions
// =============================================================================

// processImages handles image preparation for the target platform during product cloning.
// Each platform handles its own image upload strategy in the respective create*Product method:
//   - Shopee: uses UploadImageFromURL in createShopeeProduct
//   - Lazada: uses /image/migrate API in createLazadaProduct
//   - TikTok: uses multipart upload in createTiktokProduct
//
// This function simply passes through the source image URLs.
func (s *CloneService) processImages(_ context.Context, _ string, images []string) ([]string, error) {
	if len(images) == 0 {
		return images, nil
	}
	// All platforms handle image upload in their respective create*Product methods.
	// Return source URLs as-is — the platform-specific creator will process them.
	return images, nil
}
