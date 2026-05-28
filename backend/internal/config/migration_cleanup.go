package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// CleanZombieColumns REMOVED from active migration (2026-05-28).
//
// RETIRED: This function is no longer called from MigrateSystemDatabase or
// MigrateTenantDatabase. It is kept here as reference only.
//
// Reason for retirement:
//   - GORM AutoMigrate is additive-only by design (adds columns, never drops)
//   - Column ownership across 60+ model types is unproven
//   - Destructive DROP COLUMN requires disposable DB verification first
//   - GORM schema parsing with embedded structs/associations has edge cases
//
// To re-enable: Write a disposable DB integration test that proves cleanup
// is idempotent, safe, and the app boots correctly afterward.
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

	// Get expected columns from GORM model (only real DB columns, not relations)
	expectedCols := make(map[string]bool)
	for _, field := range stmt.Schema.Fields {
		if field.DBName != "" && field.DBName != "-" {
			expectedCols[field.DBName] = true
		}
	}

	// SAFETY: If model has very few columns, schema parsing likely failed — skip
	if len(expectedCols) < 3 {
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

	// SAFETY: If DB has fewer columns than model expects, something is wrong — skip
	if len(actualCols) < len(expectedCols) {
		return
	}

	// Find zombie columns (in DB but not in model)
	dropped := 0
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
				dropped++
				log.Info().
					Str("table", tableName).
					Str("column", col.ColumnName).
					Msg("Dropped zombie column")
			}
		}
	}
	if dropped > 0 {
		log.Info().Str("table", tableName).Int("dropped", dropped).Msg("Zombie column cleanup complete")
	}
}

func getCurrentSchema(db *gorm.DB) string {
	var schema string
	db.Raw("SELECT current_schema()").Scan(&schema)
	return schema
}
