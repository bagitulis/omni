# Go Backend

E-commerce multi-platform backend (Shopee, Lazada, TikTok) built with Go + Gin.

## Quick Start

```bash
# Development
go run cmd/server/main.go

# Tests
go test ./tests/... -v

# Build
go build -o server ./cmd/server
```

## Docker

```bash
# Go only
docker-compose -f docker-compose.go.yml up -d --build

# Full stack (Go + Node.js + NGINX)
docker-compose -f docker-compose.complete.yml up -d --build
```

## API Endpoints

| Platform | Method | Endpoint |
|----------|--------|----------|
| Health | GET | `/api/health` |
| Shopee | GET | `/api/shopee/orders` |
| Shopee | GET | `/api/shopee/products` |
| Shopee | POST | `/api/shopee/sync/orders` |
| Lazada | GET | `/api/lazada/orders` |
| Lazada | POST | `/api/lazada/sync/orders` |
| TikTok | GET | `/api/tiktok/orders` |
| TikTok | POST | `/api/tiktok/sync/orders` |

## Required Headers

```
Authorization: Bearer <jwt_token>
x-tenant-id: <tenant_id>
```

## Project Structure

```
backend-go/
├── cmd/server/          # Entry point
├── internal/
│   ├── config/          # Configuration
│   ├── handlers/        # HTTP handlers
│   ├── middleware/      # Auth, CORS, Tenant
│   ├── models/          # GORM models
│   ├── repositories/    # Data access
│   ├── services/        # Business logic
│   └── dto/             # Request/Response
├── pkg/                 # API clients (Shopee, Lazada, TikTok)
├── tests/               # Unit tests
└── Dockerfile           # Production image
```

## Environment

| Variable | Default | Description |
|----------|---------|-------------|
| PORT | 8080 | Server port |
| GO_ENV | development | Environment |
| DATABASE_PATH | ../backend/data | Database folder |
| JWT_SECRET | - | JWT secret |
