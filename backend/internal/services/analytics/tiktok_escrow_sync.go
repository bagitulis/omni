// Package analytics provides TikTok escrow sync service
package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omni/backend/internal/dto"
	"github.com/omni/backend/internal/services"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
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

// SyncMonth syncs escrow data for a specific month from TikTok API.
// Uses order-centric approach: SearchOrders -> GetOrderDetail (batch) -> GetOrderTransactions (per-order).
func (s *TiktokEscrowSyncService) SyncMonth(
	ctx context.Context,
	month, year int,
	forceResync bool,
) (*dto.SyncResultDTO, error) {
	log.Info().Msgf("[TiktokEscrowSync] Starting sync for %d-%02d, tenant: %s", year, month, s.tenantID)

	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return nil, fmt.Errorf("cannot sync current month, wait until month ends")
	}

	tables := TiktokEscrowTables()

	if !forceResync {
		existing, retryIDs := s.checkExistingSyncForRetry(ctx, tables, month, year)
		if existing != nil && retryIDs == nil {
			return &dto.SyncResultDTO{
				TotalOrders: existing.TotalOrders,
				Message:     "Already synced. Use forceResync to update.",
			}, nil
		}
	}

	client, err := s.getTiktokClient()
	if err != nil {
		return nil, fmt.Errorf("get tiktok client: %w", err)
	}

	orders, err := s.fetchOrdersForSettlementWindow(ctx, client, month, year)
	if err != nil {
		return nil, fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		return &dto.SyncResultDTO{
			TotalOrders: 0,
			Message:     "No completed orders found for this period",
		}, nil
	}

	log.Info().Msgf("[TiktokEscrowSync] Found %d candidate orders in sync window", len(orders))

	if forceResync {
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return nil, err
		}
	}

	totalItems, processedOrders, failedOrders := s.processOrdersConcurrent(ctx, client, orders, month, year)

	failedIDs := []string{}
	if err := s.saveSyncRecord(ctx, tables, month, year, processedOrders, failedOrders, failedIDs); err != nil {
		return nil, err
	}

	return &dto.SyncResultDTO{
		TotalOrders:  processedOrders,
		TotalItems:   totalItems,
		FailedOrders: failedOrders,
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

// deleteMonthData deletes existing data for the month
func (s *TiktokEscrowSyncService) deleteMonthData(ctx context.Context, month, year int) error {
	tables := TiktokEscrowTables()
	return s.base.DeleteMonthDataGeneric(
		ctx, tables.OrderTable, tables.ItemTable, tables.SyncTable, month, year,
	)
}

// fetchOrderTransaction fetches transaction details for an order.
func (s *TiktokEscrowSyncService) fetchOrderTransaction(
	ctx context.Context,
	client *tiktokPkg.Client,
	orderID string,
) (*tiktokPkg.OrderTransactionResponse, error) {
	resp, err := client.GetOrderTransactionsV202309(orderID)
	if err != nil || resp.Code != 0 {
		log.Warn().Str("order_id", orderID).Str("api_version", "v202309").
			Err(err).Msg("[TiktokEscrowSync] v202309 API failed, trying v202501")
		resp, err = client.GetOrderTransactions(orderID)
		if err != nil {
			return nil, fmt.Errorf("both API versions failed for order %s: %w", orderID, err)
		}
	}

	if resp.Code != 0 {
		rawResp, _ := json.Marshal(resp)
		log.Warn().Str("order_id", orderID).Int("api_code", resp.Code).
			Str("api_message", resp.Message).Str("raw_response", string(rawResp)).
			Msg("[TiktokEscrowSync] TikTok API returned error")
		return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
	}

	if len(resp.Data.SkuTransactions) == 0 {
		v2Resp, v2Err := client.GetOrderTransactions(orderID)
		if v2Err == nil && v2Resp.Code == 0 && len(v2Resp.Data.SkuTransactions) > 0 {
			resp.Data.SkuTransactions = v2Resp.Data.SkuTransactions
		}
	}

	return resp, nil
}
