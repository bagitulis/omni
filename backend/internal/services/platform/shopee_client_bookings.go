package platform

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// GetBookingList fetches booking list with cursor pagination.
// Uses update_time as time_range_field since bookings update over time.
func (c *ShopeeAPIClient) GetBookingList(ctx context.Context, days int, pageSize int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		expiry := c.config.GetTokenExpiry()
		if expiry > 0 && time.Now().UnixMilli() >= expiry {
			// Hard-expired: token is definitely expired and refresh failed — abort
			return nil, fmt.Errorf("shopee: token expired and refresh failed: %w", err)
		}
		// Buffer/unknown expiry (expiry==0 or within buffer) — warn and continue
		log.Warn().Err(err).Str("platform", "shopee").Msg("Token validation warning, continuing with existing token")
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	shopID, _ := strconv.ParseInt(c.config.ShopID, 10, 64)
	c.client.SetShopCredentials(shopID, c.config.GetAccessToken())

	timeTo := time.Now().Unix()
	timeFrom := time.Now().AddDate(0, 0, -days).Unix()

	var allBookings []map[string]interface{}
	cursor := ""

	for {
		response, err := c.client.GetBookingList(ctx, &shopeePkg.GetBookingListRequest{
			TimeRangeField: "update_time",
			TimeFrom:       timeFrom,
			TimeTo:         timeTo,
			PageSize:       pageSize,
			Cursor:         cursor,
		})
		if err != nil {
			return allBookings, fmt.Errorf("get booking list: %w", err)
		}

		if response == nil || len(response.Response.BookingList) == 0 {
			break
		}

		for _, booking := range response.Response.BookingList {
			allBookings = append(allBookings, map[string]interface{}{
				"booking_sn":     booking.BookingSn,
				"order_sn":       booking.OrderSn,
				"booking_status": booking.BookingStatus,
			})
		}

		if !response.Response.More || response.Response.NextCursor == "" {
			break
		}
		cursor = response.Response.NextCursor
	}

	if len(allBookings) == 0 {
		return []map[string]interface{}{}, nil
	}

	log.Info().Str("platform", "shopee").Int("count", len(allBookings)).Msg("Fetched booking list")
	return allBookings, nil
}

// GetBookingDetails fetches booking details by booking SNs in batches of 50 (Shopee API limit).
func (c *ShopeeAPIClient) GetBookingDetails(ctx context.Context, bookingSNs []string) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// IMPORTANT: Check token expiry and refresh if needed BEFORE API call
	if err := c.config.EnsureValidToken(ctx); err != nil {
		expiry := c.config.GetTokenExpiry()
		if expiry > 0 && time.Now().UnixMilli() >= expiry {
			// Hard-expired: token is definitely expired and refresh failed — abort
			return nil, fmt.Errorf("shopee: token expired and refresh failed: %w", err)
		}
		// Buffer/unknown expiry (expiry==0 or within buffer) — warn and continue
		log.Warn().Err(err).Str("platform", "shopee").Msg("Token validation warning, continuing with existing token")
	}

	// Reload config from DB to get latest token (after potential refresh)
	if err := c.config.LoadConfig(ctx); err != nil {
		return nil, fmt.Errorf("reload config: %w", err)
	}

	shopID, _ := strconv.ParseInt(c.config.ShopID, 10, 64)
	c.client.SetShopCredentials(shopID, c.config.GetAccessToken())

	var allBookings []map[string]interface{}

	for i := 0; i < len(bookingSNs); i += 50 {
		end := i + 50
		if end > len(bookingSNs) {
			end = len(bookingSNs)
		}
		batch := bookingSNs[i:end]

		response, err := c.client.GetBookingDetail(ctx, &shopeePkg.GetBookingDetailRequest{
			BookingSNList: strings.Join(batch, ","),
		})
		if err != nil {
			return allBookings, fmt.Errorf("get booking details at batch %d: %w", i/50, err)
		}

		if response == nil || len(response.Response.BookingList) == 0 {
			continue
		}

		for _, booking := range response.Response.BookingList {
			// Build recipient_address nested map
			recipientAddress := map[string]interface{}{}
			if booking.RecipientAddress != nil {
				recipientAddress = map[string]interface{}{
					"name":         booking.RecipientAddress.Name,
					"phone":        booking.RecipientAddress.Phone,
					"town":         booking.RecipientAddress.Town,
					"district":     booking.RecipientAddress.District,
					"city":         booking.RecipientAddress.City,
					"state":        booking.RecipientAddress.State,
					"region":       booking.RecipientAddress.Region,
					"zipcode":      booking.RecipientAddress.Zipcode,
					"full_address": booking.RecipientAddress.FullAddress,
				}
			}

			// Build items nested array
			items := make([]interface{}, 0, len(booking.ItemList))
			for _, item := range booking.ItemList {
				items = append(items, map[string]interface{}{
					"item_name":           item.ItemName,
					"model_name":          item.ModelName,
					"item_sku":            item.ItemSku,
					"model_sku":           item.ModelSku,
					"weight":              item.Weight,
					"product_location_id": item.ProductLocationId,
					"image_info": map[string]interface{}{
						"image_url": item.ImageInfo.ImageURL,
					},
				})
			}

			allBookings = append(allBookings, map[string]interface{}{
				"booking_sn":        booking.BookingSn,
				"order_sn":          booking.OrderSn,
				"booking_status":    booking.BookingStatus,
				"match_status":      booking.MatchStatus,
				"region":            booking.Region,
				"shipping_carrier":  booking.ShippingCarrier,
				"create_time":       booking.CreateTime,
				"update_time":       booking.UpdateTime,
				"fulfillment_flag":  booking.FulfillmentFlag,
				"pickup_done_time":  booking.PickupDoneTime,
				"recipient_address": recipientAddress,
				"items":             items,
			})
		}
	}

	if allBookings == nil {
		allBookings = []map[string]interface{}{}
	}

	return allBookings, nil
}
