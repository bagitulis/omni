package shopee

import (
	"context"
	"time"

	"github.com/omni/backend/internal/utils"
)

// ProcessedTransactions represents processed transaction data
type ProcessedTransactions struct {
	Transactions []map[string]interface{} `json:"transactions"`
	TotalAmount  float64                  `json:"total_amount"`
	Count        int                      `json:"count"`
	OrderNumbers []string                 `json:"order_numbers"`
}

// ProcessTransactions processes raw transactions into exportable format
func (s *WalletService) ProcessTransactions(transactions []Transaction) *ProcessedTransactions {
	result := &ProcessedTransactions{
		Transactions: make([]map[string]interface{}, 0, len(transactions)),
		OrderNumbers: make([]string, 0),
	}

	orderSet := make(map[string]bool)

	for _, tx := range transactions {
		processed := map[string]interface{}{
			"date":             utils.ToWIB(tx.CreateTime).Format("2006-01-02 15:04:05"),
			"order_sn":         tx.OrderSN,
			"description":      tx.Description,
			"amount":           tx.Amount,
			"status":           tx.Status,
			"transaction_type": tx.TransactionType,
			"tab_type":         categorizeTransaction(tx.TransactionType),
			"buyer_name":       "",
		}
		result.Transactions = append(result.Transactions, processed)
		result.TotalAmount += tx.Amount

		if tx.OrderSN != "" && !orderSet[tx.OrderSN] {
			orderSet[tx.OrderSN] = true
			result.OrderNumbers = append(result.OrderNumbers, tx.OrderSN)
		}
	}

	result.Count = len(transactions)
	return result
}

// categorizeTransaction categorizes transaction by type
func categorizeTransaction(txType string) string {
	switch txType {
	case "ESCROW_RELEASED", "SETTLEMENT":
		return "Income"
	case "COMMISSION", "SERVICE_FEE", "PLATFORM_FEE":
		return "Fee"
	case "REFUND":
		return "Refund"
	case "WITHDRAWAL":
		return "Withdrawal"
	default:
		return "Other"
	}
}

// GetMonthlyTransactions gets transactions for a specific month
func (s *WalletService) GetMonthlyTransactions(ctx context.Context, month, year int, txType string) ([]Transaction, error) {
	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0).Add(-time.Second)

	filter := TransactionFilter{
		StartDate: startDate,
		EndDate:   endDate,
		Type:      txType,
		PageSize:  100,
	}

	var allTx []Transaction
	for {
		result, err := s.GetTransactions(ctx, filter)
		if err != nil {
			return nil, err
		}
		allTx = append(allTx, result.Transactions...)
		if !result.HasMore {
			break
		}
		filter.PageToken = result.NextToken
	}

	return allTx, nil
}
