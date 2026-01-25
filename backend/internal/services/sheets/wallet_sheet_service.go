package sheets

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// WalletTransaction represents a wallet/escrow transaction
type WalletTransaction struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	TenantID        string    `gorm:"index;not null" json:"tenantId"`
	Platform        string    `gorm:"index;not null" json:"platform"`
	TransactionID   string    `gorm:"index" json:"transactionId"`
	OrderSN         string    `gorm:"index" json:"orderSn,omitempty"`
	TransactionType string    `json:"transactionType"` // escrow_release, withdrawal, fee, etc.
	Amount          float64   `json:"amount"`
	Fee             float64   `json:"fee"`
	NetAmount       float64   `json:"netAmount"`
	Currency        string    `json:"currency"`
	Status          string    `json:"status"`
	Description     string    `json:"description,omitempty"`
	TransactionDate time.Time `json:"transactionDate"`
	Source          string    `json:"source"` // api, sheet
	SheetConfigID   uint      `json:"sheetConfigId,omitempty"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (WalletTransaction) TableName() string {
	return "wallet_transactions"
}

// WalletBalance represents current wallet balance
type WalletBalance struct {
	ID             uint      `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"index;not null" json:"tenantId"`
	Platform       string    `gorm:"index;not null" json:"platform"`
	AvailableBalance float64 `json:"availableBalance"`
	PendingBalance   float64 `json:"pendingBalance"`
	TotalBalance     float64 `json:"totalBalance"`
	Currency       string    `json:"currency"`
	LastUpdated    time.Time `json:"lastUpdated"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// TableName returns the table name for GORM
func (WalletBalance) TableName() string {
	return "wallet_balances"
}

// WalletSummary represents wallet summary across platforms
type WalletSummary struct {
	TotalAvailable float64                     `json:"totalAvailable"`
	TotalPending   float64                     `json:"totalPending"`
	ByPlatform     map[string]*WalletBalance   `json:"byPlatform"`
}

// WalletSheetService handles wallet/escrow sheet operations
type WalletSheetService struct {
	db *gorm.DB
}

// NewWalletSheetService creates a new wallet sheet service
func NewWalletSheetService(db *gorm.DB) *WalletSheetService {
	return &WalletSheetService{db: db}
}

// GetTransactions retrieves wallet transactions
func (s *WalletSheetService) GetTransactions(
	ctx context.Context,
	tenantID string,
	platform string,
	limit int,
	offset int,
) ([]WalletTransaction, int64, error) {
	var transactions []WalletTransaction
	var total int64

	query := s.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	if err := query.Model(&WalletTransaction{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Order("transaction_date DESC").Find(&transactions).Error
	return transactions, total, err
}

// SaveTransactions saves wallet transactions
func (s *WalletSheetService) SaveTransactions(
	ctx context.Context,
	tenantID string,
	transactions []WalletTransaction,
) error {
	for i := range transactions {
		transactions[i].TenantID = tenantID
		transactions[i].NetAmount = transactions[i].Amount - transactions[i].Fee

		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND platform = ? AND transaction_id = ?",
				tenantID, transactions[i].Platform, transactions[i].TransactionID).
			Assign(transactions[i]).
			FirstOrCreate(&WalletTransaction{}).Error
		if err != nil {
			return err
		}
	}
	return nil
}

// GetWalletBalance retrieves wallet balance for a platform
func (s *WalletSheetService) GetWalletBalance(
	ctx context.Context,
	tenantID string,
	platform string,
) (*WalletBalance, error) {
	var balance WalletBalance

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", tenantID, platform).
		First(&balance).Error

	if err != nil {
		return nil, err
	}

	return &balance, nil
}

// SaveWalletBalance saves or updates wallet balance
func (s *WalletSheetService) SaveWalletBalance(
	ctx context.Context,
	balance *WalletBalance,
) error {
	balance.TotalBalance = balance.AvailableBalance + balance.PendingBalance
	balance.LastUpdated = time.Now()

	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ?", balance.TenantID, balance.Platform).
		Assign(balance).
		FirstOrCreate(&WalletBalance{}).Error
}

// GetWalletSummary returns wallet summary across all platforms
func (s *WalletSheetService) GetWalletSummary(
	ctx context.Context,
	tenantID string,
) (*WalletSummary, error) {
	var balances []WalletBalance

	err := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Find(&balances).Error

	if err != nil {
		return nil, err
	}

	summary := &WalletSummary{
		ByPlatform: make(map[string]*WalletBalance),
	}

	for i := range balances {
		summary.TotalAvailable += balances[i].AvailableBalance
		summary.TotalPending += balances[i].PendingBalance
		summary.ByPlatform[balances[i].Platform] = &balances[i]
	}

	return summary, nil
}

// GetTransactionsByOrder retrieves transactions for an order
func (s *WalletSheetService) GetTransactionsByOrder(
	ctx context.Context,
	tenantID string,
	orderSN string,
) ([]WalletTransaction, error) {
	var transactions []WalletTransaction

	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND order_sn = ?", tenantID, orderSN).
		Order("transaction_date DESC").
		Find(&transactions).Error

	return transactions, err
}
