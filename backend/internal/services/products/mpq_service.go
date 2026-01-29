package products

import (
	"context"
	"time"

	"gorm.io/gorm"
)

// MPQRule represents a minimum purchase quantity rule
type MPQRule struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    string    `gorm:"index;not null" json:"tenant_id"`
	Platform    string    `gorm:"index;not null" json:"platform"`
	ItemID      string    `gorm:"index" json:"item_id,omitempty"`
	SKU         string    `gorm:"index" json:"sku,omitempty"`
	CategoryID  string    `gorm:"index" json:"category_id,omitempty"`
	MinQty      int       `gorm:"not null" json:"min_qty"`
	MaxQty      int       `json:"max_qty,omitempty"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	Priority    int       `gorm:"default:0" json:"priority"` // Higher = more specific
	Description string    `json:"description,omitempty"`
	StartDate   time.Time `json:"start_date,omitempty"`
	EndDate     time.Time `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// TableName returns the table name for GORM
func (MPQRule) TableName() string {
	return "mpq_rules"
}

// MPQValidationResult represents MPQ validation result
type MPQValidationResult struct {
	IsValid      bool     `json:"is_valid"`
	MinQty       int      `json:"min_qty"`
	MaxQty       int      `json:"max_qty"`
	RequestedQty int      `json:"requested_qty"`
	RuleApplied  *MPQRule `json:"rule_applied,omitempty"`
	Message      string   `json:"message"`
}

// MPQService handles minimum purchase quantity operations
type MPQService struct {
	db *gorm.DB
}

// NewMPQService creates a new MPQ service
func NewMPQService(db *gorm.DB) *MPQService {
	return &MPQService{db: db}
}

// GetMPQRules retrieves all MPQ rules for a tenant
func (s *MPQService) GetMPQRules(
	ctx context.Context,
	tenantID string,
	platform string,
) ([]MPQRule, error) {
	var rules []MPQRule

	query := s.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		Where("is_active = ?", true).
		Order("priority DESC, created_at ASC")

	if platform != "" {
		query = query.Where("platform = ?", platform)
	}

	err := query.Find(&rules).Error
	return rules, err
}

// GetApplicableRule finds the applicable MPQ rule for an item
func (s *MPQService) GetApplicableRule(
	ctx context.Context,
	tenantID string,
	platform string,
	itemID string,
	sku string,
	categoryID string,
) (*MPQRule, error) {
	now := time.Now()

	// Try to find rule by SKU first (most specific)
	if sku != "" {
		var rule MPQRule
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND platform = ? AND sku = ? AND is_active = ?",
				tenantID, platform, sku, true).
			Where("(start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
				now, now).
			Order("priority DESC").
			First(&rule).Error
		if err == nil {
			return &rule, nil
		}
	}

	// Try by item ID
	if itemID != "" {
		var rule MPQRule
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND platform = ? AND item_id = ? AND is_active = ?",
				tenantID, platform, itemID, true).
			Where("(start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
				now, now).
			Order("priority DESC").
			First(&rule).Error
		if err == nil {
			return &rule, nil
		}
	}

	// Try by category
	if categoryID != "" {
		var rule MPQRule
		err := s.db.WithContext(ctx).
			Where("tenant_id = ? AND platform = ? AND category_id = ? AND is_active = ?",
				tenantID, platform, categoryID, true).
			Where("(start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
				now, now).
			Order("priority DESC").
			First(&rule).Error
		if err == nil {
			return &rule, nil
		}
	}

	// Try platform-wide rule
	var rule MPQRule
	err := s.db.WithContext(ctx).
		Where("tenant_id = ? AND platform = ? AND is_active = ?", tenantID, platform, true).
		Where("item_id IS NULL OR item_id = ''").
		Where("sku IS NULL OR sku = ''").
		Where("category_id IS NULL OR category_id = ''").
		Where("(start_date IS NULL OR start_date <= ?) AND (end_date IS NULL OR end_date >= ?)",
			now, now).
		Order("priority DESC").
		First(&rule).Error

	if err != nil {
		return nil, err
	}

	return &rule, nil
}

// ValidatePurchaseQuantity validates a purchase quantity against MPQ rules
func (s *MPQService) ValidatePurchaseQuantity(
	ctx context.Context,
	tenantID string,
	platform string,
	itemID string,
	sku string,
	categoryID string,
	requestedQty int,
) *MPQValidationResult {
	result := &MPQValidationResult{
		RequestedQty: requestedQty,
		IsValid:      true,
		MinQty:       1,
	}

	rule, err := s.GetApplicableRule(ctx, tenantID, platform, itemID, sku, categoryID)
	if err != nil {
		// No rule found, allow any quantity
		result.Message = "No MPQ rule applicable"
		return result
	}

	result.RuleApplied = rule
	result.MinQty = rule.MinQty
	result.MaxQty = rule.MaxQty

	if requestedQty < rule.MinQty {
		result.IsValid = false
		result.Message = "Quantity below minimum"
		return result
	}

	if rule.MaxQty > 0 && requestedQty > rule.MaxQty {
		result.IsValid = false
		result.Message = "Quantity exceeds maximum"
		return result
	}

	result.Message = "Quantity is valid"
	return result
}

// SaveMPQRule saves or updates an MPQ rule
func (s *MPQService) SaveMPQRule(
	ctx context.Context,
	rule *MPQRule,
) error {
	if rule.ID == 0 {
		return s.db.WithContext(ctx).Create(rule).Error
	}
	return s.db.WithContext(ctx).Save(rule).Error
}

// DeleteMPQRule deletes an MPQ rule
func (s *MPQService) DeleteMPQRule(
	ctx context.Context,
	tenantID string,
	id uint,
) error {
	return s.db.WithContext(ctx).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Delete(&MPQRule{}).Error
}

// DisableMPQRule disables an MPQ rule
func (s *MPQService) DisableMPQRule(
	ctx context.Context,
	tenantID string,
	id uint,
) error {
	return s.db.WithContext(ctx).
		Model(&MPQRule{}).
		Where("tenant_id = ? AND id = ?", tenantID, id).
		Update("is_active", false).Error
}
