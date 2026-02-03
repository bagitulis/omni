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
	ErrInvalidInput  = errors.New("invalid input parameters")
	ErrOrderNotFound = errors.New("order not found")
)

// Order statuses
const (
	StatusPending   = "pending"
	StatusPaid      = "paid"
	StatusShipped   = "shipped"
	StatusCancelled = "cancelled"
)

// OrderService handles order operations
type OrderService struct {
	db *sql.DB
}

// NewOrderService creates a new instance of OrderService
func NewOrderService(db *sql.DB) *OrderService {
	return &OrderService{db: db}
}

type Order struct {
	ID        int         `json:"id"`
	UserID    int         `json:"user_id"`
	Total     float64     `json:"total"`
	SubTotal  float64     `json:"sub_total"`
	Tax       float64     `json:"tax"`
	Discount  float64     `json:"discount"`
	Status    string      `json:"status"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
	Items     []OrderItem `json:"items"`
}

type OrderItem struct {
	ID        int     `json:"id"`
	OrderID   int     `json:"order_id"`
	ProductID int     `json:"product_id"`
	Qty       int     `json:"qty"`
	Price     float64 `json:"price"`
}

// CreateOrder creates a new order with transaction support and validation
func (s *OrderService) CreateOrder(ctx context.Context, userID int, items []OrderItem) (*Order, error) {
	if userID <= 0 {
		return nil, fmt.Errorf("%w: invalid user ID", ErrInvalidInput)
	}
	if len(items) == 0 {
		return nil, fmt.Errorf("%w: items cannot be empty", ErrInvalidInput)
	}

	subTotal := s.round(s.calculateSubTotal(items))
	tax := s.round(s.calculateTax(subTotal))
	discount := 0.0 // Placeholder for discount logic if needed
	total := s.round(subTotal + tax - discount)

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	query := `
		INSERT INTO orders (user_id, sub_total, tax, discount, total, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	var orderID int
	// BUG FIX: Using QueryRowContext with RETURNING id for Postgres
	err = tx.QueryRowContext(ctx, query, userID, subTotal, tax, discount, total, StatusPending, now, now).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `INSERT INTO order_items (order_id, product_id, qty, price) VALUES ($1, $2, $3, $4)`
	stmt, err := tx.PrepareContext(ctx, itemQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare item statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if item.Qty <= 0 || item.Price < 0 {
			return nil, fmt.Errorf("%w: invalid item quantity or price", ErrInvalidInput)
		}
		if _, err := stmt.ExecContext(ctx, orderID, item.ProductID, item.Qty, item.Price); err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &Order{
		ID:        orderID,
		UserID:    userID,
		Total:     total,
		SubTotal:  subTotal,
		Tax:       tax,
		Discount:  discount,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Items:     items,
	}, nil
}

// GetOrder retrieves an order by ID with context
func (s *OrderService) GetOrder(ctx context.Context, id int) (*Order, error) {
	if id <= 0 {
		return nil, fmt.Errorf("%w: invalid order ID", ErrInvalidInput)
	}

	query := `
		SELECT id, user_id, total, sub_total, tax, discount, status, created_at, updated_at
		FROM orders WHERE id = $1`

	var o Order
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.Total, &o.SubTotal, &o.Tax, &o.Discount, &o.Status, &o.CreatedAt, &o.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrOrderNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	itemsQuery := `SELECT id, product_id, qty, price FROM order_items WHERE order_id = $1`
	rows, err := s.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var item OrderItem
		item.OrderID = o.ID
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Qty, &item.Price); err != nil {
			return nil, fmt.Errorf("failed to scan item: %w", err)
		}
		o.Items = append(o.Items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return &o, nil
}

// UpdateStatus updates the order status with validation
func (s *OrderService) UpdateStatus(ctx context.Context, id int, status string) error {
	if id <= 0 {
		return fmt.Errorf("%w: invalid order ID", ErrInvalidInput)
	}

	validStatuses := map[string]bool{
		StatusPending: true, StatusPaid: true, StatusShipped: true, StatusCancelled: true,
	}
	if !validStatuses[status] {
		return fmt.Errorf("%w: invalid status %s", ErrInvalidInput, status)
	}

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
		return ErrOrderNotFound
	}

	return nil
}

// SearchOrders searches for orders with pagination
func (s *OrderService) SearchOrders(ctx context.Context, statusKeyword string, page, limit int) ([]Order, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	// Prevent SQL injection: using parameterized query
	keywordPattern := "%" + statusKeyword + "%"

	var totalCount int64
	countQuery := `SELECT count(*) FROM orders WHERE status LIKE $1`
	if err := s.db.QueryRowContext(ctx, countQuery, keywordPattern).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count orders: %w", err)
	}

	query := `
		SELECT id, user_id, total, status, created_at 
		FROM orders 
		WHERE status LIKE $1 
		ORDER BY created_at DESC 
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, keywordPattern, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search orders: %w", err)
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status, &o.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan order: %w", err)
		}
		orders = append(orders, o)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return orders, totalCount, nil
}

func (s *OrderService) calculateSubTotal(items []OrderItem) float64 {
	var total float64
	for _, item := range items {
		total += item.Price * float64(item.Qty)
	}
	return total
}

func (s *OrderService) calculateTax(amount float64) float64 {
	const taxRate = 0.10 // 10% Tax
	return amount * taxRate
}

func (s *OrderService) round(val float64) float64 {
	return math.Round(val*100) / 100
}
