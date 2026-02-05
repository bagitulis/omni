package services

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAnalyticsService(t *testing.T) {
	svc := NewAnalyticsService(nil)
	assert.NotNil(t, svc)
}

func TestCalculateAOV(t *testing.T) {
	tests := []struct {
		name        string
		totalSales  float64
		totalOrders int
		expected    float64
	}{
		{
			name:        "normal case",
			totalSales:  1000.0,
			totalOrders: 10,
			expected:    100.0,
		},
		{
			name:        "zero orders returns zero",
			totalSales:  500.0,
			totalOrders: 0,
			expected:    0,
		},
		{
			name:        "fractional result",
			totalSales:  100.0,
			totalOrders: 3,
			expected:    33.333333333333336,
		},
		{
			name:        "zero sales zero orders",
			totalSales:  0,
			totalOrders: 0,
			expected:    0,
		},
		{
			name:        "large numbers",
			totalSales:  1000000.0,
			totalOrders: 5000,
			expected:    200.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateAOV(tt.totalSales, tt.totalOrders)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscrowSyncStatus(t *testing.T) {
	status := &EscrowSyncStatus{
		Month:      1,
		Year:       2024,
		ShopeeSync: nil,
		TiktokSync: nil,
	}

	assert.Equal(t, 1, status.Month)
	assert.Equal(t, 2024, status.Year)
	assert.Nil(t, status.ShopeeSync)
	assert.Nil(t, status.TiktokSync)
}

func TestEscrowSyncStatus_AllFields(t *testing.T) {
	tests := []struct {
		name  string
		month int
		year  int
	}{
		{name: "january", month: 1, year: 2024},
		{name: "december", month: 12, year: 2024},
		{name: "mid year", month: 6, year: 2023},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := &EscrowSyncStatus{
				Month: tt.month,
				Year:  tt.year,
			}
			assert.Equal(t, tt.month, status.Month)
			assert.Equal(t, tt.year, status.Year)
		})
	}
}

func TestCalculateAOV_EdgeCases(t *testing.T) {
	// Test very small numbers
	result := calculateAOV(0.01, 1)
	assert.Equal(t, 0.01, result)

	// Test very large sales with single order
	result = calculateAOV(999999999.99, 1)
	assert.Equal(t, 999999999.99, result)

	// Test many orders with small sales
	result = calculateAOV(10.0, 1000)
	assert.Equal(t, 0.01, result)
}
