package tiktok

import (
	"context"
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildVariantNameAndData(t *testing.T) {
	attrs := []tiktokPkg.ProductSalesAttr{
		{Name: "Color", ValueName: "Red"},
		{Name: "Size", ValueName: "XL"},
	}

	assert.Equal(t, "Color:Red | Size:XL", buildVariantName(attrs))
	assert.Equal(
		t,
		models.JSONMap{"Color": "Red", "Size": "XL"},
		buildVariantData(attrs),
	)
}

func TestSyncProductSkus_PopulatesVariantFields(t *testing.T) {
	db := setupTikTokTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant_variant")
	prodRepo := repositories.NewTiktokProductRepository(db)

	product := &models.TiktokProduct{
		TenantID:  "tenant_variant",
		ProductID: "PROD_VARIANT",
		Name:      "Variant Product",
		Status:    "LIVE",
	}
	require.NoError(t, prodRepo.Upsert(ctx, product))

	savedProduct, err := prodRepo.FindByProductID(ctx, "PROD_VARIANT")
	require.NoError(t, err)
	require.NotNil(t, savedProduct)

	searchProduct := tiktokPkg.ProductSearchItem{
		ID: "PROD_VARIANT",
		Skus: []tiktokPkg.ProductSearchSku{
			{
				ID:        "SKU_VARIANT_1",
				SellerSku: "SELLER-SKU-1",
				Price: struct {
					Currency          string `json:"currency"`
					TaxExclusivePrice string `json:"tax_exclusive_price"`
					SalePrice         string `json:"sale_price"`
					OriginalPrice     string `json:"original_price"`
				}{
					SalePrice: "100000",
				},
				Inventory: []struct {
					WarehouseID string `json:"warehouse_id"`
					Quantity    int    `json:"quantity"`
				}{
					{WarehouseID: "WH-1", Quantity: 5},
				},
			},
		},
	}

	detail := &tiktokPkg.ProductDetailResponse{}
	detail.Data.Skus = []tiktokPkg.ProductDetailSku{
		{
			ID: "SKU_VARIANT_1",
			SalesAttributes: []tiktokPkg.ProductSalesAttr{
				{Name: "Color", ValueName: "Red"},
				{Name: "Size", ValueName: "XL"},
			},
		},
	}

	service.syncProductSkus(ctx, prodRepo, savedProduct.ID, searchProduct, detail)

	var savedSku models.TiktokSku
	require.NoError(t, db.WithContext(ctx).Where("sku_id = ?", "SKU_VARIANT_1").First(&savedSku).Error)
	assert.Equal(t, "Color:Red | Size:XL", savedSku.VariantName)
	assert.Equal(t, models.JSONMap{"Color": "Red", "Size": "XL"}, savedSku.VariantData)
}
