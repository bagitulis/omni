package platform

import (
	"context"
	"fmt"
	"strconv"
	"time"

	shopeePkg "github.com/omni/backend/pkg/shopee"
	"github.com/rs/zerolog/log"
)

// ShopeeAPIClient implements APIClient for Shopee
type ShopeeAPIClient struct {
	config      *ShopeeConfigManager
	client      *shopeePkg.Client
	initialized bool
}

// NewShopeeAPIClient creates a Shopee API client
func NewShopeeAPIClient(config *ShopeeConfigManager) *ShopeeAPIClient {
	c := &ShopeeAPIClient{
		config: config,
	}

	// Initialize client if config is available
	if config.IsConfigured() {
		partnerID, _ := strconv.ParseInt(config.PartnerID, 10, 64)
		shopID, _ := strconv.ParseInt(config.ShopID, 10, 64)

		c.client = shopeePkg.NewClient(partnerID, config.PartnerKey, true)
		c.client.SetShopCredentials(shopID, config.GetAccessToken())
		c.initialized = true
	}

	return c
}

// IsInitialized returns whether client is ready
func (c *ShopeeAPIClient) IsInitialized() bool {
	return c.initialized && c.config.IsConfigured()
}

// GetClient returns the underlying Shopee client for advanced operations
func (c *ShopeeAPIClient) GetClient() *shopeePkg.Client {
	return c.client
}

// GetOrderList fetches orders by status
// For PROCESSED status, uses Package API (searchPackageList) as per Node.js implementation
func (c *ShopeeAPIClient) GetOrderList(ctx context.Context, status string, days int) ([]map[string]interface{}, error) {
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

	// For PROCESSED status, use Package API (same as Node.js getProcessedOrdersWithTracking)
	if status == "PROCESSED" {
		return c.getProcessedOrdersViaPackageAPI(ctx)
	}

	// For other statuses, use regular order list API with pagination
	return c.getOrdersViaOrderListAPI(ctx, status, days)
}

// getOrdersViaOrderListAPI fetches orders using get_order_list API
func (c *ShopeeAPIClient) getOrdersViaOrderListAPI(ctx context.Context, status string, days int) ([]map[string]interface{}, error) {
	timeTo := time.Now().Unix()
	timeFrom := time.Now().AddDate(0, 0, -days).Unix()

	timeRangeField := "create_time"
	if status == "SHIPPED" || status == "COMPLETED" {
		timeRangeField = "update_time"
	}

	var allOrders []map[string]interface{}
	cursor := ""

	for {
		response, err := c.client.GetOrderList(timeFrom, timeTo, timeRangeField, status, cursor)
		if err != nil {
			return allOrders, fmt.Errorf("get order list: %w", err)
		}

		if response == nil || response.Response.OrderList == nil {
			break
		}

		for _, order := range response.Response.OrderList {
			allOrders = append(allOrders, map[string]interface{}{
				"order_sn": order.OrderSN,
				"platform": "shopee",
				"status":   status,
			})
		}

		if !response.Response.More || response.Response.NextCursor == "" {
			break
		}
		cursor = response.Response.NextCursor
	}

	return allOrders, nil
}

// getProcessedOrdersViaPackageAPI fetches processed orders using Package API
// This matches Node.js shopeeOrderManager.getProcessedOrdersWithTracking()
// Flow: searchPackageList(3) → getPackageDetail → getOrderDetails → Merge
func (c *ShopeeAPIClient) getProcessedOrdersViaPackageAPI(ctx context.Context) ([]map[string]interface{}, error) {
	// Step 1: Get all packages with package_status=3 (Processed)
	var allPackages []shopeePkg.PackageBasic
	cursor := ""

	for {
		response, err := c.client.SearchPackageList(3, cursor, 100) // 3 = Processed
		if err != nil {
			return nil, fmt.Errorf("search package list: %w", err)
		}

		if response == nil || len(response.Response.PackagesList) == 0 {
			break
		}

		allPackages = append(allPackages, response.Response.PackagesList...)

		if !response.Response.Pagination.More || response.Response.Pagination.NextCursor == "" {
			break
		}
		cursor = response.Response.Pagination.NextCursor
	}

	if len(allPackages) == 0 {
		return []map[string]interface{}{}, nil
	}

	log.Info().Str("platform", "shopee").Int("count", len(allPackages)).Msg("Found processed packages via Package API")

	// Step 2: Get package details (tracking, carrier, SKU, qty) in batches of 50
	packageInfoMap := make(map[string]*shopeePkg.PackageDetail)
	packageNumbers := make([]string, 0, len(allPackages))
	for _, pkg := range allPackages {
		if pkg.PackageNumber != "" {
			packageNumbers = append(packageNumbers, pkg.PackageNumber)
		}
	}

	for i := 0; i < len(packageNumbers); i += 50 {
		end := i + 50
		if end > len(packageNumbers) {
			end = len(packageNumbers)
		}
		batch := packageNumbers[i:end]

		detailResp, err := c.client.GetPackageDetail(batch)
		if err != nil {
			log.Warn().Err(err).Str("platform", "shopee").Msg("Failed to get package detail")
			continue
		}

		for idx := range detailResp.Response.PackageList {
			pkg := &detailResp.Response.PackageList[idx]
			packageInfoMap[pkg.OrderSN] = pkg
		}
	}

	// Step 3: Get order details for item_name and model_name
	orderSNs := make([]string, 0, len(allPackages))
	seen := make(map[string]bool)
	for _, pkg := range allPackages {
		if !seen[pkg.OrderSN] {
			orderSNs = append(orderSNs, pkg.OrderSN)
			seen[pkg.OrderSN] = true
		}
	}

	// Fetch order details in batches of 50
	orderDetails := make(map[string]map[string]interface{})
	for i := 0; i < len(orderSNs); i += 50 {
		end := i + 50
		if end > len(orderSNs) {
			end = len(orderSNs)
		}
		batch := orderSNs[i:end]

		details, err := c.GetOrderDetails(ctx, batch)
		if err != nil {
			log.Warn().Err(err).Str("platform", "shopee").Msg("Failed to get order details")
			continue
		}

		for _, detail := range details {
			if orderSN, ok := detail["order_sn"].(string); ok {
				orderDetails[orderSN] = detail
			}
		}
	}

	// Step 4: Build result - merge package info with order details
	result := make([]map[string]interface{}, 0, len(allPackages))
	for _, pkg := range allPackages {
		pkgDetail := packageInfoMap[pkg.OrderSN]
		orderDetail := orderDetails[pkg.OrderSN]

		trackingNumber := ""
		shippingCarrier := ""
		if pkgDetail != nil {
			trackingNumber = pkgDetail.TrackingNumber
			if trackingNumber == "" || trackingNumber == "-" {
				trackingNumber = pkgDetail.PackageNumber
			}
			shippingCarrier = pkgDetail.ShippingCarrier
		}

		result = append(result, map[string]interface{}{
			"order_sn":         pkg.OrderSN,
			"package_number":   pkg.PackageNumber,
			"tracking_number":  trackingNumber,
			"shipping_carrier": shippingCarrier,
			"platform":         "shopee",
			"status":           "PROCESSED",
			"order_detail":     orderDetail,
		})
	}

	log.Info().Str("platform", "shopee").Int("count", len(result)).Msg("Returning processed orders with tracking info")
	return result, nil
}

// GetOrderDetails fetches order details
func (c *ShopeeAPIClient) GetOrderDetails(ctx context.Context, orderIDs []string) ([]map[string]interface{}, error) {
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

	// Call the actual client method (GetOrderDetail, singular)
	response, err := c.client.GetOrderDetail(orderIDs)
	if err != nil {
		return nil, fmt.Errorf("get order details: %w", err)
	}

	if response == nil || response.Response.OrderList == nil {
		return []map[string]interface{}{}, nil
	}

	// Debug: Log first order details to verify API response
	if len(response.Response.OrderList) > 0 {
		first := response.Response.OrderList[0]
		log.Debug().
			Str("platform", "shopee").
			Str("order_sn", first.OrderSN).
			Float64("total_amount", first.TotalAmount).
			Str("buyer", first.BuyerUsername).
			Str("payment", first.PaymentMethod).
			Str("carrier", first.ShippingCarrier).
			Msg("Sample order details")
	}

	// Convert to generic format - includes items
	result := make([]map[string]interface{}, 0, len(response.Response.OrderList))
	for _, order := range response.Response.OrderList {
		// Convert item_list to generic format
		items := make([]interface{}, 0, len(order.ItemList))
		for _, item := range order.ItemList {
			itemMap := map[string]interface{}{
				"item_id":    item.ItemID,
				"model_id":   item.ModelID,
				"item_name":  item.ItemName,
				"model_name": item.ModelName,
				"item_sku":   item.ItemSKU,
				"model_sku":  item.ModelSKU,
				"quantity":   item.ModelQuantityPurchased,
				"price":      item.ModelOriginalPrice,
			}
			if item.ImageInfo != nil && item.ImageInfo.ImageURL != "" {
				itemMap["image_info"] = map[string]interface{}{
					"image_url": item.ImageInfo.ImageURL,
				}
			}
			items = append(items, itemMap)
		}

		// Determine best shipping carrier from available fields
		shippingCarrier := order.ShippingCarrier
		if shippingCarrier == "" {
			shippingCarrier = order.CheckoutShippingCarrier
		}

		// Determine buyer message
		buyerMessage := order.Note
		if buyerMessage == "" {
			buyerMessage = order.MessageToSeller
		}

		result = append(result, map[string]interface{}{
			"order_sn":         order.OrderSN,
			"status":           order.OrderStatus,
			"total_amount":     order.TotalAmount,
			"currency":         order.Currency,
			"buyer_user_id":    order.BuyerUserID,
			"buyer_username":   order.BuyerUsername,
			"payment_method":   order.PaymentMethod,
			"shipping_carrier": shippingCarrier,
			"buyer_message":    buyerMessage,
			"tracking_number":  order.TrackingNo,
			"ship_by_date":     order.ShipByDate,
			"days_to_ship":     order.DaysToShip,
			"create_time":      order.CreateTime,
			"platform":         "shopee",
			"items":            items,
		})
	}

	return result, nil
}

// GetProductList fetches products
func (c *ShopeeAPIClient) GetProductList(ctx context.Context, offset, limit int) ([]map[string]interface{}, error) {
	if !c.IsInitialized() {
		return nil, fmt.Errorf("shopee client not initialized")
	}

	// TODO: Implement product list
	return []map[string]interface{}{}, nil
}
