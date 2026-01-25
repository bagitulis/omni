package google

import (
	"errors"
	"sync"
	"time"
)

// ErrAccountNotFound is returned when a service account is not found
var ErrAccountNotFound = errors.New("account not found")

// QuotaService manages Google API quota tracking
type QuotaService struct {
	mu            sync.RWMutex
	accounts      map[string]*AccountQuota
	requestsToday int64
	lastReset     time.Time
}

// AccountQuota represents quota for a single service account
type AccountQuota struct {
	AccountID     string    `json:"account_id"`
	Email         string    `json:"email"`
	RequestsToday int64     `json:"requests_today"`
	RequestsLimit int64     `json:"requests_limit"`
	LastRequest   time.Time `json:"last_request"`
	IsActive      bool      `json:"is_active"`
}

// QuotaStatus represents overall quota status
type QuotaStatus struct {
	TotalRequests   int64           `json:"total_requests"`
	DailyLimit      int64           `json:"daily_limit"`
	UsagePercent    float64         `json:"usage_percent"`
	ActiveAccount   string          `json:"active_account"`
	AccountsCount   int             `json:"accounts_count"`
	ResetTime       time.Time       `json:"reset_time"`
	Accounts        []*AccountQuota `json:"accounts"`
}

// QuotaDetailedStats represents detailed quota statistics
type QuotaDetailedStats struct {
	QuotaStatus
	RequestsByHour    map[int]int64 `json:"requests_by_hour"`
	AveragePerMinute  float64       `json:"average_per_minute"`
	PeakHour          int           `json:"peak_hour"`
	EstimatedDepleted time.Time     `json:"estimated_depleted"`
}

// NewQuotaService creates a new quota service
func NewQuotaService() *QuotaService {
	return &QuotaService{
		accounts:  make(map[string]*AccountQuota),
		lastReset: time.Now().Truncate(24 * time.Hour),
	}
}

// RegisterAccount registers a service account for quota tracking
func (q *QuotaService) RegisterAccount(accountID, email string, limit int64) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.accounts[accountID] = &AccountQuota{
		AccountID:     accountID,
		Email:         email,
		RequestsLimit: limit,
		IsActive:      len(q.accounts) == 0, // First account is active by default
	}
}

// RecordRequest records an API request for quota tracking
func (q *QuotaService) RecordRequest(accountID string) {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.checkReset()

	if acc, ok := q.accounts[accountID]; ok {
		acc.RequestsToday++
		acc.LastRequest = time.Now()
	}
	q.requestsToday++
}

// GetStatus returns current quota status
func (q *QuotaService) GetStatus() *QuotaStatus {
	q.mu.RLock()
	defer q.mu.RUnlock()

	q.checkReset()

	var totalLimit int64
	var activeAccount string
	accounts := make([]*AccountQuota, 0, len(q.accounts))

	for _, acc := range q.accounts {
		totalLimit += acc.RequestsLimit
		if acc.IsActive {
			activeAccount = acc.AccountID
		}
		accounts = append(accounts, acc)
	}

	usagePercent := float64(0)
	if totalLimit > 0 {
		usagePercent = float64(q.requestsToday) / float64(totalLimit) * 100
	}

	return &QuotaStatus{
		TotalRequests: q.requestsToday,
		DailyLimit:    totalLimit,
		UsagePercent:  usagePercent,
		ActiveAccount: activeAccount,
		AccountsCount: len(q.accounts),
		ResetTime:     q.lastReset.Add(24 * time.Hour),
		Accounts:      accounts,
	}
}

// GetDetailedStats returns detailed quota statistics
func (q *QuotaService) GetDetailedStats() *QuotaDetailedStats {
	status := q.GetStatus()

	// Calculate estimates
	hoursElapsed := time.Since(q.lastReset).Hours()
	avgPerMinute := float64(0)
	if hoursElapsed > 0 {
		avgPerMinute = float64(q.requestsToday) / (hoursElapsed * 60)
	}

	var estimatedDepleted time.Time
	if avgPerMinute > 0 && status.DailyLimit > q.requestsToday {
		remaining := status.DailyLimit - q.requestsToday
		minutesToDepletion := float64(remaining) / avgPerMinute
		estimatedDepleted = time.Now().Add(time.Duration(minutesToDepletion) * time.Minute)
	}

	return &QuotaDetailedStats{
		QuotaStatus:       *status,
		RequestsByHour:    make(map[int]int64), // Would need persistent storage for real tracking
		AveragePerMinute:  avgPerMinute,
		PeakHour:          0,
		EstimatedDepleted: estimatedDepleted,
	}
}

// checkReset resets counters if a new day has started
func (q *QuotaService) checkReset() {
	now := time.Now().Truncate(24 * time.Hour)
	if now.After(q.lastReset) {
		q.requestsToday = 0
		for _, acc := range q.accounts {
			acc.RequestsToday = 0
		}
		q.lastReset = now
	}
}

// SetActiveAccount sets the active service account
func (q *QuotaService) SetActiveAccount(accountID string) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	found := false
	for id, acc := range q.accounts {
		if id == accountID {
			acc.IsActive = true
			found = true
		} else {
			acc.IsActive = false
		}
	}

	if !found {
		return ErrAccountNotFound
	}
	return nil
}
