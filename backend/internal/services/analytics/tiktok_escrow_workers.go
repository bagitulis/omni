// Package analytics provides concurrent worker functions for TikTok escrow sync
package analytics

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// processOrdersWithProgress processes orders concurrently with progress reporting.
// Returns failedOrderIDs containing only orders that errored (not legitimately skipped orders).
func (s *TiktokEscrowSyncService) processOrdersWithProgress(
	ctx context.Context, client *tiktokPkg.Client,
	orders []tiktokPkg.TiktokOrder, month, year int, onProgress ProgressCallback,
) (totalItems, processedOrders, failedOrders int, cancelled bool, failedOrderIDs []string) {
	totalOrders := len(orders)
	var atomicItems, atomicProcessed, atomicFailed, atomicDone int64

	var failedIDsMu sync.Mutex
	var failedIDsList []string

	progressDone := make(chan struct{})
	if onProgress != nil {
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()
			for {
				select {
				case <-progressDone:
					return
				case <-ticker.C:
					done := int(atomic.LoadInt64(&atomicDone))
					proc := int(atomic.LoadInt64(&atomicProcessed))
					percent := 20 + (75 * done / max(totalOrders, 1))
					onProgress(percent, proc, totalOrders,
						fmt.Sprintf("Processing orders %d/%d (%d settled this month)...", done, totalOrders, proc))
				}
			}
		}()
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 5)

	for _, order := range orders {
		select {
		case <-ctx.Done():
			cancelled = true
			break
		default:
		}
		if cancelled {
			break
		}

		wg.Add(1)
		go func(o tiktokPkg.TiktokOrder) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			items, err := s.processSingleOrder(ctx, client, o, month, year)
			if err != nil {
				atomic.AddInt64(&atomicFailed, 1)
				failedIDsMu.Lock()
				failedIDsList = append(failedIDsList, o.ID)
				failedIDsMu.Unlock()
			} else if items > 0 {
				atomic.AddInt64(&atomicItems, int64(items))
				atomic.AddInt64(&atomicProcessed, 1)
			}
			atomic.AddInt64(&atomicDone, 1)
		}(order)
	}

	wg.Wait()
	close(progressDone)

	totalItems = int(atomicItems)
	processedOrders = int(atomicProcessed)
	failedOrders = int(atomicFailed)
	failedOrderIDs = failedIDsList
	return totalItems, processedOrders, failedOrders, cancelled, failedOrderIDs
}
