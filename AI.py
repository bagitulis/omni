#!/usr/bin/env python3
"""
AI.py - OpenCode Provider Switcher v5.1
Profile-based configuration system (CLI entry point)

3 Profiles: Mix Copilot, Mix Antigravity, enowX
Delivery:
  - Mix Copilot / Mix Antigravity: Plugin (antigravity-* names via Auth Plugin)
  - enowX: Direct (enowxlabs/ provider via local proxy, no plugin)
Source: opencode-configs/opencode-profiles.json (shared + per-profile model assignments)
Output: oh-my-opencode.json (valid schema, consumed by opencode)

Helper modules (in opencode-configs/):
  ai_profiles.py - Profile loading, merging, LSP detection, plugin transform
  ai_sync.py     - Account/config file synchronization across locations
"""

import json
import os
import subprocess
import sys
from pathlib import Path

# Add opencode-configs to import path for helper modules
sys.path.insert(0, str(Path(__file__).parent / "opencode-configs"))

from ai_profiles import (
    inject_lsp_config,
    load_profiles,
    merge_profile,
    transform_for_plugin,
)
from ai_sync import copy_file, smart_sync_accounts

# Paths
SCRIPT_DIR = Path(__file__).parent
CONFIG_DIR = SCRIPT_DIR / "opencode-configs"
PROFILES_FILE = CONFIG_DIR / "opencode-profiles.json"
TARGET_DIR = Path.home() / ".config" / "opencode"

# AppData paths: only set if env vars are valid (non-empty)
_appdata = os.environ.get("APPDATA", "")
_localappdata = os.environ.get("LOCALAPPDATA", "")
APPDATA_DIR = Path(_appdata) / "opencode" if _appdata else None
LOCALAPPDATA_DIR = Path(_localappdata) / "opencode" if _localappdata else None

# Profile and CLI mappings
MENU_OPTIONS = {
    "1": "mix-copilot",
    "2": "mix-antigravity",
    "3": "enowx",
}
CLI_ARGS = {
    "mix-copilot": "1",
    "mix-antigravity": "2",
    "enowx": "3",
    "sync": "s",
    "current": "c",
}

# Profiles that use direct delivery (no antigravity plugin, no account sync)
_DIRECT_PROFILES = {"enowx"}

# Backward-compat aliases (old proxy/plugin CLI args → new profile names)
_DEPRECATED_CLI_ARGS = {
    "mix-copilot-proxy": "1",
    "mix-copilot-plugin": "1",
    "mix-antigravity-proxy": "2",
    "mix-antigravity-plugin": "2",
}


def _is_cli_mode() -> bool:
    """Return True if running with command-line arguments (non-interactive)."""
    return len(sys.argv) > 1


def clear_screen():
    os.system("cls" if os.name == "nt" else "clear")


def detect_current_provider() -> str:
    """Detect current provider from oh-my-opencode.json."""
    config_file = TARGET_DIR / "oh-my-opencode.json"
    if not config_file.exists():
        return "[None]"

    try:
        config = json.loads(config_file.read_text(encoding="utf-8"))
        default_model = config.get("default_model", "")

        if "github-copilot" in default_model:
            return "Mix Copilot (Plugin)"
        if "google/" in default_model:
            return "Mix Antigravity (Plugin)"
        if "enowxlabs/" in default_model:
            return "enowX (Direct)"

        return f"[Unknown: {default_model}]"
    except Exception:
        return "[Error]"


def apply_profile(profile_name: str) -> bool:
    """Apply a profile using the appropriate delivery method.

    Plugin profiles (mix-copilot, mix-antigravity):
      - Transform model names for antigravity plugin
      - Copy opencode-plugin.json as provider config
      - Copy antigravity.json and sync accounts

    Direct profiles (enowx):
      - No model name transform
      - Copy opencode-enowx.json as provider config
      - No antigravity.json, no account sync

    Args:
        profile_name: 'mix-copilot', 'mix-antigravity', or 'enowx'

    Returns True on success.
    """
    is_direct = profile_name in _DIRECT_PROFILES

    # Load and merge profile
    profiles_data = load_profiles(PROFILES_FILE)
    shared = profiles_data.get("shared", {})
    profile = profiles_data.get("profiles", {}).get(profile_name)

    if not profile:
        print(f"\n   [ERROR] Profile not found: {profile_name}")
        return False

    config = merge_profile(shared, profile)
    config = inject_lsp_config(config)

    # Serialize (transform only for plugin profiles)
    content = json.dumps(config, indent=2, ensure_ascii=False) + "\n"
    if not is_direct:
        content = transform_for_plugin(content)

    # Write oh-my-opencode.json
    TARGET_DIR.mkdir(parents=True, exist_ok=True)
    dst = TARGET_DIR / "oh-my-opencode.json"
    dst.write_text(content, encoding="utf-8")

    delivery = "direct" if is_direct else "plugin"
    print(f"   [OK] Generated oh-my-opencode.json ({profile_name}, {delivery})")

    # Copy provider config (opencode.json)
    if is_direct:
        src_opencode = CONFIG_DIR / "opencode-enowx.json"
        src_label = "opencode-enowx.json"
    else:
        src_opencode = CONFIG_DIR / "opencode-plugin.json"
        src_label = "opencode-plugin.json"

    if src_opencode.exists():
        copy_file(src_opencode, TARGET_DIR / "opencode.json")
        print(f"   [OK] Copied {src_label} -> opencode.json")
    else:
        print(f"   [WARN] {src_label} not found, skipping provider config")

    # Plugin-only: copy antigravity.json and sync accounts
    if not is_direct:
        src_antigravity = CONFIG_DIR / "antigravity.json"
        if src_antigravity.exists():
            copy_file(src_antigravity, TARGET_DIR / "antigravity.json")
            print("   [OK] Copied antigravity.json")

        smart_sync_accounts(CONFIG_DIR, TARGET_DIR, APPDATA_DIR, LOCALAPPDATA_DIR)
        print(f"\n   [OK] Switched to {profile_name} (plugin)")
        print("   [INFO] Using Antigravity Auth Plugin")
    else:
        print(f"\n   [OK] Switched to {profile_name} (direct)")
        print("   [INFO] Using enowX Labs proxy (localhost:1430)")

    return True


def start_opencode():
    """Start OpenCode."""
    print("\n   Starting OpenCode...")
    print()
    exe_path = SCRIPT_DIR / "opencode.exe"
    try:
        if exe_path.exists():
            subprocess.run([str(exe_path)])
        else:
            print("   [ERROR] opencode.exe not found")
    except Exception as e:
        print(f"   [ERROR] Failed to start OpenCode: {e}")


def show_menu():
    """Display the main menu."""
    clear_screen()
    print()
    print("  +==================================================================+")
    print("  |           AI.py - OpenCode Provider Switcher v5.1               |")
    print("  +==================================================================+")
    print("  |                                                                  |")
    print("  |   [1] Mix Copilot      (Copilot + Google + OpenAI)  [Plugin]    |")
    print("  |   [2] Mix Antigravity  (Google + OpenAI)            [Plugin]    |")
    print("  |   [3] enowX            (enowX Labs proxy)           [Direct]    |")
    print("  |                                                                  |")
    print("  |   [S] Sync    - Sync accounts across locations                   |")
    print("  |   [C] Current - Show current provider                            |")
    print("  |   [Q] Quit                                                       |")
    print("  |                                                                  |")
    print("  +==================================================================+")
    print()
    current = detect_current_provider()
    print(f"   Current Provider: {current}")
    print()


def main():
    """Main entry point with CLI and interactive support."""
    cli_mode = _is_cli_mode()

    if cli_mode:
        arg = sys.argv[1].lower()
        if arg in _DEPRECATED_CLI_ARGS:
            new_arg = arg.replace("-proxy", "").replace("-plugin", "")
            print(f"   [DEPRECATED] '{arg}' is deprecated. Use '{new_arg}' instead.")
            choice = _DEPRECATED_CLI_ARGS[arg]
        else:
            choice = CLI_ARGS.get(arg, arg)
    else:
        show_menu()
        choice = input("   Select [1, 2, 3, S, C, Q]: ").strip().lower()

    while True:
        if choice == "q":
            break

        elif choice == "s":
            print("\n   Syncing accounts across all locations...")
            smart_sync_accounts(CONFIG_DIR, TARGET_DIR, APPDATA_DIR, LOCALAPPDATA_DIR)
            if cli_mode:
                break
            print()
            input("   Press Enter to continue...")

        elif choice == "c":
            current = detect_current_provider()
            print(f"\n   Current Provider: {current}")
            if cli_mode:
                break
            print()
            input("   Press Enter to continue...")

        elif choice in MENU_OPTIONS:
            profile_name = MENU_OPTIONS[choice]
            if apply_profile(profile_name):
                start_opencode()
            break

        else:
            print("\n   Invalid choice!")

        # Show menu again for interactive mode only
        if not cli_mode:
            show_menu()
            choice = input("   Select [1, 2, 3, S, C, Q]: ").strip().lower()
        else:
            break


if __name__ == "__main__":
    main()
