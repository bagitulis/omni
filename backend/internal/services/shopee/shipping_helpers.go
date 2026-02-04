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

	return nil, fmt.Errorf("shipment not ready: please arrange shipment (pickup/dropoff) first to generate tracking number")
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

			// Try to find a slot for tomorrow
			tomorrow := time.Now().AddDate(0, 0, 1)
			tomorrowY, tomorrowM, tomorrowD := tomorrow.Date()

			for _, slot := range addr.TimeSlotList {
				slotTime := time.Unix(slot.Date, 0)
				y, m, d := slotTime.Date()
				if y == tomorrowY && m == tomorrowM && d == tomorrowD {
					pickup.PickupTimeID = slot.PickupTimeID
					log.Info().Str("order_sn", orderSN).Str("slot_id", slot.PickupTimeID).Msg("Selected pickup slot for tomorrow")
					return nil
				}
			}

			// Fallback to first available slot if tomorrow is not found
			pickup.PickupTimeID = addr.TimeSlotList[0].PickupTimeID
			log.Info().Str("order_sn", orderSN).Str("slot_id", pickup.PickupTimeID).Msg("Selected first available pickup slot (tomorrow not available)")
			return nil
		}
	}

	return fmt.Errorf("pickup address ID not found in available options")
}
