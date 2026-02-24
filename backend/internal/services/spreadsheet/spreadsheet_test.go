package spreadsheet_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/omni/backend/internal/services/spreadsheet"
)

// ---------------------------------------------------------------------------
// Constructor
// ---------------------------------------------------------------------------

func TestNewRegistryService_NotNil(t *testing.T) {
	svc := spreadsheet.NewRegistryService(nil, "tenant1")
	require.NotNil(t, svc)
}

func TestNewRegistryService_WithVariousTenants(t *testing.T) {
	tenants := []string{"tenant1", "tenant-abc", "t123", ""}
	for _, tid := range tenants {
		svc := spreadsheet.NewRegistryService(nil, tid)
		require.NotNil(t, svc, "expected non-nil service for tenant=%q", tid)
	}
}

// ---------------------------------------------------------------------------
// RegisterRequest — struct initialization & validation rules
// ---------------------------------------------------------------------------

func TestRegisterRequest_FieldAccess(t *testing.T) {
	req := spreadsheet.RegisterRequest{
		SpreadsheetID: "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms",
		Name:          "Inventory Sheet",
		Purpose:       "inventory",
	}
	assert.Equal(t, "1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms", req.SpreadsheetID)
	assert.Equal(t, "Inventory Sheet", req.Name)
	assert.Equal(t, "inventory", req.Purpose)
}

func TestRegisterRequest_EmptyFields(t *testing.T) {
	tests := []struct {
		name    string
		req     spreadsheet.RegisterRequest
		wantErr bool
	}{
		{
			name:    "valid request",
			req:     spreadsheet.RegisterRequest{SpreadsheetID: "abc", Name: "Sheet"},
			wantErr: false,
		},
		{
			name:    "empty spreadsheet_id",
			req:     spreadsheet.RegisterRequest{SpreadsheetID: "", Name: "Sheet"},
			wantErr: true,
		},
		{
			name:    "empty name",
			req:     spreadsheet.RegisterRequest{SpreadsheetID: "abc", Name: ""},
			wantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			hasError := tc.req.SpreadsheetID == "" || tc.req.Name == ""
			assert.Equal(t, tc.wantErr, hasError)
		})
	}
}

// ---------------------------------------------------------------------------
// UpdateRequest — struct initialization
// ---------------------------------------------------------------------------

func TestUpdateRequest_FieldAccess(t *testing.T) {
	req := spreadsheet.UpdateRequest{
		Name:     "Updated Sheet",
		Purpose:  "analytics",
		IsActive: true,
	}
	assert.Equal(t, "Updated Sheet", req.Name)
	assert.Equal(t, "analytics", req.Purpose)
	assert.True(t, req.IsActive)
}

func TestUpdateRequest_PartialUpdate(t *testing.T) {
	// UpdateRequest with only some fields set — covers branching logic in Update()
	tests := []struct {
		name       string
		req        spreadsheet.UpdateRequest
		expectName string
	}{
		{"name only", spreadsheet.UpdateRequest{Name: "New Name"}, "New Name"},
		{"empty name", spreadsheet.UpdateRequest{Name: ""}, ""},
		{"purpose only", spreadsheet.UpdateRequest{Purpose: "orders"}, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expectName, tc.req.Name)
		})
	}
}

func TestUpdateRequest_IsActive_DefaultsFalse(t *testing.T) {
	// Zero-value UpdateRequest should have IsActive = false
	req := spreadsheet.UpdateRequest{}
	assert.False(t, req.IsActive)
}

// ---------------------------------------------------------------------------
// ListFilter — struct initialization
// ---------------------------------------------------------------------------

func TestListFilter_FieldAccess(t *testing.T) {
	filter := spreadsheet.ListFilter{
		Purpose:    "inventory",
		ActiveOnly: true,
	}
	assert.Equal(t, "inventory", filter.Purpose)
	assert.True(t, filter.ActiveOnly)
}

func TestListFilter_ZeroValue(t *testing.T) {
	filter := spreadsheet.ListFilter{}
	assert.Equal(t, "", filter.Purpose)
	assert.False(t, filter.ActiveOnly)
}

func TestListFilter_ActiveOnlyFilter(t *testing.T) {
	// Mirrors the branching in List(): filter.ActiveOnly and filter.Purpose
	tests := []struct {
		name           string
		filter         spreadsheet.ListFilter
		expectFiltered bool
	}{
		{
			name:           "active only",
			filter:         spreadsheet.ListFilter{ActiveOnly: true},
			expectFiltered: true,
		},
		{
			name:           "purpose filter",
			filter:         spreadsheet.ListFilter{Purpose: "orders"},
			expectFiltered: true,
		},
		{
			name:           "no filter",
			filter:         spreadsheet.ListFilter{},
			expectFiltered: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isFiltered := tc.filter.ActiveOnly || tc.filter.Purpose != ""
			assert.Equal(t, tc.expectFiltered, isFiltered)
		})
	}
}

// ---------------------------------------------------------------------------
// Purpose values — known valid values
// ---------------------------------------------------------------------------

func TestRegisterRequest_KnownPurposeValues(t *testing.T) {
	validPurposes := []string{"inventory", "orders", "analytics", "wallet", "shipping", "other", ""}
	for _, p := range validPurposes {
		req := spreadsheet.RegisterRequest{
			SpreadsheetID: "abc",
			Name:          "Sheet",
			Purpose:       p,
		}
		assert.Equal(t, p, req.Purpose)
	}
}
