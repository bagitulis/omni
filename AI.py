#!/usr/bin/env python3
"""
AI.py - OpenCode Provider Switcher v4.2
Profile-based configuration system (CLI entry point)

2 Profiles: Mix Copilot, Mix Antigravity
Delivery: Proxy (clean names) or Plugin (antigravity-* names)
Source: opencode-configs/opencode-profiles.json (shared + per-profile model assignments)
Output: oh-my-opencode.json (valid schema, consumed by opencode)

Helper modules (in opencode-configs/):
  ai_profiles.py - Profile loading, merging, LSP detection, plugin transform
  ai_sync.py     - Account/config file synchronization across locations
"""

import json
import os
import socket
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
PROFILES = {"1": "mix-copilot", "2": "mix-antigravity"}
CLI_ARGS = {
    "mix-copilot": "1", "mix-copilot-proxy": "1p", "mix-copilot-plugin": "1l",
    "mix-antigravity": "2", "mix-antigravity-proxy": "2p", "mix-antigravity-plugin": "2l",
    "sync": "s", "current": "c",
}


def _is_cli_mode() -> bool:
    """Return True if running with command-line arguments (non-interactive)."""
    return len(sys.argv) > 1


def clear_screen():
    os.system("cls" if os.name == "nt" else "clear")


def check_proxy_running(port: int = 8045) -> bool:
    """Check if proxy is running on specified port."""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(1)
            return s.connect_ex(("127.0.0.1", port)) == 0
    except Exception:
        return False


def detect_current_provider() -> str:
    """Detect current provider from oh-my-opencode.json."""
    config_file = TARGET_DIR / "oh-my-opencode.json"
    if not config_file.exists():
        return "[None]"

    try:
        content = config_file.read_text(encoding="utf-8")
        has_copilot = "github-copilot" in content
        has_antigravity = "google/antigravity-" in content
        has_google = "google/" in content

        if has_copilot:
            if has_antigravity:
                return "Mix Copilot (Plugin)"
            if has_google:
                return "Mix Copilot (Proxy)"
            return "Mix Copilot"
        if has_antigravity:
            return "Mix Antigravity (Plugin)"
        if has_google:
            return "Mix Antigravity (Proxy)"

        return "[Unknown]"
    except Exception:
        return "[Error]"


def apply_profile(profile_name: str, delivery: str) -> bool:
    """Apply a profile with specified delivery method.

    Args:
        profile_name: 'mix-copilot' or 'mix-antigravity'
        delivery: 'proxy' or 'plugin'

    Returns True on success.
    """
    is_proxy = delivery == "proxy"
    is_plugin = delivery == "plugin"

    # Check proxy if needed (in CLI mode, just fail instead of prompting)
    if is_proxy and not check_proxy_running(8045):
        print("\n   [WARNING] Antigravity Proxy NOT running on port 8045!")
        print("   Please start Antigravity Tools and enable proxy first.")
        if _is_cli_mode():
            return False
        response = input("\n   Continue anyway? [y/N]: ").strip().lower()
        if response != "y":
            return False

    # Load and merge profile
    profiles_data = load_profiles(PROFILES_FILE)
    shared = profiles_data.get("shared", {})
    profile = profiles_data.get("profiles", {}).get(profile_name)

    if not profile:
        print(f"\n   [ERROR] Profile not found: {profile_name}")
        return False

    config = merge_profile(shared, profile)
    config = inject_lsp_config(config)

    # Serialize and optionally transform for plugin mode
    content = json.dumps(config, indent=2, ensure_ascii=False) + "\n"
    if is_plugin:
        content = transform_for_plugin(content)

    # Write oh-my-opencode.json
    TARGET_DIR.mkdir(parents=True, exist_ok=True)
    dst = TARGET_DIR / "oh-my-opencode.json"
    dst.write_text(content, encoding="utf-8")

    mode_label = "plugin" if is_plugin else "proxy"
    print(f"   [OK] Generated oh-my-opencode.json ({profile_name}, {mode_label})")

    # Copy opencode.json (proxy or plugin provider config)
    opencode_file = "opencode-plugin.json" if is_plugin else "opencode-proxy.json"
    src_opencode = CONFIG_DIR / opencode_file
    if src_opencode.exists():
        copy_file(src_opencode, TARGET_DIR / "opencode.json")
        print(f"   [OK] Copied {opencode_file} -> opencode.json")

    # Copy antigravity.json
    src_antigravity = CONFIG_DIR / "antigravity.json"
    if src_antigravity.exists():
        copy_file(src_antigravity, TARGET_DIR / "antigravity.json")
        print("   [OK] Copied antigravity.json")

    # Smart sync accounts for plugin modes
    if is_plugin:
        smart_sync_accounts(CONFIG_DIR, TARGET_DIR, APPDATA_DIR, LOCALAPPDATA_DIR)

    print(f"\n   [OK] Switched to {profile_name} ({mode_label})")
    if is_proxy:
        print("   [INFO] Using Antigravity Proxy at localhost:8045")
    if is_plugin:
        print("   [INFO] Using Antigravity Auth Plugin")

    return True


def ask_delivery() -> str:
    """Ask user for delivery method (proxy vs plugin)."""
    print()
    print("   Delivery method:")
    print("   [P] Proxy  - Via localhost:8045 (clean model names)")
    print("   [L] Plugin - Via Auth Plugin (antigravity-* names)")
    print()
    choice = input("   Select [P/L]: ").strip().lower()
    if choice in ("l", "plugin"):
        return "plugin"
    return "proxy"


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
    print("  |           AI.py - OpenCode Provider Switcher v4.2               |")
    print("  +==================================================================+")
    print("  |                                                                  |")
    print("  |   [1] Mix Copilot      - Copilot + Google + OpenAI              |")
    print("  |   [2] Mix Antigravity  - Google + OpenAI (no Copilot)           |")
    print("  |                                                                  |")
    print("  |   After selecting a profile, choose delivery:                    |")
    print("  |     [P] Proxy  - Via localhost:8045 (clean names)               |")
    print("  |     [L] Plugin - Via Auth Plugin (antigravity-* names)          |")
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
        choice = CLI_ARGS.get(arg, arg)
    else:
        show_menu()
        choice = input("   Select [1-2, S, C, Q]: ").strip().lower()

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

        elif choice in PROFILES:
            profile_name = PROFILES[choice]
            delivery = ask_delivery()
            if apply_profile(profile_name, delivery):
                start_opencode()
            break

        elif len(choice) == 2 and choice[0] in PROFILES and choice[1] in ("p", "l"):
            profile_name = PROFILES[choice[0]]
            delivery = "plugin" if choice[1] == "l" else "proxy"
            if apply_profile(profile_name, delivery):
                start_opencode()
            break

        else:
            print("\n   Invalid choice!")

        # Show menu again for interactive mode only
        if not cli_mode:
            show_menu()
            choice = input("   Select [1-2, S, C, Q]: ").strip().lower()
        else:
            break


if __name__ == "__main__":
    main()
