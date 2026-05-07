// Package master_product provides Master Product management services
package master_product

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Validation errors
var (
	ErrTitleRequired      = errors.New("title is required")
	ErrTitleTooLong       = errors.New("title exceeds 120 characters")
	ErrDescriptionTooLong = errors.New("description exceeds 5000 characters")
	ErrTooManyImages      = errors.New("maximum 8 images allowed")
	ErrTooManySKUs        = errors.New("maximum 50 SKUs allowed")
	ErrSellerSkuRequired  = errors.New("seller_sku is required for SKU")
	ErrProductNotFound    = errors.New("product not found")
	ErrSkuNotFound        = errors.New("SKU not found")
	ErrTenantIDRequired   = errors.New("tenant_id is required")
)

// Service handles Master Product business logic
type Service struct {
	repo         *repositories.MasterProductRepository
	db           *gorm.DB
	imageManager *ImageManager // Optional - for join table support
}

// NewService creates a new Master Product service
func NewService(db *gorm.DB) *Service {
	return &Service{
		repo:         repositories.NewMasterProductRepository(db),
		db:           db,
		imageManager: newDefaultImageManager(db),
	}
}

// NewServiceWithImageManager creates a service with image management support
func NewServiceWithImageManager(db *gorm.DB, imgMgr *ImageManager) *Service {
	return &Service{
		repo:         repositories.NewMasterProductRepository(db),
		db:           db,
		imageManager: imgMgr,
	}
}

// CreateInput represents input for creating a master product
type CreateInput struct {
	Title       string           `json:"title"`
	Description string           `json:"description"`
	Images      []string         `json:"images"`
	Status      string           `json:"status"`
	SKUs        []CreateSkuInput `json:"skus"`
}

// CreateSkuInput represents input for creating a SKU
type CreateSkuInput struct {
	SellerSku   string                 `json:"seller_sku"`
	VariantName string                 `json:"variant_name"`
	VariantData map[string]interface{} `json:"variant_data"`
	Price       float64                `json:"price"`
	Stock       int                    `json:"stock"`
}

// UpdateInput represents input for updating a master product
type UpdateInput struct {
	Title       *string          `json:"title,omitempty"`
	Description *string          `json:"description,omitempty"`
	Images      []string         `json:"images,omitempty"`
	Status      *string          `json:"status,omitempty"`
	SKUs        []UpdateSkuInput `json:"skus,omitempty"`
}

// ListFilter represents filters for listing products
type ListFilter struct {
	Status       string `json:"status,omitempty"`
	Search       string `json:"search,omitempty"`
	Platform     string `json:"platform,omitempty"`
	LinkedOnly   bool   `json:"linked_only,omitempty"`
	UnmappedOnly bool   `json:"unmapped_only,omitempty"` // GAP-16: products with 0 platform links
	Page         int    `json:"page"`
	Limit        int    `json:"limit"`
}

// ListResult represents paginated list result
type ListResult struct {
	Data  []models.MasterProduct `json:"data"`
	Total int64                  `json:"total"`
	Page  int                    `json:"page"`
	Limit int                    `json:"limit"`
}

// ================== CRUD Operations ==================

// Create creates a new master product with validation
func (s *Service) Create(ctx context.Context, tenantID string, input CreateInput) (*models.MasterProduct, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Validate input
	if err := s.validateCreateInput(input); err != nil {
		return nil, err
	}

	// Set default status if not provided
	status := input.Status
	if status == "" {
		status = models.MasterProductStatusDraft
	}

	// Convert images to JSONArray (list of URLs)
	var images models.JSONArray
	if len(input.Images) > 0 {
		images = make(models.JSONArray, len(input.Images))
		for i, img := range input.Images {
			images[i] = img
		}
	}

	// Create product
	product := &models.MasterProduct{
		TenantID:    tenantID,
		Title:       input.Title,
		Description: input.Description,
		Images:      images,
		Status:      status,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := s.repo.Create(ctx, product); err != nil {
		log.Error().Err(err).Str("tenant_id", tenantID).Msg("Failed to create master product")
		return nil, err
	}

	// Create SKUs if provided
	for _, skuInput := range input.SKUs {
		if err := s.createSkuForProduct(ctx, tenantID, product.ID, skuInput); err != nil {
			log.Error().Err(err).Uint("product_id", product.ID).Msg("Failed to create SKU")
			// Continue with other SKUs, don't fail the whole operation
		}
	}

	// Create join table entries (dual-write)
	if s.imageManager != nil && len(input.Images) > 0 {
		if err := s.imageManager.ProcessAndLinkImages(ctx, tenantID, product.ID, input.Images); err != nil {
			log.Warn().Err(err).Uint("product_id", product.ID).Msg("Failed to create image join entries (non-fatal)")
			// Don't fail the whole operation - JSONB is still written
		}
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", product.ID).
		Str("title", product.Title).
		Int("sku_count", len(input.SKUs)).
		Msg("Master product created")

	// Fetch complete product with SKUs
	return s.repo.FindByID(ctx, product.ID)
}

// GetByID retrieves a product by ID
func (s *Service) GetByID(ctx context.Context, tenantID string, id uint) (*models.MasterProduct, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	product, err := s.repo.FindByTenantAndID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	return product, nil
}

// List retrieves products with filtering and pagination
func (s *Service) List(ctx context.Context, tenantID string, filter ListFilter) (*ListResult, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Set defaults
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	var products []models.MasterProduct
	var total int64
	var err error

	if filter.UnmappedOnly {
		// Products with ZERO platform links (not connected to any marketplace)
		products, total, err = s.repo.FindUnmapped(ctx, tenantID, filter.Page, filter.Limit, filter.Search)
	} else if filter.LinkedOnly && filter.Platform == "" {
		// "Mapped" tab — products with at least one platform link
		products, total, err = s.repo.FindMapped(ctx, tenantID, filter.Page, filter.Limit, filter.Search)
	} else if filter.Platform != "" {
		// Platform-specific filter — uses platform links
		products, total, err = s.repo.FindLinked(
			ctx,
			tenantID,
			filter.Page,
			filter.Limit,
			filter.Status,
			filter.Search,
			filter.Platform,
		)
	} else if filter.Search != "" {
		products, total, err = s.repo.SearchByTitle(ctx, tenantID, filter.Search, filter.Page, filter.Limit)
	} else if filter.Status != "" {
		products, total, err = s.repo.FindByStatus(ctx, tenantID, filter.Status, filter.Page, filter.Limit)
	} else {
		products, total, err = s.repo.FindAll(ctx, tenantID, filter.Page, filter.Limit)
	}

	if err != nil {
		return nil, err
	}

	// Opsi A: Enrich with per-platform real prices from staging tables
	s.enrichWithPlatformPrices(ctx, tenantID, products)
	// Enrich with inventory reference price/stock from Google Sheets
	s.enrichWithInventoryPrices(ctx, tenantID, products)

	return &ListResult{
		Data:  products,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}
