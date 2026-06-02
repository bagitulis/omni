"""
ai_constants.py - Shared paths, mappings, and configuration constants

Single source of truth for all path definitions, profile mappings,
and enowX configuration used across AI.py modules.

Cross-platform: auto-detects Windows vs Linux/macOS paths.
"""

import os
import platform
from pathlib import Path

# ── Platform Detection ────────────────────────────────────────────────────────

IS_WINDOWS = platform.system() == "Windows"
IS_LINUX = platform.system() == "Linux"
IS_MACOS = platform.system() == "Darwin"


def _get_real_home() -> Path:
    """Get the real home directory, bypassing $HOME env overrides.

    On Linux/macOS, uses pwd module to read from /etc/passwd — this always
    returns the actual system home (e.g. /root) regardless of $HOME overrides
    by tools like Hermes Agent profiles, tmux sessions, or sudo.

    On Windows, falls back to USERPROFILE or Path.home().
    """
    if IS_WINDOWS:
        # Windows: USERPROFILE is stable, not overridden by tools
        userprofile = os.environ.get("USERPROFILE", "")
        if userprofile:
            return Path(userprofile)
        return Path.home()
    else:
        # Linux/macOS: use pwd to get real home from /etc/passwd
        try:
            import pwd
            return Path(pwd.getpwuid(os.getuid()).pw_dir)
        except (ImportError, KeyError):
            # Fallback: expanduser for the actual username
            try:
                username = os.environ.get("USER") or os.environ.get("LOGNAME") or "root"
                return Path(os.path.expanduser(f"~{username}"))
            except Exception:
                return Path.home()


_REAL_HOME = _get_real_home()

# ── Paths ────────────────────────────────────────────────────────────────────

# SCRIPT_DIR is set dynamically by AI.py (the root caller) via init()
SCRIPT_DIR: Path = Path(".")
CONFIG_DIR: Path = Path(".")
PROFILES_FILE: Path = Path(".")
# TARGET_DIR uses real home (from /etc/passwd), not $HOME which may be
# overridden by Hermes Agent profiles or other tools
TARGET_DIR: Path = _REAL_HOME / ".config" / "opencode"

# AppData paths: Windows-only (APPDATA/LOCALAPPDATA env vars)
_appdata = os.environ.get("APPDATA", "")
_localappdata = os.environ.get("LOCALAPPDATA", "")
APPDATA_DIR: Path | None = Path(_appdata) / "opencode" if _appdata else None
LOCALAPPDATA_DIR: Path | None = Path(_localappdata) / "opencode" if _localappdata else None


def _detect_project_root() -> Path:
    """Detect the project root directory dynamically.

    Strategy: derive from SCRIPT_DIR by walking up until we find the parent
    that contains this 'ai' project folder. No hardcoded paths.
    """
    script_resolved = SCRIPT_DIR.resolve()

    # Walk up from SCRIPT_DIR to find the "Project" parent (or equivalent)
    # AI.py lives in <PROJECT_ROOT>/ai/, so parent of SCRIPT_DIR is PROJECT_ROOT
    if script_resolved.name == "ai":
        return script_resolved.parent

    # Generic: find a parent named "Project" or similar container
    for parent in script_resolved.parents:
        # Check if this parent contains known sibling projects
        if (parent / "ai").exists() and (parent / "ai" / "AI.py").exists():
            return parent

    # Last resort: assume SCRIPT_DIR's parent is the root
    return script_resolved.parent


PROJECT_ROOT: Path = Path(".")  # Set by init()


def init(script_dir: Path):
    """Initialize paths relative to the calling script's directory.

    Must be called once from AI.py before any other module uses these paths.
    """
    global SCRIPT_DIR, CONFIG_DIR, PROFILES_FILE, PROJECT_ROOT
    SCRIPT_DIR = script_dir
    CONFIG_DIR = script_dir / "opencode-configs"
    PROFILES_FILE = CONFIG_DIR / "opencode-profiles.json"
    PROJECT_ROOT = _detect_project_root()
    _init_dynamic_paths()


# ── Profile mappings ─────────────────────────────────────────────────────────

MENU_OPTIONS = {
    "1": "mix-copilot",
    "2": "mix-antigravity",
    "3": "enowx",
    "4": "enowx-mix",
    "5": "enowx-std",
}

CLI_ARGS = {
    "mix-copilot": "1",
    "mix-antigravity": "2",
    "enowx": "3",
    "enowx-mix": "4",
    "enowx-std": "5",
    "enowx-setup": "e",
    "sync": "s",
    "config-sync": "x",
    "open-auto": "open-auto",
    "open-omni": "open-omni",
    "open-extensions": "open-extensions",
    "current": "c",
}

# Profiles that use direct delivery (no antigravity plugin, no account sync)
DIRECT_PROFILES = {"enowx", "enowx-mix", "enowx-std"}

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


def _build_enowx_config_locations() -> list[Path]:
    """Build enowX config locations based on detected PROJECT_ROOT."""
    root = PROJECT_ROOT
    return [
        root / "ai" / "opencode-configs" / "opencode-enowx.json",
        root / "extensions" / "opencode-configs" / "opencode-enowx.json",
        root / "omni" / "opencode-configs" / "opencode-enowx.json",
        root / "auto" / "opencode-configs" / "opencode-enowx.json",
        root / "ads-analytics" / "opencode-configs" / "opencode-enowx.json",
    ]


def _build_hub_sync_targets() -> dict[str, Path]:
    """Build hub sync targets based on detected PROJECT_ROOT."""
    root = PROJECT_ROOT
    return {
        "auto": root / "auto",
        "extensions": root / "extensions",
        "omni": root / "omni",
        "ads-analytics": root / "ads-analytics",
    }


# These are initialized lazily after init() sets PROJECT_ROOT
ENOWX_CONFIG_LOCATIONS: list[Path] = []
AI_HUB_PATH: Path = Path(".")
HUB_SYNC_TARGETS: dict[str, Path] = {}


def _init_dynamic_paths():
    """Initialize paths that depend on PROJECT_ROOT. Called from init()."""
    global ENOWX_CONFIG_LOCATIONS, AI_HUB_PATH, HUB_SYNC_TARGETS
    ENOWX_CONFIG_LOCATIONS = _build_enowx_config_locations()
    AI_HUB_PATH = Path(os.environ.get("AI_HUB_PATH", str(PROJECT_ROOT / "ai")))
    HUB_SYNC_TARGETS = _build_hub_sync_targets()

# Files to sync from hub to targets (relative to project root)
# NOTE: AI.py is NOT synced — it lives only in the hub (ai/ project).
#       Spoke projects are launched via [O] Open in from the hub.
HUB_SYNC_FILES = [
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

# BMAD Method directories (per-project, never synced between projects)
# Each project has its own _bmad/, .claude/skills/, and _bmad-output/
# installed via: npx bmad-method install --modules bmm --tools claude-code --yes
BMAD_DIRS_EXCLUDE = [
    "_bmad/",
    "_bmad-output/",
    ".claude/skills/",
]
