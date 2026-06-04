package main

import "encoding/json"

// Analytics & Monitoring

func handleGetSystemHealth() (interface{}, error) {
	// Note: monitoring endpoint uses auth but NOT tenant middleware
	data, err := omniClient.getNoTenant("/api/monitoring/health/detailed")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetAnalyticsDashboard() (interface{}, error) {
	data, err := omniClient.get("/api/analytics/dashboard")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetJobStats() (interface{}, error) {
	data, err := omniClient.get("/api/jobs/stats")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

// Config & Settings

func handleGetSettings() (interface{}, error) {
	data, err := omniClient.get("/api/settings/general")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleUpdateSettings(args map[string]interface{}) (interface{}, error) {
	settings := args["settings"]
	data, err := omniClient.post("/api/settings/general", settings)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetInventoryConfig() (interface{}, error) {
	data, err := omniClient.get("/api/inventory/config")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

// Tenant & Product Management

func handleListTenants() (interface{}, error) {
	data, err := omniClient.getNoTenant("/api/tenants")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleSyncProducts() (interface{}, error) {
	data, err := omniClient.post("/api/products/sync", nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
