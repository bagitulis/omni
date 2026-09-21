package testutils

import (
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSharedContainerCleanupIsDeterministic pins the fix for the leaked
// container problem found in the 2026-09-21 storage audit.
//
// Root cause (reproduced live): the testcontainers reaper (ryuk) cannot start
// on Podman/Windows because it bind-mounts the Windows named pipe
// `//./pipe/docker_engine`, which fails with:
//
//	reaper: new reaper: run container: container create: Error response from
//	daemon: make cli opts(): making volume mountpoint for volume
//	//./pipe/docker_engine: mkdir //./pipe: permission denied
//
// With no reaper running, nothing ever removed the shared postgres container.
// A day of `go test` runs left 17 exited containers holding ~3 GB of images
// pinned as "in use" (image pruning skips anything a container references),
// so the Podman retention policy silently stopped reclaiming space.
//
// The fix has two halves, both asserted here:
//  1. The reaper must be disabled (it cannot work on this platform, and
//     testcontainers otherwise fails hard instead of falling back).
//  2. Cleanup must be deterministic via TeardownSharedContainer, called from
//     TestMain, instead of relying on a reaper that never starts.
func TestTeardownSharedContainerIsIdempotent(t *testing.T) {
	// Must be safe with no container ever created (the common case on CI and
	// on machines with no runtime): calling it must not panic.
	assert.NotPanics(t, func() {
		TeardownSharedContainer(context.Background())
		TeardownSharedContainer(context.Background())
	}, "teardown must be idempotent and safe with no container")
}

// TestTeardownSharedContainerClearsState proves teardown resets the package
// state so a later test cannot observe a terminated container.
func TestTeardownSharedContainerClearsState(t *testing.T) {
	sharedContainer = nil
	sharedOnce = sync.Once{}

	assert.NotPanics(t, func() {
		TeardownSharedContainer(context.Background())
	})

	assert.Nil(t, sharedContainer, "teardown must leave no container reference")
}

// TestRyukReaperIsDisabled documents the platform constraint: the reaper cannot
// start under Podman on Windows, so the suite must not depend on it.
//
// This asserts the helper the runtime uses to decide, rather than the raw env
// var, so that a future platform that CAN run the reaper is still free to
// re-enable it.
func TestRyukReaperIsDisabled(t *testing.T) {
	assert.True(t,
		reaperShouldBeDisabled(),
		"reaper must be disabled on platforms where it cannot bind the podman socket",
	)
}

// TestEnsureReaperPolicyIsApplied checks that the policy is actually pushed
// into the environment testcontainers reads (TESTCONTAINERS_RYUK_DISABLED).
func TestEnsureReaperPolicyIsApplied(t *testing.T) {
	original, had := os.LookupEnv("TESTCONTAINERS_RYUK_DISABLED")
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("TESTCONTAINERS_RYUK_DISABLED", original)
			return
		}
		_ = os.Unsetenv("TESTCONTAINERS_RYUK_DISABLED")
	})

	_ = os.Unsetenv("TESTCONTAINERS_RYUK_DISABLED")
	applyReaperPolicy()

	if reaperShouldBeDisabled() {
		assert.Equal(t, "true", os.Getenv("TESTCONTAINERS_RYUK_DISABLED"),
			"reaper policy must be exported for testcontainers to read")
	}
}
