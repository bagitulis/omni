package services

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

func WriteCredentialMigrationReport(w io.Writer, report *CredentialMigrationReport) error {
	if report == nil {
		report = &CredentialMigrationReport{}
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("encode credential migration report: %w", err)
	}
	return nil
}

func WriteCredentialMigrationTextReport(w io.Writer, report *CredentialMigrationReport) error {
	if report == nil {
		report = &CredentialMigrationReport{}
	}
	if _, err := fmt.Fprintf(w, "Credential migration mode: %s\n", report.Mode); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Can write: %t\nCutover ready: %t\nFallback enabled: %t\n", report.CanWrite, report.CutoverReady, report.FallbackEnabled); err != nil {
		return err
	}
	if len(report.AbortReasons) > 0 {
		if _, err := fmt.Fprintf(w, "Abort reasons: %s\n", strings.Join(report.AbortReasons, "; ")); err != nil {
			return err
		}
	}
	if err := writeCountMap(w, "Source key counts", report.SourceKeyCounts); err != nil {
		return err
	}
	if err := writeCountMap(w, "Target credential counts", report.TargetCredentialCounts); err != nil {
		return err
	}
	for _, tenant := range report.Tenants {
		if _, err := fmt.Fprintf(w, "Tenant: %s schema=%s source_rows=%d target_connections=%d target_app_configs=%d duplicate_targets=%d incomplete_sources=%d unknown_shapes=%d unreadable_secrets=%d coverage_match=%t platform_match=%t\n",
			tenant.TenantID, tenant.SchemaName, tenant.SourceRows, tenant.TargetConnections, tenant.TargetAppConfigs, tenant.DuplicateTargets, tenant.IncompleteSources, tenant.UnknownShapes, tenant.UnreadableSecrets, tenant.TenantCoverageMatch, tenant.PlatformCountMatches); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintf(w, "Backfill: created_connections=%d created_app_configs=%d skipped_existing=%d\n", report.Backfill.CreatedConnections, report.Backfill.CreatedAppConfigs, report.Backfill.SkippedExisting); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "Fallback: env_var=%s enabled=%t canonical_first=%t legacy_read_only=%t audit_event=%s\n", report.Fallback.EnvVar, report.Fallback.Enabled, report.Fallback.CanonicalFirst, report.Fallback.LegacyReadOnly, report.Fallback.AuditEvent); err != nil {
		return err
	}
	return nil
}

func writeCountMap(w io.Writer, title string, counts map[string]int) error {
	if _, err := fmt.Fprintf(w, "%s:\n", title); err != nil {
		return err
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if _, err := fmt.Fprintf(w, "  %s=%d\n", key, counts[key]); err != nil {
			return err
		}
	}
	return nil
}
