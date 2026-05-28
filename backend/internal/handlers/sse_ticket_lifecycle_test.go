package handlers

import (
	"context"
	"testing"
	"time"
)

func TestSSECleanupLifecycleStartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	StartSSECleanup(ctx)
	StopSSECleanup()
	StopSSECleanup()

	StartSSECleanup(ctx)
	cancel()
	deadline := time.After(200 * time.Millisecond)
	for {
		sseLifecycleMu.Lock()
		alive := sseCleanupAlive
		sseLifecycleMu.Unlock()
		if !alive {
			return
		}
		select {
		case <-deadline:
			t.Fatal("SSE cleanup worker did not stop after context cancellation")
		case <-time.After(10 * time.Millisecond):
		}
	}
}
