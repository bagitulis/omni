package inventory

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// ParallelBatchConfig configures parallel execution limits
type ParallelBatchConfig struct {
	MaxConcurrency int           // Total goroutines (default: 10)
	MaxPerPlatform int           // Per-platform concurrency (default: 5)
	ItemTimeout    time.Duration // Per-item timeout (default: 30s)
}

// DefaultParallelBatchConfig returns sensible defaults
func DefaultParallelBatchConfig() ParallelBatchConfig {
	return ParallelBatchConfig{
		MaxConcurrency: 10,
		MaxPerPlatform: 5,
		ItemTimeout:    30 * time.Second,
	}
}

// BatchWorkItem represents a single unit of work for parallel execution
type BatchWorkItem struct {
	Index     int
	SKU       string
	Platforms []string
}

// BatchWorkResult represents the result of a single work item
type BatchWorkResult struct {
	Index  int
	Result interface{}
}

// WorkerFunc is the function signature for batch work execution
type WorkerFunc func(ctx context.Context, item BatchWorkItem) interface{}

// RunParallelBatch executes work items in parallel with concurrency limiting.
// It uses a global semaphore to limit total goroutines.
// Results are returned in the same order as input items.
func RunParallelBatch(ctx context.Context, items []BatchWorkItem, config ParallelBatchConfig, worker WorkerFunc) []interface{} {
	if len(items) == 0 {
		return nil
	}

	// For very small batches, just run sequentially
	if len(items) <= 2 {
		results := make([]interface{}, len(items))
		for i, item := range items {
			itemCtx, cancel := context.WithTimeout(ctx, config.ItemTimeout)
			results[i] = worker(itemCtx, item)
			cancel()
		}
		return results
	}

	results := make([]interface{}, len(items))
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Global concurrency semaphore
	sem := make(chan struct{}, config.MaxConcurrency)

	for _, item := range items {
		wg.Add(1)
		go func(workItem BatchWorkItem) {
			defer wg.Done()

			// Acquire global semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				mu.Lock()
				results[workItem.Index] = map[string]interface{}{
					"sku":     workItem.SKU,
					"success": false,
					"error":   "context cancelled",
				}
				mu.Unlock()
				return
			}

			// Execute with per-item timeout
			itemCtx, cancel := context.WithTimeout(ctx, config.ItemTimeout)
			defer cancel()

			// Recover from panics in worker
			var result interface{}
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Error().
							Str("sku", workItem.SKU).
							Interface("panic", r).
							Msg("[ParallelBatch] Worker panicked")
						result = map[string]interface{}{
							"sku":     workItem.SKU,
							"success": false,
							"error":   "internal error during sync",
						}
					}
				}()
				result = worker(itemCtx, workItem)
			}()

			mu.Lock()
			results[workItem.Index] = result
			mu.Unlock()
		}(item)
	}

	wg.Wait()
	return results
}
