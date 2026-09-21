package testutils

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"
	"unicode"

	_ "github.com/lib/pq" // Register "postgres" driver for admin sql.Open used by test-database bootstrap
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
// dropped on cleanup.
//
// The test is SKIPPED (never failed) when either:
//   - short mode is active, or
//   - no reachable container runtime is available.
//
// The runtime check lives here rather than only in the per-package guard so that
// every caller inherits it. Callers outside this package (for example
// services/notification_service_test.go) previously failed outright on a machine
// with no running container runtime instead of skipping.
func SetupTestPostgres(t *testing.T) *gorm.DB {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoContainerRuntime(t)

	initSharedContainer(t)

	db, _ := createTestDatabase(t)
	return db
}

// skipIfNoContainerRuntime skips the calling test unless a container runtime is
// installed AND reachable.
//
// Two distinct reasons are reported so the cause is obvious from the test log:
// no CLI at all, versus a CLI whose VM/daemon is not running.
func skipIfNoContainerRuntime(t *testing.T) {
	t.Helper()

	if resolveContainerCLI() == "" {
		t.Skip("skipping integration test: no container runtime (docker/podman) found on PATH")
	}

	if !isContainerRuntimeAvailable() {
		t.Skip("skipping integration test: container runtime found but not reachable " +
			"(start it with `podman machine start` or launch Docker Desktop)")
	}
}

// TeardownTestPostgres is retained for backward compatibility.
//
// It is intentionally a no-op: the container is process-wide, so per-test
// teardown would break every other test still using it. Real cleanup happens
// once per process in TestMain via TeardownSharedContainer (see
// container_cleanup.go) — NOT via the testcontainers reaper, which cannot
// start on Podman/Windows.
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

// startContainerSafely calls the testcontainers constructor and converts panics
// into errors.
//
// testcontainers-go v0.40.0 PANICS rather than returning an error when it cannot
// determine a usable Docker/container host (see internal/core.MustExtractDockerHost,
// which panics with "rootless Docker is not supported on Windows" on some Windows
// setups). Without this recovery the panic escapes the test binary and aborts the
// whole run, so a missing runtime looks like a crash instead of a skip.
func startContainerSafely(
	ctx context.Context,
	req testcontainers.ContainerRequest,
) (container testcontainers.Container, err error) {
	defer func() {
		if r := recover(); r != nil {
			container = nil
			err = fmt.Errorf("testcontainers could not initialise a container host: %v", r)
		}
	}()

	return testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
}

// initSharedContainer creates the shared postgres container exactly once.
func initSharedContainer(t *testing.T) {
	t.Helper()

	sharedOnce.Do(func() {
		if !isContainerRuntimeAvailable() {
			sharedErr = fmt.Errorf(
				"no container runtime available: install Docker or Podman, " +
					"or set CONTAINER_RUNTIME=docker|podman",
			)
			return
		}

		// Must run before the first container is created: on platforms where
		// the reaper cannot bind the runtime socket it makes container
		// creation FAIL instead of degrading gracefully (see container_cleanup.go).
		EnsureReaperPolicy()

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

		container, err := startContainerSafely(ctx, req)
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
// PostgreSQL folds unquoted identifiers to lowercase, so we normalize to
// lowercase up-front. Otherwise a CREATE DATABASE MyName followed by a
// connection to "MyName" would fail with "database does not exist".
func sanitizeDBName(name string) string {
	var b strings.Builder
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' {
			b.WriteRune(r)
		} else {
			b.WriteRune('_')
		}
	}
	result := strings.ToLower(b.String())

	if len(result) > 60 {
		result = result[:60]
	}

	if len(result) > 0 && unicode.IsDigit(rune(result[0])) {
		result = "t_" + result
	}

	return result
}

// resolveContainerCLI determines which container CLI to use for testcontainers.
//
// Priority matches scripts/python-build/omni_build/container_runtime.py:
//
//	CONTAINER_RUNTIME env var > docker (if on PATH) > podman (if on PATH) > ""
//
// An empty return means no runtime is available; callers should skip rather
// than fail. Detection is deliberately runtime-based, NOT OS-based: a Windows
// host running Docker Desktop or Podman can run these tests fine, and keying
// off runtime.GOOS would silently skip them.
func resolveContainerCLI() string {
	if override := strings.TrimSpace(os.Getenv("CONTAINER_RUNTIME")); override != "" {
		return override
	}

	for _, candidate := range []string{"docker", "podman"} {
		if _, err := exec.LookPath(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

// isContainerRuntimeAvailable reports whether a usable container runtime exists.
//
// Two conditions must hold:
//  1. A CLI is resolvable (see resolveContainerCLI), and
//  2. that CLI can actually reach a container host.
//
// The second check matters on Windows: Podman and Docker Desktop both ship a CLI
// that is present on PATH before the backing VM is started. Without a liveness
// probe, `isContainerRuntimeAvailable()` returns true, the test proceeds, and
// testcontainers then fails (or panics) on a machine that simply has no running
// VM — turning an expected skip into a red test run.
//
// The probe is `info`, mirroring scripts/python-build/omni_build/container_runtime.py,
// which uses the same command in ContainerRuntime.is_running().
func isContainerRuntimeAvailable() bool {
	cli := resolveContainerCLI()
	if cli == "" {
		return false
	}

	// Fast path: the CLI may not be on PATH even when CONTAINER_RUNTIME names it
	// (e.g. Podman installed to Program Files but not added to PATH).
	path, err := exec.LookPath(cli)
	if err != nil {
		path = cli
	}

	ctx, cancel := context.WithTimeout(context.Background(), runtimeProbeTimeout)
	defer cancel()

	return exec.CommandContext(ctx, path, "info").Run() == nil
}

// runtimeProbeTimeout bounds the `info` liveness probe. Starting a stopped
// Podman machine or Docker daemon is not attempted here: this helper only
// answers "can I use a runtime right now", so a slow or absent host must not
// stall the test suite.
const runtimeProbeTimeout = 10 * time.Second
