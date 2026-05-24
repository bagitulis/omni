"""
ai_apply.py - Profile application and provider detection

Handles merging profiles, writing oh-my-opencode.json, and copying provider configs.
"""

import json
from pathlib import Path

from ai_constants import (
    CONFIG_DIR,
    DIRECT_PROFILES,
    PROFILES_FILE,
    TARGET_DIR,
    APPDATA_DIR,
    LOCALAPPDATA_DIR,
)
from ai_profiles import (
    inject_lsp_config,
    load_profiles,
    merge_profile,
    transform_for_plugin,
    validate_plugin_config,
)
from ai_sync import copy_file, smart_sync_accounts



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
        if "deepseek/" in default_model:
            return "Enowx Deepseek (Direct)"
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
    config = validate_plugin_config(config)

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
        for config_name in ("oh-my-openagent.json", "oh-my-opencode.json", "oh-my-china.json"):
            dst = target_dir / config_name
            dst.write_text(content, encoding="utf-8")

    delivery = "direct" if is_direct else "plugin"
    print(f"   [OK] Generated oh-my-openagent.json ({profile_name}, {delivery})")

    if is_direct:
        if profile_name == "9router":
            src_opencode = CONFIG_DIR / "opencode-9router.json"
            src_label = "opencode-9router.json"
        else:
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
        if profile_name == "enowx":
            print("   [INFO] Sisyphus uses opencode/deepseek-v4-flash-free (FREE). Other agents use DeepSeek API.")
        elif profile_name == "enowx-mix":
            print("   [INFO] Using enowX Labs proxy + DeepSeek API")
        elif profile_name == "9router":
            print("   [INFO] Using 9 Router (localhost:20128)")
        else:
            print("   [INFO] Using enowX Labs proxy (localhost:1430)")

    return True

