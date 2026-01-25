package products

import (
	"context"
)

// CategoryMapper handles category mapping between platforms
// NO CACHING - always queries fresh data
type CategoryMapper struct {
}

// NewCategoryMapper creates a new category mapper
func NewCategoryMapper() *CategoryMapper {
	return &CategoryMapper{}
}

// GetCategories retrieves categories for a platform
// NO CACHING - always queries fresh data from API
func (m *CategoryMapper) GetCategories(ctx context.Context, api PlatformAPI, platform string, parentID int64) ([]Category, error) {
	// Always fetch fresh from API - no caching
	return api.GetCategories(ctx, parentID)
}

// GetCategoryTree retrieves full category tree
func (m *CategoryMapper) GetCategoryTree(ctx context.Context, api PlatformAPI, platform string) ([]Category, error) {
	return m.GetCategories(ctx, api, platform, 0)
}

// FindCategoryByName searches for a category by name
func (m *CategoryMapper) FindCategoryByName(categories []Category, name string) *Category {
	for i := range categories {
		if categories[i].Name == name {
			return &categories[i]
		}
		if len(categories[i].Children) > 0 {
			if found := m.FindCategoryByName(categories[i].Children, name); found != nil {
				return found
			}
		}
	}
	return nil
}

// MapCategory maps a category from source to target platform
func (m *CategoryMapper) MapCategory(sourceID int64, sourcePlatform, targetPlatform string, mappings map[string]CategoryMapping) *int64 {
	key := buildMappingKey(sourcePlatform, sourceID)
	if mapping, ok := mappings[key]; ok {
		if targetID, ok := mapping.TargetIDs[targetPlatform]; ok {
			return &targetID
		}
	}
	return nil
}

// CategoryMapping represents cross-platform category mapping
type CategoryMapping struct {
	SourcePlatform string           `json:"source_platform"`
	SourceID       int64            `json:"source_id"`
	SourceName     string           `json:"source_name"`
	TargetIDs      map[string]int64 `json:"target_ids"` // platform -> category_id
}

func buildMappingKey(platform string, id int64) string {
	return platform + "_" + string(rune(id))
}
