package operations_test

import (
	"testing"

	"github.com/omni/backend/internal/services/operations"
)

// TestShopeeOperation_TableName verifies GORM table name.
func TestShopeeOperation_TableName(t *testing.T) {
	op := operations.ShopeeOperation{}
	if op.TableName() != "shopee_operations" {
		t.Errorf("expected table name 'shopee_operations', got %q", op.TableName())
	}
}

// TestOperationBatch_TableName verifies GORM table name.
func TestOperationBatch_TableName(t *testing.T) {
	b := operations.OperationBatch{}
	if b.TableName() != "operation_batches" {
		t.Errorf("expected table name 'operation_batches', got %q", b.TableName())
	}
}

// TestOperationType_Constants verifies OperationType constant values.
func TestOperationType_Constants(t *testing.T) {
	tests := []struct {
		name     string
		got      operations.OperationType
		expected operations.OperationType
	}{
		{"OpPrintLabel", operations.OpPrintLabel, "print_label"},
		{"OpShipOrder", operations.OpShipOrder, "ship_order"},
		{"OpCancelOrder", operations.OpCancelOrder, "cancel_order"},
		{"OpUpdateStock", operations.OpUpdateStock, "update_stock"},
		{"OpUpdatePrice", operations.OpUpdatePrice, "update_price"},
		{"OpBulkUpdate", operations.OpBulkUpdate, "bulk_update"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestOperationStatus_Constants verifies OperationStatus constant values.
func TestOperationStatus_Constants(t *testing.T) {
	tests := []struct {
		name     string
		got      operations.OperationStatus
		expected operations.OperationStatus
	}{
		{"StatusPending", operations.StatusPending, "pending"},
		{"StatusProcessing", operations.StatusProcessing, "processing"},
		{"StatusCompleted", operations.StatusCompleted, "completed"},
		{"StatusFailed", operations.StatusFailed, "failed"},
		{"StatusPartial", operations.StatusPartial, "partial"},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("%s = %q, want %q", tt.name, tt.got, tt.expected)
			}
		})
	}
}

// TestShopeeOperation_FieldAssignment verifies struct fields can be set without DB.
func TestShopeeOperation_FieldAssignment(t *testing.T) {
	op := operations.ShopeeOperation{
		TenantID:      "tenant1",
		OperationType: operations.OpShipOrder,
		Status:        operations.StatusPending,
		OrderSN:       "SN123",
		RetryCount:    0,
		MaxRetries:    3,
	}
	if op.TenantID != "tenant1" {
		t.Errorf("expected TenantID='tenant1', got %q", op.TenantID)
	}
	if op.OperationType != operations.OpShipOrder {
		t.Errorf("expected OperationType=OpShipOrder, got %q", op.OperationType)
	}
	if op.Status != operations.StatusPending {
		t.Errorf("expected Status=StatusPending, got %q", op.Status)
	}
	if op.OrderSN != "SN123" {
		t.Errorf("expected OrderSN='SN123', got %q", op.OrderSN)
	}
}

// TestOperationBatch_FieldAssignment verifies OperationBatch struct fields.
func TestOperationBatch_FieldAssignment(t *testing.T) {
	b := operations.OperationBatch{
		TenantID:       "tenant2",
		OperationType:  operations.OpBulkUpdate,
		Status:         operations.StatusProcessing,
		TotalItems:     10,
		ProcessedItems: 5,
		SuccessItems:   4,
		FailedItems:    1,
	}
	if b.TenantID != "tenant2" {
		t.Errorf("expected TenantID='tenant2', got %q", b.TenantID)
	}
	if b.TotalItems != 10 {
		t.Errorf("expected TotalItems=10, got %d", b.TotalItems)
	}
	if b.SuccessItems+b.FailedItems != b.ProcessedItems {
		t.Errorf("success(%d) + failed(%d) should equal processed(%d)", b.SuccessItems, b.FailedItems, b.ProcessedItems)
	}
}

// TestNewShopeeOperationsService_Constructor verifies constructor with nil DB returns non-nil.
func TestNewShopeeOperationsService_Constructor(t *testing.T) {
	svc := operations.NewShopeeOperationsService(nil)
	if svc == nil {
		t.Fatal("expected non-nil ShopeeOperationsService")
	}
}
