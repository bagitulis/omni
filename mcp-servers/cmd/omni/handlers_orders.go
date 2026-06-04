package main

import "encoding/json"

func handleGetTodayOrders() (interface{}, error) {
	data, err := omniClient.get("/api/orders/today")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetUnpaidOrders() (interface{}, error) {
	data, err := omniClient.get("/api/orders/unpaid")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetUnprocessOrders() (interface{}, error) {
	data, err := omniClient.get("/api/orders/unprocess")
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleSyncAllOrders() (interface{}, error) {
	data, err := omniClient.post("/api/orders/sync-all", nil)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}

func handleGetOrderDetail(args map[string]interface{}) (interface{}, error) {
	orderSn := args["order_sn"].(string)
	data, err := omniClient.get("/api/orders/" + orderSn)
	if err != nil {
		return nil, err
	}
	var result interface{}
	err = json.Unmarshal(data, &result)
	return result, err
}
