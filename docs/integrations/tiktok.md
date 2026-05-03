---
last_updated: 2026-05-03
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
1. User clicks "Connect TikTok" → GET /api/platform-auth/initiate/tiktok
2. Redirect to TikTok OAuth page
3. User authorizes → TikTok redirects to callback URL
4. GET /api/platform-auth/callback/tiktok → Exchange code for tokens
5. Store access_token + refresh_token (encrypted)
6. Select active shop (TikTok supports multiple shops per account)
```

## Multi-Shop Support

TikTok accounts can have multiple shops. OMNI handles this:
- `GET /api/platform-auth/tiktok/shops` — List available shops
- `GET /api/platform-auth/tiktok/active-shop` — Get current active shop
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
