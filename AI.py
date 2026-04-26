#!/usr/bin/env python3
"""
AI.py - OpenCode Provider Switcher v5.4
Thin CLI entry point — menu display and dispatch only.

4 Profiles: Mix Copilot, Mix Antigravity, enowX, enowX Mix
Delivery:
  - Mix Copilot / Mix Antigravity: Plugin (antigravity-* names via Auth Plugin)
  - enowX / enowX Mix: Direct (enowxlabs/ provider via local proxy, no plugin)

Hub sync: On startup, auto-pulls latest configs from D:\\Project\\ai (hub).
  Disable with --no-sync flag or AI_HUB_PATH="" env var.

Helper modules (in opencode-configs/):
  ai_constants.py  - Paths, mappings, enowX config
  ai_profiles.py   - Profile loading, merging, LSP detection, plugin transform
  ai_apply.py      - Profile application, provider detection, start opencode
  ai_enowx.py      - enowX CLI operations (setup, apikey)
  ai_sync.py       - Account/config file sync across filesystem locations
  config_sync.py   - Cross-project config sync (hub→spoke + interactive)
"""

import os
import sys
from pathlib import Path

# Add opencode-configs to import path for helper modules
sys.path.insert(0, str(Path(__file__).parent / "opencode-configs"))

from ai_constants import (
    CLI_ARGS,
    DEPRECATED_CLI_ARGS,
    DIRECT_PROFILES,
    MENU_OPTIONS,
    init as init_paths,
)

# Initialize paths before importing modules that depend on them
SCRIPT_DIR = Path(__file__).parent
init_paths(SCRIPT_DIR)

from ai_apply import apply_profile, detect_current_provider, start_opencode
from ai_enowx import fetch_and_inject_apikey, setup as enowx_setup
from ai_sync import smart_sync_accounts
from ai_constants import APPDATA_DIR, CONFIG_DIR, LOCALAPPDATA_DIR, TARGET_DIR
from config_sync import run_interactive as config_sync_interactive


def _auto_sync_from_hub():
    """Pull latest configs from hub if this project is not the hub itself.

    Runs silently on success. Warns (non-blocking) on failure.
    Skipped if --no-sync flag is present or AI_HUB_PATH is empty.
    """
    if "--no-sync" in sys.argv:
        return

    from ai_constants import AI_HUB_PATH, HUB_SYNC_TARGETS

    # If AI_HUB_PATH env var is explicitly empty, skip
    if not os.environ.get("AI_HUB_PATH", "default"):
        return

    hub = AI_HUB_PATH
    if not hub.exists():
        return

    # Don't sync if we ARE the hub
    if SCRIPT_DIR.resolve() == hub.resolve():
        return

    # Don't sync if we're not a known target
    is_target = any(
        SCRIPT_DIR.resolve() == t.resolve()
        for t in HUB_SYNC_TARGETS.values()
    )
    if not is_target:
        return

    try:
        from config_sync import sync_from_hub
        copied, new, _skipped, _details, errors = sync_from_hub(hub, SCRIPT_DIR)
        if errors:
            print(f"   [HUB-SYNC] {len(errors)} error(s) syncing from hub")
        elif copied or new:
            print(f"   [HUB-SYNC] Updated {copied + new} file(s) from hub")
    except Exception as e:
        print(f"   [HUB-SYNC] Warning: {e}")


def _is_cli_mode() -> bool:
    return len(sys.argv) > 1


def _clear_screen():
    os.system("cls" if os.name == "nt" else "clear")


def _show_menu():
    _clear_screen()
    print()
    print("  +==================================================================+")
    print("  |           AI.py - OpenCode Provider Switcher v5.3               |")
    print("  +==================================================================+")
    print("  |                                                                  |")
    print("  |   [1] Mix Copilot      (Copilot + Google + OpenAI)  [Plugin]    |")
    print("  |   [2] Mix Antigravity  (Google + OpenAI)            [Plugin]    |")
    print("  |   [3] enowX            (enowX Labs proxy)           [Direct]    |")
    print("  |   [4] enowX Mix        (enowX Labs multi-model)     [Direct]    |")
    print("  |                                                                  |")
    print("  |   [E] enowX Setup  - logout/login/start (run once per PC)       |")
    print("  |   [S] Sync         - Sync accounts across locations             |")
    print("  |   [X] Config Sync  - Sync models/scripts between projects       |")
    print("  |   [C] Current      - Show current provider                      |")
    print("  |   [Q] Quit                                                      |")
    print("  |                                                                  |")
    print("  +==================================================================+")
    print()
    print(f"   Current Provider: {detect_current_provider()}")
    print()


def _prompt():
    return input("   Select [1, 2, 3, 4, E, S, X, C, Q]: ").strip().lower()


def _handle_choice(choice: str, cli_mode: bool) -> bool:
    """Handle a single menu choice. Returns True to exit, False to loop."""
    if choice == "q":
        return True

    if choice == "e":
        enowx_setup()
        if not cli_mode:
            input("\n   Press Enter to continue...")
        return cli_mode

    if choice == "s":
        print("\n   Syncing accounts across all locations...")
        smart_sync_accounts(CONFIG_DIR, TARGET_DIR, APPDATA_DIR, LOCALAPPDATA_DIR)
        if not cli_mode:
            input("\n   Press Enter to continue...")
        return cli_mode

    if choice == "x":
        config_sync_interactive(caller_dir=SCRIPT_DIR)
        if not cli_mode:
            input("\n   Press Enter to continue...")
        return cli_mode

    if choice == "c":
        print(f"\n   Current Provider: {detect_current_provider()}")
        if not cli_mode:
            input("\n   Press Enter to continue...")
        return cli_mode

    if choice in MENU_OPTIONS:
        profile_name = MENU_OPTIONS[choice]
        if profile_name in DIRECT_PROFILES:
            if not fetch_and_inject_apikey():
                print("\n   [AUTO-TRIGGER] Running enowX Setup...")
                enowx_setup()
                print("\n   [RETRY] Fetching apikey after setup...")
                if not fetch_and_inject_apikey():
                    print("\n   [ERROR] Setup completed but apikey fetch still failed")
                    if not cli_mode:
                        input("\n   Press Enter to continue...")
                    return cli_mode
        if apply_profile(profile_name):
            start_opencode()
        return True

    print("\n   Invalid choice!")
    return cli_mode


def main():
    _auto_sync_from_hub()

    # Strip --no-sync from argv so it doesn't interfere with CLI dispatch
    sys.argv = [a for a in sys.argv if a != "--no-sync"]
    cli_mode = _is_cli_mode()

    if cli_mode:
        arg = sys.argv[1].lower()
        if arg in DEPRECATED_CLI_ARGS:
            new_arg = arg.replace("-proxy", "").replace("-plugin", "")
            print(f"   [DEPRECATED] '{arg}' is deprecated. Use '{new_arg}' instead.")
            choice = DEPRECATED_CLI_ARGS[arg]
        else:
            choice = CLI_ARGS.get(arg, arg)
        _handle_choice(choice, cli_mode=True)
    else:
        while True:
            _show_menu()
            choice = _prompt()
            if _handle_choice(choice, cli_mode=False):
                break


if __name__ == "__main__":
    main()
