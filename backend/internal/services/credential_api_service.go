package services

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

type CredentialApiService struct {
	db *gorm.DB
}

type CredentialPlatformStatus struct {
	Platform      string                                      `json:"platform"`
	Region        string                                      `json:"region,omitempty"`
	Status        string                                      `json:"status"`
	AppConfigured bool                                        `json:"app_configured"`
	Stores        []models.CredentialConnectionMaskedResponse `json:"stores"`
}

type CredentialAppUpsertRequest struct {
	Platform        string `json:"-"`
	Region          string `json:"region,omitempty"`
	StoreIdentifier string `json:"store_identifier,omitempty"`
	AppKey          string `json:"app_key,omitempty"`
	AppSecret       string `json:"app_secret,omitempty"`
	PartnerID       int64  `json:"partner_id,omitempty"`
	PartnerKey      string `json:"partner_key,omitempty"`
	Reason          string `json:"reason"`
}

type CredentialOAuthInitiateRequest struct {
	Platform        string `json:"-"`
	Intent          string `json:"intent"`
	StoreIdentifier string `json:"store_identifier,omitempty"`
	RedirectPath    string `json:"redirect_path"`
}

type CredentialOAuthReconnectRequest struct {
	Platform        string `json:"-"`
	StoreIdentifier string `json:"-"`
	Intent          string `json:"intent"`
	RedirectPath    string `json:"redirect_path"`
}

type CredentialConnectionActionRequest struct {
	Platform        string `json:"-"`
	StoreIdentifier string `json:"-"`
	Reason          string `json:"reason"`
	RevokeRemote    bool   `json:"revoke_remote,omitempty"`
}

type CredentialManualTokenRequest struct {
	Platform        string `json:"-"`
	StoreIdentifier string `json:"store_identifier"`
	Region          string `json:"region,omitempty"`
	AccessToken     string `json:"access_token"`
	RefreshToken    string `json:"refresh_token,omitempty"`
	ExpiresAt       string `json:"expires_at,omitempty"`
	ShopCipher      string `json:"shop_cipher,omitempty"`
	Reason          string `json:"reason"`
}

type CredentialOAuthAttemptResponse struct {
	AuthURL   string `json:"auth_url"`
	AttemptID string `json:"attempt_id"`
	ExpiresAt string `json:"expires_at"`
}

type CredentialMutationResponse struct {
	Platform            string `json:"platform"`
	StoreIdentifierMask string `json:"store_identifier_mask,omitempty"`
	Status              string `json:"status"`
	Configured          bool   `json:"configured,omitempty"`
	SecretMask          string `json:"secret_mask,omitempty"`
	AuditEventID        string `json:"audit_event_id,omitempty"`
	RemoteRevokeStatus  string `json:"remote_revoke_status,omitempty"`
	LastRefreshAt       string `json:"last_refresh_at,omitempty"`
	ExpiresAt           string `json:"expires_at,omitempty"`
}

type CredentialAuditListResponse struct {
	Events []models.CredentialAuditEvent `json:"events"`
}

func NewCredentialApiService(db *gorm.DB) *CredentialApiService { return &CredentialApiService{db: db} }

func (s *CredentialApiService) GetPlatformStatus(ctx context.Context, tenantID, role, userID, platform, storeIdentifier string) ([]CredentialPlatformStatus, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	repo := repositories.NewCredentialRepository(s.db)
	if platform == "" {
		platforms := []CredentialPlatformStatus{{Platform: models.PlatformShopee, Status: "disconnected"}, {Platform: models.PlatformLazada, Status: "disconnected", Region: "id"}, {Platform: models.PlatformTiktok, Status: "disconnected"}}
		for i := range platforms {
			app, err := repo.GetAppConfig(ctx, tenantID, platforms[i].Platform)
			if err != nil {
				return nil, err
			}
			conns, err := repo.ListConnections(ctx, tenantID, platforms[i].Platform)
			if err != nil {
				return nil, err
			}
			platforms[i].Status = "disconnected"
			if len(conns) > 0 {
				platforms[i].Status = "connected"
			}
			platforms[i].AppConfigured = app != nil
			platforms[i].Stores = maskedConnectionsForPlatform(conns, "")
			if app != nil {
				platforms[i].Region = app.Region
			}
		}
		return platforms, nil
	}
	app, err := repo.GetAppConfig(ctx, tenantID, platform)
	if err != nil {
		return nil, err
	}
	conns, err := repo.ListConnections(ctx, tenantID, platform)
	if err != nil {
		return nil, err
	}
	status := "disconnected"
	if len(conns) > 0 {
		status = "connected"
	}
	platforms := []CredentialPlatformStatus{{Platform: platform, Status: status, AppConfigured: app != nil}}
	for i := range platforms {
		platforms[i].Stores = maskedConnectionsForPlatform(conns, storeIdentifier)
		if app != nil {
			platforms[i].Region = app.Region
		}
	}
	return platforms, nil
}

func (s *CredentialApiService) UpsertAppCredential(ctx context.Context, tenantID, role, userID string, req CredentialAppUpsertRequest, rotate bool) (*CredentialMutationResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	if role != "developer" && role != "admin" {
		return nil, fmt.Errorf("forbidden")
	}
	repo := repositories.NewCredentialRepository(s.db)
	cfg, err := repo.GetAppConfig(ctx, tenantID, req.Platform)
	if err != nil {
		return nil, err
	}
	if cfg == nil {
		cfg = &models.CredentialAppConfig{TenantID: tenantID, Platform: req.Platform}
	}
	cfg.Region = normalizeRegion(req.Region)
	cfg.UpdatedBy = userID
	if cfg.CreatedBy == "" {
		cfg.CreatedBy = userID
	}
	switch req.Platform {
	case models.PlatformShopee:
		cfg.PartnerID = req.PartnerID
		cfg.PartnerKey = req.PartnerKey
	case models.PlatformLazada, models.PlatformTiktok:
		cfg.AppKey = req.AppKey
		cfg.AppSecret = req.AppSecret
	}
	cfg.Configured = true
	if err := repo.UpsertAppConfig(ctx, cfg); err != nil {
		return nil, err
	}
	masked := cfg.ToMaskedResponse()
	secretMask := "configured"
	if rotate {
		secretMask = "rotated"
	}
	return &CredentialMutationResponse{Platform: req.Platform, Status: "connected", Configured: masked.Configured, SecretMask: secretMask, AuditEventID: uuid.NewString()}, nil
}

func (s *CredentialApiService) InitiateOAuth(ctx context.Context, tenantID, role, userID string, req CredentialOAuthInitiateRequest) (*CredentialOAuthAttemptResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	if req.RedirectPath == "" {
		return nil, fmt.Errorf("redirect_path is required")
	}
	attemptID := uuid.NewString()
	return &CredentialOAuthAttemptResponse{AuthURL: "https://example.invalid/oauth/" + req.Platform, AttemptID: attemptID, ExpiresAt: time.Now().Add(15 * time.Minute).Format(time.RFC3339)}, nil
}

func (s *CredentialApiService) ReconnectOAuth(ctx context.Context, tenantID, role, userID string, req CredentialOAuthReconnectRequest) (*CredentialOAuthAttemptResponse, error) {
	if req.StoreIdentifier == "" {
		return nil, fmt.Errorf("store_identifier is required")
	}
	return s.InitiateOAuth(ctx, tenantID, role, userID, CredentialOAuthInitiateRequest{Platform: req.Platform, Intent: req.Intent, StoreIdentifier: req.StoreIdentifier, RedirectPath: req.RedirectPath})
}

func (s *CredentialApiService) ChangeConnectionStatus(ctx context.Context, tenantID, role, userID string, req CredentialConnectionActionRequest, action string) (*CredentialMutationResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	if req.StoreIdentifier == "" {
		return nil, fmt.Errorf("store_identifier is required")
	}
	repo := repositories.NewCredentialRepository(s.db)
	conn, err := repo.GetConnection(ctx, tenantID, req.Platform, req.StoreIdentifier)
	if err != nil {
		return nil, err
	}
	if conn == nil {
		return nil, fmt.Errorf("connection not found")
	}
	if action == "refresh" {
		if err := repo.UpdateConnectionStatus(ctx, tenantID, req.Platform, req.StoreIdentifier, "connected"); err != nil {
			return nil, err
		}
		return &CredentialMutationResponse{Platform: req.Platform, StoreIdentifierMask: conn.StoreIdentifierMask(), Status: "connected", LastRefreshAt: time.Now().Format(time.RFC3339), ExpiresAt: time.UnixMilli(conn.TokenExpiry).Format(time.RFC3339)}, nil
	}
	if err := repo.DisableConnectionWithVersion(ctx, conn, conn.Version, req.Reason); err != nil {
		return nil, err
	}
	return &CredentialMutationResponse{Platform: req.Platform, StoreIdentifierMask: conn.StoreIdentifierMask(), Status: "disconnected", RemoteRevokeStatus: "not_supported", AuditEventID: uuid.NewString()}, nil
}

func (s *CredentialApiService) ListAuditEvents(ctx context.Context, tenantID, role, userID, platform, storeIdentifier, limitStr, cursor string) (*CredentialAuditListResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	limit := 50
	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	repo := repositories.NewCredentialRepository(s.db)
	events, err := repo.ListAuditEvents(ctx, tenantID, platform, limit, 0)
	if err != nil {
		return nil, err
	}
	return &CredentialAuditListResponse{Events: events}, nil
}

func (s *CredentialApiService) ApplyManualToken(ctx context.Context, tenantID, role, userID string, req CredentialManualTokenRequest) (*CredentialMutationResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	if role != "developer" && role != "admin" {
		return nil, fmt.Errorf("forbidden")
	}
	if req.Reason == "" {
		return nil, fmt.Errorf("reason is required")
	}
	repo := repositories.NewCredentialRepository(s.db)
	conn := &models.CredentialConnection{TenantID: tenantID, Platform: req.Platform, StoreIdentifier: req.StoreIdentifier, Region: normalizeRegion(req.Region), AccessToken: req.AccessToken, RefreshToken: req.RefreshToken, ShopCipher: req.ShopCipher, Status: "connected", CreatedBy: userID, UpdatedBy: userID}
	if conn.Region == "" {
		conn.Region = "id"
	}
	if err := repo.CreateConnection(ctx, conn); err != nil {
		return nil, err
	}
	return &CredentialMutationResponse{Platform: req.Platform, StoreIdentifierMask: conn.StoreIdentifierMask(), Status: conn.Status, Configured: true, AuditEventID: uuid.NewString()}, nil
}

func maskedConnectionsForPlatform(conns []models.CredentialConnection, storeIdentifier string) []models.CredentialConnectionMaskedResponse {
	result := make([]models.CredentialConnectionMaskedResponse, 0, len(conns))
	for _, conn := range conns {
		if storeIdentifier != "" && conn.StoreIdentifier != storeIdentifier {
			continue
		}
		result = append(result, conn.ToMaskedResponse())
	}
	return result
}

func normalizeRegion(region string) string {
	if strings.EqualFold(region, "indonesia") || region == "" {
		return "id"
	}
	return region
}
