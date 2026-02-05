package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test JSON marshaling for platform-specific models to ensure snake_case compliance

func TestProductTableName(t *testing.T) {
	product := Product{}
	assert.Equal(t, "products", product.TableName())
}

func TestProductSKUTableName(t *testing.T) {
	sku := ProductSKU{}
	assert.Equal(t, "product_skus", sku.TableName())
}

func TestMasterProductTableName(t *testing.T) {
	product := MasterProduct{}
	assert.Equal(t, "master_products", product.TableName())
}

func TestMasterProductSkuTableName(t *testing.T) {
	sku := MasterProductSku{}
	assert.Equal(t, "master_product_skus", sku.TableName())
}

func TestShopeeOrderTableName(t *testing.T) {
	order := ShopeeOrder{}
	assert.Equal(t, "shopee_orders", order.TableName())
}

func TestShopeeOrderItemTableName(t *testing.T) {
	item := ShopeeOrderItem{}
	assert.Equal(t, "shopee_order_items", item.TableName())
}

func TestShopeeProductTableName(t *testing.T) {
	product := ShopeeProduct{}
	assert.Equal(t, "shopee_products", product.TableName())
}

func TestShopeeSkuTableName(t *testing.T) {
	sku := ShopeeSku{}
	assert.Equal(t, "shopee_skus", sku.TableName())
}

func TestLazadaOrderTableName(t *testing.T) {
	order := LazadaOrder{}
	assert.Equal(t, "lazada_orders", order.TableName())
}

func TestLazadaOrderItemTableName(t *testing.T) {
	item := LazadaOrderItem{}
	assert.Equal(t, "lazada_order_items", item.TableName())
}

func TestLazadaProductTableName(t *testing.T) {
	product := LazadaProduct{}
	assert.Equal(t, "lazada_products", product.TableName())
}

func TestLazadaSkuTableName(t *testing.T) {
	sku := LazadaSku{}
	assert.Equal(t, "lazada_skus", sku.TableName())
}

func TestTiktokOrderTableName(t *testing.T) {
	order := TiktokOrder{}
	assert.Equal(t, "tiktok_orders", order.TableName())
}

func TestTiktokOrderItemTableName(t *testing.T) {
	item := TiktokOrderItem{}
	assert.Equal(t, "tiktok_order_items", item.TableName())
}

func TestTiktokProductTableName(t *testing.T) {
	product := TiktokProduct{}
	assert.Equal(t, "tiktok_products", product.TableName())
}

func TestTiktokSkuTableName(t *testing.T) {
	sku := TiktokSku{}
	assert.Equal(t, "tiktok_skus", sku.TableName())
}

func TestWebhookLogTableName(t *testing.T) {
	log := WebhookLog{}
	assert.Equal(t, "webhook_logs", log.TableName())
}

func TestWebhookOrderEventTableName(t *testing.T) {
	event := WebhookOrderEvent{}
	assert.Equal(t, "webhook_order_events", event.TableName())
}

func TestWebhookProductEventTableName(t *testing.T) {
	event := WebhookProductEvent{}
	assert.Equal(t, "webhook_product_events", event.TableName())
}

func TestJobTableName(t *testing.T) {
	job := Job{}
	assert.Equal(t, "jobs", job.TableName())
}

func TestJobHistoryTableName(t *testing.T) {
	history := JobHistory{}
	assert.Equal(t, "job_history", history.TableName())
}

func TestAuditLogTableName(t *testing.T) {
	log := AuditLog{}
	assert.Equal(t, "audit_logs", log.TableName())
}

func TestInventorySettingsTableName(t *testing.T) {
	settings := InventorySettings{}
	assert.Equal(t, "inventory_settings", settings.TableName())
}

func TestInventoryRecordTableName(t *testing.T) {
	record := InventoryRecord{}
	assert.Equal(t, "inventory_records", record.TableName())
}

func TestInventorySyncHistoryTableName(t *testing.T) {
	history := InventorySyncHistory{}
	assert.Equal(t, "inventory_sync_history", history.TableName())
}

// JSON marshaling tests for snake_case compliance
func TestProductJSONMarshalingSnakeCase(t *testing.T) {
	product := &Product{
		ID:        "prod123",
		TenantID:  "tenant1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(product)
	require.NoError(t, err)

	var result map[string]interface{}
	err = json.Unmarshal(data, &result)
	require.NoError(t, err)

	// All JSON keys should be snake_case
	jsonStr := string(data)
	// Should not have any camelCase or PascalCase keys
	assert.Contains(t, jsonStr, "created_at")
	assert.Contains(t, jsonStr, "updated_at")
	assert.Contains(t, jsonStr, "tenant_id")
}

func TestShopeeOrderJSONMarshalingSnakeCase(t *testing.T) {
	order := &ShopeeOrder{
		ID:        123,
		TenantID:  "tenant1",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	data, err := json.Marshal(order)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "tenant_id")
	assert.Contains(t, jsonStr, "created_at")
	assert.Contains(t, jsonStr, "updated_at")
}

func TestInventoryRecordJSONMarshalingSnakeCase(t *testing.T) {
	record := &InventoryRecord{
		ID:            "inv123",
		TenantID:      "tenant1",
		KeyColumnName: "SKU",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	data, err := json.Marshal(record)
	require.NoError(t, err)

	jsonStr := string(data)
	assert.Contains(t, jsonStr, "tenant_id")
	assert.Contains(t, jsonStr, "key_column_name")
	assert.Contains(t, jsonStr, "created_at")
	assert.Contains(t, jsonStr, "updated_at")
}
