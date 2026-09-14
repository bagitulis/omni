package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/omni/backend/internal/models"
	"gorm.io/gorm"
)

// ErrPairingCodeUnusable is returned when a pairing code is unknown, already
// consumed, or expired. These are deliberately collapsed into one error so a
// caller cannot distinguish (and thus probe for) existing codes.
var ErrPairingCodeUnusable = errors.New("pairing code is invalid, expired, or already used")

// ExtensionRepository provides data access for paired browser extensions.
//
// Every method is tenant-scoped implicitly: the *gorm.DB handed to the
// constructor already has the tenant schema selected via search_path. There is
// no tenant_id parameter because there is no tenant_id column — passing one
// would imply the caller could cross tenants.
type ExtensionRepository struct {
	db *gorm.DB
}

// NewExtensionRepository creates an ExtensionRepository bound to one tenant DB.
func NewExtensionRepository(db *gorm.DB) *ExtensionRepository {
	return &ExtensionRepository{db: db}
}

// FindPairingCode looks up a pairing code without consuming it.
//
// Used to discover which tenant owns a code before redeeming it, since the
// browser has no JWT and the code is its only handle.
func (r *ExtensionRepository) FindPairingCode(ctx context.Context, code string) (*models.PairingCode, error) {
	if code == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var found models.PairingCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&found).Error; err != nil {
		return nil, err
	}
	return &found, nil
}

// CreateExtension inserts a newly paired extension.
func (r *ExtensionRepository) CreateExtension(ctx context.Context, ext *models.Extension) error {
	return r.db.WithContext(ctx).Create(ext).Error
}

// ListExtensions returns all paired extensions, newest first.
func (r *ExtensionRepository) ListExtensions(ctx context.Context) ([]models.Extension, error) {
	var out []models.Extension
	err := r.db.WithContext(ctx).Order("id DESC").Find(&out).Error
	return out, err
}

// GetExtensionByID looks up a paired extension by its string extension_id.
func (r *ExtensionRepository) GetExtensionByID(ctx context.Context, extensionID string) (*models.Extension, error) {
	var ext models.Extension
	if err := r.db.WithContext(ctx).Where("extension_id = ?", extensionID).First(&ext).Error; err != nil {
		return nil, err
	}
	return &ext, nil
}

// FindExtensionByTokenHash resolves an extension by the hash of its pairing
// token. The token itself is never stored, so a leaked database does not yield
// usable credentials.
func (r *ExtensionRepository) FindExtensionByTokenHash(ctx context.Context, tokenHash string) (*models.Extension, error) {
	if tokenHash == "" {
		return nil, gorm.ErrRecordNotFound
	}
	var ext models.Extension
	if err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&ext).Error; err != nil {
		return nil, err
	}
	return &ext, nil
}

// UpdateExtensionStatus sets the connection status and, when connecting,
// refreshes last_seen.
func (r *ExtensionRepository) UpdateExtensionStatus(ctx context.Context, extensionID, status string) error {
	updates := map[string]any{"status": status, "updated_at": time.Now()}
	if status == models.ExtensionStatusConnected {
		updates["last_seen"] = time.Now()
	}
	return r.db.WithContext(ctx).
		Model(&models.Extension{}).
		Where("extension_id = ?", extensionID).
		Updates(updates).Error
}

// TouchExtensionLastSeen updates last_seen without changing status.
func (r *ExtensionRepository) TouchExtensionLastSeen(ctx context.Context, extensionID string) error {
	return r.db.WithContext(ctx).
		Model(&models.Extension{}).
		Where("extension_id = ?", extensionID).
		Update("last_seen", time.Now()).Error
}

// DeleteExtension removes a paired extension by extension_id.
func (r *ExtensionRepository) DeleteExtension(ctx context.Context, extensionID string) error {
	res := r.db.WithContext(ctx).
		Where("extension_id = ?", extensionID).
		Delete(&models.Extension{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// MarkAllExtensionsDisconnected is called at startup: any extension still marked
// "connected" in the DB has no live WebSocket, so the status is stale and must
// be reset. Without this, the UI shows phantom online extensions after a restart.
func (r *ExtensionRepository) MarkAllExtensionsDisconnected(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Model(&models.Extension{}).
		Where("status = ?", models.ExtensionStatusConnected).
		Update("status", models.ExtensionStatusDisconnected).Error
}

// UpdateExtension persists changes to an existing extension row (re-pairing:
// refreshed metadata, capabilities, and token hash).
func (r *ExtensionRepository) UpdateExtension(ctx context.Context, ext *models.Extension) error {
	return r.db.WithContext(ctx).Save(ext).Error
}

// CreatePairingCode inserts a new pairing code.
func (r *ExtensionRepository) CreatePairingCode(ctx context.Context, code *models.PairingCode) error {
	return r.db.WithContext(ctx).Create(code).Error
}

// ConsumePairingCode atomically redeems a pairing code.
//
// The UPDATE is conditional on consumed=false so two concurrent confirms cannot
// both win: the second sees RowsAffected == 0 and is rejected. A read-then-write
// would leave a race where one code mints two tokens.
func (r *ExtensionRepository) ConsumePairingCode(ctx context.Context, code string, now time.Time) (*models.PairingCode, error) {
	var found models.PairingCode
	if err := r.db.WithContext(ctx).Where("code = ?", code).First(&found).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrPairingCodeUnusable
		}
		return nil, err
	}

	if !found.IsUsable(now) {
		return nil, ErrPairingCodeUnusable
	}

	res := r.db.WithContext(ctx).
		Model(&models.PairingCode{}).
		Where("id = ? AND consumed = ?", found.ID, false).
		Updates(map[string]any{"consumed": true, "consumed_at": now, "updated_at": now})
	if res.Error != nil {
		return nil, res.Error
	}
	if res.RowsAffected == 0 {
		// Lost the race — another confirm consumed it first.
		return nil, ErrPairingCodeUnusable
	}

	found.Consumed = true
	found.ConsumedAt = &now
	return &found, nil
}

// DeleteExpiredPairingCodes prunes stale codes. Called opportunistically rather
// than on a schedule to avoid a background goroutine for a low-volume table.
func (r *ExtensionRepository) DeleteExpiredPairingCodes(ctx context.Context, now time.Time) error {
	return r.db.WithContext(ctx).
		Where("expires_at < ?", now).
		Delete(&models.PairingCode{}).Error
}

// InsertScrapedProducts bulk-inserts products for a job.
//
// Deduplicates on (job_id, link): re-scraping a page after a resume is expected,
// so overlapping rows are skipped rather than duplicated or failing the job.
func (r *ExtensionRepository) InsertScrapedProducts(ctx context.Context, products []models.ScrapedProduct) error {
	if len(products) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Clauses(clauseOnConflictDoNothing()).
		CreateInBatches(products, 200).Error
}

// CountScrapedProducts returns how many rows a job has collected so far.
func (r *ExtensionRepository) CountScrapedProducts(ctx context.Context, jobID string) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).
		Model(&models.ScrapedProduct{}).
		Where("job_id = ?", jobID).
		Count(&n).Error
	return n, err
}

// ListScrapedProducts returns a page of results for a job, newest first.
func (r *ExtensionRepository) ListScrapedProducts(ctx context.Context, jobID string, page, pageSize int) ([]models.ScrapedProduct, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 500 {
		pageSize = 50
	}

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.ScrapedProduct{}).
		Where("job_id = ?", jobID).
		Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var out []models.ScrapedProduct
	err := r.db.WithContext(ctx).
		Where("job_id = ?", jobID).
		Order("id DESC").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&out).Error
	return out, total, err
}
