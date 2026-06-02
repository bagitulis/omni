---
last_updated: 2026-06-02
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
1. User clicks "Connect Shopee" → POST /api/credentials/platforms/shopee/connections/oauth/initiate
2. Redirect to Shopee OAuth page
3. User authorizes → Shopee redirects to callback URL
4. GET /api/credentials/callback/shopee → Exchange code for tokens
5. Tokens stored encrypted in credential_connections (PostgreSQL, public schema)
6. App credentials (partner_id, partner_key) stored in credential_app_configs
7. Auto-refresh before expiry via TokenManager
```

### Credential Storage

| What | Table | Key Fields |
|------|-------|------------|
| App credentials | credential_app_configs | partner_id, partner_key (Fernet-encrypted) |
| Store tokens | credential_connections | access_token, refresh_token (Fernet-encrypted) |
| Audit trail | credential_audit_events | Metadata only, never secrets |

### Required Credentials

| Field | Source | Notes |
|-------|--------|-------|
| partner_id | Shopee Open Platform | Integer, per-app |
| partner_key | Shopee Open Platform | Encrypted at rest |
| access_token | OAuth flow | Auto-refreshed, ~2 hour lifetime |
| refresh_token | OAuth flow | Auto-refreshed, 25 days 5 hours lifetime |

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
