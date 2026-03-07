// Package master_product provides background sync
package master_product

import (
	"context"
	"sync"
	"time"

	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// BackgroundSyncService handles periodic sync
type BackgroundSyncService struct {
	db            *gorm.DB
	repo          *repositories.MasterProductRepository
	syncInterval  time.Duration
	maxConcurrent int
	stopChan      chan struct{}
	wg            sync.WaitGroup
	running       bool
	mu            sync.Mutex
}

// NewBackgroundSyncService creates a new service
func NewBackgroundSyncService(db *gorm.DB, syncInterval time.Duration) *BackgroundSyncService {
	return &BackgroundSyncService{
		db:            db,
		repo:          repositories.NewMasterProductRepository(db),
		syncInterval:  syncInterval,
		maxConcurrent: 3,
		stopChan:      make(chan struct{}),
	}
}

// Start begins the background sync loop
func (s *BackgroundSyncService) Start() {
	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		return
	}
	s.running = true
	s.mu.Unlock()

	s.wg.Add(1)
	go s.syncLoop()

	log.Info().Dur("interval", s.syncInterval).Msg("Background sync started")
}

// Stop gracefully stops the sync
func (s *BackgroundSyncService) Stop() {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return
	}
	s.running = false
	s.mu.Unlock()

	close(s.stopChan)
	s.wg.Wait()
	log.Info().Msg("Background sync stopped")
}

func (s *BackgroundSyncService) syncLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.syncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.stopChan:
			return
		case <-ticker.C:
			s.runSyncCycle()
		}
	}
}

func (s *BackgroundSyncService) runSyncCycle() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// TODO: Implement actual sync logic — this is currently a no-op stub.
	// When implementing, add: fetch stale products → sync from marketplace APIs → update DB
	log.Warn().Msg("Background sync cycle triggered but sync logic is not yet implemented (no-op)")
	_ = ctx
}

// IsRunning returns sync status
func (s *BackgroundSyncService) IsRunning() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.running
}
