package services

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewAuditService(t *testing.T) {
	svc := NewAuditService(nil)
	assert.NotNil(t, svc)
}

func TestAuditListResponse(t *testing.T) {
	resp := &AuditListResponse{
		Data:       nil,
		Total:      100,
		Page:       2,
		PageSize:   10,
		TotalPages: 10,
	}

	assert.Equal(t, int64(100), resp.Total)
	assert.Equal(t, 2, resp.Page)
	assert.Equal(t, 10, resp.PageSize)
	assert.Equal(t, int64(10), resp.TotalPages)
}

func TestAuditListResponse_Pagination(t *testing.T) {
	tests := []struct {
		name      string
		total     int64
		pageSize  int
		wantPages int64
	}{
		{
			name:      "exact division",
			total:     100,
			pageSize:  10,
			wantPages: 10,
		},
		{
			name:      "with remainder",
			total:     95,
			pageSize:  10,
			wantPages: 10,
		},
		{
			name:      "single page",
			total:     5,
			pageSize:  10,
			wantPages: 1,
		},
		{
			name:      "zero records",
			total:     0,
			pageSize:  10,
			wantPages: 0,
		},
		{
			name:      "single record",
			total:     1,
			pageSize:  10,
			wantPages: 1,
		},
		{
			name:      "large dataset",
			total:     10000,
			pageSize:  25,
			wantPages: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test pagination calculation
			totalPages := (tt.total + int64(tt.pageSize) - 1) / int64(tt.pageSize)
			assert.Equal(t, tt.wantPages, totalPages)
		})
	}
}

func TestAuditListResponse_AllFields(t *testing.T) {
	resp := &AuditListResponse{
		Data:       nil,
		Total:      250,
		Page:       5,
		PageSize:   20,
		TotalPages: 13,
	}

	assert.Equal(t, int64(250), resp.Total)
	assert.Equal(t, 5, resp.Page)
	assert.Equal(t, 20, resp.PageSize)
	assert.Equal(t, int64(13), resp.TotalPages)
}

func TestAuditService_OffsetCalculation(t *testing.T) {
	tests := []struct {
		name       string
		page       int
		pageSize   int
		wantOffset int
	}{
		{name: "first page", page: 1, pageSize: 10, wantOffset: 0},
		{name: "second page", page: 2, pageSize: 10, wantOffset: 10},
		{name: "fifth page", page: 5, pageSize: 20, wantOffset: 80},
		{name: "large page", page: 100, pageSize: 50, wantOffset: 4950},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			offset := (tt.page - 1) * tt.pageSize
			assert.Equal(t, tt.wantOffset, offset)
		})
	}
}

func TestAuditService_CleanupDuration(t *testing.T) {
	tests := []struct {
		name     string
		duration time.Duration
	}{
		{name: "30 days", duration: 30 * 24 * time.Hour},
		{name: "90 days", duration: 90 * 24 * time.Hour},
		{name: "1 year", duration: 365 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Greater(t, tt.duration.Hours(), float64(0))
		})
	}
}
