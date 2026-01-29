package services

import (
	"context"

	"github.com/omni/backend/internal/utils"
	"gorm.io/gorm"
)

// ReportService handles report generation
type ReportService struct {
	db *gorm.DB
}

// NewReportService creates a new report service
func NewReportService(db *gorm.DB) *ReportService {
	return &ReportService{db: db}
}

// AdsReportData represents ads report data
type AdsReportData struct {
	Summary   AdsReportSummary `json:"summary"`
	Daily     []DailyAdsReport `json:"daily"`
	Campaigns []CampaignReport `json:"campaigns"`
	DateRange DateRange        `json:"date_range"`
}

// AdsReportSummary represents ads report summary
type AdsReportSummary struct {
	TotalSpend       float64 `json:"total_spend"`
	TotalImpressions int64   `json:"total_impressions"`
	TotalClicks      int64   `json:"total_clicks"`
	TotalConversions int64   `json:"total_conversions"`
	TotalRevenue     float64 `json:"total_revenue"`
	CTR              float64 `json:"ctr"`
	CPC              float64 `json:"cpc"`
	ROAS             float64 `json:"roas"`
	ConversionRate   float64 `json:"conversion_rate"`
}

// DailyAdsReport represents daily ads report
type DailyAdsReport struct {
	Date        string  `json:"date"`
	Spend       float64 `json:"spend"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	Conversions int64   `json:"conversions"`
	Revenue     float64 `json:"revenue"`
	CTR         float64 `json:"ctr"`
	CPC         float64 `json:"cpc"`
	ROAS        float64 `json:"roas"`
}

// CampaignReport represents campaign report
type CampaignReport struct {
	CampaignID   string  `json:"campaign_id"`
	CampaignName string  `json:"campaign_name"`
	Status       string  `json:"status"`
	Spend        float64 `json:"spend"`
	Impressions  int64   `json:"impressions"`
	Clicks       int64   `json:"clicks"`
	Conversions  int64   `json:"conversions"`
	Revenue      float64 `json:"revenue"`
	ROAS         float64 `json:"roas"`
}

// DateRange represents date range
type DateRange struct {
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

// GetShopeeAdsReport gets Shopee ads report
func (s *ReportService) GetShopeeAdsReport(ctx context.Context, tenantID, startDate, endDate string) (*AdsReportData, error) {
	// TODO: Implement actual data retrieval from database
	// For now, return mock structure
	return &AdsReportData{
		Summary: AdsReportSummary{
			TotalSpend:       0,
			TotalImpressions: 0,
			TotalClicks:      0,
			TotalConversions: 0,
			TotalRevenue:     0,
		},
		Daily:     []DailyAdsReport{},
		Campaigns: []CampaignReport{},
		DateRange: DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}, nil
}

// GetLatestShopeeAdsReport gets latest Shopee ads report
func (s *ReportService) GetLatestShopeeAdsReport(ctx context.Context, tenantID string) (*AdsReportData, error) {
	now := utils.NowWIB()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -7).Format("2006-01-02")

	return s.GetShopeeAdsReport(ctx, tenantID, startDate, endDate)
}

// GetShopeeAdsReportByDate gets Shopee ads report for specific date
func (s *ReportService) GetShopeeAdsReportByDate(ctx context.Context, tenantID, date string) (*AdsReportData, error) {
	return s.GetShopeeAdsReport(ctx, tenantID, date, date)
}

// GetTiktokAdsReport gets TikTok ads report
func (s *ReportService) GetTiktokAdsReport(ctx context.Context, tenantID, startDate, endDate string) (*AdsReportData, error) {
	// TODO: Implement actual TikTok ads data retrieval
	return &AdsReportData{
		Summary: AdsReportSummary{
			TotalSpend:       0,
			TotalImpressions: 0,
			TotalClicks:      0,
			TotalConversions: 0,
			TotalRevenue:     0,
		},
		Daily:     []DailyAdsReport{},
		Campaigns: []CampaignReport{},
		DateRange: DateRange{
			StartDate: startDate,
			EndDate:   endDate,
		},
	}, nil
}

// GetLatestTiktokAdsReport gets latest TikTok ads report
func (s *ReportService) GetLatestTiktokAdsReport(ctx context.Context, tenantID string) (*AdsReportData, error) {
	now := utils.NowWIB()
	endDate := now.Format("2006-01-02")
	startDate := now.AddDate(0, 0, -7).Format("2006-01-02")

	return s.GetTiktokAdsReport(ctx, tenantID, startDate, endDate)
}
