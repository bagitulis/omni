# Lazada SDK (Go)

Go SDK for Lazada Open Platform. This package wraps the existing signed client in `backend/pkg/lazada` and provides:

- Context-aware helpers for common endpoints (Orders, Products, Logistics, Seller, Auth, DataMoat).
- A generic `CallRaw` / `CallRawMap` to call any Lazada REST endpoint (so the SDK remains "complete" even when a typed helper doesn't exist yet).

## Install

```go
import lazadasdk "github.com/omni/backend/lazada-sdk"
```

## Quick start

```go
cfg := lazadasdk.Config{
    AppKey:       "your-app-key",
    AppSecret:    "your-app-secret",
    Region:       "id", // id|my|sg|th|vn|ph
    AccessToken:  "seller-access-token",
}

client, err := lazadasdk.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
orders, err := client.GetOrders(ctx, lazadasdk.GetOrdersParams{
    Status: "packed",
    Offset: 0,
    Limit:  50,
})
if err != nil {
    log.Fatal(err)
}
fmt.Println("orders:", orders.Data.Count)
```

## Calling any endpoint (reference parity)

```go
resp, err := client.CallRawMap(ctx, lazadasdk.RawCall{
    Method: "GET",
    Path:   "/seller/get",
})
```

## References

- Lazada Open Platform docs: https://open.lazada.com/
- Reference SDK (PHP): https://github.com/laijingwu/lazop-sdk/tree/master/src
