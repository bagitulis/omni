package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CleanZombieColumns removes columns from the database that no longer exist in GORM models.
// GORM AutoMigrate adds columns but never removes them. This function handles the cleanup.
// Safe: only drops columns that are NOT referenced by any model field.
func CleanZombieColumns(db *gorm.DB, models []interface{}) {
	for _, model := range models {
		cleanZombieColumnsForModel(db, model)
	}
}

func cleanZombieColumnsForModel(db *gorm.DB, model interface{}) {
	stmt := &gorm.Statement{DB: db}
	if err := stmt.Parse(model); err != nil {
		return
	}

	tableName := stmt.Schema.Table
	if tableName == "" {
		return
	}

	// Get expected columns from GORM model
	expectedCols := make(map[string]bool)
	for _, field := range stmt.Schema.Fields {
		if field.DBName != "" && field.DBName != "-" {
			expectedCols[field.DBName] = true
		}
	}

	if len(expectedCols) == 0 {
		return
	}

	// Get actual columns from database
	currentSchema := getCurrentSchema(db)
	if currentSchema == "" {
		return
	}

	type ColumnInfo struct {
		ColumnName string `gorm:"column:column_name"`
	}

	var actualCols []ColumnInfo
	err := db.Raw(`
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position
	`, currentSchema, tableName).Scan(&actualCols).Error
	if err != nil {
		return
	}

	// Find zombie columns (in DB but not in model)
	for _, col := range actualCols {
		if !expectedCols[col.ColumnName] {
			// Drop the zombie column
			dropSQL := fmt.Sprintf(
				`ALTER TABLE "%s"."%s" DROP COLUMN IF EXISTS "%s"`,
				currentSchema, tableName, col.ColumnName,
			)
			if err := db.Exec(dropSQL).Error; err != nil {
				log.Warn().
					Str("table", tableName).
					Str("column", col.ColumnName).
					Err(err).
					Msg("Failed to drop zombie column")
			} else {
				log.Info().
					Str("table", tableName).
					Str("column", col.ColumnName).
					Msg("Dropped zombie column")
			}
		}
	}
}

func getCurrentSchema(db *gorm.DB) string {
	var schema string
	db.Raw("SELECT current_schema()").Scan(&schema)
	return schema
}
