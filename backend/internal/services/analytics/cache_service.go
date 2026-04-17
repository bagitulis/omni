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

// RefreshMV refreshes a single materialized view.
// Uses a background context so user request cancellation doesn't abort the refresh.
func (c *CacheService) RefreshMV(
	ctx context.Context,
	db *gorm.DB,
	tenantID string,
	viewName string,
) RefreshStatus {
	startTime := time.Now()

	// Use background context with timeout — user cancel must NOT kill MV refresh
	bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Check if MV has been populated (unpopulated MVs fail with CONCURRENTLY)
	populated := c.isMVPopulated(bgCtx, db, viewName)

	var err error
	if populated {
		// Try concurrent refresh first (non-blocking)
		sql := "REFRESH MATERIALIZED VIEW CONCURRENTLY " + viewName
		err = db.WithContext(bgCtx).Exec(sql).Error
	}

	// Fallback: non-concurrent refresh (blocks reads but always works)
	if !populated || err != nil {
		sql := "REFRESH MATERIALIZED VIEW " + viewName
		err = db.WithContext(bgCtx).Exec(sql).Error

		if err != nil {
			return RefreshStatus{
				ViewName: viewName,
				Status:   "ERROR: " + err.Error(),
			}
		}
	}

	refreshTime := time.Since(startTime).Milliseconds()

	// Get row count
	var rowCount int64
	db.WithContext(bgCtx).Table(viewName).
		Where("tenant_id = ?", tenantID).
		Count(&rowCount)

	// Update metadata
	c.updateMetadata(bgCtx, db, tenantID, viewName, refreshTime, rowCount)

	return RefreshStatus{
		ViewName:      viewName,
		LastRefreshed: time.Now(),
		RowCount:      rowCount,
		Status:        "SUCCESS",
	}
}

// updateMetadata updates cache metadata table.
// tenantID unused — tenant isolation is at the schema level, not row level.
func (c *CacheService) updateMetadata(
	ctx context.Context,
	db *gorm.DB,
	_, viewName string,
	refreshTimeMs int64,
	rowCount int64,
) {
	sql := `
		INSERT INTO analytics_cache_metadata 
		(view_name, last_refresh, refresh_time_ms, row_count, updated_at)
		VALUES (?, NOW(), ?, ?, NOW())
		ON CONFLICT (view_name) 
		DO UPDATE SET 
			last_refresh = NOW(),
			refresh_time_ms = EXCLUDED.refresh_time_ms,
			row_count = EXCLUDED.row_count,
			updated_at = NOW()
	`
	db.WithContext(ctx).Exec(sql, viewName, refreshTimeMs, rowCount)
}

// GetCacheStatus returns status of all cached views.
// No tenant_id filter needed — table is per-tenant schema.
func (c *CacheService) GetCacheStatus(
	ctx context.Context,
	db *gorm.DB,
	_ string,
) []CacheMetadata {
	var metadata []CacheMetadata

	db.WithContext(ctx).
		Table("analytics_cache_metadata").
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
		Where("view_name = ?", viewName).
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

// isMVPopulated checks if a materialized view has been populated at least once.
// Unpopulated MVs cannot use REFRESH CONCURRENTLY.
func (c *CacheService) isMVPopulated(ctx context.Context, db *gorm.DB, viewName string) bool {
	var populated bool
	err := db.WithContext(ctx).Raw(
		"SELECT relispopulated FROM pg_class WHERE relname = ?", viewName,
	).Scan(&populated).Error
	if err != nil {
		return false
	}
	return populated
}
