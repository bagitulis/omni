#!/usr/bin/env python3
"""
setup_claude_projects.py — Auto-setup Claude Code config for all projects.

Detects project type and generates appropriate:
- .claude/rules/project-context.md (project-specific rules)
- .claude/rules/git-safety.md (shared git safety rules)
- .claude/settings.local.json (minimal, project-specific)
- .claude/workflows/*.js (useful workflows)

Usage:
  python scripts/setup_claude_projects.py              # Setup all projects
  python scripts/setup_claude_projects.py --project X  # Setup specific project
  python scripts/setup_claude_projects.py --dry-run    # Preview only
"""

import json
import os
import shutil
import sys

PROJECTS_DIR = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))

PROJECTS = {
    "automa": {
        "type": "fullstack",
        "lang": ["go", "typescript"],
        "description": "Web automation platform with Go backend and web frontend",
        "rules": ["go-rules", "web-rules", "automation-rules"],
        "workflows": ["scan-bugs", "review-changes"],
    },
    "captcha-resolver": {
        "type": "python-ml",
        "lang": ["python"],
        "description": "CAPTCHA solving system with ML models and distributed workers",
        "rules": ["python-rules", "ml-safety", "distributed-systems"],
        "workflows": ["scan-bugs", "test-coverage"],
    },
    "enaproxy": {
        "type": "go-cli",
        "lang": ["go"],
        "description": "Proxy router and endpoint management tool",
        "rules": ["go-rules", "cli-patterns"],
        "workflows": ["scan-bugs", "review-changes"],
    },
    "extensions": {
        "type": "browser-extension",
        "lang": ["typescript", "javascript"],
        "description": "Browser extensions for various platforms",
        "rules": ["extension-rules", "browser-apis", "security-rules"],
        "workflows": ["scan-bugs", "security-audit"],
    },
    "hermes-sync": {
        "type": "coordination",
        "lang": ["python", "bash"],
        "description": "Sync coordination and scheduling tool",
        "rules": ["sync-rules", "cron-safety"],
        "workflows": ["scan-bugs"],
    },
}

# Git safety rules (shared across all projects)
GIT_SAFETY = """---
alwaysApply: true
description: Git operation safety rules - STRICT restricted operations
---

# Git Safety Rules (STRICT)

> **CRITICAL**: AI agents are ONLY allowed to run the commands listed below.
> ANY git command not explicitly listed here is FORBIDDEN.

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

| Command                  | Why Banned                                            |
| ------------------------ | ----------------------------------------------------- |
| `git rm`                 | Deletes tracked files                                 |
| `git rm --cached`        | Untracks files                                        |
| `git reset`              | Any form — rewrites history                           |
| `git reset --hard`       | Loses uncommitted work permanently                    |
| `git checkout -- <file>` | Discards file changes — data loss risk                |
| `git restore`            | Discards changes — data loss risk                     |
| `git clean`              | Deletes untracked files permanently                   |
| `git push --force`       | Destroys remote history                               |
| `git rebase`             | Rewrites history                                      |
| `git branch -D`          | Force-deletes branch                                  |
| `git stash`              | Hides changes                                         |
| `git cherry-pick`        | Rewrites history                                      |
| `git revert`             | Creates revert commits — ask user first               |
| `git merge`              | Can cause conflicts — ask user first                  |
| `git pull`               | Can trigger merge conflicts — ask user first          |

---

## Auto-Commit Policy (MANDATORY)

**After EVERY completed task unit:**

1. `git add -A` — Stage ALL changes (no cherry-picking)
2. `git commit -m "type(scope): description"` — Meaningful message
3. `git push` — Push immediately

**NO permission required. NO confirmation needed.**

---

## Commit Message Style

```
type(scope): short description
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`
"""


def generate_project_context(project_name, config):
    """Generate project-specific context rules."""
    return f"""---
alwaysApply: true
description: Project context and coding standards for {project_name}
---

# {project_name.title()} — Project Context

> **Type**: {config['type']}
> **Languages**: {', '.join(config['lang'])}
> **Description**: {config['description']}

## Code Quality Standards

| Rule | Guideline |
| --- | --- |
| **File size** | ~300 lines per file (refactor if exceeded) |
| **Functions** | Single responsibility, <50 lines |
| **Comments** | Explain WHY, not WHAT |
| **Error handling** | Never silently fail — log or propagate |
| **Testing** | Write tests for new features |

## Language-Specific Rules

{generate_lang_rules(config['lang'])}

## Project-Specific Rules

{generate_type_rules(config['type'], config['lang'])}

## Git Workflow

- Follow auto-commit policy (see git-safety.md)
- Commit message format: `type(scope): description`
- Push after every task completion
- Never force-push without explicit permission
"""


def generate_lang_rules(langs):
    """Generate language-specific coding standards."""
    rules = []

    if "go" in langs:
        rules.append("""### Go Rules
- Use `gofmt` for formatting
- Error handling: Always check errors, use `fmt.Errorf("context: %w", err)`
- Naming: CamelCase for exported, camelCase for unexported
- Interfaces: Prefer small, composable interfaces
- Packages: One package per directory, clear package names
- Run `go vet ./...` and `go build ./...` before commit""")

    if "python" in langs:
        rules.append("""### Python Rules
- Use `black` or `ruff` for formatting
- Type hints: Add to function signatures
- Docstrings: Google style for functions
- Error handling: Use specific exceptions, not bare `except:`
- Imports: Group standard library, third-party, local
- Run `python -m py_compile` before commit""")

    if "typescript" in langs or "javascript" in langs:
        rules.append("""### TypeScript/JavaScript Rules
- Use `prettier` for formatting
- TypeScript: Prefer `interface` over `type` for objects
- Error handling: Use try/catch, never empty catch blocks
- Naming: camelCase for variables/functions, PascalCase for classes
- Imports: Use ES modules (`import` not `require`)
- Run `npm run lint` and `npm run build` before commit""")

    return "\n\n".join(rules) if rules else "No language-specific rules defined."


def generate_type_rules(project_type, langs):
    """Generate project-type-specific rules."""
    if project_type == "fullstack":
        return """### Full-Stack Project Rules
- **Backend**: API-first design, RESTful endpoints
- **Frontend**: Component-based architecture, reusable components
- **Integration**: Backend types must match frontend types exactly
- **Testing**: Unit tests for backend, integration tests for API
- **Deployment**: Docker Compose for local development"""

    elif project_type == "python-ml":
        return """### Python ML Project Rules
- **Models**: Version ML models separately from code
- **Data**: Never commit large datasets (>10MB)
- **Experiments**: Document hyperparameters and results
- **Dependencies**: Pin versions in requirements.txt
- **Safety**: Test ML models on small sample before production"""

    elif project_type == "go-cli":
        return """### Go CLI Project Rules
- **Flags**: Use `cobra` or `flag` package for CLI args
- **Output**: Use structured logging (JSON for machines, pretty for humans)
- **Error codes**: Return meaningful exit codes (0=success, 1=error, 2=invalid args)
- **Help**: Always provide `--help` flag with examples
- **Config**: Support config file + environment variables + flags"""

    elif project_type == "browser-extension":
        return """### Browser Extension Rules
- **Manifest**: Keep manifest.json minimal and secure
- **Permissions**: Request only necessary permissions
- **Content scripts**: Isolate from page scripts
- **Background**: Use service workers (MV3) not background pages
- **Security**: Never eval(), sanitize all user input
- **Testing**: Test on Chrome, Firefox, Edge"""

    elif project_type == "coordination":
        return """### Coordination Tool Rules
- **Idempotency**: All operations must be idempotent (safe to retry)
- **Logging**: Log every state transition
- **Locking**: Use distributed locks for concurrent operations
- **Monitoring**: Expose health check endpoint
- **Recovery**: Implement graceful shutdown and recovery"""

    return "No type-specific rules defined."


def generate_settings(project_name, config):
    """Generate project-specific settings."""
    # Base permissions (already in global, but add project-specific)
    permissions = []

    # Add project-specific tools
    if "python" in config["lang"]:
        permissions.extend([
            "Bash(pip *)",
            "Bash(pipenv *)",
            "Bash(poetry *)",
            "Bash(pytest *)",
            "Bash(black *)",
            "Bash(ruff *)",
        ])

    if "go" in config["lang"]:
        permissions.extend([
            "Bash(go build *)",
            "Bash(go test *)",
            "Bash(go vet *)",
            "Bash(go fmt *)",
            "Bash(golangci-lint *)",
        ])

    if "typescript" in config["lang"] or "javascript" in config["lang"]:
        permissions.extend([
            "Bash(npm run *)",
            "Bash(yarn *)",
            "Bash(pnpm *)",
            "Bash(eslint *)",
            "Bash(prettier *)",
            "Bash(jest *)",
            "Bash(vitest *)",
        ])

    return {
        "permissions": {
            "allow": permissions
        }
    }


def setup_project(project_name, config, dry_run=False):
    """Setup Claude Code config for a single project."""
    project_path = os.path.join(PROJECTS_DIR, project_name)

    if not os.path.exists(project_path):
        print(f"  SKIP: {project_name} — directory not found")
        return False

    print(f"\n{'='*60}")
    print(f"Setting up: {project_name}")
    print(f"  Type: {config['type']}")
    print(f"  Languages: {', '.join(config['lang'])}")
    print(f"{'='*60}")

    # Create .claude directory
    claude_dir = os.path.join(project_path, ".claude")
    rules_dir = os.path.join(claude_dir, "rules")
    workflows_dir = os.path.join(claude_dir, "workflows")

    if not dry_run:
        os.makedirs(rules_dir, exist_ok=True)
        os.makedirs(workflows_dir, exist_ok=True)

    # 1. Generate project-context.md
    context_file = os.path.join(rules_dir, "project-context.md")
    if dry_run or not os.path.exists(context_file):
        print(f"  [+] Create: .claude/rules/project-context.md")
        if not dry_run:
            with open(context_file, "w", encoding="utf-8") as f:
                f.write(generate_project_context(project_name, config))
    else:
        print(f"  [=] Exists: .claude/rules/project-context.md")

    # 2. Generate git-safety.md
    git_file = os.path.join(rules_dir, "git-safety.md")
    if dry_run or not os.path.exists(git_file):
        print(f"  [+] Create: .claude/rules/git-safety.md")
        if not dry_run:
            with open(git_file, "w", encoding="utf-8") as f:
                f.write(GIT_SAFETY)
    else:
        print(f"  [=] Exists: .claude/rules/git-safety.md")

    # 3. Generate settings.local.json
    settings_file = os.path.join(claude_dir, "settings.local.json")
    if dry_run or not os.path.exists(settings_file):
        print(f"  [+] Create: .claude/settings.local.json")
        if not dry_run:
            with open(settings_file, "w", encoding="utf-8") as f:
                json.dump(generate_settings(project_name, config), f, indent=2)
    else:
        print(f"  [=] Exists: .claude/settings.local.json")

    print(f"  ✓ Setup complete for {project_name}")
    return True


def main():
    import argparse
    parser = argparse.ArgumentParser(description="Setup Claude Code for projects")
    parser.add_argument("--project", type=str, help="Setup specific project only")
    parser.add_argument("--dry-run", action="store_true", help="Preview without writing")
    args = parser.parse_args()

    if args.project:
        if args.project not in PROJECTS:
            print(f"ERROR: Unknown project '{args.project}'")
            print(f"Available: {', '.join(PROJECTS.keys())}")
            sys.exit(1)
        setup_project(args.project, PROJECTS[args.project], dry_run=args.dry_run)
    else:
        print("Setting up Claude Code for all projects...")
        for name, config in PROJECTS.items():
            setup_project(name, config, dry_run=args.dry_run)

    print("\n" + "="*60)
    print("✅ Setup complete!")
    if args.dry_run:
        print("   (dry-run mode — no files written)")
    else:
        print("   Next: cd into each project and test with Claude Code")


if __name__ == "__main__":
    main()
