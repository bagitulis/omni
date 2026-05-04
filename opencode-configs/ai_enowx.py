"""
ai_enowx.py - enowX Labs proxy management

Handles enowxai CLI operations: setup (logout/login/start),
apikey fetching, and injecting apikey into config files.
"""

import json
import subprocess
import time
import urllib.request
import urllib.error

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


def _wait_for_dashboard(timeout: int = 30) -> bool:
    """Wait for dashboard to be ready by polling health endpoint.

    Returns True if dashboard responds, False if timeout reached.
    """
    import socket

    start_time = time.time()
    check_interval = 1  # Check every second initially

    while time.time() - start_time < timeout:
        try:
            # Try to connect to dashboard port
            sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            sock.settimeout(2)
            result = sock.connect_ex(('127.0.0.1', 1431))
            sock.close()

            if result == 0:
                # Port is open, now try HTTP request
                try:
                    req = urllib.request.Request(
                        'http://127.0.0.1:1431',
                        method='HEAD',
                        headers={'User-Agent': 'AI.py/6.0'}
                    )
                    with urllib.request.urlopen(req, timeout=3) as response:
                        # Dashboard is responding
                        return True
                except urllib.error.HTTPError as e:
                    # HTTP error but server is responding (e.g., 401 Unauthorized)
                    if e.code in (401, 403, 302, 307):
                        # These mean dashboard is up but needs auth
                        print(f"   [OK] Dashboard ready (auth required)")
                        return True
                    # Other errors might mean still initializing
                except Exception:
                    # Connection refused or other error, keep waiting
                    pass
        except Exception:
            pass

        time.sleep(check_interval)
        # Show progress every 5 seconds
        elapsed = int(time.time() - start_time)
        if elapsed % 5 == 0 and elapsed > 0:
            print(f"   [WAIT] Still waiting... ({elapsed}s/{timeout}s)")

    return False

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
    # Wait for dashboard to be ready with health check
    print("   [WAIT] Waiting for dashboard to initialize...")
    if not _wait_for_dashboard(timeout=30):
        print("   [WARN] Dashboard not responding — may need login")
        print("   [INFO] Open http://localhost:1431 in browser and login with password: 123y")
        print("   [INFO] Then retry selecting enowX profile")
        print("   --- enowX Setup Complete (with warnings) ---\n")
        return True  # Return True because proxy started, just needs auth

    # Verify proxy is running by fetching apikey
    print("   [VERIFY] Checking proxy is responding...")
    verify_out = _run_cmd(["apikey"], capture=True)
    if verify_out and verify_out.startswith("enx-"):
        print(f"   [OK] Proxy verified — apikey: {verify_out[:12]}...")
    else:
        print("   [WARN] Proxy started but apikey not yet available")
        print("   [INFO] If dashboard shows login page, use password: 123y")
        print("   [INFO] Then retry selecting enowX profile")

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

    # Retry with exponential backoff for dashboard initialization
    max_retries = 5
    base_delay = 1
    apikey = None

    for attempt in range(max_retries):
        apikey = _run_cmd(["apikey"], capture=True)
        if apikey and apikey.startswith("enx-"):
            break
        if attempt < max_retries - 1:
            delay = base_delay * (2 ** attempt)  # 1, 2, 4, 8, 16 seconds
            print(f"   [WAIT] Dashboard not ready, retrying in {delay}s... (attempt {attempt + 1}/{max_retries})")
            time.sleep(delay)

    if not apikey:
        print("   [ERROR] Failed to retrieve enowxai apikey after retries")
        print("   [INFO] Dashboard may require authentication")
        print("   [INFO] Open http://localhost:1431 in browser and login with password: 123y")
        print("   [INFO] Then retry selecting enowX profile")
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
