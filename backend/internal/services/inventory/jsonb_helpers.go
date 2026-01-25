package inventory

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/omni/backend/internal/models"
)

// ========== JSONB Data Helpers ==========

// GetDataMap parses JSONB data field to map
func GetDataMap(record models.InventoryRecord) map[string]interface{} {
	data := make(map[string]interface{})
	if record.Data != "" {
		json.Unmarshal([]byte(record.Data), &data)
	}
	return data
}

// GetDataString gets a string value from JSONB data
func GetDataString(record models.InventoryRecord, key string) string {
	data := GetDataMap(record)
	if val, ok := data[key]; ok {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

// GetDataFloat gets a float value from JSONB data
func GetDataFloat(record models.InventoryRecord, key string) float64 {
	data := GetDataMap(record)
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case string:
			f, _ := strconv.ParseFloat(v, 64)
			return f
		case int:
			return float64(v)
		}
	}
	return 0
}

// GetDataInt gets an int value from JSONB data
func GetDataInt(record models.InventoryRecord, key string) int {
	data := GetDataMap(record)
	if val, ok := data[key]; ok {
		switch v := val.(type) {
		case float64:
			return int(v)
		case string:
			i, _ := strconv.Atoi(v)
			return i
		case int:
			return v
		}
	}
	return 0
}

// SetDataValue sets a value in JSONB data
func SetDataValue(record *models.InventoryRecord, key string, value interface{}) error {
	data := GetDataMap(*record)
	data[key] = value
	jsonBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	record.Data = string(jsonBytes)
	return nil
}

// GetSKU gets SKU from common column names
func GetSKU(record models.InventoryRecord) string {
	data := GetDataMap(record)
	for _, key := range []string{"SKU", "sku", "Sku", "KODE", "kode", "Code", "code", "Product Code"} {
		if val, ok := data[key]; ok && val != "" {
			return fmt.Sprintf("%v", val)
		}
	}
	if record.KeyColumnName == "SKU" || record.KeyColumnName == "sku" {
		return record.KeyValue
	}
	return record.KeyValue
}

// GetProductName gets product name from common column names
func GetProductName(record models.InventoryRecord) string {
	data := GetDataMap(record)
	for _, key := range []string{"Nama Barang", "Product Name", "Name", "nama", "NAMA", "product_name"} {
		if val, ok := data[key]; ok && val != "" {
			return fmt.Sprintf("%v", val)
		}
	}
	return ""
}

// GetPrice gets price from common column names
func GetPrice(record models.InventoryRecord) float64 {
	data := GetDataMap(record)
	for _, key := range []string{"HARGA", "Harga", "harga", "Price", "price", "PRICE"} {
		if val, ok := data[key]; ok {
			switch v := val.(type) {
			case float64:
				return v
			case string:
				f, _ := strconv.ParseFloat(v, 64)
				return f
			}
		}
	}
	return 0
}

// GetQuantity gets quantity/stock from common column names
func GetQuantity(record models.InventoryRecord) int {
	data := GetDataMap(record)
	for _, key := range []string{"Stock", "stock", "STOCK", "Quantity", "quantity", "QTY", "qty", "Stok", "stok"} {
		if val, ok := data[key]; ok {
			switch v := val.(type) {
			case float64:
				return int(v)
			case string:
				i, _ := strconv.Atoi(v)
				return i
			case int:
				return v
			}
		}
	}
	return 0
}
