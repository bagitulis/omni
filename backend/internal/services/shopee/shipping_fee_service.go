package shopee

import (
	"context"
	"fmt"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// ShippingFeeService handles shipping fee processing
type ShippingFeeService struct {
	client   *shopeePkg.Client
	tenantID string
}

// NewShippingFeeService creates a new shipping fee service
func NewShippingFeeService(client *shopeePkg.Client, tenantID string) *ShippingFeeService {
	return &ShippingFeeService{client: client, tenantID: tenantID}
}

// ShippingFeeData represents shipping fee data for a month
type ShippingFeeData struct {
	Fees  []map[string]interface{} `json:"fees"`
	Total float64                  `json:"total"`
	Count int                      `json:"count"`
}

// ShippingFeeResult represents shipping fee for a single order
type ShippingFeeResult struct {
	OrderSN           string  `json:"order_sn"`
	ShippingFee       float64 `json:"shipping_fee"`
	ActualShippingFee float64 `json:"actual_shipping_fee"`
	Discount          float64 `json:"discount"`
	Subsidy           float64 `json:"subsidy"`
	CreatedDate       string  `json:"created_date"`
	Success           bool    `json:"success"`
	Error             string  `json:"error,omitempty"`
}

// ProcessShippingFees processes shipping fees for orders
func (s *ShippingFeeService) ProcessShippingFees(ctx context.Context, orderSNs []string) ([]ShippingFeeResult, error) {
	results := make([]ShippingFeeResult, 0, len(orderSNs))

	// Process in batches of 50
	const batchSize = 50
	for i := 0; i < len(orderSNs); i += batchSize {
		end := i + batchSize
		if end > len(orderSNs) {
			end = len(orderSNs)
		}

		batch := orderSNs[i:end]
		batchResults, err := s.processBatch(ctx, batch)
		if err != nil {
			// Log error but continue with other batches
			for _, orderSN := range batch {
				results = append(results, ShippingFeeResult{
					OrderSN: orderSN,
					Success: false,
					Error:   err.Error(),
				})
			}
			continue
		}
		results = append(results, batchResults...)
	}

	return results, nil
}

// processBatch processes a batch of orders
func (s *ShippingFeeService) processBatch(ctx context.Context, orderSNs []string) ([]ShippingFeeResult, error) {
	escrowResp, err := s.client.GetEscrowDetails(shopeePkg.GetEscrowDetailsRequest{
		OrderSNList: orderSNs,
	})
	if err != nil {
		return nil, fmt.Errorf("get escrow details: %w", err)
	}

	results := make([]ShippingFeeResult, 0, len(orderSNs))
	// Use GetOrderList() helper for batch response
	for _, order := range escrowResp.GetOrderList() {
		// Use order_income data directly
		income := order.OrderIncome

		results = append(results, ShippingFeeResult{
			OrderSN:           order.OrderSN,
			ShippingFee:       income.BuyerPaidShippingFee,
			ActualShippingFee: income.ActualShippingFee,
			Discount:          income.SellerShippingDiscount,
			Subsidy:           income.ShopeeShippingRebate,
			CreatedDate:       time.Unix(order.PayTime, 0).Format("2006-01-02"),
			Success:           true,
		})
	}

	return results, nil
}

// GetMonthlyShippingFees gets shipping fees for a month
func (s *ShippingFeeService) GetMonthlyShippingFees(ctx context.Context, month, year int) (*ShippingFeeData, error) {
	// First, get orders for the month
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	// Get order list for the period (empty orderStatus = all statuses) with pagination
	var allOrderSNs []string
	cursor := ""

	for {
		orderResp, err := s.client.GetOrderList(ctx, startDate.Unix(), endDate.Unix(), "create_time", "", cursor)
		if err != nil {
			return nil, fmt.Errorf("get order list: %w", err)
		}

		if len(orderResp.Response.OrderList) == 0 {
			break
		}

		for _, order := range orderResp.Response.OrderList {
			allOrderSNs = append(allOrderSNs, order.OrderSN)
		}

		if !orderResp.Response.More || orderResp.Response.NextCursor == "" {
			break
		}
		cursor = orderResp.Response.NextCursor
	}

	if len(allOrderSNs) == 0 {
		return &ShippingFeeData{
			Fees:  []map[string]interface{}{},
			Total: 0,
			Count: 0,
		}, nil
	}

	// Get shipping fees
	feeResults, err := s.ProcessShippingFees(ctx, allOrderSNs)
	if err != nil {
		return nil, err
	}

	// Build response
	fees := make([]map[string]interface{}, 0, len(feeResults))
	var total float64

	for _, fee := range feeResults {
		if fee.Success {
			fees = append(fees, map[string]interface{}{
				"order_sn":            fee.OrderSN,
				"shipping_fee":        fee.ShippingFee,
				"actual_shipping_fee": fee.ActualShippingFee,
				"discount":            fee.Discount,
				"subsidy":             fee.Subsidy,
				"created_date":        fee.CreatedDate,
			})
			total += fee.ActualShippingFee
		}
	}

	return &ShippingFeeData{
		Fees:  fees,
		Total: total,
		Count: len(fees),
	}, nil
}
