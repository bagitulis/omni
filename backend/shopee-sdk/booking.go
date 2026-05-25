package shopeesdk

import (
	"context"
	"errors"
	"strings"
	"time"

	shopee "github.com/omni/backend/pkg/shopee"
)

// Type aliases for booking response types.
type (
	GetBookingListResponse   = shopee.GetBookingListResponse
	GetBookingDetailResponse = shopee.GetBookingDetailResponse
)

// BookingListParams controls booking list queries.
type BookingListParams struct {
	BookingStatus  string
	TimeFrom       time.Time
	TimeTo         time.Time
	TimeRangeField string // create_time or update_time
	PageSize       int
	Cursor         string
}

// BookingDetailParams controls booking detail lookups.
type BookingDetailParams struct {
	BookingSNs []string
}

// GetBookingList fetches a list of bookings within a time window (<=15 days).
func (c *Client) GetBookingList(ctx context.Context, params BookingListParams) (*GetBookingListResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}

	end := params.TimeTo
	if end.IsZero() {
		end = time.Now()
	}
	start := params.TimeFrom
	if start.IsZero() {
		start = end.AddDate(0, 0, -15)
	}
	timeRangeField := params.TimeRangeField
	if timeRangeField == "" {
		timeRangeField = "create_time"
	}
	pageSize := params.PageSize
	if pageSize == 0 {
		pageSize = 50
	}

	return c.api.GetBookingList(ctx, &shopee.GetBookingListRequest{
		TimeRangeField: timeRangeField,
		TimeFrom:       start.Unix(),
		TimeTo:         end.Unix(),
		PageSize:       pageSize,
		Cursor:         params.Cursor,
		BookingStatus:  params.BookingStatus,
	})
}

// GetBookingDetails fetches booking details for up to 50 booking SNs.
func (c *Client) GetBookingDetails(ctx context.Context, params BookingDetailParams) (*GetBookingDetailResponse, error) {
	if err := apiReady(ctx); err != nil {
		return nil, err
	}
	if len(params.BookingSNs) == 0 {
		return nil, errors.New("booking_sns is required")
	}

	return c.api.GetBookingDetail(ctx, &shopee.GetBookingDetailRequest{
		BookingSNList: strings.Join(params.BookingSNs, ","),
	})
}
