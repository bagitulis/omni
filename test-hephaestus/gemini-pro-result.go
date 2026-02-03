package services

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// OrderService handles order operations with proper validation and error handling.
type OrderService struct {
	db *sql.DB
}

// Order represents an order in the system.
type Order struct {
	ID        int64       `json:"id"`
	UserID    int64       `json:"user_id"`
	Total     float64     `json:"total"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	Items     []OrderItem `json:"items"`
}

// OrderItem represents a line item in an order.
type OrderItem struct {
	ProductID int64   `json:"product_id"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

// Valid order statuses
const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusCancelled = "cancelled"
)

var validStatuses = map[string]bool{
	StatusPending:   true,
	StatusPaid:      true,
	StatusShipped:   true,
	StatusCancelled: true,
}

// NewOrderService creates a new instance of OrderService.
func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

// CreateOrder creates a new order with transaction support and validation.
func (s *OrderService) CreateOrder(ctx context.Context, userID int64, items []OrderItem) (*Order, error) {
	// Input Validation
	if userID <= 0 {
		return nil, errors.New("invalid user ID")
	}
	if len(items) == 0 {
		return nil, errors.New("items cannot be empty")
	}

	// Calculate Total with precision handling
	total, err := s.CalculateTotal(items)
	if err != nil {
		return nil, err
	}

	// Begin Transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert Order (PostgreSQL style with RETURNING id)
	var orderID int64
	query := `INSERT INTO orders (user_id, total, status, created_at) VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRowContext(ctx, query, userID, total, StatusPending, time.Now()).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert Items
	itemQuery := `INSERT INTO order_items (order_id, product_id, qty, price) VALUES ($1, $2, $3, $4)`
	stmt, err := tx.PrepareContext(ctx, itemQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare item statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.ExecContext(ctx, orderID, item.ProductID, item.Qty, item.Price); err != nil {
			return nil, fmt.Errorf("failed to insert order item (product_id: %d): %w", item.ProductID, err)
		}
	}

	// Commit Transaction
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Order{
		ID:        orderID,
		UserID:    userID,
		Total:     total,
		Status:    StatusPending,
		CreatedAt: time.Now(), // Approximate for return
		Items:     items,
	}, nil
}

// GetOrder retrieves an order by ID including its items.
func (s *OrderService) GetOrder(ctx context.Context, id int64) (*Order, error) {
	if id <= 0 {
		return nil, errors.New("invalid order ID")
	}

	var order Order
	query := `SELECT id, user_id, total, status, created_at FROM orders WHERE id = $1`
	err := s.db.QueryRowContext(ctx, query, id).Scan(&order.ID, &order.UserID, &order.Total, &order.Status, &order.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("order not found")
		}
		return nil, fmt.Errorf("failed to fetch order: %w", err)
	}

	// Fetch Items
	itemsQuery := `SELECT product_id, qty, price FROM order_items WHERE order_id = $1`
	rows, err := s.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		if err := rows.Scan(&item.ProductID, &item.Qty, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		order.Items = append(order.Items, item)
	}

	return &order, nil
}

// UpdateStatus updates the status of an order safely.
func (s *OrderService) UpdateStatus(ctx context.Context, id int64, status string) error {
	if id <= 0 {
		return errors.New("invalid order ID")
	}
	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	query := `UPDATE orders SET status = $1 WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, status, id)
	if err != nil {
		return fmt.Errorf("failed to update status: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("order not found")
	}

	return nil
}

// SearchOrders searches orders with pagination and validation.
func (s *OrderService) SearchOrders(ctx context.Context, keyword string, page, limit int) ([]Order, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}
	offset := (page - 1) * limit

	// Secure search with ILIKE (Postgres case-insensitive)
	// keyword should be sanitized or handled by driver params.
	// We wrap with % for wildcards here.
	searchTerm := "%" + strings.TrimSpace(keyword) + "%"

	query := `SELECT id, user_id, total, status, created_at FROM orders WHERE status ILIKE $1 LIMIT $2 OFFSET $3`
	rows, err := s.db.QueryContext(ctx, query, searchTerm, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
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
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, nil
}

// CalculateTotal calculates the total price for items with basic validation.
func (s *OrderService) CalculateTotal(items []OrderItem) (float64, error) {
	var total float64
	for _, item := range items {
		if item.Qty <= 0 {
			return 0, fmt.Errorf("invalid quantity for product %d", item.ProductID)
		}
		if item.Price < 0 {
			return 0, fmt.Errorf("invalid price for product %d", item.ProductID)
		}
		// Simple float arithmetic (caution: floating point errors possible)
		// Ideally use decimal type or cents (int64).
		// Rounding to 2 decimal places:
		lineTotal := item.Price * float64(item.Qty)
		total += lineTotal
	}

	// Round to 2 decimals
	return math.Round(total*100) / 100, nil
}
