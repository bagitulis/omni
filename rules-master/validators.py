"""
validators.py - Integrity checks for rules-master sync system.

Provides:
  - scan_orphan_markers: Find MASTER markers in files not registered in rules.json
  - validate_compose_coverage: Verify all agents/categories have prompt_compose mappings
"""

import json
import os
import re

MARKER_SCAN_PATTERN = re.compile(r"<!-- MASTER:([a-z0-9_-]+) -->")

# Agents/categories that intentionally have no prompt_append (utility agents only).
# NOTE: oracle, librarian, explore, metis, momus DO have prompt_compose mappings — they are NOT exempt.
COMPOSE_EXEMPT = {
    "default",
    "multimodal-looker",
}


def scan_orphan_markers(inject_targets, project_root):
    """Find MASTER markers in target files that are NOT registered in rules.json.

    Returns list of (filepath, key) tuples for unregistered markers.
    """
    orphans = []
    for filepath, block_config in inject_targets.items():
        abs_path = os.path.join(project_root, filepath)
        if not os.path.exists(abs_path):
            continue

        with open(abs_path, "r", encoding="utf-8") as f:
            content = f.read()

        registered_keys = set(block_config.keys())
        found_keys = set(MARKER_SCAN_PATTERN.findall(content))
        unregistered = found_keys - registered_keys
        for key in sorted(unregistered):
            orphans.append((filepath, key))

    return orphans


def validate_compose_coverage(prompt_compose, profiles_path, project_root):
    """Verify every agent/category in shared section is composed or exempt.

    Returns list of (kind, name) tuples for unmapped entries.
    """
    abs_path = os.path.join(project_root, profiles_path)
    if not os.path.exists(abs_path):
        return []

    with open(abs_path, "r", encoding="utf-8") as f:
        data = json.load(f)

    shared = data.get("shared", {})
    gaps = []

    composed_agents = set(prompt_compose.get("agents", {}).keys())
    for agent_name in shared.get("agents", {}):
        if agent_name.startswith("_"):
            continue
        if agent_name not in composed_agents and agent_name not in COMPOSE_EXEMPT:
            gaps.append(("agent", agent_name))

    composed_cats = set(prompt_compose.get("categories", {}).keys())
    for cat_name in shared.get("categories", {}):
        if cat_name.startswith("_"):
            continue
        if cat_name not in composed_cats and cat_name not in COMPOSE_EXEMPT:
            gaps.append(("category", cat_name))

    return gaps
