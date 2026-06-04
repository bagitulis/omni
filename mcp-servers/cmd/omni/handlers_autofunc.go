package main

import (
	"encoding/json"
	"fmt"
)

func handleGetAutoFunctions() (interface{}, error) {
	data, err := omniClient.get("/api/jobs/auto-functions")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleToggleAutoFunction(args map[string]interface{}) (interface{}, error) {
	name := args["name"].(string)
	enabled := args["enabled"].(bool)
	action := "enable"
	if !enabled {
		action = "disable"
	}
	path := fmt.Sprintf("/api/jobs/auto-functions/%s/%s", name, action)
	data, err := omniClient.post(path, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleRunAutoFunction(args map[string]interface{}) (interface{}, error) {
	name := args["name"].(string)
	data, err := omniClient.post(fmt.Sprintf("/api/jobs/auto-functions/%s/run", name), nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetAutoFunctionHistory() (interface{}, error) {
	data, err := omniClient.get("/api/jobs/auto-functions/history")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleDisconnectStore(args map[string]interface{}) (interface{}, error) {
	platform := args["platform"].(string)
	store := args["store_identifier"].(string)
	path := fmt.Sprintf("/api/credentials/platforms/%s/connections/%s/disconnect", platform, store)
	data, err := omniClient.post(path, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleRefreshStoreCredential(args map[string]interface{}) (interface{}, error) {
	platform := args["platform"].(string)
	store := args["store_identifier"].(string)
	path := fmt.Sprintf("/api/credentials/platforms/%s/connections/%s/refresh", platform, store)
	data, err := omniClient.post(path, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
