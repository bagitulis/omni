package response_test

import (
	"encoding/json"
	"testing"

	"github.com/omni/backend/internal/dto/response"
)

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
}

func TestSuccessHelper(t *testing.T) {
	data := map[string]string{"id": "123"}
	r := response.Success(data)
	if !r.Success {
		t.Error("Success() should return Success=true")
	}
	if r.Data == nil {
		t.Error("Success() should set Data")
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
}
