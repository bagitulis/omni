package unit_test

import (
	"testing"

	"github.com/omni/backend/internal/dto/response"
)

func TestSuccessResponse(t *testing.T) {
	data := map[string]string{"key": "value"}
	resp := response.Success(data)

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Error != "" {
		t.Error("Expected Error to be empty")
	}
	if resp.Data == nil {
		t.Error("Expected Data to be set")
	}
}

func TestErrorResponse(t *testing.T) {
	resp := response.Error("Something went wrong")

	if resp.Success {
		t.Error("Expected Success to be false")
	}
	if resp.Error != "Something went wrong" {
		t.Errorf("Expected error message, got %s", resp.Error)
	}
	if resp.Data != nil {
		t.Error("Expected Data to be nil")
	}
}

func TestSuccessWithMeta(t *testing.T) {
	data := []string{"item1", "item2"}
	meta := &response.Meta{
		Total:      100,
		Page:       1,
		PageSize:   10,
		TotalPages: 10,
	}

	resp := response.SuccessWithMeta(data, meta)

	if !resp.Success {
		t.Error("Expected Success to be true")
	}
	if resp.Meta == nil {
		t.Error("Expected Meta to be set")
	}
	if resp.Meta.Total != 100 {
		t.Errorf("Expected Total 100, got %d", resp.Meta.Total)
	}
}
