package master_product

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHandler_BatchUpdateSkus(t *testing.T) {
	t.Run("missing tenant ID returns 401", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/skus/batch", nil)

		handler.BatchUpdateSkus(c)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})

	t.Run("valid tenant no DB returns 500 (DB check before JSON bind)", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/skus/batch",
			strings.NewReader(`not-json`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")

		handler.BatchUpdateSkus(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		// DB check happens before JSON bind, so no-DB env returns 500
		var resp map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, false, resp["success"])
	})
}

func TestHandler_UpdateSkuWithTenant(t *testing.T) {
	t.Run("valid tenant but no DB returns 500", func(t *testing.T) {
		handler := NewHandler("")

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest(http.MethodPut, "/api/master-products/skus/1",
			strings.NewReader(`{"price": 10000, "stock": 5}`))
		c.Request.Header.Set("Content-Type", "application/json")
		c.Set("tenantID", "test-tenant")
		c.Params = gin.Params{{Key: "id", Value: "1"}}

		handler.UpdateSku(c)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestUpdateSkuInput_StructFields(t *testing.T) {
	t.Run("struct can be instantiated with pointer fields and marshaled", func(t *testing.T) {
		price := 15000.0
		stock := 10
		input := UpdateSkuInput{
			Price: &price,
			Stock: &stock,
		}
		assert.Equal(t, 15000.0, *input.Price)
		assert.Equal(t, 10, *input.Stock)

		data, err := json.Marshal(input)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "price")
		assert.Contains(t, m, "stock")
	})

	t.Run("nil fields omitted from JSON", func(t *testing.T) {
		input := UpdateSkuInput{}
		data, _ := json.Marshal(input)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		_, hasPrice := m["price"]
		_, hasStock := m["stock"]
		assert.False(t, hasPrice, "nil price should be omitted")
		assert.False(t, hasStock, "nil stock should be omitted")
	})
}

func TestBatchUpdateSkusInput_StructFields(t *testing.T) {
	t.Run("struct can be instantiated with sku_ids", func(t *testing.T) {
		price := 20000.0
		stock := 50
		input := BatchUpdateSkusInput{
			SkuIDs: []uint{1, 2, 3},
			Price:  &price,
			Stock:  &stock,
		}
		assert.Equal(t, []uint{1, 2, 3}, input.SkuIDs)
		assert.Equal(t, 20000.0, *input.Price)
		assert.Equal(t, 50, *input.Stock)

		data, err := json.Marshal(input)
		assert.NoError(t, err)
		var m map[string]interface{}
		json.Unmarshal(data, &m)
		assert.Contains(t, m, "sku_ids")
	})
}
