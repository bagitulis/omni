#!/usr/bin/env python3
"""
sync_rules.py - Master rules synchronization tool.

Reads rules-master/rules.json and propagates shared rule blocks to target files.

Three modes:
  1. generate_full_doc: Generates entire files from templates (e.g., DELEGATION_RULES.md)
  2. inject_block: Replaces content between <!-- MASTER:key --> markers in existing files
  3. compose_json: Composes prompt_append strings from DRY prompt_blocks into JSON config

Usage:
  python rules-master/sync_rules.py              # Sync all targets
  python rules-master/sync_rules.py --check      # Check for drift (CI mode, exit 1 if stale)
  python rules-master/sync_rules.py --dry-run    # Show what would change without writing
  python rules-master/sync_rules.py --section failure-escalation  # Sync only this block
  python rules-master/sync_rules.py --target AGENTS.md            # Sync only this file
"""

import argparse
import json
import os
import re
import sys

SCRIPT_DIR = os.path.dirname(os.path.abspath(__file__))
PROJECT_ROOT = os.path.dirname(SCRIPT_DIR)
RULES_JSON = os.path.join(SCRIPT_DIR, "rules.json")

MARKER_OPEN = "<!-- MASTER:{key} -->"
MARKER_CLOSE = "<!-- /MASTER:{key} -->"
MARKER_PATTERN = re.compile(
    r"(<!-- MASTER:([a-z0-9_-]+) -->)\n(.*?)\n(<!-- /MASTER:\2 -->)",
    re.DOTALL,
)


def load_rules():
    """Load and validate rules.json."""
    with open(RULES_JSON, "r", encoding="utf-8") as f:
        data = json.load(f)

    blocks = data.get("blocks", {})
    targets = data.get("targets", {})
    prompt_blocks = data.get("prompt_blocks", {})
    prompt_compose = data.get("prompt_compose", {})

    if not blocks:
        print("ERROR: No blocks defined in rules.json")
        sys.exit(1)

    return blocks, targets, prompt_blocks, prompt_compose


def resolve_template(template_lines, blocks):
    """Resolve $block:name references in template lines."""
    resolved = []
    for line in template_lines:
        if line.startswith("$block:"):
            block_key = line[len("$block:"):]
            if block_key not in blocks:
                print(f"  ERROR: Unknown block reference '{block_key}'")
                sys.exit(1)
            resolved.append(blocks[block_key])
        else:
            resolved.append(line)
    return "\n".join(resolved) + "\n"


def detect_line_ending(content):
    """Detect the dominant line ending in a file."""
    crlf = content.count("\r\n")
    lf = content.count("\n") - crlf
    return "\r\n" if crlf > lf else "\n"


def normalize_to_lf(content):
    """Normalize all line endings to LF for comparison."""
    return content.replace("\r\n", "\n")


def apply_line_ending(content, ending):
    """Apply consistent line ending to content."""
    content = normalize_to_lf(content)
    if ending == "\r\n":
        content = content.replace("\n", "\r\n")
    return content


def generate_full_doc(filepath, config, blocks, dry_run=False):
    """Generate an entire file from a template."""
    abs_path = os.path.join(PROJECT_ROOT, filepath)
    template = config.get("template", [])

    new_content = resolve_template(template, blocks)

    # Preserve line ending of existing file, or use OS default
    if os.path.exists(abs_path):
        with open(abs_path, "r", encoding="utf-8", newline="") as f:
            old_raw = f.read()
        ending = detect_line_ending(old_raw)
        old_normalized = normalize_to_lf(old_raw)
    else:
        ending = "\r\n" if sys.platform == "win32" else "\n"
        old_normalized = None

    new_normalized = normalize_to_lf(new_content)

    if old_normalized == new_normalized:
        return False, "up-to-date"

    if dry_run:
        return True, "would generate"

    final_content = apply_line_ending(new_content, ending)
    os.makedirs(os.path.dirname(abs_path), exist_ok=True)
    with open(abs_path, "w", encoding="utf-8", newline="") as f:
        f.write(final_content)
    return True, "generated"


def inject_blocks(filepath, block_keys, blocks, dry_run=False):
    """Replace content between MASTER markers in a file."""
    abs_path = os.path.join(PROJECT_ROOT, filepath)

    if not os.path.exists(abs_path):
        print(f"  WARNING: Target file not found: {filepath}")
        return False, "not found"

    with open(abs_path, "r", encoding="utf-8", newline="") as f:
        raw_content = f.read()

    ending = detect_line_ending(raw_content)
    content = normalize_to_lf(raw_content)
    changed = False
    missing_markers = []

    for key in block_keys:
        if key not in blocks:
            print(f"  ERROR: Block '{key}' not found in rules.json")
            sys.exit(1)

        open_marker = MARKER_OPEN.format(key=key)
        close_marker = MARKER_CLOSE.format(key=key)

        if open_marker not in content:
            missing_markers.append(key)
            continue

        # Build pattern for this specific key
        pattern = re.compile(
            re.escape(open_marker) + r"\n(.*?)\n" + re.escape(close_marker),
            re.DOTALL,
        )

        new_block = blocks[key]
        replacement = f"{open_marker}\n{new_block}\n{close_marker}"

        new_content = pattern.sub(replacement, content)
        if new_content != content:
            changed = True
            content = new_content

    if missing_markers:
        print(f"  MISSING MARKERS in {filepath}: {missing_markers}")

    if not changed:
        return False, "up-to-date"

    if dry_run:
        return True, "would update"

    final_content = apply_line_ending(content, ending)
    with open(abs_path, "w", encoding="utf-8", newline="") as f:
        f.write(final_content)
    return True, "updated"


def compose_prompt(block_keys, prompt_blocks):
    """Join prompt_blocks by key into a single prompt_append string."""
    parts = []
    for key in block_keys:
        if key.startswith("_"):
            continue
        if key not in prompt_blocks:
            print(f"  ERROR: Unknown prompt_block '{key}'")
            sys.exit(1)
        parts.append(prompt_blocks[key])
    return "\n\n".join(parts)


def compose_json(filepath, prompt_blocks, prompt_compose, dry_run=False):
    """Compose prompt_append strings from prompt_blocks into opencode-profiles.json."""
    abs_path = os.path.join(PROJECT_ROOT, filepath)

    if not os.path.exists(abs_path):
        print(f"  WARNING: Target file not found: {filepath}")
        return False, "not found"

    with open(abs_path, "r", encoding="utf-8", newline="") as f:
        raw_content = f.read()

    ending = detect_line_ending(raw_content)
    data = json.loads(raw_content)
    changed = False

    shared = data.get("shared", {})

    # Compose agents
    agents_compose = prompt_compose.get("agents", {})
    shared_agents = shared.get("agents", {})
    for agent_name, block_keys in agents_compose.items():
        if agent_name.startswith("_"):
            continue
        new_prompt = compose_prompt(block_keys, prompt_blocks)
        if agent_name not in shared_agents:
            shared_agents[agent_name] = {}
        old_prompt = shared_agents[agent_name].get("prompt_append", "")
        if old_prompt != new_prompt:
            shared_agents[agent_name]["prompt_append"] = new_prompt
            changed = True

    # Compose categories
    cats_compose = prompt_compose.get("categories", {})
    shared_cats = shared.get("categories", {})
    for cat_name, block_keys in cats_compose.items():
        if cat_name.startswith("_"):
            continue
        new_prompt = compose_prompt(block_keys, prompt_blocks)
        if cat_name not in shared_cats:
            shared_cats[cat_name] = {}
        old_prompt = shared_cats[cat_name].get("prompt_append", "")
        if old_prompt != new_prompt:
            shared_cats[cat_name]["prompt_append"] = new_prompt
            changed = True

    if not changed:
        return False, "up-to-date"

    if dry_run:
        return True, "would update"

    new_content = json.dumps(data, indent=2, ensure_ascii=False) + "\n"
    final_content = apply_line_ending(new_content, ending)
    with open(abs_path, "w", encoding="utf-8", newline="") as f:
        f.write(final_content)
    return True, "updated"


def main():
    parser = argparse.ArgumentParser(description="Sync master rules to target files")
    parser.add_argument("--check", action="store_true", help="Check mode (exit 1 if stale)")
    parser.add_argument("--dry-run", action="store_true", help="Show changes without writing")
    parser.add_argument("--section", type=str, help="Sync only this block key")
    parser.add_argument("--target", type=str, help="Sync only this target file")
    args = parser.parse_args()

    blocks, targets, prompt_blocks, prompt_compose = load_rules()
    gen_targets = targets.get("generate_full_doc", {})
    inject_targets = targets.get("inject_block", {})
    compose_targets = targets.get("compose_json", {})

    any_changed = False
    results = []

    # --- Generate full docs ---
    for filepath, config in gen_targets.items():
        if args.target and not filepath.endswith(args.target):
            continue

        # If --section specified, skip generate targets that don't use that block
        if args.section:
            template_str = "\n".join(config.get("template", []))
            if f"$block:{args.section}" not in template_str:
                continue

        changed, status = generate_full_doc(filepath, config, blocks, dry_run=args.dry_run or args.check)
        results.append((filepath, status))
        if changed:
            any_changed = True

    # --- Inject blocks ---
    for filepath, block_config in inject_targets.items():
        if args.target and not filepath.endswith(args.target):
            continue

        block_keys = list(block_config.keys())
        if args.section:
            block_keys = [k for k in block_keys if k == args.section]
            if not block_keys:
                continue

        changed, status = inject_blocks(filepath, block_keys, blocks, dry_run=args.dry_run or args.check)
        results.append((filepath, status))
        if changed:
            any_changed = True

    # --- Compose JSON (prompt_append) ---
    if not args.section:  # compose_json applies to whole file, not individual sections
        for filepath in compose_targets:
            if args.target and not filepath.endswith(args.target):
                continue

            changed, status = compose_json(
                filepath, prompt_blocks, prompt_compose,
                dry_run=args.dry_run or args.check,
            )
            results.append((filepath + " (prompt_append)", status))
            if changed:
                any_changed = True

    # --- Report ---
    print("\n--- Sync Results ---")
    for filepath, status in results:
        icon = "*" if "update" in status or "generate" in status else " "
        print(f"  [{icon}] {filepath}: {status}")

    if args.check:
        if any_changed:
            print("\nDRIFT DETECTED: Run 'python rules-master/sync_rules.py' to fix.")
            sys.exit(1)
        else:
            print("\nAll targets up-to-date.")
            sys.exit(0)

    if args.dry_run:
        print(f"\nDry run complete. {'Changes needed.' if any_changed else 'No changes needed.'}")
    else:
        print(f"\nSync complete. {'Files updated.' if any_changed else 'Everything up-to-date.'}")


if __name__ == "__main__":
    main()
