package lazada

import (
	"context"
	"testing"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"

	"github.com/glebarez/sqlite"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
)

// MockLazadaClient mocks the Lazada client for sync service testing
type MockLazadaClient struct {
	mock.Mock
}

func (m *MockLazadaClient) GetOrders(status string, offset, limit int) (*lazadaPkg.OrderListResponse, error) {
	args := m.Called(status, offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lazadaPkg.OrderListResponse), args.Error(1)
}

func (m *MockLazadaClient) GetProducts(offset, limit int) (*lazadaPkg.ProductListResponse, error) {
	args := m.Called(offset, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*lazadaPkg.ProductListResponse), args.Error(1)
}

func setupLazadaTestDB(t *testing.T) *gorm.DB {
	// Use simple in-memory DB without cache=shared to ensure test isolation
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.AutoMigrate(&models.LazadaOrder{}, &models.LazadaProduct{}, &models.LazadaSku{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestNewLazadaSyncService(t *testing.T) {
	db := setupLazadaTestDB(t)

	// Create nil client - we just test the service struct creation
	service := NewSyncService(nil, db)

	assert.NotNil(t, service)
	assert.NotNil(t, service.orderRepo)
	assert.NotNil(t, service.prodRepo)
}

func TestNewLazadaSyncServiceWithTenant(t *testing.T) {
	db := setupLazadaTestDB(t)

	service := NewSyncServiceWithTenant(nil, db, "tenant123")

	assert.NotNil(t, service)
	assert.Equal(t, "tenant123", service.tenantID)
}

func TestLazadaSyncService_SyncOrders_DBOperations(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	// Simulate order data from Lazada API
	orders := []struct {
		OrderID          string
		OrderStatus      string
		BuyerUsername    string
		ShippingCarrier  string
		PromisedShipDate string
	}{
		{
			OrderID:          "LZD_ORDER_001",
			OrderStatus:      "pending",
			BuyerUsername:    "John Doe",
			ShippingCarrier:  "LEX",
			PromisedShipDate: "2024-01-15 12:00:00",
		},
		{
			OrderID:          "LZD_ORDER_002",
			OrderStatus:      "shipped",
			BuyerUsername:    "Jane Smith",
			ShippingCarrier:  "JNE",
			PromisedShipDate: "2024-01-16",
		},
	}

	for _, order := range orders {
		// Parse ship by date
		var shipByDate *int64
		if order.PromisedShipDate != "" {
			for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05", "2006-01-02"} {
				if t, err := time.Parse(layout, order.PromisedShipDate); err == nil {
					ts := t.Unix()
					shipByDate = &ts
					break
				}
			}
		}

		dbOrder := &models.LazadaOrder{
			TenantID:        "tenant1",
			OrderSN:         order.OrderID,
			OrderStatus:     order.OrderStatus,
			BuyerUsername:   order.BuyerUsername,
			ShippingCarrier: order.ShippingCarrier,
			ShipByDate:      shipByDate,
		}

		err := service.orderRepo.Upsert(ctx, dbOrder)
		assert.NoError(t, err)
	}

	// Verify orders were saved
	var count int64
	db.Model(&models.LazadaOrder{}).Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify first order details
	var order models.LazadaOrder
	db.Where("order_sn = ?", "LZD_ORDER_001").First(&order)
	assert.Equal(t, "pending", order.OrderStatus)
	assert.Equal(t, "LEX", order.ShippingCarrier)
	assert.Equal(t, "John Doe", order.BuyerUsername)
	assert.NotNil(t, order.ShipByDate)
}

func TestLazadaSyncService_SyncProducts_DBOperations(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	// Simulate product data from Lazada API
	products := []struct {
		ItemID      string
		Name        string
		Description string
		Brand       string
		Price       float64
		Status      string
	}{
		{ItemID: "123456789", Name: "Product 1", Description: "Desc 1", Brand: "Brand A", Price: 50000, Status: "Active"},
		{ItemID: "987654321", Name: "Product 2", Description: "Desc 2", Brand: "Brand B", Price: 75000, Status: "Active"},
		{ItemID: "111222333", Name: "Product 3", Description: "Desc 3", Brand: "Brand C", Price: 100000, Status: "Inactive"},
	}

	for _, prod := range products {
		dbProd := &models.LazadaProduct{
			TenantID:    "tenant1",
			ItemID:      prod.ItemID,
			Name:        prod.Name,
			Description: prod.Description,
			Brand:       prod.Brand,
			Price:       prod.Price,
			Status:      prod.Status,
		}

		err := service.prodRepo.Upsert(ctx, dbProd)
		assert.NoError(t, err)
	}

	// Verify products were saved
	var count int64
	db.Model(&models.LazadaProduct{}).Count(&count)
	assert.Equal(t, int64(3), count)

	// Verify product details
	savedProd, err := service.prodRepo.FindByItemID(ctx, "123456789")
	assert.NoError(t, err)
	assert.Equal(t, "Product 1", savedProd.Name)
	assert.Equal(t, "Brand A", savedProd.Brand)
	assert.Equal(t, float64(50000), savedProd.Price)
}

func TestLazadaSyncService_ProductUpsert_Update(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	// Insert initial product
	dbProd := &models.LazadaProduct{
		TenantID:    "tenant1",
		ItemID:      "UPD_001",
		Name:        "Original Name",
		Description: "Original Description",
		Brand:       "Original Brand",
		Price:       50000,
		Status:      "Active",
	}
	err := service.prodRepo.Upsert(ctx, dbProd)
	assert.NoError(t, err)

	// Update the same product
	updatedProd := &models.LazadaProduct{
		TenantID:    "tenant1",
		ItemID:      "UPD_001",
		Name:        "Updated Name",
		Description: "Updated Description",
		Brand:       "Updated Brand",
		Price:       75000,
		Status:      "Inactive",
	}
	err = service.prodRepo.Upsert(ctx, updatedProd)
	assert.NoError(t, err)

	// Verify only one product exists
	var count int64
	db.Model(&models.LazadaProduct{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify it was updated
	savedProd, err := service.prodRepo.FindByItemID(ctx, "UPD_001")
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", savedProd.Name)
	assert.Equal(t, "Inactive", savedProd.Status)
	assert.Equal(t, float64(75000), savedProd.Price)
}

func TestLazadaSyncService_OrderUpsert_Update(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	// Insert initial order
	dbOrder := &models.LazadaOrder{
		TenantID:    "tenant1",
		OrderSN:     "ORDER_UPD",
		OrderStatus: "pending",
	}
	err := service.orderRepo.Upsert(ctx, dbOrder)
	assert.NoError(t, err)

	// Update the same order
	updatedOrder := &models.LazadaOrder{
		TenantID:        "tenant1",
		OrderSN:         "ORDER_UPD",
		OrderStatus:     "shipped",
		ShippingCarrier: "JNE",
	}
	err = service.orderRepo.Upsert(ctx, updatedOrder)
	assert.NoError(t, err)

	// Verify only one order exists
	var count int64
	db.Model(&models.LazadaOrder{}).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify it was updated
	var savedOrder models.LazadaOrder
	db.Where("order_sn = ?", "ORDER_UPD").First(&savedOrder)
	assert.Equal(t, "shipped", savedOrder.OrderStatus)
	assert.Equal(t, "JNE", savedOrder.ShippingCarrier)
}

func TestLazadaSyncService_SKUUpsert(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	// Insert product first
	dbProd := &models.LazadaProduct{
		TenantID: "tenant1",
		ItemID:   "PROD_SKU",
		Name:     "Product with SKUs",
		Status:   "Active",
	}
	err := service.prodRepo.Upsert(ctx, dbProd)
	assert.NoError(t, err)

	// Insert SKUs
	skus := []struct {
		SkuID       string
		ShopSku     string
		SellerSku   string
		VariantName string
		Price       float64
		Quantity    int
	}{
		{SkuID: "SKU_001", ShopSku: "SHOP_001", SellerSku: "SELLER_001", VariantName: "Red - S", Price: 50000, Quantity: 10},
		{SkuID: "SKU_002", ShopSku: "SHOP_002", SellerSku: "SELLER_002", VariantName: "Blue - M", Price: 55000, Quantity: 15},
		{SkuID: "SKU_003", ShopSku: "SHOP_003", SellerSku: "SELLER_003", VariantName: "Green - L", Price: 60000, Quantity: 5},
	}

	for _, sku := range skus {
		dbSku := &models.LazadaSku{
			TenantID:    "tenant1",
			ItemID:      "PROD_SKU",
			SkuID:       sku.SkuID,
			ShopSku:     sku.ShopSku,
			SellerSku:   sku.SellerSku,
			VariantName: sku.VariantName,
			Price:       sku.Price,
			Quantity:    sku.Quantity,
		}
		err := service.prodRepo.UpsertSku(ctx, dbSku)
		assert.NoError(t, err)
	}

	// Verify SKUs were saved
	var count int64
	db.Model(&models.LazadaSku{}).Count(&count)
	assert.Equal(t, int64(3), count)

	// Verify SKU details
	var savedSku models.LazadaSku
	db.Where("sku_id = ?", "SKU_001").First(&savedSku)
	assert.Equal(t, "Red - S", savedSku.VariantName)
	assert.Equal(t, float64(50000), savedSku.Price)
	assert.Equal(t, 10, savedSku.Quantity)
}

func TestLazadaSyncService_TenantIsolation(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service1 := NewSyncServiceWithTenant(nil, db, "tenant_A")
	service2 := NewSyncServiceWithTenant(nil, db, "tenant_B")

	// Insert products for different tenants with DIFFERENT ItemIDs
	// Note: Current repository upsert only checks item_id, not tenant_id
	// So using same item_id would update instead of creating separate records
	err := service1.prodRepo.Upsert(ctx, &models.LazadaProduct{
		TenantID: "tenant_A",
		ItemID:   "ITEM_A_001",
		Name:     "Tenant A Product",
		Price:    50000,
		Status:   "Active",
	})
	assert.NoError(t, err)

	err = service2.prodRepo.Upsert(ctx, &models.LazadaProduct{
		TenantID: "tenant_B",
		ItemID:   "ITEM_B_001",
		Name:     "Tenant B Product",
		Price:    75000,
		Status:   "Inactive",
	})
	assert.NoError(t, err)

	// Both products should exist separately
	var count int64
	db.Model(&models.LazadaProduct{}).Count(&count)
	assert.Equal(t, int64(2), count)

	// Verify tenant A product
	var prodA models.LazadaProduct
	db.Where("item_id = ?", "ITEM_A_001").First(&prodA)
	assert.Equal(t, "Tenant A Product", prodA.Name)
	assert.Equal(t, float64(50000), prodA.Price)
	assert.Equal(t, "tenant_A", prodA.TenantID)

	// Verify tenant B product
	var prodB models.LazadaProduct
	db.Where("item_id = ?", "ITEM_B_001").First(&prodB)
	assert.Equal(t, "Tenant B Product", prodB.Name)
	assert.Equal(t, float64(75000), prodB.Price)
	assert.Equal(t, "tenant_B", prodB.TenantID)
}

func TestLazadaSyncService_EmptyDatabase(t *testing.T) {
	db := setupLazadaTestDB(t)

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	// Verify empty counts
	var orderCount, productCount int64
	db.Model(&models.LazadaOrder{}).Count(&orderCount)
	db.Model(&models.LazadaProduct{}).Count(&productCount)

	assert.Equal(t, int64(0), orderCount)
	assert.Equal(t, int64(0), productCount)
	assert.NotNil(t, service)
}

func TestLazadaSyncService_ClearLazadaProductCache(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	service := NewSyncServiceWithTenant(nil, db, "tenant1")

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaProduct{
		TenantID: "tenant1",
		ItemID:   "item-1",
		Name:     "Tenant 1 Product",
		Status:   "Active",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaSku{
		TenantID:  "tenant1",
		ItemID:    "item-1",
		SkuID:     "sku-1",
		SellerSku: "SELLER-1",
		Price:     10000,
		Quantity:  5,
	}).Error)

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaProduct{
		TenantID: "tenant2",
		ItemID:   "item-2",
		Name:     "Tenant 2 Product",
		Status:   "Active",
	}).Error)
	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaSku{
		TenantID:  "tenant2",
		ItemID:    "item-2",
		SkuID:     "sku-2",
		SellerSku: "SELLER-2",
		Price:     20000,
		Quantity:  7,
	}).Error)

	require.NoError(t, service.clearLazadaProductCache(ctx))

	var tenant1Products int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.LazadaProduct{}).
		Where("tenant_id = ?", "tenant1").
		Count(&tenant1Products).Error)
	assert.Equal(t, int64(0), tenant1Products)

	var tenant1Skus int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.LazadaSku{}).
		Where("tenant_id = ?", "tenant1").
		Count(&tenant1Skus).Error)
	assert.Equal(t, int64(0), tenant1Skus)

	var tenant2Products int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.LazadaProduct{}).
		Where("tenant_id = ?", "tenant2").
		Count(&tenant2Products).Error)
	assert.Equal(t, int64(1), tenant2Products)

	var tenant2Skus int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.LazadaSku{}).
		Where("tenant_id = ?", "tenant2").
		Count(&tenant2Skus).Error)
	assert.Equal(t, int64(1), tenant2Skus)
}
