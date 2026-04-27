"""
ai_enowx.py - enowX Labs proxy management

Handles enowxai CLI operations: setup (logout/login/start),
apikey fetching, and injecting apikey into config files.
"""

import json
import subprocess
import time

from ai_constants import ENOWX_CONFIG_LOCATIONS, ENOWX_LICENSE_KEY


def _run_cmd(args: list[str], capture: bool = False) -> str | None:
    """Run an enowxai CLI command.

    Returns stdout if capture=True and command succeeds, else None.
    Checks returncode — prints stderr and returns None on failure.
    """
    try:
        result = subprocess.run(
            ["enowxai"] + args,
            capture_output=True,
            text=True,
            timeout=30,
        )
        if result.returncode != 0:
            stderr = result.stderr.strip()
            label = " ".join(args[:1])  # Show command name, not license key
            print(f"   [ERROR] enowxai {label} failed (exit code {result.returncode})")
            if stderr:
                print(f"   [STDERR] {stderr}")
            return None
        if capture:
            return result.stdout.strip()
        return None
    except FileNotFoundError:
        print("   [ERROR] enowxai not found in PATH — install from https://enowxlabs.com")
        return None
    except subprocess.TimeoutExpired:
        print(f"   [ERROR] enowxai {args[0] if args else ''} timed out (30s)")
        return None


def _mask_key(key: str) -> str:
    """Mask license key showing first 5 and last 5 chars only."""
    if len(key) <= 12:
        return key[:3] + "***" + key[-3:]
    return key[:5] + "***" + key[-5:]


def setup() -> bool:
    """Re-activate enowxai proxy: logout -> login -> start.

    Run manually when switching PCs or when the proxy needs a fresh start.
    Safe to skip if enowxai is already running.
    Returns True on success, False on failure.
    """
    print("\n   --- enowX Setup (logout -> login -> start) ---")

    # Validate license key format
    if not ENOWX_LICENSE_KEY.startswith("ENOWX-"):
        print(f"   [ERROR] Invalid license key format (expected ENOWX-...)")
        print(f"   [INFO] Check ENOWX_LICENSE_KEY in ai_constants.py")
        return False

    masked = _mask_key(ENOWX_LICENSE_KEY)
    print(f"   [KEY] Using license: {masked}")

    # Step 1: Logout
    print("   [1/3] enowxai logout...")
    _run_cmd(["logout"])
    time.sleep(1)

    # Step 2: Login with license key
    print("   [2/3] enowxai login...")
    login_out = _run_cmd(["login", ENOWX_LICENSE_KEY], capture=True)
    if login_out is None:
        print(f"   [ERROR] Login failed with key {masked}")
        print("   [INFO] Check network connection and license key validity")
        return False
    if "error" in login_out.lower():
        print(f"   [ERROR] Login returned error: {login_out}")
        return False
    print(f"   [OK] {login_out}")
    time.sleep(1)

    # Step 3: Start proxy
    print("   [3/3] enowxai start...")
    start_out = _run_cmd(["start"], capture=True)
    if start_out is None:
        print("   [ERROR] Failed to start enowxai proxy")
        return False
    if "error" in start_out.lower():
        print(f"   [ERROR] Start returned error: {start_out}")
        return False
    print(f"   [OK] {start_out}")

    # Wait for proxy server to fully initialize
    print("   [WAIT] Waiting for proxy server to initialize (5 seconds)...")
    time.sleep(5)

    # Verify proxy is running by fetching apikey
    print("   [VERIFY] Checking proxy is responding...")
    verify_out = _run_cmd(["apikey"], capture=True)
    if verify_out and verify_out.startswith("enx-"):
        print(f"   [OK] Proxy verified — apikey: {verify_out[:12]}...")
    else:
        print("   [WARN] Proxy started but apikey not yet available")
        print("   [INFO] Try selecting a profile (option 3/4) to retry apikey fetch")

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
        print("   [INFO] Auto-triggering enowX Setup...")
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
