---
status: accepted
date: 2025-01-01
---

# ADR-003: Local Platform SDKs

## Context

OMNI integrates with Shopee, Lazada, and TikTok Shop APIs. These platforms have complex authentication, rate limiting, and frequently changing APIs.

## Decision

Maintain **local SDK implementations** in the repository rather than using third-party Go packages.

## SDK Locations

| Platform | Path | Size |
|----------|------|------|
| Shopee | `backend/shopee-sdk/` | ~10 files |
| Lazada | `backend/lazada-sdk/` | ~10 files |
| Lazada IOP | `backend/lazada_sdk/iop-sdk-go/` | Official SDK |
| TikTok | `backend/tiktok_sdk/` | 100+ files |

## Consequences

- Full control over API integration behavior
- Can patch issues immediately without waiting for upstream
- Must maintain SDK ourselves (no community updates)
- **Rule**: Always check local SDK before external docs
- Larger repository size but faster development cycle
