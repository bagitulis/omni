package tiktok

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	tiktokPkg "github.com/omni/backend/pkg/tiktok"
	"github.com/rs/zerolog"
)

// fetchProductPageWithRetry calls SearchProductsV202502 with exponential backoff retry.
// Retries up to 3 times with base delay 2s, max delay 30s, and jitter.
func (s *SyncService) fetchProductPageWithRetry(ctx context.Context, pageToken string) (*tiktokPkg.SearchProductsResponse, error) {
	const maxRetries = 3
	const baseDelay = 2 * time.Second
	const maxDelay = 30 * time.Second

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			delay := baseDelay * time.Duration(1<<uint(attempt-1))
			if delay > maxDelay {
				delay = maxDelay
			}
			jitter := time.Duration(rand.Intn(500)) * time.Millisecond
			delay += jitter

			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry backoff: %w", ctx.Err())
			case <-time.After(delay):
			}
		}

		resp, err := s.client.SearchProductsV202502("ACTIVATE", 100, pageToken)
		if err == nil {
			return resp, nil
		}

		lastErr = err
		zerolog.Ctx(ctx).Warn().
			Int("attempt", attempt+1).
			Int("max_retries", maxRetries).
			Err(err).
			Msg("TikTok SearchProducts API call failed, retrying")
	}

	return nil, fmt.Errorf("SearchProductsV202502 failed after %d retries: %w", maxRetries, lastErr)
}
