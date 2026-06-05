---
alwaysApply: true
description: Mandatory coding rules and quality standards (synced from project root)
synced_from: RULES.md
---

<!-- AUTO-SYNCED from RULES.md by .claude/sync_project_rules.py -->
<!-- Source: RULES.md | DO NOT EDIT — edit the source file and re-run sync -->

# MANDATORY RULES (Read Before EVERY Action)

| Rule              | Requirement                                     | Check               |
| ----------------- | ----------------------------------------------- | ------------------- |
| **SRP**           | One function = one purpose                      | Before writing code |
| **DRY**           | No duplicate logic, extract to utilities        | Before writing code |
| **OOP**           | Proper encapsulation, use interfaces            | Before writing code |
| **~300 Lines**    | Quality signal — review SRP/DRY/OOP if exceeded | After editing file  |
| **Commit ALL**    | Include ALL changed files, no cherry-pick       | Before committing   |
| **Push Always**   | Push immediately after commit                   | After committing    |
| **Parallel Work** | Fire delegation in background, continue working | When delegating     |

**VIOLATION = TASK FAILURE. No exceptions.**
