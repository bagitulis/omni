package config

import (
	"fmt"

	"github.com/rs/zerolog/log"
)

// ShopeeCredentials holds Shopee platform credentials
type ShopeeCredentials struct {
	PartnerID      string
	PartnerKey     string
	PushPartnerKey string
}

// TiktokCredentials holds TikTok platform credentials
type TiktokCredentials struct {
	AppKey    string
	AppSecret string
}

// LazadaCredentials holds Lazada platform credentials
type LazadaCredentials struct {
	AppKey    string
	AppSecret string
}

// GetShopeeCredentials returns Shopee platform credentials
func (s *GlobalConfigService) GetShopeeCredentials() (*ShopeeCredentials, error) {
	partnerId, _ := s.GetConfig("shopee", "partnerId")
	partnerKey, _ := s.GetConfig("shopee", "partnerKey")
	pushPartnerKey, _ := s.GetConfig("shopee", "pushPartnerKey")

	if partnerId == "" || partnerKey == "" {
		log.Warn().Str("platform", "shopee").Msg("missing credentials in global_config")
		return nil, fmt.Errorf("shopee credentials not found in global_config: partnerId=%q, partnerKey=%q", partnerId, partnerKey)
	}

	return &ShopeeCredentials{
		PartnerID:      partnerId,
		PartnerKey:     partnerKey,
		PushPartnerKey: pushPartnerKey,
	}, nil
}

// GetTiktokCredentials returns TikTok platform credentials
func (s *GlobalConfigService) GetTiktokCredentials() (*TiktokCredentials, error) {
	appKey, _ := s.GetConfig("tiktok", "appKey")
	appSecret, _ := s.GetConfig("tiktok", "appSecret")

	if appKey == "" || appSecret == "" {
		log.Warn().Str("platform", "tiktok").Msg("missing credentials in global_config")
		return nil, fmt.Errorf("tiktok credentials not found in global_config: appKey=%q, appSecret=%q", appKey, appSecret)
	}

	return &TiktokCredentials{
		AppKey:    appKey,
		AppSecret: appSecret,
	}, nil
}

// GetLazadaCredentials returns Lazada platform credentials
func (s *GlobalConfigService) GetLazadaCredentials() (*LazadaCredentials, error) {
	appKey, _ := s.GetConfig("lazada", "appKey")
	appSecret, _ := s.GetConfig("lazada", "appSecret")

	if appKey == "" || appSecret == "" {
		log.Warn().Str("platform", "lazada").Msg("missing credentials in global_config")
		return nil, fmt.Errorf("lazada credentials not found in global_config: appKey=%q, appSecret=%q", appKey, appSecret)
	}

	return &LazadaCredentials{
		AppKey:    appKey,
		AppSecret: appSecret,
	}, nil
}

// HasShopeeCredentials checks if Shopee credentials are configured
func (s *GlobalConfigService) HasShopeeCredentials() bool {
	creds, err := s.GetShopeeCredentials()
	if err != nil {
		return false
	}
	return creds.PartnerID != "" && creds.PartnerKey != ""
}

// HasTiktokCredentials checks if TikTok credentials are configured
func (s *GlobalConfigService) HasTiktokCredentials() bool {
	creds, err := s.GetTiktokCredentials()
	if err != nil {
		return false
	}
	return creds.AppKey != "" && creds.AppSecret != ""
}

// HasLazadaCredentials checks if Lazada credentials are configured
func (s *GlobalConfigService) HasLazadaCredentials() bool {
	creds, err := s.GetLazadaCredentials()
	if err != nil {
		return false
	}
	return creds.AppKey != "" && creds.AppSecret != ""
}
