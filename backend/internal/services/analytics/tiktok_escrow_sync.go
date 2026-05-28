package analytics

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/services"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type TiktokEscrowSyncService struct {
	systemDB *gorm.DB
	tenantDB *gorm.DB
	tenantID string
	client   TiktokEscrowClient
}

func NewTiktokEscrowSyncService(systemDB *gorm.DB, tenantDB *gorm.DB, tenantID string) *TiktokEscrowSyncService {
	return &TiktokEscrowSyncService{
		systemDB: systemDB,
		tenantDB: tenantDB,
		tenantID: tenantID,
	}
}

func (s *TiktokEscrowSyncService) SetClient(client TiktokEscrowClient) {
	s.client = client
}

func (s *TiktokEscrowSyncService) SyncMonthWithProgress(
	ctx context.Context, month, year int, forceResync bool,
	onProgress func(processed, total int, message string),
) error {
	log.Info().Msgf("[TiktokEscrowSync] Starting sync for %d-%02d, tenant: %s", year, month, s.tenantID)
	onProgress(0, 100, "Starting sync...")

	if err := s.validateMonth(month, year); err != nil {
		return err
	}

	tables := tiktokEscrowTables()

	if !forceResync {
		if s.isAlreadySynced(ctx, month, year) {
			onProgress(100, 100, "Already synced")
			return nil
		}
	}

	client, err := s.getClient(ctx)
	if err != nil {
		return fmt.Errorf("get tiktok client: %w", err)
	}

	onProgress(10, 100, "Fetching completed orders...")
	orders, err := s.fetchOrdersForSettlementWindow(ctx, client, month, year)
	if err != nil {
		return fmt.Errorf("fetch orders: %w", err)
	}

	if len(orders) == 0 {
		onProgress(100, 100, "No completed orders found")
		return nil
	}

	if forceResync {
		onProgress(15, 100, "Clearing existing data...")
		if err := s.deleteMonthData(ctx, month, year); err != nil {
			return err
		}
	}

	onProgress(18, 100, "Enriching order details...")
	s.enrichOrdersWithDetails(ctx, client, orders)

	onProgress(20, 100, "Processing orders...")
	totalItems, processedOrders, failedOrders, failedIDs := s.processOrdersConcurrent(ctx, client, orders, month, year)

	onProgress(95, 100, "Saving sync record...")
	if err := s.saveSyncRecord(ctx, tables, month, year, processedOrders, failedOrders, failedIDs); err != nil {
		return err
	}

	onProgress(100, 100, "Sync completed")
	log.Info().Msgf("[TiktokEscrowSync] Completed: %d orders (%d failed), %d items", processedOrders, failedOrders, totalItems)
	return nil
}

func (s *TiktokEscrowSyncService) validateMonth(month, year int) error {
	now := time.Now()
	if month == int(now.Month()) && year == now.Year() {
		return fmt.Errorf("cannot sync current month, wait until month ends")
	}
	return nil
}

func (s *TiktokEscrowSyncService) isAlreadySynced(ctx context.Context, month, year int) bool {
	var existing models.TiktokEscrowSync
	err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		First(&existing).Error
	return err == nil
}

func (s *TiktokEscrowSyncService) getClient(_ context.Context) (TiktokEscrowClient, error) {
	if s.client != nil {
		return s.client, nil
	}
	credService := services.NewCredentialService("")
	creds, err := credService.GetPlatformCredentials(s.tenantID, "tiktok")
	if err != nil {
		return nil, fmt.Errorf("failed to get TikTok credentials: %w", err)
	}
	if creds.AppKey == "" || creds.AppSecret == "" {
		return nil, fmt.Errorf("TikTok credentials not configured for tenant %s", s.tenantID)
	}
	tkClient := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	tkClient.SetCredentials(creds.AccessToken, creds.ShopCipher)
	return tkClient, nil
}

func (s *TiktokEscrowSyncService) fetchOrdersForSettlementWindow(
	ctx context.Context, client TiktokEscrowClient, month, year int,
) ([]tiktokPkg.TiktokOrder, error) {
	monthStart := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := monthStart.AddDate(0, 1, 0)
	startDate := monthStart.AddDate(0, 0, -21)

	req := tiktokPkg.OrderSearchRequest{
		OrderStatus:  "COMPLETED",
		CreateTimeGe: startDate.Unix(),
		CreateTimeLt: endDate.Unix(),
	}

	var allOrders []tiktokPkg.TiktokOrder
	pageToken := ""
	for {
		resp, err := client.SearchOrders(ctx, req, 100, pageToken)
		if err != nil {
			return nil, fmt.Errorf("SearchOrders: %w", err)
		}
		if resp.Code != 0 {
			return nil, fmt.Errorf("TikTok API error: %d - %s", resp.Code, resp.Message)
		}
		allOrders = append(allOrders, resp.Data.Orders...)
		pageToken = resp.Data.NextPageToken
		if pageToken == "" {
			break
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(200 * time.Millisecond):
		}
	}
	return allOrders, nil
}

func (s *TiktokEscrowSyncService) enrichOrdersWithDetails(
	ctx context.Context, client TiktokEscrowClient, orders []tiktokPkg.TiktokOrder,
) {
	const batchSize = 50
	for i := 0; i < len(orders); i += batchSize {
		end := min(i+batchSize, len(orders))
		ids := make([]string, 0, end-i)
		for _, o := range orders[i:end] {
			ids = append(ids, o.ID)
		}
		detail, err := client.GetOrderDetail(ctx, ids)
		if err != nil {
			log.Warn().Err(err).Int("batch", i/batchSize).
				Msg("[TiktokEscrowSync] GetOrderDetail failed, skipping enrichment")
			continue
		}
		detailMap := make(map[string]*tiktokPkg.OrderDetailData, len(detail.Data.Orders))
		for j := range detail.Data.Orders {
			detailMap[detail.Data.Orders[j].ID] = &detail.Data.Orders[j]
		}
		for j := i; j < end; j++ {
			d, ok := detailMap[orders[j].ID]
			if !ok {
				continue
			}
			enrichOrderFromDetail(&orders[j], d)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(200 * time.Millisecond):
		}
	}
}

func enrichOrderFromDetail(order *tiktokPkg.TiktokOrder, d *tiktokPkg.OrderDetailData) {
	if d.PaymentInfo != nil {
		if order.PaymentInfo.ShippingFee == "" {
			order.PaymentInfo.ShippingFee = d.PaymentInfo.ShippingFee
		}
		if order.PaymentInfo.TotalAmount == "" {
			order.PaymentInfo.TotalAmount = d.PaymentInfo.TotalAmount
		}
		if order.PaymentInfo.SubTotal == "" {
			order.PaymentInfo.SubTotal = d.PaymentInfo.SubTotal
		}
		if order.PaymentInfo.Currency == "" {
			order.PaymentInfo.Currency = d.PaymentInfo.Currency
		}
	}
	if d.RecipientAddress != nil && order.RecipientAddress.Name == "" {
		order.RecipientAddress = *d.RecipientAddress
	}
	if d.Status != "" && order.Status == "" {
		order.Status = d.Status
	}
	if d.CreateTime > 0 && order.CreateTime == 0 {
		order.CreateTime = d.CreateTime
	}
	if len(order.LineItems) == 0 && len(d.LineItems) > 0 {
		for _, li := range d.LineItems {
			order.LineItems = append(order.LineItems, tiktokPkg.TiktokOrderItem{
				ID:            li.ID,
				SkuID:         li.SkuID,
				SkuName:       li.SkuName,
				ProductID:     li.ProductID,
				ProductName:   li.ProductName,
				SellerSku:     li.SellerSku,
				Quantity:      li.Quantity,
				OriginalPrice: li.OriginalPrice,
				SalePrice:     li.SalePrice,
			})
		}
	}
}

func (s *TiktokEscrowSyncService) processOrdersConcurrent(
	ctx context.Context, client TiktokEscrowClient,
	orders []tiktokPkg.TiktokOrder, month, year int,
) (totalItems, processedOrders, failedOrders int, failedIDs []string) {
	var (
		mu         sync.Mutex
		wg         sync.WaitGroup
		processed  int64
		failed     int64
		itemsCount int64
		failIDs    []string
	)
	sem := make(chan struct{}, 5)

	for _, order := range orders {
		wg.Add(1)
		go func(o tiktokPkg.TiktokOrder) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			items, err := s.processSingleOrder(ctx, client, o, month, year)
			if err != nil {
				atomic.AddInt64(&failed, 1)
				mu.Lock()
				failIDs = append(failIDs, o.ID)
				mu.Unlock()
				return
			}
			if items > 0 {
				atomic.AddInt64(&itemsCount, int64(items))
				atomic.AddInt64(&processed, 1)
			}
		}(order)
	}
	wg.Wait()

	return int(itemsCount), int(processed), int(failed), failIDs
}

func (s *TiktokEscrowSyncService) processSingleOrder(
	ctx context.Context, client TiktokEscrowClient, order tiktokPkg.TiktokOrder, month, year int,
) (int, error) {
	transaction, err := client.GetOrderTransactions(ctx, order.ID)
	if err != nil {
		return 0, fmt.Errorf("TikTok v202501 order transactions failed for order %s: %w", order.ID, err)
	}
	if transaction == nil || transaction.Data.OrderID == "" {
		return 0, fmt.Errorf("no transaction data for order %s", order.ID)
	}
	if transaction.Code != 0 {
		return 0, fmt.Errorf("TikTok API error: %d - %s", transaction.Code, transaction.Message)
	}
	return s.saveEscrowOrder(ctx, order, transaction, month, year)
}

func (s *TiktokEscrowSyncService) saveSyncRecord(
	ctx context.Context, tables TiktokEscrowTables,
	month, year, totalOrders, failedOrders int, failedIDs []string,
) error {
	syncTable := tables.SyncTable
	now := time.Now()

	s.tenantDB.WithContext(ctx).Table(syncTable).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(nil)

	record := map[string]any{
		"id":               uuid.New().String(),
		"tenant_id":        s.tenantID,
		"month":            month,
		"year":             year,
		"total_orders":     totalOrders,
		"failed_orders":    failedOrders,
		"failed_order_ids": marshalFailedIDs(failedIDs),
		"synced_at":        now,
		"created_at":       now,
		"updated_at":       now,
	}
	return s.tenantDB.WithContext(ctx).Table(syncTable).Create(record).Error
}

func (s *TiktokEscrowSyncService) deleteMonthData(ctx context.Context, month, year int) error {
	if err := s.tenantDB.WithContext(ctx).
		Where("escrow_order_id IN (?)",
			s.tenantDB.WithContext(ctx).Model(&models.TiktokEscrowOrder{}).
				Select("id").
				Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year),
		).
		Delete(&models.TiktokEscrowItem{}).Error; err != nil {
		return fmt.Errorf("delete items: %w", err)
	}
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(&models.TiktokEscrowOrder{}).Error; err != nil {
		return fmt.Errorf("delete orders: %w", err)
	}
	if err := s.tenantDB.WithContext(ctx).
		Where("tenant_id = ? AND month = ? AND year = ?", s.tenantID, month, year).
		Delete(&models.TiktokEscrowSync{}).Error; err != nil {
		return fmt.Errorf("delete sync: %w", err)
	}
	return nil
}

func marshalFailedIDs(ids []string) *string {
	if len(ids) == 0 {
		return nil
	}
	data, err := json.Marshal(ids)
	if err != nil {
		return nil
	}
	s := string(data)
	return &s
}
