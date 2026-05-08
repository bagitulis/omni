# Evaluate Project

> **Trigger:** User says "evaluate project", "evaluate project [name]", "run project evaluation", or "audit project"
> **Agent:** Sisyphus (orchestrator) loads this skill
> **Purpose:** Comprehensive project evaluation using heavy explore + librarian research

---

## Overview

This skill conducts a deep, phased evaluation of a project's codebase. It fires 50-100+ explore/librarian agents to discover issues, research best practices, and produce a structured evaluation report.

**This is NOT bug fixing.** Evaluation = research + analysis + report. Fixes come AFTER, as separate tasks.

---

## Trigger Detection

Load this skill when user says any of:
- "evaluate project"
- "evaluate project [name]"
- "run project evaluation"
- "audit project [name]"
- "evaluate codebase"
- "deep audit"
- "tech debt analysis"

---

## Workflow

### Phase 0: Setup

1. Determine target project (ask if not specified):
   - `ai` (hub) — Python
   - `omni` — Go + React
   - `auto` — Go + Svelte
   - `extensions` — Go + Svelte
   - `ads-analytics` — Go + Python + Svelte

2. Check for existing progress:
   - Read `.sisyphus/evaluation/progress.json` if exists
   - Resume from last checkpoint if incomplete
   - Start fresh if no progress file

3. Ask user which categories to evaluate (default: all):
   - code-quality
   - bugs
   - tests
   - security
   - performance
   - consistency
   - documentation
   - dependencies
   - architecture

4. Create directory structure:
   ```
   .sisyphus/evaluation/
   ├── progress.json
   └── findings/
   ```

---

### Phase 1: DISCOVERY (explore heavy — fire 3 per category)

For each selected category, fire 3 explore agents in parallel:

#### Category: code-quality
```
explore 1: "Find all files > 300 lines. List file paths and line counts."
explore 2: "Find duplicated code patterns (same logic in multiple files). Look for copy-paste."
explore 3: "Find dead code: unused functions, unreachable branches, commented-out code blocks."
```

#### Category: bugs
```
explore 1: "Find potential null/nil pointer dereferences, unhandled errors, missing error checks."
explore 2: "Find race conditions: shared state without mutex, concurrent map access, goroutine leaks."
explore 3: "Find logic errors: off-by-one, wrong comparison operators, inverted conditions."
```

#### Category: tests
```
explore 1: "Find source files that have NO corresponding test file. List untested modules."
explore 2: "Find test files with only happy-path tests (no error cases, no edge cases)."
explore 3: "Find tests that are skipped, commented out, or have TODO markers."
```

#### Category: security
```
explore 1: "Find hardcoded secrets: API keys, passwords, tokens in source code (not .env)."
explore 2: "Find SQL injection risks: string concatenation in queries, unsanitized user input."
explore 3: "Find auth/authz gaps: endpoints without middleware, missing permission checks."
```

#### Category: performance
```
explore 1: "Find N+1 query patterns: database calls inside loops."
explore 2: "Find blocking operations: synchronous I/O in hot paths, missing timeouts."
explore 3: "Find memory issues: unbounded slices/arrays, missing resource cleanup (defer/close)."
```

#### Category: consistency
```
explore 1: "Find naming inconsistencies: mixed camelCase/snake_case, inconsistent prefixes."
explore 2: "Find error handling inconsistencies: some functions return error, others panic/log."
explore 3: "Find API response format inconsistencies: mixed JSON field naming, inconsistent status codes."
```

#### Category: documentation
```
explore 1: "Find exported functions/types without docstrings or comments."
explore 2: "Find README.md files that are outdated (reference non-existent files/features)."
explore 3: "Find API endpoints without documentation (no swagger, no comments)."
```

#### Category: dependencies
```
explore 1: "Check go.mod/package.json for outdated dependencies (compare versions)."
librarian 1: "Search for known vulnerabilities in our dependency versions."
explore 2: "Find unused dependencies: imported but never used packages."
```

#### Category: architecture
```
explore 1: "Find circular dependencies: package A imports B, B imports A."
explore 2: "Find layer violations: handlers calling repositories directly (skipping service layer)."
explore 3: "Find god objects: files/structs with too many responsibilities (>10 methods)."
```

**After each category completes:**
- Aggregate results
- Write to `.sisyphus/evaluation/findings/{category}.md`
- Update `progress.json` (mark category as done)
- Commit checkpoint: `git add -A && git commit -m "eval(project): complete {category} scan"`

---

### Phase 2: RESEARCH (librarian — for top findings)

After Phase 1, take the top 10 most critical findings and fire librarian agents:

```
For each critical finding:
  librarian: "Research best practice for [issue]. Find how mature projects solve this.
              Compare our approach vs industry standard. Provide specific recommendations."
```

Write research results to `.sisyphus/evaluation/findings/research.md`

---

### Phase 3: SYNTHESIS (aggregate + score)

1. Read all findings from `.sisyphus/evaluation/findings/*.md`
2. Deduplicate (same issue found by multiple agents)
3. Score each finding:
   - **CRITICAL** (score 70-100): Security vulnerability, data loss risk, crash bug
   - **HIGH** (score 45-69): Performance bottleneck, missing auth, broken feature
   - **MODERATE** (score 20-44): Code smell, missing tests, inconsistency
   - **LOW** (score 1-19): Style issue, minor docs gap, trivial improvement
4. Calculate health score: `100 - (weighted_sum_of_findings / files_analyzed * 100)`
5. Generate `EVALUATION_REPORT.md` at project root
6. Generate `.sisyphus/evaluation/findings.json` (machine-readable)

---

### Phase 4: REPORT & TRACK

1. Write `EVALUATION_REPORT.md` with format:
   ```markdown
   # Project Evaluation Report
   
   **Date:** YYYY-MM-DD
   **Project:** [name]
   **Health Score:** XX/100
   **Total Findings:** N (X critical, Y high, Z moderate, W low)
   
   ## Executive Summary
   [2-3 sentences]
   
   ## Top 5 Priorities
   1. [CRITICAL] Title — file:line — effort estimate
   ...
   
   ## Findings by Category
   | Category | Critical | High | Moderate | Low |
   |----------|----------|------|----------|-----|
   ...
   
   ## Detailed Findings
   ### Code Quality
   ...
   ### Security
   ...
   
   ## Recommendations
   1. ...
   
   ## Comparison with Previous Evaluation
   [If previous EVALUATION_REPORT.md exists, compare scores]
   ```

2. Update `CHANGELOG.md`:
   ```markdown
   ## [Unreleased]
   ### Added
   - Project evaluation report (health score: XX/100, N findings)
   ```

3. Commit and push:
   ```
   git add -A
   git commit -m "eval(project): complete evaluation — health score XX/100"
   git push
   ```

4. Report to user with summary

---

## Progress File Format

`.sisyphus/evaluation/progress.json`:
```json
{
  "project": "omni",
  "started": "2026-05-08T10:00:00Z",
  "categories": {
    "code-quality": {"status": "done", "findings": 12},
    "bugs": {"status": "done", "findings": 5},
    "tests": {"status": "in_progress", "findings": 0},
    "security": {"status": "pending"},
    "performance": {"status": "pending"},
    "consistency": {"status": "pending"},
    "documentation": {"status": "pending"},
    "dependencies": {"status": "pending"},
    "architecture": {"status": "pending"}
  },
  "total_findings": 17,
  "health_score": null
}
```

---

## Rules for This Skill

1. **DO NOT FIX** issues during evaluation. Only document them.
2. **Fire explore/librarian aggressively** — 3 per category minimum, more for large projects.
3. **Checkpoint after each category** — commit findings so progress is not lost.
4. **Deduplicate** — same issue found by multiple agents = 1 finding.
5. **Evidence required** — every finding must cite file:line with specific code snippet.
6. **Compare with previous** — if EVALUATION_REPORT.md already exists, show delta.
7. **Respect false_positives.json** — skip known false positives from previous runs.
8. **Time budget** — if evaluation is taking too long, complete current category and checkpoint. User can resume later.

---

## False Positives Management

`.sisyphus/evaluation/false_positives.json`:
```json
{
  "entries": [
    {
      "id": "FP-001",
      "file": "internal/legacy/old_handler.go",
      "reason": "Legacy code scheduled for removal in Q3",
      "added": "2026-05-08",
      "category": "code-quality"
    }
  ]
}
```

When a finding matches a false_positives entry, skip it and note "X false positives suppressed" in report.

---

## Cross-Project Evaluation

If user says "evaluate all projects":
1. Run evaluation for each project sequentially
2. After all done, generate cross-project summary:
   ```
   docs/evaluation/cross-project-summary.md
   ```
3. Identify patterns that appear in multiple projects (systemic issues)
