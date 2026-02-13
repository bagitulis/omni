package label

import (
	"context"
	"fmt"
	"strings"
	"unicode"

	"github.com/omni/backend/internal/config"
	lazadaHandler "github.com/omni/backend/internal/handlers/lazada"
	"github.com/omni/backend/internal/models"
	shopeeService "github.com/omni/backend/internal/services/shopee"
	tiktokService "github.com/omni/backend/internal/services/tiktok"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// LabelResult represents a unified shipping label result across all platforms
type LabelResult struct {
	OrderSN      string `json:"order_sn"`
	Platform     string `json:"platform"`
	Status       string `json:"status"`
	FileData     string `json:"file_data,omitempty"`     // Base64 encoded PDF or document URL
	ErrorMessage string `json:"error_message,omitempty"` // Error description if failed
}

// LabelService provides unified multi-platform label printing
type LabelService struct {
	basePath string
}

// NewLabelService creates a new LabelService
func NewLabelService(basePath string) *LabelService {
	return &LabelService{basePath: basePath}
}

// GetLabel retrieves a shipping label for an order, routing to the correct platform.
// If platform is empty, it auto-detects by searching the DB.
func (s *LabelService) GetLabel(ctx context.Context, tenantID, orderSN, platform string) *LabelResult {
	if platform == "" {
		detected, err := s.detectPlatform(ctx, tenantID, orderSN)
		if err != nil {
			return &LabelResult{
				OrderSN:      orderSN,
				Platform:     "unknown",
				Status:       "FAILED",
				ErrorMessage: fmt.Sprintf("failed to detect platform: %s", err.Error()),
			}
		}
		platform = detected
	}

	switch platform {
	case "shopee":
		return s.getShopeeLabel(ctx, tenantID, orderSN)
	case "tiktok":
		return s.getTikTokLabel(ctx, tenantID, orderSN)
	case "lazada":
		return s.getLazadaLabel(ctx, tenantID, orderSN)
	default:
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     platform,
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("unsupported platform: %s", platform),
		}
	}
}

// detectPlatform searches platform order tables to find which platform an order belongs to
func (s *LabelService) detectPlatform(ctx context.Context, tenantID, orderSN string) (string, error) {
	db, err := config.GetTenantDB(tenantID, s.basePath)
	if err != nil {
		return "", fmt.Errorf("failed to get tenant DB: %w", err)
	}

	if found := s.existsInTable(ctx, db, &models.ShopeeOrder{}, orderSN); found {
		return "shopee", nil
	}
	if found := s.existsInTable(ctx, db, &models.TiktokOrder{}, orderSN); found {
		return "tiktok", nil
	}
	if found := s.existsInTable(ctx, db, &models.LazadaOrder{}, orderSN); found {
		return "lazada", nil
	}

	var orderToday models.OrderTodayItem
	if err := db.WithContext(ctx).
		Where("order_sn = ?", orderSN).
		Order("updated_at DESC").
		First(&orderToday).Error; err == nil {
		platform := strings.ToLower(strings.TrimSpace(orderToday.Platform))
		if platform == "shopee" || platform == "tiktok" || platform == "lazada" {
			return platform, nil
		}
	}

	if isLikelyTikTokOrderID(orderSN) {
		log.Warn().Str("order_sn", orderSN).Msg("Platform not found in DB, defaulting to TikTok by order ID format")
		return "tiktok", nil
	}

	return "", fmt.Errorf("order %s not found in any platform", orderSN)
}

func isLikelyTikTokOrderID(orderSN string) bool {
	trimmed := strings.TrimSpace(orderSN)
	if len(trimmed) < 16 {
		return false
	}

	for _, ch := range trimmed {
		if !unicode.IsDigit(ch) {
			return false
		}
	}

	return true
}

// existsInTable checks if an order_sn exists in the given platform order table
func (s *LabelService) existsInTable(ctx context.Context, db *gorm.DB, model interface{}, orderSN string) bool {
	var count int64
	db.WithContext(ctx).Model(model).Where("order_sn = ?", orderSN).Count(&count)
	return count > 0
}

// getShopeeLabel retrieves a shipping label via Shopee's shipping service
func (s *LabelService) getShopeeLabel(ctx context.Context, tenantID, orderSN string) *LabelResult {
	svc := shopeeService.NewShippingServiceWithCreds(tenantID, s.basePath)
	result, err := svc.GetShippingLabel(ctx, orderSN, "", "THERMAL_AIR_WAYBILL")
	if err != nil {
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     "shopee",
			Status:       "FAILED",
			ErrorMessage: err.Error(),
		}
	}

	return &LabelResult{
		OrderSN:      result.OrderSN,
		Platform:     "shopee",
		Status:       result.Status,
		FileData:     result.FileData,
		ErrorMessage: result.ErrorMessage,
	}
}

// getTikTokLabel retrieves a shipping label via TikTok's shipping service
func (s *LabelService) getTikTokLabel(ctx context.Context, tenantID, orderSN string) *LabelResult {
	svc := tiktokService.NewShippingService(s.basePath)
	docURL, _, err := svc.GetShippingLabelByOrder(ctx, tenantID, orderSN, "SHIPPING_LABEL")
	if err != nil {
		// Fallback: try with packageID directly
		docURL2, err2 := svc.GetShippingLabel(ctx, tenantID, orderSN, "SHIPPING_LABEL")
		if err2 != nil {
			return &LabelResult{
				OrderSN:      orderSN,
				Platform:     "tiktok",
				Status:       "FAILED",
				ErrorMessage: fmt.Sprintf("failed to get TikTok label: %s", err.Error()),
			}
		}
		docURL = docURL2
	}

	log.Info().Str("order_sn", orderSN).Str("platform", "tiktok").Msg("Retrieved TikTok shipping label")
	return &LabelResult{
		OrderSN:  orderSN,
		Platform: "tiktok",
		Status:   "SUCCESS",
		FileData: docURL, // URL to document
	}
}

// getLazadaLabel retrieves a shipping document via Lazada's API
func (s *LabelService) getLazadaLabel(ctx context.Context, tenantID, orderSN string) *LabelResult {
	// Get order items for Lazada (Lazada uses order_item_ids for document retrieval)
	db, err := config.GetTenantDB(tenantID, s.basePath)
	if err != nil {
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     "lazada",
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to get tenant DB: %s", err.Error()),
		}
	}

	var items []models.LazadaOrderItem
	db.WithContext(ctx).Where("order_sn = ?", orderSN).Find(&items)
	if len(items) == 0 {
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     "lazada",
			Status:       "FAILED",
			ErrorMessage: "no order items found for Lazada order",
		}
	}

	// Collect item IDs
	itemIDs := make([]string, len(items))
	for i, item := range items {
		itemIDs[i] = fmt.Sprintf("%d", item.ItemID)
	}

	// Get Lazada client
	client, err := lazadaHandler.GetLazadaClient(tenantID, s.basePath)
	if err != nil {
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     "lazada",
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to get Lazada client: %s", err.Error()),
		}
	}

	// Fetch document
	docResp, err := client.GetDocument(lazadaPkg.GetDocumentRequest{
		OrderItemIDs: itemIDs,
		DocType:      "shippingLabel",
	})
	if err != nil {
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     "lazada",
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("failed to get Lazada document: %s", err.Error()),
		}
	}
	if docResp.Code != "0" && docResp.Code != "" {
		return &LabelResult{
			OrderSN:      orderSN,
			Platform:     "lazada",
			Status:       "FAILED",
			ErrorMessage: fmt.Sprintf("Lazada API error: code %s", docResp.Code),
		}
	}

	// Prefer base64 file data, fall back to URL
	fileData := docResp.Data.Document.File
	if fileData == "" {
		fileData = docResp.Data.Document.URL
	}

	log.Info().Str("order_sn", orderSN).Str("platform", "lazada").Msg("Retrieved Lazada shipping document")
	return &LabelResult{
		OrderSN:  orderSN,
		Platform: "lazada",
		Status:   "SUCCESS",
		FileData: fileData,
	}
}
