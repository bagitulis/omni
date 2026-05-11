# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- User Management frontend: full CRUD page with table, create/edit modals, role-based actions
- Admin role now has users.create and users.delete permissions
- Sidebar shows User Management menu item for authorized roles (admin+)
- Route /users protected by users.list permission guard

### Fixed

- Disable dead Price Sync button with "coming soon" message instead of empty callback
- Add retry logic with timeout to image download service (retry on 5xx/network errors, no retry on 4xx)
- Fix silent error swallowing in inventory sync_service (json.Marshal) with proper zerolog error logging
- Fix silent error swallowing in route scanner_service (filepath.Walk) with proper zerolog warning

## [1.0.0] - 2026-05-02

### Added

- Multi-platform e-commerce integration (Shopee, Lazada, TikTok Shop)
- Smart PostgreSQL backup/restore with change detection and SHA-256 checksums
- React 19 frontend with Ant Design 5
- JWT authentication with refresh tokens and rate limiting
- Multi-tenant schema isolation (tenant_{id} pattern)
- Automated build system with 22+ error pattern recovery
- MCP server implementations for platform APIs

### Security

- SHA-256 backup integrity verification
- Pre-restore safety backups
- Data encryption for sensitive fields (Fernet)
- CSRF protection and CORS configuration
- Account lockout after failed login attempts
