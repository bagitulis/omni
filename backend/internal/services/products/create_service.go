package products

import (
	"context"
	"fmt"

	"gorm.io/gorm"
)

// PlatformAPI defines interface for platform product operations
type PlatformAPI interface {
	CreateProduct(ctx context.Context, data map[string]interface{}) (map[string]interface{}, error)
	UploadImage(ctx context.Context, imageURL string) (string, error)
	GetCategories(ctx context.Context, parentID int64) ([]Category, error)
}

// Category represents a product category
type Category struct {
	ID       int64      `json:"category_id"`
	Name     string     `json:"name"`
	ParentID int64      `json:"parent_id"`
	HasChild bool       `json:"has_child"`
	Children []Category `json:"children,omitempty"`
}

// CreateProductRequest represents product creation request
type CreateProductRequest struct {
	Name        string            `json:"name" binding:"required"`
	Description string            `json:"description"`
	Price       float64           `json:"price" binding:"required"`
	Stock       int               `json:"stock"`
	SKU         string            `json:"sku"`
	CategoryID  int64             `json:"category_id" binding:"required"`
	Images      []string          `json:"images"`
	Variations  []VariationCreate `json:"variations,omitempty"`
	Weight      float64           `json:"weight"`
	Attributes  map[string]string `json:"attributes,omitempty"`
}

// VariationCreate represents variation data for creation
type VariationCreate struct {
	Name  string  `json:"name"`
	SKU   string  `json:"sku"`
	Price float64 `json:"price"`
	Stock int     `json:"stock"`
}

// CreateProductResult represents the result of product creation
type CreateProductResult struct {
	Success   bool   `json:"success"`
	ProductID int64  `json:"product_id,omitempty"`
	ItemID    int64  `json:"item_id,omitempty"`
	Message   string `json:"message,omitempty"`
	Error     string `json:"error,omitempty"`
}

// CreateService handles product creation across platforms
type CreateService struct {
	db       *gorm.DB
	tenantID string
}

// NewCreateService creates a new product creation service
func NewCreateService(db *gorm.DB, tenantID string) *CreateService {
	return &CreateService{db: db, tenantID: tenantID}
}

// CreateOnShopee creates a product on Shopee
func (s *CreateService) CreateOnShopee(ctx context.Context, api PlatformAPI, req CreateProductRequest) (*CreateProductResult, error) {
	// Upload images first
	imageIDs := make([]string, 0, len(req.Images))
	for _, imgURL := range req.Images {
		imgID, err := api.UploadImage(ctx, imgURL)
		if err != nil {
			return &CreateProductResult{Success: false, Error: err.Error()}, nil
		}
		imageIDs = append(imageIDs, imgID)
	}

	// Build create payload
	payload := map[string]interface{}{
		"item_name":      req.Name,
		"description":    req.Description,
		"original_price": req.Price,
		"normal_stock":   req.Stock,
		"item_sku":       req.SKU,
		"category_id":    req.CategoryID,
		"image":          map[string]interface{}{"image_id_list": imageIDs},
		"weight":         req.Weight,
	}

	// Add variations if present
	if len(req.Variations) > 0 {
		payload["tier_variation"] = s.buildShopeeVariations(req.Variations)
	}

	resp, err := api.CreateProduct(ctx, payload)
	if err != nil {
		return &CreateProductResult{Success: false, Error: err.Error()}, nil
	}

	return s.parseShopeeResponse(resp), nil
}

// CreateOnLazada creates a product on Lazada
func (s *CreateService) CreateOnLazada(ctx context.Context, api PlatformAPI, req CreateProductRequest) (*CreateProductResult, error) {
	// Upload images
	imageURLs := make([]string, 0, len(req.Images))
	for _, imgURL := range req.Images {
		uploaded, err := api.UploadImage(ctx, imgURL)
		if err != nil {
			return &CreateProductResult{Success: false, Error: err.Error()}, nil
		}
		imageURLs = append(imageURLs, uploaded)
	}

	// Build Lazada XML-style payload
	payload := map[string]interface{}{
		"Request": map[string]interface{}{
			"Product": map[string]interface{}{
				"PrimaryCategory": req.CategoryID,
				"Attributes": map[string]interface{}{
					"name":              req.Name,
					"description":       req.Description,
					"short_description": truncateString(req.Description, 200),
				},
				"Skus": s.buildLazadaSKUs(req, imageURLs),
			},
		},
	}

	resp, err := api.CreateProduct(ctx, payload)
	if err != nil {
		return &CreateProductResult{Success: false, Error: err.Error()}, nil
	}

	return s.parseLazadaResponse(resp), nil
}

// CreateOnTiktok creates a product on TikTok Shop
func (s *CreateService) CreateOnTiktok(ctx context.Context, api PlatformAPI, req CreateProductRequest) (*CreateProductResult, error) {
	// Upload images
	imageIDs := make([]string, 0, len(req.Images))
	for _, imgURL := range req.Images {
		imgID, err := api.UploadImage(ctx, imgURL)
		if err != nil {
			return &CreateProductResult{Success: false, Error: err.Error()}, nil
		}
		imageIDs = append(imageIDs, imgID)
	}

	// Build TikTok payload
	payload := map[string]interface{}{
		"product_name": req.Name,
		"description":  req.Description,
		"category_id":  fmt.Sprintf("%d", req.CategoryID),
		"images":       s.buildTiktokImages(imageIDs),
		"skus":         s.buildTiktokSKUs(req),
	}

	resp, err := api.CreateProduct(ctx, payload)
	if err != nil {
		return &CreateProductResult{Success: false, Error: err.Error()}, nil
	}

	return s.parseTiktokResponse(resp), nil
}

// buildShopeeVariations builds Shopee variation structure
func (s *CreateService) buildShopeeVariations(variations []VariationCreate) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(variations))
	for _, v := range variations {
		result = append(result, map[string]interface{}{
			"name":  v.Name,
			"sku":   v.SKU,
			"price": v.Price,
			"stock": v.Stock,
		})
	}
	return result
}

// buildLazadaSKUs builds Lazada SKU structure
func (s *CreateService) buildLazadaSKUs(req CreateProductRequest, imageURLs []string) []map[string]interface{} {
	if len(req.Variations) == 0 {
		return []map[string]interface{}{{
			"SellerSku": req.SKU,
			"price":     req.Price,
			"quantity":  req.Stock,
			"Images":    map[string]interface{}{"Image": imageURLs},
		}}
	}

	result := make([]map[string]interface{}, 0, len(req.Variations))
	for _, v := range req.Variations {
		result = append(result, map[string]interface{}{
			"SellerSku": v.SKU,
			"price":     v.Price,
			"quantity":  v.Stock,
		})
	}
	return result
}

// buildTiktokImages builds TikTok image structure
func (s *CreateService) buildTiktokImages(imageIDs []string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(imageIDs))
	for _, id := range imageIDs {
		result = append(result, map[string]interface{}{"id": id})
	}
	return result
}

// buildTiktokSKUs builds TikTok SKU structure
func (s *CreateService) buildTiktokSKUs(req CreateProductRequest) []map[string]interface{} {
	if len(req.Variations) == 0 {
		return []map[string]interface{}{{
			"seller_sku":     req.SKU,
			"original_price": req.Price,
			"stock_infos":    []map[string]interface{}{{"available_stock": req.Stock}},
		}}
	}

	result := make([]map[string]interface{}, 0, len(req.Variations))
	for _, v := range req.Variations {
		result = append(result, map[string]interface{}{
			"seller_sku":     v.SKU,
			"original_price": v.Price,
			"stock_infos":    []map[string]interface{}{{"available_stock": v.Stock}},
		})
	}
	return result
}

func (s *CreateService) parseShopeeResponse(resp map[string]interface{}) *CreateProductResult {
	if itemID, ok := resp["item_id"].(float64); ok {
		return &CreateProductResult{Success: true, ItemID: int64(itemID)}
	}
	if errMsg, ok := resp["error"].(string); ok {
		return &CreateProductResult{Success: false, Error: errMsg}
	}
	return &CreateProductResult{Success: false, Error: fmt.Sprintf("unknown Shopee response format: %v", resp)}
}

func (s *CreateService) parseLazadaResponse(resp map[string]interface{}) *CreateProductResult {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if itemID, ok := data["item_id"].(float64); ok {
			return &CreateProductResult{Success: true, ItemID: int64(itemID)}
		}
	}
	return &CreateProductResult{Success: false, Error: fmt.Sprintf("failed to parse Lazada response: %v", resp)}
}

func (s *CreateService) parseTiktokResponse(resp map[string]interface{}) *CreateProductResult {
	if data, ok := resp["data"].(map[string]interface{}); ok {
		if productID, ok := data["product_id"].(string); ok {
			return &CreateProductResult{Success: true, Message: productID}
		}
	}
	return &CreateProductResult{Success: false, Error: fmt.Sprintf("failed to parse TikTok response: %v", resp)}
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen]
}
