---
alwaysApply: true
description: Git operation safety rules - prevents destructive operations
---

# Git Safety Rules

## Allowed Operations

| Command        | Safe | Notes                |
| -------------- | ---- | -------------------- |
| `git status`   | ✅   | Always safe          |
| `git diff`     | ✅   | Always safe          |
| `git log`      | ✅   | Always safe          |
| `git add`      | ✅   | Safe                 |
| `git commit`   | ✅   | Safe                 |
| `git push`     | ✅   | Safe (normal push)   |
| `git pull`     | ✅   | Safe                 |
| `git fetch`    | ✅   | Safe                 |
| `git branch`   | ✅   | Create/list safe     |
| `git checkout` | ✅   | Switch branches safe |

---

## FORBIDDEN Operations

| Command            | Risk                    | Alternative       |
| ------------------ | ----------------------- | ----------------- |
| `git reset --hard` | Loses uncommitted work  | `git stash` first |
| `git push --force` | Destroys remote history | Never use         |
| `git push -f`      | Same as above           | Never use         |
| `git rebase`       | Rewrites history        | Merge instead     |
| `git branch -D`    | Deletes unmerged branch | Use `-d` (safe)   |
| `git clean -fd`    | Deletes untracked files | Manual cleanup    |

---

## Auto-Commit Policy (MANDATORY)

**After EVERY completed task:**

1. `git add -A` - Stage all changes
2. `git commit -m "type: description"` - Commit with meaningful message
3. `git push` - Push immediately

**NO permission required. NO confirmation needed.**

> **Exception — Hephaestus**: Hephaestus (testing agent) MUST NOT auto-commit or auto-push. Requires explicit user confirmation before any git operation. See AGENTS.md constitution.

## Before Committing

1. Check what's staged: `git status`
2. Review changes: `git diff --staged`
3. Ensure no secrets in commit
4. Write meaningful commit message

---

## Commit Message Style

```
type: short description

- Detail 1
- Detail 2
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

---

## If Something Goes Wrong

```bash
# Undo last commit (keep changes)
git reset --soft HEAD~1

# Recover deleted branch
git reflog
git checkout -b branch-name <commit-hash>

# Discard unstaged changes to file
git checkout -- <file>
```

---

## Never Commit

- `.env` files with secrets
- `credentials.json`
- API keys
- Passwords
- Personal data
