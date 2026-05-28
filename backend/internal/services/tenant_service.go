package services

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// ErrTenantNotFound is returned when a tenant does not exist in system.tenants.
var ErrTenantNotFound = fmt.Errorf("tenant not found")

// TenantInfo represents tenant information
type TenantInfo struct {
	ID       string `json:"id"`
	ShopName string `json:"shop_name"`
	IsGlobal bool   `json:"is_global,omitempty"`
}

// TenantService manages multi-tenant operations
type TenantService struct {
	basePath string
}

// NewTenantService creates a new tenant service
func NewTenantService(basePath string) *TenantService {
	return &TenantService{basePath: basePath}
}

// GetAvailableTenants returns list of active tenants available for login.
// Queries system.tenants table and only returns tenants where is_active = true.
func (s *TenantService) GetAvailableTenants(ctx context.Context) ([]TenantInfo, error) {
	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system db: %w", err)
	}

	var tenants []struct {
		TenantID string `gorm:"column:tenant_id"`
		IsActive bool   `gorm:"column:is_active"`
	}

	err = systemDB.WithContext(ctx).Table("system.tenants").
		Select("tenant_id, is_active").
		Where("is_active = ?", true).
		Find(&tenants).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get active tenants: %w", err)
	}

	result := make([]TenantInfo, 0, len(tenants))
	for _, t := range tenants {
		shopName := formatShopName(t.TenantID)
		result = append(result, TenantInfo{
			ID:       t.TenantID,
			ShopName: shopName,
		})
	}

	return result, nil
}

// GetTenantDB returns database connection for a specific tenant
func (s *TenantService) GetTenantDB(tenantID string) (*gorm.DB, error) {
	return config.GetTenantDBWithContext(tenantID, s.basePath)
}

// GetSystemDB returns database connection for system schema (developer accounts, global config)
func (s *TenantService) GetSystemDB() (*gorm.DB, error) {
	return config.GetSystemDB(s.basePath)
}

// GetUserRepoForTenant creates a user repository for specific tenant
func (s *TenantService) GetUserRepoForTenant(tenantID string) (*repositories.UserRepository, error) {
	db, err := s.GetTenantDB(tenantID)
	if err != nil {
		return nil, err
	}
	return repositories.NewUserRepository(db), nil
}

// TenantExists checks if tenant exists
func (s *TenantService) TenantExists(ctx context.Context, tenantID string) bool {
	tenants, err := s.GetAvailableTenants(ctx)
	if err != nil {
		return false
	}
	for _, t := range tenants {
		if t.ID == tenantID {
			return true
		}
	}
	return false
}

// GetActiveTenantsByID returns active tenant metadata for an explicit tenant scope.
func (s *TenantService) GetActiveTenantsByID(ctx context.Context, tenantIDs []string) (map[string]TenantInfo, error) {
	if len(tenantIDs) == 0 {
		return map[string]TenantInfo{}, nil
	}
	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system db: %w", err)
	}
	var tenants []struct {
		TenantID string `gorm:"column:tenant_id"`
		IsActive bool   `gorm:"column:is_active"`
	}
	if err := systemDB.WithContext(ctx).Table("system.tenants").
		Select("tenant_id, is_active").
		Where("tenant_id IN ? AND is_active = ?", tenantIDs, true).
		Find(&tenants).Error; err != nil {
		return nil, fmt.Errorf("failed to get scoped tenants: %w", err)
	}
	result := make(map[string]TenantInfo, len(tenants))
	for _, tenant := range tenants {
		result[tenant.TenantID] = TenantInfo{ID: tenant.TenantID, ShopName: formatShopName(tenant.TenantID)}
	}
	return result, nil
}

// DeactivateTenant soft-deletes a tenant by setting is_active=false.
// Does NOT drop schema or delete any data.
func (s *TenantService) DeactivateTenant(ctx context.Context, tenantID string) error {
	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return fmt.Errorf("failed to get system db: %w", err)
	}

	// Verify tenant exists in system.tenants
	var count int64
	err = systemDB.WithContext(ctx).Table("system.tenants").
		Where("tenant_id = ?", tenantID).
		Count(&count).Error
	if err != nil {
		return fmt.Errorf("failed to check tenant existence: %w", err)
	}
	if count == 0 {
		return ErrTenantNotFound
	}

	// Set is_active = false and deactivated_at = now
	err = systemDB.WithContext(ctx).Table("system.tenants").
		Where("tenant_id = ?", tenantID).
		Updates(map[string]interface{}{
			"is_active":      false,
			"deactivated_at": time.Now(),
		}).Error
	if err != nil {
		return fmt.Errorf("failed to deactivate tenant %s: %w", tenantID, err)
	}

	log.Info().
		Str("service", "tenant").
		Str("tenant_id", tenantID).
		Msg("Tenant deactivated (soft-delete)")

	return nil
}

// MigrateTenant runs migrations for a specific tenant
func (s *TenantService) MigrateTenant(tenantID string) error {
	db, err := config.GetTenantDB(tenantID, s.basePath)
	if err != nil {
		return err
	}
	return config.MigrateTenantDatabase(db, tenantID)
}

// MigrateAllTenants runs migrations for all tenants
func (s *TenantService) MigrateAllTenants(ctx context.Context) error {
	tenants, err := s.GetAvailableTenants(ctx)
	if err != nil {
		return err
	}

	for _, t := range tenants {
		log.Info().
			Str("service", "tenant").
			Str("tenant_id", t.ID).
			Msg("Migrating tenant")
		if err := s.MigrateTenant(t.ID); err != nil {
			log.Error().
				Str("service", "tenant").
				Str("tenant_id", t.ID).
				Err(err).
				Msg("Failed to migrate tenant")
			// Continue with other tenants
		}
	}

	return nil
}

// formatShopName converts tenant_id to readable shop name
func formatShopName(tenantID string) string {
	// yumna_bertigamart -> Bertigamart
	// tika_nusseyba -> Nusseyba Shop
	parts := strings.Split(tenantID, "_")
	if len(parts) >= 2 {
		name := parts[len(parts)-1]
		// Capitalize first letter
		if len(name) > 0 {
			return strings.ToUpper(name[:1]) + name[1:]
		}
	}
	return tenantID
}

// tenantNamePattern validates tenant name: lowercase letter start, alphanumeric + underscore, 3-50 chars.
var tenantNamePattern = regexp.MustCompile(`^[a-z][a-z0-9_]{2,49}$`)

// CreateTenantResult holds the result of tenant creation.
type CreateTenantResult struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
}

// CreateTenant creates a new tenant: validates name, creates schema, runs migrations, registers in system.tenants.
func (s *TenantService) CreateTenant(ctx context.Context, name string) (*CreateTenantResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Validate name format
	if !tenantNamePattern.MatchString(name) {
		return nil, &TenantValidationError{
			Msg: "tenant name must start with lowercase letter, contain only lowercase alphanumeric/underscore, 3-50 chars",
		}
	}

	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system db: %w", err)
	}

	// Check if schema already exists
	schemaName := fmt.Sprintf("tenant_%s", name)
	var schemaCount int64
	err = systemDB.WithContext(ctx).Raw(
		"SELECT COUNT(*) FROM information_schema.schemata WHERE schema_name = ?",
		schemaName,
	).Scan(&schemaCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to check schema existence: %w", err)
	}
	if schemaCount > 0 {
		return nil, &TenantDuplicateError{Name: name}
	}

	// Check system.tenants table for duplicate
	var tenantCount int64
	err = systemDB.WithContext(ctx).Table("system.tenants").
		Where("tenant_id = ?", name).
		Count(&tenantCount).Error
	if err != nil {
		log.Warn().Err(err).Msg("Could not check system.tenants table")
	} else if tenantCount > 0 {
		return nil, &TenantDuplicateError{Name: name}
	}

	// Create tenant schema and run migrations
	log.Info().
		Str("service", "tenant").
		Str("tenant_name", name).
		Msg("Creating new tenant")

	tenantDB, err := config.GetTenantDB(name, s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create tenant connection: %w", err)
	}

	if err := config.MigrateTenantDatabase(tenantDB, name); err != nil {
		s.rollbackTenantSchema(systemDB, name)
		return nil, fmt.Errorf("failed to migrate tenant database: %w", err)
	}

	// Register in system.tenants table
	now := time.Now().UTC()
	err = systemDB.WithContext(ctx).Exec(
		`INSERT INTO system.tenants (tenant_id, is_active, created_at) VALUES (?, ?, ?)`,
		name, true, now,
	).Error
	if err != nil {
		s.rollbackTenantSchema(systemDB, name)
		return nil, fmt.Errorf("failed to register tenant in system: %w", err)
	}

	log.Info().
		Str("service", "tenant").
		Str("action", "create_tenant").
		Str("tenant_name", name).
		Msg("Tenant created successfully")

	return &CreateTenantResult{
		ID:        name,
		Name:      name,
		IsActive:  true,
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}

// rollbackTenantSchema drops a tenant schema on creation failure.
func (s *TenantService) rollbackTenantSchema(systemDB *gorm.DB, name string) {
	schemaName := fmt.Sprintf("tenant_%s", name)
	err := systemDB.Exec("DROP SCHEMA IF EXISTS " + quoteID(schemaName) + " CASCADE").Error
	if err != nil {
		log.Error().
			Str("service", "tenant").
			Str("schema", schemaName).
			Err(err).
			Msg("Failed to rollback tenant schema")
	} else {
		log.Info().
			Str("service", "tenant").
			Str("schema", schemaName).
			Msg("Rolled back tenant schema")
	}
	config.ResetTenantConnection(name)
}

// quoteID safely quotes a SQL identifier to prevent injection.
func quoteID(identifier string) string {
	return `"` + strings.ReplaceAll(identifier, `"`, `""`) + `"`
}

// TenantValidationError represents a tenant name validation failure.
type TenantValidationError struct {
	Msg string
}

func (e *TenantValidationError) Error() string {
	return e.Msg
}

// TenantDuplicateError represents a duplicate tenant name.
type TenantDuplicateError struct {
	Name string
}

func (e *TenantDuplicateError) Error() string {
	return fmt.Sprintf("tenant '%s' already exists", e.Name)
}
