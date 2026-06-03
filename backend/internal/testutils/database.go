package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	postgresdriver "gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// PostgresContainer holds the postgres container and GORM connection
type PostgresContainer struct {
	Container testcontainers.Container
	DB        *gorm.DB
	URI       string
}

// sharedContainer holds a process-wide shared postgres container.
// All tests share this single container; each test gets its own database.
var (
	sharedOnce      sync.Once
	sharedContainer testcontainers.Container
	sharedBaseDSN   string
	sharedPort      string
	sharedErr       error
)

// SetupTestPostgres starts (or reuses) a shared postgres container and returns
// a GORM DB connection to a per-test isolated database.
// The container lives for the entire test process; only the test database is
// dropped on cleanup. In short mode or when Docker is unavailable, the test
// is skipped.
func SetupTestPostgres(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	initSharedContainer(t)

	db, _ := createTestDatabase(t)
	return db
}

// TeardownTestPostgres terminates the postgres container.
// Kept for backward compatibility; the shared container is automatically
// cleaned up when the process exits (via testcontainers reaper).
func TeardownTestPostgres(t *testing.T, container testcontainers.Container) {
	// No-op for shared container - it lives for the process lifetime.
	// Individual test databases are cleaned up by t.Cleanup in SetupTestPostgres.
}

// SetupTestPostgresWithModels starts postgres and auto-migrates models
func SetupTestPostgresWithModels(t *testing.T, models ...interface{}) *gorm.DB {
	db := SetupTestPostgres(t)

	if len(models) > 0 {
		err := db.AutoMigrate(models...)
		require.NoError(t, err, "failed to auto-migrate models")
	}

	return db
}

// initSharedContainer creates the shared postgres container exactly once.
func initSharedContainer(t *testing.T) {
	t.Helper()

	sharedOnce.Do(func() {
		if !isDockerAvailable() {
			sharedErr = fmt.Errorf("Docker is not available on this platform")
			return
		}

		ctx := context.Background()

		req := testcontainers.ContainerRequest{
			Image:        "postgres:15-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_USER":     "testuser",
				"POSTGRES_PASSWORD": "testpass",
				"POSTGRES_DB":       "testdb",
			},
			WaitingFor: wait.ForAll(
				wait.ForLog("database system is ready to accept connections"),
				wait.ForListeningPort("5432/tcp"),
			),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			sharedErr = fmt.Errorf("failed to start shared postgres container: %w", err)
			return
		}

		port, err := container.MappedPort(ctx, "5432/tcp")
		if err != nil {
			sharedErr = fmt.Errorf("failed to get mapped port: %w", err)
			return
		}

		sharedContainer = container
		sharedPort = port.Port()
		sharedBaseDSN = fmt.Sprintf(
			"host=127.0.0.1 port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
			sharedPort,
		)
	})

	require.NoError(t, sharedErr, "failed to initialize shared postgres container")
}

// createTestDatabase creates a unique database for this test and registers
// cleanup to drop it when the test finishes.
func createTestDatabase(t *testing.T) (*gorm.DB, string) {
	t.Helper()

	adminDB, err := sql.Open("postgres", sharedBaseDSN)
	require.NoError(t, err, "failed to open admin connection")
	require.NoError(t, adminDB.Ping(), "failed to ping admin connection")

	dbName := sanitizeDBName(t.Name())

	// Drop any leftover database from a previous run (WITH FORCE kills connections).
	_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbName))

	// Create the test database.
	_, err = adminDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err, "failed to create test database %s", dbName)

	testDSN := fmt.Sprintf(
		"host=127.0.0.1 port=%s user=testuser password=testpass dbname=%s sslmode=disable",
		sharedPort, dbName,
	)

	var db *gorm.DB
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgresdriver.Open(testDSN), &gorm.Config{})
		if err == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Second)
		}
	}
	require.NoError(t, err, "failed to connect to test database after retries")

	t.Cleanup(func() {
		if sqlDB, dbErr := db.DB(); dbErr == nil && sqlDB != nil {
			_ = sqlDB.Close()
		}
		_, _ = adminDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s WITH (FORCE)", dbName))
		_ = adminDB.Close()
	})

	return db, dbName
}

// sanitizeDBName turns a Go test name into a valid PostgreSQL identifier.
func sanitizeDBName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	result := b.String()

	if len(result) > 60 {
		result = result[:60]
	}

	if len(result) > 0 && unicode.IsDigit(rune(result[0])) {
		result = "t_" + result
	}

	return result
}

// isDockerAvailable checks if Docker is running by executing docker ps
func isDockerAvailable() bool {
	cmd := exec.Command("docker", "ps")
	return cmd.Run() == nil
}
