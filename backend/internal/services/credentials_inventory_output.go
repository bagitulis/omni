package services

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
)

// WriteInventoryReport writes a redacted dry-run report as deterministic JSON.
func WriteInventoryReport(w io.Writer, report *InventoryReport) error {
	if report == nil {
		report = &InventoryReport{Rows: []InventoryRow{}}
	}
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(report); err != nil {
		return fmt.Errorf("encode inventory report: %w", err)
	}
	return nil
}

// WriteInventoryTextReport writes a human-readable redacted dry-run report.
func WriteInventoryTextReport(w io.Writer, report *InventoryReport) error {
	if report == nil {
		report = &InventoryReport{Rows: []InventoryRow{}}
	}
	byTenant := make(map[string][]InventoryRow)
	for _, row := range report.Rows {
		byTenant[row.TenantID] = append(byTenant[row.TenantID], row)
	}
	tenants := make([]string, 0, len(byTenant))
	for tenantID := range byTenant {
		tenants = append(tenants, tenantID)
	}
	sort.Strings(tenants)

	for _, tenantID := range tenants {
		if _, err := fmt.Fprintf(w, "Tenant: %s\n", tenantID); err != nil {
			return err
		}
		platforms := rowsByPlatform(byTenant[tenantID])
		platformNames := make([]string, 0, len(platforms))
		for platform := range platforms {
			platformNames = append(platformNames, platform)
		}
		sort.Strings(platformNames)
		for _, platform := range platformNames {
			rows := platforms[platform]
			if _, err := fmt.Fprintf(w, "  Platform: %s\n", platform); err != nil {
				return err
			}
			mixedMarker := ""
			if len(rows) > 1 {
				mixedMarker = " ⚠️ MIXED"
			}
			for _, row := range rows {
				missing := "none"
				if len(row.MissingRequiredFields) > 0 {
					missing = strings.Join(row.MissingRequiredFields, ", ")
				}
				if _, err := fmt.Fprintf(w, "    Row shape: %s (count: %d)%s\n", row.RowShape, row.Count, mixedMarker); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "    Missing fields: %s\n", missing); err != nil {
					return err
				}
				if _, err := fmt.Fprintf(w, "    Duplicate store IDs: %d\n", row.DuplicateStoreCount); err != nil {
					return err
				}
				if row.RedactedValueSummary != "" {
					if _, err := fmt.Fprintf(w, "    Secret values: %s\n", row.RedactedValueSummary); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func rowsByPlatform(rows []InventoryRow) map[string][]InventoryRow {
	byPlatform := make(map[string][]InventoryRow)
	for _, row := range rows {
		byPlatform[row.Platform] = append(byPlatform[row.Platform], row)
	}
	for platform := range byPlatform {
		sort.Slice(byPlatform[platform], func(i, j int) bool {
			return byPlatform[platform][i].RowShape < byPlatform[platform][j].RowShape
		})
	}
	return byPlatform
}
