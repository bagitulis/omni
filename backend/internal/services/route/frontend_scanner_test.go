package route

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRoutePath(t *testing.T) {
	assert.Equal(t, "/api/orders", NormalizeRoutePath("/orders", true))
	assert.Equal(t, "/api/orders/:param", NormalizeRoutePath("/orders/${orderSn}", true))
	assert.Equal(t, "/api/orders/:param", NormalizeRoutePath("/api/orders/:id", true))
	assert.Equal(t, "", NormalizeRoutePath("https://example.com/api/orders", true))
}

func TestNormalizeRoutePath_TableDriven(t *testing.T) {
	testCases := []struct {
		name            string
		input           string
		ensureAPIPrefix bool
		expected        string
	}{
		{
			name:            "trailing slash normalized",
			input:           "/orders/",
			ensureAPIPrefix: true,
			expected:        "/api/orders",
		},
		{
			name:            "query string stripped",
			input:           "/orders?id=1",
			ensureAPIPrefix: true,
			expected:        "/api/orders",
		},
		{
			name:            "frontend template segment normalized",
			input:           "/orders/${order_sn}",
			ensureAPIPrefix: true,
			expected:        "/api/orders/:param",
		},
		{
			name:            "backend param segment normalized",
			input:           "/api/orders/:id",
			ensureAPIPrefix: false,
			expected:        "/api/orders/:param",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, NormalizeRoutePath(tc.input, tc.ensureAPIPrefix))
		})
	}
}

func TestParseFrontendRoutes(t *testing.T) {
	content := `
import api from "./client"
export async function run(orderSn: string) {
  await api.get("/orders")
  await apiClient.post("/products/${orderSn}")
}
`

	routes := parseFrontendRoutes(content, "orders.ts")
	if assert.Len(t, routes, 2) {
		assert.Equal(t, "GET", routes[0].Method)
		assert.Equal(t, "/api/orders", routes[0].Endpoint)
		assert.Equal(t, "orders.ts", routes[0].Source)

		assert.Equal(t, "POST", routes[1].Method)
		assert.Equal(t, "/api/products/:param", routes[1].Endpoint)
	}
}

func TestParseFrontendRoutes_NormalizationCases(t *testing.T) {
	content := `
import api from "./client"
export async function run(orderSn: string) {
  await api.get("/orders/")
  await api.get("/orders?id=1")
  await api.get("/orders/${orderSn}")
}
`

	routes := parseFrontendRoutes(content, "orders.ts")
	if assert.Len(t, routes, 3) {
		assert.Equal(t, "/api/orders", routes[0].Endpoint)
		assert.Equal(t, "/api/orders", routes[1].Endpoint)
		assert.Equal(t, "/api/orders/:param", routes[2].Endpoint)
	}
}
