package handlers

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"github.com/omni/backend/internal/services/oauth"
	"github.com/omni/backend/internal/testutils"
	"github.com/omni/backend/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupOAuthSecurityTest(t *testing.T) (*OAuthHandler, *repositories.OAuthRepository, *repositories.CredentialRepository, *gin.Context) {
	t.Helper()
	key, err := utils.GenerateKey()
	require.NoError(t, err)
	t.Setenv("ENCRYPTION_KEY", key)
	t.Setenv("JWT_SECRET", "test-oauth-state-secret-32-bytes")
	db := testutils.SetupTestPostgresWithModels(t,
		&models.OAuthState{},
		&models.OAuthLog{},
		&models.OAuthConnectionAttempt{},
		&models.CredentialConnection{},
		&models.CredentialAppConfig{},
		&models.CredentialAuditEvent{},
	)
	db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_credential_connections_active ON credential_connections (tenant_id, platform, store_identifier) WHERE disabled_at IS NULL`)
	oauthRepo := repositories.NewOAuthRepository(db)
	credentialRepo := repositories.NewCredentialRepository(db)
	handler := NewOAuthHandler(oauthRepo, nil, "http://frontend.test")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/platform-auth/callback/tiktok", nil)
	c.Request.Host = "backend.test"
	c.Set("tenant_id", "tenant-a")
	c.Set("userID", "user-1")
	return handler, oauthRepo, credentialRepo, c
}

func createSignedOAuthAttempt(t *testing.T, repo *repositories.OAuthRepository, credentialRepo *repositories.CredentialRepository, tenantID, platform, intent, storeID string, expiresAt time.Time) string {
	t.Helper()
	attemptID := uuid.New().String()
	nonce, err := oauth.NewNonce()
	require.NoError(t, err)
	claims := oauth.StateClaims{TenantID: tenantID, Platform: platform, AttemptID: attemptID, Intent: intent, StoreID: storeID, UserID: "user-1", CSRFNonce: nonce, Nonce: nonce, RedirectPath: "/settings", RedirectURI: "/settings", ExpiresAt: expiresAt.Unix()}
	signedState, err := oauth.BuildSignedState(claims)
	require.NoError(t, err)
	_, err = repo.CreateBoundState(contextWithTestTimeout(t), repositories.OAuthStateCreateParams{TenantID: tenantID, Platform: platform, AttemptID: attemptID, Intent: intent, StoreID: storeID, CSRFNonce: nonce, State: signedState, RedirectURL: "http://frontend.test/settings", ExpiresAt: expiresAt})
	require.NoError(t, err)
	require.NoError(t, credentialRepo.CreateAttempt(contextWithTestTimeout(t), &models.OAuthConnectionAttempt{TenantID: tenantID, Platform: platform, AttemptID: attemptID, Status: oauthAttemptPending, Intent: intent, IntendedStoreID: storeID, SignedState: signedState, CSRFNonce: nonce, RedirectPath: "/settings", ExpiresAt: expiresAt, CreatedBy: "user-1"}))
	return signedState
}

func contextWithTestTimeout(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestOAuthSignedStateReplayAndScopeValidation(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	state := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformTiktok, "connect", "", time.Now().Add(10*time.Minute))

	validated, status, err := handler.validateCallbackState(c, models.PlatformTiktok, state)
	require.NoError(t, err)
	require.Equal(t, oauthAttemptCompleted, status)
	require.Equal(t, "tenant-a", validated.TenantID)

	replayed, replayStatus, replayErr := handler.validateCallbackState(c, models.PlatformTiktok, state)
	require.Error(t, replayErr)
	assert.Nil(t, replayed)
	assert.Equal(t, "replayed_state", replayStatus)
}

func TestOAuthSignedStateRejectsTenantMismatch(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	state := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-b", models.PlatformTiktok, "connect", "", time.Now().Add(10*time.Minute))

	validated, status, err := handler.validateCallbackState(c, models.PlatformTiktok, state)
	require.Error(t, err)
	assert.Nil(t, validated)
	assert.Equal(t, "tenant_mismatch", status)
}

func TestOAuthSignedStateRejectsNoSession(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	c.Set("tenant_id", "")
	state := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformTiktok, "connect", "", time.Now().Add(10*time.Minute))

	validated, status, err := handler.validateCallbackState(c, models.PlatformTiktok, state)
	require.Error(t, err)
	assert.Nil(t, validated)
	assert.Equal(t, "no_session", status)
}

func TestOAuthSignedStateRejectsWrongPlatform(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	state := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformTiktok, "connect", "", time.Now().Add(10*time.Minute))

	validated, status, err := handler.validateCallbackState(c, models.PlatformShopee, state)
	require.Error(t, err)
	assert.Nil(t, validated)
	assert.Equal(t, "wrong_platform", status)
}

func TestOAuthSignedStateRejectsExpiredAndCancelledAttempts(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	expiredClaims := oauth.StateClaims{TenantID: "tenant-a", Platform: models.PlatformTiktok, AttemptID: uuid.New().String(), Intent: "connect", UserID: "user-1", CSRFNonce: "nonce", Nonce: "nonce", RedirectPath: "/settings", RedirectURI: "/settings", ExpiresAt: time.Now().Add(-time.Minute).Unix()}
	expired, err := oauth.BuildSignedState(expiredClaims)
	require.NoError(t, err)
	_, status, err := handler.validateCallbackState(c, models.PlatformTiktok, expired)
	require.Error(t, err)
	assert.Equal(t, "expired_state", status)

	cancelled := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformTiktok, "connect", "", time.Now().Add(10*time.Minute))
	claims, err := oauth.ParseSignedState(cancelled)
	require.NoError(t, err)
	require.NoError(t, credentialRepo.CompleteAttempt(contextWithTestTimeout(t), claims.TenantID, claims.Platform, claims.AttemptID, oauthAttemptCancelled))
	validated, status, err := handler.validateCallbackState(c, models.PlatformTiktok, cancelled)
	require.Error(t, err)
	assert.Nil(t, validated)
	assert.Equal(t, "replayed_state", status)
}

func TestOAuthSignedStateRejectsDisallowedRedirect(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	attemptID := uuid.New().String()
	nonce, err := oauth.NewNonce()
	require.NoError(t, err)
	expiresAt := time.Now().Add(10 * time.Minute)
	claims := oauth.StateClaims{TenantID: "tenant-a", Platform: models.PlatformTiktok, Marketplace: models.PlatformTiktok, AttemptID: attemptID, Intent: "connect", UserID: "user-1", CSRFNonce: nonce, Nonce: nonce, RedirectPath: "https://evil.test/callback", RedirectURI: "https://evil.test/callback", ExpiresAt: expiresAt.Unix()}
	signedState, err := oauth.BuildSignedState(claims)
	require.NoError(t, err)
	_, err = repo.CreateBoundState(contextWithTestTimeout(t), repositories.OAuthStateCreateParams{TenantID: "tenant-a", Platform: models.PlatformTiktok, AttemptID: attemptID, Intent: "connect", UserID: "user-1", CSRFNonce: nonce, State: signedState, RedirectURL: "https://evil.test/callback", ExpiresAt: expiresAt})
	require.NoError(t, err)
	require.NoError(t, credentialRepo.CreateAttempt(contextWithTestTimeout(t), &models.OAuthConnectionAttempt{TenantID: "tenant-a", Platform: models.PlatformTiktok, AttemptID: attemptID, Status: oauthAttemptPending, Intent: "connect", SignedState: signedState, CSRFNonce: nonce, RedirectPath: "https://evil.test/callback", ExpiresAt: expiresAt, CreatedBy: "user-1"}))

	validated, status, err := handler.validateCallbackState(c, models.PlatformTiktok, signedState)
	require.Error(t, err)
	assert.Nil(t, validated)
	assert.Equal(t, "invalid_redirect", status)
}

func TestOAuthDuplicateStoreReconciliation(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	require.NoError(t, credentialRepo.CreateConnection(contextWithTestTimeout(t), &models.CredentialConnection{TenantID: "tenant-a", Platform: models.PlatformShopee, StoreIdentifier: "shop-123", Status: "connected", AccessToken: "raw-access", RefreshToken: "raw-refresh"}))

	newConnect := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformShopee, "connect", "", time.Now().Add(10*time.Minute))
	state, _, err := handler.validateCallbackState(c, models.PlatformShopee, newConnect)
	require.NoError(t, err)
	assert.EqualError(t, handler.ensureStoreWriteAllowed(c, state, "shop-123"), "duplicate_store")

	reconnect := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformShopee, "reconnect", "shop-123", time.Now().Add(10*time.Minute))
	state, _, err = handler.validateCallbackState(c, models.PlatformShopee, reconnect)
	require.NoError(t, err)
	assert.NoError(t, handler.ensureStoreWriteAllowed(c, state, "shop-123"))
}

func TestOAuthStateAndLogsRedactSecrets(t *testing.T) {
	handler, repo, credentialRepo, c := setupOAuthSecurityTest(t)
	state := createSignedOAuthAttempt(t, repo, credentialRepo, "tenant-a", models.PlatformTiktok, "connect", "", time.Now().Add(10*time.Minute))
	attempt, err := credentialRepo.GetAttempt(contextWithTestTimeout(t), "tenant-a", models.PlatformTiktok, mustParseAttemptID(t, state))
	require.NoError(t, err)
	data, err := json.Marshal(attempt)
	require.NoError(t, err)
	assert.NotContains(t, string(data), state)

	handler.logOAuthCallback(c, "tenant-a", models.PlatformTiktok, models.OAuthStatusFailed, "auth_code=raw-code access_token=raw-token state="+state)
	logs, err := repo.FindLogsByTenant(contextWithTestTimeout(t), "tenant-a", 10)
	require.NoError(t, err)
	require.Len(t, logs, 1)
	assert.Equal(t, "failed", logs[0].ErrorMsg)
	assert.False(t, strings.Contains(logs[0].ErrorMsg, "raw-code"))
	assert.False(t, strings.Contains(logs[0].ErrorMsg, "raw-token"))
}

func mustParseAttemptID(t *testing.T, state string) string {
	t.Helper()
	claims, err := oauth.ParseSignedState(state)
	require.NoError(t, err)
	return claims.AttemptID
}
