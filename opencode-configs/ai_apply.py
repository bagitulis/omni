"""
ai_apply.py - Profile application and provider detection

Handles merging profiles, writing oh-my-opencode.json, copying provider
configs, detecting current provider, and starting opencode.
"""

import json
import os
import shutil
import subprocess
from pathlib import Path

from ai_constants import (
    CONFIG_DIR,
    DIRECT_PROFILES,
    PROFILES_FILE,
    SCRIPT_DIR,
    TARGET_DIR,
    APPDATA_DIR,
    LOCALAPPDATA_DIR,
)
from ai_profiles import (
    inject_lsp_config,
    load_profiles,
    merge_profile,
    transform_for_plugin,
)
from ai_sync import copy_file, smart_sync_accounts

# Minimum OpenCode version required for oh-my-openagent plugin support
_MIN_OPENCODE_VERSION = "1.0.133"


def _inject_small_model(merged_config: dict, opencode_json_path: Path):
    """Inject small_model from profile into opencode.json.

    Reads the copied opencode.json, sets the small_model key from the
    merged profile config, and writes it back. This ensures opencode uses
    the profile's preferred lightweight model for title generation instead
    of auto-selecting from available models.
    """
    small_model = merged_config.get("small_model")
    if not small_model:
        return

    try:
        data = json.loads(opencode_json_path.read_text(encoding="utf-8"))
        data["small_model"] = small_model
        opencode_json_path.write_text(
            json.dumps(data, indent=2, ensure_ascii=False) + "\n",
            encoding="utf-8",
        )
    except (json.JSONDecodeError, OSError) as e:
        print(f"   [WARN] Could not inject small_model: {e}")


def detect_current_provider() -> str:
    """Detect current provider from oh-my-openagent.json (or legacy oh-my-opencode.json)."""
    config_file = TARGET_DIR / "oh-my-openagent.json"
    if not config_file.exists():
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

    Direct profiles (enowx, enowx-mix):
      - No model name transform
      - Copy opencode-enowx.json as provider config
      - No antigravity.json, no account sync

    Returns True on success.
    """
    is_direct = profile_name in DIRECT_PROFILES

    profiles_data = load_profiles(PROFILES_FILE)
    shared = profiles_data.get("shared", {})
    profile = profiles_data.get("profiles", {}).get(profile_name)

    if not profile:
        print(f"\n   [ERROR] Profile not found: {profile_name}")
        return False

    config = merge_profile(shared, profile)
    config = inject_lsp_config(config)

    content = json.dumps(config, indent=2, ensure_ascii=False) + "\n"
    if not is_direct:
        content = transform_for_plugin(content)

    TARGET_DIR.mkdir(parents=True, exist_ok=True)

    # Write to all locations the plugin may search:
    #   - ~/.config/opencode/ (Linux/macOS default, also used by opencode core)
    #   - %APPDATA%/opencode/ (Windows, where oh-my-openagent reads on Windows)
    # Write both filenames: oh-my-openagent.json (current) + oh-my-opencode.json (legacy)
    write_dirs = [TARGET_DIR]
    if APPDATA_DIR and APPDATA_DIR != TARGET_DIR:
        APPDATA_DIR.mkdir(parents=True, exist_ok=True)
        write_dirs.append(APPDATA_DIR)

    for target_dir in write_dirs:
        for config_name in ("oh-my-openagent.json", "oh-my-opencode.json"):
            dst = target_dir / config_name
            dst.write_text(content, encoding="utf-8")

    delivery = "direct" if is_direct else "plugin"
    print(f"   [OK] Generated oh-my-openagent.json ({profile_name}, {delivery})")

    if is_direct:
        src_opencode = CONFIG_DIR / "opencode-enowx.json"
        src_label = "opencode-enowx.json"
    else:
        src_opencode = CONFIG_DIR / "opencode-plugin.json"
        src_label = "opencode-plugin.json"

    if src_opencode.exists():
        copy_file(src_opencode, TARGET_DIR / "opencode.json")
        _inject_small_model(profile, TARGET_DIR / "opencode.json")
        print(f"   [OK] Copied {src_label} -> opencode.json")
    else:
        print(f"   [WARN] {src_label} not found, skipping provider config")

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


def _resolve_opencode_binary(launch_dir: Path) -> str | None:
    """Find the best OpenCode binary using smart resolution order.

    Priority:
      1. Local opencode.exe in launch_dir (self-updating, user's preferred)
      2. Global 'opencode' in PATH (npm/bun global install)
      3. LOCALAPPDATA/opencode/opencode-cli.exe (Windows Desktop installer)

    Returns the path string or None if not found.
    """
    # 1. Local opencode.exe in project (self-updates, user's primary binary)
    local_exe = launch_dir / "opencode.exe"
    if local_exe.exists():
        return str(local_exe)

    # 2. Global PATH (npm/bun install)
    global_cmd = shutil.which("opencode")
    if global_cmd:
        return str(global_cmd)

    # 3. Windows Desktop installer CLI
    localappdata = os.environ.get("LOCALAPPDATA", "")
    if localappdata:
        desktop_cli = Path(localappdata) / "opencode" / "opencode-cli.exe"
        if desktop_cli.exists():
            return str(desktop_cli)

    return None


def _parse_version(version_str: str) -> tuple[int, ...]:
    """Parse version string like '1.14.31' into comparable tuple."""
    try:
        parts = version_str.strip().lstrip("v").split(".")
        return tuple(int(p) for p in parts[:3])
    except (ValueError, IndexError):
        return (0, 0, 0)


def _check_opencode_version(binary_path: str) -> str | None:
    """Get OpenCode version from binary. Returns version string or None."""
    try:
        result = subprocess.run(
            [binary_path, "--version"],
            capture_output=True, text=True, timeout=10,
        )
        if result.returncode == 0:
            return result.stdout.strip()
    except (subprocess.TimeoutExpired, OSError):
        pass
    return None


def _auto_upgrade_if_outdated(binary_path: str, version: str | None) -> str | None:
    """Auto-upgrade OpenCode if version is below minimum for plugin support.

    Returns the new version string after upgrade, or None if upgrade
    was not needed or failed.
    """
    if not version:
        return None
    current = _parse_version(version)
    minimum = _parse_version(_MIN_OPENCODE_VERSION)
    if current >= minimum:
        return None

    print(f"   [UPGRADE] OpenCode {version} is outdated (min: {_MIN_OPENCODE_VERSION})")
    print(f"   [UPGRADE] Auto-upgrading to latest...")

    try:
        result = subprocess.run(
            [binary_path, "upgrade"],
            capture_output=True, text=True, timeout=120,
        )
        if result.returncode == 0:
            new_version = _check_opencode_version(binary_path)
            print(f"   [OK] Upgraded to {new_version}")
            return new_version
        else:
            print(f"   [ERROR] Upgrade failed (exit {result.returncode})")
            if result.stderr:
                print(f"   [STDERR] {result.stderr[:200]}")
            print(f"   [INFO] Manual fix: opencode upgrade")
    except subprocess.TimeoutExpired:
        print("   [ERROR] Upgrade timed out (120s)")
        print(f"   [INFO] Manual fix: opencode upgrade")
    except OSError as e:
        print(f"   [ERROR] Upgrade failed: {e}")

    return None


def _ensure_opencode_config():
    """Ensure opencode.json exists at ~/.config/opencode/ with plugin list.

    On a fresh PC, opencode.json may not exist yet. Without it, OpenCode
    won't know which plugins to load (no oh-my-openagent = no custom agents).

    If missing, copies from the hub's opencode-enowx.json as a sensible
    default that includes the full plugin list.
    """
    target_config = TARGET_DIR / "opencode.json"
    if target_config.exists():
        return

    # Try enowx config first (has full plugin list), fallback to plugin config
    for src_name in ("opencode-enowx.json", "opencode-plugin.json"):
        src = CONFIG_DIR / src_name
        if src.exists():
            TARGET_DIR.mkdir(parents=True, exist_ok=True)
            copy_file(src, target_config)
            print(f"   [BOOTSTRAP] Created opencode.json from {src_name}")
            return

    print("   [WARN] No opencode.json found — run AI.py and select a profile")


def start_opencode(target_dir: Path | None = None):
    """Start OpenCode with smart binary resolution.

    Resolution order:
      1. Local opencode.exe in launch_dir (self-updating, user's preferred)
      2. Global 'opencode' in PATH (npm/bun global install)
      3. LOCALAPPDATA/opencode/opencode-cli.exe (Windows Desktop installer)

    Plugins (oh-my-openagent, etc.) are auto-installed by OpenCode itself
    at runtime from the "plugin" array in opencode.json.

    Args:
        target_dir: If provided, launch opencode from that directory.
                    If None, launch from SCRIPT_DIR.
    """
    launch_dir = target_dir if target_dir else SCRIPT_DIR

    # Ensure opencode.json exists (plugin list needed for agents)
    _ensure_opencode_config()

    # Resolve binary
    binary = _resolve_opencode_binary(launch_dir)
    if not binary:
        print("\n   [ERROR] OpenCode not found anywhere!")
        print("   [INFO] Install options:")
        print("          npm install -g opencode")
        print("          bun install -g opencode")
        print("          curl -fsSL https://opencode.ai/install | bash")
        return

    # Version check — auto-upgrade if too old
    version = _check_opencode_version(binary)
    new_version = _auto_upgrade_if_outdated(binary, version)
    if new_version:
        version = new_version

    if target_dir:
        print(f"\n   Starting OpenCode in {launch_dir.name}...")
    else:
        print("\n   Starting OpenCode...")

    if version:
        print(f"   [INFO] Using: {binary} (v{version})")
    print()

    try:
        subprocess.run([binary], cwd=str(launch_dir))
    except Exception as e:
        print(f"   [ERROR] Failed to start OpenCode: {e}")
