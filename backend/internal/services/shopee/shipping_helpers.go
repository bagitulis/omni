package shopee

import (
	"context"
	"fmt"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// ensureShipmentReady validates if tracking number exists with retries
func (s *ShippingService) ensureShipmentReady(ctx context.Context, client shippingClient, orderSN, packageNumber string) (*shopeePkg.GetTrackingNumberResponse, error) {
	var resp *shopeePkg.GetTrackingNumberResponse
	var err error

	for i := 0; i < 3; i++ {
		resp, err = client.GetTrackingNumber(orderSN)
		if err == nil && resp != nil && resp.Response.TrackingNumber != "" {
			return resp, nil
		}

		if i < 2 {
			log.Warn().
				Str("order_sn", orderSN).
				Str("package_number", packageNumber).
				Int("attempt", i+1).
				Msg("Tracking number not ready, retrying...")
			time.Sleep(500 * time.Millisecond)
		}
	}

	if err != nil {
		return nil, err
	}

	return resp, nil
}

// applyDefaultPickupTime selects the preferred pickup time slot (tomorrow) or first available
func (s *ShippingService) applyDefaultPickupTime(ctx context.Context, client shippingClient, orderSN string, pickup *shopeePkg.PickupInfo) error {
	if pickup == nil || pickup.AddressID == 0 {
		return nil
	}

	if pickup.PickupTimeID != "" {
		return nil
	}

	resp, err := client.GetShippingParameter(orderSN)
	if err != nil {
		return fmt.Errorf("get shipping parameter: %w", err)
	}

	if resp == nil {
		return fmt.Errorf("invalid shipping parameter response")
	}

	for _, addr := range resp.Response.Pickup.AddressList {
		if addr.AddressID == pickup.AddressID {
			if len(addr.TimeSlotList) == 0 {
				return fmt.Errorf("no pickup time slots available")
			}

			// Pick the first (earliest) available slot
			earliestSlot := addr.TimeSlotList[0]
			pickup.PickupTimeID = earliestSlot.PickupTimeID

			// Log which slot was selected with date and time info
			slotTime := time.Unix(earliestSlot.Date, 0)
			log.Info().
				Str("order_sn", orderSN).
				Str("pickup_time_id", earliestSlot.PickupTimeID).
				Time("slot_date", slotTime).
				Str("slot_time_text", earliestSlot.TimeText).
				Msg("Selected earliest available pickup slot")

			return nil
		}
	}

	return fmt.Errorf("pickup address ID not found in available options")
}
