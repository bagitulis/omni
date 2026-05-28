package services

import "time"

// ExternalOperationStatus represents the lifecycle state of an irreversible
// external API operation (platform ship, product create, stock/price push).
//
// State transitions:
//
//	pending → succeeded  (external API confirmed success)
//	pending → failed     (external API returned an error)
//	pending → canceled   (operation aborted before external call)
//	pending → unknown    (external API unreachable / ambiguous response)
//	unknown → succeeded  (reconciliation confirmed success)
//	unknown → failed     (reconciliation confirmed failure)
//
// Retry policy:
//   - succeeded: MUST NOT retry (operation already completed externally)
//   - failed: safe to retry (external action did not execute or was rejected)
//   - canceled: safe to retry (external action never attempted)
//   - unknown: MUST reconcile before retrying (external state unclear)
type ExternalOperationStatus string

const (
	// ExtOpPending indicates the operation is in-flight or queued.
	ExtOpPending ExternalOperationStatus = "pending"

	// ExtOpSucceeded indicates the external API confirmed the action completed.
	ExtOpSucceeded ExternalOperationStatus = "succeeded"

	// ExtOpFailed indicates the external API rejected the action or returned
	// a definitive error. The external state is known to be unchanged.
	ExtOpFailed ExternalOperationStatus = "failed"

	// ExtOpCanceled indicates the operation was aborted before the external
	// API call was made (e.g., context cancellation, pre-flight validation).
	ExtOpCanceled ExternalOperationStatus = "canceled"

	// ExtOpUnknown indicates the external API call was made but the outcome
	// could not be determined (e.g., network timeout, ambiguous response).
	// MUST be reconciled before retrying.
	ExtOpUnknown ExternalOperationStatus = "unknown"
)

// CanRetry returns true if it is safe to retry the operation.
// Operations with unknown status MUST be reconciled before retry.
func (s ExternalOperationStatus) CanRetry() bool {
	return s == ExtOpFailed || s == ExtOpCanceled
}

// NeedsReconciliation returns true if the operation requires status
// verification against the external platform before any retry attempt.
func (s ExternalOperationStatus) NeedsReconciliation() bool {
	return s == ExtOpUnknown
}

// IsTerminal returns true if the operation reached a final state.
func (s ExternalOperationStatus) IsTerminal() bool {
	return s == ExtOpSucceeded || s == ExtOpFailed || s == ExtOpCanceled
}

// ExternalOperationType categorizes the kind of external operation.
type ExternalOperationType string

const (
	ExtOpBulkShip       ExternalOperationType = "bulk_ship"
	ExtOpSingleShip     ExternalOperationType = "single_ship"
	ExtOpCancelOrder    ExternalOperationType = "cancel_order"
	ExtOpCreateProduct  ExternalOperationType = "create_product"
	ExtOpUpdateStock    ExternalOperationType = "update_stock"
	ExtOpUpdatePrice    ExternalOperationType = "update_price"
	ExtOpSyncOrders     ExternalOperationType = "sync_orders"
	ExtOpSyncProducts   ExternalOperationType = "sync_products"
	ExtOpSyncBookings   ExternalOperationType = "sync_bookings"
	ExtOpArrangeShipment ExternalOperationType = "arrange_shipment"
)

// ExternalOperationRecord represents a single external operation with its
// outcome. Used for audit and retry tracking.
type ExternalOperationRecord struct {
	OperationID   string                `json:"operation_id"`
	OperationType ExternalOperationType `json:"operation_type"`
	Platform      string                `json:"platform"`
	TenantID      string                `json:"tenant_id"`
	ItemID        string                `json:"item_id"` // order_sn, product_id, sku, etc.
	Status        ExternalOperationStatus `json:"status"`
	Error         string                `json:"error,omitempty"`
	IdempotencyKey string               `json:"idempotency_key,omitempty"`
	AttemptedAt   time.Time             `json:"attempted_at"`
	CompletedAt   *time.Time            `json:"completed_at,omitempty"`
}

// ClassifyBulkShipStatus maps bulk ship result statuses to ExternalOperationStatus.
// The bulk ship handler uses: "shipped", "failed", "cancelled", "shipped_but_local_failed".
func ClassifyBulkShipStatus(resultStatus string) ExternalOperationStatus {
	switch resultStatus {
	case "shipped", "shipped_but_local_failed":
		return ExtOpSucceeded // external action succeeded regardless of local DB state
	case "failed":
		return ExtOpFailed
	case "cancelled":
		return ExtOpCanceled
	default:
		return ExtOpUnknown
	}
}

// IsStockPriceIdempotent documents that stock and price updates are
// naturally idempotent: they set an absolute value (not increment),
// so retrying with the same value produces the same external state.
// This is "last-writer-wins" semantics.
func IsStockPriceIdempotent() bool {
	return true
}

// IsOrderSyncIdempotent documents that order/product/booking sync
// operations are read-only from the platform perspective. They fetch
// data and save to local DB using upsert, so retries are always safe.
func IsOrderSyncIdempotent() bool {
	return true
}
