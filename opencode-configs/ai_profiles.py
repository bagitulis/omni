"""
ai_profiles.py - Profile loading, merging, and configuration utilities

Handles loading opencode-profiles.json, deep-merging shared config with
profile overrides, LSP server detection, and plugin model name transformation.
"""

import copy
import json
import shutil
import sys
from pathlib import Path


def transform_for_plugin(content: str) -> str:
    """Transform model names for plugin mode (google/* -> google/antigravity-*)."""
    content = content.replace("google/claude-", "google/antigravity-claude-")
    content = content.replace("google/gemini-", "google/antigravity-gemini-")
    return content


def load_profiles(profiles_file: Path) -> dict:
    """Load opencode-profiles.json and return parsed dict.

    Exits with error message if file is missing or contains invalid JSON.
    """
    if not profiles_file.exists():
        print(f"   [ERROR] Profiles file not found: {profiles_file}")
        sys.exit(1)

    try:
        with open(profiles_file, "r", encoding="utf-8") as f:
            return json.load(f)
    except json.JSONDecodeError as e:
        print(f"   [ERROR] Invalid JSON in {profiles_file.name}: {e}")
        sys.exit(1)


def merge_profile(shared: dict, profile: dict) -> dict:
    """Deep-merge shared config with profile overrides.

    Shared provides: $schema, google_auth, browser_automation_engine,
                     agents (prompt_append, skills), categories (prompt_append, skills).
    Profile provides: default_model, variant,
                      agents (model, variant, temperature, reasoningEffort),
                      categories (model, variant, temperature, reasoningEffort).

    Returns a valid oh-my-opencode.json dict (no internal keys like _comment).
    """
    result = {}

    # Copy top-level shared keys
    for key in ("$schema", "google_auth", "browser_automation_engine"):
        if key in shared:
            result[key] = copy.deepcopy(shared[key])

    # Copy profile top-level keys
    for key in ("default_model", "variant"):
        if key in profile:
            result[key] = profile[key]

    # Merge agents and categories: shared props + profile props
    result["agents"] = _merge_section(
        shared.get("agents", {}), profile.get("agents", {})
    )
    result["categories"] = _merge_section(
        shared.get("categories", {}), profile.get("categories", {})
    )

    return result


def _merge_section(shared_section: dict, profile_section: dict) -> dict:
    """Merge a shared section (agents or categories) with profile overrides.

    For each key present in shared OR profile:
      - Start with deep copy of shared entry
      - Update with profile entry (model, variant, temperature, reasoningEffort)
    """
    all_keys = set(list(shared_section.keys()) + list(profile_section.keys()))
    merged = {}

    for key in all_keys:
        entry = copy.deepcopy(shared_section.get(key, {}))
        entry.update(profile_section.get(key, {}))
        merged[key] = entry

    return merged


def detect_lsp_servers() -> dict:
    """Detect LSP servers available on this machine.

    Uses shutil.which() to verify availability, but writes portable binary
    names (not full paths) so configs work across machines.
    Returns a dict suitable for the 'lsp' key in oh-my-opencode.json.
    """
    servers = {}

    # (server_id, binary, extra_args, extensions, priority)
    lsp_candidates = [
        ("gopls", "gopls", [], [".go"], 10),
        ("typescript", "typescript-language-server", ["--stdio"], [".ts", ".tsx"], 10),
        ("biome", "biome", ["lsp-proxy", "--stdio"], [".ts", ".tsx", ".js", ".jsx", ".json", ".css"], 5),
    ]

    for server_id, binary, extra_args, extensions, priority in lsp_candidates:
        if shutil.which(binary):
            cmd = [binary] + extra_args if extra_args else [binary]
            servers[server_id] = {
                "command": cmd,
                "extensions": extensions,
                "priority": priority,
            }

    return servers


def inject_lsp_config(config: dict) -> dict:
    """Inject dynamically detected LSP server paths into config dict.

    Merges detected servers into the 'lsp' key. User-defined entries take precedence.
    """
    lsp_servers = detect_lsp_servers()
    if not lsp_servers:
        return config

    existing_lsp = config.get("lsp", {})
    for server_id, server_config in lsp_servers.items():
        if server_id not in existing_lsp:
            existing_lsp[server_id] = server_config
    config["lsp"] = existing_lsp
    return config
