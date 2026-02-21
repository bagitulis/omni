package models

import "time"

// JobStatus represents job execution status
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// Job represents a queued job
// Matches PostgreSQL table: tenant_{name}.jobs
// Note: No tenant_id column - schema isolation handles multi-tenancy
type Job struct {
	ID              string     `gorm:"column:id;primaryKey" json:"id"`
	Type            string     `gorm:"column:type;not null" json:"type"`
	Status          JobStatus  `gorm:"column:status;not null;default:pending" json:"status"`
	Priority        string     `gorm:"column:priority;default:normal" json:"priority"`
	Data            string     `gorm:"column:data;type:jsonb;not null" json:"data"`
	ErrorMessage    string     `gorm:"column:error_message" json:"error_message,omitempty"`
	ProgressPercent int        `gorm:"column:progress_percent;default:0" json:"progress_percent"`
	ProgressMessage string     `gorm:"column:progress_message" json:"progress_message,omitempty"`
	TotalItems      int        `gorm:"column:total_items;default:0" json:"total_items"`
	ProcessedItems  int        `gorm:"column:processed_items;default:0" json:"processed_items"`
	ResultData      string     `gorm:"column:result_data;type:text" json:"result_data,omitempty"`
	StartedAt       *time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt     *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName returns the table name
func (Job) TableName() string {
	return "jobs"
}

// JobHistory represents completed job history
// Matches PostgreSQL table: tenant_{name}.job_history
// Note: No tenant_id column - schema isolation handles multi-tenancy
type JobHistory struct {
	ID           int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	JobID        string     `gorm:"column:job_id;not null" json:"job_id"`
	JobType      string     `gorm:"column:job_type" json:"job_type"`
	Status       JobStatus  `gorm:"column:status;not null" json:"status"`
	ErrorMessage string     `gorm:"column:error_message" json:"error_message,omitempty"`
	DurationMs   int        `gorm:"column:duration_ms" json:"duration_ms"`
	StartedAt    *time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt  *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
	CreatedAt    time.Time  `gorm:"column:created_at" json:"created_at"`
}

// TableName returns the table name
func (JobHistory) TableName() string {
	return "job_history"
}

// CreateJobRequest represents job creation request
type CreateJobRequest struct {
	ID       string `json:"id"`
	Type     string `json:"type" binding:"required"`
	Data     string `json:"data"`
	Priority string `json:"priority"`
}

// JobFilter represents job query filters
type JobFilter struct {
	Status   JobStatus `form:"status"`
	Type     string    `form:"type"`
	Page     int       `form:"page"`
	PageSize int       `form:"page_size"`
}
