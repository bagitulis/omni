package sync

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/utils/logger"
)

const bookingSyncBatchSize = 50

var bookingServiceLogger = logger.Named("BookingSyncService")

// BookingSyncService orchestrates Shopee booking order sync and reads.
type BookingSyncService struct {
	manager    BookingManager
	repository BookingRepository
	tenantID   string
	shopID     uint64
}

// GetBookingSyncService creates a Shopee booking sync service for a tenant.
func GetBookingSyncService(tenantID string, coordinationService any) *BookingSyncService {
	service := &BookingSyncService{tenantID: tenantID}
	if tenantID == "" {
		bookingServiceLogger.Error("Cannot create booking sync service: missing tenant ID")
		return service
	}

	coordService := resolveBookingCoordinationService(tenantID, coordinationService)
	if coordService != nil {
		if err := coordService.InitializePlatforms(context.Background()); err != nil {
			bookingServiceLogger.WithTenantID(tenantID).Warn("Failed to initialize platform coordination service: " + err.Error())
		}
		if shopeeClient := coordService.GetShopeeClient(); shopeeClient != nil {
			service.manager = &ShopeeBookingManager{client: shopeeClient}
		}
		service.shopID = resolveBookingShopID(coordService)
	}

	db, err := config.GetTenantDBByID(tenantID)
	if err != nil {
		bookingServiceLogger.WithTenantID(tenantID).Error("Failed to get tenant database for booking repository: " + err.Error())
		return service
	}
	service.repository = NewGormBookingRepository(db, tenantID)
	return service
}

// ShopeeBookingManager adapts ShopeeAPIClient to BookingManager.
type ShopeeBookingManager struct {
	client *platform.ShopeeAPIClient
}

func (m *ShopeeBookingManager) GetBookingList(ctx context.Context, days int) ([]Booking, error) {
	if m.client == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}
	rawBookings, err := m.client.GetBookingList(ctx, days, bookingSyncBatchSize)
	if err != nil {
		return nil, err
	}

	bookings := make([]Booking, 0, len(rawBookings))
	for _, raw := range rawBookings {
		bookings = append(bookings, Booking{
			BookingSN:     getString(raw, "booking_sn"),
			OrderSN:       getString(raw, "order_sn"),
			BookingStatus: getString(raw, "booking_status"),
		})
	}
	return bookings, nil
}

func (m *ShopeeBookingManager) GetBookingDetails(ctx context.Context, bookingSNs []string) (map[string]Booking, error) {
	if m.client == nil {
		return nil, fmt.Errorf("shopee client not configured")
	}
	rawDetails, err := m.client.GetBookingDetails(ctx, bookingSNs)
	if err != nil {
		return nil, err
	}

	bookings := make(map[string]Booking, len(rawDetails))
	for _, raw := range rawDetails {
		booking, items, err := rawToBooking(raw)
		if err != nil {
			return bookings, err
		}
		booking.Items = items
		bookings[booking.BookingSN] = booking
	}
	return bookings, nil
}

func (m *ShopeeBookingManager) GetPlatform() string { return "shopee" }

// SyncBookings fetches booking list, hydrates details, and persists successes.
func (s *BookingSyncService) SyncBookings(ctx context.Context, days int) *BookingSyncResult {
	result := &BookingSyncResult{Platform: "shopee"}
	if s.manager == nil || s.repository == nil {
		result.Errors = append(result.Errors, "booking sync service is not fully configured")
		result.FailedCount = 1
		return result
	}

	list, err := s.manager.GetBookingList(ctx, days)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("booking list fetch failed: %s", err.Error()))
		result.FailedCount = 1
		return result
	}

	bookingSNs := collectBookingSNs(list)
	bookings := make([]Booking, 0, len(bookingSNs))
	for i := 0; i < len(bookingSNs); i += bookingSyncBatchSize {
		end := i + bookingSyncBatchSize
		end = min(end, len(bookingSNs))
		details, err := s.manager.GetBookingDetails(ctx, bookingSNs[i:end])
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("booking detail batch %d failed: %s", i/bookingSyncBatchSize+1, err.Error()))
			result.FailedCount += end - i
			continue
		}
		for _, bookingSN := range bookingSNs[i:end] {
			booking, ok := details[bookingSN]
			if !ok {
				result.Errors = append(result.Errors, fmt.Sprintf("booking detail missing for %s", bookingSN))
				result.FailedCount++
				continue
			}
			booking.TenantID = s.tenantID
			booking.ShopID = s.shopID
			booking.SyncedAt = time.Now()
			bookings = append(bookings, booking)
		}
	}

	for _, booking := range bookings {
		items := booking.Items
		booking.Items = nil
		if err := s.repository.UpsertBookingWithItems(ctx, s.tenantID, s.shopID, booking, items); err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("persist booking %s failed: %s", booking.BookingSN, err.Error()))
			result.FailedCount++
			continue
		}
		result.Count++
	}

	result.Success = len(result.Errors) == 0
	bookingServiceLogger.WithFields(map[string]any{"count": result.Count, "failed_count": result.FailedCount}).Info("Booking sync completed")
	return result
}

func (s *BookingSyncService) ListBookings(ctx context.Context, params BookingListParams) ([]Booking, int64, error) {
	if s.repository == nil {
		return nil, 0, fmt.Errorf("booking repository not configured")
	}
	return s.repository.ListBookings(ctx, s.tenantID, params)
}

func (s *BookingSyncService) GetBookingDetail(ctx context.Context, bookingSN string) (*Booking, []BookingItem, error) {
	if s.repository == nil {
		return nil, nil, fmt.Errorf("booking repository not configured")
	}
	booking, err := s.repository.GetBookingDetail(ctx, s.tenantID, s.shopID, bookingSN)
	if err != nil || booking == nil {
		return nil, nil, err
	}
	parentOrderExists, parentOrderStatus, err := s.resolveParentOrderStatus(ctx, booking.OrderSN)
	if err != nil {
		return nil, nil, err
	}
	booking.ParentOrderExists = parentOrderExists
	booking.ParentOrderStatus = parentOrderStatus
	return booking, booking.Items, nil
}

func (s *BookingSyncService) resolveParentOrderStatus(ctx context.Context, orderSN string) (bool, string, error) {
	if orderSN == "" {
		return false, "no_parent", nil
	}
	exists, err := s.repository.ParentOrderExists(ctx, s.tenantID, orderSN)
	if err != nil {
		return false, "", fmt.Errorf("resolve parent order status: %w", err)
	}
	if exists {
		return true, "synced", nil
	}
	return false, "not_synced", nil
}

func rawToBooking(raw map[string]any) (Booking, []BookingItem, error) {
	bookingSN := getString(raw, "booking_sn")
	if bookingSN == "" {
		return Booking{}, nil, fmt.Errorf("booking_sn is required")
	}
	recipientAddressJSON, recipientName, recipientPhone := parseRecipientAddress(raw)
	rawJSON, err := json.Marshal(raw)
	if err != nil {
		return Booking{}, nil, fmt.Errorf("marshal booking raw data: %w", err)
	}

	booking := Booking{
		BookingSN:            bookingSN,
		OrderSN:              getString(raw, "order_sn"),
		BookingStatus:        getString(raw, "booking_status"),
		MatchStatus:          getString(raw, "match_status"),
		Region:               getString(raw, "region"),
		ShippingCarrier:      getString(raw, "shipping_carrier"),
		RecipientName:        recipientName,
		RecipientPhone:       recipientPhone,
		RecipientAddressJSON: recipientAddressJSON,
		FulfillmentFlag:      getString(raw, "fulfillment_flag"),
		CreateTime:           getInt64(raw, "create_time"),
		UpdateTime:           getInt64(raw, "update_time"),
		PickupDoneTime:       getInt64(raw, "pickup_done_time"),
		RawData:              string(rawJSON),
	}
	return booking, parseBookingItems(raw, bookingSN), nil
}

func parseRecipientAddress(raw map[string]any) (string, string, string) {
	recipientAddress, ok := raw["recipient_address"].(map[string]any)
	if !ok {
		return "{}", "", ""
	}
	addressJSON, err := json.Marshal(recipientAddress)
	if err != nil {
		return "{}", getString(recipientAddress, "name"), getString(recipientAddress, "phone")
	}
	return string(addressJSON), getString(recipientAddress, "name"), getString(recipientAddress, "phone")
}

func parseBookingItems(raw map[string]any, bookingSN string) []BookingItem {
	rawItems, ok := raw["items"].([]any)
	if !ok {
		return []BookingItem{}
	}
	items := make([]BookingItem, 0, len(rawItems))
	for _, rawItem := range rawItems {
		itemMap, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		item := BookingItem{
			BookingSN:         bookingSN,
			ItemID:            getInt64(itemMap, "item_id"),
			ModelID:           getInt64(itemMap, "model_id"),
			ItemName:          getString(itemMap, "item_name"),
			ModelName:         getString(itemMap, "model_name"),
			ItemSku:           getString(itemMap, "item_sku"),
			ModelSku:          getString(itemMap, "model_sku"),
			Quantity:          getInt(itemMap, "quantity"),
			Weight:            getFloat64(itemMap, "weight"),
			ProductLocationID: getString(itemMap, "product_location_id"),
			ImageURL:          getString(itemMap, "image_info.image_url"),
		}
		item.SKU = item.ModelSku
		if item.SKU == "" {
			item.SKU = item.ItemSku
		}
		item.LineKey = GenerateBookingItemLineKey(item.ItemID, item.ModelID, item.ItemSku, item.ModelSku, item.ItemName, item.ModelName)
		if rawJSON, err := json.Marshal(itemMap); err == nil {
			item.RawData = string(rawJSON)
		}
		items = append(items, item)
	}
	return items
}

func collectBookingSNs(bookings []Booking) []string {
	bookingSNs := make([]string, 0, len(bookings))
	seen := make(map[string]bool, len(bookings))
	for _, booking := range bookings {
		if booking.BookingSN == "" || seen[booking.BookingSN] {
			continue
		}
		seen[booking.BookingSN] = true
		bookingSNs = append(bookingSNs, booking.BookingSN)
	}
	return bookingSNs
}

func resolveBookingCoordinationService(tenantID string, coordinationService any) *platform.PlatformCoordinationService {
	if typed, ok := coordinationService.(*platform.PlatformCoordinationService); ok && typed != nil {
		return typed
	}
	return platform.GetPlatformCoordinationService(tenantID)
}

func resolveBookingShopID(coordService *platform.PlatformCoordinationService) uint64 {
	configManager, err := coordService.GetConfigManager(platform.PlatformShopee)
	if err != nil {
		return 0
	}
	shopID, ok := configManager.GetConfig("shopId")
	if !ok || shopID == "" {
		return 0
	}
	parsed, err := strconv.ParseUint(shopID, 10, 64)
	if err != nil {
		return 0
	}
	return parsed
}
