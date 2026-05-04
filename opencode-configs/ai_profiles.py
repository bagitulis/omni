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
            data = json.load(f)
        if "shared" not in data or "profiles" not in data:
            print(f"   [ERROR] {profiles_file.name} missing 'shared' or 'profiles' key")
            sys.exit(1)
        return data
    except json.JSONDecodeError as e:
        print(f"   [ERROR] Invalid JSON in {profiles_file.name}: {e}")
        sys.exit(1)


def merge_profile(shared: dict, profile: dict) -> dict:
    """Deep-merge shared config with profile overrides.

    Shared provides all top-level config keys: $schema, google_auth,
    browser_automation_engine, model_fallback, runtime_fallback, hashline_edit,
    experimental, notification, auto_update, background_task, plus
    agents (prompt_append, skills, permission) and categories (prompt_append, skills).

    Profile provides: default_model, small_model, variant,
                      agents (model, variant, temperature, reasoningEffort, fallback_models),
                      categories (model, variant, temperature, reasoningEffort).

    Returns a valid oh-my-opencode.json dict (no internal keys like _comment).
    """
    # Keys that are internal to opencode-profiles.json, not part of output
    _internal_keys = {"_comment", "_version", "_shared_warning", "_description"}
    # Keys that require deep-merge (shared + profile sections combined)
    _merge_keys = {"agents", "categories"}
    # Profile keys that are ONLY for internal use (not valid in oh-my-openagent.json)
    # 'model' in profile root = legacy reference model, not a valid output key
    _profile_only_keys = {"model"}

    result = {}

    # Copy ALL shared top-level keys (except internal and merge-keys)
    for key, value in shared.items():
        if key in _internal_keys or key in _merge_keys:
            continue
        result[key] = copy.deepcopy(value)

    # Copy profile top-level keys (except internal, merge-keys, and profile-only)
    # Profile keys override shared keys (e.g. profile can override auto_update)
    for key, value in profile.items():
        if key in _internal_keys or key in _merge_keys or key in _profile_only_keys:
            continue
        result[key] = copy.deepcopy(value)

    # Deep-merge agents and categories: shared props + profile props
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


def validate_plugin_config(config: dict) -> dict:
    """Validate and auto-fix dynamic_context_pruning to match plugin Zod schema."""
    dcp = config.get("experimental", {}).get("dynamic_context_pruning")
    if not dcp:
        return config

    if isinstance(dcp.get("turn_protection"), (int, float)):
        dcp["turn_protection"] = {"enabled": True, "turns": dcp["turn_protection"]}
        print("   [AUTO-FIX] turn_protection: number -> {enabled, turns}")

    strategies = dcp.get("strategies", {})
    for key in ("deduplication", "supersede_writes", "purge_errors"):
        if isinstance(strategies.get(key), bool):
            strategies[key] = {"enabled": strategies[key]}
            print(f"   [AUTO-FIX] strategies.{key}: bool -> {{enabled}}")

    return config
