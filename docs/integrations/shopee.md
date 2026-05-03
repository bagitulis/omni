---
last_updated: 2026-05-03
updated_by: agent
relates_to: backend/shopee-sdk/
stale_if_changed:
  - backend/shopee-sdk/*.go
  - backend/internal/handler/*shopee*
  - backend/internal/service/*shopee*
---

# Shopee Integration Guide

## Overview

OMNI integrates with Shopee Open Platform for order management, product catalog, shipping, and wallet/escrow operations.

## SDK Location

```
backend/shopee-sdk/
├── client.go          # HTTP client, auth, signing
├── orders.go          # Order API operations
├── products.go        # Product API operations
├── shipping.go        # Shipping/logistics API
├── wallet.go          # Wallet/escrow API
└── types.go           # Request/response types
```

**ALWAYS check local SDK before external docs.**

## Authentication Flow

```
1. User clicks "Connect Shopee" → GET /api/platform-auth/initiate/shopee
2. Redirect to Shopee OAuth page
3. User authorizes → Shopee redirects to callback URL
4. GET /api/platform-auth/callback/shopee → Exchange code for tokens
5. Store access_token + refresh_token (encrypted)
6. Auto-refresh before expiry
```

## Key API Operations

| Operation | SDK Method | API Endpoint |
|-----------|-----------|--------------|
| Get orders | `GetOrderList()` | `/api/shopee/orders` |
| Sync orders | `SyncOrders()` | `/api/shopee/sync/orders` |
| Get products | `GetProductList()` | `/api/shopee/products` |
| Sync products | `SyncProducts()` | `/api/shopee/sync/products` |
| Arrange shipping | `ArrangeShipment()` | `/api/shopee/shipping/arrange` |
| Get wallet | `GetWalletBalance()` | `/api/shopee/wallet/balance` |
| Get escrow | `GetEscrowDetail()` | `/api/shopee/wallet/escrow-detail` |

## Webhook Events

| Event | Handler | Description |
|-------|---------|-------------|
| `order.status_update` | WebhookOrderEvent | Order status changed |
| `product.update` | WebhookProductEvent | Product modified |
| `return.create` | WebhookReturnEvent | Return/refund created |
| `shop.authorization` | WebhookShopeeEvent | Auth status change |

## Rate Limiting

- Shopee enforces per-shop rate limits
- OMNI respects these via SDK-level throttling
- Sync operations have 300s nginx timeout for large catalogs

## Official Documentation

- https://open.shopee.com/documents
