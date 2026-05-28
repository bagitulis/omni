package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/middleware"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

const (
	oauthAttemptPending   = "pending"
	oauthAttemptCompleted = "completed"
	oauthAttemptReplayed  = "replayed"
	oauthAttemptExpired   = "expired"
	oauthAttemptCancelled = "cancelled"
	oauthAttemptFailed    = "failed"
)

type OAuthHandler struct {
	oauthRepo    *repositories.OAuthRepository
	configRepo   *repositories.GlobalConfigRepository
	platformRepo *repositories.PlatformConfigRepository
	db           *gorm.DB
	frontendURL  string
	basePath     string
}

func NewOAuthHandler(
	oauthRepo *repositories.OAuthRepository,
	configRepo *repositories.GlobalConfigRepository,
	platformRepo *repositories.PlatformConfigRepository,
	frontendURL string,
) *OAuthHandler {
	basePath := os.Getenv("DB_BASE_PATH")
	if basePath == "" {
		basePath = "./data"
	}
	var db *gorm.DB
	if oauthRepo != nil {
		db = oauthRepo.DB()
	}
	return &OAuthHandler{oauthRepo: oauthRepo, configRepo: configRepo, platformRepo: platformRepo, db: db, frontendURL: frontendURL, basePath: basePath}
}

func (h *OAuthHandler) InitiateAuth(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}
	platform := c.Param("platform")
	if !isSupportedOAuthPlatform(platform) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid platform"})
		return
	}

	state, err := h.createBoundOAuthState(c, tenantID, platform)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	authURL, err := h.buildPlatformAuthURL(c, platform, state.State)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": gin.H{"auth_url": authURL, "attempt_id": state.AttemptID, "expires_at": state.ExpiresAt}})
}

func (h *OAuthHandler) HandleCallback(c *gin.Context) {
	platform := c.Param("platform")
	stateValue := c.Query("state")
	code := c.Query("code")

	state, status, err := h.validateCallbackState(c, platform, stateValue)
	if err != nil || state == nil {
		h.logOAuthCallback(c, "", platform, models.OAuthStatusFailed, status)
		c.JSON(oauthCallbackStatusCode(status), gin.H{"success": false, "error": sanitizeOAuthErrorText(status)})
		return
	}
	h.logOAuthCallback(c, state.TenantID, state.Platform, models.OAuthStatusReceived, "callback_received")

	exchangeErr := h.exchangeAndReconcile(c, state, code)
	if exchangeErr != nil {
		_ = h.markAttempt(c, state, oauthAttemptFailed)
		h.logOAuthCallback(c, state.TenantID, state.Platform, models.OAuthStatusFailed, sanitizeOAuthError(exchangeErr))
		h.redirectOAuthResult(c, state.RedirectURL, sanitizeOAuthError(exchangeErr))
		return
	}

	_ = h.markAttempt(c, state, oauthAttemptCompleted)
	h.redirectOAuthResult(c, state.RedirectURL, "success")
}

func (h *OAuthHandler) createBoundOAuthState(c *gin.Context, tenantID, platform string) (*models.OAuthState, error) {
	if h.oauthRepo == nil {
		return nil, fmt.Errorf("oauth repository not configured")
	}
	attemptID := uuid.New().String()
	nonce, err := oauth.NewNonce()
	if err != nil {
		return nil, err
	}
	intent := c.DefaultQuery("intent", "connect")
	storeID := c.Query("store_identifier")
	redirectPath := sanitizeRedirectPath(c.Query("redirect_path"))
	redirectURL := h.frontendURL + redirectPath
	expiresAt := oauth.OAuthStateExpiry()
	claims := oauth.StateClaims{
		TenantID:     tenantID,
		Platform:     platform,
		Marketplace:  platform,
		AttemptID:    attemptID,
		Intent:       intent,
		StoreID:      storeID,
		UserID:       middleware.GetUserID(c),
		SessionID:    c.GetHeader("X-Session-ID"),
		CSRFNonce:    nonce,
		Nonce:        nonce,
		RedirectPath: redirectPath,
		RedirectURI:  redirectPath,
		ExpiresAt:    expiresAt.Unix(),
	}
	signedState, err := oauth.BuildSignedState(claims)
	if err != nil {
		return nil, err
	}
	state, err := h.oauthRepo.CreateBoundState(c.Request.Context(), repositories.OAuthStateCreateParams{
		TenantID: tenantID, Platform: platform, AttemptID: attemptID, Intent: intent, StoreID: storeID,
		UserID: claims.UserID, SessionID: claims.SessionID, CSRFNonce: nonce, State: signedState,
		RedirectURL: redirectURL, ExpiresAt: expiresAt,
	})
	if err != nil {
		return nil, err
	}
	err = h.credentialRepo(c).CreateAttempt(c.Request.Context(), &models.OAuthConnectionAttempt{
		TenantID: tenantID, Platform: platform, AttemptID: attemptID, Status: oauthAttemptPending,
		Intent: intent, IntendedStoreID: storeID, SignedState: signedState, CSRFNonce: nonce,
		RedirectPath: redirectPath, ExpiresAt: expiresAt, CreatedBy: claims.UserID,
	})
	if err != nil {
		return nil, err
	}
	return state, nil
}

func (h *OAuthHandler) buildPlatformAuthURL(c *gin.Context, platform, state string) (string, error) {
	switch platform {
	case models.PlatformShopee:
		creds, err := h.configRepo.GetShopeeCredentials(c.Request.Context())
		if err != nil {
			return "", err
		}
		callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/shopee", h.getBackendURL(c))
		return oauth.NewShopeeOAuthService(creds.PartnerID, creds.PartnerKey, callbackURL, false).GetAuthURL(state), nil
	case models.PlatformLazada:
		creds, err := h.configRepo.GetLazadaCredentials(c.Request.Context())
		if err != nil {
			return "", err
		}
		callbackURL := fmt.Sprintf("%s/api/platform-auth/callback/lazada", h.getBackendURL(c))
		return oauth.NewLazadaOAuthService(creds.AppKey, creds.AppSecret, callbackURL, false).GetAuthURL(state), nil
	case models.PlatformTiktok:
		creds, err := h.configRepo.GetTiktokCredentials(c.Request.Context())
		if err != nil {
			return "", err
		}
		callbackURL := oauth.TiktokCallbackURL(h.getBackendURL(c))
		if callbackURL == "" {
			return "", fmt.Errorf("callback_url is required")
		}
		return oauth.NewTiktokOAuthService(creds.AppKey, creds.AppSecret, callbackURL, false).GetAuthURL(state), nil
	default:
		return "", fmt.Errorf("invalid platform")
	}
}

func (h *OAuthHandler) validateCallbackState(c *gin.Context, callbackPlatform, stateValue string) (*models.OAuthState, string, error) {
	currentTenantID := middleware.GetTenantID(c)
	currentUserID := c.GetString("userID")
	if currentTenantID == "" || currentUserID == "" {
		return nil, "no_session", fmt.Errorf("no_session")
	}
	claims, err := oauth.ParseSignedState(stateValue)
	if err != nil {
		return nil, err.Error(), err
	}
	if err := validateStateClaimsForCallback(claims); err != nil {
		return nil, err.Error(), err
	}
	if claims.TenantID != currentTenantID {
		return nil, "tenant_mismatch", fmt.Errorf("tenant_mismatch")
	}
	if claims.UserID != currentUserID {
		return nil, "user_mismatch", fmt.Errorf("user_mismatch")
	}
	if claims.Platform != callbackPlatform {
		return nil, "wrong_platform", fmt.Errorf("wrong_platform")
	}
	if claims.Marketplace != "" && claims.Marketplace != claims.Platform {
		return nil, "wrong_platform", fmt.Errorf("wrong_platform")
	}
	attempt, err := h.credentialRepo(c).GetAttempt(c.Request.Context(), claims.TenantID, claims.Platform, claims.AttemptID)
	if err != nil || attempt == nil || attempt.CSRFNonce != claims.Nonce || attempt.SignedState != stateValue {
		return nil, "invalid_attempt", fmt.Errorf("invalid_attempt")
	}
	if time.Now().After(attempt.ExpiresAt) {
		_ = h.markAttemptByClaims(c, claims, oauthAttemptExpired)
		return nil, "expired_state", fmt.Errorf("expired_state")
	}
	if attempt.Status != oauthAttemptPending || attempt.CompletedAt != nil {
		return nil, "replayed_state", fmt.Errorf("replayed_state")
	}
	state, status, err := h.oauthRepo.ConsumeState(c.Request.Context(), stateValue)
	if err != nil || state == nil {
		if err == nil {
			err = fmt.Errorf("%s", status)
		}
		return nil, status, err
	}
	if status != oauthAttemptCompleted {
		_ = h.markAttempt(c, state, status)
		return nil, status, fmt.Errorf("%s", status)
	}
	if !stateMatchesClaims(state, claims) {
		_ = h.markAttempt(c, state, oauthAttemptFailed)
		return nil, "wrong_scope", fmt.Errorf("wrong_scope")
	}
	if attempt.CSRFNonce != state.CSRFNonce || claims.Nonce != state.CSRFNonce {
		return nil, "invalid_attempt", fmt.Errorf("invalid_attempt")
	}
	if state.UserID != currentUserID {
		_ = h.markAttempt(c, state, oauthAttemptFailed)
		return nil, "user_mismatch", fmt.Errorf("user_mismatch")
	}
	if !redirectURLAllowed(h.frontendURL, state.RedirectURL, claims.RedirectURI) {
		_ = h.markAttempt(c, state, oauthAttemptFailed)
		return nil, "invalid_redirect", fmt.Errorf("invalid_redirect")
	}
	return state, oauthAttemptCompleted, nil
}

func (h *OAuthHandler) exchangeAndReconcile(c *gin.Context, state *models.OAuthState, code string) error {
	switch state.Platform {
	case models.PlatformShopee:
		shopID := c.Query("shop_id")
		if err := h.ensureStoreWriteAllowed(c, state, shopID); err != nil {
			return err
		}
		return h.exchangeShopeeToken(c, state.TenantID, code, shopID)
	case models.PlatformLazada:
		if err := h.ensureStoreWriteAllowed(c, state, c.Query("seller_id")); err != nil {
			return err
		}
		return h.exchangeLazadaToken(c, state.TenantID, code)
	case models.PlatformTiktok:
		if err := h.ensureStoreWriteAllowed(c, state, c.Query("shop_id")); err != nil {
			return err
		}
		return h.exchangeTiktokToken(c, state.TenantID, code)
	default:
		return fmt.Errorf("invalid_platform")
	}
}

func (h *OAuthHandler) ensureStoreWriteAllowed(c *gin.Context, state *models.OAuthState, providerStoreID string) error {
	if providerStoreID == "" {
		return nil
	}
	if state.StoreID != "" && state.StoreID != providerStoreID {
		return fmt.Errorf("store_mismatch")
	}
	existing, err := h.credentialRepo(c).GetConnection(c.Request.Context(), state.TenantID, state.Platform, providerStoreID)
	if err != nil {
		return err
	}
	if existing != nil && !(state.Intent == "reconnect" && state.StoreID == providerStoreID) {
		return fmt.Errorf("duplicate_store")
	}
	return nil
}

func (h *OAuthHandler) credentialRepo(c *gin.Context) *repositories.CredentialRepository {
	return repositories.NewCredentialRepository(h.db)
}

func (h *OAuthHandler) markAttempt(c *gin.Context, state *models.OAuthState, status string) error {
	if state == nil || status == "" {
		return nil
	}
	return h.credentialRepo(c).CompleteAttempt(c.Request.Context(), state.TenantID, state.Platform, state.AttemptID, status)
}

func (h *OAuthHandler) markAttemptByClaims(c *gin.Context, claims oauth.StateClaims, status string) error {
	if status == "" {
		return nil
	}
	return h.credentialRepo(c).CompleteAttempt(c.Request.Context(), claims.TenantID, claims.Platform, claims.AttemptID, status)
}

func (h *OAuthHandler) logOAuthCallback(c *gin.Context, tenantID, platform, status, code string) {
	if h.oauthRepo == nil || tenantID == "" || platform == "" {
		return
	}
	_ = h.oauthRepo.CreateLog(c.Request.Context(), &models.OAuthLog{TenantID: tenantID, Platform: platform, EventType: models.OAuthEventCallback, Status: status, ErrorMsg: sanitizeOAuthErrorText(code)})
}

func (h *OAuthHandler) redirectOAuthResult(c *gin.Context, redirectURL, result string) {
	if redirectURL == "" {
		redirectURL = h.frontendURL + "/settings"
	}
	parsed, err := url.Parse(redirectURL)
	if err != nil {
		parsed, _ = url.Parse(h.frontendURL + "/settings")
	}
	query := parsed.Query()
	if result == "success" {
		query.Set("success", "true")
	} else {
		query.Set("error", sanitizeOAuthErrorText(result))
	}
	parsed.RawQuery = query.Encode()
	c.Redirect(http.StatusFound, parsed.String())
}

func (h *OAuthHandler) getBackendURL(c *gin.Context) string {
	scheme := "https"
	if c.Request.TLS == nil {
		scheme = "http"
	}
	return fmt.Sprintf("%s://%s", scheme, c.Request.Host)
}

func (h *OAuthHandler) GetOAuthLogs(c *gin.Context) {
	tenantID := middleware.GetTenantID(c)
	if tenantID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"success": false, "error": "Missing tenant_id"})
		return
	}
	platform := c.Query("platform")
	limit := 50
	var logs []models.OAuthLog
	var err error
	if platform != "" {
		logs, err = h.oauthRepo.FindLogsByPlatform(c.Request.Context(), tenantID, platform, limit)
	} else {
		logs, err = h.oauthRepo.FindLogsByTenant(c.Request.Context(), tenantID, limit)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": logs})
}

func isSupportedOAuthPlatform(platform string) bool {
	return platform == models.PlatformShopee || platform == models.PlatformLazada || platform == models.PlatformTiktok
}

func sanitizeRedirectPath(path string) string {
	if path == "" {
		return "/settings"
	}
	if !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return "/settings"
	}
	if !isAllowedOAuthRedirectPath(path) {
		return "/settings"
	}
	return path
}

func isAllowedOAuthRedirectPath(path string) bool {
	switch path {
	case "/settings", "/settings/platforms":
		return true
	default:
		return false
	}
}

func oauthCallbackStatusCode(status string) int {
	switch status {
	case "no_session":
		return http.StatusUnauthorized
	case "tenant_mismatch", "user_mismatch":
		return http.StatusForbidden
	case "expired_state", "expired", "replayed_state", "invalid_state", "invalid_attempt", "wrong_platform", "wrong_scope", "invalid_redirect":
		return http.StatusBadRequest
	default:
		return http.StatusBadRequest
	}
}

func stateMatchesClaims(state *models.OAuthState, claims oauth.StateClaims) bool {
	return state.TenantID == claims.TenantID && state.Platform == claims.Platform && state.AttemptID == claims.AttemptID && state.Intent == claims.Intent && state.StoreID == claims.StoreID && state.CSRFNonce == claims.CSRFNonce && state.ExpiresAt.Unix() == claims.ExpiresAt
}

func sanitizeOAuthError(err error) string {
	if err == nil {
		return "failed"
	}
	return sanitizeOAuthErrorText(err.Error())
}

func sanitizeOAuthErrorText(value string) string {
	safe := strings.ToLower(strings.TrimSpace(value))
	safe = strings.ReplaceAll(safe, " ", "_")
	allowed := map[string]bool{"success": true, "no_session": true, "tenant_mismatch": true, "user_mismatch": true, "invalid_state": true, "invalid_redirect": true, "expired_state": true, "expired": true, "replayed": true, "replayed_state": true, "wrong_platform": true, "wrong_scope": true, "invalid_attempt": true, "duplicate_store": true, "store_mismatch": true, "invalid_platform": true, "failed": true}
	if allowed[safe] {
		return safe
	}
	log.Warn().Str("oauth_error_code", "redacted").Msg("OAuth callback failed")
	return "failed"
}
