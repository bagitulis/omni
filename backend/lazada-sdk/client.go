package lazadasdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	lazada "github.com/omni/backend/pkg/lazada"
)

// Client wraps the low-level Lazada client with context-aware helpers.
type Client struct {
	api    *lazada.Client
	appKey string
}

func NewClient(cfg Config) (*Client, error) {
	if cfg.AppKey == "" {
		return nil, errors.New("app_key is required")
	}
	if cfg.AppSecret == "" {
		return nil, errors.New("app_secret is required")
	}
	if cfg.Region == "" {
		cfg.Region = "id"
	}

	api := lazada.NewClient(cfg.AppKey, cfg.AppSecret, cfg.Region)
	if cfg.AccessToken != "" {
		api.SetAccessToken(cfg.AccessToken)
	}

	return &Client{api: api, appKey: cfg.AppKey}, nil
}

func (c *Client) SetAccessToken(token string) {
	c.api.SetAccessToken(token)
}

func requireContext(ctx context.Context) error {
	if ctx == nil {
		return errors.New("context is required")
	}
	return ctx.Err()
}

func (c *Client) apiKey() string {
	return c.appKey
}

func (c *Client) callRaw(ctx context.Context, method, path string, params map[string]string, out interface{}) error {
	if err := requireContext(ctx); err != nil {
		return err
	}
	if path == "" || !strings.HasPrefix(path, "/") {
		return fmt.Errorf("path must start with '/'")
	}
	if method == "" {
		method = "GET"
	}
	return c.api.RawRequest(ctx, method, path, params, out)
}
