package services

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)

// AnalyticsService handles analytics operations
type AnalyticsService struct {
	repo *repositories.AnalyticsRepository
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(repo *repositories.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

// GetSettings gets analytics settings
func (s *AnalyticsService) GetSettings(ctx context.Context, tenantID, platform string) (*models.AnalyticsSettings, error) {
	return s.repo.GetSettings(ctx, tenantID, platform)
}

// UpdateSettings updates analytics settings
func (s *AnalyticsService) UpdateSettings(ctx context.Context, settings *models.AnalyticsSettings) error {
	return s.repo.CreateOrUpdateSettings(ctx, settings)
}

// GetDashboardSummary gets dashboard summary data
func (s *AnalyticsService) GetDashboardSummary(ctx context.Context, tenantID string, startDate, endDate time.Time) (*models.SalesAnalytics, error) {
	orderCounts, err := s.repo.GetOrderCountByPlatform(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	salesByPlatform, err := s.repo.GetTotalSalesByPlatform(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	totalOrders := 0
	totalSales := 0.0

	for _, count := range orderCounts {
		totalOrders += count
	}

	for _, sales := range salesByPlatform {
		totalSales += sales
	}

	return &models.SalesAnalytics{
		TotalSales:   totalSales,
		TotalOrders:  totalOrders,
		AverageOrder: calculateAOV(totalSales, totalOrders),
		ByPlatform:   salesByPlatform,
		PeriodStart:  startDate,
		PeriodEnd:    endDate,
	}, nil
}

// GetOrderAnalytics gets order analytics
func (s *AnalyticsService) GetOrderAnalytics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*models.OrderAnalytics, error) {
	orderCounts, err := s.repo.GetOrderCountByPlatform(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	totalOrders := 0
	for _, count := range orderCounts {
		totalOrders += count
	}

	return &models.OrderAnalytics{
		TotalOrders:     totalOrders,
		ByPlatform:      orderCounts,
		ByStatus:        make(map[string]int),
		FulfillmentRate: 100.0, // Placeholder
		ReturnRate:      0,
		AverageValue:    0,
	}, nil
}

// GetRevenueAnalytics gets revenue analytics
func (s *AnalyticsService) GetRevenueAnalytics(ctx context.Context, tenantID string, startDate, endDate time.Time) (*models.RevenueAnalytics, error) {
	salesByPlatform, err := s.repo.GetTotalSalesByPlatform(ctx, tenantID, startDate, endDate)
	if err != nil {
		return nil, err
	}

	totalRevenue := 0.0
	for _, sales := range salesByPlatform {
		totalRevenue += sales
	}

	return &models.RevenueAnalytics{
		GrossRevenue:   totalRevenue,
		NetRevenue:     totalRevenue * 0.85, // Placeholder 15% fees
		TotalFees:      totalRevenue * 0.15,
		TotalDiscounts: 0,
		ByPlatform:     salesByPlatform,
		ByMonth:        make(map[string]float64),
		PeriodStart:    startDate,
		PeriodEnd:      endDate,
	}, nil
}

// GetEscrowSyncStatus gets escrow sync status
func (s *AnalyticsService) GetEscrowSyncStatus(ctx context.Context, tenantID string, month, year int) (*EscrowSyncStatus, error) {
	shopeeSync, err := s.repo.GetShopeeEscrowSync(ctx, tenantID, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get Shopee escrow sync status: %w", err)
	}

	tiktokSync, err := s.repo.GetTiktokEscrowSync(ctx, tenantID, month, year)
	if err != nil {
		return nil, fmt.Errorf("failed to get TikTok escrow sync status: %w", err)
	}

	return &EscrowSyncStatus{
		Month:      month,
		Year:       year,
		ShopeeSync: shopeeSync,
		TiktokSync: tiktokSync,
	}, nil
}

// EscrowSyncStatus represents escrow sync status
type EscrowSyncStatus struct {
	Month      int                      `json:"month"`
	Year       int                      `json:"year"`
	ShopeeSync *models.ShopeeEscrowSync `json:"shopee_sync"`
	TiktokSync *models.TiktokEscrowSync `json:"tiktok_sync"`
}

// UpdateShopeeEscrowSync updates Shopee escrow sync status
func (s *AnalyticsService) UpdateShopeeEscrowSync(ctx context.Context, tenantID string, month, year, totalOrders int) error {
	return s.repo.CreateOrUpdateShopeeEscrowSync(ctx, tenantID, month, year, totalOrders)
}

// calculateAOV calculates average order value
func calculateAOV(totalSales float64, totalOrders int) float64 {
	if totalOrders == 0 {
		return 0
	}
	return totalSales / float64(totalOrders)
}
