package handlers

// SSE ticket system: exchange JWT for a short-lived one-time ticket.
// This prevents JWT exposure in SSE URL query params.

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// SSETicket represents a short-lived one-time SSE auth ticket.
type SSETicket struct {
	UserID    string
	TenantID  string
	ExpiresAt time.Time
}

// SSETicketStore is an in-memory store for SSE tickets.
type SSETicketStore struct {
	mu      sync.RWMutex
	tickets map[string]*SSETicket
}

var sseTicketStore = &SSETicketStore{
	tickets: make(map[string]*SSETicket),
}

var (
	sseShutdown       = make(chan struct{})
	sseLifecycleMu    sync.Mutex
	sseCleanupAlive   bool
	sseShutdownClosed bool
)

// StartSSECleanup starts lifecycle-managed cleanup for expired SSE tickets.
func StartSSECleanup(ctx context.Context) {
	sseLifecycleMu.Lock()
	defer sseLifecycleMu.Unlock()

	if sseCleanupAlive {
		return
	}
	sseShutdown = make(chan struct{})
	sseCleanupAlive = true
	sseShutdownClosed = false
	shutdown := sseShutdown

	go func() {
		defer func() {
			sseLifecycleMu.Lock()
			sseCleanupAlive = false
			sseShutdownClosed = false
			sseLifecycleMu.Unlock()
		}()

		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				sseTicketStore.cleanup()
			case <-ctx.Done():
				return
			case <-shutdown:
				return
			}
		}
	}()
}

// StopSSECleanup stops the SSE ticket cleanup worker.
func StopSSECleanup() {
	sseLifecycleMu.Lock()
	defer sseLifecycleMu.Unlock()

	if !sseCleanupAlive || sseShutdownClosed {
		return
	}
	close(sseShutdown)
	sseShutdownClosed = true
}

// Create generates a new one-time ticket for the given user/tenant.
func (s *SSETicketStore) Create(userID, tenantID string) string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	ticket := hex.EncodeToString(b)

	s.mu.Lock()
	s.tickets[ticket] = &SSETicket{
		UserID:    userID,
		TenantID:  tenantID,
		ExpiresAt: time.Now().Add(30 * time.Second),
	}
	s.mu.Unlock()
	return ticket
}

// Validate checks and consumes a ticket (one-time use).
func (s *SSETicketStore) Validate(ticket string) (*SSETicket, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t, exists := s.tickets[ticket]
	if !exists {
		return nil, false
	}

	// One-time use: delete immediately
	delete(s.tickets, ticket)

	if time.Now().After(t.ExpiresAt) {
		return nil, false
	}

	return t, true
}

func (s *SSETicketStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for k, v := range s.tickets {
		if now.After(v.ExpiresAt) {
			delete(s.tickets, k)
		}
	}
}

// CreateSSETicket handles POST /api/auth/sse-ticket.
// Requires Auth middleware (JWT validated).
func CreateSSETicket(c *gin.Context) {
	userID := c.GetString("userID")
	tenantID := c.GetString("tenant_id")

	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}

	ticket := sseTicketStore.Create(userID, tenantID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"ticket": ticket,
		},
	})
}

// ValidateSSETicket validates a ticket and returns user info.
// Used by the notification stream handler.
func ValidateSSETicket(ticket string) (userID string, tenantID string, valid bool) {
	t, ok := sseTicketStore.Validate(ticket)
	if !ok {
		return "", "", false
	}
	return t.UserID, t.TenantID, true
}
