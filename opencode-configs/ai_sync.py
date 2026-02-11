"""
ai_sync.py - Account and config file synchronization utilities

Handles syncing antigravity-accounts.json and antigravity.json
across multiple filesystem locations (newest file wins).
"""

from __future__ import annotations

import shutil
from pathlib import Path


def get_file_mtime(filepath: Path) -> float:
    """Get file modification time, return 0 if not exists."""
    try:
        return filepath.stat().st_mtime if filepath.exists() else 0
    except Exception:
        return 0


def copy_file(src: Path, dst: Path):
    """Copy file, creating parent directories if needed."""
    dst.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dst)


def smart_sync_accounts(
    config_dir: Path,
    target_dir: Path,
    appdata_dir: Path | None,
    localappdata_dir: Path | None,
):
    """Sync antigravity-accounts.json across all locations - newest wins.

    Args:
        config_dir: Local opencode-configs directory
        target_dir: ~/.config/opencode directory
        appdata_dir: AppData/Roaming/opencode directory (may be invalid)
        localappdata_dir: AppData/Local/opencode directory (may be invalid)
    """
    locations = [
        (config_dir / "antigravity-accounts.json", "opencode-configs"),
        (target_dir / "antigravity-accounts.json", ".config/opencode"),
    ]

    # Only include AppData locations if paths are valid (not empty)
    if appdata_dir and str(appdata_dir) != "opencode":
        locations.append(
            (appdata_dir / "antigravity-accounts.json", "AppData/Roaming/opencode")
        )
    if localappdata_dir and str(localappdata_dir) != "opencode":
        locations.append(
            (localappdata_dir / "antigravity-accounts.json", "AppData/Local/opencode")
        )

    target_dir.mkdir(parents=True, exist_ok=True)
    if appdata_dir and str(appdata_dir) != "opencode":
        appdata_dir.mkdir(parents=True, exist_ok=True)

    mtimes = [(loc, name, get_file_mtime(loc)) for loc, name in locations]
    valid_files = [(loc, name, mtime) for loc, name, mtime in mtimes if mtime > 0]

    if not valid_files:
        print("   [SKIP] No antigravity-accounts.json found in any location")
        return

    newest = max(valid_files, key=lambda x: x[2])
    print(f"   [SYNC] Newest accounts: {newest[1]}")

    for loc, name in locations:
        if loc != newest[0]:
            if "Local" in name and localappdata_dir and not localappdata_dir.exists():
                continue
            try:
                copy_file(newest[0], loc)
                print(f"   [OK] -> {name}")
            except Exception as e:
                print(f"   [ERROR] -> {name}: {e}")

    sync_antigravity_json(config_dir, target_dir, appdata_dir)


def sync_antigravity_json(
    config_dir: Path,
    target_dir: Path,
    appdata_dir: Path | None,
):
    """Copy antigravity.json from config_dir (source of truth) to other locations."""
    source = config_dir / "antigravity.json"
    if not source.exists():
        print("   [SKIP] No antigravity.json in opencode-configs")
        return

    destinations = [target_dir / "antigravity.json"]
    if appdata_dir and str(appdata_dir) != "opencode":
        destinations.append(appdata_dir / "antigravity.json")

    for dst in destinations:
        if dst.parent.exists():
            try:
                copy_file(source, dst)
            except Exception:
                pass
