package shopee

import (
	"context"
	"fmt"
	"time"

	"github.com/omni/backend/internal/services"
	"github.com/omni/backend/internal/utils"
	shopeePkg "github.com/omni/backend/pkg/shopee"
)

// WalletService handles Shopee wallet operations
type WalletService struct {
	apiClient   APIClient
	tenantID    string
	credService *services.CredentialService
	dbPath      string
}

// NewWalletService creates a new wallet service
func NewWalletService(apiClient APIClient, tenantID string) *WalletService {
	return &WalletService{apiClient: apiClient, tenantID: tenantID}
}

// NewWalletServiceWithCreds creates a wallet service with credential support
func NewWalletServiceWithCreds(tenantID, dbPath string) *WalletService {
	return &WalletService{
		tenantID:    tenantID,
		dbPath:      dbPath,
		credService: services.NewCredentialService(dbPath),
	}
}

// WalletBalance represents wallet balance
type WalletBalance struct {
	TotalBalance     float64 `json:"totalBalance"`
	PendingBalance   float64 `json:"pendingBalance"`
	AvailableBalance float64 `json:"availableBalance"`
	Currency         string  `json:"currency"`
	UpdatedAt        string  `json:"updatedAt"`
}

// Transaction represents a wallet transaction
type Transaction struct {
	TransactionID   string    `json:"transactionId"`
	TransactionType string    `json:"transactionType"`
	Amount          float64   `json:"amount"`
	Status          string    `json:"status"`
	OrderSN         string    `json:"orderSn,omitempty"`
	Description     string    `json:"description"`
	CreateTime      time.Time `json:"createTime"`
}

// TransactionFilter represents filter for transaction queries
type TransactionFilter struct {
	StartDate time.Time
	EndDate   time.Time
	Type      string
	PageSize  int
	PageToken string
}

// TransactionResult represents paginated transaction result
type TransactionResult struct {
	Transactions []Transaction `json:"transactions"`
	TotalCount   int           `json:"totalCount"`
	HasMore      bool          `json:"hasMore"`
	NextToken    string        `json:"nextPageToken,omitempty"`
}

// getClient creates a Shopee client with credentials
func (s *WalletService) getClient() (*shopeePkg.Client, error) {
	if s.credService == nil {
		return nil, fmt.Errorf("credential service not initialized")
	}

	creds, err := s.credService.GetPlatformCredentials(s.tenantID, "shopee")
	if err != nil {
		return nil, err
	}

	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	return client, nil
}

// GetBalance gets current wallet balance
func (s *WalletService) GetBalance(ctx context.Context) (*WalletBalance, error) {
	// Get transactions and calculate balance
	filter := TransactionFilter{
		StartDate: time.Now().AddDate(0, -1, 0),
		EndDate:   time.Now(),
		PageSize:  100,
	}

	result, err := s.GetTransactions(ctx, filter)
	if err != nil {
		return nil, err
	}

	var total, pending, available float64
	for _, tx := range result.Transactions {
		total += tx.Amount
		if tx.Status == "COMPLETED" {
			available += tx.Amount
		} else {
			pending += tx.Amount
		}
	}

	return &WalletBalance{
		TotalBalance:     total,
		PendingBalance:   pending,
		AvailableBalance: available,
		Currency:         "IDR",
		UpdatedAt:        utils.NowWIB().Format(time.RFC3339),
	}, nil
}

// GetTransactions gets wallet transactions
func (s *WalletService) GetTransactions(ctx context.Context, filter TransactionFilter) (*TransactionResult, error) {
	client, err := s.getClient()
	if err != nil {
		return nil, err
	}

	pageNo := 1
	if filter.PageToken != "" {
		fmt.Sscanf(filter.PageToken, "%d", &pageNo)
	}

	pageSize := filter.PageSize
	if pageSize == 0 {
		pageSize = 50
	}

	req := shopeePkg.GetWalletTransactionRequest{
		PageNo:    pageNo,
		PageSize:  pageSize,
		StartDate: filter.StartDate.Unix(),
		EndDate:   filter.EndDate.Unix(),
	}

	if filter.Type != "" {
		req.TransactionTypes = []string{filter.Type}
	}

	result, err := client.GetWalletTransactions(req)
	if err != nil {
		return nil, fmt.Errorf("get wallet transactions: %w", err)
	}

	transactions := make([]Transaction, len(result.Response.TransactionList))
	for i, tx := range result.Response.TransactionList {
		transactions[i] = Transaction{
			TransactionID:   fmt.Sprintf("%d", tx.TransactionID),
			TransactionType: tx.TransactionType,
			Amount:          tx.Amount,
			Status:          tx.Status,
			OrderSN:         tx.OrderSN,
			Description:     tx.Description,
			CreateTime:      time.Unix(tx.CreateTime, 0),
		}
	}

	nextToken := ""
	if result.Response.More {
		nextToken = fmt.Sprintf("%d", pageNo+1)
	}

	return &TransactionResult{
		Transactions: transactions,
		TotalCount:   len(transactions),
		HasMore:      result.Response.More,
		NextToken:    nextToken,
	}, nil
}

// CalculateNetIncome calculates net income from transactions
func (s *WalletService) CalculateNetIncome(ctx context.Context, startDate, endDate time.Time) (*NetIncome, error) {
	filter := TransactionFilter{
		StartDate: startDate,
		EndDate:   endDate,
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

	return s.calculateIncome(allTx), nil
}

// NetIncome represents net income calculation
type NetIncome struct {
	GrossIncome  float64 `json:"grossIncome"`
	TotalFees    float64 `json:"totalFees"`
	NetIncome    float64 `json:"netIncome"`
	OrderCount   int     `json:"orderCount"`
	RefundAmount float64 `json:"refundAmount"`
	RefundCount  int     `json:"refundCount"`
}

func (s *WalletService) calculateIncome(transactions []Transaction) *NetIncome {
	result := &NetIncome{}
	for _, tx := range transactions {
		switch tx.TransactionType {
		case "ESCROW_RELEASED", "SETTLEMENT":
			result.GrossIncome += tx.Amount
			result.OrderCount++
		case "COMMISSION", "SERVICE_FEE", "PLATFORM_FEE":
			result.TotalFees += tx.Amount
		case "REFUND":
			result.RefundAmount += tx.Amount
			result.RefundCount++
		}
	}
	result.NetIncome = result.GrossIncome - result.TotalFees - result.RefundAmount
	return result
}

// Note: ProcessedTransactions, ProcessTransactions, GetMonthlyTransactions, categorizeTransaction
// are defined in wallet_transactions.go
