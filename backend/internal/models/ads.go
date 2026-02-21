package models

import "time"

// ============================================
// SHOPEE ADS MODELS
// Matches PostgreSQL schema exactly
// ============================================

// ShopeeAdsUploadBatch represents a batch of uploaded Shopee ads data
type ShopeeAdsUploadBatch struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	FileName     string    `gorm:"column:file_name;type:varchar(500)" json:"file_name"`
	PeriodStart  time.Time `gorm:"column:period_start;type:timestamptz" json:"period_start"`
	PeriodEnd    time.Time `gorm:"column:period_end;type:timestamptz" json:"period_end"`
	PeriodLabel  string    `gorm:"column:period_label;type:varchar(20);index" json:"period_label"`
	TotalRows    int       `gorm:"column:total_rows" json:"total_rows"`
	InsertedRows int       `gorm:"column:inserted_rows" json:"inserted_rows"`
	SkippedRows  int       `gorm:"column:skipped_rows" json:"skipped_rows"`
	UpdatedRows  int       `gorm:"column:updated_rows" json:"updated_rows"`
	Status       string    `gorm:"column:status;type:varchar(100)" json:"status"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`
	UploadedBy   string    `gorm:"column:uploaded_by;type:varchar(255)" json:"uploaded_by"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz" json:"updated_at"`
}

func (ShopeeAdsUploadBatch) TableName() string { return GetTableName("ShopeeAdsUploadBatch") }

// ShopeeAdsProductData represents Shopee ads product-level data
type ShopeeAdsProductData struct {
	ID                      int       `gorm:"column:id;primaryKey" json:"id"`
	TenantID                string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	UploadBatchID           string    `gorm:"column:upload_batch_id;type:varchar(255);index" json:"upload_batch_id"`
	PeriodStart             time.Time `gorm:"column:period_start;type:timestamptz" json:"period_start"`
	PeriodEnd               time.Time `gorm:"column:period_end;type:timestamptz" json:"period_end"`
	PeriodLabel             string    `gorm:"column:period_label;type:varchar(100);index" json:"period_label"`
	ProductID               string    `gorm:"column:product_id;type:varchar(255);index" json:"product_id"`
	ProductName             string    `gorm:"column:product_name;type:text" json:"product_name"`
	Status                  string    `gorm:"column:status;type:varchar(100)" json:"status"`
	BiddingMode             string    `gorm:"column:bidding_mode;type:varchar(255)" json:"bidding_mode"`
	Placement               string    `gorm:"column:placement;type:varchar(255)" json:"placement"`
	StartDate               time.Time `gorm:"column:start_date;type:timestamptz" json:"start_date"`
	EndDate                 string    `gorm:"column:end_date;type:varchar(100)" json:"end_date"`
	Impressions             int       `gorm:"column:impressions;type:integer" json:"impressions"`
	Clicks                  int       `gorm:"column:clicks;type:integer" json:"clicks"`
	CTR                     float64   `gorm:"column:ctr" json:"ctr"`
	Conversions             int       `gorm:"column:conversions" json:"conversions"`
	DirectConversions       int       `gorm:"column:direct_conversions" json:"direct_conversions"`
	ConversionRate          float64   `gorm:"column:conversion_rate" json:"conversion_rate"`
	DirectConversionRate    float64   `gorm:"column:direct_conversion_rate" json:"direct_conversion_rate"`
	CostPerConversion       float64   `gorm:"column:cost_per_conversion" json:"cost_per_conversion"`
	CostPerDirectConversion float64   `gorm:"column:cost_per_direct_conversion" json:"cost_per_direct_conversion"`
	UnitsSold               int       `gorm:"column:units_sold" json:"units_sold"`
	DirectUnitsSold         int       `gorm:"column:direct_units_sold" json:"direct_units_sold"`
	Revenue                 float64   `gorm:"column:revenue" json:"revenue"`
	DirectRevenue           float64   `gorm:"column:direct_revenue" json:"direct_revenue"`
	Cost                    float64   `gorm:"column:cost" json:"cost"`
	ROAS                    float64   `gorm:"column:roas" json:"roas"`
	DirectROAS              float64   `gorm:"column:direct_roas" json:"direct_roas"`
	ACOS                    float64   `gorm:"column:acos" json:"acos"`
	DirectACOS              float64   `gorm:"column:direct_acos" json:"direct_acos"`
	CreatedAt               time.Time `gorm:"column:created_at;type:timestamptz" json:"created_at"`
	UpdatedAt               time.Time `gorm:"column:updated_at;type:timestamptz" json:"updated_at"`
}

func (ShopeeAdsProductData) TableName() string { return GetTableName("ShopeeAdsProductData") }

// ============================================
// TIKTOK ADS MODELS
// Matches PostgreSQL schema exactly
// ============================================

// TiktokAdsUploadBatch represents a batch of uploaded TikTok ads data
type TiktokAdsUploadBatch struct {
	ID           string    `gorm:"column:id;primaryKey;type:varchar(255)" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	FileName     string    `gorm:"column:file_name;type:varchar(500)" json:"file_name"`
	PeriodStart  time.Time `gorm:"column:period_start;type:timestamptz" json:"period_start"`
	PeriodEnd    time.Time `gorm:"column:period_end;type:timestamptz" json:"period_end"`
	PeriodLabel  string    `gorm:"column:period_label;type:varchar(20);index" json:"period_label"`
	TotalRows    int       `gorm:"column:total_rows" json:"total_rows"`
	InsertedRows int       `gorm:"column:inserted_rows" json:"inserted_rows"`
	SkippedRows  int       `gorm:"column:skipped_rows" json:"skipped_rows"`
	UpdatedRows  int       `gorm:"column:updated_rows" json:"updated_rows"`
	Status       string    `gorm:"column:status;type:varchar(100)" json:"status"`
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message"`
	UploadedBy   string    `gorm:"column:uploaded_by;type:varchar(255)" json:"uploaded_by"`
	CreatedAt    time.Time `gorm:"column:created_at;type:timestamptz" json:"created_at"`
	UpdatedAt    time.Time `gorm:"column:updated_at;type:timestamptz" json:"updated_at"`
}

func (TiktokAdsUploadBatch) TableName() string { return GetTableName("TiktokAdsUploadBatch") }

// TiktokAdsCreativeData represents TikTok ads creative-level data
type TiktokAdsCreativeData struct {
	ID                int       `gorm:"column:id;primaryKey" json:"id"`
	TenantID          string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	UploadBatchID     string    `gorm:"column:upload_batch_id;type:varchar(255);index" json:"upload_batch_id"`
	PeriodStart       time.Time `gorm:"column:period_start;type:timestamptz" json:"period_start"`
	PeriodEnd         time.Time `gorm:"column:period_end;type:timestamptz" json:"period_end"`
	PeriodLabel       string    `gorm:"column:period_label;type:varchar(20);index" json:"period_label"`
	CampaignID        string    `gorm:"column:campaign_id;type:varchar(255)" json:"campaign_id"`
	CampaignName      string    `gorm:"column:campaign_name;type:text" json:"campaign_name"`
	ProductID         string    `gorm:"column:product_id;type:varchar(255);index" json:"product_id"`
	CreativeType      string    `gorm:"column:creative_type;type:varchar(255)" json:"creative_type"`
	VideoTitle        string    `gorm:"column:video_title;type:text" json:"video_title"`
	VideoID           string    `gorm:"column:video_id;type:varchar(255)" json:"video_id"`
	TiktokAccount     string    `gorm:"column:tiktok_account;type:varchar(500)" json:"tiktok_account"`
	PostingTime       time.Time `gorm:"column:posting_time;type:timestamptz" json:"posting_time"`
	Status            string    `gorm:"column:status;type:varchar(100)" json:"status"`
	AuthorizationType string    `gorm:"column:authorization_type;type:varchar(255)" json:"authorization_type"`
	Cost              float64   `gorm:"column:cost;type:double precision" json:"cost"`
	OrdersSKU         int       `gorm:"column:orders_sku;type:integer" json:"orders_sku"`
	CostPerOrder      float64   `gorm:"column:cost_per_order;type:double precision" json:"cost_per_order"`
	GrossRevenue      float64   `gorm:"column:gross_revenue;type:double precision" json:"gross_revenue"`
	ROI               float64   `gorm:"column:roi;type:double precision" json:"roi"`
	Impressions       int       `gorm:"column:impressions;type:integer" json:"impressions"`
	Clicks            int       `gorm:"column:clicks;type:integer" json:"clicks"`
	CTR               float64   `gorm:"column:ctr;type:double precision" json:"ctr"`
	ConversionRate    float64   `gorm:"column:conversion_rate;type:double precision" json:"conversion_rate"`
	WatchRate2s       float64   `gorm:"column:watch_rate_2s;type:double precision" json:"watch_rate_2s"`
	WatchRate6s       float64   `gorm:"column:watch_rate_6s;type:double precision" json:"watch_rate_6s"`
	WatchRate25Pct    float64   `gorm:"column:watch_rate_25pct;type:double precision" json:"watch_rate_25pct"`
	WatchRate50Pct    float64   `gorm:"column:watch_rate_50pct;type:double precision" json:"watch_rate_50pct"`
	WatchRate75Pct    float64   `gorm:"column:watch_rate_75pct;type:double precision" json:"watch_rate_75pct"`
	WatchRate100Pct   float64   `gorm:"column:watch_rate_100pct;type:double precision" json:"watch_rate_100pct"`
	Currency          string    `gorm:"column:currency;type:varchar(20)" json:"currency"`
	CreatedAt         time.Time `gorm:"column:created_at;type:timestamptz" json:"created_at"`
	UpdatedAt         time.Time `gorm:"column:updated_at;type:timestamptz" json:"updated_at"`
}

func (TiktokAdsCreativeData) TableName() string { return GetTableName("TiktokAdsCreativeData") }

// TiktokAdsProductSummary represents product-level summary from TikTok ads
type TiktokAdsProductSummary struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	ProductID      string    `gorm:"column:product_id;type:varchar(255);index" json:"product_id"`
	ProductName    string    `gorm:"column:product_name;type:text" json:"product_name"`
	TotalCost      float64   `gorm:"column:total_cost" json:"total_cost"`
	TotalRevenue   float64   `gorm:"column:total_revenue" json:"total_revenue"`
	TotalConv      int       `gorm:"column:total_conv" json:"total_conversions"`
	AvgROI         float64   `gorm:"column:avg_roi" json:"avg_roi"`
	AvgCPA         float64   `gorm:"column:avg_cpa" json:"avg_cpa"`
	TotalCreatives int       `gorm:"column:total_creatives" json:"total_creatives"`
	PeriodLabel    string    `gorm:"column:period_label;type:varchar(20);index" json:"period_label"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamptz" json:"created_at"`
}

func (TiktokAdsProductSummary) TableName() string { return GetTableName("TiktokAdsProductSummary") }

// TiktokAdsMLPrediction represents ML predictions for TikTok ads
type TiktokAdsMLPrediction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"column:tenant_id;type:varchar(255);index;not null" json:"tenant_id"`
	ProductID       string    `gorm:"column:product_id;type:varchar(255);index" json:"product_id"`
	PredictedROI    float64   `gorm:"column:predicted_roi" json:"predicted_roi"`
	PredictedConv   int       `gorm:"column:predicted_conv" json:"predicted_conversions"`
	Confidence      float64   `gorm:"column:confidence" json:"confidence"`
	RecommendBudget float64   `gorm:"column:recommend_budget" json:"recommended_budget"`
	PredictedAt     time.Time `gorm:"column:predicted_at;type:timestamptz" json:"predicted_at"`
	ValidUntil      time.Time `gorm:"column:valid_until;type:timestamptz" json:"valid_until"`
	CreatedAt       time.Time `gorm:"column:created_at;type:timestamptz" json:"created_at"`
}

func (TiktokAdsMLPrediction) TableName() string { return GetTableName("TiktokAdsMLPrediction") }
