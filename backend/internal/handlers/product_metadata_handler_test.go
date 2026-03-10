package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// TestProductMetadataHandler_AllEndpointsReturn501 verifies that all product metadata
// endpoints return 501 Not Implemented since they previously returned hardcoded dummy data.
func TestProductMetadataHandler_AllEndpointsReturn501(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewProductMetadataHandler("")

	tests := []struct {
		name    string
		method  string
		path    string
		handler gin.HandlerFunc
	}{
		{"GetShopeeCategories", "GET", "/shopee/categories", handler.GetShopeeCategories},
		{"GetShopeeAttributes", "GET", "/shopee/attributes/123", handler.GetShopeeAttributes},
		{"GetShopeeBrands", "GET", "/shopee/brands/123", handler.GetShopeeBrands},
		{"GetShopeeLogistics", "GET", "/shopee/logistics", handler.GetShopeeLogistics},
		{"GetLazadaCategories", "GET", "/lazada/categories", handler.GetLazadaCategories},
		{"GetLazadaAttributes", "GET", "/lazada/attributes/123", handler.GetLazadaAttributes},
		{"GetLazadaBrands", "GET", "/lazada/brands/123", handler.GetLazadaBrands},
		{"GetTiktokCategories", "GET", "/tiktok/categories", handler.GetTiktokCategories},
		{"GetTiktokAttributes", "GET", "/tiktok/attributes/123", handler.GetTiktokAttributes},
		{"GetTiktokBrands", "GET", "/tiktok/brands", handler.GetTiktokBrands},
		{"GetTiktokWarehouses", "GET", "/tiktok/warehouses", handler.GetTiktokWarehouses},
		{"UploadImage", "POST", "/upload-image", handler.UploadImage},
		{"ValidateProduct", "POST", "/validate", handler.ValidateProduct},
		{"GetTemplates", "GET", "/templates", handler.GetTemplates},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			// Set tenantID so we don't stop at the auth check
			r.Use(func(c *gin.Context) {
				c.Set("tenantID", "test-tenant")
				c.Next()
			})
			if tt.method == "GET" {
				r.GET(tt.path, tt.handler)
			} else {
				r.POST(tt.path, tt.handler)
			}

			req, _ := http.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			assert.Equal(t, http.StatusNotImplemented, w.Code, "Expected 501 for %s", tt.name)

			var resp map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &resp)
			assert.NoError(t, err)
			assert.Equal(t, false, resp["success"])
		})
	}
}
