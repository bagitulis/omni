package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

// OrderService handles order operations
type OrderService struct {
	db      *sql.DB
	taxRate float64
}

// NewOrderService creates a new OrderService with proper initialization
func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db, taxRate: 0.10}
}

// Order represents a customer order
type Order struct {
	ID        int
	UserID    int
	Total     float64
	Status    string
	CreatedAt time.Time
	Items     []OrderItem
}

// OrderItem represents an item in an order
type OrderItem struct {
	ProductID int
	Qty       int
	Price     float64
	Discount  float64
}

var validStatuses = map[string]bool{
	"pending": true, "confirmed": true, "processing": true,
	"shipped": true, "delivered": true, "cancelled": true,
}

var (
	ErrInvalidUserID  = errors.New("invalid user ID: must be positive")
	ErrEmptyItems     = errors.New("order must contain at least one item")
	ErrInvalidItem    = errors.New("invalid item: quantity and price must be positive")
	ErrInvalidStatus  = errors.New("invalid order status")
	ErrOrderNotFound  = errors.New("order not found")
	ErrInvalidOrderID = errors.New("invalid order ID: must be positive")
	ErrInvalidLimit   = errors.New("invalid limit: must be between 1 and 100")
)

func validateItems(items []OrderItem) error {
	if len(items) == 0 {
		return ErrEmptyItems
	}
	for _, item := range items {
		if item.Qty <= 0 || item.Price < 0 || item.ProductID <= 0 {
			return ErrInvalidItem
		}
		if item.Discount < 0 || item.Discount > 1 {
			return fmt.Errorf("invalid discount: must be between 0 and 1")
		}
	}
	return nil
}

// CreateOrder creates a new order with validation, transaction, and context
func (s *OrderService) CreateOrder(ctx context.Context, userID int, items []OrderItem) (*Order, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	if err := validateItems(items); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	total := s.CalculateTotal(items)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
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
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	orderID, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get order ID: %w", err)
	}

	for _, item := range items {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO order_items (order_id, product_id, quantity, price, discount) VALUES ($1, $2, $3, $4, $5)",
			orderID, item.ProductID, item.Qty, item.Price, item.Discount)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Order{
		ID: int(orderID), UserID: userID, Total: total,
		Status: "pending", CreatedAt: time.Now(), Items: items,
	}, nil
}

// GetOrder retrieves an order by ID with proper error handling
func (s *OrderService) GetOrder(ctx context.Context, id int) (*Order, error) {
	if id <= 0 {
		return nil, ErrInvalidOrderID
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	row := s.db.QueryRowContext(ctx,
		"SELECT id, user_id, total, status, created_at FROM orders WHERE id = $1", id)

	var order Order
	err := row.Scan(&order.ID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	rows, err := s.db.QueryContext(ctx,
		"SELECT product_id, quantity, price, discount FROM order_items WHERE order_id = $1", id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.Qty, &item.Price, &item.Discount); err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return &order, nil
}

// UpdateStatus updates order status with validation and audit trail
func (s *OrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	if id <= 0 {
		return ErrInvalidOrderID
	}
	if !validStatuses[status] {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
	}

	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
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
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get current status: %w", err)
	}

	result, err := tx.ExecContext(ctx,
		"UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3", status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	if rowsAffected, _ := result.RowsAffected(); rowsAffected == 0 {
		return ErrOrderNotFound
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO order_status_history (order_id, old_status, new_status, changed_at) VALUES ($1, $2, $3, $4)",
		id, oldStatus, status, time.Now())
	if err != nil {
		return fmt.Errorf("failed to create audit record: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

// SearchOrders searches orders with parameterization and pagination
func (s *OrderService) SearchOrders(ctx context.Context, keyword string, limit, offset int) ([]Order, error) {
	if limit <= 0 || limit > 100 {
		return nil, ErrInvalidLimit
	}
	if offset < 0 {
		offset = 0
	}

	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	rows, err := s.db.QueryContext(ctx,
		"SELECT id, user_id, total, status, created_at FROM orders WHERE status LIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3",
		"%"+keyword+"%", limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to search orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}
	return orders, nil
}

// CalculateTotal calculates order total with discounts, tax, and proper precision
func (s *OrderService) CalculateTotal(items []OrderItem) float64 {
	var subtotal float64
	for _, item := range items {
		itemTotal := item.Price * float64(item.Qty) * (1 - item.Discount)
		subtotal += itemTotal
	}
	total := subtotal * (1 + s.taxRate)
	return math.Round(total*100) / 100
}
