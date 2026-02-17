package models

import "time"

// Allowed values for MarketplaceSyncHistory enum fields
var (
	AllowedSyncPlatforms  = []string{"shopee", "tiktok", "lazada"}
	AllowedSyncOperations = []string{"stock_update", "price_update", "wholesale_update", "mpq_update", "clone"}
	AllowedSyncStatuses   = []string{"success", "failed", "partial"}
)

// IsValidSyncPlatform checks if platform is in the allowed set
func IsValidSyncPlatform(v string) bool {
	for _, p := range AllowedSyncPlatforms {
		if v == p {
			return true
		}
	}
	return false
}

// IsValidSyncOperation checks if operation is in the allowed set
func IsValidSyncOperation(v string) bool {
	for _, o := range AllowedSyncOperations {
		if v == o {
			return true
		}
	}
	return false
}

// IsValidSyncStatus checks if status is in the allowed set
func IsValidSyncStatus(v string) bool {
	for _, s := range AllowedSyncStatuses {
		if v == s {
			return true
		}
	}
	return false
}

// MarketplaceSyncHistory tracks marketplace sync operations per SKU per platform
type MarketplaceSyncHistory struct {
	ID           string    `gorm:"primaryKey;column:id;size:255" json:"id"`
	TenantID     string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	SKU          string    `gorm:"column:sku;index;not null" json:"sku"`
	Platform     string    `gorm:"column:platform;index;not null" json:"platform"`   // shopee, tiktok, lazada
	Operation    string    `gorm:"column:operation;index;not null" json:"operation"` // stock_update, price_update, wholesale_update, mpq_update, clone
	Status       string    `gorm:"column:status;not null" json:"status"`             // success, failed, partial
	RequestData  *string   `gorm:"column:request_data;type:text" json:"request_data,omitempty"`
	ResponseData *string   `gorm:"column:response_data;type:text" json:"response_data,omitempty"`
	ErrorMessage *string   `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	CreatedAt    time.Time `gorm:"column:created_at;index" json:"created_at"`
}

// TableName returns the table name for MarketplaceSyncHistory
func (MarketplaceSyncHistory) TableName() string {
	return GetTableName("MarketplaceSyncHistory")
}

// MarketplaceSyncHistoryFilter holds query parameters for filtering sync history
type MarketplaceSyncHistoryFilter struct {
	TenantID  string `json:"tenant_id"`
	SKUSearch string `json:"sku_search,omitempty"`
	Platform  string `json:"platform,omitempty"`
	Operation string `json:"operation,omitempty"`
	Status    string `json:"status,omitempty"`
	DateFrom  string `json:"date_from,omitempty"`
	DateTo    string `json:"date_to,omitempty"`
	Page      int    `json:"page"`
	PageSize  int    `json:"page_size"`
}

// MarketplaceSyncHistoryListResult holds paginated list results
type MarketplaceSyncHistoryListResult struct {
	Entries  []MarketplaceSyncHistory `json:"entries"`
	Total    int64                    `json:"total"`
	Page     int                      `json:"page"`
	PageSize int                      `json:"page_size"`
}
