package shopee

import (
	"context"
	"fmt"

	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// EscrowService handles escrow detail operations
type EscrowService struct {
	client   *shopeePkg.Client
	tenantID string
}

// NewEscrowService creates a new escrow service
func NewEscrowService(client *shopeePkg.Client, tenantID string) *EscrowService {
	return &EscrowService{client: client, tenantID: tenantID}
}

// EscrowDetail represents detailed escrow information for an order
type EscrowDetail struct {
	OrderSN       string                 `json:"order_sn"`
	BuyerUsername string                 `json:"buyer_username"`
	PayTime       int64                  `json:"pay_time"`
	OrderStatus   string                 `json:"order_status"`
	TotalAmount   float64                `json:"total_amount"`
	Incomes       []EscrowIncome         `json:"incomes"`
	Items         []EscrowItem           `json:"items"`
	RawData       map[string]interface{} `json:"raw_data,omitempty"`
}

// EscrowIncome represents an income/fee component
type EscrowIncome struct {
	EscrowAccountType string  `json:"escrow_account_type"`
	Amount            float64 `json:"amount"`
	Description       string  `json:"description"`
}

// EscrowItem represents an order item in escrow
type EscrowItem struct {
	ItemID          int64   `json:"item_id"`
	ItemName        string  `json:"item_name"`
	SKU             string  `json:"sku"`
	Quantity        int     `json:"quantity"`
	OriginalPrice   float64 `json:"original_price"`
	DiscountedPrice float64 `json:"discounted_price"`
}

// GetEscrowDetail gets escrow detail for a single order
func (s *EscrowService) GetEscrowDetail(ctx context.Context, orderSN string) (*EscrowDetail, error) {
	details, err := s.GetEscrowDetailsBatch(ctx, []string{orderSN})
	if err != nil {
		return nil, err
	}

	if len(details) == 0 {
		return nil, nil
	}

	return details[0], nil
}

// GetEscrowDetailsBatch gets escrow details for multiple orders
func (s *EscrowService) GetEscrowDetailsBatch(ctx context.Context, orderSNs []string) ([]*EscrowDetail, error) {
	// Process in batches of 50
	const batchSize = 50
	results := make([]*EscrowDetail, 0, len(orderSNs))

	for i := 0; i < len(orderSNs); i += batchSize {
		end := i + batchSize
		if end > len(orderSNs) {
			end = len(orderSNs)
		}

		batch := orderSNs[i:end]
		batchResults, err := s.processBatch(ctx, s.client, batch)
		if err != nil {
			return nil, err
		}
		results = append(results, batchResults...)
	}

	return results, nil
}

// processBatch processes a batch of order SNs
func (s *EscrowService) processBatch(ctx context.Context, client *shopeePkg.Client, orderSNs []string) ([]*EscrowDetail, error) {
	escrowResp, err := client.GetEscrowDetails(shopeePkg.GetEscrowDetailsRequest{
		OrderSNList: orderSNs,
	})
	if err != nil {
		return nil, fmt.Errorf("get escrow details: %w", err)
	}

	// Use GetOrderList() helper for batch response
	orderList := escrowResp.GetOrderList()
	results := make([]*EscrowDetail, 0, len(orderList))

	for _, order := range orderList {
		detail := &EscrowDetail{
			OrderSN:       order.OrderSN,
			BuyerUsername: order.BuyerUsername,
			PayTime:       order.PayTime,
			OrderStatus:   order.OrderStatus,
			TotalAmount:   order.OrderIncome.EscrowAmount,
			Incomes:       make([]EscrowIncome, 0),
			Items:         make([]EscrowItem, 0),
		}

		// Add main income from order_income
		detail.Incomes = append(detail.Incomes, EscrowIncome{
			EscrowAccountType: "escrow_amount",
			Amount:            order.OrderIncome.EscrowAmount,
			Description:       "Escrow Amount",
		})

		// Process items from order_income.items
		for _, item := range order.OrderIncome.Items {
			detail.Items = append(detail.Items, EscrowItem{
				ItemID:          item.ItemID,
				ItemName:        item.ItemName,
				SKU:             item.ModelSKU,
				Quantity:        item.QuantityPurchased,
				OriginalPrice:   item.OriginalPrice,
				DiscountedPrice: item.DiscountedPrice,
			})
		}

		results = append(results, detail)
	}

	return results, nil
}
