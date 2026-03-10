package handlers

import (
	"fmt"

	"github.com/omni/backend/internal/config"
	"github.com/omni/backend/internal/services"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	shopeePkg "github.com/omni/backend/pkg/shopee"
	tiktokPkg "github.com/omni/backend/pkg/tiktok"
)

// platformOrderTable maps platform names to their DB table names.
var platformOrderTable = map[string]string{
	"shopee": "ShopeeOrder",
	"tiktok": "TiktokOrder",
	"lazada": "LazadaOrder",
}

// ===== Platform Client Helpers =====

func (h *OrderManagerHandler) getShopeeClient(tenantID string) (*shopeePkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "shopee")
	if err != nil {
		return nil, err
	}
	client := shopeePkg.NewClient(creds.PartnerID, creds.PartnerKey, creds.IsProduction)
	client.SetShopCredentials(creds.ShopID, creds.AccessToken)
	return client, nil
}

func (h *OrderManagerHandler) getTikTokClient(tenantID string) (*tiktokPkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "tiktok")
	if err != nil {
		return nil, err
	}
	client := tiktokPkg.NewClient(creds.AppKey, creds.AppSecret)
	client.SetCredentials(creds.AccessToken, creds.ShopCipher)
	return client, nil
}

func (h *OrderManagerHandler) getLazadaClient(tenantID string) (*lazadaPkg.Client, error) {
	credService := services.NewCredentialService(h.basePath)
	creds, err := credService.GetPlatformCredentials(tenantID, "lazada")
	if err != nil {
		return nil, err
	}
	region := creds.Region
	if region == "" {
		region = "id"
	}
	client := lazadaPkg.NewClient(creds.AppKey, creds.AppSecret, region)
	client.SetAccessToken(creds.AccessToken)
	return client, nil
}

// ===== Status Helpers =====

func (h *OrderManagerHandler) validateOrderStatus(tenantID, orderSN, platform string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}

	table, ok := platformOrderTable[platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	var status string
	if err := db.Table(table).Select("order_status").Where("order_sn = ?", orderSN).Scan(&status).Error; err != nil {
		return err
	}

	expectedStatuses := map[string][]string{
		"shopee": {"READY_TO_SHIP"},
		"tiktok": {"AWAITING_SHIPMENT"},
		"lazada": {"pending", "packed"},
	}

	allowed := expectedStatuses[platform]
	for _, s := range allowed {
		if status == s {
			return nil
		}
	}
	return fmt.Errorf("order status is %s, expected one of %v", status, allowed)
}

func (h *OrderManagerHandler) updateOrderStatus(tenantID, orderSN, status, platform string) error {
	db, err := config.GetTenantDB(tenantID, h.basePath)
	if err != nil {
		return err
	}

	table, ok := platformOrderTable[platform]
	if !ok {
		return fmt.Errorf("unsupported platform: %s", platform)
	}

	return db.Table(table).Where("order_sn = ?", orderSN).Update("order_status", status).Error
}

// ===== Utility Functions =====

func allFailed(orderSNs []string, errMsg string) []BulkShipItemResult {
	results := make([]BulkShipItemResult, len(orderSNs))
	for i, sn := range orderSNs {
		results[i] = BulkShipItemResult{
			OrderSN: sn,
			Status:  "failed",
			Error:   errMsg,
		}
	}
	return results
}

func buildBulkSummary(results []BulkShipItemResult) BulkShipSummary {
	s := BulkShipSummary{Total: len(results)}
	for _, r := range results {
		switch r.Status {
		case "shipped", "shipped_but_local_failed":
			s.Shipped++
		default:
			s.Failed++
		}
	}
	return s
}
