---
status: in-progress
date: 2025-06-01
---

# ADR-004: React Migration (from Vue)

## Context

Original frontend was Vue.js. Migrating to React 19 for better ecosystem, TypeScript support, and developer experience.

## Decision

Full rewrite in React 19 + Vite 6 + Ant Design 5 + Zustand + TanStack Query.

## Key Design Choices

- **SPA only** (no SSR/Next.js) — Vite for fast dev, static build for production
- **Ant Design 5** — Enterprise-grade components, data-dense tables
- **Zustand** — Lightweight global state (not Redux)
- **TanStack Query** — Server state management with caching
- **System fonts** — No Google Fonts (12px base, data-dense OMS)
- **Ginee-style** — Flat navigation, task-first dashboard

## Constraints

- Must integrate with existing Go backend API (snake_case JSON)
- Windows development (case-sensitivity issues with directories)
- Lowercase directory names only (`modals/` not `Modals/`)

## Status

Migration in progress. Vue legacy code still exists but is being replaced page by page.
