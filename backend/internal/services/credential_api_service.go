package services

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/omni/backend/internal/services/platform"
	"github.com/omni/backend/internal/services/sync"
)

func NewCredentialApiService(dbPath string) *CredentialApiService { return &CredentialApiService{dbPath: dbPath} }

func (s *CredentialApiService) GetPlatformStatus(ctx context.Context, tenantID, role, userID, platform, storeIdentifier string) ([]CredentialPlatformStatus, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	tenantDB, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}
	repo := repositories.NewCredentialRepository(tenantDB)
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
	if role != "developer" && role != "admin" && role != "owner" {
		return nil, fmt.Errorf("forbidden")
	}
	tenantDB, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}
	repo := repositories.NewCredentialRepository(tenantDB)
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
	eventType := "app_credential_upsert"
	if rotate {
		eventType = "app_credential_rotate"
	}
	auditEvent := &models.CredentialAuditEvent{
		TenantID: tenantID, Platform: req.Platform,
		EventType: eventType, Status: "success", Actor: userID, ActorRole: role,
		Metadata: models.JSONMap{"reason": req.Reason},
	}
	auditEventID := ""
	if err := repo.CreateAuditEvent(ctx, auditEvent); err != nil {
		log.Error().Err(err).Str("event_type", auditEvent.EventType).Msg("Failed to persist audit event")
	} else {
		auditEventID = auditEvent.ID
	}
	return &CredentialMutationResponse{Platform: req.Platform, Status: "connected", Configured: masked.Configured, SecretMask: secretMask, AuditEventID: auditEventID}, nil
}

func (s *CredentialApiService) InitiateOAuth(ctx context.Context, tenantID, role, userID, callbackBaseURL string, req CredentialOAuthInitiateRequest) (*CredentialOAuthAttemptResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	if req.RedirectPath == "" {
		return nil, fmt.Errorf("redirect_path is required")
	}
	if role != "developer" && role != "admin" && role != "owner" {
		return nil, fmt.Errorf("forbidden")
	}
	tenantDB, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}
	repo := repositories.NewCredentialRepository(tenantDB)
	attemptID := uuid.NewString()

	// Build CSRF nonce and signed state
	csrfNonce, err := oauth.NewNonce()
	if err != nil {
		return nil, fmt.Errorf("generate csrf nonce: %w", err)
	}
	expiresAt := time.Now().Add(10 * time.Minute)
	claims := oauth.StateClaims{
		TenantID:     tenantID,
		Platform:     req.Platform,
		Marketplace:  req.Platform,
		AttemptID:    attemptID,
		Intent:       req.Intent,
		StoreID:      req.StoreIdentifier,
		UserID:       userID,
		CSRFNonce:    csrfNonce,
		Nonce:        csrfNonce,
		RedirectPath: req.RedirectPath,
		RedirectURI:  req.RedirectPath,
		ExpiresAt:    expiresAt.Unix(),
	}
	signedState, err := oauth.BuildSignedState(claims)
	if err != nil {
		return nil, fmt.Errorf("build signed state: %w", err)
	}

	// Persist OAuth connection attempt
	attempt := &models.OAuthConnectionAttempt{
		TenantID:        tenantID,
		Platform:        req.Platform,
		AttemptID:       attemptID,
		Status:          "pending",
		Intent:          req.Intent,
		IntendedStoreID: req.StoreIdentifier,
		SignedState:     signedState,
		CSRFNonce:       csrfNonce,
		RedirectPath:    req.RedirectPath,
		ExpiresAt:       expiresAt,
		CreatedBy:       userID,
	}
	if err := repo.CreateAttempt(ctx, attempt); err != nil {
		return nil, fmt.Errorf("create oauth attempt: %w", err)
	}

	// Get app credentials to build real platform auth URL
	if callbackBaseURL == "" {
		callbackBaseURL = os.Getenv("APP_URL")
	}
	if callbackBaseURL == "" {
		callbackBaseURL = "https://yndigital.my.id"
		log.Warn().Msg("InitiateOAuth: APP_URL not set, using fallback")
	}
	callbackURL := callbackBaseURL + "/api/credentials/callback/" + req.Platform
	var authURL string

	switch req.Platform {
	case models.PlatformShopee:
		cfg, _ := repo.GetAppConfig(ctx, tenantID, req.Platform)
		if cfg != nil && cfg.PartnerID > 0 && cfg.PartnerKey != "" {
			svc := oauth.NewShopeeOAuthService(cfg.PartnerID, cfg.PartnerKey, callbackURL, false)
			authURL = svc.GetAuthURL(signedState)
		} else {
			return nil, fmt.Errorf("shopee app credentials not configured for tenant %s", tenantID)
		}
	case models.PlatformLazada:
		cfg, _ := repo.GetAppConfig(ctx, tenantID, req.Platform)
		if cfg != nil && cfg.AppKey != "" && cfg.AppSecret != "" {
			svc := oauth.NewLazadaOAuthService(cfg.AppKey, cfg.AppSecret, callbackURL, false)
			authURL = svc.GetAuthURL(signedState)
		} else {
			return nil, fmt.Errorf("lazada app credentials not configured for tenant %s", tenantID)
		}
	case models.PlatformTiktok:
		cfg, _ := repo.GetAppConfig(ctx, tenantID, req.Platform)
		if cfg != nil && cfg.AppKey != "" && cfg.AppSecret != "" {
			svc := oauth.NewTiktokOAuthService(cfg.AppKey, cfg.AppSecret, callbackURL, false)
			authURL = svc.GetAuthURL(signedState)
		} else {
			return nil, fmt.Errorf("tiktok app credentials not configured for tenant %s", tenantID)
		}
	default:
		return nil, fmt.Errorf("unsupported platform: %s", req.Platform)
	}

	// Write audit event
	auditEvent := &models.CredentialAuditEvent{
		TenantID: tenantID, Platform: req.Platform,
		StoreIdentifier: req.StoreIdentifier,
		EventType:       "oauth_initiate", Status: "initiated",
		Actor: userID, ActorRole: role,
		Metadata: models.JSONMap{"intent": req.Intent, "attempt_id": attemptID},
	}
	auditEventID := ""
	if err := repo.CreateAuditEvent(ctx, auditEvent); err != nil {
		log.Error().Err(err).Str("event_type", auditEvent.EventType).Msg("Failed to persist audit event")
	} else {
		auditEventID = auditEvent.ID
	}

	return &CredentialOAuthAttemptResponse{
		AuthURL:      authURL,
		AttemptID:    attemptID,
		ExpiresAt:    expiresAt.Format(time.RFC3339),
		AuditEventID: auditEventID,
	}, nil
}

func (s *CredentialApiService) ReconnectOAuth(ctx context.Context, tenantID, role, userID, callbackBaseURL string, req CredentialOAuthReconnectRequest) (*CredentialOAuthAttemptResponse, error) {
	if req.StoreIdentifier == "" {
		return nil, fmt.Errorf("store_identifier is required")
	}
	return s.InitiateOAuth(ctx, tenantID, role, userID, callbackBaseURL, CredentialOAuthInitiateRequest{
		Platform: req.Platform, Intent: req.Intent, StoreIdentifier: req.StoreIdentifier, RedirectPath: req.RedirectPath,
	})
}

func (s *CredentialApiService) ChangeConnectionStatus(ctx context.Context, tenantID, role, userID string, req CredentialConnectionActionRequest, action string) (*CredentialMutationResponse, error) {
	if tenantID == "" {
		return nil, fmt.Errorf("Missing tenant_id")
	}
	if role != "developer" && role != "admin" && role != "owner" {
		return nil, fmt.Errorf("forbidden")
	}
	if req.StoreIdentifier == "" {
		return nil, fmt.Errorf("store_identifier is required")
	}
	tenantDB, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}
	repo := repositories.NewCredentialRepository(tenantDB)
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
		invalidateCredentialCaches(tenantID)
		auditEvent := &models.CredentialAuditEvent{
			TenantID: tenantID, Platform: req.Platform, StoreIdentifier: req.StoreIdentifier,
			EventType: "connection_refresh", Status: "success", Actor: userID, ActorRole: role,
			Metadata: models.JSONMap{"reason": req.Reason},
		}
		auditEventID := ""
		if err := repo.CreateAuditEvent(ctx, auditEvent); err != nil {
			log.Error().Err(err).Str("event_type", auditEvent.EventType).Msg("Failed to persist audit event")
		} else {
			auditEventID = auditEvent.ID
		}
		return &CredentialMutationResponse{Platform: req.Platform, StoreIdentifierMask: conn.StoreIdentifierMask(), Status: "connected", LastRefreshAt: time.Now().Format(time.RFC3339), ExpiresAt: time.UnixMilli(conn.TokenExpiry).Format(time.RFC3339), AuditEventID: auditEventID}, nil
	}
	if err := repo.DisableConnectionWithVersion(ctx, conn, conn.Version, req.Reason); err != nil {
		return nil, err
	}
	invalidateCredentialCaches(tenantID)
	auditEvent := &models.CredentialAuditEvent{
		TenantID: tenantID, Platform: req.Platform, StoreIdentifier: req.StoreIdentifier,
		EventType: "connection_disconnect", Status: "success", Actor: userID, ActorRole: role,
		Metadata: models.JSONMap{"reason": req.Reason},
	}
	auditEventID := ""
	if err := repo.CreateAuditEvent(ctx, auditEvent); err != nil {
		log.Error().Err(err).Str("event_type", auditEvent.EventType).Msg("Failed to persist audit event")
	} else {
		auditEventID = auditEvent.ID
	}
	return &CredentialMutationResponse{Platform: req.Platform, StoreIdentifierMask: conn.StoreIdentifierMask(), Status: "disconnected", RemoteRevokeStatus: "not_supported", AuditEventID: auditEventID}, nil
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
	tenantDB, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}
	repo := repositories.NewCredentialRepository(tenantDB)
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
	if role != "developer" && role != "admin" && role != "owner" {
		return nil, fmt.Errorf("forbidden")
	}
	if req.Reason == "" {
		return nil, fmt.Errorf("reason is required")
	}
	tenantDB, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("get tenant DB: %w", err)
	}
	repo := repositories.NewCredentialRepository(tenantDB)

	// Check for existing connection (including disabled ones — upsert pattern)
	var existing *models.CredentialConnection
	allConns, _ := repo.ListConnections(ctx, tenantID, req.Platform)
	for _, c := range allConns {
		if c.StoreIdentifier == req.StoreIdentifier {
			existing = &c
			break
		}
	}

	region := normalizeRegion(req.Region)
	if region == "" {
		region = "id"
	}

	var conn *models.CredentialConnection
	if existing != nil {
		// Update existing connection with new token
		existing.AccessToken = req.AccessToken
		existing.RefreshToken = req.RefreshToken
		existing.ShopCipher = req.ShopCipher
		existing.Status = "connected"
		existing.UpdatedBy = userID
		wasDisabled := existing.DisabledAt != nil
		existing.DisabledAt = nil
		existing.DisabledReason = ""
		existing.Region = region
		if req.ExpiresAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("invalid expires_at: must be RFC3339 format")
			}
			existing.TokenExpiry = parsed.UnixMilli()
		}
		conn = existing
		if wasDisabled {
			// Re-enable disabled connection
			if err := repo.ReconnectWithToken(ctx, conn); err != nil {
				return nil, err
			}
			if err := repo.UpdateManualConnectionWithVersion(ctx, conn, conn.Version); err != nil {
				return nil, err
			}
	} else {
		// Create new connection
		conn = &models.CredentialConnection{TenantID: tenantID, Platform: req.Platform, StoreIdentifier: req.StoreIdentifier, Region: region, AccessToken: req.AccessToken, RefreshToken: req.RefreshToken, ShopCipher: req.ShopCipher, Status: "connected", CreatedBy: userID, UpdatedBy: userID}
		if req.ExpiresAt != "" {
			parsed, err := time.Parse(time.RFC3339, req.ExpiresAt)
			if err != nil {
				return nil, fmt.Errorf("invalid expires_at: must be RFC3339 format")
			}
			conn.TokenExpiry = parsed.UnixMilli()
		}
		if err := repo.CreateConnection(ctx, conn); err != nil {
			return nil, err
		}
	}
	invalidateCredentialCaches(tenantID)
	auditEvent := &models.CredentialAuditEvent{
		TenantID: tenantID, Platform: req.Platform, StoreIdentifier: req.StoreIdentifier,
		EventType: "manual_token_apply", Status: "success", Actor: userID, ActorRole: role,
		Metadata: models.JSONMap{"reason": req.Reason, "action": func() string { if existing != nil { return "updated" }; return "created" }()},
	}
	auditEventID := ""
	if err := repo.CreateAuditEvent(ctx, auditEvent); err != nil {
		log.Error().Err(err).Str("event_type", auditEvent.EventType).Msg("Failed to persist audit event")
	} else {
		auditEventID = auditEvent.ID
	}
	return &CredentialMutationResponse{Platform: req.Platform, StoreIdentifierMask: conn.StoreIdentifierMask(), Status: conn.Status, Configured: true, AuditEventID: auditEventID}, nil
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

func invalidateCredentialCaches(tenantID string) {
	platform.InvalidateTenantPlatformService(tenantID)
	sync.InvalidateTenantInstance(tenantID)
}
