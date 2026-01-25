package main

import "time"

// MigrationConfig holds migration configuration
type MigrationConfig struct {
	SQLitePath      string
	PostgresDSN     string
	TenantID        string
	BatchSize       int
	VerifyOnly      bool
	DryRun          bool
	ContinueOnError bool
}

// MigrationStats holds statistics for a table migration
type MigrationStats struct {
	TableName    string
	SourceRows   int64
	MigratedRows int64
	SkippedRows  int64
	ErrorRows    int64
	Duration     time.Duration
	Errors       []string
}

// MigrationReport holds the overall migration report
type MigrationReport struct {
	StartTime     time.Time
	EndTime       time.Time
	TenantID      string
	TotalTables   int
	SuccessTables int
	FailedTables  int
	Stats         []MigrationStats
}

// TableMapping maps SQLite tables to PostgreSQL tables
type TableMapping struct {
	SQLiteTable   string
	PostgresTable string
	Columns       []ColumnMapping
	HasTenantID   bool
}

// ColumnMapping maps SQLite columns to PostgreSQL columns
type ColumnMapping struct {
	SQLiteCol   string
	PostgresCol string
	Transform   func(interface{}) interface{}
}
