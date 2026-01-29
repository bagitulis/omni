package analytics

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// CacheService handles Materialized View refresh operations
type CacheService struct {
	basePath string
}

// NewCacheService creates a new cache service
func NewCacheService(basePath string) *CacheService {
	return &CacheService{basePath: basePath}
}

// RefreshStatus contains MV refresh status
type RefreshStatus struct {
	ViewName      string    `json:"view_name"`
	LastRefreshed time.Time `json:"last_refreshed"`
	RowCount      int64     `json:"row_count"`
	Status        string    `json:"status"`
}

// CacheMetadata contains cache metadata
type CacheMetadata struct {
	TenantID    string    `json:"tenant_id"`
	ViewName    string    `json:"view_name"`
	LastRefresh time.Time `json:"last_refresh"`
	RefreshTime int64     `json:"refresh_time_ms"`
	RowCount    int64     `json:"row_count"`
}

// MaterializedViews lists all analytics MVs
var MaterializedViews = []string{
	"mv_tiktok_ads_summary",
	"mv_tiktok_ads_period_summary",
	"mv_ml_product_analysis",
	"mv_ml_portfolio_summary",
	"mv_shopee_ads_summary",
	"mv_shopee_ads_product_analysis",
}

// RefreshAllMVs refreshes all materialized views for a tenant
func (c *CacheService) RefreshAllMVs(ctx context.Context, db *gorm.DB, tenantID string) []RefreshStatus {
	results := make([]RefreshStatus, 0, len(MaterializedViews))

	for _, viewName := range MaterializedViews {
		status := c.RefreshMV(ctx, db, tenantID, viewName)
		results = append(results, status)
	}

	return results
}

// RefreshMV refreshes a single materialized view
func (c *CacheService) RefreshMV(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
	viewName string,
) RefreshStatus {
	startTime := time.Now()

	// Execute REFRESH MATERIALIZED VIEW CONCURRENTLY
	sql := "REFRESH MATERIALIZED VIEW CONCURRENTLY " + viewName
	err := db.WithContext(ctx).Exec(sql).Error

	refreshTime := time.Since(startTime).Milliseconds()

	if err != nil {
		// Try non-concurrent refresh if concurrent fails
		sql = "REFRESH MATERIALIZED VIEW " + viewName
		err = db.WithContext(ctx).Exec(sql).Error

		if err != nil {
			return RefreshStatus{
				ViewName: viewName,
				Status:   "ERROR: " + err.Error(),
			}
		}
	}

	// Get row count
	var rowCount int64
	db.WithContext(ctx).Table(viewName).
		Where("tenant_id = ?", tenantID).
		Count(&rowCount)

	// Update metadata
	c.updateMetadata(ctx, db, tenantID, viewName, refreshTime, rowCount)

	return RefreshStatus{
		ViewName:      viewName,
		LastRefreshed: time.Now(),
		RowCount:      rowCount,
		Status:        "SUCCESS",
	}
}

// updateMetadata updates cache metadata table
func (c *CacheService) updateMetadata(
	ctx context.Context,
	db *gorm.DB,
	tenantID, viewName string,
	refreshTimeMs int64,
	rowCount int64,
) {
	sql := `
		INSERT INTO analytics_cache_metadata 
		(tenant_id, view_name, last_refresh, refresh_time_ms, row_count, updated_at)
		VALUES (?, ?, NOW(), ?, ?, NOW())
		ON CONFLICT (tenant_id, view_name) 
		DO UPDATE SET 
			last_refresh = NOW(),
			refresh_time_ms = EXCLUDED.refresh_time_ms,
			row_count = EXCLUDED.row_count,
			updated_at = NOW()
	`
	db.WithContext(ctx).Exec(sql, tenantID, viewName, refreshTimeMs, rowCount)
}

// GetCacheStatus returns status of all cached views
func (c *CacheService) GetCacheStatus(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
) []CacheMetadata {
	var metadata []CacheMetadata

	db.WithContext(ctx).
		Table("analytics_cache_metadata").
		Where("tenant_id = ?", tenantID).
		Find(&metadata)

	return metadata
}

// GetLastRefreshTime returns last refresh time for a view
func (c *CacheService) GetLastRefreshTime(
	ctx context.Context,
	db *gorm.DB,
	tenantID, viewName string,
) *time.Time {
	var metadata CacheMetadata

	err := db.WithContext(ctx).
		Table("analytics_cache_metadata").
		Where("tenant_id = ? AND view_name = ?", tenantID, viewName).
		First(&metadata).Error

	if err != nil {
		return nil
	}

	return &metadata.LastRefresh
}

// IsCacheStale checks if cache is stale (older than threshold)
func (c *CacheService) IsCacheStale(
	ctx context.Context,
	db *gorm.DB,
	tenantID, viewName string,
	maxAgeMinutes int,
) bool {
	lastRefresh := c.GetLastRefreshTime(ctx, db, tenantID, viewName)
	if lastRefresh == nil {
		return true
	}

	staleThreshold := time.Now().Add(-time.Duration(maxAgeMinutes) * time.Minute)
	return lastRefresh.Before(staleThreshold)
}

// AutoRefreshIfStale refreshes MV if stale
func (c *CacheService) AutoRefreshIfStale(
	ctx context.Context,
	db *gorm.DB,
	tenantID, viewName string,
	maxAgeMinutes int,
) RefreshStatus {
	if c.IsCacheStale(ctx, db, tenantID, viewName, maxAgeMinutes) {
		return c.RefreshMV(ctx, db, tenantID, viewName)
	}

	lastRefresh := c.GetLastRefreshTime(ctx, db, tenantID, viewName)
	refreshTime := time.Time{}
	if lastRefresh != nil {
		refreshTime = *lastRefresh
	}

	return RefreshStatus{
		ViewName:      viewName,
		LastRefreshed: refreshTime,
		Status:        "CACHED",
	}
}
