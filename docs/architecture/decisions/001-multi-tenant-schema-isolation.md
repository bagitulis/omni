---
status: accepted
date: 2025-01-01
---

# ADR-001: Multi-Tenant Schema Isolation

## Context

OMNI serves multiple e-commerce sellers. Each seller's data (orders, products, inventory) must be completely isolated for security and compliance.

## Decision

Use **PostgreSQL schema-based isolation** — each tenant gets its own schema (`tenant_{tenantID}`) within a single database.

## Alternatives Considered

| Approach | Pros | Cons |
|----------|------|------|
| **Schema isolation** (chosen) | Strong isolation, shared infra, easy backup per tenant | Schema management complexity |
| Database per tenant | Strongest isolation | Operational overhead, connection limits |
| Row-level (tenant_id column) | Simplest code | Risk of data leaks, complex queries |

## Consequences

- Every query is scoped via `SET search_path TO tenant_{id}`
- No default tenant fallback (hard error if missing)
- PgBouncer handles connection pooling in transaction mode
- Migrations must run per-schema
- Backup/restore can target individual tenants
