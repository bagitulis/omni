package shopee

import (
	"context"
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func setupProductTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to create test database: %v", err)
	}

	err = db.AutoMigrate(&models.ShopeeProduct{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func TestNewProductService(t *testing.T) {
	db := setupProductTestDB(t)
	service := NewProductService(db)

	assert.NotNil(t, service)
	assert.NotNil(t, service.repo)
}

func TestProductService_GetProductByItemID_Success(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()

	// Seed test data
	product := models.ShopeeProduct{
		ItemID:      12345,
		Name:        "Test Product",
		Description: "Test Description",
		Status:      "NORMAL",
		Price:       100.50,
		Quantity:    10,
	}
	db.Create(&product)

	service := NewProductService(db)

	result, err := service.GetProductByItemID(ctx, 12345)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(12345), result.ItemID)
	assert.Equal(t, "Test Product", result.Name)
	assert.Equal(t, 100.50, result.Price)
}

func TestProductService_GetProductByItemID_NotFound(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()

	service := NewProductService(db)

	result, err := service.GetProductByItemID(ctx, 99999)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestProductService_GetProductByItemID_MultipleProducts(t *testing.T) {
	db := setupProductTestDB(t)
	ctx := context.Background()

	// Seed multiple products
	product1 := models.ShopeeProduct{ItemID: 11111, Name: "Product 1", Status: "NORMAL", Price: 50.0}
	product2 := models.ShopeeProduct{ItemID: 22222, Name: "Product 2", Status: "NORMAL", Price: 75.0}
	product3 := models.ShopeeProduct{ItemID: 33333, Name: "Product 3", Status: "BANNED", Price: 100.0}
	db.Create(&product1)
	db.Create(&product2)
	db.Create(&product3)

	service := NewProductService(db)

	// Fetch each product
	result1, err1 := service.GetProductByItemID(ctx, 11111)
	result2, err2 := service.GetProductByItemID(ctx, 22222)
	result3, err3 := service.GetProductByItemID(ctx, 33333)

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.Equal(t, "Product 1", result1.Name)
	assert.Equal(t, "Product 2", result2.Name)
	assert.Equal(t, "BANNED", result3.Status)
}
