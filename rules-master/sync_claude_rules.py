#!/usr/bin/env python3
"""
sync_claude_rules.py — Auto-sync Claude Code rules from rules-master.

Convenience wrapper that ONLY syncs .claude/rules/*.md targets.
Use this as a git hook, CI step, or manual trigger to keep
Claude Code rules in sync with the master source of truth.

Usage:
  python rules-master/sync_claude_rules.py              # Sync .claude/rules only
  python rules-master/sync_claude_rules.py --check      # CI mode (exit 1 if drift)
  python rules-master/sync_claude_rules.py --dry-run    # Preview only
"""

import subprocess
import sys
import os

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
SYNC_SCRIPT = os.path.join(SCRIPT_DIR, "sync_rules.py")

CLAUDE_TARGETS = [
    ".claude/rules/delegation-rules.md",
    ".claude/rules/code-quality.md",
    ".claude/rules/react-rules.md",
]


def main():
    # Pass through all args but add --target for each Claude file
    extra_args = []
    passthrough = []
    for arg in sys.argv[1:]:
        if arg in ("--check", "--dry-run"):
            passthrough.append(arg)
        else:
            extra_args.append(arg)

    any_changed = False
    has_error = False

    for target in CLAUDE_TARGETS:
        cmd = [sys.executable, SYNC_SCRIPT] + passthrough + ["--target", target]
        result = subprocess.run(cmd, capture_output=False)
        if result.returncode != 0:
            has_error = True
            if "--check" in passthrough:
                any_changed = True

    if has_error and "--check" in passthrough:
        print("\n❌ Claude Code rules are out of sync!")
        print("   Run: python rules-master/sync_claude_rules.py")
        sys.exit(1)
    else:
        print("\n✅ Claude Code rules are in sync.")
        sys.exit(0)


if __name__ == "__main__":
    main()
