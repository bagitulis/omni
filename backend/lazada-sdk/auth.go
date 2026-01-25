package lazadasdk

import (
	"context"
	"net/url"
)

const AuthorizeURL = "https://auth.lazada.com/oauth/authorize"

type AuthURLParams struct {
	RedirectURI string
	State       string
	UUID        string
	Country     string // id|my|sg|th|vn|ph
	ForceAuth   bool
}

// BuildAuthURL builds the Lazada OAuth authorization URL (response_type=code).
func (c *Client) BuildAuthURL(p AuthURLParams) (string, error) {
	if p.RedirectURI == "" {
		return "", ErrBadRequest("redirect_uri is required")
	}

	v := url.Values{}
	v.Set("client_id", c.apiKey())
	v.Set("redirect_uri", p.RedirectURI)
	v.Set("response_type", "code")
	if p.ForceAuth {
		v.Set("force_auth", "true")
	}
	if p.State != "" {
		v.Set("state", p.State)
	}
	if p.UUID != "" {
		v.Set("uuid", p.UUID)
	}
	if p.Country != "" {
		v.Set("country", p.Country)
	}

	return AuthorizeURL + "?" + v.Encode(), nil
}

// GenerateAccessToken exchanges an authorization code for an access token.
// Endpoint: POST /auth/token/create (auth host).
func (c *Client) GenerateAccessToken(ctx context.Context, code string, uuid string) (*TokenResponse, error) {
	params := map[string]string{
		"code": code,
	}
	if uuid != "" {
		params["uuid"] = uuid
	}

	var resp TokenResponse
	if err := c.callRaw(ctx, "POST", "/auth/token/create", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// RefreshAccessToken refreshes a Lazada access token.
// Endpoint: POST /auth/token/refresh (auth host).
func (c *Client) RefreshAccessToken(ctx context.Context, refreshToken string, uuid string) (*TokenResponse, error) {
	params := map[string]string{
		"refresh_token": refreshToken,
	}
	if uuid != "" {
		params["uuid"] = uuid
	}

	var resp TokenResponse
	if err := c.callRaw(ctx, "POST", "/auth/token/refresh", params, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}
