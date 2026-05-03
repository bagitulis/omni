# Middleware Execution Chain

## Overview

The OMNI backend uses a layered middleware architecture to handle cross-cutting concerns like authentication, CORS, logging, rate limiting, and error handling. This document describes the exact execution order, what each middleware does, and how tenant_id flows through the request lifecycle.

## Middleware Registration Order

The middleware chain is registered in `backend/cmd/server/main.go` (lines 53-173). **Order matters** — middlewares execute in the order they are registered.

### Global Middleware (Applied to All Routes)

```go
// backend/cmd/server/main.go:53-55
router := gin.Default()
router.Use(middleware.CORS())
router.Use(middleware.Logger())
```

**Execution Order:**
1. **CORS** (line 54)
2. **Logger** (line 55)

### Protected Routes Middleware (Applied to `/api/*` Protected Routes)

```go
// backend/cmd/server/main.go:171-173
protected := api.Group("")
protected.Use(middleware.Auth())
protected.Use(middleware.Tenant())
```

**Execution Order (for protected routes only):**
1. **Auth** (line 172)
2. **Tenant** (line 173)

---

## Middleware Details

### 1. CORS Middleware

**File:** `backend/internal/middleware/cors.go`

**Purpose:** Handle Cross-Origin Resource Sharing (CORS) and set security headers.

**Execution Stage:** First (global)

**What It Does:**
- Validates incoming `Origin` header against allowed origins
- Sets CORS response headers (`Access-Control-Allow-*`)
- Adds security headers (`X-Content-Type-Options`, `X-Frame-Options`, etc.)
- Handles preflight OPTIONS requests

**Configuration:**
- **Allowed Origins:** Hardcoded defaults + environment variable `CORS_ORIGINS`
- **Default Origins:**
  - `http://localhost:5173` (Vite dev port)
  - `http://localhost:5174` (Vite alt port)
  - `http://localhost:3000` (React dev)
  - `http://localhost:80` / `http://localhost`
  - `https://yndigital.my.id` (production)
  - `https://www.yndigital.my.id`
  - `http://yndigital.my.id`
  - `http://www.yndigital.my.id`

**Headers Set:**
```
Access-Control-Allow-Origin: [origin]
Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE, OPTIONS
Access-Control-Allow-Headers: Origin, Content-Type, Accept, Authorization, x-tenant-id, x-csrf-token, Cache-Control, Pragma
Access-Control-Allow-Credentials: true
Access-Control-Max-Age: 86400
X-Content-Type-Options: nosniff
X-Frame-Options: SAMEORIGIN
X-XSS-Protection: 1; mode=block
Referrer-Policy: strict-origin-when-cross-origin
```

**Context Mutations:** None

---

### 2. Logger Middleware

**File:** `backend/internal/middleware/logging.go`

**Purpose:** Log all incoming requests with structured logging.

**Execution Stage:** Second (global)

**What It Does:**
- Records request start time
- Logs HTTP method, path, status code, and latency
- Uses zerolog for structured logging

**Log Output:**
```json
{
  "level": "info",
  "method": "GET",
  "path": "/api/shopee/orders",
  "status": 200,
  "latency": "125ms",
  "message": "Request"
}
```

**Context Mutations:** None

---

### 3. Auth Middleware

**File:** `backend/internal/middleware/auth.go`

**Purpose:** Validate JWT token and extract user identity.

**Execution Stage:** Third (protected routes only)

**When Applied:**
- All routes under `/api/*` protected group (line 172 in main.go)
- Specifically: `/api/shopee/*`, `/api/lazada/*`, `/api/tiktok/*`, etc.

**What It Does:**
1. Extracts JWT token from:
   - `Authorization: Bearer <token>` header (preferred)
   - `?token=<token>` query parameter (fallback for SSE/EventSource)
2. Validates JWT signature using `JWT_SECRET`
3. Extracts claims from token
4. Sets context values for downstream handlers

**JWT Validation:**
- **Secret Source:** Environment variable `JWT_SECRET` (min 32 chars)
- **Validation:** Uses `utils.JWTService.ValidateToken()`
- **Error Handling:** Returns 401 Unauthorized if token is missing, invalid, or expired

**Context Keys Set:**
```go
c.Set("userID", claims.UserID)        // User identifier
c.Set("tenant_id", claims.TenantID)   // Tenant identifier (from JWT)
c.Set("role", claims.Role)            // User role (admin, user, etc.)
```

**Error Responses:**
```json
// Missing token
{
  "success": false,
  "error": "Missing authorization token"
}

// Invalid format
{
  "success": false,
  "error": "Invalid authorization format"
}

// Invalid/expired token
{
  "success": false,
  "error": "Unauthorized"
}
```

**Security Notes:**
- Never logs token values or parts
- Returns generic "Unauthorized" message (doesn't expose internal details)
- Requires minimum 32-char secret in production
- Fails fast in production if JWT_SECRET is missing

---

### 4. Tenant Middleware

**File:** `backend/internal/middleware/tenant.go`

**Purpose:** Extract and validate tenant ID, enforce multi-tenant isolation.

**Execution Stage:** Fourth (protected routes only)

**When Applied:**
- All routes under `/api/*` protected group (line 173 in main.go)
- Runs AFTER Auth middleware

**What It Does:**
1. Checks if `tenant_id` was already set by Auth middleware (from JWT)
2. If not set, extracts from `x-tenant-id` header (legacy flow)
3. Validates tenant ID format and existence
4. Sets canonical `tenant_id` context key

**Tenant ID Extraction Flow:**
```
┌─────────────────────────────────────────────────────────┐
│ Request arrives                                         │
└────────────────┬────────────────────────────────────────┘
                 │
                 ▼
        ┌────────────────────┐
        │ Auth middleware    │
        │ runs first         │
        └────────┬───────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ JWT contains tenant_id?                │
        └────────┬──────────────────────────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ YES: Set c.Set("tenant_id", ...)      │
        │ NO: tenant_id remains empty           │
        └────────┬──────────────────────────────┘
                 │
                 ▼
        ┌────────────────────┐
        │ Tenant middleware  │
        │ runs second        │
        └────────┬───────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ tenant_id already set from JWT?       │
        └────────┬──────────────────────────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ YES: Skip header check, use JWT value │
        │ NO: Check x-tenant-id header          │
        └────────┬──────────────────────────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ tenant_id empty?                      │
        └────────┬──────────────────────────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ YES: Return 401 "Missing tenant_id"   │
        │ NO: Validate format/existence         │
        └────────┬──────────────────────────────┘
                 │
        ┌────────▼──────────────────────────────┐
        │ Valid? Set c.Set("tenant_id", ...)    │
        │ Invalid? Return 401 "Invalid tenant"  │
        └────────────────────────────────────────┘
```

**Validation Rules:**
- **Format:** Alphanumeric + underscore only (`^[a-zA-Z0-9_]+$`)
- **Existence:** Checked against `config.ValidateTenant()` if tenants are loaded
- **Fallback:** Format-only validation if tenant config not loaded

**Context Keys Set:**
```go
c.Set("tenant_id", tenantID)  // Canonical tenant identifier
```

**Error Responses:**
```json
// Missing tenant_id
{
  "success": false,
  "error": "Missing tenant_id"
}

// Invalid tenant (from config)
{
  "success": false,
  "error": "Invalid tenant ID"
}

// Invalid format
{
  "success": false,
  "error": "Invalid tenant ID format"
}
```

**CRITICAL RULE:** No default tenant. If tenant_id is missing, request is rejected with 401.

---

## Tenant ID Flow Through Request Lifecycle

### Flow Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│ 1. Client sends request                                          │
│    Authorization: Bearer <JWT>                                   │
│    x-tenant-id: tenant_123 (optional, legacy)                    │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 2. CORS Middleware                                               │
│    - Validates origin                                            │
│    - Sets CORS headers                                           │
│    - No tenant_id handling                                       │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 3. Logger Middleware                                             │
│    - Records request start                                       │
│    - No tenant_id handling                                       │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 4. Auth Middleware (Protected Routes Only)                       │
│    - Extracts JWT token                                          │
│    - Validates signature                                         │
│    - Extracts claims:                                            │
│      • userID                                                    │
│      • tenant_id ◄─── TENANT ID SOURCE #1                        │
│      • role                                                      │
│    - Sets context: c.Set("tenant_id", claims.TenantID)           │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 5. Tenant Middleware (Protected Routes Only)                     │
│    - Checks if tenant_id already set by Auth                     │
│    - If YES: Use JWT value (already validated)                   │
│    - If NO: Extract from x-tenant-id header ◄─── TENANT ID #2    │
│    - Validate format/existence                                   │
│    - Set canonical context: c.Set("tenant_id", tenantID)         │
│    - Reject if missing (NO DEFAULT TENANT)                       │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 6. Handler receives request                                      │
│    - Extracts tenant_id via middleware.GetTenantID(c)            │
│    - Uses for database schema selection                          │
│    - Passes to service layer                                     │
│    - Service uses for data isolation                             │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 7. Service Layer                                                 │
│    - Receives tenant_id from handler                             │
│    - Queries database with tenant schema: tenant_{tenantID}      │
│    - All data operations scoped to tenant                        │
└────────────────────┬─────────────────────────────────────────────┘
                     │
                     ▼
┌──────────────────────────────────────────────────────────────────┐
│ 8. Response sent back to client                                  │
│    - Tenant-scoped data only                                     │
└──────────────────────────────────────────────────────────────────┘
```

### Tenant ID Sources (Priority Order)

1. **JWT Claims (Primary)** — Extracted by Auth middleware
   - Most secure (cryptographically signed)
   - Already validated by JWT signature check
   - Preferred source

2. **x-tenant-id Header (Legacy/Fallback)** — Extracted by Tenant middleware
   - Used if JWT doesn't contain tenant_id
   - Validated against tenant config
   - Less secure (can be spoofed if not over HTTPS)

### Context Key Naming

**Canonical Key:** `"tenant_id"` (snake_case)

All middleware and handlers use this single key:
```go
// Auth middleware sets it
c.Set("tenant_id", claims.TenantID)

// Tenant middleware validates and re-sets it
c.Set("tenant_id", tenantID)

// Handlers extract it
tenantID := middleware.GetTenantID(c)

// Or directly
tenantID := c.GetString("tenant_id")
```

---

## Additional Middleware (Not in Global Chain)

These middleware are available but not automatically applied to all routes:

### CSRF Protection Middleware

**File:** `backend/internal/middleware/csrf.go`

**Purpose:** Prevent Cross-Site Request Forgery attacks.

**When Applied:** Not applied globally; must be explicitly added to routes

**What It Does:**
- Implements double-submit cookie pattern
- Validates CSRF token in `x-csrf-token` header
- Skips validation for safe methods (GET, HEAD, OPTIONS)
- Skips validation for exempt paths (webhooks, auth, health)

**Exempt Paths:**
```go
/api/webhooks/
/api/platform-auth/
/api/health
/api/n8n/
/api/auth/login
/api/auth/register
/api/auth/refresh
/api/auth/dev-login
/api/csrf-token
/api/status
```

**Token Generation:**
```go
token, err := middleware.GenerateCSRFToken()
// Returns 32-byte base64-encoded token
// Stored with 24-hour TTL
```

**Context Mutations:** None (validates only)

---

### Error Handler Middleware

**File:** `backend/internal/middleware/error_handler.go`

**Purpose:** Recover from panics and handle errors consistently.

**When Applied:** Not applied globally; typically used in error recovery

**What It Does:**
- Recovers from panics
- Logs panic with stack trace (dev only)
- Converts errors to standard response format
- Handles `AppError` types with proper HTTP status codes

**Error Response Format:**
```json
{
  "success": false,
  "error": "Error message",
  "type": "error_type",
  "details": {}
}
```

**Context Mutations:** None

---

### Request Logger Middleware

**File:** `backend/internal/middleware/request_logger.go`

**Purpose:** Structured request/response logging with request ID tracking.

**When Applied:** Not applied globally; can be added for detailed logging

**What It Does:**
- Generates or extracts request ID
- Logs request start with method, path, client IP
- Logs response with status code, duration, errors
- Adds request ID to response header

**Context Keys Set:**
```go
c.Set("requestId", requestID)
c.Header("X-Request-ID", requestID)
```

**Log Levels:**
- 5xx errors → ERROR
- 4xx errors → WARN
- 2xx/3xx → INFO

---

### Rate Limiting Middleware

**File:** `backend/internal/middleware/rate_limit.go`

**Purpose:** Prevent abuse and protect against DoS attacks.

**When Applied:** Not applied globally; must be explicitly added to specific routes

**Algorithm:** Token bucket with configurable rate and burst size

**Available Rate Limiters:**

#### 1. AuthRateLimitMiddleware()
- **Limit:** 30 attempts per minute per IP
- **Burst:** 50
- **Use Case:** General auth endpoints
- **Rationale:** Prevents brute force while allowing office users on shared NAT

#### 2. LoginRateLimitMiddleware()
- **Limit:** 10 attempts per minute per IP
- **Burst:** 20
- **Use Case:** Login endpoint specifically
- **Rationale:** Stricter than general auth; combined with account lockout

#### 3. APIRateLimitMiddleware()
- **Limit:** 1000 requests per minute per user
- **Burst:** 2000
- **Use Case:** General API endpoints
- **Rationale:** Allows sync operations (100-500 calls per batch)

#### 4. SyncRateLimitMiddleware()
- **Limit:** 5000 requests per minute per tenant
- **Burst:** 10000
- **Use Case:** Platform sync endpoints (Shopee, Lazada, TikTok)
- **Rationale:** Very high for API-intensive sync operations

#### 5. WebhookRateLimitMiddleware()
- **Limit:** 1000 requests per minute per source IP
- **Burst:** 2000
- **Use Case:** Incoming webhooks
- **Rationale:** Handles flash sale bursts (200+ orders in 5 min × 3 platforms)

#### 6. PublicRateLimitMiddleware()
- **Limit:** 30 requests per minute per IP
- **Burst:** 60
- **Use Case:** Public/unauthenticated endpoints
- **Rationale:** Prevents scraping and enumeration

**Response Headers:**
```
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
Retry-After: 60
```

**Error Response:**
```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "message": "Too many requests, please try again later",
  "retry_after": 60
}
```

---

## Complete Request Lifecycle Example

### Request: GET /api/shopee/orders

```
1. Client sends:
   GET /api/shopee/orders HTTP/1.1
   Authorization: Bearer eyJhbGc...
   Origin: http://localhost:5173

2. CORS Middleware
   ✓ Validates origin (localhost:5173 allowed)
   ✓ Sets CORS headers
   → Continue

3. Logger Middleware
   ✓ Records start time
   ✓ Logs: GET /api/shopee/orders
   → Continue

4. Auth Middleware (protected route)
   ✓ Extracts token from Authorization header
   ✓ Validates JWT signature
   ✓ Extracts claims:
     - userID: "user_123"
     - tenant_id: "tenant_abc"
     - role: "admin"
   ✓ Sets context:
     c.Set("userID", "user_123")
     c.Set("tenant_id", "tenant_abc")
     c.Set("role", "admin")
   → Continue

5. Tenant Middleware (protected route)
   ✓ Checks if tenant_id already set: YES (from JWT)
   ✓ Skips header check (JWT is authoritative)
   ✓ Validates tenant_id format: ✓ Valid
   ✓ Sets context: c.Set("tenant_id", "tenant_abc")
   → Continue

6. Handler (shopee.GetOrders)
   ✓ Extracts tenant_id: middleware.GetTenantID(c) → "tenant_abc"
   ✓ Calls service layer with tenant_id
   → Service queries database schema: tenant_abc

7. Service Layer
   ✓ Queries: SELECT * FROM tenant_abc.orders
   ✓ Returns tenant-scoped data

8. Response
   HTTP/1.1 200 OK
   Access-Control-Allow-Origin: http://localhost:5173
   X-Request-ID: abc-123-def
   Content-Type: application/json
   
   {
     "success": true,
     "data": [...]
   }
```

---

## Security Considerations

### 1. Tenant Isolation
- **Enforced at:** Tenant middleware (rejects missing tenant_id)
- **Validated at:** Service layer (queries use tenant schema)
- **Fallback:** Database schema isolation (`tenant_{tenantID}`)

### 2. Authentication
- **JWT Secret:** Minimum 32 characters, required in production
- **Token Validation:** Cryptographic signature check
- **Token Extraction:** Bearer header (preferred) or query param (SSE)

### 3. CORS
- **Allowed Origins:** Whitelist + environment variable
- **Credentials:** Allowed (cookies sent with requests)
- **Preflight:** Handled automatically

### 4. Rate Limiting
- **Per-IP:** Auth endpoints (prevents brute force)
- **Per-User:** API endpoints (prevents abuse)
- **Per-Tenant:** Sync endpoints (prevents one tenant affecting others)

### 5. CSRF Protection
- **Pattern:** Double-submit cookie
- **Exempt Paths:** Webhooks, auth, health checks
- **Safe Methods:** GET, HEAD, OPTIONS (skipped)

---

## Debugging Middleware Issues

### Check Tenant ID Flow

```go
// In handler
tenantID := middleware.GetTenantID(c)
if tenantID == "" {
    log.Error().Msg("tenant_id missing from context")
    // This means Tenant middleware didn't run or failed
}
```

### Check Auth Status

```go
// In handler
userID := middleware.GetUserID(c)
role := middleware.GetRole(c)
if userID == "" {
    log.Error().Msg("userID missing - Auth middleware failed")
}
```

### Check Request ID

```go
// In handler
requestID := middleware.GetRequestID(c)
log.Info().Str("request_id", requestID).Msg("Processing request")
```

### Common Issues

| Issue | Cause | Solution |
|-------|-------|----------|
| 401 "Missing tenant_id" | Tenant middleware ran but tenant_id not set | Check JWT claims or x-tenant-id header |
| 401 "Unauthorized" | JWT validation failed | Check JWT_SECRET, token expiry, signature |
| CORS error | Origin not in whitelist | Add to CORS_ORIGINS env var or defaults |
| Rate limit exceeded | Too many requests | Check rate limit config for endpoint |
| CSRF token invalid | Token expired or missing | Regenerate token, check TTL (24h) |

---

## Configuration Reference

### Environment Variables

| Variable | Purpose | Default | Required |
|----------|---------|---------|----------|
| `JWT_SECRET` | JWT signing key | dev-secret-key-... | Yes (prod) |
| `JWT_REFRESH_SECRET` | Refresh token key | (from JWT_SECRET) | No |
| `GO_ENV` | Environment mode | development | No |
| `CORS_ORIGINS` | Additional allowed origins | (empty) | No |

### Middleware Configuration

| Middleware | Config Location | Configurable |
|------------|-----------------|--------------|
| CORS | `cors.go` lines 17-27 | Yes (env var) |
| Logger | `logging.go` | No |
| Auth | `auth.go` lines 15-43 | Yes (env var) |
| Tenant | `tenant.go` lines 15-16 | No |
| CSRF | `csrf.go` lines 14-34 | Yes (constants) |
| Rate Limit | `rate_limit.go` lines 206-295 | Yes (functions) |

---

## Summary

The OMNI middleware chain enforces:

1. **CORS** — Cross-origin requests validated
2. **Logging** — All requests logged
3. **Authentication** — JWT token validated (protected routes)
4. **Tenant Isolation** — tenant_id extracted and validated (protected routes)
5. **Optional:** CSRF, Error Handling, Rate Limiting

**Tenant ID Flow:**
- Extracted from JWT claims (Auth middleware)
- Validated by Tenant middleware
- Passed to handlers via context
- Used for database schema selection in service layer

**Critical Rule:** No default tenant — missing tenant_id results in 401 Unauthorized.
