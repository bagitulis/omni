// Package analytics provides TikTok escrow sync service
package analytics

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"gorm.io/gorm"
)

// TiktokEscrowSyncService handles syncing escrow data from TikTok API
type TiktokEscrowSyncService struct {
	base     BaseEscrowService
	tenantID string
}

// NewTiktokEscrowSyncService creates a new TikTok escrow sync service
func NewTiktokEscrowSyncService(db *gorm.DB, tenantID, dbPath string) *TiktokEscrowSyncService {
	return &TiktokEscrowSyncService{
		base:     NewBaseEscrowService(db, tenantID, dbPath),
		tenantID: tenantID,
	}
}

// SyncMonth syncs escrow data for a specific month from TikTok API
func (s *TiktokEscrowSyncService) SyncMonth(
	ctx context.Context,
	month, year int,
	forceResync bool,
) (*dto.SyncResultDTO, error) {
	log.Printf("[TiktokEscrowSync] Starting sync for %d-%02d, tenant: %s", year, month, s.tenantID)

	// Check if current month (cannot sync)
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return nil, fmt.Errorf("cannot sync current month, wait until month ends")
	}

	// Check if already synced
	tables := TiktokEscrowTables()
	if !forceResync {
		var existing models.TiktokEscrowSync
		err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
			Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
			First(&existing).Error
		if err == nil {
			return &dto.SyncResultDTO{
				TotalOrders: existing.TotalOrders,
				Message:     "Already synced. Use forceResync to update.",
			}, nil
		}
	}

	// Get TikTok client
	client, err := s.getTiktokClient()
	if err != nil {
		return nil, fmt.Errorf("get tiktok client: %w", err)
	}

	// Fetch completed orders for the month
	orders, err := s.fetchOrdersByMonth(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No completed orders found for this period",
		}, nil
	}

	log.Printf("[TiktokEscrowSync] Found %d completed orders", len(orders))

	// Delete existing data if force resync
	if forceResync {
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	// Process orders
	totalItems, processedOrders, failedOrders := s.processOrders(ctx, client, orders, month, year)

	// Create sync record
	syncRecord := models.TiktokEscrowSync{
		ID:          uuid.New().String(),
		TenantID:    s.tenantID,
		Month:       month,
		Year:        year,
		TotalOrders: processedOrders,
		SyncedAt:    time.Now(),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := s.base.DB.WithContext(ctx).Table(s.base.Table(tables.SyncTable)).
		Create(&syncRecord).Error; err != nil {
		return nil, err
	}

	return &dto.SyncResultDTO{
		TotalOrders: processedOrders,
		TotalItems:  totalItems,
		Message: fmt.Sprintf("Synced %d/%d orders (%d failed), %d items from TikTok API",
			processedOrders, len(orders), failedOrders, totalItems),
	}, nil
}

// getTiktokClient creates TikTok client with tenant credentials
func (s *TiktokEscrowSyncService) getTiktokClient() (*tiktokPkg.Client, error) {
	credService := services.NewCredentialService(s.base.DBPath)
	creds, err := credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		return nil, err
	}

	if creds.AppKey == "" || creds.AppSecret == "" {
		return nil, fmt.Errorf("tiktok credentials not configured for tenant")
	}

	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)
	return client, nil
}

// fetchOrdersByMonth fetches completed orders for a month
func (s *TiktokEscrowSyncService) fetchOrdersByMonth(
	ctx context.Context,
	client *tiktokPkg.Client,
	month, year int,
) ([]tiktokPkg.TiktokOrder, error) {
	var allOrders []tiktokPkg.TiktokOrder

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, 0)

	req := tiktokPkg.OrderSearchRequest{
		OrderStatus:  "COMPLETED",
		CreateTimeGe: startDate.Unix(),
		CreateTimeLt: endDate.Unix(),
	}

	pageToken := ""
	pageCount := 0

	for {
		resp, err := client.SearchOrders(req, 100, pageToken)
		if err != nil {
			return nil, err
		}

		if resp.Code != 0 {
			return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
		}

		allOrders = append(allOrders, resp.Data.Orders...)
		pageCount++
		log.Printf("[TiktokEscrowSync] Fetched page %d: %d orders", pageCount, len(resp.Data.Orders))

		pageToken = resp.Data.NextPageToken
		if pageToken == "" {
			break
		}

		// Rate limiting
		time.Sleep(500 * time.Millisecond)
	}

	return allOrders, nil
}

// processOrders processes all orders and returns counts
func (s *TiktokEscrowSyncService) processOrders(
	ctx context.Context,
	client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder,
	month, year int,
) (totalItems, processedOrders, failedOrders int) {
	for _, order := range orders {
		orderID := order.ID
		transaction, err := s.fetchOrderTransaction(ctx, client, orderID)
		if err != nil {
			log.Printf("[TiktokEscrowSync] Error fetching transaction for order %s: %v", orderID, err)
			failedOrders++
			continue
		}

		if transaction == nil || transaction.Data.OrderID == "" {
			log.Printf("[TiktokEscrowSync] No transaction data for order %s", orderID)
			failedOrders++
			continue
		}

		itemsCount, err := s.saveEscrowOrder(ctx, order, transaction, month, year)
		if err != nil {
			log.Printf("[TiktokEscrowSync] Error saving order %s: %v", orderID, err)
			failedOrders++
			continue
		}

		totalItems += itemsCount
		processedOrders++

		// Rate limiting (reduced from 500ms to avoid timeout on large months)
		time.Sleep(150 * time.Millisecond)
	}
	return totalItems, processedOrders, failedOrders
}

// fetchOrderTransaction fetches transaction details for an order
func (s *TiktokEscrowSyncService) fetchOrderTransaction(
	ctx context.Context,
	client *tiktokPkg.Client,
	orderID string,
) (*tiktokPkg.OrderTransactionResponse, error) {
	// Try v202501 API first
	resp, err := client.GetOrderTransactions(orderID)
	if err != nil {
		// Fallback to v202309
		resp, err = client.GetOrderTransactionsV202309(orderID)
		if err != nil {
			return nil, err
		}
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
	}

	return resp, nil
}

// deleteMonthData deletes existing data for the month
func (s *TiktokEscrowSyncService) deleteMonthData(ctx context.Context, month, year int) error {
	tables := TiktokEscrowTables()
	return s.base.DeleteMonthDataGeneric(
		ctx, tables.OrderTable, tables.ItemTable, tables.SyncTable, month, year,
	)
}
