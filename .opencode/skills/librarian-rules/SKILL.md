---
description: Librarian research rules for OMNI project - documentation and OSS lookup specialist
---

# Librarian Rules

<!-- MASTER:skill-librarian-role -->
> **Role:** Multi-repo analysis, documentation lookup, OSS implementation examples.
> **Constraints:** Cannot write, edit, or delegate.
<!-- /MASTER:skill-librarian-role -->

---

## Research Priority Order (MANDATORY!)

<!-- MASTER:skill-librarian-research-priority-order -->
```
1️⃣ LOCAL SDK FIRST (HIGHEST PRIORITY)
   backend/shopee-sdk/     - Shopee API
   backend/lazada-sdk/     - Lazada API
   backend/tiktok_sdk/     - TikTok API (100+ files)

2️⃣ EXISTING CODEBASE
   internal/               - Existing implementations

3️⃣ EXTERNAL DOCS (LAST RESORT)
   Official API docs       - Only if not found locally
```
<!-- /MASTER:skill-librarian-research-priority-order -->

---

## SDK Folder Paths

<!-- MASTER:skill-librarian-sdk-folder-paths -->
| Platform   | Local Path                       | Key Files                         |
| ---------- | -------------------------------- | --------------------------------- |
| Shopee     | `backend/shopee-sdk/`            | orders.go, products.go, client.go |
| Lazada     | `backend/lazada-sdk/`            | order.go, product.go, auth.go     |
| Lazada IOP | `backend/lazada_sdk/iop-sdk-go/` | Official IOP SDK                  |
| TikTok     | `backend/tiktok_sdk/`            | Comprehensive SDK (100+ files)    |
<!-- /MASTER:skill-librarian-sdk-folder-paths -->

---

## Official Documentation (Backup Reference)

<!-- MASTER:skill-librarian-official-documentation -->
| Platform  | Documentation URL                   |
| --------- | ----------------------------------- |
| Shopee    | https://open.shopee.com/documents   |
| Lazada    | https://open.lazada.com/doc/api.htm |
| TikTok    | https://partner.tiktokshop.com/doc  |
| Tokopedia | https://developer.tokopedia.com/    |
<!-- /MASTER:skill-librarian-official-documentation -->

---

## When to Use

<!-- MASTER:skill-librarian-when-to-use -->
| Trigger               | Action                       |
| --------------------- | ---------------------------- |
| External API error    | Find error code meaning      |
| Format mismatch       | Find official API spec       |
| OAuth/Auth issues     | Find auth flow documentation |
| Rate limiting         | Find best practices          |
| Unfamiliar Go library | Find usage examples          |
<!-- /MASTER:skill-librarian-when-to-use -->

---

## Output Format

<!-- MASTER:skill-librarian-output-format -->
```markdown
## Research Findings

### Source

[Local SDK / Codebase / External Docs]

### Relevant Code/Docs

[Code snippets or doc excerpts]

### Key Insights

- [Finding 1]
- [Finding 2]

### Recommended Implementation

[Based on research]

### References

- [File path or URL]
```
<!-- /MASTER:skill-librarian-output-format -->

---

## Anti-Patterns

<!-- MASTER:skill-librarian-anti-patterns -->
| Don't                  | Do Instead                     |
| ---------------------- | ------------------------------ |
| Skip local SDK search  | ALWAYS check local SDK first   |
| Jump to external docs  | Exhaust local options first    |
| Return vague findings  | Provide specific code examples |
| Forget to cite sources | Always include file path/URL   |
<!-- /MASTER:skill-librarian-anti-patterns -->
