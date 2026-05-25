package sync

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

type mockBookingManager struct {
	list    []Booking
	details map[string]Booking
	err     error
}

func (m *mockBookingManager) GetBookingList(context.Context, int) ([]Booking, error) {
	return m.list, m.err
}

func (m *mockBookingManager) GetBookingDetails(_ context.Context, bookingSNs []string) (map[string]Booking, error) {
	if m.err != nil {
		return nil, m.err
	}
	result := make(map[string]Booking)
	for _, sn := range bookingSNs {
		if booking, ok := m.details[sn]; ok {
			result[sn] = booking
		}
	}
	return result, nil
}

func (m *mockBookingManager) GetPlatform() string { return "shopee" }

type mockBookingRepository struct {
	upserted      []Booking
	replacedItems map[string][]BookingItem
	upsertErr     error
	replaceErr    map[string]error
}

func (r *mockBookingRepository) UpsertBookings(_ context.Context, _ string, _ uint64, bookings []Booking) error {
	r.upserted = append(r.upserted, bookings...)
	return r.upsertErr
}

func (r *mockBookingRepository) ReplaceBookingItems(_ context.Context, _ string, _ uint64, bookingSN string, items []BookingItem) error {
	if err := r.replaceErr[bookingSN]; err != nil {
		return err
	}
	r.replacedItems[bookingSN] = items
	return nil
}

func (r *mockBookingRepository) ListBookings(context.Context, string, BookingListParams) ([]Booking, int64, error) {
	return r.upserted, int64(len(r.upserted)), nil
}

func (r *mockBookingRepository) GetBookingDetail(_ context.Context, _ string, _ uint64, bookingSN string) (*Booking, error) {
	for _, booking := range r.upserted {
		if booking.BookingSN == bookingSN {
			return &booking, nil
		}
	}
	return nil, nil
}

func TestRawToBookingParsesBookingAndItems(t *testing.T) {
	raw := map[string]any{
		"booking_sn":       "BSN-1",
		"order_sn":         "OSN-1",
		"booking_status":   "READY_TO_SHIP",
		"match_status":     "MATCHED",
		"region":           "ID",
		"shipping_carrier": "SPX",
		"create_time":      int64(1710000000),
		"update_time":      float64(1710000100),
		"recipient_address": map[string]any{
			"name":  "Test User",
			"phone": "+1-555-0100",
		},
		"items": []any{map[string]any{
			"item_id":    int64(10),
			"model_id":   int64(20),
			"item_name":  "Test Item",
			"model_name": "Blue",
			"item_sku":   "ITEM-SKU",
			"model_sku":  "MODEL-SKU",
			"quantity":   float64(2),
			"image_info": map[string]any{"image_url": "https://example.com/image.jpg"},
		}},
	}

	booking, items, err := rawToBooking(raw)

	require.NoError(t, err)
	require.Equal(t, "BSN-1", booking.BookingSN)
	require.Equal(t, "Test User", booking.RecipientName)
	require.JSONEq(t, `{"name":"Test User","phone":"+1-555-0100"}`, booking.RecipientAddressJSON)
	require.Len(t, items, 1)
	require.Equal(t, "MODEL-SKU", items[0].SKU)
	require.Equal(t, "10|20|ITEM-SKU|MODEL-SKU|Test Item|Blue", items[0].LineKey)
}

func TestSyncBookingsPersistsSuccessesAndReportsItemFailures(t *testing.T) {
	repo := &mockBookingRepository{replacedItems: map[string][]BookingItem{}, replaceErr: map[string]error{"BSN-2": errors.New("insert failed")}}
	service := &BookingSyncService{
		manager: &mockBookingManager{
			list: []Booking{{BookingSN: "BSN-1"}, {BookingSN: "BSN-2"}},
			details: map[string]Booking{
				"BSN-1": {BookingSN: "BSN-1", Items: []BookingItem{{LineKey: "1|0||||"}}},
				"BSN-2": {BookingSN: "BSN-2", Items: []BookingItem{{LineKey: "2|0||||"}}},
			},
		},
		repository: repo,
		tenantID:   "tenant_test",
		shopID:     123,
	}

	result := service.SyncBookings(context.Background(), 15)

	require.False(t, result.Success)
	require.Equal(t, 1, result.Count)
	require.Equal(t, 1, result.FailedCount)
	require.Len(t, result.Errors, 1)
	require.Len(t, repo.upserted, 2)
	require.Len(t, repo.replacedItems["BSN-1"], 1)
}
