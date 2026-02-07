#!/usr/bin/env python3
"""
AI.py - OpenCode Provider Switcher v3.1
Python version with full functionality

5 Modes: Copilot, Antigravity Proxy/Plugin, Mix Proxy/Plugin
Smart Sync: 4 locations for antigravity configs
Dynamic Transform: For plugin modes
"""

import os
import sys
import shutil
import socket
import subprocess
from pathlib import Path
from datetime import datetime

# Paths
SCRIPT_DIR = Path(__file__).parent
CONFIG_DIR = SCRIPT_DIR / "opencode-configs"
TARGET_DIR = Path.home() / ".config" / "opencode"
APPDATA_DIR = Path(os.environ.get("APPDATA", "")) / "opencode"
LOCALAPPDATA_DIR = Path(os.environ.get("LOCALAPPDATA", "")) / "opencode"


def clear_screen():
    os.system('cls' if os.name == 'nt' else 'clear')


def transform_for_plugin(content: str) -> str:
    """Transform model names for plugin mode (google/* -> google/antigravity-*)"""
    content = content.replace('google/claude-', 'google/antigravity-claude-')
    content = content.replace('google/gemini-', 'google/antigravity-gemini-')
    return content


def check_proxy_running(port: int = 8045) -> bool:
    """Check if proxy is running on specified port"""
    try:
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.settimeout(1)
            result = s.connect_ex(('127.0.0.1', port))
            return result == 0
    except:
        return False


def get_file_mtime(filepath: Path) -> float:
    """Get file modification time, return 0 if not exists"""
    try:
        return filepath.stat().st_mtime if filepath.exists() else 0
    except:
        return 0


def copy_file(src: Path, dst: Path):
    """Copy file, creating parent directories if needed"""
    dst.parent.mkdir(parents=True, exist_ok=True)
    shutil.copy2(src, dst)


def smart_sync_accounts():
    """Sync antigravity-accounts.json across all locations - newest wins"""
    locations = [
        (CONFIG_DIR / "antigravity-accounts.json", "opencode-configs"),
        (TARGET_DIR / "antigravity-accounts.json", ".config/opencode"),
        (APPDATA_DIR / "antigravity-accounts.json", "AppData/Roaming/opencode"),
        (LOCALAPPDATA_DIR / "antigravity-accounts.json", "AppData/Local/opencode"),
    ]
    
    # Ensure directories exist
    TARGET_DIR.mkdir(parents=True, exist_ok=True)
    APPDATA_DIR.mkdir(parents=True, exist_ok=True)
    
    # Find newest file
    mtimes = [(loc, name, get_file_mtime(loc)) for loc, name in locations]
    valid_files = [(loc, name, mtime) for loc, name, mtime in mtimes if mtime > 0]
    
    if not valid_files:
        print("   [SKIP] No antigravity-accounts.json found in any location")
        return
    
    newest = max(valid_files, key=lambda x: x[2])
    print(f"   [SYNC] Newest accounts: {newest[1]}")
    
    # Copy to all other locations
    for loc, name in locations:
        if loc != newest[0]:
            # Skip LOCALAPPDATA if directory doesn't exist
            if "Local" in name and not LOCALAPPDATA_DIR.exists():
                continue
            try:
                copy_file(newest[0], loc)
                print(f"   [OK] -> {name}")
            except Exception as e:
                print(f"   [ERROR] -> {name}: {e}")
    
    # Also sync antigravity.json
    sync_antigravity_json()


def sync_antigravity_json():
    """Sync antigravity.json across locations"""
    locations = [
        CONFIG_DIR / "antigravity.json",
        TARGET_DIR / "antigravity.json",
        APPDATA_DIR / "antigravity.json",
    ]
    
    mtimes = [(loc, get_file_mtime(loc)) for loc in locations]
    valid_files = [(loc, mtime) for loc, mtime in mtimes if mtime > 0]
    
    if not valid_files:
        return
    
    newest = max(valid_files, key=lambda x: x[1])
    
    for loc, _ in mtimes:
        if loc != newest[0] and loc.parent.exists():
            try:
                copy_file(newest[0], loc)
            except:
                pass


def detect_current_provider() -> str:
    """Detect current provider from oh-my-opencode.json"""
    config_file = TARGET_DIR / "oh-my-opencode.json"
    
    if not config_file.exists():
        return "[None]"
    
    try:
        content = config_file.read_text(encoding='utf-8')
        
        has_copilot = 'github-copilot' in content
        has_antigravity = 'google/antigravity-' in content
        has_google = 'google/' in content
        
        if has_copilot:
            if has_antigravity:
                return "Mix Plugin"
            elif has_google:
                return "Mix Proxy"
            else:
                return "Copilot"
        elif has_antigravity:
            return "Antigravity Plugin"
        elif has_google:
            return "Antigravity Proxy"
        
        return "[Unknown]"
    except:
        return "[Error]"


def apply_config(provider: str, ohmyopencode_file: str, opencode_file: str,
                 need_proxy: bool, need_plugin: bool, transform_plugin: bool):
    """Apply the selected configuration"""
    
    src_ohmyopencode = CONFIG_DIR / ohmyopencode_file
    
    if not src_ohmyopencode.exists():
        print(f"\n   [ERROR] Config not found: {ohmyopencode_file}")
        input("   Press Enter to continue...")
        return False
    
    # Check proxy if needed
    if need_proxy:
        if not check_proxy_running(8045):
            print("\n   [WARNING] Antigravity Proxy NOT running on port 8045!")
            print("   Please start Antigravity Tools and enable proxy first.")
            response = input("\n   Continue anyway? [y/N]: ").strip().lower()
            if response != 'y':
                return False
    
    # Ensure target directory exists
    TARGET_DIR.mkdir(parents=True, exist_ok=True)
    
    # Copy/Transform oh-my-opencode config
    dst_ohmyopencode = TARGET_DIR / "oh-my-opencode.json"
    
    if transform_plugin:
        # Transform for plugin mode
        content = src_ohmyopencode.read_text(encoding='utf-8')
        transformed = transform_for_plugin(content)
        dst_ohmyopencode.write_text(transformed, encoding='utf-8')
        print(f"   [OK] Transformed {ohmyopencode_file} -> oh-my-opencode.json (plugin mode)")
    else:
        copy_file(src_ohmyopencode, dst_ohmyopencode)
        print(f"   [OK] Copied {ohmyopencode_file} -> oh-my-opencode.json")
    
    # Copy opencode.json if specified
    if opencode_file:
        src_opencode = CONFIG_DIR / opencode_file
        if src_opencode.exists():
            copy_file(src_opencode, TARGET_DIR / "opencode.json")
            print(f"   [OK] Copied {opencode_file} -> opencode.json")
    
    # Copy antigravity.json for non-Copilot modes
    if provider != "Copilot":
        src_antigravity = CONFIG_DIR / "antigravity.json"
        if src_antigravity.exists():
            copy_file(src_antigravity, TARGET_DIR / "antigravity.json")
            print("   [OK] Copied antigravity.json")
    
    # Smart sync accounts for plugin modes
    if need_plugin:
        smart_sync_accounts()
    
    print(f"\n   [OK] Switched to {provider}")
    if need_proxy:
        print("   [INFO] Using Antigravity Proxy at localhost:8045")
    if need_plugin:
        print("   [INFO] Using Antigravity Auth Plugin")
    
    return True


def start_opencode():
    """Start OpenCode"""
    print("\n   Starting OpenCode...")
    print()
    exe_path = SCRIPT_DIR / "opencode.exe"
    try:
        if exe_path.exists():
            subprocess.run([str(exe_path)])
        else:
            print(f"   [ERROR] Failed to start OpenCode: {e}")
    except Exception as e:
        print(f"   [ERROR] Failed to start OpenCode: {e}")


def show_menu():
    """Display the main menu"""
    clear_screen()
    print()
    print("  +==================================================================+")
    print("  |           AI.py - OpenCode Provider Switcher v3.1               |")
    print("  +==================================================================+")
    print("  |                                                                  |")
    print("  |   [1] Copilot            - GitHub Copilot Only                   |")
    print("  |                                                                  |")
    print("  |   [2] Antigravity Proxy  - Via localhost:8045 (clean names)      |")
    print("  |   [3] Antigravity Plugin - Via Auth Plugin (antigravity-* names) |")
    print("  |                                                                  |")
    print("  |   [4] Mix Proxy          - Copilot + Antigravity Proxy           |")
    print("  |   [5] Mix Plugin         - Copilot + Antigravity Plugin          |")
    print("  |                                                                  |")
    print("  |   [S] Sync               - Sync accounts across 4 locations      |")
    print("  |   [Q] Quit                                                       |")
    print("  |                                                                  |")
    print("  +==================================================================+")
    print()
    
    current = detect_current_provider()
    print(f"   Current Provider: {current}")
    print()


# Provider configurations
PROVIDERS = {
    '1': {
        'name': 'Copilot',
        'ohmyopencode': 'oh-my-opencode-copilot.json',
        'opencode': '',
        'need_proxy': False,
        'need_plugin': False,
        'transform': False,
    },
    '2': {
        'name': 'Antigravity Proxy',
        'ohmyopencode': 'oh-my-opencode-antigravity.json',
        'opencode': 'opencode-proxy.json',
        'need_proxy': True,
        'need_plugin': False,
        'transform': False,
    },
    '3': {
        'name': 'Antigravity Plugin',
        'ohmyopencode': 'oh-my-opencode-antigravity.json',
        'opencode': 'opencode-plugin.json',
        'need_proxy': False,
        'need_plugin': True,
        'transform': True,
    },
    '4': {
        'name': 'Mix Proxy',
        'ohmyopencode': 'oh-my-opencode-mix.json',
        'opencode': 'opencode-proxy.json',
        'need_proxy': True,
        'need_plugin': False,
        'transform': False,
    },
    '5': {
        'name': 'Mix Plugin',
        'ohmyopencode': 'oh-my-opencode-mix.json',
        'opencode': 'opencode-plugin.json',
        'need_proxy': False,
        'need_plugin': True,
        'transform': True,
    },
}


def main():
    # Handle command line arguments
    if len(sys.argv) > 1:
        arg = sys.argv[1].lower()
        arg_map = {
            'copilot': '1',
            'antigravity-proxy': '2',
            'antigravity-plugin': '3',
            'mix-proxy': '4',
            'mix-plugin': '5',
            'sync': 's',
            'current': 'c',
        }
        choice = arg_map.get(arg, arg)
    else:
        show_menu()
        choice = input("   Select [1-5, S, Q]: ").strip().lower()
    
    while True:
        if choice == 'q':
            break
        elif choice == 's':
            print("\n   Syncing accounts across all locations...")
            smart_sync_accounts()
            print()
            input("   Press Enter to continue...")
        elif choice == 'c':
            current = detect_current_provider()
            print(f"\n   Current Provider: {current}")
            print()
            input("   Press Enter to continue...")
        elif choice in PROVIDERS:
            p = PROVIDERS[choice]
            success = apply_config(
                provider=p['name'],
                ohmyopencode_file=p['ohmyopencode'],
                opencode_file=p['opencode'],
                need_proxy=p['need_proxy'],
                need_plugin=p['need_plugin'],
                transform_plugin=p['transform'],
            )
            if success:
                start_opencode()
            break
        else:
            print("\n   Invalid choice!")
        
        # Show menu again for interactive mode
        if len(sys.argv) <= 1:
            show_menu()
            choice = input("   Select [1-5, S, Q]: ").strip().lower()
        else:
            break


if __name__ == '__main__':
    main()
