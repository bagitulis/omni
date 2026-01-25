package tiktok

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/dto/response"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// ProductSearchHandler handles TikTok product search/sync operations
type ProductSearchHandler struct {
	basePath string
}

// NewProductSearchHandler creates a new product search handler
func NewProductSearchHandler(basePath string) *ProductSearchHandler {
	return &ProductSearchHandler{basePath: basePath}
}

// SearchProductsRequest represents the request body for product search
type SearchProductsRequest struct {
	Status   string `json:"status"`
	PageSize int    `json:"page_size"`
	Limit    int    `json:"limit,omitempty"`
	SyncToDB bool   `json:"sync_to_db"`
}

// SearchProducts handles POST /api/tiktok/products/search
// Fetches products from TikTok API and syncs to database
func (h *ProductSearchHandler) SearchProducts(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, response.Error("Missing tenantId"))
		return
	}

	// Parse request body
	var req SearchProductsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// Default values if no body provided
		req.Status = "ACTIVATE"
		req.PageSize = 100
		req.SyncToDB = true
	}

	// Set defaults
	if req.PageSize <= 0 {
		req.PageSize = 100
	}
	if req.Status == "" {
		req.Status = "ACTIVATE"
	}

	// Get TikTok client
	client, err := h.getTiktokClient(tenantID)
	if err != nil {
		log.Printf("Failed to get TikTok client: %v", err)
		c.JSON(http.StatusInternalServerError, response.Error("Failed to get TikTok client: "+err.Error()))
		return
	}

	// Get tenant database
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error("Database connection failed"))
		return
	}

	ctx := c.Request.Context()
	productRepo := repositories.NewTiktokProductRepository(db)
	skuRepo := repositories.NewTiktokSkuRepository(db)

	// Fetch all products with pagination
	var allProducts []tiktokPkg.ProductSearchItem
	pageToken := ""
	pageCount := 0

	for {
		resp, err := client.SearchProductsV202502(req.Status, req.PageSize, pageToken)
		if err != nil {
			log.Printf("TikTok API error: %v", err)
			c.JSON(http.StatusInternalServerError, response.Error("TikTok API error: "+err.Error()))
			return
		}

		if resp.Code != 0 {
			log.Printf("TikTok API returned error code %d: %s", resp.Code, resp.Message)
			c.JSON(http.StatusInternalServerError, response.Error("TikTok API error: "+resp.Message))
			return
		}

		allProducts = append(allProducts, resp.Data.Products...)
		pageCount++

		// Check if more pages
		if resp.Data.NextPageToken == "" {
			break
		}
		pageToken = resp.Data.NextPageToken

		// Safety limit
		if pageCount >= 50 {
			log.Printf("Reached page limit of 50")
			break
		}
	}

	log.Printf("Fetched %d products from TikTok API in %d pages", len(allProducts), pageCount)

	// Sync to database if requested
	if req.SyncToDB && len(allProducts) > 0 {
		savedProducts := 0
		savedSkus := 0

		for _, prod := range allProducts {
			// Fetch product detail for variant info (sales_attributes)
			detailResp, err := client.GetProductDetail(prod.ID)
			if err != nil {
				log.Printf("Failed to get detail for product %s: %v", prod.ID, err)
				// Continue without detail
				detailResp = nil
			}

			// Save product
			dbProduct := &models.TiktokProduct{
				TenantID:    tenantID,
				ProductID:   prod.ID,
				Name:        prod.Title,
				Description: prod.Description,
				Status:      prod.Status,
			}

			// Calculate total quantity from SKUs
			totalQty := 0
			for _, sku := range prod.Skus {
				for _, inv := range sku.Inventory {
					totalQty += inv.Quantity
				}
			}
			dbProduct.Quantity = totalQty

			// Get price from first SKU
			if len(prod.Skus) > 0 {
				price, _ := strconv.ParseFloat(prod.Skus[0].Price.SalePrice, 64)
				dbProduct.Price = price
			}

			if err := productRepo.Upsert(ctx, dbProduct); err != nil {
				log.Printf("Failed to save product %s: %v", prod.ID, err)
				continue
			}
			savedProducts++

			// Save SKUs with variant_name from product detail
			skuMap := make(map[string][]tiktokPkg.ProductSalesAttr)
			if detailResp != nil && detailResp.Code == 0 {
				for _, detailSku := range detailResp.Data.Skus {
					skuMap[detailSku.ID] = detailSku.SalesAttributes
				}
			}

			for _, sku := range prod.Skus {
				dbSku := &models.TiktokSku{
					TenantID:  tenantID,
					ProductID: dbProduct.ID,
					SkuID:     sku.ID,
					SellerSku: sku.SellerSku,
				}

				// Set price
				price, _ := strconv.ParseFloat(sku.Price.SalePrice, 64)
				dbSku.Price = price

				// Set quantity
				for _, inv := range sku.Inventory {
					dbSku.Quantity += inv.Quantity
				}

				// Set variant_name from sales_attributes
				if attrs, ok := skuMap[sku.ID]; ok && len(attrs) > 0 {
					dbSku.VariantName = buildVariantName(attrs)
					dbSku.VariantData = buildVariantData(attrs)
				}

				if err := skuRepo.Upsert(ctx, dbSku); err != nil {
					log.Printf("Failed to save SKU %s: %v", sku.ID, err)
					continue
				}
				savedSkus++
			}
		}

		log.Printf("Saved %d products and %d SKUs to database", savedProducts, savedSkus)

		c.JSON(http.StatusOK, gin.H{
			"success":       true,
			"pageCount":     pageCount,
			"totalProducts": len(allProducts),
			"savedProducts": savedProducts,
			"savedSkus":     savedSkus,
		})
		return
	}

	// Return without sync
	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"pageCount":     pageCount,
		"totalProducts": len(allProducts),
		"products":      allProducts,
	})
}

// buildVariantName creates variant name from sales attributes
// Format: "Color:Red | Size:XL"
func buildVariantName(attrs []tiktokPkg.ProductSalesAttr) string {
	if len(attrs) == 0 {
		return ""
	}

	parts := make([]string, 0, len(attrs))
	for _, attr := range attrs {
		if attr.Name != "" && attr.ValueName != "" {
			parts = append(parts, attr.Name+":"+attr.ValueName)
		}
	}
	return strings.Join(parts, " | ")
}

// buildVariantData creates JSONMap variant data for JSONB column
func buildVariantData(attrs []tiktokPkg.ProductSalesAttr) models.JSONMap {
	if len(attrs) == 0 {
		return nil
	}

	result := make(models.JSONMap)
	for _, attr := range attrs {
		if attr.Name != "" && attr.ValueName != "" {
			result[attr.Name] = attr.ValueName
		}
	}
	return result
}

// getTiktokClient creates TikTok API client for tenant
func (h *ProductSearchHandler) getTiktokClient(tenantID string) (*tiktokPkg.Client, error) {
	ctx := context.Background()
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return nil, err
	}

	// Get platform credentials from tenant's key-value storage
	// NOTE: PostgreSQL uses schema isolation, NOT tenant_id column
	credRepo := repositories.NewPlatformCredentialsRepository(db)
	tenantCreds, err := credRepo.GetTiktokCredentials(ctx)
	if err != nil {
		return nil, err
	}
	if tenantCreds.AccessToken == "" || tenantCreds.ShopCipher == "" {
		return nil, fmt.Errorf("missing TikTok credentials: accessToken or shopCipher not configured")
	}

	// Use tenant credentials for appKey/appSecret if available, otherwise fall back to global
	appKey := tenantCreds.AppKey
	appSecret := tenantCreds.AppSecret

	if appKey == "" || appSecret == "" {
		// Fall back to global credentials
		systemDB, err := config.GetSystemDB(h.basePath)
		if err != nil {
			return nil, err
		}
		globalRepo := repositories.NewGlobalConfigRepository(systemDB)
		globalCreds, err := globalRepo.GetTiktokCredentials(ctx)
		if err != nil {
			return nil, err
		}
		appKey = globalCreds.AppKey
		appSecret = globalCreds.AppSecret
	}

	client := tiktokPkg.NewClient(appKey, appSecret)
	client.SetCredentials(tenantCreds.AccessToken, tenantCreds.ShopCipher)
	return client, nil
}
