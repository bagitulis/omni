package sync

import (
	"context"
	"fmt"
	"time"
)

// BookingManager handles Shopee booking API operations
type BookingManager interface {
	GetBookingList(ctx context.Context, days int) ([]Booking, error)
	GetBookingDetails(ctx context.Context, bookingSNs []string) (map[string]Booking, error)
	GetPlatform() string
}

// Booking represents a booking order in the sync layer
type Booking struct {
	ID                   string        `json:"id"`
	TenantID             string        `json:"tenant_id"`
	ShopID               uint64        `json:"shop_id"`
	BookingSN            string        `json:"booking_sn"`
	OrderSN              string        `json:"order_sn"`
	BookingStatus        string        `json:"booking_status"`
	MatchStatus          string        `json:"match_status"`
	Region               string        `json:"region"`
	ShippingCarrier      string        `json:"shipping_carrier"`
	RecipientName        string        `json:"recipient_name"`
	RecipientPhone       string        `json:"recipient_phone"`
	RecipientAddressJSON string        `json:"recipient_address_json"`
	FulfillmentFlag      string        `json:"fulfillment_flag"`
	CreateTime           int64         `json:"create_time"`
	UpdateTime           int64         `json:"update_time"`
	PickupDoneTime       int64         `json:"pickup_done_time"`
	RawData              string        `json:"-"`
	SyncedAt             time.Time     `json:"synced_at"`
	ItemCount            int           `json:"item_count,omitempty"`
	HasParentOrder       bool          `json:"has_parent_order"`
	ParentOrderExists    bool          `json:"parent_order_exists"`
	ParentOrderStatus    string        `json:"parent_order_status,omitempty"`
	Items                []BookingItem `json:"items,omitempty" gorm:"-"`
}

// BookingItem represents an item within a booking in the sync layer
type BookingItem struct {
	ID                string  `json:"id"`
	TenantID          string  `json:"tenant_id"`
	ShopID            uint64  `json:"shop_id"`
	BookingSN         string  `json:"booking_sn"`
	LineKey           string  `json:"line_key"`
	ItemID            int64   `json:"item_id"`
	ModelID           int64   `json:"model_id"`
	ItemName          string  `json:"item_name"`
	ModelName         string  `json:"model_name"`
	ItemSku           string  `json:"item_sku"`
	ModelSku          string  `json:"model_sku"`
	SKU               string  `json:"sku"`
	Quantity          int     `json:"quantity"`
	Weight            float64 `json:"weight"`
	ProductLocationID string  `json:"product_location_id"`
	ImageURL          string  `json:"image_url"`
	RawData           string  `json:"-"`
}

// BookingRepository handles booking persistence
type BookingRepository interface {
	UpsertBookings(ctx context.Context, tenantID string, shopID uint64, bookings []Booking) error
	UpsertBookingWithItems(ctx context.Context, tenantID string, shopID uint64, booking Booking, items []BookingItem) error
	ReplaceBookingItems(ctx context.Context, tenantID string, shopID uint64, bookingSN string, items []BookingItem) error
	ListBookings(ctx context.Context, tenantID string, params BookingListParams) ([]Booking, int64, error)
	GetBookingDetail(ctx context.Context, tenantID string, shopID uint64, bookingSN string) (*Booking, error)
	ParentOrderExists(ctx context.Context, tenantID string, orderSN string) (bool, error)
}

// BookingListParams represents query parameters for listing bookings
type BookingListParams struct {
	Page          int    `json:"page"`
	PageSize      int    `json:"page_size"`
	Search        string `json:"search"`
	BookingStatus string `json:"booking_status"`
	MatchStatus   string `json:"match_status"`
	Platform      string `json:"platform"`
}

// BookingSyncResult represents the result of a booking sync operation
type BookingSyncResult struct {
	Platform    string   `json:"platform"`
	Success     bool     `json:"success"`
	Count       int      `json:"count"`
	FailedCount int      `json:"failed_count"`
	Errors      []string `json:"errors,omitempty"`
}

// GenerateBookingItemLineKey creates a deterministic unique key for a booking item.
// Format: item_id|model_id|item_sku|model_sku|item_name|model_name
// Empty components are preserved to ensure uniqueness.
func GenerateBookingItemLineKey(itemID int64, modelID int64, itemSku, modelSku, itemName, modelName string) string {
	return fmt.Sprintf("%d|%d|%s|%s|%s|%s", itemID, modelID, itemSku, modelSku, itemName, modelName)
}
