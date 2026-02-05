package services

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewReportService(t *testing.T) {
	svc := NewReportService(nil)
	assert.NotNil(t, svc)
}

func TestReportService_GetShopeeAdsReport(t *testing.T) {
	svc := NewReportService(nil)

	report, err := svc.GetShopeeAdsReport(context.Background(), "tenant1", "2024-01-01", "2024-01-31")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "2024-01-01", report.DateRange.StartDate)
	assert.Equal(t, "2024-01-31", report.DateRange.EndDate)
	assert.NotNil(t, report.Summary)
	assert.NotNil(t, report.Daily)
	assert.NotNil(t, report.Campaigns)
}

func TestReportService_GetLatestShopeeAdsReport(t *testing.T) {
	svc := NewReportService(nil)

	report, err := svc.GetLatestShopeeAdsReport(context.Background(), "tenant1")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotEmpty(t, report.DateRange.StartDate)
	assert.NotEmpty(t, report.DateRange.EndDate)
}

func TestReportService_GetShopeeAdsReportByDate(t *testing.T) {
	svc := NewReportService(nil)

	report, err := svc.GetShopeeAdsReportByDate(context.Background(), "tenant1", "2024-01-15")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "2024-01-15", report.DateRange.StartDate)
	assert.Equal(t, "2024-01-15", report.DateRange.EndDate)
}

func TestReportService_GetTiktokAdsReport(t *testing.T) {
	svc := NewReportService(nil)

	report, err := svc.GetTiktokAdsReport(context.Background(), "tenant1", "2024-02-01", "2024-02-28")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.Equal(t, "2024-02-01", report.DateRange.StartDate)
	assert.Equal(t, "2024-02-28", report.DateRange.EndDate)
}

func TestReportService_GetLatestTiktokAdsReport(t *testing.T) {
	svc := NewReportService(nil)

	report, err := svc.GetLatestTiktokAdsReport(context.Background(), "tenant1")

	require.NoError(t, err)
	assert.NotNil(t, report)
	assert.NotEmpty(t, report.DateRange.StartDate)
	assert.NotEmpty(t, report.DateRange.EndDate)
}

func TestAdsReportData(t *testing.T) {
	data := AdsReportData{
		Summary: AdsReportSummary{
			TotalSpend:       1000.0,
			TotalImpressions: 50000,
			TotalClicks:      500,
			TotalConversions: 50,
			TotalRevenue:     5000.0,
			CTR:              1.0,
			CPC:              2.0,
			ROAS:             5.0,
			ConversionRate:   10.0,
		},
		Daily: []DailyAdsReport{
			{
				Date:        "2024-01-01",
				Spend:       100.0,
				Impressions: 5000,
				Clicks:      50,
				Conversions: 5,
				Revenue:     500.0,
				CTR:         1.0,
				CPC:         2.0,
				ROAS:        5.0,
			},
		},
		Campaigns: []CampaignReport{
			{
				CampaignID:   "camp-1",
				CampaignName: "Campaign 1",
				Status:       "active",
				Spend:        500.0,
				Impressions:  25000,
				Clicks:       250,
				Conversions:  25,
				Revenue:      2500.0,
				ROAS:         5.0,
			},
		},
		DateRange: DateRange{
			StartDate: "2024-01-01",
			EndDate:   "2024-01-31",
		},
	}

	// Test Summary fields
	assert.Equal(t, 1000.0, data.Summary.TotalSpend)
	assert.Equal(t, int64(50000), data.Summary.TotalImpressions)
	assert.Equal(t, int64(500), data.Summary.TotalClicks)
	assert.Equal(t, int64(50), data.Summary.TotalConversions)
	assert.Equal(t, 5000.0, data.Summary.TotalRevenue)
	assert.Equal(t, 1.0, data.Summary.CTR)
	assert.Equal(t, 2.0, data.Summary.CPC)
	assert.Equal(t, 5.0, data.Summary.ROAS)
	assert.Equal(t, 10.0, data.Summary.ConversionRate)

	// Test Daily reports
	assert.Len(t, data.Daily, 1)
	assert.Equal(t, "2024-01-01", data.Daily[0].Date)

	// Test Campaign reports
	assert.Len(t, data.Campaigns, 1)
	assert.Equal(t, "camp-1", data.Campaigns[0].CampaignID)
	assert.Equal(t, "Campaign 1", data.Campaigns[0].CampaignName)

	// Test DateRange
	assert.Equal(t, "2024-01-01", data.DateRange.StartDate)
	assert.Equal(t, "2024-01-31", data.DateRange.EndDate)
}

func TestAdsReportSummary(t *testing.T) {
	summary := AdsReportSummary{
		TotalSpend:       2500.0,
		TotalImpressions: 100000,
		TotalClicks:      1000,
		TotalConversions: 100,
		TotalRevenue:     10000.0,
		CTR:              1.0,
		CPC:              2.5,
		ROAS:             4.0,
		ConversionRate:   10.0,
	}

	assert.Equal(t, 2500.0, summary.TotalSpend)
	assert.Equal(t, int64(100000), summary.TotalImpressions)
	assert.Equal(t, int64(1000), summary.TotalClicks)
	assert.Equal(t, int64(100), summary.TotalConversions)
	assert.Equal(t, 10000.0, summary.TotalRevenue)
}

func TestDailyAdsReport(t *testing.T) {
	daily := DailyAdsReport{
		Date:        "2024-02-15",
		Spend:       200.0,
		Impressions: 10000,
		Clicks:      100,
		Conversions: 10,
		Revenue:     1000.0,
		CTR:         1.0,
		CPC:         2.0,
		ROAS:        5.0,
	}

	assert.Equal(t, "2024-02-15", daily.Date)
	assert.Equal(t, 200.0, daily.Spend)
	assert.Equal(t, int64(10000), daily.Impressions)
	assert.Equal(t, int64(100), daily.Clicks)
}

func TestCampaignReport(t *testing.T) {
	campaign := CampaignReport{
		CampaignID:   "campaign-123",
		CampaignName: "Summer Sale 2024",
		Status:       "active",
		Spend:        1500.0,
		Impressions:  75000,
		Clicks:       750,
		Conversions:  75,
		Revenue:      7500.0,
		ROAS:         5.0,
	}

	assert.Equal(t, "campaign-123", campaign.CampaignID)
	assert.Equal(t, "Summer Sale 2024", campaign.CampaignName)
	assert.Equal(t, "active", campaign.Status)
	assert.Equal(t, 1500.0, campaign.Spend)
	assert.Equal(t, int64(75000), campaign.Impressions)
}

func TestDateRange(t *testing.T) {
	dateRange := DateRange{
		StartDate: "2024-03-01",
		EndDate:   "2024-03-31",
	}

	assert.Equal(t, "2024-03-01", dateRange.StartDate)
	assert.Equal(t, "2024-03-31", dateRange.EndDate)
}
