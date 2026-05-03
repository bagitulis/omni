---
status: accepted
date: 2025-01-01
---

# ADR-002: Layered Architecture (Handler → Service → Repository)

## Context

Need clear separation of concerns for a complex OMS with multiple platform integrations.

## Decision

Strict 3-layer architecture:
- **Handler**: HTTP parsing, validation, response formatting
- **Service**: Business logic, orchestration
- **Repository**: Database access only

## Rules

- Handler CANNOT call Repository directly
- Service CANNOT access HTTP request/response objects
- Repository CANNOT contain business logic
- Each layer only calls the layer below it

## Consequences

- Testable: Each layer can be unit tested independently
- Maintainable: Changes in one layer don't cascade
- Clear ownership: Bugs are easy to locate by layer
- Slightly more boilerplate than flat architecture
