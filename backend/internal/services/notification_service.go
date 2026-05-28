package services

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
)

// SSEClient represents a connected SSE client
type SSEClient chan string

// SSEManager manages real-time notification streams for all tenants
type SSEManager struct {
	// Map of tenantID -> list of clients
	clients map[string][]SSEClient
	mu      sync.RWMutex
}

var globalSSEManager = &SSEManager{
	clients: make(map[string][]SSEClient),
}

// NotificationService manages in-app notifications per tenant.
type NotificationService struct {
	repo     *repositories.NotificationRepository
	tenantID string
}

// NewNotificationService creates a new notification service.
func NewNotificationService(repo *repositories.NotificationRepository) *NotificationService {
	return &NotificationService{repo: repo}
}

// WithTenant sets the tenant ID for SSE routing
func (s *NotificationService) WithTenant(tenantID string) *NotificationService {
	s.tenantID = tenantID
	return s
}

// Push creates a new notification in the database and broadcasts it via SSE.
func (s *NotificationService) Push(notifType, category, title, message, actionURL string) (*models.Notification, error) {
	notif := &models.Notification{
		Type:      notifType,
		Category:  category,
		Title:     title,
		Message:   message,
		ActionURL: actionURL,
		Read:      false,
		CreatedAt: time.Now(),
	}

	// 1. Save to Database
	if err := s.repo.Create(context.Background(), notif); err != nil {
		log.Error().Err(err).Str("title", title).Msg("Failed to push notification to DB")
		return nil, err
	}

	// 2. Broadcast via SSE if tenantID is known
	if s.tenantID != "" {
		s.broadcast(notif)
	}
	return notif, nil
}

// PushJobResult creates a notification from a completed job using standardized helpers.
func (s *NotificationService) PushJobResult(job *models.Job, success bool, detail string) error {
	notifType := models.NotifTypeSuccess
	if !success {
		notifType = models.NotifTypeError
	}

	title := models.JobTypeToTitle(job.Type, success)
	category := models.JobTypeToCategory(job.Type)

	s.Push(
		notifType,
		category,
		title,
		detail,
		"", // ActionURL can be added based on category later
	)
	return nil
}

// broadcast sends the notification to all active SSE connections for the tenant
func (s *NotificationService) broadcast(notif *models.Notification) {
	// Copy client slice under lock to avoid race with UnregisterClient
	globalSSEManager.mu.RLock()
	original, ok := globalSSEManager.clients[s.tenantID]
	if !ok || len(original) == 0 {
		globalSSEManager.mu.RUnlock()
		return
	}
	// Snapshot the slice so we can iterate safely after releasing the lock
	clients := make([]SSEClient, len(original))
	copy(clients, original)
	globalSSEManager.mu.RUnlock()

	data, err := json.Marshal(notif)
	if err != nil {
		return
	}

	event := fmt.Sprintf("data: %s\n\n", string(data))

	log.Info().Str("tenant", s.tenantID).Int("clients", len(clients)).Msg("Broadcasting notification via SSE")

	for _, client := range clients {
		select {
		case client <- event:
		default:
			log.Warn().Str("tenant", s.tenantID).Msg("[SSE] Client buffer full, notification dropped")
		}
	}
}

// RegisterClient registers a new SSE client for a tenant
func (s *NotificationService) RegisterClient(tenantID string) SSEClient {
	client := make(SSEClient, 50)
	globalSSEManager.mu.Lock()
	globalSSEManager.clients[tenantID] = append(globalSSEManager.clients[tenantID], client)
	globalSSEManager.mu.Unlock()
	return client
}

// UnregisterClient removes an SSE client
func (s *NotificationService) UnregisterClient(tenantID string, client SSEClient) {
	globalSSEManager.mu.Lock()
	defer globalSSEManager.mu.Unlock()

	clients := globalSSEManager.clients[tenantID]
	for i, c := range clients {
		if c == client {
			globalSSEManager.clients[tenantID] = append(clients[:i], clients[i+1:]...)
			close(c)
			break
		}
	}
}

// List returns recent notifications.
func (s *NotificationService) List(limit int, sinceID int64, unreadOnly bool) ([]models.Notification, error) {
	return s.repo.List(context.Background(), limit, sinceID, unreadOnly)
}

// GetByID returns a single notification by ID.
func (s *NotificationService) GetByID(id int64) (*models.Notification, error) {
	return s.repo.GetByID(context.Background(), id)
}

// UnreadCount returns the number of unread notifications.
func (s *NotificationService) UnreadCount() (int64, error) {
	return s.repo.UnreadCount(context.Background())
}

// MarkAsRead marks a single notification as read.
func (s *NotificationService) MarkAsRead(id int64) error {
	return s.repo.MarkAsRead(context.Background(), id)
}

// MarkAllAsRead marks all notifications as read.
func (s *NotificationService) MarkAllAsRead() error {
	return s.repo.MarkAllAsRead(context.Background())
}

// Delete removes a single notification.
func (s *NotificationService) Delete(id int64) error {
	return s.repo.Delete(context.Background(), id)
}

// DeleteAll removes all notifications.
func (s *NotificationService) DeleteAll() error {
	return s.repo.DeleteAll(context.Background())
}

// CleanupOlderThan removes notifications older than the given number of days.
func (s *NotificationService) CleanupOlderThan(days int) int64 {
	rowsAffected, err := s.repo.CleanupOlderThan(context.Background(), days)
	if err != nil {
		log.Error().Err(err).Int("days", days).Msg("Failed to cleanup old notifications")
	}
	return rowsAffected
}

// GetSettings returns notification settings for the current tenant.
func (s *NotificationService) GetSettings() (*models.NotificationSettings, error) {
	return s.repo.GetSettings(context.Background())
}

// SaveSettings upserts notification settings.
func (s *NotificationService) SaveSettings(retentionDays int) error {
	return s.repo.SaveSettings(context.Background(), retentionDays)
}
