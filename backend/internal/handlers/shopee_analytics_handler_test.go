package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/omni/backend/internal/dto"
	"github.com/stretchr/testify/assert"
)

// mockShopeeService is a test double implementing ShopeeAnalyticsService
type mockShopeeService struct {
	settings       *dto.AnalyticsSettingsDTO
	syncStatus     *dto.SyncStatusDTO
	reconciliation *dto.ReconciliationResultDTO
	shippingFee    *dto.ShopeeShippingFeeResultDTO
	jobID          string
	err            error
}

func (m *mockShopeeService) GetSettings(_ context.Context, _ string) (*dto.AnalyticsSettingsDTO, error) {
	return m.settings, m.err
}

func (m *mockShopeeService) SaveSettings(_ context.Context, _ string, settings *dto.AnalyticsSettingsDTO) (*dto.AnalyticsSettingsDTO, error) {
	return settings, m.err
}

func (m *mockShopeeService) GetSyncStatus(_ context.Context, _ string, _, _ int) (*dto.SyncStatusDTO, error) {
	return m.syncStatus, m.err
}

func (m *mockShopeeService) SyncEscrow(_ context.Context, _ string, _, _ int, _ bool) (string, error) {
	return m.jobID, m.err
}

func (m *mockShopeeService) DeleteSyncData(_ context.Context, _ string, _, _ int) error {
	return m.err
}

func (m *mockShopeeService) GetReconciliation(_ context.Context, _ string, _, _ int) (*dto.ReconciliationResultDTO, error) {
	return m.reconciliation, m.err
}

func (m *mockShopeeService) GetShippingFeeAnalysis(_ context.Context, _ string, _, _ int) (*dto.ShopeeShippingFeeResultDTO, error) {
	return m.shippingFee, m.err
}

func (m *mockShopeeService) RepopulateItems(_ context.Context, _ string, _ string) error {
	return m.err
}

func TestShopeeAnalyticsHandler_GetSettings_MissingTenant_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/settings", nil)

	handler.GetSettings(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], "Missing tenant_id")
}

func TestShopeeAnalyticsHandler_GetSettings_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{
			settings: &dto.AnalyticsSettingsDTO{
				PriceColumn:        "price",
				FormulaDeduction:   5000,
				FormulaMultiplier:  1.0,
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/settings", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.GetSettings(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "price", data["price_column"])
	assert.Equal(t, 5000.0, data["formula_deduction"])
	assert.Equal(t, 1.0, data["formula_multiplier"])
}

func TestShopeeAnalyticsHandler_GetSettings_ServiceError_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{
			err: assert.AnError,
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/settings", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.GetSettings(c)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
	assert.Contains(t, resp["error"], assert.AnError.Error())
}

func TestShopeeAnalyticsHandler_SaveSettings_MissingTenant_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/settings", nil)

	handler.SaveSettings(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestShopeeAnalyticsHandler_SaveSettings_ValidBody_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{},
	}

	body := `{"price_column": "price", "formula_deduction": 5000, "formula_multiplier": 1.5}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant_id", "test-tenant-123")

	handler.SaveSettings(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "price", data["price_column"])
	assert.Equal(t, 5000.0, data["formula_deduction"])
	assert.Equal(t, 1.5, data["formula_multiplier"])
}

func TestShopeeAnalyticsHandler_SaveSettings_InvalidBody_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{},
	}

	// Invalid JSON body
	body := `{invalid json}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/settings", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant_id", "test-tenant-123")

	handler.SaveSettings(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, false, resp["success"])
}

func TestShopeeAnalyticsHandler_GetSyncStatus_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{
			syncStatus: &dto.SyncStatusDTO{
				Synced:       true,
				TotalOrders:  100,
				FailedOrders: 2,
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/sync-status", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.GetSyncStatus(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, true, data["synced"])
	assert.Equal(t, float64(100), data["total_orders"])
}

func TestShopeeAnalyticsHandler_SyncEscrow_ValidTenant_Returns202(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{
			jobID: "job-12345",
		},
	}

	body := `{"month": 1, "year": 2025, "force_resync": true}`
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/sync", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set("tenant_id", "test-tenant-123")

	handler.SyncEscrow(c)

	assert.Equal(t, http.StatusAccepted, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	assert.Equal(t, "job-12345", data["job_id"])
}

func TestShopeeAnalyticsHandler_DeleteSyncData_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/sync?month=1&year=2025", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.DeleteSyncData(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Contains(t, resp["message"], "deleted")
}

func TestShopeeAnalyticsHandler_GetReconciliation_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{
			reconciliation: &dto.ReconciliationResultDTO{
				Summary: dto.ReconciliationSummaryDTO{
					TotalSKU: 10,
					SKUOk:    8,
				},
				Details: []dto.SkuGroupDTO{
					{SKU: "SKU001", TotalQuantity: 5, TotalAmount: 100.0},
				},
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/reconciliation?month=1&year=2025", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.GetReconciliation(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	data := resp["data"].(map[string]interface{})
	summary := data["summary"].(map[string]interface{})
	assert.Equal(t, float64(10), summary["total_sku"])
	assert.Equal(t, float64(8), summary["sku_ok"])
}

func TestShopeeAnalyticsHandler_GetShippingFeeAnalysis_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{
			shippingFee: &dto.ShopeeShippingFeeResultDTO{
				Summary: dto.ShippingFeeSummaryDTO{
					TotalOrders:          50,
					OrdersWithDifference: 3,
					NetImpact:            -15000.0,
				},
				Details: []dto.ShopeeShippingOrderDTO{
					{OrderSN: "ORD001", PlatformFee: 10000, ActualFee: 8000, Difference: -2000},
				},
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/shipping-fee?month=1&year=2025", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.GetShippingFeeAnalysis(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
}

func TestShopeeAnalyticsHandler_RepopulateItems_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{
		svc: &mockShopeeService{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/repopulate-items?period=2025-01", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.RepopulateItems(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Contains(t, resp["message"], "repopulation")
}

func TestShopeeAnalyticsHandler_SyncEscrow_MissingTenant_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/sync", nil)

	handler.SyncEscrow(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestShopeeAnalyticsHandler_GetReconciliation_MissingTenant_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &ShopeeAnalyticsHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/reconciliation", nil)

	handler.GetReconciliation(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
