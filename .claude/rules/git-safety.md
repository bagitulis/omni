---
alwaysApply: true
description: Git operation safety rules - STRICT restricted operations
---

# Git Safety Rules (STRICT)

> **CRITICAL**: AI agents are ONLY allowed to run the commands listed below.
> ANY git command not explicitly listed here is FORBIDDEN.
> This rule was enacted after an incident where Oracle recommended `git rm --cached` which deleted tracked backup files.

---

## ALLOWED Operations (Exhaustive List)

### Write Operations (Auto-Commit Policy)

| Command      | Purpose           | Notes                           |
| ------------ | ----------------- | ------------------------------- |
| `git add`    | Stage changes     | `git add -A` for all changes    |
| `git add -A` | Stage all changes | Preferred for commits           |
| `git commit` | Create commit     | Always with `-m "message"`      |
| `git push`   | Push to remote    | Normal push only, never --force |

### Read-Only Operations

| Command      | Purpose              | Notes           |
| ------------ | -------------------- | --------------- |
| `git status` | View working tree    | Always safe     |
| `git diff`   | View changes         | Always safe     |
| `git log`    | View commit history  | Always safe     |
| `git show`   | View commit contents | Always safe     |
| `git blame`  | View line authorship | Always safe     |
| `git branch` | List branches        | List only, safe |

---

## FORBIDDEN Operations (ZERO EXCEPTIONS)

**ANY command not in the ALLOWED list above is FORBIDDEN. Key banned commands include:**

| Command                  | Why Banned                                            |
| ------------------------ | ----------------------------------------------------- |
| `git rm`                 | Deletes tracked files — caused backup loss            |
| `git rm --cached`        | Untracks files — caused backup loss                   |
| `git reset`              | Any form (--soft, --mixed, --hard) — rewrites history |
| `git reset --hard`       | Loses uncommitted work permanently                    |
| `git reset --soft`       | Rewrites commit history                               |
| `git checkout -- <file>` | Discards file changes — data loss risk                |
| `git restore`            | Discards changes — data loss risk                     |
| `git clean`              | Deletes untracked files permanently                   |
| `git push --force`       | Destroys remote history                               |
| `git push -f`            | Same as above                                         |
| `git rebase`             | Rewrites history                                      |
| `git branch -D`          | Force-deletes branch                                  |
| `git stash`              | Hides changes — can cause confusion                   |
| `git cherry-pick`        | Rewrites history                                      |
| `git revert`             | Creates revert commits — ask user first               |
| `git merge`              | Can cause conflicts — ask user first                  |
| `git pull`               | Can trigger merge conflicts — ask user first          |
| `git fetch`              | Generally safe but not needed for AI workflow         |

---

## Auto-Commit Policy (MANDATORY)

**After EVERY completed task unit:**

1. `git add -A` — Stage ALL changes (no cherry-picking)
2. `git commit -m "type(scope): description"` — Meaningful message
3. `git push` — Push immediately

**NO permission required. NO confirmation needed.**

> **Exception — Hephaestus**: Requires explicit user confirmation before any git operation.

---

## Commit Message Style

```
type(scope): short description
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

---

## Oracle / Advisory Agents — Git Restrictions

- Oracle is a **READ-ONLY advisor**
- Oracle may **recommend** git operations, but recommendations MUST be confirmed with the user before execution
- NO advisory agent (Oracle, Librarian, Explore) may execute git write commands
- If Oracle recommends removing/untracking files: **STOP and ask user first**

---

## If Something Goes Wrong

**DO NOT attempt to fix git issues yourself.**

1. Report the issue to the user
2. Describe what happened
3. Let the user decide the recovery action
4. NEVER run recovery commands (reset, checkout --, restore, clean)

---

## Never Commit

- `.env` files with secrets
- `credentials.json`
- API keys, passwords, tokens
- Personal data

## Intentionally Tracked Files (DO NOT REMOVE)

These files/directories are tracked in git intentionally by the user, even if `.gitignore` patterns match them:

- `backups/smart/` — User's database backups
- `nginx/logs/error.log` — Nginx error logs

**NEVER run `git rm` or `git rm --cached` on these paths.**
