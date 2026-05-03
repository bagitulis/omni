---
last_updated: 2026-05-03
updated_by: agent
relates_to: backend/internal/models/
stale_if_changed:
  - backend/internal/models/*.go
  - backend/internal/config/migration.go
  - backend/migrations/
---

# Database Schema Documentation

## Architecture

| Aspect | Detail |
|--------|--------|
| **ORM** | GORM (Go) |
| **Production DB** | PostgreSQL 16 |
| **Development DB** | SQLite (per-tenant files) |
| **Multi-Tenant** | Schema-based isolation (`tenant_{tenantID}`) |
| **Migrations** | GORM AutoMigrate + SQL scripts |
| **JSON Support** | Custom JSONMap/JSONArray types for JSONB |

---

## Schema Layout

```
PostgreSQL Database: omni
├── Schema: public (shared)
│   ├── users
│   └── audit_logs
├── Schema: tenant_abc123
│   ├── shopee_orders, shopee_order_items
│   ├── shopee_products, shopee_skus
│   ├── lazada_orders, lazada_order_items
│   ├── lazada_products, lazada_skus
│   ├── tiktok_orders, tiktok_order_items
│   ├── tiktok_products, tiktok_skus
│   ├── master_products, master_product_skus
│   ├── master_product_platform_links
│   ├── products, product_skus
│   ├── inventory_records, inventory_settings
│   ├── jobs, job_history
│   ├── webhook_logs, webhook_*_events
│   ├── notifications, notification_settings
│   └── ... (all tenant-specific tables)
└── Schema: tenant_def456
    └── ... (same structure, isolated data)
```

---

## System Tables (Shared)

### users

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | User ID |
| username | string | unique, not null | Login username |
| email | string | unique | Email address |
| password_hash | string | not null | Bcrypt hash |
| role | string | | User role |
| tenant_id | string | not null | Primary tenant |
| is_active | bool | default true | Account active |
| failed_login_attempts | int | default 0 | Lockout counter |
| locked_until | *time | | Lockout expiry |
| last_login_at | *time | | Last login timestamp |
| oauth_provider | string | | OAuth provider (github, copilot) |
| oauth_id | string | | OAuth external ID |
| created_at | time | auto | Created timestamp |
| updated_at | time | auto | Updated timestamp |

### audit_logs

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Log ID |
| tenant_id | string | index | Tenant context |
| user_id | uint | index | Acting user |
| action | string | index | Action performed |
| resource | string | | Resource type |
| resource_id | string | | Resource identifier |
| details | JSONMap | | Action details (JSONB) |
| ip_address | string | | Client IP |
| user_agent | string | | Client user agent |
| created_at | time | auto | Timestamp |

---

## Tenant Tables — Platform Orders

### shopee_orders

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| order_sn | string | uniqueIndex | Shopee order number |
| order_status | string | index | Current status |
| buyer_username | string | | Buyer name |
| total_amount | float64 | | Order total |
| currency | string | | Currency code |
| shipping_carrier | string | | Carrier name |
| tracking_number | string | | Tracking number |
| payment_method | string | | Payment method |
| create_time | int64 | | Shopee timestamp |
| update_time | int64 | | Last update timestamp |
| raw_data | JSONMap | | Full API response |
| created_at | time | auto | DB created |
| updated_at | time | auto | DB updated |

### shopee_order_items

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| order_sn | string | index | Parent order |
| item_id | int64 | | Shopee item ID |
| item_name | string | | Product name |
| model_id | int64 | | Variant model ID |
| model_name | string | | Variant name |
| sku | string | index | SKU code |
| quantity | int | | Quantity ordered |
| price | float64 | | Item price |

### lazada_orders / lazada_order_items

Similar structure to Shopee with Lazada-specific fields (order_id instead of order_sn, etc.)

### tiktok_orders / tiktok_order_items

Similar structure with TikTok-specific fields (order_id, package_id, etc.)

---

## Tenant Tables — Products

### master_products

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| name | string | not null | Product name |
| description | string | | Product description |
| category | string | | Product category |
| brand | string | | Brand name |
| status | string | default "active" | active/inactive/draft |
| main_image | string | | Primary image URL |
| images | JSONArray | | Additional images |
| attributes | JSONMap | | Custom attributes |
| created_at | time | auto | Created |
| updated_at | time | auto | Updated |

### master_product_skus

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| master_product_id | uint | FK, index | Parent product |
| sku | string | uniqueIndex | SKU code |
| name | string | | Variant name |
| price | float64 | | Base price |
| stock | int | | Current stock |
| weight | float64 | | Weight (grams) |
| dimensions | JSONMap | | L x W x H |
| attributes | JSONMap | | Variant attributes |
| status | string | default "active" | Status |

**Constraint**: Max 50 SKUs per master product.

### master_product_platform_links

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| master_product_id | uint | FK, index | Master product |
| master_sku_id | uint | FK | Master SKU |
| platform | string | index | shopee/lazada/tiktok |
| platform_product_id | string | | Platform item ID |
| platform_sku_id | string | | Platform SKU ID |
| platform_sku | string | | Platform SKU code |
| sync_status | string | | synced/pending/error |
| last_synced_at | *time | | Last sync time |

### shopee_products / lazada_products / tiktok_products

Platform-specific product tables with full API data stored in `raw_data` JSONMap column.

### shopee_skus / lazada_skus / tiktok_skus

Platform-specific SKU/variant tables linked to their parent product.

---

## Tenant Tables — Inventory

### inventory_records

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| key_column | string | index | Primary key column name |
| key_value | string | uniqueIndex | Primary key value (usually SKU) |
| data | JSONMap | | Dynamic fields from spreadsheet |
| sync_status | string | | synced/pending/conflict |
| last_synced_at | *time | | Last sync time |
| source | string | | Data source (sheets/manual) |
| created_at | time | auto | Created |
| updated_at | time | auto | Updated |

### inventory_settings

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| spreadsheet_id | string | | Google Sheets ID |
| sheet_name | string | | Sheet tab name |
| key_column | string | | Primary key column |
| sync_direction | string | | bidirectional/to-sheets/from-sheets |
| auto_sync | bool | | Enable auto-sync |
| sync_interval | int | | Sync interval (minutes) |

### inventory_sku_platform_status

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| sku | string | index | SKU code |
| platform | string | index | Platform name |
| item_id | string | | Platform item ID |
| model_id | string | | Platform model ID |
| status | string | | active/inactive/not_found |
| stock | int | | Platform stock level |
| price | float64 | | Platform price |
| checked_at | time | | Last check time |

---

## Tenant Tables — Webhooks

### webhook_logs

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| platform | string | index | Source platform |
| event_type | string | index | Event type |
| payload | JSONMap | | Raw webhook payload |
| status | string | | processed/failed/ignored |
| error | string | | Error message if failed |
| processed_at | *time | | Processing time |
| created_at | time | auto | Received time |

### webhook_order_events / webhook_product_events / webhook_return_events

Specialized event tables for different webhook types with platform-specific fields.

---

## Tenant Tables — Jobs & Automation

### jobs

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | string | PK | Job UUID |
| type | string | index | Job type |
| status | string | index | pending/running/completed/failed/cancelled |
| payload | JSONMap | | Job parameters |
| result | JSONMap | | Job result |
| error | string | | Error message |
| attempts | int | default 0 | Retry count |
| max_attempts | int | default 3 | Max retries |
| started_at | *time | | Start time |
| completed_at | *time | | Completion time |
| created_at | time | auto | Created |

### auto_functions_config

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| name | string | uniqueIndex | Function name |
| description | string | | Description |
| enabled | bool | default false | Active status |
| schedule | string | | Cron expression |
| config | JSONMap | | Function configuration |
| last_run_at | *time | | Last execution |
| next_run_at | *time | | Next scheduled run |

---

## Tenant Tables — Analytics

### shopee_escrow_orders / tiktok_escrow_orders

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| order_sn | string | uniqueIndex | Order number |
| escrow_amount | float64 | | Escrow amount |
| buyer_total | float64 | | Buyer paid total |
| commission | float64 | | Platform commission |
| service_fee | float64 | | Service fee |
| shipping_fee | float64 | | Shipping fee |
| release_date | *time | | Escrow release date |
| sync_month | string | index | YYYY-MM format |

---

## Tenant Tables — Notifications

### notifications

| Column | Type | Constraints | Description |
|--------|------|-------------|-------------|
| id | uint | PK, auto | Internal ID |
| type | string | index | Notification type |
| title | string | | Title |
| message | string | | Message body |
| data | JSONMap | | Additional data |
| read | bool | default false | Read status |
| read_at | *time | | Read timestamp |
| created_at | time | auto | Created |

---

## Key Relationships

```
MasterProduct (1) ──── (N) MasterProductSku
       │
       └──── (N) MasterProductPlatformLink ────── ShopeeProduct
                                            ────── LazadaProduct
                                            ────── TiktokProduct

ShopeeOrder (1) ──── (N) ShopeeOrderItem
LazadaOrder (1) ──── (N) LazadaOrderItem
TiktokOrder (1) ──── (N) TiktokOrderItem

ShopeeProduct (1) ──── (N) ShopeeSku
LazadaProduct (1) ──── (N) LazadaSku
TiktokProduct (1) ──── (N) TiktokSku
```

---

## Multi-Tenancy Rules

1. **No default tenant**: Missing `tenant_id` = hard error (never fallback)
2. **Schema isolation**: `SET search_path TO tenant_{id}` per connection
3. **Connection pooling**: PgBouncer with transaction-mode pooling
4. **System tables**: Users and audit logs in shared schema
5. **Tenant creation**: Auto-creates schema + runs migrations

---

## Migration Strategy

- **Primary**: GORM `AutoMigrate()` in `backend/internal/config/migration.go`
- **Complex changes**: SQL scripts in `backend/migrations/`
- **Idempotent**: All migrations use `IF NOT EXISTS`
- **After schema change**: MUST run `python build.py backup`

---

## Custom Types

```go
// JSONMap — stores arbitrary JSON objects (PostgreSQL JSONB)
type JSONMap map[string]interface{}

// JSONArray — stores JSON arrays (PostgreSQL JSONB)
type JSONArray []interface{}
```

Used for dynamic/flexible data (raw API responses, custom attributes, etc.)
