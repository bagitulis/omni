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

# Agents/categories that intentionally have no prompt_append (advisory, skill-only, etc.)
COMPOSE_EXEMPT = {
    "oracle", "librarian", "explore", "multimodal-looker", "metis", "momus",
    "default", "artistry", "writing", "review", "testing", "security",
    "unspecified-low", "unspecified-high",
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


def validate_doc_freshness(docs_dir="docs", project_root="."):
    """Check if documentation files are stale based on stale_if_changed paths.

    Compares doc's last_updated date against file modification times.
    Returns list of stale docs with details.
    """
    from pathlib import Path
    from datetime import datetime

    stale_docs = []
    docs_path = Path(project_root) / docs_dir

    if not docs_path.exists():
        return stale_docs

    for md_file in docs_path.rglob("*.md"):
        content = md_file.read_text(encoding="utf-8")
        if not content.startswith("---"):
            continue

        # Extract frontmatter
        end_idx = content.find("---", 3)
        if end_idx == -1:
            continue

        fm_str = content[3:end_idx].strip()

        # Parse last_updated
        match = re.search(r"last_updated:\s*(.+)", fm_str)
        if not match:
            continue
        try:
            doc_date = datetime.strptime(match.group(1).strip(), "%Y-%m-%d")
        except ValueError:
            continue

        # Parse stale_if_changed
        stale_match = re.search(r"stale_if_changed:\s*\n((?:\s+-\s+.+\n?)+)", fm_str)
        if not stale_match:
            continue
        patterns = re.findall(r"-\s+(.+)", stale_match.group(1))
        patterns = [p.strip() for p in patterns]

        # Check each path/glob
        changed_files = []
        root = Path(project_root)
        for pattern in patterns:
            if "*" in pattern:
                matched = list(root.glob(pattern))
            else:
                p = root / pattern
                matched = [p] if p.exists() else []

            for f in matched:
                if not f.exists() or f.is_dir():
                    continue
                file_mtime = datetime.fromtimestamp(f.stat().st_mtime)
                if file_mtime.date() > doc_date.date():
                    changed_files.append(str(f))

        if changed_files:
            stale_docs.append({
                "doc": str(md_file.relative_to(project_root)),
                "last_updated": match.group(1).strip(),
                "changed_count": len(changed_files),
                "examples": changed_files[:3],
            })

    return stale_docs
