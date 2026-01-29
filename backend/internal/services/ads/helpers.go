// Package ads provides ads analytics services - helper functions
package ads

import (
	"bufio"
	"bytes"
	"strconv"
	"strings"

	"github.com/omni/backend/internal/models"
)

// splitLines splits byte data into lines
func splitLines(data []byte) []string {
	var lines []string
	scanner := bufio.NewScanner(bytes.NewReader(data))
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines
}

// buildColumnIndex creates index map from headers
func buildColumnIndex(headers []string, mapping map[string]string) map[string]int {
	idx := make(map[string]int)
	for i, h := range headers {
		h = strings.TrimSpace(h)
		if field, ok := mapping[h]; ok {
			idx[field] = i
		}
	}
	return idx
}

// parseShopeeRow converts a CSV row to ShopeeAdsProductData
func parseShopeeRow(row []string, colIdx map[string]int, period *Period, tenantID string) models.ShopeeAdsProductData {
	return models.ShopeeAdsProductData{
		TenantID:                tenantID,
		ProductID:               getStrVal(row, colIdx, "productId"),
		ProductName:             getStrVal(row, colIdx, "productName"),
		Status:                  getStrVal(row, colIdx, "status"),
		BiddingMode:             getStrVal(row, colIdx, "biddingMode"),
		Placement:               getStrVal(row, colIdx, "placement"),
		Impressions:             getIntVal(row, colIdx, "impressions"),
		Clicks:                  getIntVal(row, colIdx, "clicks"),
		CTR:                     getFloatVal(row, colIdx, "ctr"),
		Conversions:             getIntVal(row, colIdx, "conversions"),
		DirectConversions:       getIntVal(row, colIdx, "directConversions"),
		ConversionRate:          getFloatVal(row, colIdx, "conversionRate"),
		DirectConversionRate:    getFloatVal(row, colIdx, "directConversionRate"),
		CostPerConversion:       getFloatVal(row, colIdx, "costPerConversion"),
		CostPerDirectConversion: getFloatVal(row, colIdx, "costPerDirectConversion"),
		UnitsSold:               getIntVal(row, colIdx, "unitsSold"),
		DirectUnitsSold:         getIntVal(row, colIdx, "directUnitsSold"),
		Revenue:                 getFloatVal(row, colIdx, "revenue"),
		DirectRevenue:           getFloatVal(row, colIdx, "directRevenue"),
		Cost:                    getFloatVal(row, colIdx, "cost"),
		ROAS:                    getFloatVal(row, colIdx, "roas"),
		DirectROAS:              getFloatVal(row, colIdx, "directRoas"),
		ACOS:                    getFloatVal(row, colIdx, "acos"),
		DirectACOS:              getFloatVal(row, colIdx, "directAcos"),
		PeriodLabel:             period.Label,
		PeriodStart:             period.Start,
		PeriodEnd:               period.End,
	}
}

// getStrVal gets string value from row by field name
func getStrVal(row []string, idx map[string]int, field string) string {
	if i, ok := idx[field]; ok && i < len(row) {
		return strings.TrimSpace(row[i])
	}
	return ""
}

// getIntVal gets integer value from row by field name
func getIntVal(row []string, idx map[string]int, field string) int {
	s := getStrVal(row, idx, field)
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	v, _ := strconv.Atoi(s)
	return v
}

// getFloatVal gets float value from row by field name
func getFloatVal(row []string, idx map[string]int, field string) float64 {
	s := getStrVal(row, idx, field)
	s = strings.ReplaceAll(s, ",", ".")
	s = strings.ReplaceAll(s, "%", "")
	s = strings.ReplaceAll(s, "Rp", "")
	s = strings.TrimSpace(s)
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// padLeft pads number with zeros to specified width
func padLeft(n, width int) string {
	s := strconv.Itoa(n)
	for len(s) < width {
		s = "0" + s
	}
	return s
}

// AdsSummary represents aggregated ads summary
type AdsSummary struct {
	TotalCost        float64 `json:"total_cost"`
	TotalRevenue     float64 `json:"total_revenue"`
	TotalClicks      int     `json:"total_clicks"`
	TotalImpressions int     `json:"total_impressions"`
	TotalConversions int     `json:"total_conversions"`
	AvgROAS          float64 `json:"avg_roas"`
	AvgACOS          float64 `json:"avg_acos"`
	AvgCTR           float64 `json:"avg_ctr"`
	ProductCount     int     `json:"product_count"`
}

// calculateShopeeAdsSummary calculates summary from ads data
func calculateShopeeAdsSummary(data []models.ShopeeAdsProductData) *AdsSummary {
	sum := &AdsSummary{ProductCount: len(data)}
	for _, d := range data {
		sum.TotalCost += d.Cost
		sum.TotalRevenue += d.Revenue
		sum.TotalClicks += d.Clicks
		sum.TotalImpressions += d.Impressions
		sum.TotalConversions += d.Conversions
	}
	if sum.TotalCost > 0 {
		sum.AvgROAS = sum.TotalRevenue / sum.TotalCost
	}
	if sum.TotalRevenue > 0 {
		sum.AvgACOS = (sum.TotalCost / sum.TotalRevenue) * 100
	}
	if sum.TotalImpressions > 0 {
		sum.AvgCTR = (float64(sum.TotalClicks) / float64(sum.TotalImpressions)) * 100
	}
	return sum
}

// ParseError represents a parsing error
type ParseError struct {
	Message string
}

// Error implements error interface
func (e *ParseError) Error() string {
	return e.Message
}
