package models

import "time"

// AutoFunctionConfig represents auto function configuration
// JSON tags use snake_case as per AGENTS.md standard
type AutoFunctionConfig struct {
	ID                     uint       `gorm:"primaryKey" json:"id"`
	Name                   string     `gorm:"column:name;not null" json:"name"`
	Enabled                bool       `gorm:"column:enabled;default:false" json:"enabled"`
	IntervalMinutes        int        `gorm:"column:interval_minutes;default:30" json:"interval_minutes"`
	StartTime              *string    `gorm:"column:start_time" json:"start_time,omitempty"`
	EndTime                *string    `gorm:"column:end_time" json:"end_time,omitempty"`
	LastExecuted           *time.Time `gorm:"column:last_executed" json:"last_executed,omitempty"`
	NextScheduledExecution *time.Time `gorm:"column:next_scheduled_execution" json:"next_scheduled_execution,omitempty"`
	CreatedAt              time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt              time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name
func (AutoFunctionConfig) TableName() string {
	return "auto_functions_config"
}

// AutoFunctionHistory represents auto function execution history
type AutoFunctionHistory struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	FunctionName string    `gorm:"column:function_name;not null" json:"function_name"`
	ExecutedAt   time.Time `gorm:"column:executed_at" json:"executed_at"`
	Status       string    `gorm:"column:status;not null" json:"status"` // success, failed
	ErrorMessage string    `gorm:"column:error_message;type:text" json:"error_message,omitempty"`
	DurationMs   int       `gorm:"column:duration_ms" json:"duration_ms"`
}

// TableName returns the table name
func (AutoFunctionHistory) TableName() string {
	return "auto_functions_history"
}

// RouteExecutionConfig represents route-specific execution config
type RouteExecutionConfig struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  string    `gorm:"column:tenant_id;index;not null" json:"-"`
	RouteKey  string    `gorm:"column:route_key;not null;uniqueIndex" json:"route_key"`
	RoutePath string    `gorm:"column:route_path;not null" json:"route_path"`
	IsEnabled bool      `gorm:"column:is_enabled;default:true" json:"is_enabled"`
	Mode      string    `gorm:"column:mode;default:normal" json:"mode"` // normal, priority, disabled
	Priority  int       `gorm:"column:priority;default:0" json:"priority"`
	RateLimit int       `gorm:"column:rate_limit" json:"rate_limit"` // Requests per minute
	Timeout   int       `gorm:"column:timeout" json:"timeout"`       // Seconds
	Category  string    `gorm:"column:category" json:"category"`     // Route category for grouping
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name
func (RouteExecutionConfig) TableName() string {
	return "route_execution_config"
}

// AutoFunctionRequest represents auto function creation/update request
// JSON tags use snake_case as per AGENTS.md standard
type AutoFunctionRequest struct {
	Name            string  `json:"name" binding:"required"`
	Enabled         bool    `json:"enabled"`
	IntervalMinutes int     `json:"interval_minutes" binding:"required,min=1"`
	StartTime       *string `json:"start_time,omitempty"`
	EndTime         *string `json:"end_time,omitempty"`
}
