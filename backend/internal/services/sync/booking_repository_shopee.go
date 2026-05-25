package sync

import (
	"context"
	"errors"
	"fmt"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/utils/logger"
	"gorm.io/gorm"
)

var bookingRepoLogger = logger.Named("GormBookingRepository")

// GormBookingRepository implements BookingRepository using GORM
type GormBookingRepository struct {
	db       *gorm.DB
	tenantID string
}

// NewGormBookingRepository creates a new booking repository
func NewGormBookingRepository(db *gorm.DB, tenantID string) *GormBookingRepository {
	return &GormBookingRepository{
		db:       db,
		tenantID: tenantID,
	}
}

// UpsertBookings upserts multiple bookings, supporting idempotent re-sync.
// Existing rows are updated; new rows are created without duplication.
func (r *GormBookingRepository) UpsertBookings(ctx context.Context, tenantID string, shopID uint64, bookings []Booking) error {
	for _, b := range bookings {
		model := bookingToModel(b)
		model.TenantID = tenantID
		model.ShopID = shopID

		if err := r.db.WithContext(ctx).
			Where("tenant_id = ? AND shop_id = ? AND booking_sn = ?", tenantID, shopID, b.BookingSN).
			Assign(model).
			FirstOrCreate(&model).Error; err != nil {
			return fmt.Errorf("upsert bookings: %w", err)
		}
	}
	return nil
}

// ReplaceBookingItems replaces all items for a booking in a single transaction.
// Deletes existing items then inserts the new set.
func (r *GormBookingRepository) ReplaceBookingItems(ctx context.Context, tenantID string, shopID uint64, bookingSN string, items []BookingItem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing items
		if err := tx.
			Where("tenant_id = ? AND shop_id = ? AND booking_sn = ?", tenantID, shopID, bookingSN).
			Delete(&models.ShopeeBookingItem{}).Error; err != nil {
			return fmt.Errorf("replace booking items - delete: %w", err)
		}

		// Insert new items
		for _, item := range items {
			model := bookingItemToModel(item)
			model.TenantID = tenantID
			model.ShopID = shopID
			model.BookingSN = bookingSN

			if err := tx.Create(&model).Error; err != nil {
				return fmt.Errorf("replace booking items - insert: %w", err)
			}
		}

		return nil
	})
}

// ListBookings returns paginated bookings filtered by params.
// Returns (bookings, total_count, error). total_count is before pagination.
func (r *GormBookingRepository) ListBookings(ctx context.Context, tenantID string, params BookingListParams) ([]Booking, int64, error) {
	query := r.db.WithContext(ctx).
		Model(&models.ShopeeBooking{}).
		Where("tenant_id = ?", tenantID)

	if params.BookingStatus != "" {
		query = query.Where("booking_status = ?", params.BookingStatus)
	}
	if params.MatchStatus != "" {
		query = query.Where("match_status = ?", params.MatchStatus)
	}
	if params.Search != "" {
		search := "%" + params.Search + "%"
		query = query.Where(
			"(booking_sn LIKE ? OR order_sn LIKE ? OR recipient_name LIKE ?)",
			search, search, search,
		)
	}

	// Total count before pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("list bookings - count: %w", err)
	}

	// Pagination defaults
	page := params.Page
	if page < 1 {
		page = 1
	}
	pageSize := params.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	// Fetch paginated bookings
	var bookingModels []models.ShopeeBooking
	if err := query.
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Order("create_time DESC").
		Find(&bookingModels).Error; err != nil {
		return nil, 0, fmt.Errorf("list bookings - find: %w", err)
	}

	if len(bookingModels) == 0 {
		return []Booking{}, total, nil
	}

	// Batch query item counts for all returned bookings
	bookingSNs := make([]string, len(bookingModels))
	for i, m := range bookingModels {
		bookingSNs[i] = m.BookingSN
	}

	type itemCountRow struct {
		BookingSN string `gorm:"column:booking_sn"`
		Count     int64  `gorm:"column:cnt"`
	}
	var itemCounts []itemCountRow
	if err := r.db.WithContext(ctx).
		Model(&models.ShopeeBookingItem{}).
		Select("booking_sn, COUNT(*) as cnt").
		Where("tenant_id = ? AND booking_sn IN ?", tenantID, bookingSNs).
		Group("booking_sn").
		Find(&itemCounts).Error; err != nil {
		bookingRepoLogger.WithFields(map[string]interface{}{
			"tenant_id":    tenantID,
			"booking_sn_count": len(bookingSNs),
		}).Warn("Failed to fetch booking item counts: " + err.Error())
	}

	itemCountMap := make(map[string]int64)
	for _, ic := range itemCounts {
		itemCountMap[ic.BookingSN] = ic.Count
	}

	// Convert to sync-layer DTOs
	bookings := make([]Booking, 0, len(bookingModels))
	for _, m := range bookingModels {
		b := bookingFromModel(m)
		b.ItemCount = int(itemCountMap[m.BookingSN])
		b.HasParentOrder = m.OrderSN != ""
		bookings = append(bookings, b)
	}

	return bookings, total, nil
}

// GetBookingDetail returns a single booking with its items.
// Returns nil, nil if the booking is not found.
func (r *GormBookingRepository) GetBookingDetail(ctx context.Context, tenantID string, shopID uint64, bookingSN string) (*Booking, error) {
	var booking models.ShopeeBooking
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND shop_id = ? AND booking_sn = ?", tenantID, shopID, bookingSN).
		First(&booking).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("get booking detail: %w", err)
	}

	var itemModels []models.ShopeeBookingItem
	if err := r.db.WithContext(ctx).
		Where("tenant_id = ? AND shop_id = ? AND booking_sn = ?", tenantID, shopID, bookingSN).
		Find(&itemModels).Error; err != nil {
		return nil, fmt.Errorf("get booking detail - items: %w", err)
	}

	result := bookingFromModel(booking)
	result.Items = make([]BookingItem, len(itemModels))
	for i, im := range itemModels {
		result.Items[i] = bookingItemFromModel(im)
	}

	return &result, nil
}

// --- Model <-> DTO transform helpers ---

func bookingToModel(b Booking) models.ShopeeBooking {
	return models.ShopeeBooking{
		BookingSN:           b.BookingSN,
		OrderSN:             b.OrderSN,
		BookingStatus:       b.BookingStatus,
		MatchStatus:         b.MatchStatus,
		Region:              b.Region,
		ShippingCarrier:     b.ShippingCarrier,
		RecipientName:       b.RecipientName,
		RecipientPhone:      b.RecipientPhone,
		RecipientAddressJSON: b.RecipientAddressJSON,
		FulfillmentFlag:     b.FulfillmentFlag,
		CreateTime:          b.CreateTime,
		UpdateTime:          b.UpdateTime,
		PickupDoneTime:      b.PickupDoneTime,
		RawData:             b.RawData,
	}
}

func bookingFromModel(m models.ShopeeBooking) Booking {
	return Booking{
		ID:                  m.ID,
		TenantID:            m.TenantID,
		ShopID:              m.ShopID,
		BookingSN:           m.BookingSN,
		OrderSN:             m.OrderSN,
		BookingStatus:       m.BookingStatus,
		MatchStatus:         m.MatchStatus,
		Region:              m.Region,
		ShippingCarrier:     m.ShippingCarrier,
		RecipientName:       m.RecipientName,
		RecipientPhone:      m.RecipientPhone,
		RecipientAddressJSON: m.RecipientAddressJSON,
		FulfillmentFlag:     m.FulfillmentFlag,
		CreateTime:          m.CreateTime,
		UpdateTime:          m.UpdateTime,
		PickupDoneTime:      m.PickupDoneTime,
		RawData:             m.RawData,
		SyncedAt:            m.SyncedAt,
	}
}

func bookingItemToModel(item BookingItem) models.ShopeeBookingItem {
	return models.ShopeeBookingItem{
		ItemID:           item.ItemID,
		ModelID:          item.ModelID,
		ItemName:         item.ItemName,
		ModelName:        item.ModelName,
		ItemSku:          item.ItemSku,
		ModelSku:         item.ModelSku,
		SKU:              item.SKU,
		Quantity:         item.Quantity,
		Weight:           item.Weight,
		ProductLocationID: item.ProductLocationID,
		ImageURL:         item.ImageURL,
		RawData:          item.RawData,
	}
}

func bookingItemFromModel(m models.ShopeeBookingItem) BookingItem {
	return BookingItem{
		ID:               m.ID,
		TenantID:         m.TenantID,
		ShopID:           m.ShopID,
		BookingSN:        m.BookingSN,
		LineKey:          m.LineKey,
		ItemID:           m.ItemID,
		ModelID:          m.ModelID,
		ItemName:         m.ItemName,
		ModelName:        m.ModelName,
		ItemSku:          m.ItemSku,
		ModelSku:         m.ModelSku,
		SKU:              m.SKU,
		Quantity:         m.Quantity,
		Weight:           m.Weight,
		ProductLocationID: m.ProductLocationID,
		ImageURL:         m.ImageURL,
		RawData:          m.RawData,
	}
}
