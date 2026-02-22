// Package master_product provides Master Product management services
package master_product

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
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
		// GAP-16: Find products whose SKUs are NOT in inventory_records
		products, total, err = s.repo.FindUnmapped(ctx, tenantID, filter.Page, filter.Limit, filter.Search)
	} else if filter.LinkedOnly && filter.Platform == "" {
		// GAP-16: "Mapped" tab — find products whose SKUs ARE in inventory_records
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

// enrichWithInventoryPrices backfills price/stock from inventory_records (Google Sheets)
// and sets InventoryPrice/InventoryStock virtual fields for reference display.
func (s *Service) enrichWithInventoryPrices(ctx context.Context, tenantID string, products []models.MasterProduct) {
	allSKUs := collectSKUs(products)
	if len(allSKUs) == 0 {
		return
	}

	// Batch query inventory_records
	inventoryTable := models.GetTableName("InventoryRecord")
	var inventoryRows []struct {
		KeyValue string `gorm:"column:key_value"`
		Data     string `gorm:"column:data"`
	}
	if err := s.db.WithContext(ctx).
		Table(inventoryTable).
		Select("key_value, data").
		Where("tenant_id = ? AND LOWER(key_value) IN (?)", tenantID, allSKUs).
		Find(&inventoryRows).Error; err != nil {
		log.Warn().Err(err).Msg("Failed to query inventory for price enrichment")
		return
	}

	// Build lookup map: lowercase SKU → inventory data
	invMap := make(map[string]map[string]interface{}, len(inventoryRows))
	for _, row := range inventoryRows {
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(row.Data), &data); err != nil {
			continue
		}
		invMap[strings.ToLower(row.KeyValue)] = data
	}

	for i := range products {
		for j := range products[i].SKUs {
			sku := &products[i].SKUs[j]
			invData, ok := invMap[strings.ToLower(sku.SellerSku)]
			if !ok {
				continue
			}

			// Set inventory reference price (virtual field)
			if hargaStr, ok := invData["HARGA"].(string); ok {
				if harga, err := strconv.ParseFloat(strings.TrimSpace(hargaStr), 64); err == nil && harga > 0 {
					sku.InventoryPrice = harga
					// Backfill master price if 0
					if sku.Price == 0 {
						sku.Price = harga
					}
				}
			}

			// Set inventory reference stock (virtual field)
			if stokStr, ok := invData["Sisa Stok"].(string); ok {
				if stok, err := strconv.Atoi(strings.TrimSpace(stokStr)); err == nil && stok > 0 {
					sku.InventoryStock = stok
					// Backfill master stock if 0
					if sku.Stock == 0 {
						sku.Stock = stok
					}
				}
			}
		}
	}
}

// enrichWithPlatformPrices populates PlatformPrices virtual field from staging tables.
// Queries shopee_skus, lazada_skus, tiktok_skus for real marketplace prices.
func (s *Service) enrichWithPlatformPrices(ctx context.Context, tenantID string, products []models.MasterProduct) {
	allSKUs := collectSKUs(products)
	if len(allSKUs) == 0 {
		return
	}

	type stagingRow struct {
		SellerSku string  `gorm:"column:seller_sku"`
		Price     float64 `gorm:"column:price"`
		Quantity  int     `gorm:"column:quantity"`
	}

	// platformMap: lowercase_sku → platform → {price, stock}
	platformMap := make(map[string][]models.PlatformPrice)

	// Query each platform staging table
	platforms := []struct {
		name  string
		table string
	}{
		{"shopee", models.GetTableName("ShopeeSku")},
		{"lazada", models.GetTableName("LazadaSku")},
		{"tiktok", models.GetTableName("TiktokSku")},
	}

	for _, p := range platforms {
		var rows []stagingRow
		if err := s.db.WithContext(ctx).
			Table(p.table).
			Select("seller_sku, price, quantity").
			Where("tenant_id = ? AND LOWER(seller_sku) IN (?)", tenantID, allSKUs).
			Find(&rows).Error; err != nil {
			log.Warn().Err(err).Str("platform", p.name).Msg("Failed to query platform staging prices")
			continue
		}

		for _, row := range rows {
			key := strings.ToLower(row.SellerSku)
			platformMap[key] = append(platformMap[key], models.PlatformPrice{
				Platform: p.name,
				Price:    row.Price,
				Stock:    row.Quantity,
			})
		}
	}

	// Attach platform prices to each SKU
	for i := range products {
		for j := range products[i].SKUs {
			sku := &products[i].SKUs[j]
			if prices, ok := platformMap[strings.ToLower(sku.SellerSku)]; ok {
				sku.PlatformPrices = prices
			}
		}
	}
}

// collectSKUs gathers all lowercase seller_skus from products.
func collectSKUs(products []models.MasterProduct) []string {
	var skus []string
	for _, p := range products {
		for _, sku := range p.SKUs {
			if sku.SellerSku != "" {
				skus = append(skus, strings.ToLower(sku.SellerSku))
			}
		}
	}
	return skus
}

// Update updates a master product
func (s *Service) Update(ctx context.Context, tenantID string, id uint, input UpdateInput) (*models.MasterProduct, error) {
	if tenantID == "" {
		return nil, ErrTenantIDRequired
	}

	// Fetch existing product
	product, err := s.repo.FindByTenantAndID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, err
	}

	// Update fields if provided
	if input.Title != nil {
		if len(*input.Title) == 0 {
			return nil, ErrTitleRequired
		}
		if len(*input.Title) > models.MasterProductMaxTitleLength {
			return nil, ErrTitleTooLong
		}
		product.Title = *input.Title
	}

	if input.Description != nil {
		if len(*input.Description) > models.MasterProductMaxDescriptionLength {
			return nil, ErrDescriptionTooLong
		}
		product.Description = *input.Description
	}

	if input.Images != nil {
		if len(input.Images) > models.MasterProductMaxImages {
			return nil, ErrTooManyImages
		}
		images := make(models.JSONArray, len(input.Images))
		for i, img := range input.Images {
			images[i] = img
		}
		product.Images = images
	}

	if input.Status != nil {
		product.Status = *input.Status
	}

	if input.SKUs != nil {
		if err := s.replaceProductSKUs(ctx, tenantID, id, input.SKUs); err != nil {
			return nil, err
		}
	}

	product.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, product); err != nil {
		log.Error().Err(err).Uint("product_id", id).Msg("Failed to update master product")
		return nil, err
	}

	// Update join table entries (dual-write)
	if s.imageManager != nil && input.Images != nil {
		if err := s.imageManager.ProcessAndLinkImages(ctx, tenantID, id, input.Images); err != nil {
			log.Warn().Err(err).Uint("product_id", id).Msg("Failed to update image join entries (non-fatal)")
		}
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", id).
		Msg("Master product updated")

	return s.repo.FindByID(ctx, product.ID)
}

// Delete deletes a master product and its SKUs
func (s *Service) Delete(ctx context.Context, tenantID string, id uint) error {
	if tenantID == "" {
		return ErrTenantIDRequired
	}

	// Verify product exists and belongs to tenant
	_, err := s.repo.FindByTenantAndID(ctx, tenantID, id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrProductNotFound
		}
		return err
	}

	// Delete image join entries and update ref counts
	if s.imageManager != nil {
		if err := s.imageManager.DeleteJoinTableEntries(ctx, id); err != nil {
			log.Warn().Err(err).Uint("product_id", id).Msg("Failed to delete image join entries (non-fatal)")
		}
	}

	// Delete associated SKUs first
	skus, _ := s.repo.FindSkusByProductID(ctx, id)
	for _, sku := range skus {
		// Delete platform links for this SKU
		links, _ := s.repo.FindPlatformLinks(ctx, id)
		for _, link := range links {
			if link.MasterSkuID != nil && *link.MasterSkuID == sku.ID {
				_ = s.repo.DeletePlatformLink(ctx, link.ID)
			}
		}
		_ = s.repo.DeleteSku(ctx, sku.ID)
	}

	// Delete product-level platform links
	links, _ := s.repo.FindPlatformLinks(ctx, id)
	for _, link := range links {
		_ = s.repo.DeletePlatformLink(ctx, link.ID)
	}

	// Delete product
	if err := s.repo.Delete(ctx, id); err != nil {
		log.Error().Err(err).Uint("product_id", id).Msg("Failed to delete master product")
		return err
	}

	log.Info().
		Str("tenant_id", tenantID).
		Uint("product_id", id).
		Msg("Master product deleted")

	return nil
}
