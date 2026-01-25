package lazadasdk

import (
	"encoding/json"
	"time"

	lazada "github.com/omni/backend/pkg/lazada"
)

type Config struct {
	AppKey      string
	AppSecret   string
	Region      string // id|my|sg|th|vn|ph
	AccessToken string
}

// RawCall represents an arbitrary Lazada REST invocation.
// Path should start with "/..." and is appended to the Lazada "/rest" base.
type RawCall struct {
	Method string
	Path   string
	Params map[string]string
}

// RawResponse is the common Lazada API envelope.
type RawResponse struct {
	Code      string          `json:"code"`
	Type      string          `json:"type,omitempty"`
	Message   string          `json:"message,omitempty"`
	RequestID string          `json:"request_id,omitempty"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// GetOrdersParams matches /orders/get.
type GetOrdersParams struct {
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	UpdatedAfter  *time.Time
	UpdatedBefore *time.Time
	Status        string
	SortDirection string
	SortBy        string
	Offset        int
	Limit         int
}

// Type aliases to reuse existing transport models where we already have them.
type (
	TokenResponse       = lazada.TokenResponse
	OrderListResponse   = lazada.OrderListResponse
	ProductListResponse = lazada.ProductListResponse
	Product             = lazada.Product
	ProductSku          = lazada.ProductSku
)
