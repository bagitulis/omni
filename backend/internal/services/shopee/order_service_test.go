package shopee

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupOrderTestDB(t *testing.T) *gorm.DB {
	// Use unique DB per test to ensure isolation - no cache=shared to prevent data pollution
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.AutoMigrate(&models.ShopeeOrder{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestNewOrderService(t *testing.T) {
	db := setupOrderTestDB(t)
	service := NewOrderService(db)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repo)
}

func TestOrderService_GetOrders(t *testing.T) {
	db := setupOrderTestDB(t)
	ctx := context.Background()

	// Seed test data
	order1 := models.ShopeeOrder{OrderSN: "ORDER001", OrderStatus: "READY_TO_SHIP"}
	order2 := models.ShopeeOrder{OrderSN: "ORDER002", OrderStatus: "SHIPPED"}
	db.Create(&order1)
	db.Create(&order2)

	service := NewOrderService(db)

	// Test GetOrders with pagination
	orders, total, err := service.GetOrders(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(2), total)
	assert.Len(t, orders, 2)
}

func TestOrderService_GetOrders_Empty(t *testing.T) {
	db := setupOrderTestDB(t)
	ctx := context.Background()

	service := NewOrderService(db)

	orders, total, err := service.GetOrders(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(0), total)
	assert.Len(t, orders, 0)
}

func TestOrderService_GetOrderBySN_Success(t *testing.T) {
	db := setupOrderTestDB(t)
	ctx := context.Background()

	// Seed test data
	order := models.ShopeeOrder{OrderSN: "ORDER123", OrderStatus: "READY_TO_SHIP", BuyerUsername: "buyer1"}
	db.Create(&order)

	service := NewOrderService(db)

	result, err := service.GetOrderBySN(ctx, "ORDER123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "ORDER123", result.OrderSN)
	assert.Equal(t, "buyer1", result.BuyerUsername)
}

func TestOrderService_GetOrderBySN_NotFound(t *testing.T) {
	db := setupOrderTestDB(t)
	ctx := context.Background()

	service := NewOrderService(db)

	result, err := service.GetOrderBySN(ctx, "NONEXISTENT")

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestOrderService_GetOrders_Pagination(t *testing.T) {
	db := setupOrderTestDB(t)
	ctx := context.Background()

	// Seed 15 test orders
	for i := 0; i < 15; i++ {
		order := models.ShopeeOrder{OrderSN: "ORDER" + string(rune('A'+i)), OrderStatus: "READY_TO_SHIP"}
		db.Create(&order)
	}

	service := NewOrderService(db)

	// Test first page
	orders, total, err := service.GetOrders(ctx, 1, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(15), total)
	assert.Len(t, orders, 10)

	// Test second page
	orders2, total2, err := service.GetOrders(ctx, 2, 10)

	assert.NoError(t, err)
	assert.Equal(t, int64(15), total2)
	assert.Len(t, orders2, 5)
}
