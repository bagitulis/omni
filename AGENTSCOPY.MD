# ⚠️ PENTING: BACA SEBELUM MELAKUKAN PERUBAHAN

**JANGAN EDIT FILE INI TANPA IZIN EKSPLISIT. PELAJARI FILE INI DENGAN SAKSAMA SEBELUM MEMBUAT PERUBAHAN APAPUN.**

---

## Panduan Pengembangan untuk AI Agent

> [!CAUTION]
> Proyek ini dalam **FASE MIGRASI AKTIF**:
>
> - **Backend:** Node.js (TypeScript/Prisma) → **Golang (GORM)**
> - **Database:** SQLite → **PostgreSQL 16**
>
> Semua kode baru **WAJIB** ditulis di Go backend (`backend/`). Jangan edit `backend-node/` kecuali untuk referensi.

---

## 📑 Daftar Isi

0. [Panduan Kilat AI (Go + Omni-PostgreSQL)](#0-panduan-kilat-ai-go--omni-postgresql)
1. [Standar Kode](#1-standar-kode)
2. [Database & Multi-Tenancy](#3-database--multi-tenancy)
3. [Model & Schema Database](#4-model--schema-database)
4. [Pola Penulisan Kode Go](#5-pola-penulisan-kode-go)
5. [Error Handling & Logging](#6-error-handling--logging)
6. [Testing](#7-testing)
7. [Security](#8-security)
8. [Docker & Deployment](#9-docker--deployment)
9. [Frontend Integration](#10-frontend-integration)
10. [Workflow untuk AI Agent](#11-workflow-untuk-ai-agent)

---

## 0. Panduan Kilat AI (Go + Omni-PostgreSQL)

> [!IMPORTANT]
> Fokus backend Golang + Omni-PostgreSQL. Semua instruksi di bawah **WAJIB diikuti berurutan** dan menjadi baseline setiap task.

### 0.1 Prinsip Wajib

- Clean Code, DRY, OOP, SRP; tidak ada duplicate/dead/unused code.
- Batas file maksimal 300 baris (pecah ke file baru jika perlu).
- Perbaiki semua error, warning, dan lint issues sebelum selesai.
- Multi-tenant: tidak ada default tenant; selalu pakai `context.Context`.
- Penamaan: DB kolom `snake_case`, JSON response `snake_case`, Go field `PascalCase`; jangan pakai alias untuk nama salah.

### 0.2 ToDo List Implementasi Detail (Urutan Wajib)

1. Analisis kode & database
   - Baca referensi `backend-node/` + SQLite schema untuk mapping ke Go/GORM.
   - Telusuri flow TikTok/Lazada SDK Golang dan panggilan route dari frontend.
   - Audit `backend/internal/models/`, `table_naming.go`, dan migration `scripts/postgres/` agar konsisten dengan Omni-PostgreSQL.
2. Perbaikan masalah fungsional
   - Identifikasi bug, nil pointer, race, tenant leak, atau query salah schema.
   - Betulkan snake_case/camelCase pada JSON/DB tag dan TableName() mapping.
   - Hapus duplikasi/unused; pastikan SRP di handler/service/repo.
3. Build & test dasar
   - `cd backend && go build ./...`
   - `go test ./...` (tambah/adjust test bila perlu).
   - Jalankan lint statis jika tersedia (golangci-lint, staticcheck).
4. Perbaikan lanjut
   - Tindak semua error/warning build/test; lengkapi migrasi SQL bila ada perubahan schema.
   - Pastikan logging terstruktur dan tidak ada data sensitif.
5. Cleanup akhir (SRP/DRY/OOP)
   - Refaktor agar fungsi tunggal tujuan, hilangkan duplikasi, pecah file >300 baris.
   - Pastikan handler tidak berisi business logic; gunakan service/repo.
6. Testing final
   - Ulangi `go test ./...` + sanity curl ke endpoint terkait (snake_case response).
   - Jika relevan, `docker compose up -d --build backend-go` lalu cek log aplikasi.
7. Iterasi
   - Ulangi siklus sampai semua langkah bersih tanpa warning/error.

### 0.3 Kriteria Keberhasilan (Butuh 2 Bukti)

- Bukti 1: Log Docker menampilkan data valid (cek timestamp/tanggal).
- Bukti 2: Script/command menunjukkan keterangan keberhasilan.
- Jika hanya ada 1 bukti, berhenti dan konfirmasi.

### 0.4 Rujukan Teknis Wajib

- Selalu baca `backend-node/` untuk logic legacy dan mapping SQLite → PostgreSQL.
- Cek SDK TikTok/Lazada Golang yang digunakan sebelum mengubah flow.
- Pelajari alur pemanggilan route dari frontend agar response tetap backward compatible.
- Validasi ulang snake_case vs camelCase di setiap DTO/JSON tag sebelum merge.
- Dilarang caching; semua read langsung ke DB tenant yang benar.

---

## 1. Standar Kode

### 1.1 Aturan Umum

| Aturan             | Penjelasan                                                          |
| ------------------ | ------------------------------------------------------------------- |
| **PascalCase**     | Struct, exported functions, methods                                 |
| **camelCase**      | Unexported functions, local variables                               |
| **snake_case**     | Database columns (PostgreSQL), JSON tags API response               |
| **CLEAN Code**     | Kode harus bersih, mudah dibaca, dan self-documenting               |
| **DRY**            | Don't Repeat Yourself - hindari duplikasi kode                      |
| **SRP**            | Single Responsibility Principle - satu fungsi, satu tujuan          |
| **Max 300 baris**  | Setiap file maksimal 300 baris, pecah jika lebih                    |
| **Context selalu** | Semua function yang akses DB/network WAJIB terima `context.Context` |

### 1.2 JSON Naming Convention

> [!IMPORTANT]
> **Semua JSON response API WAJIB menggunakan `snake_case`** - tanpa exception.

```go
type Order struct {
    OrderSN       string    `json:"order_sn"`       // ✅ snake_case
    TotalAmount   float64   `json:"total_amount"`   // ✅ snake_case
    BuyerUsername string    `json:"buyer_username"` // ✅ snake_case
    CreatedAt     time.Time `json:"created_at"`     // ✅ snake_case
}

type AutoFunctionConfig struct {
    ID              uint       `json:"id"`
    Name            string     `json:"name"`
    Enabled         bool       `json:"enabled"`
    IntervalMinutes int        `json:"interval_minutes"` // ✅ snake_case
    StartTime       *string    `json:"start_time"`       // ✅ snake_case
    EndTime         *string    `json:"end_time"`         // ✅ snake_case
    LastExecuted    *time.Time `json:"last_executed"`    // ✅ snake_case
    CreatedAt       time.Time  `json:"created_at"`       // ✅ snake_case
    UpdatedAt       time.Time  `json:"updated_at"`       // ✅ snake_case
}
```

> [!IMPORTANT]
> **Jangan** campur business logic di handler. Handler hanya untuk:
>
> 1. Parse request
> 2. Validasi input
> 3. Call service
> 4. Format response

---

## 3. Database & Multi-Tenancy

### 3.1 ⚠️ ATURAN MUTLAK DATABASE

pakai VARCHAR(255) supaya lebih kompatibel dengan berbagai tipe data di SQLite lama.

> [!CAUTION]
>
> ```
> ❌ JANGAN PERNAH:
> ├── Reset/drop database production
> ├── Tambah database baru tanpa izin
> ├── Edit scripts/postgres/*.sql tanpa izin
> ├── Hapus atau rename tenant schema tanpa izin (OTOMATIS DIBERI IZIN KETIKA MEMPERBAIKI FUNGSI YANG SALAH)
> ├── TIDAK ADA DEFAULT TENANT (lihat 3.4)
> ├── Raw SQL untuk DDL tanpa migration script
> └── Gunakan CACHE dalam bentuk apapun (lihat 3.6)
> ```

### 3.2 Strategi Multi-Tenancy: Schema per Tenant

**PostgreSQL menggunakan SCHEMA SEPARATION:**

```
omni_main (database)
├── system (schema)           → Global config, tenant registry
├── tenant_yumna_bertigamart  → Tenant: Bertigamart
└── tenant_tika_nusseyba      → Tenant: Nusseyba Shop
```

**SQL untuk set tenant context:**

```sql
SELECT set_tenant_context('yumna_bertigamart');
-- Ini mengubah search_path ke tenant_yumna_bertigamart
```

### 3.3 Akses Database di Go

```go
// ✅ BENAR - Dengan tenant context
func (s *OrderService) GetOrders(ctx context.Context, tenantID string) ([]models.ShopeeOrder, error) {
    db := s.dbManager.GetTenantDB(tenantID)

    var orders []models.ShopeeOrder
    err := db.WithContext(ctx).Find(&orders).Error
    return orders, err
}

// ✅ BENAR - Global config (system schema)
func (s *ConfigService) GetGlobalConfig(ctx context.Context, platform, key string) (*models.GlobalConfig, error) {
    db := s.dbManager.GetSystemDB()

    var config models.GlobalConfig
    err := db.WithContext(ctx).
        Where("platform = ? AND config_key = ?", platform, key).
        First(&config).Error
    return &config, err
}

// ❌ SALAH - Hardcode tenant
func (s *OrderService) GetOrders(ctx context.Context) ([]models.ShopeeOrder, error) {
    db := s.dbManager.GetTenantDB("yumna_bertigamart") // NEVER hardcode!
    // ...
}
```

### 3.4 ⛔ TIDAK ADA DEFAULT TENANT

**Jika `tenantID` kosong, WAJIB throw error:**

```go
// ❌ SALAH - Menyembunyikan bug!
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
    tenantID := c.Locals("tenant_id")
    if tenantID == nil || tenantID == "" {
        tenantID = "default" // NEVER DO THIS!
    }
}

// ✅ BENAR - Explicit error
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
    tenantID, ok := c.Locals("tenant_id").(string)
    if !ok || tenantID == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "error":   "Missing tenant_id - authentication required",
        })
    }
    // proceed...
}
```

### 3.5 Tenant Registry

| Tenant ID           | Schema                     | Shop Name     |
| ------------------- | -------------------------- | ------------- |
| `yumna_bertigamart` | `tenant_yumna_bertigamart` | Bertigamart   |
| `tika_nusseyba`     | `tenant_tika_nusseyba`     | Nusseyba Shop |

**Format tenant ID:** `{owner}_{shopname}` (lowercase, underscore)

### 3.6 ⛔ DILARANG: Pengelolaan Cache

> [!CAUTION]
> **JANGAN gunakan cache dalam bentuk APAPUN!**
>
> ```
> ❌ DILARANG:
> ├── Redis / Memcached
> ├── In-memory cache (sync.Map, map dengan mutex)
> ├── Local file cache
> ├── Query result caching
> ├── Session cache
> └── Third-party caching libraries (go-cache, bigcache, dll)
> ```

**Alasan:**

1. **Kompleksitas Multi-Tenant** - Cache invalidation antar tenant sangat rumit
2. **Konsistensi Data** - Risiko data stale antara cache dan database
3. **Memory Overhead** - Sistem multi-tenant akan mengkonsumsi memory berlebih
4. **Debug Difficulty** - Bug cache sangat sulit di-trace
5. **Skala Kecil** - Untuk skala saat ini, PostgreSQL sudah cukup cepat

**Alternatif yang DIIZINKAN:**

```go
// ✅ BENAR - Query langsung ke database
func (s *ProductService) GetProduct(ctx context.Context, tenantID, productID string) (*models.Product, error) {
    db := s.dbManager.GetTenantDB(tenantID)
    var product models.Product
    err := db.WithContext(ctx).Where("id = ?", productID).First(&product).Error
    return &product, err
}

// ❌ SALAH - Menggunakan cache
func (s *ProductService) GetProduct(ctx context.Context, tenantID, productID string) (*models.Product, error) {
    cacheKey := fmt.Sprintf("%s:product:%s", tenantID, productID)
    if cached, ok := s.cache.Get(cacheKey); ok {  // JANGAN!
        return cached.(*models.Product), nil
    }
    // ...
}
```

> [!TIP]
> Jika performa menjadi masalah di masa depan, diskusikan dulu dengan tim sebelum implementasi caching.

---

## 4. Model & Schema Database

### 4.1 Pemetaan Tipe SQLite → PostgreSQL

| SQLite          | PostgreSQL            | GORM Tag                       |
| --------------- | --------------------- | ------------------------------ |
| `INTEGER`       | `INTEGER` / `BIGINT`  | `gorm:"type:integer"`          |
| `TEXT`          | `TEXT` / `VARCHAR(n)` | `gorm:"type:text"`             |
| `REAL`          | `DOUBLE PRECISION`    | `gorm:"type:double precision"` |
| `BLOB`          | `BYTEA`               | `gorm:"type:bytea"`            |
| `DATETIME`      | `TIMESTAMPTZ`         | `gorm:"type:timestamptz"`      |
| `BOOLEAN (0/1)` | `BOOLEAN`             | `gorm:"type:boolean"`          |
| `JSON (TEXT)`   | `JSONB`               | `gorm:"type:jsonb"`            |

### 4.2 Pemetaan Nama Tabel

**Prisma (PascalCase) → PostgreSQL (snake_case):**

Gunakan `table_naming.go` untuk mapping:

```go
// backend/internal/models/table_naming.go

var tableNameMapping = map[string]string{
    "User":              "users",
    "ShopeeOrder":       "shopee_orders",
    "ShopeeOrderItem":   "shopee_order_items",
    "ShopeeProduct":     "shopee_products",
    "ShopeeSku":         "shopee_skus",
    // ... lengkap di file
}

// Model WAJIB implement TableName():
func (ShopeeOrder) TableName() string {
    return GetTableName("ShopeeOrder") // Returns "shopee_orders" for PostgreSQL
}
```

### 4.3 Pola Definisi Model

**Template untuk setiap model:**

```go
package models

import "time"

// ShopeeOrder represents a Shopee order in database
// Migration: backend-node/prisma/schema.prisma → ShopeeOrder model
type ShopeeOrder struct {
    ID              uint       `gorm:"primaryKey" json:"id"`
    TenantID        string     `gorm:"column:tenant_id;index;not null" json:"tenant_id"`
    OrderSN         string     `gorm:"column:order_sn;uniqueIndex;not null" json:"order_sn"`
    ShopID          *int64     `gorm:"column:shop_id" json:"shop_id,omitempty"`
    OrderStatus     string     `gorm:"column:order_status" json:"order_status"`

    // Timestamps - ALWAYS use these patterns
    CreatedAt       time.Time  `gorm:"column:created_at" json:"created_at"`
    UpdatedAt       time.Time  `gorm:"column:updated_at" json:"updated_at"`
}

// TableName specifies PostgreSQL table name
func (ShopeeOrder) TableName() string {
    return GetTableName("ShopeeOrder")
}
```

### 4.4 Checklist Model Baru

Saat membuat model baru, pastikan:

- [ ] Struct fields menggunakan PascalCase
- [ ] GORM tags dengan `column:snake_case`
- [ ] JSON tags dengan snake_case (response compatibility)
- [ ] `TenantID` field ada untuk tenant-scoped tables
- [ ] Implement `TableName()` method
- [ ] Tambah mapping di `table_naming.go`
- [ ] Buat migration SQL di `scripts/postgres/`
- [ ] Index pada columns yang sering di-query

### 4.5 Tabel yang Valid per Schema (dari omni-postgres)

**📁 `tenant_{name}` (56 tabel - Data Tenant):**

| Kategori             | Tabel                                                                                                                                                                                          |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Users & Config**   | `users`, `platform_configs`, `google_sheets_settings`                                                                                                                                          |
| **Shopee Orders**    | `shopee_orders`, `shopee_order_items`                                                                                                                                                          |
| **Shopee Products**  | `shopee_products`, `shopee_skus`                                                                                                                                                               |
| **Shopee Escrow**    | `shopee_escrow_sync`, `shopee_escrow_orders`, `shopee_escrow_items`                                                                                                                            |
| **Shopee Ads**       | `shopee_ads_upload_batches`, `shopee_ads_product_data`                                                                                                                                         |
| **Lazada Orders**    | `lazada_orders`, `lazada_order_items`                                                                                                                                                          |
| **Lazada Products**  | `lazada_products`, `lazada_skus`                                                                                                                                                               |
| **TikTok Orders**    | `tiktok_orders`, `tiktok_order_items`                                                                                                                                                          |
| **TikTok Products**  | `tiktok_products`, `tiktok_skus`                                                                                                                                                               |
| **TikTok Escrow**    | `tiktok_escrow_sync`, `tiktok_escrow_orders`, `tiktok_escrow_items`                                                                                                                            |
| **TikTok Ads**       | `tiktok_ads_upload_batches`, `tiktok_ads_creative_data`, `tiktok_ads_product_summaries`, `tiktok_ads_ml_predictions`                                                                           |
| **Unified Products** | `products`, `product_skus`                                                                                                                                                                     |
| **Inventory**        | `inventory_settings`, `inventory_records`, `inventory_sync_history`, `inventory_sku_platform_status`, `sheet_snapshots`                                                                        |
| **Orders Today**     | `order_today_items`, `locked_orders`                                                                                                                                                           |
| **Settings**         | `filter_preferences`, `wholesale_settings`, `analytics_settings`, `spreadsheets`, `route_configs`, `route_execution_config`                                                                    |
| **OAuth**            | `oauth_states`, `oauth_logs`                                                                                                                                                                   |
| **Webhooks**         | `webhook_logs`, `webhook_order_events`, `webhook_product_events`, `webhook_return_events`, `webhook_marketing_events`, `webhook_shopee_events`, `webhook_webchat_events`, `webhook_fbs_events` |
| **Jobs**             | `jobs`, `job_history`, `auto_functions_config`, `auto_functions_history`                                                                                                                       |

**📁 `system` (8 tabel - Global):**

| Tabel           | Deskripsi                          |
| --------------- | ---------------------------------- |
| `global_config` | Konfigurasi global (snake_case) ✅ |
| `GlobalConfig`  | Legacy - akan dihapus              |
| `tenants`       | Registry semua tenant              |
| `audit_log`     | Audit log (snake_case) ✅          |
| `audit_logs`    | Alias - gunakan `audit_log`        |
| `AuditLog`      | Legacy - akan dihapus              |
| `users`         | System users (snake_case) ✅       |
| `User`          | Legacy - akan dihapus              |

> [!WARNING]
> **Tabel Legacy di system schema:**
>
> - `GlobalConfig`, `AuditLog`, `User` → Format PascalCase (legacy dari SQLite/Prisma)
> - `audit_logs` → Alias yang tidak standar
>
> **Gunakan versi snake_case:** `global_config`, `audit_log`, `users`, `tenants`

---

## 5. Pola Penulisan Kode Go

### 5.1 Handler Pattern

```go
// internal/handlers/order_handler.go
package handlers

import (
    "github.com/gofiber/fiber/v2"
    "github.com/omni/backend/internal/services"
)

type OrderHandler struct {
    orderService *services.OrderService
}

func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
    return &OrderHandler{orderService: orderService}
}

// GetOrders godoc
// @Summary Get all orders
// @Tags Orders
// @Accept json
// @Produce json
// @Param platform path string true "Platform (shopee/lazada/tiktok)"
// @Success 200 {object} dto.OrderListResponse
// @Router /api/orders/{platform} [get]
func (h *OrderHandler) GetOrders(c *fiber.Ctx) error {
    tenantID, ok := c.Locals("tenant_id").(string)
    if !ok || tenantID == "" {
        return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
            "success": false,
            "error":   "Missing tenant_id",
        })
    }

    platform := c.Params("platform")
    page := c.QueryInt("page", 1)
    limit := c.QueryInt("limit", 20)

    orders, total, err := h.orderService.GetOrders(c.Context(), tenantID, platform, page, limit)
    if err != nil {
        return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "success": false,
            "error":   err.Error(),
        })
    }

    return c.JSON(fiber.Map{
        "success": true,
        "data":    orders,
        "meta": fiber.Map{
            "total": total,
            "page":  page,
            "limit": limit,
        },
    })
}
```

### 5.2 Service Pattern

```go
// internal/services/order_service.go
package services

import (
    "context"

    "github.com/omni/backend/internal/models"
    "github.com/omni/backend/internal/repositories"
)

type OrderService struct {
    shopeeRepo  *repositories.ShopeeOrderRepository
    lazadaRepo  *repositories.LazadaOrderRepository
    tiktokRepo  *repositories.TiktokOrderRepository
    dbManager   *DatabaseManager
}

func NewOrderService(dbManager *DatabaseManager) *OrderService {
    return &OrderService{
        dbManager: dbManager,
        // repositories initialized per-request with tenant DB
    }
}

func (s *OrderService) GetOrders(ctx context.Context, tenantID, platform string, page, limit int) ([]interface{}, int64, error) {
    db := s.dbManager.GetTenantDB(tenantID)

    switch platform {
    case "shopee":
        repo := repositories.NewShopeeOrderRepository(db)
        return toInterfaceSlice(repo.FindAll(ctx, page, limit))
    case "lazada":
        repo := repositories.NewLazadaOrderRepository(db)
        return toInterfaceSlice(repo.FindAll(ctx, page, limit))
    case "tiktok":
        repo := repositories.NewTiktokOrderRepository(db)
        return toInterfaceSlice(repo.FindAll(ctx, page, limit))
    default:
        return nil, 0, errors.New("unknown platform")
    }
}
```

### 5.3 Repository Pattern

```go
// internal/repositories/shopee_order_repository.go
package repositories

import (
    "context"

    "github.com/omni/backend/internal/models"
    "gorm.io/gorm"
)

type ShopeeOrderRepository struct {
    db *gorm.DB
}

func NewShopeeOrderRepository(db *gorm.DB) *ShopeeOrderRepository {
    return &ShopeeOrderRepository{db: db}
}

// FindAll returns paginated orders for tenant
func (r *ShopeeOrderRepository) FindAll(ctx context.Context, page, pageSize int) ([]models.ShopeeOrder, int64, error) {
    var orders []models.ShopeeOrder
    var total int64

    // Count first
    if err := r.db.WithContext(ctx).Model(&models.ShopeeOrder{}).Count(&total).Error; err != nil {
        return nil, 0, err
    }

    // Then paginate
    offset := (page - 1) * pageSize
    err := r.db.WithContext(ctx).
        Order("created_at DESC").
        Offset(offset).
        Limit(pageSize).
        Find(&orders).Error

    return orders, total, err
}

// FindByOrderSN finds order by serial number
func (r *ShopeeOrderRepository) FindByOrderSN(ctx context.Context, orderSN string) (*models.ShopeeOrder, error) {
    var order models.ShopeeOrder
    err := r.db.WithContext(ctx).Where("order_sn = ?", orderSN).First(&order).Error
    if err == gorm.ErrRecordNotFound {
        return nil, nil
    }
    return &order, err
}

// Upsert creates or updates order
func (r *ShopeeOrderRepository) Upsert(ctx context.Context, order *models.ShopeeOrder) error {
    return r.db.WithContext(ctx).
        Where("order_sn = ?", order.OrderSN).
        Assign(*order).
        FirstOrCreate(order).Error
}
```

### 5.4 DTO Pattern (Request/Response)

```go
// internal/dto/order_dto.go
package dto

import "time"

// OrderListRequest untuk pagination
type OrderListRequest struct {
    Page     int    `query:"page" validate:"min=1"`
    Limit    int    `query:"limit" validate:"min=1,max=100"`
    Status   string `query:"status"`
    Platform string `params:"platform" validate:"required,oneof=shopee lazada tiktok"`
}

// OrderResponse untuk API response
type OrderResponse struct {
    ID          uint      `json:"id"`
    OrderSN     string    `json:"orderSn"`
    OrderStatus string    `json:"orderStatus"`
    TotalAmount float64   `json:"totalAmount"`
    CreatedAt   time.Time `json:"createdAt"`
}

// OrderListResponse untuk paginated response
type OrderListResponse struct {
    Success bool            `json:"success"`
    Data    []OrderResponse `json:"data"`
    Meta    PaginationMeta  `json:"meta"`
}

type PaginationMeta struct {
    Total int `json:"total"`
    Page  int `json:"page"`
    Limit int `json:"limit"`
}
```

### 5.5 Interface Definitions

```go
// internal/services/interfaces.go
package services

import "context"

// OrderServiceInterface defines order operations
type OrderServiceInterface interface {
    GetOrders(ctx context.Context, tenantID, platform string, page, limit int) ([]interface{}, int64, error)
    GetOrderByID(ctx context.Context, tenantID, platform, orderID string) (interface{}, error)
    SyncOrders(ctx context.Context, tenantID, platform string) error
}

// InventoryServiceInterface defines inventory operations
type InventoryServiceInterface interface {
    GetRecords(ctx context.Context, tenantID string) ([]models.InventoryRecord, error)
    SyncFromSheet(ctx context.Context, tenantID string) error
    UpdateStock(ctx context.Context, tenantID, sku string, quantity int) error
}
```

---

## 6. Error Handling & Logging

### 6.1 Custom Error Types

```go
// internal/errors/errors.go
package errors

type AppError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Status  int    `json:"-"`
}

func (e *AppError) Error() string {
    return e.Message
}

// Predefined errors
var (
    ErrNotFound          = &AppError{Code: "NOT_FOUND", Message: "Resource not found", Status: 404}
    ErrUnauthorized      = &AppError{Code: "UNAUTHORIZED", Message: "Authentication required", Status: 401}
    ErrForbidden         = &AppError{Code: "FORBIDDEN", Message: "Access denied", Status: 403}
    ErrBadRequest        = &AppError{Code: "BAD_REQUEST", Message: "Invalid request", Status: 400}
    ErrInternalServer    = &AppError{Code: "INTERNAL_ERROR", Message: "Internal server error", Status: 500}
    ErrMissingTenantID   = &AppError{Code: "MISSING_TENANT", Message: "Tenant ID is required", Status: 401}
    ErrInvalidPlatform   = &AppError{Code: "INVALID_PLATFORM", Message: "Unknown platform", Status: 400}
)

// Wrap error with context
func Wrap(err error, message string) error {
    return fmt.Errorf("%s: %w", message, err)
}
```

### 6.2 Logging Pattern

```go
// Gunakan structured logging
import "github.com/rs/zerolog/log"

// ✅ BENAR - Structured logging
log.Info().
    Str("tenant_id", tenantID).
    Str("platform", platform).
    Int("order_count", len(orders)).
    Msg("Orders synced successfully")

log.Error().
    Err(err).
    Str("tenant_id", tenantID).
    Str("order_sn", orderSN).
    Msg("Failed to process order")

// ❌ SALAH - String formatting
log.Printf("Orders synced for tenant %s: %d orders", tenantID, len(orders))
```

### 6.3 Error Response Pattern

```go
// Helper function untuk error response
func errorResponse(c *fiber.Ctx, err error) error {
    var appErr *errors.AppError
    if errors.As(err, &appErr) {
        return c.Status(appErr.Status).JSON(fiber.Map{
            "success": false,
            "error":   appErr.Message,
            "code":    appErr.Code,
        })
    }

    // Generic error - don't expose details in production
    log.Error().Err(err).Msg("Unhandled error")
    return c.Status(500).JSON(fiber.Map{
        "success": false,
        "error":   "Internal server error",
    })
}
```

---

## 7. Testing

### 7.1 Struktur Test

| Folder                       | Isi               |
| ---------------------------- | ----------------- |
| `backend/tests/unit/`        | Unit tests        |
| `backend/tests/integration/` | Integration tests |
| `backend/tests/e2e/`         | End-to-end tests  |

### 7.2 Perintah Test

```bash
cd backend

# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/services/...

# Run with verbose
go test -v ./internal/handlers/...
```

### 7.3 Test Pattern

```go
// internal/services/order_service_test.go
package services_test

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestOrderService_GetOrders(t *testing.T) {
    // Arrange
    mockRepo := new(MockOrderRepository)
    service := NewOrderService(mockRepo)

    expectedOrders := []models.ShopeeOrder{
        {ID: 1, OrderSN: "TEST001"},
    }
    mockRepo.On("FindAll", mock.Anything, 1, 20).Return(expectedOrders, int64(1), nil)

    // Act
    orders, total, err := service.GetOrders(context.Background(), "test_tenant", "shopee", 1, 20)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, int64(1), total)
    assert.Len(t, orders, 1)
    mockRepo.AssertExpectations(t)
}
```

---

## 8. Security

### 8.1 Security Features

| Fitur             | Implementasi                   | File                               |
| ----------------- | ------------------------------ | ---------------------------------- |
| **JWT Auth**      | Token validation               | `internal/middleware/auth.go`      |
| **CORS**          | Origin whitelist               | `internal/middleware/cors.go`      |
| **Rate Limiting** | Request throttling             | `internal/middleware/ratelimit.go` |
| **SQL Injection** | GORM parameterized queries     | Semua repositories                 |
| **XSS**           | JSON response (no HTML render) | Semua handlers                     |
| **Encryption**    | Fernet encryption for secrets  | `internal/utils/encryption.go`     |

### 8.2 Protected vs Public Endpoints

**🔒 Protected (Butuh JWT):**

```
/api/orders/*
/api/products/*
/api/inventory/*
/api/settings/*
/api/analytics/*
```

**🌐 Public (No Auth):**

```
/api/health
/api/auth/login
/api/auth/register
/api/webhooks/*
```

### 8.3 Sensitive Data Handling

```go
// ✅ BENAR - Jangan log sensitive data
log.Info().
    Str("user_id", user.ID).
    Str("action", "login").
    Msg("User logged in")

// ❌ SALAH - Jangan log password/token
log.Info().
    Str("password", password).     // NEVER!
    Str("token", accessToken).     // NEVER!
    Msg("Auth attempt")

// ✅ BENAR - Response tanpa password
type UserResponse struct {
    ID       string `json:"id"`
    Username string `json:"username"`
    Email    string `json:"email"`
    // Password tidak ada!
}
```

---

### 9.2 Environment Variables

| Variable         | Required | Deskripsi                        |
| ---------------- | -------- | -------------------------------- |
| `GO_ENV`         | ✅       | `development` / `production`     |
| `PORT`           | ✅       | Port server (default: 3000)      |
| `DB_DRIVER`      | ✅       | `sqlite` atau `postgres`         |
| `PG_HOST`        | ✅       | PostgreSQL host                  |
| `PG_PORT`        | ✅       | PostgreSQL port (5432)           |
| `PG_USER`        | ✅       | PostgreSQL user                  |
| `PG_PASSWORD`    | ✅       | PostgreSQL password              |
| `PG_DATABASE`    | ✅       | Database name                    |
| `JWT_SECRET`     | ✅       | JWT signing secret (min 32 char) |
| `ENCRYPTION_KEY` | ✅       | Fernet encryption key (base64)   |

### 9.3 Health Check

```bash
# Check API health
curl http://localhost:3000/api/health

# Expected response:
{
  "status": "healthy",
  "database": "connected",
  "version": "1.0.0"
}
```

---

## 10. Frontend Integration

### 10.1 API Response Format (WAJIB KONSISTEN)

```json
// Success response
{
  "success": true,
  "data": { ... },
  "meta": {
    "total": 100,
    "page": 1,
    "limit": 20
  }
}

// Error response
{
  "success": false,
  "error": "Human readable message",
  "code": "ERROR_CODE"
}
```

### 10.2 Backward Compatibility

Saat migrasi, pastikan response format konsisten:

```go
// ✅ BENAR - Format yang konsisten
return c.JSON(fiber.Map{
    "success": true,
    "data":    orders,
    "meta": fiber.Map{
        "total": total,
        "page":  page,
        "limit": limit,
    },
})

// ❌ SALAH - Format berbeda dari Node.js
return c.JSON(fiber.Map{
    "status": "OK",           // Harusnya "success": true
    "orders": orders,          // Harusnya "data"
    "count": len(orders),      // Format berbeda
})
```

### 10.3 Frontend Testing Setelah Migrasi

```bash
# Build frontend
cd frontend
npm run build

# Deploy
cd ..
docker compose -f docker-compose.n8n.production-full.yml up -d --build frontend

# Test dengan Lighthouse
cd backend
node tests/e2e/lighthouse-quick-test.js operation inventory
```

---

## 11. Workflow untuk AI Agent

### 11.1 Sebelum Menulis Kode

```
┌──────────────────────────────────────────────────────────────────┐
│  CHECKLIST SEBELUM CODING                                        │
├──────────────────────────────────────────────────────────────────┤
│  □ Baca AGENTS.MD ini dengan saksama                             │
│  □ Cek referensi Node.js di backend-node/                        │
│  □ Identifikasi model/schema yang dibutuhkan                     │
│  □ Cek apakah model sudah ada di backend/internal/models/        │
│  □ Cek table_naming.go untuk mapping nama tabel                  │
│  □ Pastikan ada migration SQL jika tabel baru                    │
│  □ Identifikasi apakah perlu multi-tenant handling               │
└──────────────────────────────────────────────────────────────────┘
```

### 11.4 Ringkasan Workflow

```
┌─────────────────────────────────────────────────────────────────┐
│  SEBELUM EDIT                                                   │
├─────────────────────────────────────────────────────────────────┤
│  1. Baca AGENTS.MD (file ini)                                   │
│  2. Cek referensi di backend-node/ untuk logic yang benar       │
│  3. Pastikan model & table mapping sudah ada                    │
│  4. Verifikasi apakah perlu update migration SQL                │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│  SAAT EDIT                                                      │
├─────────────────────────────────────────────────────────────────┤
│  1. Ikuti pola Repository → Service → Handler                  │
│  2. SELALU handle tenant_id (no default!)                       │
│  3. SELALU gunakan context.Context                              │
│  4. SELALU structured logging                                   │
│  5. Response format harus konsisten dengan Node.js              │
└─────────────────────────────────────────────────────────────────┘
                            ↓
┌─────────────────────────────────────────────────────────────────┐
│  SETELAH EDIT                                                   │
├─────────────────────────────────────────────────────────────────┤
│  1. Build: go build ./...                                       │
│  2. Test: go test ./...                                         │
│  3. Docker: docker compose up -d --build backend-go             │
│  4. Verifikasi endpoint dengan curl/Postman                     │
│  5. Cek frontend masih berfungsi (jika API berubah)             │
└─────────────────────────────────────────────────────────────────┘
```

---

## 🚨 Ringkasan Penting

> [!WARNING]
> **6 Aturan yang TIDAK BOLEH Dilanggar:**
>
> 1. **JANGAN** gunakan default tenant - throw error jika kosong
> 2. **JANGAN** edit `backend-node/` - hanya untuk referensi
> 3. **JANGAN** buat tabel baru tanpa update migration SQL
> 4. **JANGAN** ubah format API response tanpa sync frontend
> 5. **JANGAN** hardcode connection string atau credentials
> 6. **JANGAN** gunakan cache dalam bentuk apapun (Redis, in-memory, dll)

> [!CAUTION]
> **Aturan Penamaan - TIDAK BOLEH PAKAI ALIAS:**
>
> Jika ada kesalahan penamaan (variable, function, file, table, column), **PERBAIKI TOTAL** - jangan gunakan alias atau workaround!
>
> ```go
> // ❌ SALAH - Menggunakan alias untuk nama yang salah
> type ShopeeOrdr struct { ... }  // Typo
> type ShopeeOrder = ShopeeOrdr   // Alias - JANGAN!
>
> // ✅ BENAR - Perbaiki langsung nama yang benar
> type ShopeeOrder struct { ... }  // Fix the name directly
> ```

---

**⚠️ INGAT: Ini adalah sistem multi-tenant dengan migrasi aktif. Setiap perubahan harus backward-compatible dengan frontend yang ada.**
