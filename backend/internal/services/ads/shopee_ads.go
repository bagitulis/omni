// Package ads provides ads analytics services
package ads

import (
	"regexp"
	"strconv"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ShopeeAdsService handles Shopee ads data operations
type ShopeeAdsService struct {
	db       *gorm.DB
	tenantID string
}

// NewShopeeAdsService creates a new Shopee ads service
func NewShopeeAdsService(db *gorm.DB, tenantID string) *ShopeeAdsService {
	return &ShopeeAdsService{db: db, tenantID: tenantID}
}

// Period represents a time period for ads data
type Period struct {
	Start time.Time
	End   time.Time
	Label string
}

// ParseResult represents CSV parsing result
type ParseResult struct {
	Data    []models.ShopeeAdsProductData
	Period  Period
	BatchID uint
}

// TrendDataPoint represents a data point for charts
type TrendDataPoint struct {
	Label       string  `json:"label"`        // Date or Week Label
	PeriodStart string  `json:"period_start"` // ISO Date
	Spend       float64 `json:"spend"`
	GMV         float64 `json:"gmv"`
	ROAS        float64 `json:"roas"`
	OrderCount  int     `json:"order_count"`
}

// ProductPerformance represents aggregated product performance with scores
type ProductPerformance struct {
	ProductID   string      `json:"product_id"`
	ProductName string      `json:"product_name"`
	ThumbUrl    string      `json:"thumb_url,omitempty"`
	Spend       float64     `json:"spend"`
	GMV         float64     `json:"gmv"`
	ROAS        float64     `json:"roas"`
	Impressions int         `json:"impressions"`
	Clicks      int         `json:"clicks"`
	CTR         float64     `json:"ctr"`
	Conversions int         `json:"conversions"`
	Sold        int         `json:"sold"`
	Score       ScoreResult `json:"score"`
}

// ColumnMap maps CSV headers to struct fields (Indonesian Shopee export format)
var shopeeColumnMap = map[string]string{
	// Product identification
	"Kode Produk": "productId",
	"Nama Iklan":  "productName",
	"Status":      "status",

	// Campaign settings
	"Mode Bidding":     "biddingMode",
	"Penempatan Iklan": "placement",
	"Tanggal Mulai":    "startDate",
	"Tanggal Selesai":  "endDate",

	// Traffic metrics
	"Dilihat":         "impressions",
	"Jumlah Klik":     "clicks",
	"Persentase Klik": "ctr",

	// Conversion metrics
	"Konversi":                  "conversions",
	"Konversi Langsung":         "directConversions",
	"Tingkat konversi":          "conversionRate",
	"Tingkat Konversi Langsung": "directConversionRate",

	// Cost metrics
	"Biaya":                       "cost",
	"Biaya per Konversi":          "costPerConversion",
	"Biaya per Konversi Langsung": "costPerDirectConversion",

	// Sales metrics
	"Produk Terjual":                    "unitsSold",
	"Terjual Langsung":                  "directUnitsSold",
	"Omzet Penjualan":                   "revenue",
	"Penjualan Langsung (GMV Langsung)": "directRevenue",

	// Performance metrics
	"Efektivitas Iklan":    "roas",
	"Efektifitas Iklan":    "roas",
	"Efektivitas Langsung": "directRoas",
	"Efektifitas Langsung": "directRoas",
	"Persentase Biaya Iklan terhadap Penjualan dari Iklan (ACOS)":                   "acos",
	"Persentase Biaya Iklan terhadap Penjualan dari Iklan Langsung (ACOS Langsung)": "directAcos",
}

// ExtractPeriodFromFilename extracts period from Shopee ads filename
func ExtractPeriodFromFilename(filename string) *Period {
	pattern := regexp.MustCompile(`(\d{2})_(\d{2})_(\d{4})-(\d{2})_(\d{2})_(\d{4})`)
	match := pattern.FindStringSubmatch(filename)
	if match == nil {
		return nil
	}

	y1, _ := strconv.Atoi(match[3])
	m1, _ := strconv.Atoi(match[2])
	d1, _ := strconv.Atoi(match[1])
	y2, _ := strconv.Atoi(match[6])
	m2, _ := strconv.Atoi(match[5])
	d2, _ := strconv.Atoi(match[4])

	startDate := time.Date(y1, time.Month(m1), d1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(y2, time.Month(m2), d2, 0, 0, 0, 0, time.UTC)

	_, week := startDate.ISOWeek()
	label := strconv.Itoa(y1) + "-W" + padLeft(week, 2)

	return &Period{
		Start: startDate,
		End:   endDate,
		Label: label,
	}
}
