package products

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- NewCategoryMapper ---

func TestNewCategoryMapper_ReturnsNonNil(t *testing.T) {
	m := NewCategoryMapper()
	assert.NotNil(t, m)
}

// --- buildMappingKey ---

func TestBuildMappingKey_ProducesNonEmptyKey(t *testing.T) {
	key := buildMappingKey("shopee", 123)
	assert.NotEmpty(t, key)
	assert.Contains(t, key, "shopee")
}

func TestBuildMappingKey_DifferentPlatforms_DifferentKeys(t *testing.T) {
	k1 := buildMappingKey("shopee", 1)
	k2 := buildMappingKey("lazada", 1)
	assert.NotEqual(t, k1, k2)
}

// --- FindCategoryByName ---

func TestFindCategoryByName_EmptyList_ReturnsNil(t *testing.T) {
	m := NewCategoryMapper()
	result := m.FindCategoryByName([]Category{}, "Electronics")
	assert.Nil(t, result)
}

func TestFindCategoryByName_Found_ReturnsCategory(t *testing.T) {
	m := NewCategoryMapper()
	cats := []Category{
		{ID: 1, Name: "Electronics"},
		{ID: 2, Name: "Clothing"},
	}
	result := m.FindCategoryByName(cats, "Clothing")
	require.NotNil(t, result)
	assert.Equal(t, int64(2), result.ID)
	assert.Equal(t, "Clothing", result.Name)
}

func TestFindCategoryByName_NotFound_ReturnsNil(t *testing.T) {
	m := NewCategoryMapper()
	cats := []Category{
		{ID: 1, Name: "Electronics"},
	}
	result := m.FindCategoryByName(cats, "Books")
	assert.Nil(t, result)
}

func TestFindCategoryByName_FindsInChildren(t *testing.T) {
	m := NewCategoryMapper()
	cats := []Category{
		{
			ID:   1,
			Name: "Electronics",
			Children: []Category{
				{ID: 10, Name: "Phones"},
				{ID: 11, Name: "Tablets"},
			},
		},
	}
	result := m.FindCategoryByName(cats, "Tablets")
	require.NotNil(t, result)
	assert.Equal(t, int64(11), result.ID)
}

func TestFindCategoryByName_DeepNested(t *testing.T) {
	m := NewCategoryMapper()
	cats := []Category{
		{
			ID:   1,
			Name: "Root",
			Children: []Category{
				{
					ID:   2,
					Name: "Mid",
					Children: []Category{
						{ID: 3, Name: "Leaf"},
					},
				},
			},
		},
	}
	result := m.FindCategoryByName(cats, "Leaf")
	require.NotNil(t, result)
	assert.Equal(t, int64(3), result.ID)
}

// --- MapCategory ---

func TestMapCategory_KeyFound_ReturnsTargetID(t *testing.T) {
	m := NewCategoryMapper()

	sourceID := int64(100)
	targetID := int64(999)
	key := buildMappingKey("shopee", sourceID)

	mappings := map[string]CategoryMapping{
		key: {
			SourcePlatform: "shopee",
			SourceID:       sourceID,
			TargetIDs:      map[string]int64{"lazada": targetID},
		},
	}

	result := m.MapCategory(sourceID, "shopee", "lazada", mappings)
	require.NotNil(t, result)
	assert.Equal(t, targetID, *result)
}

func TestMapCategory_KeyNotFound_ReturnsNil(t *testing.T) {
	m := NewCategoryMapper()
	mappings := map[string]CategoryMapping{}

	result := m.MapCategory(99, "shopee", "lazada", mappings)
	assert.Nil(t, result)
}

func TestMapCategory_TargetPlatformNotInMapping_ReturnsNil(t *testing.T) {
	m := NewCategoryMapper()

	sourceID := int64(10)
	key := buildMappingKey("shopee", sourceID)

	mappings := map[string]CategoryMapping{
		key: {
			SourcePlatform: "shopee",
			SourceID:       sourceID,
			TargetIDs:      map[string]int64{"lazada": 200}, // no tiktok
		},
	}

	result := m.MapCategory(sourceID, "shopee", "tiktok", mappings)
	assert.Nil(t, result)
}

func TestMapCategory_EmptyMappings_ReturnsNil(t *testing.T) {
	m := NewCategoryMapper()
	result := m.MapCategory(1, "shopee", "tiktok", nil)
	assert.Nil(t, result)
}

// --- CategoryMapping struct ---

func TestCategoryMapping_JSONFields(t *testing.T) {
	cm := CategoryMapping{
		SourcePlatform: "shopee",
		SourceID:       42,
		SourceName:     "Electronics",
		TargetIDs:      map[string]int64{"lazada": 99, "tiktok": 77},
	}
	assert.Equal(t, "shopee", cm.SourcePlatform)
	assert.Equal(t, int64(42), cm.SourceID)
	assert.Equal(t, int64(99), cm.TargetIDs["lazada"])
	assert.Equal(t, int64(77), cm.TargetIDs["tiktok"])
}

// Category struct helper (needed for tests — uses the PlatformAPI interface's Category type)
// Ensure Category has ID, Name, Children fields as required by the tests
func init() {
	// Verify Category can be constructed with expected fields
	_ = Category{ID: 1, Name: "test"}
	_ = fmt.Sprintf // prevent unused import if needed
}
