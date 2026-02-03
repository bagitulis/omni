package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

// OrderService handles order operations - NEEDS REFACTORING
// Issues: unused imports, missing error handling, type inconsistencies,
// hardcoded values, missing context, poor naming

type OrderService struct {
	db *sql.DB
}

type Order struct {
	ID        int
	UserID    int
	Total     float64
	Status    string
	CreatedAt time.Time
	Items     []OrderItem
}

type OrderItem struct {
	ProductID int
	Qty       int
	Price     float64
}

// CreateOrder - has multiple issues that need fixing
func (s *OrderService) CreateOrder(userId int, items []OrderItem) (*Order, error) {
	// BUG 1: No input validation
	// BUG 2: Hardcoded timeout
	// BUG 3: No transaction
	// BUG 4: SQL injection vulnerable
	// BUG 5: No context propagation

	total := 0.0
	for _, item := range items {
		total += item.Price * float64(item.Qty)
	}

	query := fmt.Sprintf("INSERT INTO orders (user_id, total, status) VALUES (%d, %f, 'pending')", userId, total)
	result, err := s.db.Exec(query)
	if err != nil {
		return nil, err
	}

	id, _ := result.LastInsertId() // BUG 6: Ignoring error

	// BUG 7: Items not saved

	return &Order{
		ID:     int(id),
		UserID: userId,
		Total:  total,
		Status: "pending",
	}, nil
}

// GetOrder - missing proper error handling
func (s *OrderService) GetOrder(id int) *Order {
	// BUG 8: Returns nil on error instead of error
	// BUG 9: No context

	row := s.db.QueryRow("SELECT id, user_id, total, status FROM orders WHERE id = ?", id)

	var order Order
	row.Scan(&order.ID, &order.UserID, &order.Total, &order.Status) // BUG 10: Ignoring scan error

	return &order
}

// UpdateStatus - inconsistent error handling
func (s *OrderService) UpdateStatus(id int, status string) {
	// BUG 11: No return value for error
	// BUG 12: No status validation
	// BUG 13: No audit trail

	s.db.Exec("UPDATE orders SET status = ? WHERE id = ?", status, id)
}

// SearchOrders - inefficient and buggy
func (s *OrderService) SearchOrders(keyword string, limit int) []Order {
	// BUG 14: SQL injection via keyword
	// BUG 15: No pagination
	// BUG 16: Returns empty slice on error

	query := "SELECT * FROM orders WHERE status LIKE '%" + keyword + "%' LIMIT " + fmt.Sprint(limit)
	rows, _ := s.db.Query(query)

	var orders []Order
	for rows.Next() {
		var o Order
		rows.Scan(&o.ID, &o.UserID, &o.Total, &o.Status)
		orders = append(orders, o)
	}

	return orders
}

// CalculateTotal - wrong calculation
func (s *OrderService) CalculateTotal(items []OrderItem) float64 {
	// BUG 17: Doesn't handle discounts
	// BUG 18: No tax calculation
	// BUG 19: Precision issues with float64

	var total float64
	for i := 0; i < len(items); i++ {
		total = total + items[i].Price*float64(items[i].Qty)
	}
	return total
}
