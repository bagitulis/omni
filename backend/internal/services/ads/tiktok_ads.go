// Package ads provides ads analytics services
package ads

import (
	"fmt"
	"regexp"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// TiktokAdsService handles TikTok ads data operations
type TiktokAdsService struct {
	db       *gorm.DB
	tenantID string
}

// NewTiktokAdsService creates a new TikTok ads service
func NewTiktokAdsService(db *gorm.DB, tenantID string) *TiktokAdsService {
	return &TiktokAdsService{db: db, tenantID: tenantID}
}

// TiktokColumnMap maps Excel headers to struct fields (English + Indonesian TikTok export format)
var tiktokColumnMap = map[string]string{
	// Campaign & Product identification
	"Campaign name": "campaignName",
	"Nama kampanye": "campaignName",
	"Campaign ID":   "campaignId",
	"ID Campaign":   "campaignId",
	"Product ID":    "productId",
	"ID produk":     "productId",

	// Creative info
	"Creative type":      "creativeType",
	"Jenis materi iklan": "creativeType",
	"Video title":        "videoTitle",
	"Judul video":        "videoTitle",
	"Video ID":           "videoId",
	"ID video":           "videoId",

	// Account & Status
	"TikTok account":     "tiktokAccount",
	"Akun TikTok":        "tiktokAccount",
	"Time posted":        "postingTime",
	"Waktu posting":      "postingTime",
	"Status":             "status",
	"Authorization type": "authorizationType",
	"Jenis otorisasi":    "authorizationType",

	// Cost & Revenue
	"Cost":              "cost",
	"Biaya":             "cost",
	"SKU orders":        "ordersSku",
	"Pesanan SKU":       "ordersSku",
	"Cost per order":    "costPerOrder",
	"Biaya per pesanan": "costPerOrder",
	"Gross revenue":     "grossRevenue",
	"Pendapatan kotor":  "grossRevenue",
	"ROI":               "roi",

	// Traffic metrics
	"Product ad impressions":    "impressions",
	"Impresi iklan produk":      "impressions",
	"Product ad clicks":         "clicks",
	"Jumlah klik iklan produk":  "clicks",
	"Product ad click rate":     "ctr",
	"Tingkat klik iklan produk": "ctr",
	"Ad conversion rate":        "conversionRate",
	"Rasio konversi iklan":      "conversionRate",

	// Video watch rates
	"2-second ad video view rate":      "watchRate2s",
	"Rasio tayang video iklan 2 detik": "watchRate2s",
	"6-second ad video view rate":      "watchRate6s",
	"Rasio tayang video iklan 6 detik": "watchRate6s",
	"25% ad video view rate":           "watchRate25pct",
	"Rasio tayang video iklan 25%":     "watchRate25pct",
	"50% ad video view rate":           "watchRate50pct",
	"Rasio tayang video iklan 50%":     "watchRate50pct",
	"75% ad video view rate":           "watchRate75pct",
	"Rasio tayang video iklan 75%":     "watchRate75pct",
	"100% ad video view rate":          "watchRate100pct",
	"Rasio tayang video iklan 100%":    "watchRate100pct",

	// Currency
	"Currency":  "currency",
	"Mata uang": "currency",
}

// TiktokParseResult represents Excel parsing result
type TiktokParseResult struct {
	Data    []models.TiktokAdsCreativeData
	Period  Period
	BatchID uint
}

// ExtractTiktokPeriod extracts period from TikTok ads filename
func ExtractTiktokPeriod(filename string) *Period {
	// Pattern 1: "creative data for product campaigns YYYY-MM-DD HH ~ YYYY-MM-DD HH.xlsx"
	pattern := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})\s+\d+\s*~\s*(\d{4})-(\d{2})-(\d{2})`)
	match := pattern.FindStringSubmatch(filename)
	if match != nil {
		y1, _ := strconv.Atoi(match[1])
		m1, _ := strconv.Atoi(match[2])
		d1, _ := strconv.Atoi(match[3])
		y2, _ := strconv.Atoi(match[4])
		m2, _ := strconv.Atoi(match[5])
		d2, _ := strconv.Atoi(match[6])

		startDate := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

		// Use 4-bin period label based on day of month
		label := generateTiktokPeriodLabel(startDate)
		return &Period{Start: startDate, End: endDate, Label: label}
	}

	// Pattern 2: "report_YYYY-MM-DD_YYYY-MM-DD.xlsx"
	pattern2 := regexp.MustCompile(`(\d{4})-(\d{2})-(\d{2})_(\d{4})-(\d{2})-(\d{2})`)
	match = pattern2.FindStringSubmatch(filename)
	if match != nil {
		y1, _ := strconv.Atoi(match[1])
		m1, _ := strconv.Atoi(match[2])
		d1, _ := strconv.Atoi(match[3])
		y2, _ := strconv.Atoi(match[4])
		m2, _ := strconv.Atoi(match[5])
		d2, _ := strconv.Atoi(match[6])

		startDate := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

		label := generateTiktokPeriodLabel(startDate)
		return &Period{Start: startDate, End: endDate, Label: label}
	}

	// Pattern 3: "DD-MM-YYYY_DD-MM-YYYY"
	pattern3 := regexp.MustCompile(`(\d{2})-(\d{2})-(\d{4})_(\d{2})-(\d{2})-(\d{4})`)
	match = pattern3.FindStringSubmatch(filename)
	if match != nil {
		y1, _ := strconv.Atoi(match[3])
		m1, _ := strconv.Atoi(match[2])
		d1, _ := strconv.Atoi(match[1])
		y2, _ := strconv.Atoi(match[6])
		m2, _ := strconv.Atoi(match[5])
		d2, _ := strconv.Atoi(match[4])

		startDate := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

		label := generateTiktokPeriodLabel(startDate)
		return &Period{Start: startDate, End: endDate, Label: label}
	}

	return nil
}

// generateTiktokPeriodLabel generates 4-bin period label based on day of month
// Bin 1: Early (1-7) - Just after payday
// Bin 2: Mid-I (8-15) - First half middle
// Bin 3: Mid-II (16-23) - Second half middle
// Bin 4: Late (24-31) - Waiting for payday
func generateTiktokPeriodLabel(date time.Time) string {
	year := date.Year()
	month := int(date.Month())
	day := date.Day()

	var bin string
	switch {
	case day <= 7:
		bin = "A" // Early
	case day <= 15:
		bin = "B" // Mid-I
	case day <= 23:
		bin = "C" // Mid-II
	default:
		bin = "D" // Late
	}

	return fmt.Sprintf("%d-%02d-%s", year, month, bin)
}

// parseTiktokRow parses a single row from TikTok Excel
func parseTiktokRow(row []string, colIdx map[string]int, period *Period, tenantID string) models.TiktokAdsCreativeData {
	return models.TiktokAdsCreativeData{
		TenantID:          tenantID,
		CampaignID:        getStrVal(row, colIdx, "campaignId"),
		CampaignName:      getStrVal(row, colIdx, "campaignName"),
		ProductID:         getStrVal(row, colIdx, "productId"),
		CreativeType:      getStrVal(row, colIdx, "creativeType"),
		VideoTitle:        getStrVal(row, colIdx, "videoTitle"),
		VideoID:           getStrVal(row, colIdx, "videoId"),
		TiktokAccount:     getStrVal(row, colIdx, "tiktokAccount"),
		Status:            getStrVal(row, colIdx, "status"),
		AuthorizationType: getStrVal(row, colIdx, "authorizationType"),
		Cost:              getFloatVal(row, colIdx, "cost"),
		OrdersSKU:         getIntVal(row, colIdx, "ordersSku"),
		CostPerOrder:      getFloatVal(row, colIdx, "costPerOrder"),
		GrossRevenue:      getFloatVal(row, colIdx, "grossRevenue"),
		ROI:               getFloatVal(row, colIdx, "roi"),
		Impressions:       getIntVal(row, colIdx, "impressions"),
		Clicks:            getIntVal(row, colIdx, "clicks"),
		CTR:               getFloatVal(row, colIdx, "ctr"),
		ConversionRate:    getFloatVal(row, colIdx, "conversionRate"),
		WatchRate2s:       getFloatVal(row, colIdx, "watchRate2s"),
		WatchRate6s:       getFloatVal(row, colIdx, "watchRate6s"),
		WatchRate25Pct:    getFloatVal(row, colIdx, "watchRate25pct"),
		WatchRate50Pct:    getFloatVal(row, colIdx, "watchRate50pct"),
		WatchRate75Pct:    getFloatVal(row, colIdx, "watchRate75pct"),
		WatchRate100Pct:   getFloatVal(row, colIdx, "watchRate100pct"),
		Currency:          getStrVal(row, colIdx, "currency"),
		PeriodLabel:       period.Label,
		PeriodStart:       period.Start,
		PeriodEnd:         period.End,
	}
}
