// Package models provides database models for analytics aggregations
package models

import "time"

// SalesAnalytics represents aggregated sales analytics data
type SalesAnalytics struct {
	TotalSales   float64            `json:"total_sales"`
	TotalOrders  int                `json:"total_orders"`
	AverageOrder float64            `json:"average_order"`
	ByPlatform   map[string]float64 `json:"by_platform"`
	PeriodStart  time.Time          `json:"period_start"`
	PeriodEnd    time.Time          `json:"period_end"`
}

// OrderAnalytics represents aggregated order analytics data
type OrderAnalytics struct {
	TotalOrders     int            `json:"total_orders"`
	ByPlatform      map[string]int `json:"by_platform"`
	ByStatus        map[string]int `json:"by_status"`
	FulfillmentRate float64        `json:"fulfillment_rate"`
	ReturnRate      float64        `json:"return_rate"`
	AverageValue    float64        `json:"average_value"`
}

// RevenueAnalytics represents aggregated revenue analytics data
type RevenueAnalytics struct {
	GrossRevenue   float64            `json:"gross_revenue"`
	NetRevenue     float64            `json:"net_revenue"`
	TotalFees      float64            `json:"total_fees"`
	TotalDiscounts float64            `json:"total_discounts"`
	ByPlatform     map[string]float64 `json:"by_platform"`
	ByMonth        map[string]float64 `json:"by_month"`
	PeriodStart    time.Time          `json:"period_start"`
	PeriodEnd      time.Time          `json:"period_end"`
}
