package master_product

import (
	"testing"

	"github.com/omni/backend/internal/testutils"
)

// TestMain wires deterministic cleanup of the shared testcontainers postgres
// container. Without it each `go test` run leaks one container that pins its
// image as "in use" (see testutils/container_cleanup.go).
func TestMain(m *testing.M) { testutils.RunTestMain(m) }
