package services

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ShopeeEscrowService handles Shopee escrow analytics
type ShopeeEscrowService struct {
	db *gorm.DB
}

// NewShopeeEscrowService creates a new Shopee escrow service
func NewShopeeEscrowService(db *gorm.DB) *ShopeeEscrowService {
	return &ShopeeEscrowService{db: db}
}

// EscrowData represents escrow data response
type EscrowData struct {
	Orders     []EscrowOrderSummary `json:"orders"`
	Summary    EscrowSummary        `json:"summary"`
	Pagination PaginationInfo       `json:"pagination"`
}

// EscrowOrderSummary represents a single escrow order summary
type EscrowOrderSummary struct {
	OrderSN           string    `json:"orderSn"`
	OrderStatus       string    `json:"orderStatus"`
	ReleaseDate       time.Time `json:"releaseDate"`
	OriginalPrice     float64   `json:"originalPrice"`
	SellerDiscount    float64   `json:"sellerDiscount"`
	ShippingFee       float64   `json:"shippingFee"`
	Commission        float64   `json:"commission"`
	ServiceFee        float64   `json:"serviceFee"`
	FinalAmount       float64   `json:"finalAmount"`
	EscrowTax         float64   `json:"escrowTax"`
	ActualShippingFee float64   `json:"actualShippingFee"`
}

// EscrowSummary represents summary of escrow data
type EscrowSummary struct {
	TotalOrders     int     `json:"totalOrders"`
	TotalOriginal   float64 `json:"totalOriginal"`
	TotalDiscount   float64 `json:"totalDiscount"`
	TotalShipping   float64 `json:"totalShipping"`
	TotalCommission float64 `json:"totalCommission"`
	TotalServiceFee float64 `json:"totalServiceFee"`
	TotalFinal      float64 `json:"totalFinal"`
}

// PaginationInfo represents pagination info
type PaginationInfo struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// SyncResult represents sync result
type SyncResult struct {
	SyncID    string    `json:"syncId"`
	Status    string    `json:"status"`
	Synced    int       `json:"synced"`
	Failed    int       `json:"failed"`
	StartedAt time.Time `json:"startedAt"`
	EndedAt   time.Time `json:"endedAt"`
}

// PriceReconciliationResult represents price reconciliation result
type PriceReconciliationResult struct {
	Items         []PriceDiscrepancy `json:"items"`
	TotalItems    int                `json:"totalItems"`
	Discrepancies int                `json:"discrepancies"`
}

// PriceDiscrepancy represents a price discrepancy
type PriceDiscrepancy struct {
	OrderSN        string  `json:"orderSn"`
	SKU            string  `json:"sku"`
	ExpectedPrice  float64 `json:"expectedPrice"`
	ActualPrice    float64 `json:"actualPrice"`
	Difference     float64 `json:"difference"`
	DiffPercentage float64 `json:"diffPercentage"`
}

// ShippingAnalysisResult represents shipping analysis result
type ShippingAnalysisResult struct {
	Summary          ShippingSummary        `json:"summary"`
	ByCarrier        []CarrierStats         `json:"byCarrier"`
	ByStatus         []StatusStats          `json:"byStatus"`
	CostDistribution []CostDistributionItem `json:"costDistribution"`
}

// ShippingSummary represents shipping summary
type ShippingSummary struct {
	TotalOrders       int     `json:"totalOrders"`
	TotalShippingCost float64 `json:"totalShippingCost"`
	AvgShippingCost   float64 `json:"avgShippingCost"`
	TotalSellerPaid   float64 `json:"totalSellerPaid"`
	TotalBuyerPaid    float64 `json:"totalBuyerPaid"`
}

// CarrierStats represents carrier statistics
type CarrierStats struct {
	Carrier    string  `json:"carrier"`
	Orders     int     `json:"orders"`
	TotalCost  float64 `json:"totalCost"`
	AvgCost    float64 `json:"avgCost"`
	Percentage float64 `json:"percentage"`
}

// StatusStats represents status statistics
type StatusStats struct {
	Status     string  `json:"status"`
	Count      int     `json:"count"`
	Percentage float64 `json:"percentage"`
}

// CostDistributionItem represents cost distribution item
type CostDistributionItem struct {
	Range string `json:"range"`
	Count int    `json:"count"`
}

// GetEscrowData retrieves escrow data for a tenant
func (s *ShopeeEscrowService) GetEscrowData(ctx context.Context, tenantID string, month, year, page, limit int) (*EscrowData, error) {
	var escrowOrders []models.ShopeeEscrowOrder
	query := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("month = ? AND year = ?", month, year)

	var total int64
	query.Model(&models.ShopeeEscrowOrder{}).Count(&total)

	offset := (page - 1) * limit
	query.Offset(offset).Limit(limit).Order("synced_at DESC").Find(&escrowOrders)

	orders := make([]EscrowOrderSummary, len(escrowOrders))
	summary := EscrowSummary{}

	for i, order := range escrowOrders {
		orders[i] = EscrowOrderSummary{
			OrderSN:        order.OrderSN,
			OrderStatus:    "", // No OrderStatus in ShopeeEscrowOrder model
			ReleaseDate:    order.SyncedAt,
			OriginalPrice:  order.EscrowAmount,
			SellerDiscount: 0,
			ShippingFee:    order.BuyerPaidShippingFee,
			Commission:     order.CommissionFee,
			ServiceFee:     order.ServiceFee,
			FinalAmount:    order.EscrowAmount,
		}

		summary.TotalOriginal += order.EscrowAmount
		summary.TotalShipping += order.BuyerPaidShippingFee
		summary.TotalCommission += order.CommissionFee
		summary.TotalServiceFee += order.ServiceFee
		summary.TotalFinal += order.EscrowAmount
	}

	summary.TotalOrders = len(orders)
	totalPages := int(total) / limit
	if int(total)%limit > 0 {
		totalPages++
	}

	return &EscrowData{
		Orders:  orders,
		Summary: summary,
		Pagination: PaginationInfo{
			Page:       page,
			Limit:      limit,
			Total:      int(total),
			TotalPages: totalPages,
		},
	}, nil
}

// SyncEscrow syncs escrow data from Shopee API
func (s *ShopeeEscrowService) SyncEscrow(ctx context.Context, tenantID string, month, year int) (*SyncResult, error) {
	syncID := fmt.Sprintf("sync_%s_%d_%02d", tenantID, year, month)

	// Create sync record
	sync := models.ShopeeEscrowSync{
		ID:          syncID,
		TenantID:    tenantID,
		Month:       month,
		Year:        year,
		TotalOrders: 0,
		SyncedAt:    time.Now(),
	}
	s.db.WithContext(ctx).Create(&sync)

	// TODO: Implement actual API sync logic
	startedAt := time.Now()
	endedAt := time.Now()

	return &SyncResult{
		SyncID:    syncID,
		Status:    "completed",
		Synced:    0,
		Failed:    0,
		StartedAt: startedAt,
		EndedAt:   endedAt,
	}, nil
}

// DeleteSyncData deletes sync data
func (s *ShopeeEscrowService) DeleteSyncData(ctx context.Context, tenantID, syncID string) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, syncID).
		Delete(&models.ShopeeEscrowSync{}).Error
}

// GetPriceReconciliation gets price reconciliation data
func (s *ShopeeEscrowService) GetPriceReconciliation(ctx context.Context, tenantID string, month, year int) (*PriceReconciliationResult, error) {
	// TODO: Implement price reconciliation logic
	return &PriceReconciliationResult{
		Items:         []PriceDiscrepancy{},
		TotalItems:    0,
		Discrepancies: 0,
	}, nil
}

// GetShippingAnalysis gets shipping analysis data
func (s *ShopeeEscrowService) GetShippingAnalysis(ctx context.Context, tenantID string, month, year int) (*ShippingAnalysisResult, error) {
	var escrowOrders []models.ShopeeEscrowOrder
	s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("month = ? AND year = ?", month, year).
		Find(&escrowOrders)

	summary := ShippingSummary{TotalOrders: len(escrowOrders)}
	carrierMap := make(map[string]*CarrierStats)

	for _, order := range escrowOrders {
		summary.TotalShippingCost += order.ActualShippingFee
		summary.TotalBuyerPaid += order.BuyerPaidShippingFee

		// Use "Standard" as default carrier since ShippingCarrier is not in model
		carrier := "Standard"
		if _, exists := carrierMap[carrier]; !exists {
			carrierMap[carrier] = &CarrierStats{Carrier: carrier}
		}
		carrierMap[carrier].Orders++
		carrierMap[carrier].TotalCost += order.ActualShippingFee
	}

	if summary.TotalOrders > 0 {
		summary.AvgShippingCost = summary.TotalShippingCost / float64(summary.TotalOrders)
	}

	byCarrier := make([]CarrierStats, 0, len(carrierMap))
	for _, stats := range carrierMap {
		if stats.Orders > 0 {
			stats.AvgCost = stats.TotalCost / float64(stats.Orders)
			stats.Percentage = float64(stats.Orders) / float64(summary.TotalOrders) * 100
		}
		byCarrier = append(byCarrier, *stats)
	}

	return &ShippingAnalysisResult{
		Summary:          summary,
		ByCarrier:        byCarrier,
		ByStatus:         []StatusStats{},
		CostDistribution: []CostDistributionItem{},
	}, nil
}
