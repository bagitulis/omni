# Shopee SDK (Go)

Lightweight Go SDK that wraps the Shopee Open Platform v2 APIs used in this project. It reuses the signed HTTP client from `backend/pkg/shopee` and exposes typed helpers for orders, products, logistics, and wallet.

## Install

The code lives inside this repo under `backend/shopee-sdk`. Import paths follow the main module name (`github.com/omni/backend/shopee-sdk`).

```go
import shopeesdk "github.com/omni/backend/shopee-sdk"
```

## Quick start

```go
cfg := shopeesdk.Config{
    PartnerID:     123456,
    PartnerKey:    "your-partner-key",
    ShopID:        9999999,
    AccessToken:   "shop-access-token",
    UseProduction: true,
}

client, err := shopeesdk.NewClient(cfg)
if err != nil {
    log.Fatal(err)
}

ctx := context.Background()
orders, err := client.GetOrders(ctx, shopeesdk.OrderListParams{
    From:           time.Now().AddDate(0, 0, -7),
    To:             time.Now(),
    Status:         "READY_TO_SHIP",
    TimeRangeField: "create_time",
})
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Fetched %d orders\n", len(orders.Response.OrderList))
```

## Provided helpers

- Orders: list + detail with optional status filter.
- Products: list, base info, model list, images.
- Logistics: shipping parameter, tracking number, ship/cancel.
- Pricing & stock: update stock and price.
- Wallet: transactions listing.
- Media: image upload.
- Generic: `CallRaw` / `CallRawMap` to hit any v2 endpoint (ads, discount/voucher/deal/prize, bundles, shop/merchant info, returns, first-mile/pickup, finance/escrow, logistics extras, push, etc.).
- Docs-driven: `CallAPI` / `CallAPIMap` to call endpoints by official `api_name` using an embedded catalog generated from Shopee's docs API (`catalog_v2.json`).

All calls accept `context.Context`; cancellation propagates before the request.

### Call endpoints not explicitly modeled (parity with https://github.com/Faiznurullah/shopee/tree/main/src/request)

```go
resp, err := client.CallRawMap(ctx, shopeesdk.RawCall{
    Method: "POST",
    Path:   "/api/v2/discount/add_discount",
    Body: map[string]interface{}{
        "discount_name": "New Year",
        "start_time":    1735689600,
        "end_time":      1735776000,
    },
})
```

### Call by official api_name (from docs)

```go
resp, err := client.CallAPIMap(ctx, "v2.product.get_recommend_attribute", nil, map[string]interface{}{
    "item_name":    "Iphone11",
    "category_id":  14695,
    "cover_image_id": 0, // optional per docs
})
```

## References

- Upstream request/response reference repository: https://github.com/nguyenduccanh011/Web_Shopee_api_v2/tree/main
- Official Shopee Open Platform docs: https://open.shopee.com/documents?version=2
- Official docs API (scrapeable): `https://open.shopee.com/api/v1/doc/module/?version=2` and `https://open.shopee.com/api/v1/doc/api/?api_id=...`

### Regenerating the embedded catalog

From `backend/`:

```bash
go run ./cmd/gen_shopee_catalog -version 2 -out shopee-sdk/catalog_v2.json
```
