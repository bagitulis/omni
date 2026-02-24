package products

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// --- GetPlatformLimits ---

func TestGetPlatformLimits_Shopee(t *testing.T) {
	titleLimit, descLimit := GetPlatformLimits("shopee")
	assert.Equal(t, ShopeeTitleLimit, titleLimit)
	assert.Equal(t, ShopeeDescLimit, descLimit)
}

func TestGetPlatformLimits_Tiktok(t *testing.T) {
	titleLimit, descLimit := GetPlatformLimits("tiktok")
	assert.Equal(t, TiktokTitleLimit, titleLimit)
	assert.Equal(t, TiktokDescLimit, descLimit)
}

func TestGetPlatformLimits_Lazada(t *testing.T) {
	titleLimit, descLimit := GetPlatformLimits("lazada")
	assert.Equal(t, LazadaTitleLimit, titleLimit)
	assert.Equal(t, LazadaDescLimit, descLimit)
}

func TestGetPlatformLimits_Unknown_ReturnsSafeDefaults(t *testing.T) {
	titleLimit, descLimit := GetPlatformLimits("unknown")
	assert.Equal(t, 255, titleLimit)
	assert.Equal(t, 10000, descLimit)
}

// --- truncateTitle ---

func TestTruncateTitle_WithinLimit_NoChange(t *testing.T) {
	title := "Short title"
	result := truncateTitle(title, 120)
	assert.Equal(t, title, result)
}

func TestTruncateTitle_AtExactLimit_NoChange(t *testing.T) {
	title := strings.Repeat("a", 120)
	result := truncateTitle(title, 120)
	assert.Equal(t, title, result)
}

func TestTruncateTitle_ExceedsLimit_Truncated(t *testing.T) {
	title := strings.Repeat("a", 200)
	result := truncateTitle(title, 120)
	assert.LessOrEqual(t, len(result), 120)
	assert.True(t, strings.HasSuffix(result, "..."))
}

func TestTruncateTitle_WordBoundaryRespected(t *testing.T) {
	// Build a title that exceeds 80 chars, ends with a space before limit
	title := strings.Repeat("word ", 20) // 100 chars
	result := truncateTitle(title, 80)
	assert.LessOrEqual(t, len(result), 80)
	assert.True(t, strings.HasSuffix(result, "..."))
}

func TestTruncateTitle_EmptyString_ReturnsEmpty(t *testing.T) {
	result := truncateTitle("", 120)
	assert.Equal(t, "", result)
}

// --- truncateDescription ---

func TestTruncateDescription_WithinLimit_NoChange(t *testing.T) {
	desc := "Short description."
	result := truncateDescription(desc, 3000)
	assert.Equal(t, desc, result)
}

func TestTruncateDescription_ExceedsLimit_Truncated(t *testing.T) {
	desc := strings.Repeat("x", 5000)
	result := truncateDescription(desc, 3000)
	assert.LessOrEqual(t, len(result), 3000)
}

func TestTruncateDescription_SentenceBoundary(t *testing.T) {
	// Build a string that exceeds limit and has a sentence boundary before it
	base := strings.Repeat("word ", 100) // ~500 chars with spaces
	desc := "Hello world. This is a sentence. " + base
	// If the sentence boundary is hit, result should not end with "..."
	result := truncateDescription(desc, 100)
	assert.LessOrEqual(t, len(result), 100)
}

func TestTruncateDescription_EmptyString_ReturnsEmpty(t *testing.T) {
	result := truncateDescription("", 3000)
	assert.Equal(t, "", result)
}

// --- AdjustForPlatform ---

func TestAdjustForPlatform_NilInput_ReturnsNil(t *testing.T) {
	result := AdjustForPlatform(nil, "shopee")
	assert.Nil(t, result)
}

func TestAdjustForPlatform_Shopee_TruncatesLongTitle(t *testing.T) {
	pd := &ProductData{
		Name:        strings.Repeat("a", 200),
		Description: "Short desc",
	}
	result := AdjustForPlatform(pd, "shopee")
	assert.LessOrEqual(t, len(result.Name), ShopeeTitleLimit)
}

func TestAdjustForPlatform_Tiktok_TruncatesLongTitle(t *testing.T) {
	pd := &ProductData{
		Name:        strings.Repeat("a", 200),
		Description: "Short desc",
	}
	result := AdjustForPlatform(pd, "tiktok")
	assert.LessOrEqual(t, len(result.Name), TiktokTitleLimit)
}

func TestAdjustForPlatform_Lazada_TruncatesLongTitle(t *testing.T) {
	pd := &ProductData{
		Name:        strings.Repeat("a", 400),
		Description: "Short desc",
	}
	result := AdjustForPlatform(pd, "lazada")
	assert.LessOrEqual(t, len(result.Name), LazadaTitleLimit)
}

func TestAdjustForPlatform_DoesNotMutateOriginal(t *testing.T) {
	originalName := strings.Repeat("x", 200)
	pd := &ProductData{Name: originalName}
	_ = AdjustForPlatform(pd, "shopee")
	assert.Equal(t, originalName, pd.Name) // original unchanged
}

func TestAdjustForPlatform_ShortTitle_Unchanged(t *testing.T) {
	pd := &ProductData{Name: "Short", Description: "Brief"}
	result := AdjustForPlatform(pd, "shopee")
	assert.Equal(t, "Short", result.Name)
	assert.Equal(t, "Brief", result.Description)
}

func TestAdjustForPlatform_UnknownPlatform_Unchanged(t *testing.T) {
	pd := &ProductData{
		Name:        strings.Repeat("x", 400),
		Description: "desc",
	}
	result := AdjustForPlatform(pd, "unknown")
	// Unknown platform: no truncation applied
	assert.Equal(t, pd.Name, result.Name)
}

// --- WillTruncate ---

func TestWillTruncate_NilInput_ReturnsFalse(t *testing.T) {
	titleTrunc, descTrunc := WillTruncate(nil, "shopee")
	assert.False(t, titleTrunc)
	assert.False(t, descTrunc)
}

func TestWillTruncate_ShortContent_BothFalse(t *testing.T) {
	pd := &ProductData{Name: "Short", Description: "Brief"}
	titleTrunc, descTrunc := WillTruncate(pd, "shopee")
	assert.False(t, titleTrunc)
	assert.False(t, descTrunc)
}

func TestWillTruncate_LongTitle_TitleTrue(t *testing.T) {
	pd := &ProductData{
		Name:        strings.Repeat("a", 200),
		Description: "Brief",
	}
	titleTrunc, descTrunc := WillTruncate(pd, "shopee")
	assert.True(t, titleTrunc)
	assert.False(t, descTrunc)
}

func TestWillTruncate_LongDesc_DescTrue(t *testing.T) {
	pd := &ProductData{
		Name:        "Short title",
		Description: strings.Repeat("x", 5000),
	}
	titleTrunc, descTrunc := WillTruncate(pd, "shopee")
	assert.False(t, titleTrunc)
	assert.True(t, descTrunc)
}

func TestWillTruncate_BothLong_BothTrue(t *testing.T) {
	pd := &ProductData{
		Name:        strings.Repeat("a", 300),
		Description: strings.Repeat("x", 5000),
	}
	titleTrunc, descTrunc := WillTruncate(pd, "shopee")
	assert.True(t, titleTrunc)
	assert.True(t, descTrunc)
}
