package testutils

import (
	"testing"

	"github.com/glebarez/sqlite"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// SetupTestSQLite opens a fresh in-memory SQLite database with the given models
// auto-migrated, and returns a GORM handle.
//
// Why this exists alongside SetupTestPostgres:
//   - Postgres-backed tests SKIP when no container runtime is reachable, so on a
//     machine without Docker/Podman running they provide zero signal.
//   - SQLite in-memory always runs, so pure data-access logic stays covered.
//
// Scope: this is for UNIT-level data-access tests. It is NOT a substitute for
// the Postgres suite. Postgres-specific behaviour — jsonb semantics,
// ON CONFLICT ... DO UPDATE, ILIKE, regex ~, gen_random_uuid(), partial indexes
// and schema/search_path handling — is NOT exercised here and must still be
// verified against Postgres.
//
// The database is named after the test, so each call gets its own database and
// tests never share state. When a single test needs several INDEPENDENT
// databases (for example to model two tenants that must not see each other's
// rows), use SetupNamedTestSQLite — otherwise the per-test name is reused and
// both handles point at the same in-memory database.
func SetupTestSQLite(t *testing.T, models ...any) *gorm.DB {
	t.Helper()
	return SetupNamedTestSQLite(t, sanitizeDBName(t.Name()), models...)
}

// SetupNamedTestSQLite opens an in-memory SQLite database under an explicit
// name.
//
// Needed when one test models multiple isolated stores: SetupTestSQLite alone
// would hand back the same database for every call in that test, which silently
// destroys any isolation assertion built on top of it.
func SetupNamedTestSQLite(t *testing.T, name string, models ...any) *gorm.DB {
	t.Helper()

	dsn := "file:" + name + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err, "failed to open in-memory sqlite")

	if len(models) > 0 {
		require.NoError(t, db.AutoMigrate(models...), "failed to auto-migrate models")
	}

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
	})

	return db
}
