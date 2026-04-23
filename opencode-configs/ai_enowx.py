"""
ai_enowx.py - enowX Labs proxy management

Handles enowxai CLI operations: setup (logout/login/start),
apikey fetching, and injecting apikey into config files.
"""

import json
import subprocess
from pathlib import Path

from ai_constants import ENOWX_CONFIG_LOCATIONS, ENOWX_LICENSE_KEY


def _run_cmd(args: list[str], capture: bool = False) -> str | None:
    """Run an enowxai CLI command. Returns stdout if capture=True, else None."""
    try:
        result = subprocess.run(
            ["enowxai"] + args,
            capture_output=capture,
            text=True,
            timeout=30,
        )
        if capture:
            return result.stdout.strip()
        return None
    except FileNotFoundError:
        print("   [ERROR] enowxai not found in PATH")
        return None
    except subprocess.TimeoutExpired:
        print(f"   [ERROR] enowxai {' '.join(args)} timed out")
        return None


def setup() -> bool:
    """Re-activate enowxai proxy: logout -> login -> start.

    Run manually when switching PCs or when the proxy needs a fresh start.
    Safe to skip if enowxai is already running.
    """
    print("\n   --- enowX Setup (logout -> login -> start) ---")

    print("   [1/3] enowxai logout...")
    _run_cmd(["logout"])

    print("   [2/3] enowxai login...")
    login_out = _run_cmd(["login", ENOWX_LICENSE_KEY], capture=True)
    if login_out:
        print(f"   {login_out}")

    print("   [3/3] enowxai start...")
    start_out = _run_cmd(["start"], capture=True)
    if start_out:
        print(f"   {start_out}")

    print("   --- enowX Setup Complete ---\n")
    return True


def fetch_and_inject_apikey() -> bool:
    """Fetch dynamic apikey from enowxai and inject into all config locations.

    The apikey differs per PC. Updates opencode-enowx.json at all known
    project locations so the correct key is used regardless of which project
    opencode is started from.

    Returns True on success.
    """
    print("   [enowX] Fetching apikey...")
    apikey = _run_cmd(["apikey"], capture=True)
    if not apikey:
        print("   [ERROR] Failed to retrieve enowxai apikey")
        print("   [HINT] Run [E] enowX Setup first if enowxai is not active")
        return False

    if not apikey.startswith("enx-"):
        print(f"   [ERROR] Unexpected apikey format: {apikey[:20]}...")
        return False

    print(f"   [OK] Got apikey: {apikey[:12]}...{apikey[-6:]}")

    updated = 0
    for config_path in ENOWX_CONFIG_LOCATIONS:
        if not config_path.exists():
            print(f"   [SKIP] {config_path} (not found)")
            continue

        try:
            data = json.loads(config_path.read_text(encoding="utf-8"))
            data["provider"]["enowxlabs"]["options"]["apiKey"] = apikey
            config_path.write_text(
                json.dumps(data, indent=2, ensure_ascii=False) + "\n",
                encoding="utf-8",
            )
            print(f"   [OK] Updated apiKey in {config_path.name} ({config_path.parent})")
            updated += 1
        except Exception as e:
            print(f"   [ERROR] Failed to update {config_path}: {e}")

    if updated == 0:
        print("   [ERROR] No config files were updated")
        return False

    print(f"   [OK] apiKey injected into {updated} file(s)")
    return True
