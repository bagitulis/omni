"""
Verification test for opencode/rules-master cleanup.

TDD gate — run BEFORE cleanup (must FAIL for the right reasons)
and AFTER cleanup (must PASS). Ephemeral: delete after commit.

Checks (all MUST pass after cleanup):
  1. Removed dirs are gone: rules-master/, opencode-configs/, .opencode/
  2. Legacy backup file is gone: AGENTS.MD.bak-20260915-before-rag-remove
  3. No `<!-- MASTER:xxx -->` markers remain in retained *.md files
  4. No live path references to deleted dirs in retained docs
     (CLAUDE.md, AGENTS.MD, ARCHITECTURE.md, DELEGATION_RULES.md)
  5. Retained rules dirs still exist: .claude/rules/, .agents/, backend/AGENTS.md
"""
from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

REMOVED_DIRS = ["rules-master", "opencode-configs", ".opencode"]
REMOVED_FILES = ["AGENTS.MD.bak-20260915-before-rag-remove"]
RETAINED_DIRS = [".claude/rules", ".agents", "backend"]
DOC_FILES_MUST_BE_CLEAN = [
    "CLAUDE.md",
    "AGENTS.MD",
    "ARCHITECTURE.md",
    "DELEGATION_RULES.md",
    "SISYPHUS_RULES.md",
    "PROMETHEUS_RULES.md",
    "EXECUTOR_RULES.md",
    "RULES.md",
    ".claude/rules/architecture.md",
    ".claude/rules/react-rules.md",
    ".claude/rules/delegation-rules.md",
    ".claude/rules/code-quality.md",
    ".claude/rules/project-context.md",
    ".claude/rules/executor-rules.md",
    ".claude/rules/prometheus-rules.md",
    ".claude/rules/sisyphus-rules.md",
]
MARKER_RE = re.compile(r"<!--\s*/?MASTER:", re.IGNORECASE)
# Live path patterns that must not appear in retained docs (case-insensitive).
# Historical mentions in CHANGELOG are allowed (immutable history).
DELETED_PATH_PATTERNS = [
    r"\brules-master/",
    r"\brules-master\\",
    r"\bopencode-configs/",
    r"\bopencode-configs\\",
    r"\.opencode/",
    r"\.opencode\\",
]

failures: list[str] = []


def _fail(msg: str) -> None:
    failures.append(msg)


def check_removed_dirs() -> None:
    for rel in REMOVED_DIRS:
        p = ROOT / rel
        if p.exists():
            _fail(f"FAIL: directory should be deleted but still exists: {rel}")


def check_removed_files() -> None:
    for rel in REMOVED_FILES:
        p = ROOT / rel
        if p.exists():
            _fail(f"FAIL: file should be deleted but still exists: {rel}")


def check_retained_dirs() -> None:
    for rel in RETAINED_DIRS:
        p = ROOT / rel
        if not p.exists():
            _fail(f"FAIL: retained directory missing: {rel}")


def check_no_master_markers() -> None:
    """Every .md in the retained tree must be free of <!-- MASTER:*--> markers."""
    for md in ROOT.rglob("*.md"):
        # Skip anything inside soon-to-be-removed dirs (they may still be
        # present before delete-step runs but should be cleaned).
        rel = md.relative_to(ROOT).as_posix()
        if any(rel.startswith(f"{d}/") for d in REMOVED_DIRS):
            continue
        # Skip venv, node_modules, and codegraph internals.
        if any(seg in rel for seg in (".venv/", "node_modules/", ".codegraph/",
                                       "backups/", "temp_repo/")):
            continue
        try:
            text = md.read_text(encoding="utf-8", errors="ignore")
        except Exception as e:  # pragma: no cover — defensive
            _fail(f"FAIL: unreadable file {rel}: {e}")
            continue
        if MARKER_RE.search(text):
            # Report first occurrence with line number.
            for i, line in enumerate(text.splitlines(), 1):
                if MARKER_RE.search(line):
                    _fail(
                        f"FAIL: MASTER marker still present in {rel}:{i} → "
                        f"{line.strip()[:120]}"
                    )
                    break


def check_docs_clean() -> None:
    """The listed top-level docs must not reference deleted paths."""
    combined = re.compile("|".join(DELETED_PATH_PATTERNS), re.IGNORECASE)
    for rel in DOC_FILES_MUST_BE_CLEAN:
        p = ROOT / rel
        if not p.exists():
            # Missing doc is a different failure — flag it.
            _fail(f"FAIL: required doc missing: {rel}")
            continue
        text = p.read_text(encoding="utf-8", errors="ignore")
        for i, line in enumerate(text.splitlines(), 1):
            if combined.search(line):
                _fail(
                    f"FAIL: {rel}:{i} references deleted path → "
                    f"{line.strip()[:120]}"
                )


def main() -> int:
    check_removed_dirs()
    check_removed_files()
    check_retained_dirs()
    check_no_master_markers()
    check_docs_clean()

    if failures:
        print("=" * 60)
        print(f"  {len(failures)} FAILURE(S)")
        print("=" * 60)
        for f in failures:
            print(f)
        return 1

    print("=" * 60)
    print("  ALL CLEANUP CHECKS PASSED")
    print("=" * 60)
    return 0


if __name__ == "__main__":
    sys.exit(main())
