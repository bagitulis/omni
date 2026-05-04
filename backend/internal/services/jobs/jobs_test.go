package jobs_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/omni/backend/internal/services/jobs"
)


// ---------------------------------------------------------------------------
// Executor
// ---------------------------------------------------------------------------

// TestNewExecutor_Constructor verifies executor creation returns non-nil.
func TestNewExecutor_Constructor(t *testing.T) {
	e := jobs.NewExecutor(nil, 3)
	if e == nil {
		t.Fatal("expected non-nil Executor")
	}
}

// TestExecutor_RegisterHandler_NoNoPanic verifies handler registration does not panic.
func TestExecutor_RegisterHandler_NoNoPanic(t *testing.T) {
	e := jobs.NewExecutor(nil, 1)
	e.RegisterHandler("some_job", func(_ context.Context, _ string) (string, error) {
		return "ok", nil
	})
	// No assertion needed beyond no panic
}

// TestExecutor_SetPollInterval_NoPanic verifies pure setter does not panic.
func TestExecutor_SetPollInterval_NoPanic(t *testing.T) {
	e := jobs.NewExecutor(nil, 1)
	e.SetPollInterval(10 * time.Second)
	e.SetPollInterval(500 * time.Millisecond)
}

// TestExecutor_SetTimeout_NoPanic verifies pure setter does not panic.
func TestExecutor_SetTimeout_NoPanic(t *testing.T) {
	e := jobs.NewExecutor(nil, 1)
	e.SetTimeout(30 * time.Second)
	e.SetTimeout(1 * time.Minute)
}

// ---------------------------------------------------------------------------
// SerializePayload
// ---------------------------------------------------------------------------

// TestSerializePayload_Struct verifies struct serializes to valid JSON.
func TestSerializePayload_Struct(t *testing.T) {
	type sample struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}
	s, err := jobs.SerializePayload(sample{Name: "test", Value: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		t.Fatalf("result is not valid JSON: %v", err)
	}
	if out["name"] != "test" {
		t.Errorf("expected name='test', got %v", out["name"])
	}
	if out["value"] != float64(42) {
		t.Errorf("expected value=42, got %v", out["value"])
	}
}

// TestSerializePayload_Map verifies map serializes correctly.
func TestSerializePayload_Map(t *testing.T) {
	m := map[string]interface{}{"key": "value", "count": 1}
	s, err := jobs.SerializePayload(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(s, `"key"`) {
		t.Errorf("expected serialized JSON to contain 'key', got: %s", s)
	}
}

// ---------------------------------------------------------------------------
// DTO structs
// ---------------------------------------------------------------------------

// TestHistoryItemResponse_FieldAssignment verifies struct can be populated.
func TestHistoryItemResponse_FieldAssignment(t *testing.T) {
	now := time.Now()
	r := jobs.HistoryItemResponse{
		ID:           1,
		JobID:        "job-001",
		JobType:      "escrow_sync",
		Status:       "completed",
		ErrorMessage: "",
		DurationMs:   150,
		StartedAt:    &now,
		CompletedAt:  &now,
		CreatedAt:    now,
	}
	if r.JobID != "job-001" {
		t.Errorf("expected JobID='job-001', got %q", r.JobID)
	}
	if r.DurationMs != 150 {
		t.Errorf("expected DurationMs=150, got %d", r.DurationMs)
	}
}

// TestPaginatedFilter_FieldAssignment verifies struct can be populated.
func TestPaginatedFilter_FieldAssignment(t *testing.T) {
	f := jobs.PaginatedFilter{
		Page:      2,
		PageSize:  25,
		Status:    "completed",
		JobType:   "escrow_sync",
		SortBy:    "created_at",
		SortOrder: "desc",
	}
	if f.Page != 2 {
		t.Errorf("expected Page=2, got %d", f.Page)
	}
	if f.SortOrder != "desc" {
		t.Errorf("expected SortOrder='desc', got %q", f.SortOrder)
	}
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

// TestConstants verifies GracePeriod and JobTypeImageCleanup have expected values.
func TestConstants(t *testing.T) {
	if jobs.GracePeriod != 24*time.Hour {
		t.Errorf("expected GracePeriod=24h, got %v", jobs.GracePeriod)
	}
	if jobs.JobTypeImageCleanup != "image_cleanup" {
		t.Errorf("expected JobTypeImageCleanup='image_cleanup', got %q", jobs.JobTypeImageCleanup)
	}
}

// ---------------------------------------------------------------------------
// ImageCleanupHandler
// ---------------------------------------------------------------------------

// TestNewImageCleanupHandler_Constructor verifies constructor with nil DB returns non-nil.
func TestNewImageCleanupHandler_Constructor(t *testing.T) {
	h := jobs.NewImageCleanupHandler(nil)
	if h == nil {
		t.Fatal("expected non-nil ImageCleanupHandler")
	}
}
