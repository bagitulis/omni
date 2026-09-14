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

// nowFunc is the clock used by the resolver. A package-level variable so tests
// can pin it without threading a clock through every call site.
var nowFunc = time.Now

// Tenant resolution for unauthenticated extension traffic.
//
// Both resolvers below search every tenant, which looks alarming given that the
// rest of the platform is strictly schema-isolated. It is necessary because the
// browser holds no JWT: a pairing code or a token is the only handle it has, so
// the owning tenant must be discovered rather than supplied.
//
// Two properties keep this safe:
//
//  1. The search returns the tenant that owns the credential, never a tenant the
//     caller asked for. There is no client input in the tenant decision.
//  2. The credentials are unguessable and single-use: a pairing code carries 40
//     bits of CSPRNG and expires in 5 minutes; a token carries 256 bits.
//
// The alternative — asking the client which tenant it belongs to — is precisely
// the cross-tenant leak this design exists to prevent.

// TenantLister returns the tenant IDs to search.
type TenantLister func() []string

// TenantDBFunc resolves a tenant-scoped database handle.
type TenantDBFunc func(tenantID string) (*gorm.DB, error)

// ErrTenantNotFoundForCredential means no tenant owns the supplied credential.
//
// Callers must surface this as a generic failure. Reporting which tenant failed,
// or distinguishing "not found" from "expired", would turn these endpoints into
// an oracle for probing valid codes and tokens.
var ErrTenantNotFoundForCredential = errors.New("credential does not belong to any tenant")

// ResolveTenantForPairingCode finds the tenant that owns a pairing code.
func ResolveTenantForPairingCode(
	ctx context.Context,
	code string,
	tenantDB TenantDBFunc,
	listTenants TenantLister,
) (string, error) {
	if code == "" {
		return "", ErrTenantNotFoundForCredential
	}
	tenants := listTenants()

	for _, tenantID := range tenants {
		db, err := tenantDB(tenantID)
		if err != nil {
			// One tenant's database being unreachable must not abort the search:
			// the code may belong to another tenant.
			continue
		}

		repo := repositories.NewExtensionRepository(db)
		found, lookupErr := repo.FindPairingCode(ctx, code)
		if lookupErr != nil || found == nil {
			continue
		}

		// A code that exists but is no longer redeemable must not resolve, or the
		// caller would proceed to a confusing failure inside ConfirmPairing.
		if !found.IsUsable(nowFunc()) {
			return "", ErrTenantNotFoundForCredential
		}
		return tenantID, nil
	}

	return "", ErrTenantNotFoundForCredential
}

// ResolveExtensionByToken finds the tenant and extension owning a pairing token.
//
// The token is hashed before lookup, so the search never compares plaintext.
func ResolveExtensionByToken(
	ctx context.Context,
	token string,
	tenantDB TenantDBFunc,
	listTenants TenantLister,
) (string, *models.Extension, error) {
	if token == "" {
		return "", nil, ErrTenantNotFoundForCredential
	}
	hash := HashPairingToken(token)

	for _, tenantID := range listTenants() {
		db, err := tenantDB(tenantID)
		if err != nil {
			continue
		}

		repo := repositories.NewExtensionRepository(db)
		ext, lookupErr := repo.FindExtensionByTokenHash(ctx, hash)
		if lookupErr != nil || ext == nil {
			continue
		}
		return tenantID, ext, nil
	}

	return "", nil, fmt.Errorf("token lookup: %w", ErrTenantNotFoundForCredential)
}
