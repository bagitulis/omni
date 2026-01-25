package inventory

import (
	"context"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// PlatformProductIds represents product IDs for all platforms
type PlatformProductIds struct {
	Shopee *ShopeeProductIds `json:"shopee,omitempty"`
	Lazada *LazadaProductIds `json:"lazada,omitempty"`
	Tiktok *TiktokProductIds `json:"tiktok,omitempty"`
}

// ShopeeProductIds represents Shopee product IDs required for stock update
type ShopeeProductIds struct {
	ItemID    string `json:"item_id"`
	ModelID   string `json:"model_id,omitempty"` // Can be null for single variant products
	SellerSku string `json:"seller_sku,omitempty"`
}

// LazadaProductIds represents Lazada product IDs required for stock update
type LazadaProductIds struct {
	ItemID    string `json:"item_id"`
	SkuID     string `json:"sku_id"`
	SellerSku string `json:"seller_sku,omitempty"`
}

// TiktokProductIds represents TikTok product IDs required for stock update
type TiktokProductIds struct {
	ProductID string `json:"product_id"`
	SkuID     string `json:"sku_id"`
	SellerSku string `json:"seller_sku,omitempty"`
}

// ProductIdFetcher fetches platform product IDs from database
type ProductIdFetcher struct {
	db       *gorm.DB
	tenantID string
}

// NewProductIdFetcher creates a new ProductIdFetcher
func NewProductIdFetcher(db *gorm.DB, tenantID string) *ProductIdFetcher {
	return &ProductIdFetcher{db: db, tenantID: tenantID}
}

// FetchBySku fetches all platform product IDs by seller SKU
func (f *ProductIdFetcher) FetchBySku(ctx context.Context, sku string) (*PlatformProductIds, error) {
	result := &PlatformProductIds{}

	// Fetch all platforms concurrently (simplified - sequential for safety)
	result.Shopee = f.fetchShopeeIds(ctx, sku)
	result.Lazada = f.fetchLazadaIds(ctx, sku)
	result.Tiktok = f.fetchTiktokIds(ctx, sku)

	return result, nil
}

// fetchShopeeIds fetches Shopee product IDs
func (f *ProductIdFetcher) fetchShopeeIds(ctx context.Context, sku string) *ShopeeProductIds {
	var shopeeSku models.ShopeeSku

	// Try sellerSku first
	err := f.db.WithContext(ctx).
		Where("tenant_id = ? AND seller_sku = ?", f.tenantID, sku).
		First(&shopeeSku).Error

	if err == gorm.ErrRecordNotFound {
		// Try modelId as BigInt
		if modelID, parseErr := strconv.ParseInt(sku, 10, 64); parseErr == nil {
			err = f.db.WithContext(ctx).
				Where("tenant_id = ? AND model_id = ?", f.tenantID, modelID).
				First(&shopeeSku).Error
		}
	}

	if err != nil {
		return nil
	}

	result := &ShopeeProductIds{
		ItemID:    fmt.Sprintf("%d", shopeeSku.ItemID),
		SellerSku: shopeeSku.SellerSku,
	}

	if shopeeSku.ModelID != nil {
		result.ModelID = fmt.Sprintf("%d", *shopeeSku.ModelID)
	}

	return result
}

// fetchLazadaIds fetches Lazada product IDs
func (f *ProductIdFetcher) fetchLazadaIds(ctx context.Context, sku string) *LazadaProductIds {
	var lazadaSku models.LazadaSku

	// Try skuId or sellerSku
	err := f.db.WithContext(ctx).
		Where("tenant_id = ? AND (sku_id = ? OR seller_sku = ?)", f.tenantID, sku, sku).
		First(&lazadaSku).Error

	if err != nil {
		return nil
	}

	return &LazadaProductIds{
		ItemID:    lazadaSku.ItemID,
		SkuID:     lazadaSku.SkuID,
		SellerSku: lazadaSku.SellerSku,
	}
}

// fetchTiktokIds fetches TikTok product IDs
func (f *ProductIdFetcher) fetchTiktokIds(ctx context.Context, sku string) *TiktokProductIds {
	var tiktokSku models.TiktokSku
	var product models.TiktokProduct

	// Try skuId or sellerSku, order by id desc to get latest
	err := f.db.WithContext(ctx).
		Where("tenant_id = ? AND (sku_id = ? OR seller_sku = ?)", f.tenantID, sku, sku).
		Order("id DESC").
		First(&tiktokSku).Error

	if err != nil {
		return nil
	}

	// Get the product to retrieve product_id
	err = f.db.WithContext(ctx).
		Where("id = ?", tiktokSku.ProductID).
		First(&product).Error

	if err != nil {
		return nil
	}

	return &TiktokProductIds{
		ProductID: product.ProductID,
		SkuID:     tiktokSku.SkuID,
		SellerSku: tiktokSku.SellerSku,
	}
}

// FetchBatch fetches IDs for multiple SKUs
func (f *ProductIdFetcher) FetchBatch(ctx context.Context, skus []string) map[string]*PlatformProductIds {
	results := make(map[string]*PlatformProductIds)

	for _, sku := range skus {
		ids, _ := f.FetchBySku(ctx, sku)
		results[sku] = ids
	}

	return results
}
