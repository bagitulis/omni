package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

// Common errors
var (
	ErrInvalidInput = errors.New("invalid input")
	ErrNotFound     = errors.New("order not found")
	ErrInternal     = errors.New("internal error")
)

// OrderService handles order operations
type OrderService struct {
	db *sql.DB
}

// NewOrderService creates a new OrderService
func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

// Order represents an order record
type Order struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	Total     float64     `json:"total"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Items     []OrderItem `json:"items"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID int64   `json:"product_id"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

// CalculateTotal calculates the total amount with precision handling
// Fixes BUG 17 (Discounts), BUG 18 (Tax), BUG 19 (Precision)
// Note: In a real system, tax and discount rules would be injected or configured.
// Here we assume a simple calculation for demonstration.
func (s *OrderService) CalculateTotal(items []OrderItem, taxRate float64, discount float64) float64 {
	var subtotal float64
	for _, item := range items {
		subtotal += item.Price * float64(item.Qty)
	}

	tax := subtotal * taxRate
	total := subtotal + tax - discount

	// Round to 2 decimal places to fix precision issues
	return math.Round(total*100) / 100
}

// CreateOrder creates a new order with transaction support
// Fixes BUG 1-7
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, items []OrderItem) (*Order, error) {
	// BUG 1: Input validation
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: no items provided", ErrInvalidInput)
	}
	for _, item := range items {
		if item.Qty <= 0 || item.Price < 0 {
			return nil, fmt.Errorf("%w: invalid item quantity or price", ErrInvalidInput)
		}
	}

	// BUG 2: Use context for timeout (caller should provide context with timeout)
	// BUG 5: Context propagation used throughout

	total := s.CalculateTotal(items, 0.0, 0.0) // Default 0 tax/discount for basic create

	// BUG 3: Transaction support
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	// Defer rollback in case of panic or error
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	// BUG 4: Parameterized queries (PostgreSQL syntax $1, $2...)
	// BUG 6: Handling insert ID correctly for Postgres (RETURNING id)
	var orderID int64
	createdAt := time.Now()

	query := `
		INSERT INTO orders (user_id, total, status, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5) 
		RETURNING id`

	err = tx.QueryRowContext(ctx, query, userID, total, "pending", createdAt, createdAt).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// BUG 7: Save items
	itemQuery := `INSERT INTO order_items (order_id, product_id, qty, price) VALUES ($1, $2, $3, $4)`
	stmt, err := tx.PrepareContext(ctx, itemQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare item statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, orderID, item.ProductID, item.Qty, item.Price); err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Order{
		ID:        orderID,
		UserID:    userID,
		Total:     total,
		Status:    "pending",
		CreatedAt: createdAt,
		UpdatedAt: createdAt,
		Items:     items,
	}, nil
}

// GetOrder retrieves an order by ID
// Fixes BUG 8-10
func (s *OrderService) GetOrder(ctx context.Context, id int64) (*Order, error) {
	// BUG 9: Context propagation
	if id <= 0 {
		return nil, ErrInvalidInput
	}

	query := `SELECT id, user_id, total, status, created_at, updated_at FROM orders WHERE id = $1`

	var order Order
	// BUG 10: Handle scan error
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&order.ID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// BUG 8: Return proper error instead of nil/panic
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	return &order, nil
}

// UpdateStatus updates the order status
// Fixes BUG 11-13
func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status string) error {
	// BUG 12: Status validation
	validStatuses := map[string]bool{
		"pending": true, "paid": true, "shipped": true, "delivered": true, "cancelled": true,
	}
	if !validStatuses[status] {
		// BUG 11: Return error value
		return fmt.Errorf("%w: invalid status '%s'", ErrInvalidInput, status)
	}

	// BUG 13: Audit trail (using updated_at)
	query := `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`

	result, err := s.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// SearchOrders searches orders by status with pagination
// Fixes BUG 14-16
func (s *OrderService) SearchOrders(ctx context.Context, status string, page, limit int) ([]Order, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// BUG 14: SQL injection prevention (parameterized)
	// BUG 15: Pagination implemented
	query := `
		SELECT id, user_id, total, status, created_at, updated_at 
		FROM orders 
		WHERE status ILIKE $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, "%"+status+"%", limit, offset)
	if err != nil {
		// BUG 16: Return error
		return nil, fmt.Errorf("failed to search orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("iteration error: %w", err)
	}

	if orders == nil {
		orders = []Order{}
	}

	return orders, nil
}
