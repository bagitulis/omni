// Package analytics provides base escrow sync utilities shared by platform-specific services
package analytics

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// BaseEscrowService contains common fields for escrow sync services
type BaseEscrowService struct {
	DB       *gorm.DB
	TenantID string
	Schema   string
	DBPath   string
}

// NewBaseEscrowService creates a base escrow service
func NewBaseEscrowService(db *gorm.DB, tenantID, dbPath string) BaseEscrowService {
	return BaseEscrowService{
		DB:       db,
		TenantID: tenantID,
		Schema:   fmt.Sprintf("tenant_%s", tenantID),
		DBPath:   dbPath,
	}
}

// Table returns table name with schema prefix
func (b *BaseEscrowService) Table(name string) string {
	return fmt.Sprintf("%s.%s", b.Schema, name)
}

// DeleteMonthDataGeneric deletes existing data for the month using provided table names
// orderTable: e.g., "shopee_escrow_orders" or "tiktok_escrow_orders"
// itemTable: e.g., "shopee_escrow_items" or "tiktok_escrow_items"
// syncTable: e.g., "shopee_escrow_sync" or "tiktok_escrow_sync"
// orderIDColumn: e.g., "order_sn" for Shopee, "order_id" for TikTok
func (b *BaseEscrowService) DeleteMonthDataGeneric(
	ctx context.Context,
	orderTable, itemTable, syncTable string,
	month, year int,
) error {
	// Get order IDs
	var orderIDs []string
	b.DB.WithContext(ctx).Table(b.Table(orderTable)).
		Where("tenant_id = ? AND month = ? AND year = ?", b.TenantID, month, year).
		Pluck("id", &orderIDs)

	// Delete items
	if len(orderIDs) > 0 {
		b.DB.WithContext(ctx).Table(b.Table(itemTable)).
			Where("escrow_order_id IN ?", orderIDs).
			Delete(nil)
	}

	// Delete orders
	b.DB.WithContext(ctx).Table(b.Table(orderTable)).
		Where("tenant_id = ? AND month = ? AND year = ?", b.TenantID, month, year).
		Delete(nil)

	// Delete sync record
	b.DB.WithContext(ctx).Table(b.Table(syncTable)).
		Where("tenant_id = ? AND month = ? AND year = ?", b.TenantID, month, year).
		Delete(nil)

	return nil
}

// ParseFloat safely parses a string to float64, returns 0 on error
func ParseFloat(s string) float64 {
	if s == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

// StringPtr returns a pointer to a string
func StringPtr(s string) *string {
	return &s
}

// Int64Ptr returns a pointer to an int64
func Int64Ptr(i int64) *int64 {
	return &i
}

// IntPtr returns a pointer to an int
func IntPtr(i int) *int {
	return &i
}

// EscrowSyncTables holds table names for escrow sync operations
type EscrowSyncTables struct {
	OrderTable string
	ItemTable  string
	SyncTable  string
}

// ShopeeEscrowTables returns table names for Shopee escrow
func ShopeeEscrowTables() EscrowSyncTables {
	return EscrowSyncTables{
		OrderTable: "shopee_escrow_orders",
		ItemTable:  "shopee_escrow_items",
		SyncTable:  "shopee_escrow_sync",
	}
}

// TiktokEscrowTables returns table names for TikTok escrow
func TiktokEscrowTables() EscrowSyncTables {
	return EscrowSyncTables{
		OrderTable: "tiktok_escrow_orders",
		ItemTable:  "tiktok_escrow_items",
		SyncTable:  "tiktok_escrow_sync",
	}
}
