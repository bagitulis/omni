package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

// Errors for order operations
var (
	ErrInvalidUserID  = errors.New("invalid user ID: must be positive")
	ErrEmptyItems     = errors.New("order must contain at least one item")
	ErrInvalidItem    = errors.New("invalid item: quantity and price must be positive")
	ErrOrderNotFound  = errors.New("order not found")
	ErrInvalidStatus  = errors.New("invalid status")
	ErrInvalidOrderID = errors.New("invalid order ID: must be positive")
	ErrInvalidLimit   = errors.New("invalid limit: must be between 1 and 100")
	ErrInvalidOffset  = errors.New("invalid offset: must be non-negative")
)

var validStatuses = map[string]bool{
	"pending": true, "confirmed": true, "shipped": true, "delivered": true, "cancelled": true,
}

const (
	DefaultTimeout = 30 * time.Second
	TaxRate        = 0.10 // 10% tax
	MaxLimit       = 100
)

// OrderService handles order operations with proper error handling and context support
type OrderService struct {
	db *sql.DB
}

// NewOrderService creates a new OrderService instance
func NewOrderService(db *sql.DB) *OrderService { return &OrderService{db: db} }

// Order represents an order in the system
type Order struct {
	ID        int
	UserID    int
	Total     float64
	Status    string
	CreatedAt time.Time
	Items     []OrderItem
}

// OrderItem represents a single item in an order
type OrderItem struct {
	ProductID int
	Qty       int
	Price     float64
	Discount  float64 // Discount percentage (0-100)
}

// CreateOrder creates a new order with proper validation, transaction, and context
func (s *OrderService) CreateOrder(ctx context.Context, userID int, items []OrderItem) (*Order, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("CreateOrder: %w", ErrInvalidUserID)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("CreateOrder: %w", ErrEmptyItems)
	}
	for i, item := range items {
		if item.ProductID <= 0 || item.Qty <= 0 || item.Price < 0 {
			return nil, fmt.Errorf("CreateOrder: item %d: %w", i, ErrInvalidItem)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	total := s.CalculateTotal(items)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	result, err := tx.ExecContext(ctx,
		"INSERT INTO orders (user_id, total, status, created_at) VALUES ($1, $2, $3, $4)",
		userID, total, "pending", time.Now())
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: failed to insert order: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("CreateOrder: failed to get order ID: %w", err)
	}

	for _, item := range items {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO order_items (order_id, product_id, quantity, price, discount) VALUES ($1, $2, $3, $4, $5)",
			id, item.ProductID, item.Qty, item.Price, item.Discount)
		if err != nil {
			return nil, fmt.Errorf("CreateOrder: failed to insert order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("CreateOrder: failed to commit transaction: %w", err)
	}

	return &Order{
		ID: int(id), UserID: userID, Total: total, Status: "pending", CreatedAt: time.Now(), Items: items,
	}, nil
}

// GetOrder retrieves an order by ID with proper error handling
func (s *OrderService) GetOrder(ctx context.Context, id int) (*Order, error) {
	if id <= 0 {
		return nil, fmt.Errorf("GetOrder: %w", ErrInvalidOrderID)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	row := s.db.QueryRowContext(ctx,
		"SELECT id, user_id, total, status, created_at FROM orders WHERE id = $1", id)

	var order Order
	err := row.Scan(&order.ID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("GetOrder: %w", ErrOrderNotFound)
		}
		return nil, fmt.Errorf("GetOrder: failed to scan order: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT product_id, quantity, price, discount FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("GetOrder: failed to query order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.Qty, &item.Price, &item.Discount); err != nil {
			return nil, fmt.Errorf("GetOrder: failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("GetOrder: error iterating order items: %w", err)
	}

	return &order, nil
}

// UpdateStatus updates order status with validation and audit trail
func (s *OrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	if id <= 0 {
		return fmt.Errorf("UpdateStatus: %w", ErrInvalidOrderID)
	}
	if !validStatuses[status] {
		return fmt.Errorf("UpdateStatus: %w", ErrInvalidStatus)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateStatus: failed to begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var oldStatus string
	err = tx.QueryRowContext(ctx, "SELECT status FROM orders WHERE id = $1", id).Scan(&oldStatus)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("UpdateStatus: %w", ErrOrderNotFound)
		}
		return fmt.Errorf("UpdateStatus: failed to get current status: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3", status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("UpdateStatus: failed to update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("UpdateStatus: failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("UpdateStatus: %w", ErrOrderNotFound)
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO order_audit_log (order_id, old_status, new_status, changed_at) VALUES ($1, $2, $3, $4)",
		id, oldStatus, status, time.Now())
	if err != nil {
		return fmt.Errorf("UpdateStatus: failed to create audit trail: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("UpdateStatus: failed to commit transaction: %w", err)
	}
	return nil
}

// SearchOrders searches orders with proper parameterization and pagination
func (s *OrderService) SearchOrders(ctx context.Context, keyword string, limit, offset int) ([]Order, error) {
	if limit <= 0 || limit > MaxLimit {
		return nil, fmt.Errorf("SearchOrders: %w", ErrInvalidLimit)
	}
	if offset < 0 {
		return nil, fmt.Errorf("SearchOrders: %w", ErrInvalidOffset)
	}

	ctx, cancel := context.WithTimeout(ctx, DefaultTimeout)
	defer cancel()

	query := `SELECT id, user_id, total, status, created_at FROM orders 
              WHERE status LIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.db.QueryContext(ctx, query, "%"+keyword+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("SearchOrders: failed to query orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("SearchOrders: failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("SearchOrders: error iterating orders: %w", err)
	}
	return orders, nil
}

// CalculateTotal calculates order total with discounts and tax
func (s *OrderService) CalculateTotal(items []OrderItem) float64 {
	var subtotal float64
	for _, item := range items {
		discountMultiplier := 1.0 - (item.Discount / 100.0)
		subtotal += item.Price * float64(item.Qty) * discountMultiplier
	}
	return math.Round(subtotal*(1+TaxRate)*100) / 100
}
