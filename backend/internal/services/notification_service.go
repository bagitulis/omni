package services

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
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
	db       *gorm.DB
	tenantID string
}

// NewNotificationService creates a new notification service.
// It requires a DB connection and the tenantID to route SSE events correctly.
func NewNotificationService(db *gorm.DB) *NotificationService {
	// Try to extract tenantID from DB session if possible, or expect it to be set later
	// For simplicity, we'll assume the caller ensures the DB is tenant-scoped.
	return &NotificationService{db: db}
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
	if err := s.db.Create(notif).Error; err != nil {
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
	
	// If the job has a tenant ID in models, we should use it
	// Assuming MultiTenantExecutor sets the tenant-scoped DB before calling this
	
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
	globalSSEManager.mu.RLock()
	clients, ok := globalSSEManager.clients[s.tenantID]
	globalSSEManager.mu.RUnlock()

	if !ok || len(clients) == 0 {
		return
	}

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
			// Client buffer full, skip or handle as needed
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
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := s.db.Order("created_at DESC").Limit(limit)
	if sinceID > 0 {
		query = query.Where("id < ?", sinceID) // Pagination: items older than sinceID
	}
	if unreadOnly {
		query = query.Where("read = false")
	}

	var items []models.Notification
	err := query.Find(&items).Error
	return items, err
}

// UnreadCount returns the number of unread notifications.
func (s *NotificationService) UnreadCount() (int64, error) {
	var count int64
	err := s.db.Model(&models.Notification{}).Where("read = false").Count(&count).Error
	return count, err
}

// MarkAsRead marks a single notification as read.
func (s *NotificationService) MarkAsRead(id int64) error {
	return s.db.Model(&models.Notification{}).
		Where("id = ?", id).
		Update("read", true).Error
}

// MarkAllAsRead marks all notifications as read.
func (s *NotificationService) MarkAllAsRead() error {
	return s.db.Model(&models.Notification{}).
		Where("read = false").
		Update("read", true).Error
}

// Delete removes a single notification.
func (s *NotificationService) Delete(id int64) error {
	return s.db.Delete(&models.Notification{}, id).Error
}

// DeleteAll removes all notifications.
func (s *NotificationService) DeleteAll() error {
	return s.db.Where("1=1").Delete(&models.Notification{}).Error
}

// CleanupOlderThan removes notifications older than the given number of days.
func (s *NotificationService) CleanupOlderThan(days int) int64 {
	if days <= 0 {
		return 0
	}
	cutoff := time.Now().AddDate(0, 0, -days)
	result := s.db.Where("created_at < ?", cutoff).Delete(&models.Notification{})
	if result.Error != nil {
		log.Error().Err(result.Error).Int("days", days).Msg("Failed to cleanup old notifications")
	}
	return result.RowsAffected
}

// GetSettings returns notification settings for the current tenant.
func (s *NotificationService) GetSettings() (*models.NotificationSettings, error) {
	var settings models.NotificationSettings
	err := s.db.First(&settings).Error
	if err == gorm.ErrRecordNotFound {
		return &models.NotificationSettings{RetentionDays: 30}, nil
	}
	return &settings, err
}

// SaveSettings upserts notification settings.
func (s *NotificationService) SaveSettings(retentionDays int) error {
	var existing models.NotificationSettings
	err := s.db.First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		return s.db.Create(&models.NotificationSettings{
			RetentionDays: retentionDays,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}).Error
	}
	if err != nil {
		return err
	}

	existing.RetentionDays = retentionDays
	existing.UpdatedAt = time.Now()
	return s.db.Save(&existing).Error
}
