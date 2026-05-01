# OMNI

[![Go Version](https://img.shields.io/badge/Go-1.24-blue.svg)](https://golang.org/)
[![React Version](https://img.shields.io/badge/React-19-blue.svg)](https://react.dev/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-blue.svg)](https://www.postgresql.org/)
[![Docker](https://img.shields.io/badge/Docker-Ready-blue.svg)](https://www.docker.com/)

A multi-platform e-commerce order management system (OMS) integrating Shopee, Lazada, and TikTok Shop into a unified dashboard.

## Features

- **Multi-Platform Integration**: Connect and manage orders from Shopee, Lazada, and TikTok Shop
- **Unified Dashboard**: Single interface for all marketplace operations
- **Smart Database Backup**: PostgreSQL backup with change detection, SHA-256 checksums, and automatic chunking
- **Multi-Tenant Architecture**: Schema-based tenant isolation for secure data separation
- **JWT Authentication**: Secure access with refresh tokens and rate limiting
- **Automated Build System**: Python-based Docker orchestration with 22+ error recovery patterns

## Quick Start

```bash
# 1. Clone and navigate to project
cd omni

# 2. Configure environment
cp .env.example .env
# Edit .env with your API keys and secrets

# 3. Build and run
python build.py smart

# 4. Access the application
# Frontend: http://localhost/
# API: http://localhost/api/
```

## Project Structure

```
omni/
├── backend/              # Go backend (Gin + GORM)
│   ├── cmd/server/       # Application entry point
│   ├── internal/         # Internal packages
│   │   ├── handler/      # HTTP handlers
│   │   ├── service/      # Business logic
│   │   ├── repository/   # Database access
│   │   └── model/        # Data models
│   ├── pkg/              # Shared utilities
│   └── shopee-sdk/       # Platform SDKs
├── frontend/             # React 19 frontend
│   ├── src/
│   │   ├── components/   # UI components
│   │   ├── pages/        # Route pages
│   │   ├── api/          # API clients
│   │   └── stores/       # Zustand state
│   └── public/
├── scripts/              # Build and deployment scripts
├── docker-compose.yml    # Docker orchestration
└── build.py              # Main build entry point
```

## Database Operations

```bash
# Create backup (smart - only if changes detected)
python build.py backup

# Restore from latest backup
python build.py restore

# Full rebuild with restore
python build.py full --restore
```

## Development Setup

### Prerequisites

- Go 1.24+
- Node.js 20+
- Python 3.10+
- Docker and Docker Compose

### Backend Development

```bash
cd backend
go mod download
go run ./cmd/server/
```

### Frontend Development

```bash
cd frontend
npm install
npm run dev
```

## Architecture

OMNI follows a layered architecture pattern:

```
Request → Handler → Service → Repository → Database
```

| Layer | Responsibility |
|-------|---------------|
| Handler | HTTP request parsing, validation, response formatting |
| Service | Business logic, orchestration, transformations |
| Repository | Database access and query execution |

## Configuration

Key environment variables (see `.env.example`):

- `JWT_SECRET` - JWT signing key (min 32 chars)
- `JWT_REFRESH_SECRET` - Refresh token signing key
- `ENCRYPTION_KEY` - Data encryption key (32 chars)
- `SHOPEE_*` - Shopee API credentials
- `LAZADA_*` - Lazada API credentials
- `TIKTOK_*` - TikTok Shop API credentials

## Documentation

- [Backend Documentation](backend/README.md)
- [Frontend Documentation](frontend/README.md)
- [Build System Guide](scripts/python-build/README.md)
- [API Documentation](backend/docs/api.md)

## License

Proprietary - All rights reserved.
