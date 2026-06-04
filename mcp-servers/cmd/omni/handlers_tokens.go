package main

import "encoding/json"

func handleGetTokenStatus() (interface{}, error) {
	data, err := omniClient.get("/api/tokens/status")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetPlatformTokenStatus(args map[string]interface{}) (interface{}, error) {
	platform := args["platform"].(string)
	data, err := omniClient.get("/api/tokens/status/" + platform)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleRefreshToken(args map[string]interface{}) (interface{}, error) {
	platform := args["platform"].(string)
	data, err := omniClient.post("/api/tokens/refresh/"+platform, nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleRefreshAllTokens() (interface{}, error) {
	data, err := omniClient.post("/api/tokens/refresh-all", nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
