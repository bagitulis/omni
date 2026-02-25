package master_product

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestImportRequest_StructFields(t *testing.T) {
	t.Run("struct can be instantiated with fields", func(t *testing.T) {
		req := ImportRequest{
			Platform: "shopee",
			ItemID:   "123456",
		}
		assert.Equal(t, "shopee", req.Platform)
		assert.Equal(t, "123456", req.ItemID)

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "platform")
		assert.Contains(t, m, "item_id")
	})
}

func TestFileImportRequest_StructFields(t *testing.T) {
	t.Run("struct with rows can be marshaled", func(t *testing.T) {
		req := FileImportRequest{
			Rows: []FileImportRowRequest{
				{
					RowNumber: 1,
					ItemName:  "Test Product",
					ItemSku:   "SKU-001",
					Price:     10000.0,
					Stock:     5,
					Valid:     true,
				},
			},
		}
		assert.Len(t, req.Rows, 1)
		assert.Equal(t, "Test Product", req.Rows[0].ItemName)

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "rows")
	})
}

func TestFileImportRowRequest_StructFields(t *testing.T) {
	t.Run("all fields marshal to snake_case JSON", func(t *testing.T) {
		row := FileImportRowRequest{
			RowNumber:   2,
			ItemName:    "My Item",
			ItemSku:     "SKU-002",
			VariantName: "Red",
			Price:       9999.0,
			Stock:       10,
			BatchKey:    "batch-1",
			Description: "Item description",
			ImageUrls:   []string{"http://img.example.com/1.jpg"},
			Valid:       true,
			Errors:      []string{},
		}
		data, err := json.Marshal(row)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "row_number")
		assert.Contains(t, m, "item_name")
		assert.Contains(t, m, "item_sku")
		assert.Contains(t, m, "price")
		assert.Contains(t, m, "stock")
		assert.Contains(t, m, "valid")
	})
}

func TestAutoMapRequest_StructFields(t *testing.T) {
	t.Run("struct marshals to snake_case JSON", func(t *testing.T) {
		req := AutoMapRequest{
			SellerSku: "SELLER-SKU-001",
		}
		assert.Equal(t, "SELLER-SKU-001", req.SellerSku)

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "seller_sku")
	})
}

func TestAutoMapBatchRequest_StructFields(t *testing.T) {
	t.Run("struct with skus slice can be marshaled", func(t *testing.T) {
		req := AutoMapBatchRequest{
			Skus: []string{"SKU-001", "SKU-002", "SKU-003"},
		}
		assert.Len(t, req.Skus, 3)

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "skus")
	})
}

func TestManualLinkRequest_StructFields(t *testing.T) {
	t.Run("struct marshals to snake_case JSON", func(t *testing.T) {
		req := ManualLinkRequest{
			MasterSkuID:    1,
			Platform:       "shopee",
			PlatformItemID: "item-123",
			PlatformSkuID:  "sku-456",
		}
		assert.Equal(t, uint(1), req.MasterSkuID)
		assert.Equal(t, "shopee", req.Platform)

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "master_sku_id")
		assert.Contains(t, m, "platform")
		assert.Contains(t, m, "platform_item_id")
		assert.Contains(t, m, "platform_sku_id")
	})
}

func TestUnlinkRequest_StructFields(t *testing.T) {
	t.Run("struct marshals to snake_case JSON", func(t *testing.T) {
		req := UnlinkRequest{
			MasterSkuID: 5,
			Platform:    "tiktok",
		}
		assert.Equal(t, uint(5), req.MasterSkuID)
		assert.Equal(t, "tiktok", req.Platform)

		data, err := json.Marshal(req)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "master_sku_id")
		assert.Contains(t, m, "platform")
	})
}
