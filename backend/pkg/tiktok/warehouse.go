package tiktok

import (
	"log"
)

// WarehouseResponse represents TikTok warehouse list response
type WarehouseResponse struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	RequestID string `json:"request_id"`
	Data      struct {
		Warehouses []Warehouse `json:"warehouses"`
	} `json:"data"`
}

// Warehouse represents a TikTok warehouse
type Warehouse struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	WarehouseType string `json:"warehouse_type"` // SELLER_WAREHOUSE, SALES_WAREHOUSE, etc
	IsDefault    bool   `json:"is_default"`
	Address      struct {
		Region        string `json:"region"`
		State         string `json:"state"`
		City          string `json:"city"`
		District      string `json:"district"`
		Town          string `json:"town"`
		FullAddress   string `json:"full_address"`
		Zipcode       string `json:"zipcode"`
		Phone         string `json:"phone"`
		ContactPerson string `json:"contact_person"`
	} `json:"address"`
}

// GetWarehouses fetches warehouse list from TikTok API
// API: GET /logistics/202309/warehouses
func (c *Client) GetWarehouses() (*WarehouseResponse, error) {
	params := map[string]string{}

	var result WarehouseResponse
	err := c.doRequest("GET", "/logistics/202309/warehouses", params, &result)
	if err != nil {
		log.Printf("[TikTok Warehouse] API error: %v", err)
		return nil, err
	}

	log.Printf("[TikTok Warehouse] Response: code=%d, message=%s, warehouse_count=%d", 
		result.Code, result.Message, len(result.Data.Warehouses))
	
	for i, w := range result.Data.Warehouses {
		log.Printf("[TikTok Warehouse] [%d] id=%s, name=%s, type=%s, is_default=%v", 
			i, w.ID, w.Name, w.WarehouseType, w.IsDefault)
	}

	if result.Code != 0 {
		return &result, nil // Return result anyway for debugging
	}

	return &result, nil
}

// GetDefaultWarehouseID returns the default warehouse ID, or the first one if no default
func (c *Client) GetDefaultWarehouseID() (string, error) {
	resp, err := c.GetWarehouses()
	if err != nil {
		return "", err
	}

	if resp.Code != 0 {
		log.Printf("[TikTok Warehouse] API returned error code %d: %s", resp.Code, resp.Message)
		return "", nil // No warehouse, let API handle it
	}

	// Find default warehouse
	for _, w := range resp.Data.Warehouses {
		if w.IsDefault {
			log.Printf("[TikTok Warehouse] Found default warehouse: %s", w.ID)
			return w.ID, nil
		}
	}

	// Return first warehouse if no default
	if len(resp.Data.Warehouses) > 0 {
		log.Printf("[TikTok Warehouse] Using first warehouse: %s", resp.Data.Warehouses[0].ID)
		return resp.Data.Warehouses[0].ID, nil
	}

	log.Printf("[TikTok Warehouse] No warehouses found")
	return "", nil
}
