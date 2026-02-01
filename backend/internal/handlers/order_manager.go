package handlers

import (
	"github.com/omni/backend/internal/utils/logger"
)

var orderManagerLogger = logger.Named("OrderManagerHandler")

// OrderManagerHandler handles order manager endpoints for frontend compatibility
// SRP: Frontend-facing endpoints for Order Manager UI
type OrderManagerHandler struct{}

// NewOrderManagerHandler creates a new order manager handler
func NewOrderManagerHandler() *OrderManagerHandler {
	return &OrderManagerHandler{}
}
