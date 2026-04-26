"""
ai_constants.py - Shared paths, mappings, and configuration constants

Single source of truth for all path definitions, profile mappings,
and enowX configuration used across AI.py modules.
"""

import os
from pathlib import Path

# ── Paths ────────────────────────────────────────────────────────────────────

# SCRIPT_DIR is set dynamically by AI.py (the root caller) via init()
SCRIPT_DIR: Path = Path(".")
CONFIG_DIR: Path = Path(".")
PROFILES_FILE: Path = Path(".")
TARGET_DIR: Path = Path.home() / ".config" / "opencode"

# AppData paths: only set if env vars are valid (non-empty)
_appdata = os.environ.get("APPDATA", "")
_localappdata = os.environ.get("LOCALAPPDATA", "")
APPDATA_DIR: Path | None = Path(_appdata) / "opencode" if _appdata else None
LOCALAPPDATA_DIR: Path | None = Path(_localappdata) / "opencode" if _localappdata else None


def init(script_dir: Path):
    """Initialize paths relative to the calling script's directory.

    Must be called once from AI.py before any other module uses these paths.
    """
    global SCRIPT_DIR, CONFIG_DIR, PROFILES_FILE
    SCRIPT_DIR = script_dir
    CONFIG_DIR = script_dir / "opencode-configs"
    PROFILES_FILE = CONFIG_DIR / "opencode-profiles.json"


# ── Profile mappings ─────────────────────────────────────────────────────────

MENU_OPTIONS = {
    "1": "mix-copilot",
    "2": "mix-antigravity",
    "3": "enowx",
    "4": "enowx-mix",
}

CLI_ARGS = {
    "mix-copilot": "1",
    "mix-antigravity": "2",
    "enowx": "3",
    "enowx-mix": "4",
    "enowx-setup": "e",
    "sync": "s",
    "config-sync": "x",
    "current": "c",
}

# Profiles that use direct delivery (no antigravity plugin, no account sync)
DIRECT_PROFILES = {"enowx", "enowx-mix"}

# Backward-compat aliases (old proxy/plugin CLI args -> new profile names)
DEPRECATED_CLI_ARGS = {
    "mix-copilot-proxy": "1",
    "mix-copilot-plugin": "1",
    "mix-antigravity-proxy": "2",
    "mix-antigravity-plugin": "2",
}

# ── enowX configuration ─────────────────────────────────────────────────────

# License key (same across all PCs; apikey differs per PC)
ENOWX_LICENSE_KEY = "ENOWX-BOVG9-DQTCC-5CW5Z-9L20N"

# All opencode-enowx.json locations to update with dynamic apikey
ENOWX_CONFIG_LOCATIONS = [
    Path("D:/Project/extensions/opencode-configs/opencode-enowx.json"),
    Path("D:/Project/omni/opencode-configs/opencode-enowx.json"),
    Path("D:/Project/Auto/opencode-configs/opencode-enowx.json"),
]

# ── Hub Sync ─────────────────────────────────────────────────────────────────

# Hub path: the single source of truth for AI.py + opencode-configs
AI_HUB_PATH = Path(os.environ.get("AI_HUB_PATH", r"D:\Project\ai"))

# Projects managed by the hub
HUB_SYNC_TARGETS = {
    "auto": Path("D:/Project/auto"),
    "extensions": Path("D:/Project/extensions"),
    "omni": Path("D:/Project/omni"),
}

# Files to sync from hub to targets (relative to project root)
HUB_SYNC_FILES = [
    "AI.py",
    "opencode-configs/ai_constants.py",
    "opencode-configs/ai_profiles.py",
    "opencode-configs/ai_apply.py",
    "opencode-configs/ai_enowx.py",
    "opencode-configs/ai_sync.py",
    "opencode-configs/config_sync.py",
    "opencode-configs/opencode-profiles.json",
    "opencode-configs/opencode-enowx.json",
    "opencode-configs/opencode-plugin.json",
    "opencode-configs/transform_config.py",
    "opencode-configs/test-accounts.js",
    "opencode-configs/test-accounts-helpers.js",
]

# Files to NEVER sync (project-specific or sensitive)
HUB_SYNC_EXCLUDE = [
    "opencode-configs/antigravity-accounts.json",
    "opencode-configs/antigravity-accounts copy.json",
    "opencode-configs/antigravity.json",
    "opencode-configs/refresh_token.json",
    "opencode-configs/AGENTS.md",
]
