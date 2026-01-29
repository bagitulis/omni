package sheets

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// ShippingFeeRecord represents a shipping fee record
type ShippingFeeRecord struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"index;not null" json:"tenant_id"`
	Platform       string    `gorm:"index;not null" json:"platform"`
	OrderSN        string    `gorm:"index" json:"order_sn"`
	TrackingNumber string    `json:"tracking_number,omitempty"`
	Courier        string    `json:"courier"`
	ShippingMethod string    `json:"shipping_method"`
	Weight         float64   `json:"weight"`
	WeightUnit     string    `json:"weight_unit"`
	BaseFee        float64   `json:"base_fee"`
	WeightFee      float64   `json:"weight_fee"`
	InsuranceFee   float64   `json:"insurance_fee"`
	CodFee         float64   `json:"cod_fee"`
	OtherFees      float64   `json:"other_fees"`
	Discount       float64   `json:"discount"`
	TotalFee       float64   `json:"total_fee"`
	Currency       string    `json:"currency"`
	PaidBy         string    `json:"paid_by"` // seller, buyer, platform
	Status         string    `json:"status"`
	ShipDate       time.Time `json:"ship_date,omitempty"`
	DeliverDate    time.Time `json:"deliver_date,omitempty"`
	Source         string    `json:"source"` // api, sheet
	SheetConfigID  uint      `json:"sheet_config_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// TableName returns the table name for GORM
func (ShippingFeeRecord) TableName() string {
	return "shipping_fee_records"
}

// ShippingFeeSummary represents shipping fee summary
type ShippingFeeSummary struct {
	TotalOrders    int                    `json:"total_orders"`
	TotalFees      float64                `json:"total_fees"`
	TotalDiscounts float64                `json:"total_discounts"`
	NetFees        float64                `json:"net_fees"`
	AvgFeePerOrder float64                `json:"avg_fee_per_order"`
	ByCourier      map[string]CourierStat `json:"by_courier"`
	ByPlatform     map[string]float64     `json:"by_platform"`
}

// CourierStat represents statistics for a courier
type CourierStat struct {
	OrderCount int     `json:"order_count"`
	TotalFees  float64 `json:"total_fees"`
	AvgFee     float64 `json:"avg_fee"`
}

// ShippingFeeSheetService handles shipping fee sheet operations
type ShippingFeeSheetService struct {
	db *gorm.DB
}

// NewShippingFeeSheetService creates a new shipping fee sheet service
func NewShippingFeeSheetService(db *gorm.DB) *ShippingFeeSheetService {
	return &ShippingFeeSheetService{db: db}
}

// GetShippingFees retrieves shipping fee records
func (s *ShippingFeeSheetService) GetShippingFees(
	ctx context.Context,
	tenantID string,
	platform string,
	limit int,
	offset int,
) ([]ShippingFeeRecord, int64, error) {
	var records []ShippingFeeRecord
	var total int64

	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	if err := query.Model(&ShippingFeeRecord{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("created_at DESC").Find(&records).Error
	return records, total, err
}

// GetShippingFeeByOrder retrieves shipping fee for an order
func (s *ShippingFeeSheetService) GetShippingFeeByOrder(
	ctx context.Context,
	tenantID string,
	orderSN string,
) (*ShippingFeeRecord, error) {
	var record ShippingFeeRecord

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND order_sn = ?", tenantID, orderSN).
		First(&record).Error

	if err != nil {
		return nil, err
	}

	return &record, nil
}

// SaveShippingFees saves shipping fee records
func (s *ShippingFeeSheetService) SaveShippingFees(
	ctx context.Context,
	tenantID string,
	records []ShippingFeeRecord,
) error {
	for i := range records {
		records[i].TenantID = tenantID
		records[i].TotalFee = records[i].BaseFee + records[i].WeightFee +
			records[i].InsuranceFee + records[i].CodFee +
			records[i].OtherFees - records[i].Discount

		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND platform = ? AND order_sn = ?",
				tenantID, records[i].Platform, records[i].OrderSN).
			Assign(records[i]).
			FirstOrCreate(&ShippingFeeRecord{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetShippingFeeSummary returns shipping fee summary
func (s *ShippingFeeSheetService) GetShippingFeeSummary(
	ctx context.Context,
	tenantID string,
	startDate time.Time,
	endDate time.Time,
) (*ShippingFeeSummary, error) {
	summary := &ShippingFeeSummary{
		ByCourier:  make(map[string]CourierStat),
		ByPlatform: make(map[string]float64),
	}

	query := s.db.WithContext(ctx).
		Model(&ShippingFeeRecord{}).
		Where("tenant_id = ?", tenantID)

	if !startDate.IsZero() {
		query = query.Where("created_at >= ?", startDate)
	}
	if !endDate.IsZero() {
		query = query.Where("created_at <= ?", endDate)
	}

	// Total summary
	var totalResult struct {
		Count     int
		Fees      float64
		Discounts float64
	}
	query.Select("COUNT(*) as count, COALESCE(SUM(total_fee), 0) as fees, COALESCE(SUM(discount), 0) as discounts").
		Scan(&totalResult)
	summary.TotalOrders = totalResult.Count
	summary.TotalFees = totalResult.Fees
	summary.TotalDiscounts = totalResult.Discounts
	summary.NetFees = summary.TotalFees
	if summary.TotalOrders > 0 {
		summary.AvgFeePerOrder = summary.TotalFees / float64(summary.TotalOrders)
	}

	// By courier
	type courierResult struct {
		Courier string
		Count   int
		Fees    float64
	}
	var courierResults []courierResult
	s.db.WithContext(ctx).
		Model(&ShippingFeeRecord{}).
		Where("tenant_id = ?", tenantID).
		Select("courier, COUNT(*) as count, COALESCE(SUM(total_fee), 0) as fees").
		Group("courier").
		Scan(&courierResults)
	for _, cr := range courierResults {
		avgFee := float64(0)
		if cr.Count > 0 {
			avgFee = cr.Fees / float64(cr.Count)
		}
		summary.ByCourier[cr.Courier] = CourierStat{
			OrderCount: cr.Count,
			TotalFees:  cr.Fees,
			AvgFee:     avgFee,
		}
	}

	// By platform
	type platformResult struct {
		Platform string
		Fees     float64
	}
	var platformResults []platformResult
	s.db.WithContext(ctx).
		Model(&ShippingFeeRecord{}).
		Where("tenant_id = ?", tenantID).
		Select("platform, COALESCE(SUM(total_fee), 0) as fees").
		Group("platform").
		Scan(&platformResults)
	for _, pr := range platformResults {
		summary.ByPlatform[pr.Platform] = pr.Fees
	}

	return summary, nil
}
