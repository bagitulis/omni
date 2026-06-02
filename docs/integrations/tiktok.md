---
last_updated: 2026-06-02
updated_by: agent
relates_to: backend/tiktok_sdk/
stale_if_changed:
  - backend/tiktok_sdk/*.go
  - backend/internal/handler/*tiktok*
  - backend/internal/service/*tiktok*
---

# TikTok Shop Integration Guide

## Overview

OMNI integrates with TikTok Shop Open Platform for order management, product catalog, shipping, and wholesale operations. This is the most comprehensive SDK (100+ files).

## SDK Location

```
backend/tiktok_sdk/           # Comprehensive SDK (100+ files)
├── client.go                 # HTTP client, auth, signing
├── orders.go                 # Order operations
├── products.go               # Product operations
├── shipping.go               # Shipping/logistics
├── categories.go             # Category tree
├── brands.go                 # Brand management
├── warehouses.go             # Warehouse management
├── images.go                 # Image upload
└── ...                       # Many more modules
```

**Largest SDK** — always check here first for TikTok-related questions.

## Authentication Flow

```
1. User clicks "Connect TikTok" → POST /api/credentials/platforms/tiktok/connections/oauth/initiate
2. Redirect to TikTok OAuth page
3. User authorizes → TikTok redirects to callback URL
4. GET /api/credentials/callback/tiktok → Exchange code for tokens
5. Tokens stored encrypted in credential_connections (PostgreSQL, public schema)
6. App credentials (app_key, app_secret) stored in credential_app_configs
7. Select active shop (TikTok supports multiple shops per account)
```

### Credential Storage

| What | Table | Key Fields |
|------|-------|------------|
| App credentials | credential_app_configs | app_key, app_secret (Fernet-encrypted) |
| Store tokens | credential_connections | access_token, refresh_token, shop_cipher (Fernet-encrypted) |
| Audit trail | credential_audit_events | Metadata only, never secrets |

### Required Credentials

| Field | Source | Notes |
|-------|--------|-------|
| app_key | TikTok Shop Open Platform | String, per-app |
| app_secret | TikTok Shop Open Platform | Encrypted at rest |
| access_token | OAuth flow | Auto-refreshed |
| refresh_token | OAuth flow | Auto-refreshed |
| shop_cipher | TikTok-specific | Encrypted shop identifier for API calls |

## Multi-Shop Support

TikTok accounts can have multiple shops. OMNI handles this:
- `GET /api/credentials/platforms/tiktok/shops` — List available shops
- `GET /api/credentials/platforms/tiktok/active-shop` — Get current active shop
- Shop selection affects all subsequent API calls

## Key API Operations

| Operation | API Endpoint | Notes |
|-----------|--------------|-------|
| Get orders | `/api/tiktok/orders` | With filters |
| Sync orders | `/api/tiktok/sync/orders` | Full sync |
| Get products | `/api/tiktok/products` | From API |
| Sync products | `/api/tiktok/sync/products` | Full sync |
| Create product | `/api/tiktok/products` (POST) | With categories/attributes |
| Upload image | `/api/tiktok/products/upload-image` | For product images |
| Arrange shipping | `/api/tiktok/shipping/arrange` | Ship order |
| Get categories | `/api/tiktok/products/categories` | Category tree |
| Get brands | `/api/tiktok/products/brands` | Brand list |
| Wholesale MPQ | `/api/wholesale/tiktok/batch-mpq` | Minimum purchase qty |

## Product Creation Flow

```
1. Get categories → /api/tiktok/products/categories
2. Get category attributes → /api/tiktok/products/categories/:id/attributes
3. Get category rules → /api/tiktok/products/categories/:id/rules
4. Upload images → /api/tiktok/products/upload-image
5. Create product → POST /api/tiktok/products
6. (Optional) Save as draft → /api/tiktok/products/draft
7. Publish draft → /api/tiktok/products/publish/:draftId
```

## Webhook Events

| Event | Handler | Description |
|-------|---------|-------------|
| Order status update | WebhookOrderEvent | Order state change |
| Product update | WebhookProductEvent | Product modified |
| Return/refund | WebhookReturnEvent | Return created |

## Global Products

TikTok supports global products (cross-border):
- `GET /api/tiktok/products/global-products` — List global products
- `POST /api/tiktok/products/publish-global` — Publish to local market

## Official Documentation

- https://partner.tiktokshop.com/doc
