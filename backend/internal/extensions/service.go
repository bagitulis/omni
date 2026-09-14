package extensions

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
	"gorm.io/gorm"
)

// Service orchestrates pairing and connection lifecycle.
//
// The tenant is always resolved server-side: a caller supplies a tenant-scoped
// *gorm.DB, and every repository call is implicitly confined to that schema.
// No method accepts a tenant from the caller, which is what prevents one
// tenant's extension request from reaching another tenant's data.
type Service struct {
	// tenantDB resolves the schema-scoped DB for a tenant. Provided as a
	// function so this package does not depend on config/connection plumbing.
	tenantDB func(tenantID string) (*gorm.DB, error)

	hub *Hub
	now func() time.Time

	// resolveByToken maps a pairing token to (tenant, extension). Set by the
	// wiring layer, which owns tenant resolution. Kept as a field so this
	// package does not embed the tenant-resolution machinery.
	resolveByToken func(ctx context.Context, token string) (string, *models.Extension, error)

	// resolveTenantForCode maps a pairing code to its owning tenant, for the
	// unauthenticated confirm call.
	resolveTenantForCode func(ctx context.Context, code string) (string, error)
}

// NewService creates a Service. now defaults to time.Now and exists so pairing
// expiry can be tested without sleeping.
func NewService(tenantDB func(tenantID string) (*gorm.DB, error), hub *Hub) *Service {
	return &Service{tenantDB: tenantDB, hub: hub, now: time.Now}
}

// SetClock overrides the clock. Test-only seam.
func (s *Service) SetClock(now func() time.Time) { s.now = now }

// StartPairing creates a single-use pairing code for a tenant.
//
// The tenant comes from the authenticated caller (the JWT claim), never from
// the request body.
func (s *Service) StartPairing(ctx context.Context, tenantID, userID string) (code string, expiresAt time.Time, err error) {
	if tenantID == "" {
		return "", time.Time{}, errors.New("extensions: tenant_id is required")
	}
	db, err := s.tenantDB(tenantID)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("extensions: tenant database: %w", err)
	}
	repo := repositories.NewExtensionRepository(db)

	now := s.now()
	// Opportunistic pruning: codes are low-volume, so a background goroutine
	// would cost more than it saves.
	if err := repo.DeleteExpiredPairingCodes(ctx, now); err != nil {
		// Pruning is housekeeping; a failure must not block pairing.
		_ = err
	}

	code, err = GeneratePairingCode()
	if err != nil {
		return "", time.Time{}, err
	}

	expiresAt = PairingCodeExpiry(now)
	var uid *int64
	if userID != "" {
		// userID arrives as a string from JWT claims; store the numeric form
		// when it parses, else leave it unset rather than guessing a value.
		if parsed, parseErr := parseUserID(userID); parseErr == nil {
			uid = &parsed
		}
	}

	pair := &models.PairingCode{
		Code:      code,
		UserID:    uid,
		ExpiresAt: expiresAt,
	}
	if err := repo.CreatePairingCode(ctx, pair); err != nil {
		return "", time.Time{}, fmt.Errorf("extensions: create pairing code: %w", err)
	}

	return code, expiresAt, nil
}

// ConfirmPairing redeems a code and returns the token to hand to the extension.
//
// The tenant comes from the caller's tenant-scoped DB. To find which tenant owns
// a code, callers resolve it with ResolvePairingTenant first (the extension has
// no JWT, so the code is the only handle it has).
func (s *Service) ConfirmPairing(ctx context.Context, tenantID string, req PairRequest) (token string, ext *models.Extension, err error) {
	if tenantID == "" {
		return "", nil, errors.New("extensions: tenant_id is required")
	}
	if err := ValidateExtensionID(req.ExtensionID); err != nil {
		return "", nil, err
	}

	db, err := s.tenantDB(tenantID)
	if err != nil {
		return "", nil, fmt.Errorf("extensions: tenant database: %w", err)
	}
	repo := repositories.NewExtensionRepository(db)

	now := s.now()
	consumed, err := repo.ConsumePairingCode(ctx, req.Code, now)
	if err != nil {
		return "", nil, err // already the shared unusable sentinel
	}

	token, hash, err := GeneratePairingToken()
	if err != nil {
		return "", nil, err
	}

	// A user may pair several browsers, so an existing extension_id is updated
	// in place rather than rejected: re-pairing a reloaded extension is normal.
	existing, lookupErr := repo.GetExtensionByID(ctx, req.ExtensionID)
	if lookupErr == nil && existing != nil {
		existing.Hostname = req.Hostname
		existing.BrowserInfo = req.BrowserInfo
		existing.ChromeVersion = req.ChromeVersion
		existing.ExtensionVersion = req.ExtensionVersion
		existing.ProtocolVersion = req.ProtocolVersion
		existing.Capabilities = NormalizeCapabilities(req.Capabilities)
		existing.TokenHash = hash
		existing.UserID = consumed.UserID
		if saveErr := repo.UpdateExtension(ctx, existing); saveErr != nil {
			return "", nil, fmt.Errorf("extensions: re-pair extension: %w", saveErr)
		}
		return token, existing, nil
	}

	ext = &models.Extension{
		ExtensionID:      req.ExtensionID,
		UserID:           consumed.UserID,
		Hostname:         req.Hostname,
		BrowserInfo:      req.BrowserInfo,
		ChromeVersion:    req.ChromeVersion,
		ExtensionVersion: req.ExtensionVersion,
		ProtocolVersion:  req.ProtocolVersion,
		Capabilities:     NormalizeCapabilities(req.Capabilities),
		Status:           models.ExtensionStatusDisconnected,
		TokenHash:        hash,
		PairedAt:         now,
	}
	if err := repo.CreateExtension(ctx, ext); err != nil {
		return "", nil, fmt.Errorf("extensions: create extension: %w", err)
	}

	return token, ext, nil
}

// Authenticate resolves a connect request to a tenant-bound identity.
//
// The lookup is by token hash across the tenants the caller is allowed to see.
// In practice a deployment resolves the tenant from the token first (via the
// tenant resolver), then looks up the extension inside that schema — the token
// is never used to select a tenant from client input.
func (s *Service) Authenticate(ctx context.Context, req ConnectRequest) (*ConnectionIdentity, error) {
	if req.Token == "" {
		return nil, errors.New("extensions: token is required")
	}
	if s.resolveByToken == nil {
		return nil, errors.New("extensions: token resolver not configured")
	}

	tenantID, ext, err := s.resolveByToken(ctx, req.Token)
	if err != nil {
		return nil, err
	}
	if tenantID == "" {
		// Fail closed: there is no default tenant.
		return nil, errors.New("extensions: token resolved to an empty tenant")
	}

	db, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("extensions: tenant database: %w", err)
	}

	var userID string
	if ext.UserID != nil {
		userID = fmt.Sprintf("%d", *ext.UserID)
	}

	return &ConnectionIdentity{
		ExtensionID: ext.ExtensionID,
		TenantID:    tenantID,
		UserID:      userID,
		DB:          db,
	}, nil
}

// SetTokenResolver configures how a pairing token maps to (tenant, extension).
func (s *Service) SetTokenResolver(fn func(ctx context.Context, token string) (string, *models.Extension, error)) {
	s.resolveByToken = fn
}

// SetPairingCodeResolver configures how a pairing code maps to its owning
// tenant.
//
// This lookup is unavoidable: the confirm call arrives from a browser that has
// no JWT, so the code is the only handle it holds and the tenant must be
// discovered from it. The resolver therefore MUST search every tenant's codes
// and return exactly one tenant — the code's uniqueness is what makes that
// unambiguous, and it is why codes are generated from 40 bits of CSPRNG.
func (s *Service) SetPairingCodeResolver(fn func(ctx context.Context, code string) (string, error)) {
	s.resolveTenantForCode = fn
}

// ResolvePairingTenant returns the tenant that owns a pairing code, or an error
// when the code is unknown. Callers use it before ConfirmPairing.
func (s *Service) ResolvePairingTenant(ctx context.Context, code string) (string, error) {
	if code == "" {
		return "", errors.New("extensions: pairing code is required")
	}
	if s.resolveTenantForCode == nil {
		return "", errors.New("extensions: pairing code resolver not configured")
	}
	return s.resolveTenantForCode(ctx, code)
}

// MarkConnected records that an extension's socket is live.
func (s *Service) MarkConnected(ctx context.Context, tenantID, extensionID string) error {
	return s.setStatus(ctx, tenantID, extensionID, models.ExtensionStatusConnected)
}

// MarkDisconnected records that an extension's socket closed.
func (s *Service) MarkDisconnected(ctx context.Context, tenantID, extensionID string) error {
	return s.setStatus(ctx, tenantID, extensionID, models.ExtensionStatusDisconnected)
}

func (s *Service) setStatus(ctx context.Context, tenantID, extensionID, status string) error {
	if tenantID == "" {
		return errors.New("extensions: tenant_id is required")
	}
	db, err := s.tenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("extensions: tenant database: %w", err)
	}
	return repositories.NewExtensionRepository(db).
		UpdateExtensionStatus(ctx, extensionID, status)
}

// ListExtensions returns the tenant's paired extensions.
func (s *Service) ListExtensions(ctx context.Context, tenantID string) ([]models.Extension, error) {
	if tenantID == "" {
		return nil, errors.New("extensions: tenant_id is required")
	}
	db, err := s.tenantDB(tenantID)
	if err != nil {
		return nil, fmt.Errorf("extensions: tenant database: %w", err)
	}
	return repositories.NewExtensionRepository(db).ListExtensions(ctx)
}

// Unpair removes an extension and force-disconnects any live socket, so the
// browser cannot keep acting after the operator revokes it.
func (s *Service) Unpair(ctx context.Context, tenantID, extensionID string) error {
	if tenantID == "" {
		return errors.New("extensions: tenant_id is required")
	}
	db, err := s.tenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("extensions: tenant database: %w", err)
	}
	if err := repositories.NewExtensionRepository(db).DeleteExtension(ctx, extensionID); err != nil {
		return err
	}
	if s.hub != nil {
		s.hub.Disconnect(extensionID)
	}
	return nil
}

// ResetStaleConnections clears "connected" rows at startup: after a restart no
// socket exists, so a persisted connected status is stale and would show phantom
// online extensions.
func (s *Service) ResetStaleConnections(ctx context.Context, tenantID string) error {
	if tenantID == "" {
		return errors.New("extensions: tenant_id is required")
	}
	db, err := s.tenantDB(tenantID)
	if err != nil {
		return fmt.Errorf("extensions: tenant database: %w", err)
	}
	return repositories.NewExtensionRepository(db).MarkAllExtensionsDisconnected(ctx)
}
