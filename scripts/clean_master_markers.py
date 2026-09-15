"""
Strip `<!-- MASTER:key -->` and `<!-- /MASTER:key -->` marker lines from
markdown files, keeping the content between them intact.

Rationale: rules-master sync system was removed. The generated content is
frozen in place; only the marker lines need to disappear so a future reader
does not think the file is still auto-generated.

Safe by design:
  * Only lines whose ENTIRE trimmed content matches `<!-- MASTER:*-->` or
    `<!-- /MASTER:*-->` are removed. Inline markers are left untouched.
  * Idempotent: running twice on a clean file is a no-op.
  * Reports every modified path and the number of markers stripped.

Usage:
  python scripts/clean_master_markers.py                 # apply to defaults
  python scripts/clean_master_markers.py --dry-run       # show what would change
  python scripts/clean_master_markers.py FILE [FILE...]  # explicit targets
"""
from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent

# Files retained after the cleanup that still contain MASTER markers.
DEFAULT_TARGETS = [
    "AGENTS.MD",
    "SISYPHUS_RULES.md",
    "PROMETHEUS_RULES.md",
    "EXECUTOR_RULES.md",
    "backend/internal/handlers/AGENTS.md",
    "backend/internal/services/AGENTS.md",
    ".claude/rules/executor-rules.md",
    ".claude/rules/prometheus-rules.md",
    ".claude/rules/sisyphus-rules.md",
    ".claude/rules/project-context.md",
]

# Whole-line marker: `<!-- MASTER:xxx -->` or `<!-- /MASTER:xxx -->`,
# with optional leading/trailing whitespace.
MARKER_LINE = re.compile(r"^\s*<!--\s*/?MASTER:[^>]+-->\s*$")


def strip_markers(text: str) -> tuple[str, int]:
    kept: list[str] = []
    removed = 0
    for line in text.splitlines(keepends=True):
        # `.splitlines(keepends=True)` returns each line with its trailing
        # newline (if any). Strip it only for the regex check so leading
        # whitespace is preserved in the retained lines.
        stripped_for_match = line.rstrip("\r\n")
        if MARKER_LINE.match(stripped_for_match):
            removed += 1
            continue
        kept.append(line)
    return "".join(kept), removed


def process(paths: list[Path], dry_run: bool) -> int:
    total_removed = 0
    files_touched = 0
    for p in paths:
        if not p.is_file():
            print(f"[SKIP] not a file: {p}")
            continue
        original = p.read_text(encoding="utf-8")
        cleaned, removed = strip_markers(original)
        if removed == 0:
            print(f"[OK] no markers: {p.relative_to(ROOT)}")
            continue
        files_touched += 1
        total_removed += removed
        if dry_run:
            print(
                f"[DRY-RUN] would strip {removed} marker line(s) from "
                f"{p.relative_to(ROOT)}"
            )
        else:
            p.write_text(cleaned, encoding="utf-8")
            print(
                f"[WRITE] stripped {removed} marker line(s) from "
                f"{p.relative_to(ROOT)}"
            )

    verb = "would strip" if dry_run else "stripped"
    print("-" * 60)
    print(f"{verb} {total_removed} marker line(s) across {files_touched} file(s)")
    return 0


def main() -> int:
    parser = argparse.ArgumentParser(description="Strip MASTER markers from markdown")
    parser.add_argument("--dry-run", action="store_true", help="show without writing")
    parser.add_argument(
        "files",
        nargs="*",
        help="explicit files to process (defaults to the built-in target list)",
    )
    args = parser.parse_args()

    targets = args.files or DEFAULT_TARGETS
    paths = [ROOT / t if not Path(t).is_absolute() else Path(t) for t in targets]
    return process(paths, args.dry_run)


if __name__ == "__main__":
    sys.exit(main())
