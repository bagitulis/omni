package testutils

import (
	"testing"
)

// TestMain applies the reaper policy and guarantees deterministic teardown of
// the shared postgres container (see container_cleanup.go).
func TestMain(m *testing.M) { RunTestMain(m) }
