---
last_updated: 2026-06-02
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
1. User clicks "Connect Lazada" → POST /api/credentials/platforms/lazada/connections/oauth/initiate
2. Redirect to Lazada OAuth page
3. User authorizes → Lazada redirects to callback URL
4. GET /api/credentials/callback/lazada → Exchange code for tokens
5. Tokens stored encrypted in credential_connections (PostgreSQL, public schema)
6. App credentials (app_key, app_secret) stored in credential_app_configs
7. Auto-refresh before expiry via TokenManager
```

### Credential Storage

| What | Table | Key Fields |
|------|-------|------------|
| App credentials | credential_app_configs | app_key, app_secret (Fernet-encrypted) |
| Store tokens | credential_connections | access_token, refresh_token (Fernet-encrypted) |
| Audit trail | credential_audit_events | Metadata only, never secrets |

### Required Credentials

| Field | Source | Notes |
|-------|--------|-------|
| app_key | Lazada Open Platform | String, per-app |
| app_secret | Lazada Open Platform | Encrypted at rest |
| access_token | OAuth flow | Auto-refreshed |
| refresh_token | OAuth flow | Auto-refreshed |

### Connection Row Note

Lazada does **not** create a `credential_connections` row during OAuth. Unlike Shopee and TikTok, Lazada doesn't expose a `shopId` or `storeIdentifier` via its OAuth flow. The access/refresh tokens are stored, but the connection record uses an empty store identifier. This is expected behavior, not a bug.

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
