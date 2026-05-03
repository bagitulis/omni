---
last_updated: 2026-05-03
updated_by: agent
relates_to: backend/lazada-sdk/, backend/lazada_sdk/
stale_if_changed:
  - backend/lazada-sdk/*.go
  - backend/lazada_sdk/iop-sdk-go/
  - backend/internal/handler/*lazada*
  - backend/internal/service/*lazada*
---

# Lazada Integration Guide

## Overview

OMNI integrates with Lazada Open Platform for order management and product catalog operations.

## SDK Locations

```
backend/lazada-sdk/           # Custom SDK
├── client.go                 # HTTP client, auth
├── order.go                  # Order operations
├── product.go                # Product operations
└── auth.go                   # OAuth flow

backend/lazada_sdk/iop-sdk-go/  # Official IOP SDK
└── ...                         # Lazada official SDK
```

**Two SDKs exist**: Custom wrapper + official IOP SDK. Check both.

## Authentication Flow

```
1. User clicks "Connect Lazada" → GET /api/platform-auth/initiate/lazada
2. Redirect to Lazada OAuth page
3. User authorizes → Lazada redirects to callback URL
4. GET /api/platform-auth/callback/lazada → Exchange code for tokens
5. Store access_token + refresh_token (encrypted)
6. Auto-refresh before expiry
```

## Key API Operations

| Operation | API Endpoint | Notes |
|-----------|--------------|-------|
| Get orders | `/api/lazada/orders` | With filters |
| Sync orders | `/api/lazada/sync/orders` | Full sync |
| Get products | `/api/lazada/products` | From API |
| Sync products | `/api/lazada/sync/products` | Full sync |
| Ship order | `/api/lazada/orders/ship` | Arrange shipment |
| Get categories | `/api/lazada/products/categories` | For product creation |

## Webhook Events

| Event | Handler | Description |
|-------|---------|-------------|
| Order status update | WebhookOrderEvent | Order state change |
| Product update | WebhookProductEvent | Product modified |

## Official Documentation

- https://open.lazada.com/doc/api.htm
