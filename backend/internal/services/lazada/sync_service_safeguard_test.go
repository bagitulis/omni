package lazada

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"unsafe"

	"github.com/omni/backend/internal/models"
	lazadaPkg "github.com/omni/backend/pkg/lazada"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLazadaSyncProducts_SkipsDeleteWhenProductListEmpty(t *testing.T) {
	db := setupLazadaTestDB(t)
	ctx := context.Background()

	require.NoError(t, db.WithContext(ctx).Create(&models.LazadaProduct{
		TenantID: "tenant1",
		ItemID:   "existing-item-1",
		Name:     "Existing Product",
		Status:   "active",
		Price:    10000,
	}).Error)

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if request.URL.Path != "/products/get" {
			writer.WriteHeader(http.StatusNotFound)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		_, _ = writer.Write([]byte(`{"code":"0","data":{"total_products":0,"products":[]}}`))
	}))
	defer server.Close()

	client := lazadaPkg.NewClient("test-app-key", "test-app-secret", "id")
	setLazadaClientBaseURL(client, server.URL)

	service := NewSyncServiceWithTenant(client, db, "tenant1")

	count, err := service.SyncProducts(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, count)

	var productCount int64
	require.NoError(t, db.WithContext(ctx).
		Model(&models.LazadaProduct{}).
		Where("tenant_id = ?", "tenant1").
		Count(&productCount).Error)
	assert.Equal(t, int64(1), productCount)
}

func setLazadaClientBaseURL(client *lazadaPkg.Client, baseURL string) {
	value := reflect.ValueOf(client).Elem().FieldByName("baseURL")
	reflect.NewAt(value.Type(), unsafe.Pointer(value.UnsafeAddr())).Elem().SetString(baseURL)
}
