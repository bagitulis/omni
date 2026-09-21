package testutils

import (
	"context"
	"os"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
)

// This file fixes the leaked-container problem found in the 2026-09-21
// storage audit. See container_cleanup_test.go for the full root-cause note.
//
// Summary: the testcontainers reaper (ryuk) cannot start on Podman/Windows —
// it bind-mounts the Windows named pipe //./pipe/docker_engine, which fails
// with `mkdir //./pipe: permission denied`. With no reaper, the shared
// postgres container was never removed, and every `go test` run leaked one
// container that kept its image pinned as "in use". Because image pruning
// skips anything a container references, the Podman retention policy then
// silently stopped reclaiming space (17 leaked containers observed).
//
// Two-part fix:
//  1. Disable the reaper on platforms where it cannot work (it otherwise makes
//     container creation FAIL rather than degrade gracefully).
//  2. Terminate the shared container deterministically from TestMain.

var (
	reaperPolicyOnce sync.Once
	// teardownTimeout bounds container termination so a stuck runtime cannot
	// hang the whole test binary forever.
	teardownTimeout = 60 * time.Second
)

// reaperShouldBeDisabled reports whether the testcontainers reaper must be
// turned off on this platform.
//
// The reaper is genuinely useful on Linux/Docker (it survives a hard test
// crash). It is disabled only where it provably cannot work: Windows hosts,
// where testcontainers mounts the Docker named pipe and Podman rejects it.
func reaperShouldBeDisabled() bool {
	return runtime.GOOS == "windows"
}

// applyReaperPolicy exports the reaper decision to the environment variable
// testcontainers reads, so container creation succeeds instead of failing.
//
// Note: this is called from TestMain, before any container is created.
func applyReaperPolicy() {
	if !reaperShouldBeDisabled() {
		return
	}
	if err := os.Setenv("TESTCONTAINERS_RYUK_DISABLED", "true"); err != nil {
		log.Warn().Err(err).Msg("could not disable testcontainers reaper")
	}
}

// EnsureReaperPolicy applies the reaper policy exactly once per process.
// Safe to call from TestMain; later calls are no-ops.
func EnsureReaperPolicy() {
	reaperPolicyOnce.Do(applyReaperPolicy)
}

// TeardownSharedContainer terminates the process-wide shared postgres container
// and resets package state so a subsequent call is a no-op.
//
// This is the deterministic replacement for the reaper: TestMain calls it after
// m.Run(). It is deliberately idempotent and safe to call when no container was
// ever created (no runtime available, or all tests skipped).
func TeardownSharedContainer(ctx context.Context) {
	container := sharedContainer
	if container == nil {
		return
	}

	// Clear state first so a concurrent/repeated call cannot double-terminate.
	sharedContainer = nil

	if ctx == nil {
		ctx = context.Background()
	}
	// Bound termination: Terminate can block on a wedged runtime.
	ctx, cancel := context.WithTimeout(ctx, teardownTimeout)
	defer cancel()

	if err := container.Terminate(ctx); err != nil {
		// Leftover containers are exactly what caused the disk leak, so make
		// the failure visible instead of silently swallowing it.
		log.Warn().Err(err).Msg("failed to terminate shared test container; " +
			"a leaked container may remain (check `podman ps -a`)")
		return
	}
	log.Info().Msg("terminated shared test postgres container")
}

// RunTestMain runs a package's tests with the shared-container reaper policy
// applied and deterministic teardown guaranteed.
//
// Go builds ONE test binary per package, so every package that calls
// SetupTestPostgres starts its own container. Without this, each of the 7
// such packages leaked one container per run — the mechanism behind the 17
// containers found in the 2026-09-21 storage audit.
//
// Usage in a package that creates containers:
//
//	func TestMain(m *testing.M) { testutils.RunTestMain(m) }
//
// Packages that do NOT create containers need no TestMain; the call is a
// harmless no-op if no container was ever started.
func RunTestMain(m *testing.M) {
	EnsureReaperPolicy()

	code := m.Run()

	// Runs even when tests fail, so a red run cannot leave a container behind.
	TeardownSharedContainer(context.Background())

	os.Exit(code)
}
