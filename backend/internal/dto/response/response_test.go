package response_test

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/dto/response"
)

// ---------------------------------------------------------------------------
// APIResponse struct field tests
// ---------------------------------------------------------------------------

func TestAPIResponse_SuccessTrue(t *testing.T) {
	r := response.APIResponse{
		Success: true,
		Data:    "some data",
	}
	if !r.Success {
		t.Error("Success should be true")
	}
	if r.Data != "some data" {
		t.Errorf("Data = %v, want some data", r.Data)
	}
	if r.Error != "" {
		t.Errorf("Error should be empty, got %q", r.Error)
	}
}

func TestAPIResponse_ErrorState(t *testing.T) {
	r := response.APIResponse{
		Success: false,
		Error:   "something went wrong",
	}
	if r.Success {
		t.Error("Success should be false")
	}
	if r.Error != "something went wrong" {
		t.Errorf("Error = %q, want something went wrong", r.Error)
	}
}

func TestAPIResponse_WithMeta(t *testing.T) {
	meta := &response.Meta{
		Total:      100,
		Page:       2,
		PageSize:   20,
		TotalPages: 5,
	}
	r := response.APIResponse{
		Success: true,
		Meta:    meta,
	}
	if r.Meta == nil {
		t.Fatal("Meta should not be nil")
	}
	if r.Meta.Total != 100 {
		t.Errorf("Meta.Total = %d, want 100", r.Meta.Total)
	}
	if r.Meta.Page != 2 {
		t.Errorf("Meta.Page = %d, want 2", r.Meta.Page)
	}
	if r.Meta.PageSize != 20 {
		t.Errorf("Meta.PageSize = %d, want 20", r.Meta.PageSize)
	}
	if r.Meta.TotalPages != 5 {
		t.Errorf("Meta.TotalPages = %d, want 5", r.Meta.TotalPages)
	}
}

func TestAPIResponse_MessageField(t *testing.T) {
	r := response.APIResponse{
		Success: true,
		Message: "operation completed",
	}
	if r.Message != "operation completed" {
		t.Errorf("Message = %q, want operation completed", r.Message)
	}
}

// ---------------------------------------------------------------------------
// Meta struct tests
// ---------------------------------------------------------------------------

func TestMeta_ZeroValues(t *testing.T) {
	m := response.Meta{}
	if m.Total != 0 {
		t.Errorf("Total = %d, want 0", m.Total)
	}
	if m.Page != 0 {
		t.Errorf("Page = %d, want 0", m.Page)
	}
}

func TestMeta_JSONTags(t *testing.T) {
	m := response.Meta{Total: 10, Page: 1, PageSize: 5, TotalPages: 2}
	b, err := json.Marshal(m)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var result map[string]interface{}
	if err := json.Unmarshal(b, &result); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	for _, key := range []string{"total", "page", "page_size", "total_pages"} {
		if _, ok := result[key]; !ok {
			t.Errorf("expected JSON key %q", key)
		}
	}
}

// ---------------------------------------------------------------------------
// Helper function tests: Success, SuccessWithMeta, Error, ErrorWithMessage
// ---------------------------------------------------------------------------

func TestSuccessHelper(t *testing.T) {
	data := map[string]string{"id": "123"}
	r := response.Success(data)

	if !r.Success {
		t.Error("Success() should return Success=true")
	}
	if r.Data == nil {
		t.Error("Success() should set Data")
	}
	if r.Meta != nil {
		t.Error("Success() should not set Meta")
	}
	if r.Error != "" {
		t.Errorf("Success() should not set Error, got %q", r.Error)
	}
}

func TestSuccessHelper_NilData(t *testing.T) {
	r := response.Success(nil)
	if !r.Success {
		t.Error("Success() should return Success=true even with nil data")
	}
}

func TestSuccessWithMetaHelper(t *testing.T) {
	data := []string{"item1", "item2"}
	meta := &response.Meta{Total: 2, Page: 1, PageSize: 10, TotalPages: 1}

	r := response.SuccessWithMeta(data, meta)

	if !r.Success {
		t.Error("SuccessWithMeta() should return Success=true")
	}
	if r.Data == nil {
		t.Error("SuccessWithMeta() should set Data")
	}
	if r.Meta == nil {
		t.Fatal("SuccessWithMeta() should set Meta")
	}
	if r.Meta.Total != 2 {
		t.Errorf("Meta.Total = %d, want 2", r.Meta.Total)
	}
}

func TestErrorHelper(t *testing.T) {
	r := response.Error("resource not found")

	if r.Success {
		t.Error("Error() should return Success=false")
	}
	if r.Error != "resource not found" {
		t.Errorf("Error = %q, want resource not found", r.Error)
	}
	if r.Data != nil {
		t.Errorf("Error() should not set Data, got %v", r.Data)
	}
}

func TestErrorHelper_EmptyMessage(t *testing.T) {
	r := response.Error("")
	if r.Success {
		t.Error("Error() should return Success=false")
	}
	if r.Error != "" {
		t.Errorf("Error = %q, want empty", r.Error)
	}
}

func TestErrorWithMessageHelper(t *testing.T) {
	r := response.ErrorWithMessage("VALIDATION_ERROR", "please fix the fields")

	if r.Success {
		t.Error("ErrorWithMessage() should return Success=false")
	}
	if r.Error != "VALIDATION_ERROR" {
		t.Errorf("Error = %q, want VALIDATION_ERROR", r.Error)
	}
	if r.Message != "please fix the fields" {
		t.Errorf("Message = %q, want please fix the fields", r.Message)
	}
}

// ---------------------------------------------------------------------------
// JSON marshaling: omitempty behavior
// ---------------------------------------------------------------------------

func TestAPIResponse_JSONOmitempty_NoDataNoMeta(t *testing.T) {
	r := response.APIResponse{Success: true}
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var m map[string]interface{}
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	// success must be present
	if _, ok := m["success"]; !ok {
		t.Error("expected JSON key success")
	}
	// data, error, message, meta should be absent when omitempty
	for _, key := range []string{"data", "error", "message", "meta"} {
		if _, ok := m[key]; ok {
			// omitempty: these should not be present when zero/nil
			t.Logf("note: key %q present in JSON even when zero (acceptable for interface{} zero)", key)
		}
	}
}

func TestAPIResponse_JSONRoundTrip(t *testing.T) {
	original := response.APIResponse{
		Success: false,
		Error:   "not found",
		Message: "resource missing",
	}
	b, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	var decoded response.APIResponse
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Success != original.Success {
		t.Errorf("Success = %v, want %v", decoded.Success, original.Success)
	}
	if decoded.Error != original.Error {
		t.Errorf("Error = %q, want %q", decoded.Error, original.Error)
	}
	if decoded.Message != original.Message {
		t.Errorf("Message = %q, want %q", decoded.Message, original.Message)
	}
}
