package handlers

import "time"

// RouteConfig represents a route configuration
type RouteConfig struct {
	ID          string                 `json:"id"`
	TenantID    string                 `json:"tenant_id"`
	RouteID     string                 `json:"route_id"`
	Platform    string                 `json:"platform"`
	Category    string                 `json:"category"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Enabled     bool                   `json:"enabled"`
	Schedule    string                 `json:"schedule,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
	LastRun     *time.Time             `json:"last_run,omitempty"`
	NextRun     *time.Time             `json:"next_run,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// UpdateConfigRequest represents update config request
type UpdateConfigRequest struct {
	Name        string                 `json:"name,omitempty"`
	Description string                 `json:"description,omitempty"`
	Enabled     *bool                  `json:"enabled,omitempty"`
	Schedule    string                 `json:"schedule,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// CreateConfigRequest represents create config request
type CreateConfigRequest struct {
	RouteID     string                 `json:"route_id" binding:"required"`
	Platform    string                 `json:"platform" binding:"required"`
	Category    string                 `json:"category" binding:"required"`
	Name        string                 `json:"name" binding:"required"`
	Description string                 `json:"description,omitempty"`
	Enabled     bool                   `json:"enabled"`
	Schedule    string                 `json:"schedule,omitempty"`
	Config      map[string]interface{} `json:"config,omitempty"`
}

// RouteExecutionLog represents route execution log
type RouteExecutionLog struct {
	ID        string    `json:"id"`
	RouteID   string    `json:"route_id"`
	Status    string    `json:"status"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Duration  int64     `json:"duration"`
	Message   string    `json:"message,omitempty"`
	Error     string    `json:"error,omitempty"`
}
