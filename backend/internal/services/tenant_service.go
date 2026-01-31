package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

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

// GetAvailableTenants returns list of tenants available for login
// Fetches from PostgreSQL schemas (tenant_*)
func (s *TenantService) GetAvailableTenants(ctx context.Context) ([]TenantInfo, error) {
	// Get schemas from PostgreSQL
	systemDB, err := config.GetSystemDB(s.basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get system db: %w", err)
	}

	var schemas []string
	err = systemDB.Raw(`
		SELECT schema_name 
		FROM information_schema.schemata 
		WHERE schema_name LIKE 'tenant_%'
		ORDER BY schema_name
	`).Scan(&schemas).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant schemas: %w", err)
	}

	tenants := make([]TenantInfo, 0, len(schemas))
	for _, schema := range schemas {
		// Extract tenant ID from schema name (tenant_xxx -> xxx)
		tenantID := strings.TrimPrefix(schema, "tenant_")
		shopName := formatShopName(tenantID)
		tenants = append(tenants, TenantInfo{
			ID:       tenantID,
			ShopName: shopName,
		})
	}

	return tenants, nil
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
