package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"
)

// OrderService handles order operations with proper error handling and context propagation
type OrderService struct {
	db         *sql.DB
	taxRate    float64
	timeoutSec int
}

// Order represents an order entity
type Order struct {
	ID        int
	UserID    int
	Total     float64
	Status    string
	CreatedAt time.Time
	Items     []OrderItem
}

// OrderItem represents an item within an order
type OrderItem struct {
	ProductID int
	Qty       int
	Price     float64
	Discount  float64
}

// SearchParams holds pagination and search parameters
type SearchParams struct {
	Keyword string
	Limit   int
	Offset  int
}

// Valid order statuses
var validStatuses = map[string]bool{
	"pending": true, "confirmed": true, "processing": true,
	"shipped": true, "delivered": true, "cancelled": true,
}

// Custom errors
var (
	ErrInvalidUserID = errors.New("invalid user ID: must be positive")
	ErrEmptyItems    = errors.New("order must contain at least one item")
	ErrInvalidItem   = errors.New("invalid item: quantity and price must be positive")
	ErrInvalidStatus = errors.New("invalid order status")
	ErrOrderNotFound = errors.New("order not found")
)

// NewOrderService creates a new OrderService with configuration
func NewOrderService(db *sql.DB, taxRate float64, timeoutSec int) *OrderService {
	if timeoutSec <= 0 {
		timeoutSec = 30
	}
	if taxRate < 0 {
		taxRate = 0
	}
	return &OrderService{db: db, taxRate: taxRate, timeoutSec: timeoutSec}
}

// CreateOrder creates a new order with validation, transaction, and context
func (s *OrderService) CreateOrder(ctx context.Context, userID int, items []OrderItem) (*Order, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	if len(items) == 0 {
		return nil, ErrEmptyItems
	}
	for _, item := range items {
		if item.Qty <= 0 || item.Price < 0 || item.ProductID <= 0 {
			return nil, fmt.Errorf("%w: productID=%d", ErrInvalidItem, item.ProductID)
		}
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeoutSec)*time.Second)
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

	query := `INSERT INTO orders (user_id, total, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	createdAt := time.Now()
	var orderID int
	if err = tx.QueryRowContext(ctx, query, userID, total, "pending", createdAt).Scan(&orderID); err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (order_id, product_id, quantity, price, discount) VALUES ($1, $2, $3, $4, $5)`
	for _, item := range items {
		if _, err = tx.ExecContext(ctx, itemQuery, orderID, item.ProductID, item.Qty, item.Price, item.Discount); err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Order{ID: orderID, UserID: userID, Total: total, Status: "pending", CreatedAt: createdAt, Items: items}, nil
}

// GetOrder retrieves an order by ID with proper error handling
func (s *OrderService) GetOrder(ctx context.Context, id int) (*Order, error) {
	if id <= 0 {
		return nil, fmt.Errorf("invalid order ID: %d", id)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeoutSec)*time.Second)
	defer cancel()

	query := `SELECT id, user_id, total, status, created_at FROM orders WHERE id = $1`
	var order Order
	err := s.db.QueryRowContext(ctx, query, id).Scan(&order.ID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	items, err := s.getOrderItems(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	order.Items = items

	return &order, nil
}

// getOrderItems retrieves items for a specific order
func (s *OrderService) getOrderItems(ctx context.Context, orderID int) ([]OrderItem, error) {
	query := `SELECT product_id, quantity, price, COALESCE(discount, 0) FROM order_items WHERE order_id = $1`
	rows, err := s.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.Qty, &item.Price, &item.Discount); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// UpdateStatus updates order status with validation and audit trail
func (s *OrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	if !validStatuses[status] {
		return fmt.Errorf("%w: %s", ErrInvalidStatus, status)
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeoutSec)*time.Second)
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
	if err = tx.QueryRowContext(ctx, `SELECT status FROM orders WHERE id = $1`, id).Scan(&oldStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrOrderNotFound
		}
		return fmt.Errorf("failed to get current status: %w", err)
	}

	result, err := tx.ExecContext(ctx, `UPDATE orders SET status = $1, updated_at = $2 WHERE id = $3`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return ErrOrderNotFound
	}

	auditQuery := `INSERT INTO order_status_audit (order_id, old_status, new_status, changed_at) VALUES ($1, $2, $3, $4)`
	if _, err = tx.ExecContext(ctx, auditQuery, id, oldStatus, status, time.Now()); err != nil {
		return fmt.Errorf("failed to insert audit record: %w", err)
	}

	return tx.Commit()
}

// SearchOrders searches orders with pagination and parameterized queries
func (s *OrderService) SearchOrders(ctx context.Context, params SearchParams) ([]Order, error) {
	if params.Limit <= 0 {
		params.Limit = 20
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(s.timeoutSec)*time.Second)
	defer cancel()

	query := `SELECT id, user_id, total, status, created_at FROM orders WHERE status LIKE $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3`
	rows, err := s.db.QueryContext(ctx, query, "%"+params.Keyword+"%", params.Limit, params.Offset)
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

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating orders: %w", err)
	}
	return orders, nil
}

// CalculateTotal calculates total with discounts and tax using proper precision
func (s *OrderService) CalculateTotal(items []OrderItem) float64 {
	var subtotal float64
	for _, item := range items {
		itemTotal := item.Price * float64(item.Qty)
		if item.Discount > 0 && item.Discount <= 100 {
			itemTotal -= itemTotal * (item.Discount / 100)
		}
		subtotal += itemTotal
	}
	total := subtotal + (subtotal * s.taxRate)
	return math.Round(total*100) / 100
}
