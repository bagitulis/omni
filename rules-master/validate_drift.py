#!/usr/bin/env python3
"""
validate_drift.py - Independent structural validation for rules-master.

Independently validates (does NOT trust sync_rules.py --check alone):
  1. All declared targets exist on disk
  2. All MASTER markers have open/close pairs
  3. No orphan markers outside registered targets
  4. Generated compose output matches source prompt_blocks

Usage:
  python rules-master/validate_drift.py                  # Validate this repo
  python rules-master/validate_drift.py --repo /path     # Validate another repo
  python rules-master/validate_drift.py --ci             # Exit 1 on any failure
  python rules-master/validate_drift.py --seeded         # Run seeded failure injection

Returns exit code 0 on clean, 1 on any issue.
"""

import argparse
import json
import os
import re
import sys
import tempfile

MARKER_OPEN_PATTERN = re.compile(r"<!-- MASTER:([a-z0-9_-]+) -->")
MARKER_CLOSE_PATTERN = re.compile(r"<!-- /MASTER:([a-z0-9_-]+) -->")


def load_rules(project_root):
    rules_path = os.path.join(project_root, "rules-master", "rules.json")
    if not os.path.exists(rules_path):
        print(f"  FAIL: rules.json not found at {rules_path}")
        return None
    with open(rules_path, "r", encoding="utf-8") as f:
        return json.load(f)


def is_not_metadata(key):
    return not key.startswith("_")


def parse_targets(targets):
    if "generate_full_doc" in targets or "inject_block" in targets:
        return (
            targets.get("generate_full_doc", {}),
            targets.get("inject_block", {}),
            targets.get("compose_json", {}),
        )
    gen_targets = {}
    inject_targets = {}
    compose_targets = {}
    for key, value in targets.items():
        if key.startswith("_"):
            continue
        if key == "compose_json":
            if isinstance(value, dict):
                compose_targets = value
            continue
        if isinstance(value, dict) and "mode" in value:
            mode = value.get("mode", "inject_block")
            if mode == "generate_full_doc":
                gen_targets[key] = value
            elif mode == "inject_block":
                markers = value.get("markers", [])
                inject_targets[key] = {m: {} for m in markers}
    return gen_targets, inject_targets, compose_targets


def get_all_target_paths(gen_targets, inject_targets, compose_targets):
    paths = set()
    for fp in list(gen_targets.keys()) + list(inject_targets.keys()) + list(compose_targets.keys()):
        if is_not_metadata(fp):
            paths.add(fp)
    return sorted(paths)


def resolve_path(fp, project_root):
    if fp.startswith("~"):
        return os.path.expanduser(fp)
    return os.path.join(project_root, fp)


def check_target_existence(gen_targets, inject_targets, compose_targets, project_root):
    missing = []
    for fp in get_all_target_paths(gen_targets, inject_targets, compose_targets):
        abs_path = resolve_path(fp, project_root)
        if not os.path.exists(abs_path):
            mode = "unknown"
            if fp in gen_targets:
                mode = "generate_full_doc"
            elif fp in inject_targets:
                mode = "inject_block"
            elif fp in compose_targets:
                mode = "compose_json"
            missing.append((fp, mode, abs_path))
    return missing


def collect_registered_keys(gen_targets, inject_targets, blocks):
    registered = set()
    for fp, config in inject_targets.items():
        if not is_not_metadata(fp):
            continue
        registered.update(k for k in config.keys() if is_not_metadata(k))
    if blocks:
        registered.update(blocks.keys())
    return registered


def scan_orphan_markers(gen_targets, inject_targets, project_root, blocks):
    registered = collect_registered_keys(gen_targets, inject_targets, blocks)
    orphans = []
    all_fps = get_all_target_paths(gen_targets, inject_targets, {})
    for fp in all_fps:
        abs_path = resolve_path(fp, project_root)
        if not os.path.exists(abs_path):
            continue
        with open(abs_path, "r", encoding="utf-8") as f:
            content = f.read()
        found = set(MARKER_OPEN_PATTERN.findall(content))
        unregistered = found - registered
        for key in sorted(unregistered):
            orphans.append((fp, key))
    return orphans


def scan_unclosed_markers(gen_targets, inject_targets, project_root):
    unclosed = []
    all_fps = get_all_target_paths(gen_targets, inject_targets, {})
    for fp in all_fps:
        abs_path = resolve_path(fp, project_root)
        if not os.path.exists(abs_path):
            continue
        with open(abs_path, "r", encoding="utf-8") as f:
            content = f.read()
        open_keys = set(MARKER_OPEN_PATTERN.findall(content))
        close_keys = set(MARKER_CLOSE_PATTERN.findall(content))
        for key in sorted(open_keys):
            close_marker = "<!-- /MASTER:" + key + " -->"
            if close_marker not in content:
                unclosed.append((fp, key, "open_without_close"))
        for key in sorted(close_keys):
            open_marker = "<!-- MASTER:" + key + " -->"
            if open_marker not in content:
                unclosed.append((fp, key, "close_without_open"))
    return unclosed


def compose_prompt(block_keys, prompt_blocks):
    parts = []
    for key in block_keys:
        if key.startswith("_"):
            continue
        if key not in prompt_blocks:
            return None, "Unknown prompt_block '" + key + "'"
        parts.append(prompt_blocks[key])
    return "\n\n".join(parts), None


def check_compose_drift(prompt_blocks, prompt_compose, profiles_path, project_root):
    issues = []
    abs_path = resolve_path(profiles_path, project_root)
    if not os.path.exists(abs_path):
        return ["Profile not found: " + profiles_path]
    with open(abs_path, "r", encoding="utf-8") as f:
        try:
            data = json.load(f)
        except json.JSONDecodeError as e:
            return ["Invalid JSON in " + profiles_path + ": " + str(e)]
    shared = data.get("shared", {})
    agents_compose = prompt_compose.get("agents", {})
    shared_agents = shared.get("agents", {})
    for agent_name, block_keys in agents_compose.items():
        if agent_name.startswith("_"):
            continue
        expected, err = compose_prompt(block_keys, prompt_blocks)
        if err:
            issues.append("  agent '" + agent_name + "': " + err)
            continue
        actual = shared_agents.get(agent_name, {}).get("prompt_append", "")
        if actual != expected:
            issues.append(
                "  agent '" + agent_name + "': prompt_append drift "
                + "(" + str(len(actual)) + " actual vs " + str(len(expected)) + " expected chars)"
            )
    cats_compose = prompt_compose.get("categories", {})
    shared_cats = shared.get("categories", {})
    for cat_name, block_keys in cats_compose.items():
        if cat_name.startswith("_"):
            continue
        expected, err = compose_prompt(block_keys, prompt_blocks)
        if err:
            issues.append("  category '" + cat_name + "': " + err)
            continue
        actual = shared_cats.get(cat_name, {}).get("prompt_append", "")
        if actual != expected:
            issues.append(
                "  category '" + cat_name + "': prompt_append drift "
                + "(" + str(len(actual)) + " actual vs " + str(len(expected)) + " expected chars)"
            )
    return issues


def validate_repo(project_root, ci_mode=False):
    print("\n" + ("=" * 60))
    print("VALIDATING: " + project_root)
    print("=" * 60)

    data = load_rules(project_root)
    if data is None:
        return False

    blocks = data.get("blocks", {})
    targets = data.get("targets", {})
    prompt_blocks = data.get("prompt_blocks", {})
    prompt_compose = data.get("prompt_compose", {})

    gen_targets, inject_targets, compose_targets = parse_targets(targets)
    all_ok = True

    print("\n[1] Checking declared target existence...")
    missing = check_target_existence(gen_targets, inject_targets, compose_targets, project_root)
    if missing:
        all_ok = False
        for fp, mode, ap in missing:
            print("  FAIL: Declared target '" + fp + "' (" + mode + ") not found at " + ap)
    else:
        print("  OK: All declared targets exist on disk")

    print("\n[2] Checking for orphan markers...")
    orphans = scan_orphan_markers(gen_targets, inject_targets, project_root, blocks)
    if orphans:
        all_ok = False
        for fp, key in orphans:
            print("  FAIL: Orphan marker <!-- MASTER:" + key + " --> in " + fp + " (not registered)")
    else:
        print("  OK: No orphan markers found")

    print("\n[3] Checking for unclosed markers...")
    unclosed = scan_unclosed_markers(gen_targets, inject_targets, project_root)
    if unclosed:
        all_ok = False
        for item in unclosed:
            if len(item) == 3:
                fp, key, reason = item
                if reason == "close_without_open":
                    print("  FAIL: Close <!-- /MASTER:" + key + " --> in " + fp + " without open marker")
                else:
                    print("  FAIL: Open <!-- MASTER:" + key + " --> in " + fp + " without closing marker")
            else:
                fp, key = item
                print("  FAIL: Open marker <!-- MASTER:" + key + " --> in " + fp + " without closing marker")
    else:
        print("  OK: All markers have proper open/close pairs")

    print("\n[4] Checking compose output drift...")
    for fp in compose_targets:
        if not is_not_metadata(fp):
            continue
        issues = check_compose_drift(prompt_blocks, prompt_compose, fp, project_root)
        if issues:
            all_ok = False
            print("  FAIL: Compose drift in " + fp + ":")
            for issue in issues:
                print("    " + issue)
        else:
            print("  OK: " + fp + " compose output matches source prompt_blocks")

    print("\n" + ("-" * 60))
    print("RESULT: ALL CHECKS PASSED" if all_ok else "RESULT: SOME CHECKS FAILED")
    print("-" * 60)
    return all_ok


def inject_seeded_failures(project_root, tmp_dir):
    """
    Create seeded failures using temp files and verify validators catch them.
    All temp files are written to tmp_dir to avoid permission issues with
    read-only target directories.
    """
    print("\n" + ("=" * 60))
    print("SEEDED FAILURE TESTS: " + project_root)
    print("=" * 60)

    results = {}

    data = load_rules(project_root)
    if data is None:
        return {"load_rules": "FAIL"}

    targets = data.get("targets", {})
    blocks = data.get("blocks", {})
    prompt_blocks = data.get("prompt_blocks", {})
    prompt_compose = data.get("prompt_compose", {})

    gen_targets, inject_targets, compose_targets = parse_targets(targets)

    # --- Seeded 1: Missing target ---
    print("\n  [Seeded 1] Missing declared target...")
    fake_key = "ztest-missing-target-file.md"
    fake_inject = dict(inject_targets)
    fake_inject[fake_key] = {"ztest_key": {}}
    new_missing = check_target_existence(gen_targets, fake_inject, compose_targets, project_root)
    detected = any(fake_key in m[0] for m in new_missing)
    if detected:
        print("  DETECTED - validator returns nonzero for missing declared target")
        results["missing_target"] = "DETECTED"
    else:
        print("  NOT DETECTED - validator did not catch missing target")
        results["missing_target"] = "NOT_DETECTED"

    # --- Seeded 2: Orphan marker ---
    # Write test file to tmp_dir, point inject_targets there
    print("\n  [Seeded 2] Orphan marker...")
    test_orphan_path = os.path.join(tmp_dir, "ztest-orphan-file.md")
    with open(test_orphan_path, "w") as f:
        f.write("Some content\n<!-- MASTER:ztest-orphan-key -->\nMore content\n")
    # Use a fake inject_targets pointing to the tmp path
    fake_inject2 = {"ztest-orphan-file.md": {}}
    orphans = scan_orphan_markers({}, fake_inject2, tmp_dir, blocks)
    detected = any("ztest-orphan-key" in str(o) for o in orphans)
    if detected:
        print("  DETECTED - validator catches orphan marker")
        results["orphan_marker"] = "DETECTED"
    else:
        print("  NOT DETECTED - validator missed orphan marker")
        results["orphan_marker"] = "NOT_DETECTED"

    # --- Seeded 3: Missing close marker ---
    print("\n  [Seeded 3] Missing close marker...")
    test_unclosed_path = os.path.join(tmp_dir, "ztest-unclosed-file.md")
    with open(test_unclosed_path, "w") as f:
        f.write("Before\n<!-- MASTER:ztest-unclosed-key -->\nNo closing marker\n")
    fake_inject3 = {"ztest-unclosed-file.md": {"ztest-unclosed-key": {}}}
    unclosed = scan_unclosed_markers({}, fake_inject3, tmp_dir)
    detected = any("ztest-unclosed-key" in str(u) for u in unclosed)
    if detected:
        print("  DETECTED - validator catches missing close marker")
        results["unclosed_marker"] = "DETECTED"
    else:
        print("  NOT DETECTED - validator missed missing close marker")
        results["unclosed_marker"] = "NOT_DETECTED"

    # --- Seeded 4: Compose drift ---
    print("\n  [Seeded 4] Compose drift...")
    if prompt_blocks and prompt_compose:
        compose_fp_list = [fp for fp in compose_targets if is_not_metadata(fp)]
        if compose_fp_list:
            profiles_path = compose_fp_list[0]
            abs_profile = resolve_path(profiles_path, project_root)
            if os.path.exists(abs_profile):
                # Inject a drift via modified prompt_compose that references unknown block
                for agent_name in prompt_compose.get("agents", {}):
                    if not agent_name.startswith("_"):
                        drifted_compose = dict(prompt_compose)
                        drifted_compose["agents"] = dict(prompt_compose["agents"])
                        drifted_compose["agents"][agent_name] = ["ztest-undefined-block"]
                        issues = check_compose_drift(
                            prompt_blocks, drifted_compose, profiles_path, project_root
                        )
                        if issues:
                            print("  DETECTED - compose drift caught by validator")
                            results["compose_drift"] = "DETECTED"
                            break
                else:
                    print("  NOT DETECTED")
                    results["compose_drift"] = "NOT_DETECTED"
            else:
                print("  SKIP: Profile not found")
        else:
            print("  SKIP: No compose targets")
    else:
        print("  SKIP: No compose data available")

    print()
    for k, v in sorted(results.items()):
        icon = "PASS" if v == "DETECTED" else "FAIL"
        print("  [" + icon + "] " + k + ": " + v)

    n_detected = sum(1 for v in results.values() if v == "DETECTED")
    print("\n  Summary: " + str(n_detected) + "/" + str(len(results)) + " seeded failures detected")
    return results


def main():
    parser = argparse.ArgumentParser(description="Independent structural validation for rules-master")
    parser.add_argument("--repo", type=str, default=".", help="Project root path")
    parser.add_argument("--ci", action="store_true", help="Exit 1 on failure")
    parser.add_argument("--seeded", action="store_true", help="Run seeded failure injection tests")
    args = parser.parse_args()

    project_root = os.path.abspath(args.repo)
    all_ok = validate_repo(project_root, ci_mode=args.ci)

    if args.seeded:
        with tempfile.TemporaryDirectory(prefix="validate_drift_") as tmp_dir:
            inject_seeded_failures(project_root, tmp_dir)

    if args.ci:
        sys.exit(0 if all_ok else 1)


if __name__ == "__main__":
    main()
