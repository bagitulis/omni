package services

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/omni/backend/internal/models"
	"github.com/omni/backend/internal/repositories"
)
	"gorm.io/gorm"
)

type CredentialMigrationService struct {
	tenantService *TenantService
}

func NewCredentialMigrationService(basePath string) *CredentialMigrationService {
	return &CredentialMigrationService{tenantService: NewTenantService(basePath)}
}

func (s *CredentialMigrationService) Run(ctx context.Context, opts CredentialMigrationOptions) (*CredentialMigrationReport, error) {
	report := newCredentialMigrationReport(opts.Mode)
	tenants, err := s.tenantService.GetAvailableTenants(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tenants: %w", err)
	}
	for _, tenant := range tenants {
		if err := s.inspectTenant(ctx, tenant.ID, report); err != nil {
			return nil, err
		}
	}
	report.AbortReasons = uniqueStrings(report.AbortReasons)
	report.CanWrite = len(report.AbortReasons) == 0
	report.CutoverReady = report.CanWrite && countsMatch(report.SourceKeyCounts, report.TargetCredentialCounts)
	if opts.Mode == CredentialMigrationModeBackfill && opts.AllowWrite {
		if !report.CanWrite {
			return report, fmt.Errorf("credential migration abort gates failed: %s", strings.Join(report.AbortReasons, "; "))
		}
		if err := s.backfill(ctx, tenants, report, opts); err != nil {
			return report, err
		}
		report.CutoverReady = countsMatch(report.SourceKeyCounts, report.TargetCredentialCounts)
	}
	if opts.Mode == CredentialMigrationModeCutover && !report.CutoverReady {
		return report, fmt.Errorf("credential cutover preflight failed")
	}
	return report, nil
}

func newCredentialMigrationReport(mode CredentialMigrationMode) *CredentialMigrationReport {
	if mode == "" {
		mode = CredentialMigrationModeDryRun
	}
	return &CredentialMigrationReport{
		Mode:                   string(mode),
		SourceKeyCounts:        map[string]int{},
		TargetCredentialCounts: map[string]int{},
		Rollback: CredentialMigrationRollback{
			LegacyReadOnly:        true,
			PreservesMigrated:     true,
			PreservesNewlyCreated: true,
			PreservesRotated:      true,
			PreservesDisconnected: true,
			OperatorSteps: []string{
				"Keep canonical credential tables intact; do not delete migrated, newly created, rotated, or disconnected rows.",
				"Restore from canonical credential backup data; legacy plaintext credential reads are not supported.",
			},
		},
	}
}

func (s *CredentialMigrationService) inspectTenant(ctx context.Context, tenantID string, report *CredentialMigrationReport) error {
	db, err := s.tenantService.GetTenantDB(tenantID)
	if err != nil {
		report.AbortReasons = append(report.AbortReasons, "missing_tenant_schema")
		report.Tenants = append(report.Tenants, CredentialMigrationTenant{TenantID: tenantID, SchemaName: "tenant_" + tenantID, MissingTenantSchema: true, AbortReasons: []string{"missing_tenant_schema"}})
		return nil
	}
	columns, err := platformConfigColumns(ctx, db)
	if err != nil {
		return fmt.Errorf("inspect tenant %s platform_configs columns: %w", tenantID, err)
	}
	tenantReport := CredentialMigrationTenant{TenantID: tenantID, SchemaName: "tenant_" + tenantID}
	if len(columns) == 0 {
		tenantReport.MissingTenantSchema = true
		tenantReport.AbortReasons = append(tenantReport.AbortReasons, "missing_platform_configs")
		report.AbortReasons = append(report.AbortReasons, "missing_tenant_schema")
		report.Tenants = append(report.Tenants, tenantReport)
		return nil
	}
	rows, err := readPlatformConfigRows(ctx, db, columns)
	if err != nil {
		return fmt.Errorf("read tenant %s legacy rows: %w", tenantID, err)
	}
	bundles, rowAbortReasons := buildLegacyBundles(tenantID, rows)
	tenantReport.SourceRows = len(rows)
	tenantReport.IncompleteSources = countReason(rowAbortReasons, "incomplete_source")
	tenantReport.UnknownShapes = countReason(rowAbortReasons, "unknown_row_shape")
	tenantReport.UnreadableSecrets = countReason(rowAbortReasons, "unreadable_encrypted_value")
	for _, reason := range rowAbortReasons {
		tenantReport.AbortReasons = append(tenantReport.AbortReasons, reason)
		report.AbortReasons = append(report.AbortReasons, reason)
	}
	for _, bundle := range bundles {
		report.SourceKeyCounts[bundle.Platform]++
	}
	if err := s.inspectTargets(ctx, db, tenantID, bundles, report, &tenantReport); err != nil {
		return err
	}
	report.Tenants = append(report.Tenants, tenantReport)
	return nil
}

func (s *CredentialMigrationService) inspectTargets(ctx context.Context, db *gorm.DB, tenantID string, bundles []legacyCredentialBundle, report *CredentialMigrationReport, tenantReport *CredentialMigrationTenant) error {
	var conns []models.CredentialConnection
	if err := db.WithContext(ctx).Find(&conns).Error; err != nil {
		return fmt.Errorf("inspect target connections for %s: %w", tenantID, err)
	}
	var apps []models.CredentialAppConfig
	if err := db.WithContext(ctx).Find(&apps).Error; err != nil {
		return fmt.Errorf("inspect target app configs for %s: %w", tenantID, err)
	}
	tenantReport.TargetConnections = len(conns)
	tenantReport.TargetAppConfigs = len(apps)
	for _, conn := range conns {
		report.TargetCredentialCounts[conn.Platform]++
	}
	for _, app := range apps {
		if app.Configured {
			report.TargetCredentialCounts[app.Platform+"_app"]++
		}
	}
	tenantReport.DuplicateTargets = duplicateCanonicalCount(conns, apps)
	if tenantReport.DuplicateTargets > 0 {
		tenantReport.AbortReasons = append(tenantReport.AbortReasons, "duplicate_canonical_keys")
		report.AbortReasons = append(report.AbortReasons, "duplicate_canonical_keys")
	}
	tenantReport.TenantCoverageMatch = len(bundles) == 0 || tenantReport.TargetConnections+tenantReport.TargetAppConfigs > 0 || len(conns)+len(apps) == 0
	tenantReport.PlatformCountMatches = tenantReport.DuplicateTargets == 0
	return nil
}

func (s *CredentialMigrationService) backfill(ctx context.Context, tenants []TenantInfo, report *CredentialMigrationReport, opts CredentialMigrationOptions) error {
	for _, tenant := range tenants {
		db, err := s.tenantService.GetTenantDB(tenant.ID)
		if err != nil {
			return fmt.Errorf("get tenant db for backfill %s: %w", tenant.ID, err)
		}
		columns, err := platformConfigColumns(ctx, db)
		if err != nil {
			return err
		}
		rows, err := readPlatformConfigRows(ctx, db, columns)
		if err != nil {
			return err
		}
		bundles, _ := buildLegacyBundles(tenant.ID, rows)
		if err := s.backfillTenant(ctx, db, bundles, report, opts); err != nil {
			return err
		}
	}
	return nil
}

func (s *CredentialMigrationService) backfillTenant(ctx context.Context, db *gorm.DB, bundles []legacyCredentialBundle, report *CredentialMigrationReport, opts CredentialMigrationOptions) error {
	repo := repositories.NewCredentialRepository(db)
	actor := opts.Actor
	if actor == "" {
		actor = "credential_migration"
	}
	for _, bundle := range bundles {
		if bundle.StoreIdentifier != "" && bundle.AccessToken != "" {
			existing, err := repo.GetConnection(ctx, bundle.TenantID, bundle.Platform, bundle.StoreIdentifier)
			if err != nil {
				return err
			}
			if existing == nil {
				if err := repo.CreateConnection(ctx, legacyBundleToConnection(bundle, actor)); err != nil {
					return err
				}
				report.Backfill.CreatedConnections++
			} else {
				report.Backfill.SkippedExisting++
			}
		}
		if bundle.AppKey != "" || bundle.AppSecret != "" || bundle.PartnerKey != "" || bundle.PartnerID != 0 {
			existingApp, err := repo.GetAppConfig(ctx, bundle.TenantID, bundle.Platform)
			if err != nil {
				return err
			}
			if existingApp == nil {
				if err := repo.UpsertAppConfig(ctx, legacyBundleToAppConfig(bundle, actor)); err != nil {
					return err
				}
				report.Backfill.CreatedAppConfigs++
			} else {
				report.Backfill.SkippedExisting++
			}
		}
	}
	return nil
}

func legacyBundleToConnection(bundle legacyCredentialBundle, actor string) *models.CredentialConnection {
	status := "connected"
	if bundle.AccessToken == "" || bundle.RefreshToken == "" {
		status = "incomplete"
	}
	return &models.CredentialConnection{TenantID: bundle.TenantID, Platform: bundle.Platform, StoreIdentifier: bundle.StoreIdentifier, StoreName: bundle.StoreName, Region: normalizeRegion(bundle.Region), AccessToken: bundle.AccessToken, RefreshToken: bundle.RefreshToken, ShopCipher: bundle.ShopCipher, TokenExpiry: bundle.TokenExpiry, Status: status, CreatedBy: actor, UpdatedBy: actor}
}

func legacyBundleToAppConfig(bundle legacyCredentialBundle, actor string) *models.CredentialAppConfig {
	return &models.CredentialAppConfig{TenantID: bundle.TenantID, Platform: bundle.Platform, Region: normalizeRegion(bundle.Region), AppKey: bundle.AppKey, AppSecret: bundle.AppSecret, PartnerID: bundle.PartnerID, PartnerKey: bundle.PartnerKey, Configured: true, CreatedBy: actor, UpdatedBy: actor}
}

func countsMatch(left, right map[string]int) bool {
	for key, value := range left {
		if right[key] < value {
			return false
		}
	}
	return true
}

func duplicateCanonicalCount(conns []models.CredentialConnection, apps []models.CredentialAppConfig) int {
	seen := map[string]int{}
	duplicates := 0
	for _, conn := range conns {
		if conn.DisabledAt != nil {
			continue
		}
		seen["conn|"+conn.TenantID+"|"+conn.Platform+"|"+conn.StoreIdentifier]++
	}
	for _, app := range apps {
		seen["app|"+app.TenantID+"|"+app.Platform+"|"+app.StoreIdentifier]++
	}
	for _, count := range seen {
		if count > 1 {
			duplicates += count - 1
		}
	}
	return duplicates
}

func uniqueStrings(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func countReason(reasons []string, target string) int {
	count := 0
	for _, reason := range reasons {
		if reason == target {
			count++
		}
	}
	return count
}
