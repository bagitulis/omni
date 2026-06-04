package main

import (
	"encoding/json"
	"fmt"
)

// Config Update Handlers

func handleUpdateInventorySettings(args map[string]interface{}) (interface{}, error) {
	settings := args["settings"]
	data, err := omniClient.put("/api/settings/inventory", settings)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetAnalyticsSettings() (interface{}, error) {
	data, err := omniClient.get("/api/analytics/settings")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleUpdateAnalyticsSettings(args map[string]interface{}) (interface{}, error) {
	settings := args["settings"]
	data, err := omniClient.put("/api/analytics/settings", settings)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleUpdateAutoFunction(args map[string]interface{}) (interface{}, error) {
	name := args["name"].(string)
	settings := args["settings"]
	path := fmt.Sprintf("/api/jobs/auto-functions/%s", name)
	data, err := omniClient.put(path, settings)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetGoogleSheetsSettings() (interface{}, error) {
	data, err := omniClient.get("/api/settings/detailed")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleUpdateGoogleSheetsSettings(args map[string]interface{}) (interface{}, error) {
	settings := args["settings"]
	data, err := omniClient.post("/api/settings/update-detailed", settings)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleUpdateInventoryConfig(args map[string]interface{}) (interface{}, error) {
	settings := args["settings"]
	data, err := omniClient.put("/api/inventory/config", settings)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
