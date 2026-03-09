package handlers

import (
	"time"
)

// RouteExecutionConfig represents execution mode config for routes.
// NOTE: This is the handler-level DTO used by route_execution_config.go.
// The canonical model is models.RouteExecutionConfig (auto_function.go).
// This DTO intentionally has a different field set to serve the handler API shape.
type RouteExecutionConfig struct {
	ID            int64     `json:"id" gorm:"primaryKey;autoIncrement"`
	TenantID      string    `json:"-" gorm:"column:tenant_id;index"`
	RouteKey      string    `json:"route_key" gorm:"column:route_key;not null"`
	RouteName     string    `json:"route_name" gorm:"column:route_name;not null"`
	Description   string    `json:"description,omitempty" gorm:"column:description"`
	ExecutionMode string    `json:"execution_mode" gorm:"column:execution_mode;default:direct"` // queue or direct
	Priority      string    `json:"priority" gorm:"column:priority;default:normal"`             // low, normal, high
	Enabled       bool      `json:"enabled" gorm:"column:enabled;default:true"`
	Icon          string    `json:"icon,omitempty" gorm:"column:icon;default:📋"`
	Category      string    `json:"category,omitempty" gorm:"column:category;default:general"`
	CreatedAt     time.Time `json:"created_at" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"column:updated_at;autoUpdateTime"`
}

// TableName returns the PostgreSQL table name
func (RouteExecutionConfig) TableName() string {
	return "route_execution_config"
}

// RouteExecutionConfigRequest represents the request body for creating config
type RouteExecutionConfigRequest struct {
	RouteKey      string `json:"route_key" binding:"required"`
	RouteName     string `json:"route_name" binding:"required"`
	Description   string `json:"description"`
	ExecutionMode string `json:"execution_mode" binding:"required"`
	Priority      string `json:"priority"`
	Enabled       *bool  `json:"enabled"`
	Icon          string `json:"icon"`
	Category      string `json:"category"`
}

// RouteExecutionConfigUpdateRequest represents the request body for updating config
type RouteExecutionConfigUpdateRequest struct {
	RouteName     string `json:"route_name"`
	Description   string `json:"description"`
	ExecutionMode string `json:"execution_mode"`
	Priority      string `json:"priority"`
	Enabled       *bool  `json:"enabled"`
	Icon          string `json:"icon"`
	Category      string `json:"category"`
}
