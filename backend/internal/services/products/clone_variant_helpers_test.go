package products

import (
	"testing"

	"github.com/omni/backend/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- EnsureDescription ---

func TestEnsureDescription_LongEnough_ReturnsOriginal(t *testing.T) {
	desc := "This description is long enough."
	result := EnsureDescription(desc, "Product Name", 25)
	assert.Equal(t, desc, result)
}

func TestEnsureDescription_TooShort_ReturnsDefault(t *testing.T) {
	result := EnsureDescription("Short", "My Product", 25)
	assert.NotEqual(t, "Short", result)
	assert.Contains(t, result, "My Product")
}

func TestEnsureDescription_Empty_ReturnsDefault(t *testing.T) {
	result := EnsureDescription("", "Widget X", 25)
	assert.Contains(t, result, "Widget X")
}

func TestEnsureDescription_ExactlyAtLimit_ReturnsOriginal(t *testing.T) {
	desc := "12345678901234567890abcde" // 25 chars
	result := EnsureDescription(desc, "Some Product", 25)
	assert.Equal(t, desc, result)
}

func TestEnsureDescription_DefaultIncludesProductName(t *testing.T) {
	result := EnsureDescription("", "Super Widget", 25)
	assert.Contains(t, result, "Super Widget")
}

// --- buildVariants (generic, tested via buildVariantsFromShopee) ---

func TestBuildVariantsFromShopee_Empty(t *testing.T) {
	result := buildVariantsFromShopee([]models.ShopeeSku{})
	assert.Empty(t, result)
}

func TestBuildVariantsFromShopee_Single(t *testing.T) {
	skus := []models.ShopeeSku{
		{SellerSku: "SKU-A", VariantName: "Red", Price: 10.0, Quantity: 5},
	}
	result := buildVariantsFromShopee(skus)
	require.Len(t, result, 1)
	assert.Equal(t, "SKU-A", result[0].SKU)
	assert.Equal(t, "Red", result[0].Name)
	assert.Equal(t, 10.0, result[0].Price)
	assert.Equal(t, 5, result[0].Stock)
}

func TestBuildVariantsFromShopee_Multiple(t *testing.T) {
	skus := []models.ShopeeSku{
		{SellerSku: "SKU-A", VariantName: "Red", Price: 10.0, Quantity: 3},
		{SellerSku: "SKU-B", VariantName: "Blue", Price: 20.0, Quantity: 7},
	}
	result := buildVariantsFromShopee(skus)
	require.Len(t, result, 2)
	assert.Equal(t, "SKU-A", result[0].SKU)
	assert.Equal(t, "SKU-B", result[1].SKU)
}

// --- buildVariantsFromLazada ---

func TestBuildVariantsFromLazada_Empty(t *testing.T) {
	result := buildVariantsFromLazada([]models.LazadaSku{})
	assert.Empty(t, result)
}

func TestBuildVariantsFromLazada_UsesSkuIDWhenSellerSkuEmpty(t *testing.T) {
	skus := []models.LazadaSku{
		{SellerSku: "", SkuID: "LAZADA-SKU-ID", VariantName: "Green", Price: 15.0, Quantity: 2},
	}
	result := buildVariantsFromLazada(skus)
	require.Len(t, result, 1)
	assert.Equal(t, "LAZADA-SKU-ID", result[0].SKU)
}

func TestBuildVariantsFromLazada_PrefersSellerSku(t *testing.T) {
	skus := []models.LazadaSku{
		{SellerSku: "MY-SELLER-SKU", SkuID: "LAZADA-SKU-ID", VariantName: "Green", Price: 15.0, Quantity: 2},
	}
	result := buildVariantsFromLazada(skus)
	require.Len(t, result, 1)
	assert.Equal(t, "MY-SELLER-SKU", result[0].SKU)
}

// --- buildVariantsFromTiktok ---

func TestBuildVariantsFromTiktok_Empty(t *testing.T) {
	result := buildVariantsFromTiktok([]models.TiktokSku{})
	assert.Empty(t, result)
}

func TestBuildVariantsFromTiktok_UsesSkuIDWhenSellerSkuEmpty(t *testing.T) {
	skus := []models.TiktokSku{
		{SellerSku: "", SkuID: "TK-SKU-ID", VariantName: "Yellow", Price: 25.0, Quantity: 10},
	}
	result := buildVariantsFromTiktok(skus)
	require.Len(t, result, 1)
	assert.Equal(t, "TK-SKU-ID", result[0].SKU)
}

func TestBuildVariantsFromTiktok_PrefersSellerSku(t *testing.T) {
	skus := []models.TiktokSku{
		{SellerSku: "MY-TK-SELLER-SKU", SkuID: "TK-SKU-ID", VariantName: "Yellow", Price: 25.0, Quantity: 10},
	}
	result := buildVariantsFromTiktok(skus)
	require.Len(t, result, 1)
	assert.Equal(t, "MY-TK-SELLER-SKU", result[0].SKU)
}
