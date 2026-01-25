package models

import "time"

// AuditLog represents an audit log entry
type AuditLog struct {
	ID             string    `gorm:"primaryKey" json:"id"`
	TenantID       string    `gorm:"index" json:"tenantId"`
	Action         string    `gorm:"index;not null" json:"action"`
	UserID         string    `gorm:"index" json:"userId"`
	TargetUserID   string    `json:"targetUserId,omitempty"`
	TargetTenantID string    `json:"targetTenantId,omitempty"`
	Details        string    `json:"details,omitempty"` // JSON string
	Status         string    `gorm:"index;not null" json:"status"` // success, failed
	ErrorMessage   string    `json:"errorMessage,omitempty"`
	IPAddress      string    `json:"ipAddress,omitempty"`
	UserAgent      string    `json:"userAgent,omitempty"`
	CreatedAt      time.Time `gorm:"index" json:"createdAt"`
}

// TableName specifies the table name for GORM
func (AuditLog) TableName() string {
	return GetTableName("AuditLog")
}

// AuditAction constants
const (
	AuditActionUserCreated   = "USER_CREATED"
	AuditActionUserUpdated   = "USER_UPDATED"
	AuditActionUserDeleted   = "USER_DELETED"
	AuditActionUserLogin     = "USER_LOGIN"
	AuditActionUserLogout    = "USER_LOGOUT"
	AuditActionLoginFailed   = "LOGIN_FAILED"
	AuditActionPasswordReset = "PASSWORD_RESET"
	AuditActionTokenRefresh  = "TOKEN_REFRESH"
	AuditActionOAuthConnect  = "OAUTH_CONNECT"
	AuditActionOrderSync     = "ORDER_SYNC"
	AuditActionProductSync   = "PRODUCT_SYNC"
	AuditActionSettingChange = "SETTING_CHANGE"
)

// AuditStatus constants
const (
	AuditStatusSuccess = "success"
	AuditStatusFailed  = "failed"
)

// AuditLogEntry is used to create new audit log
type AuditLogEntry struct {
	TenantID       string
	Action         string
	UserID         string
	TargetUserID   string
	TargetTenantID string
	Details        map[string]interface{}
	Status         string
	ErrorMessage   string
	IPAddress      string
	UserAgent      string
}
