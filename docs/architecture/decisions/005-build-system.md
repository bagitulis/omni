---
status: accepted
date: 2025-03-01
---

# ADR-005: Python Build Orchestrator (build.py)

## Context

Docker Compose alone doesn't handle complex build scenarios: conditional rebuilds, database backup/restore, error recovery, resource management.

## Decision

Custom Python build orchestrator (`build.py`) that wraps Docker Compose with intelligent build logic.

## Features

- **Smart builds**: Only rebuild changed images (hash-based detection)
- **22+ error recovery patterns**: Automatic retry and fallback
- **Database backup**: SHA-256 checksums, change detection, chunking
- **Resource profiles**: Standard vs low-spec deployments
- **Interactive menu**: For manual operation

## Consequences

- Single entry point for all build/deploy operations
- Requires Python 3.10+ (additional dependency)
- More complex than raw `docker compose up`
- But handles edge cases that raw Docker cannot
