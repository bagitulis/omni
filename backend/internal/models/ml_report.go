package models

import (
	"time"
)

// MLReport represents a generated ML analysis report
type MLReport struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	Platform    string    `gorm:"column:platform;index;not null" json:"platform"` // shopee, tiktok
	ReportType  string    `gorm:"column:report_type;not null" json:"report_type"` // full, executive
	PeriodStart time.Time `gorm:"column:period_start" json:"period_start"`
	PeriodEnd   time.Time `gorm:"column:period_end" json:"period_end"`
	PeriodLabel string    `gorm:"column:period_label;index" json:"period_label"`
	FilePath    string    `gorm:"column:file_path;not null" json:"file_path"`
	FileName    string    `gorm:"column:file_name;not null" json:"file_name"`
	FileSize    int64     `gorm:"column:file_size" json:"file_size"`
	Status      string    `gorm:"column:status;default:'completed'" json:"status"` // pending, completed, failed
	ErrorMsg    string    `gorm:"column:error_msg" json:"error_msg,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies the table name
func (MLReport) TableName() string {
	return "ml_reports"
}

// MLJob represents an async ML job
type MLJob struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	TenantID    string     `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
	JobType     string     `gorm:"column:job_type;not null" json:"job_type"` // generate_report
	Platform    string     `gorm:"column:platform;not null" json:"platform"`
	Status      string     `gorm:"column:status;default:'pending'" json:"status"` // pending, running, completed, failed
	Progress    int        `gorm:"column:progress;default:0" json:"progress"`     // 0-100
	ResultID    *uint      `gorm:"column:result_id" json:"result_id,omitempty"`   // MLReport ID
	ErrorMsg    string     `gorm:"column:error_msg" json:"error_msg,omitempty"`
	CreatedAt   time.Time  `gorm:"column:created_at" json:"created_at"`
	StartedAt   *time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	CompletedAt *time.Time `gorm:"column:completed_at" json:"completed_at,omitempty"`
}

// TableName specifies the table name
func (MLJob) TableName() string {
	return "ml_jobs"
}
