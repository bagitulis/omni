package testutils

import (
	"context"
	"fmt"
	"testing"
	"time"

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

// SetupTestPostgres starts a postgres container and returns a GORM DB connection
// The container and connection should be cleaned up by calling TeardownTestPostgres
// In short mode, the test is skipped automatically.
func SetupTestPostgres(t *testing.T) *gorm.DB {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()

	// Create postgres container request
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

	// Start container
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	require.NoError(t, err, "failed to start postgres container")

	// Get connection string
	port, err := container.MappedPort(ctx, "5432/tcp")
	require.NoError(t, err, "failed to get mapped port")

	// On Windows with Docker Desktop, use 127.0.0.1 instead of container host
	// to avoid DNS resolution issues
	dsn := fmt.Sprintf("host=127.0.0.1 port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
		port.Port())

	// Connect to database with retry logic
	var db *gorm.DB
	var connErr error
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		db, connErr = gorm.Open(postgresdriver.Open(dsn), &gorm.Config{})
		if connErr == nil {
			break
		}
		if i < maxRetries-1 {
			time.Sleep(time.Second)
		}
	}
	require.NoError(t, connErr, "failed to connect to postgres after retries")

	// Store container reference for cleanup
	t.Cleanup(func() {
		TeardownTestPostgres(t, container)
	})

	return db
}

// TeardownTestPostgres terminates the postgres container
func TeardownTestPostgres(t *testing.T, container testcontainers.Container) {
	ctx := context.Background()
	err := container.Terminate(ctx)
	if err != nil {
		t.Logf("failed to terminate postgres container: %v", err)
	}
}

// SetupTestPostgresWithModels starts postgres and auto-migrates models
func SetupTestPostgresWithModels(t *testing.T, models ...interface{}) *gorm.DB {
	db := SetupTestPostgres(t)

	// Auto-migrate all provided models
	if len(models) > 0 {
		err := db.AutoMigrate(models...)
		require.NoError(t, err, "failed to auto-migrate models")
	}

	return db
}
