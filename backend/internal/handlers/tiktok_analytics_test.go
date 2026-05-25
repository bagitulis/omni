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

// mockTiktokService is a test double implementing TiktokAnalyticsService
type mockTiktokService struct {
	settings       *dto.AnalyticsSettingsDTO
	syncStatus     *dto.SyncStatusDTO
	reconciliation *dto.TiktokReconciliationResultDTO
	shippingFee    *dto.TiktokShippingFeeResultDTO
	jobID          string
	err            error
}

func (m *mockTiktokService) GetSettings(_ context.Context, _ string) (*dto.AnalyticsSettingsDTO, error) {
	return m.settings, m.err
}

func (m *mockTiktokService) SaveSettings(_ context.Context, _ string, settings *dto.AnalyticsSettingsDTO) (*dto.AnalyticsSettingsDTO, error) {
	return settings, m.err
}

func (m *mockTiktokService) GetSyncStatus(_ context.Context, _ string, _, _ int) (*dto.SyncStatusDTO, error) {
	return m.syncStatus, m.err
}

func (m *mockTiktokService) SyncEscrow(_ context.Context, _ string, _, _ int, _ bool) (string, error) {
	return m.jobID, m.err
}

func (m *mockTiktokService) DeleteSyncData(_ context.Context, _ string, _, _ int) error {
	return m.err
}

func (m *mockTiktokService) GetReconciliation(_ context.Context, _ string, _, _ int) (*dto.TiktokReconciliationResultDTO, error) {
	return m.reconciliation, m.err
}

func (m *mockTiktokService) GetShippingFeeAnalysis(_ context.Context, _ string, _, _ int) (*dto.TiktokShippingFeeResultDTO, error) {
	return m.shippingFee, m.err
}

func (m *mockTiktokService) RepopulateItems(_ context.Context, _ string, _ string) error {
	return m.err
}

func TestTiktokAnalyticsHandler_GetSettings_MissingTenant_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{}

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

func TestTiktokAnalyticsHandler_GetSettings_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{
			settings: &dto.AnalyticsSettingsDTO{
				PriceColumn:        "cost_price",
				FormulaDeduction:   3000,
				FormulaMultiplier:  1.2,
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
	assert.Equal(t, "cost_price", data["price_column"])
	assert.Equal(t, 3000.0, data["formula_deduction"])
	assert.Equal(t, 1.2, data["formula_multiplier"])
}

func TestTiktokAnalyticsHandler_GetSettings_ServiceError_Returns500(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{
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
}

func TestTiktokAnalyticsHandler_SaveSettings_ValidBody_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{},
	}

	body := `{"price_column": "cost_price", "formula_deduction": 3000, "formula_multiplier": 1.2}`
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
	assert.Equal(t, "cost_price", data["price_column"])
	assert.Equal(t, 3000.0, data["formula_deduction"])
	assert.Equal(t, 1.2, data["formula_multiplier"])
}

func TestTiktokAnalyticsHandler_SaveSettings_InvalidBody_Returns400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{},
	}

	body := `{bad json}`
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

func TestTiktokAnalyticsHandler_GetSyncStatus_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{
			syncStatus: &dto.SyncStatusDTO{
				Synced:       false,
				TotalOrders:  50,
				FailedOrders: 0,
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
	assert.Equal(t, false, data["synced"])
	assert.Equal(t, float64(50), data["total_orders"])
}

func TestTiktokAnalyticsHandler_SyncEscrow_ValidTenant_Returns202(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{
			jobID: "tiktok-job-67890",
		},
	}

	body := `{"month": 2, "year": 2025, "force_resync": false}`
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
	assert.Equal(t, "tiktok-job-67890", data["job_id"])
}

func TestTiktokAnalyticsHandler_DeleteSyncData_MissingTenant_Returns401(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/sync", nil)

	handler.DeleteSyncData(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestTiktokAnalyticsHandler_GetReconciliation_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{
			reconciliation: &dto.TiktokReconciliationResultDTO{
				Summary: dto.ReconciliationSummaryDTO{
					TotalSKU:          20,
					TotalTransactions: 100,
					SKUOk:             18,
				},
				Details: []dto.TiktokSkuGroupDTO{
					{SKU: "TIKTOK-SKU-001", TotalQuantity: 10, TotalAmount: 250.0},
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
	assert.Equal(t, float64(20), summary["total_sku"])
	assert.Equal(t, float64(100), summary["total_transactions"])
	assert.Equal(t, float64(18), summary["sku_ok"])
}

func TestTiktokAnalyticsHandler_GetShippingFeeAnalysis_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{
			shippingFee: &dto.TiktokShippingFeeResultDTO{
				Summary: dto.ShippingFeeSummaryDTO{
					TotalOrders:          30,
					OrdersWithDifference: 5,
					NetImpact:            -5000.0,
				},
				Details: []dto.TiktokShippingOrderDTO{
					{OrderSN: "TK-ORD-001", ShippingFee: 15000, ActualFee: 12000, Difference: -3000},
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

func TestTiktokAnalyticsHandler_RepopulateItems_ValidTenant_Returns200(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &TiktokAnalyticsHandler{
		svc: &mockTiktokService{},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/repopulate-items?period=2025-02", nil)
	c.Set("tenant_id", "test-tenant-123")

	handler.RepopulateItems(c)

	assert.Equal(t, http.StatusOK, w.Code)
	var resp map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Equal(t, true, resp["success"])
	assert.Contains(t, resp["message"], "repopulation")
}
