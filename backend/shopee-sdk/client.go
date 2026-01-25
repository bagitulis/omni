package shopeesdk

import (
	"context"
	"errors"

	shopee "github.com/omni/backend/pkg/shopee"
)

// Client wraps the low-level Shopee client with context-aware helpers.
type Client struct {
	api *shopee.Client
}

// NewClient builds a Shopee client using shop-scoped credentials.
func NewClient(cfg Config) (*Client, error) {
	if cfg.PartnerID == 0 {
		return nil, errors.New("partner_id is required")
	}
	if cfg.PartnerKey == "" {
		return nil, errors.New("partner_key is required")
	}

	api := shopee.NewClient(cfg.PartnerID, cfg.PartnerKey, cfg.UseProduction)
	if cfg.ShopID > 0 && cfg.AccessToken != "" {
		api.SetShopCredentials(cfg.ShopID, cfg.AccessToken)
	}

	return &Client{api: api}, nil
}

// SetShopCredentials updates shop ID and access token.
func (c *Client) SetShopCredentials(shopID int64, accessToken string) {
	c.api.SetShopCredentials(shopID, accessToken)
}

// apiReady ensures context not cancelled before network calls.
func apiReady(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	return ctx.Err()
}
