package handlers

// LazadaTokenResponse represents Lazada OAuth token response
type LazadaTokenResponse struct {
	AccessToken      string `json:"access_token"`
	RefreshToken     string `json:"refresh_token"`
	Country          string `json:"country"`
	RefreshExpiresIn int64  `json:"refresh_expires_in"`
	ExpiresIn        int64  `json:"expires_in"`
	AccountPlatform  string `json:"account_platform"`
	Account          string `json:"account"`
	Code             string `json:"code"`
	RequestID        string `json:"request_id"`
}

// TiktokTokenResponse represents TikTok OAuth token response
type TiktokTokenResponse struct {
	Code      int             `json:"code"`
	Message   string          `json:"message"`
	Data      TiktokTokenData `json:"data"`
	RequestID string          `json:"request_id"`
}

// TiktokTokenData represents the data field in TikTok token response
type TiktokTokenData struct {
	AccessToken          string `json:"access_token"`
	RefreshToken         string `json:"refresh_token"`
	AccessTokenExpireIn  int64  `json:"access_token_expire_in"`
	RefreshTokenExpireIn int64  `json:"refresh_token_expire_in"`
	OpenID               string `json:"open_id"`
	SellerName           string `json:"seller_name"`
}

// ShopeeTokenResponse represents Shopee OAuth token response
type ShopeeTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpireIn     int64  `json:"expire_in"`
	RequestID    string `json:"request_id"`
	Error        string `json:"error"`
	Message      string `json:"message"`
	PartnerID    int64  `json:"partner_id"`
	ShopID       int64  `json:"shop_id"`
}
