#!/usr/bin/env python3
"""
Traffic Split Control Script - Cross-platform.

Manages gradual migration from Node.js to Go backend.
Replaces traffic-split.ps1 with a Python equivalent that works on Windows and Linux.

Usage:
    python traffic-split.py              # Show current status
    python traffic-split.py 10           # Set 10% to Go (testing)
    python traffic-split.py 50           # Set 50% to Go (parallel)
    python traffic-split.py 100          # Set 100% to Go (full migration)
    python traffic-split.py rollback     # Rollback to 0% Go
"""
import re
import subprocess
import sys
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
CONFIG_FILE = SCRIPT_DIR / "nginx-traffic-split.conf"

VALID_ACTIONS = ("status", "10", "50", "100", "rollback")


def show_banner():
    print()
    print("╔══════════════════════════════════════════════════════════╗")
    print("║         TRAFFIC SPLIT CONTROLLER: Node.js → Go           ║")
    print("╚══════════════════════════════════════════════════════════╝")
    print()


def get_current_split():
    """Read current traffic split percentage from nginx config."""
    if not CONFIG_FILE.exists():
        print(f"  ERROR: Config file not found: {CONFIG_FILE}")
        sys.exit(1)

    content = CONFIG_FILE.read_text(encoding="utf-8")

    # Match percentage-based split
    match = re.search(r"(?m)^split_clients.*\{[\s\S]*?(\d+)%\s+go_backend", content)
    if match:
        return int(match.group(1))

    # Match wildcard (100%) split
    if re.search(r"(?m)^split_clients.*\{\s*\*\s+go_backend", content):
        return 100

    return 0


def set_traffic_split(percentage):
    """Update nginx config to set traffic split percentage."""
    content = CONFIG_FILE.read_text(encoding="utf-8")

    # Comment out all split_clients blocks
    content = re.sub(r"(?m)^(split_clients)", r"# \1", content)

    # Uncomment the appropriate block
    if percentage == 10:
        content = re.sub(
            r"(?m)^# (split_clients.*\{[\s\S]*?10%\s+go_backend[\s\S]*?\})",
            r"\1", content,
        )
    elif percentage == 50:
        content = re.sub(
            r"(?m)^# (split_clients.*\{[\s\S]*?50%\s+go_backend[\s\S]*?\})",
            r"\1", content,
        )
    elif percentage == 100:
        content = re.sub(
            r"(?m)^# (split_clients.*\{\s*\*\s+go_backend[\s\S]*?\})",
            r"\1", content,
        )
    # percentage == 0 means all blocks stay commented (rollback)

    CONFIG_FILE.write_text(content, encoding="utf-8")


def reload_nginx():
    """Test and reload nginx configuration via Docker."""
    print("  Reloading nginx configuration...")

    # Test config first
    result = subprocess.run(
        ["docker", "exec", "omni-nginx", "nginx", "-t"],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        print("  ✗ Nginx config test failed!")
        print(result.stderr)
        sys.exit(1)

    # Reload
    result = subprocess.run(
        ["docker", "exec", "omni-nginx", "nginx", "-s", "reload"],
        capture_output=True, text=True,
    )
    if result.returncode != 0:
        print(f"  ✗ Nginx reload failed: {result.stderr}")
        sys.exit(1)

    print("  ✓ Nginx reloaded successfully")


def show_status():
    """Display current traffic distribution."""
    current_split = get_current_split()

    print("  Current Traffic Distribution:")
    print()

    go_bar = "█" * (current_split // 5)
    node_bar = "█" * ((100 - current_split) // 5)

    print(f"  Go Backend:      {go_bar} {current_split}%")
    print(f"  Node.js Backend: {node_bar} {100 - current_split}%")
    print()
    print("  Commands:")
    print("    python traffic-split.py 10       # Set 10% to Go (testing)")
    print("    python traffic-split.py 50       # Set 50% to Go (parallel)")
    print("    python traffic-split.py 100      # Set 100% to Go (full migration)")
    print("    python traffic-split.py rollback # Rollback to 0% Go")
    print()


def main():
    action = sys.argv[1] if len(sys.argv) > 1 else "status"

    if action not in VALID_ACTIONS:
        print(f"  ERROR: Invalid action '{action}'")
        print(f"  Valid actions: {', '.join(VALID_ACTIONS)}")
        sys.exit(1)

    show_banner()

    if action == "status":
        show_status()
    elif action == "10":
        print("  Setting traffic split to 10% Go / 90% Node.js...")
        set_traffic_split(10)
        reload_nginx()
        show_status()
    elif action == "50":
        print("  Setting traffic split to 50% Go / 50% Node.js...")
        set_traffic_split(50)
        reload_nginx()
        show_status()
    elif action == "100":
        print("  Setting traffic split to 100% Go / 0% Node.js...")
        print()
        print("  ⚠️  WARNING: This will route ALL traffic to Go backend!")
        confirm = input("  Type 'yes' to confirm: ").strip()
        if confirm == "yes":
            set_traffic_split(100)
            reload_nginx()
            show_status()
        else:
            print("  Cancelled.")
    elif action == "rollback":
        print("  Rolling back to 0% Go (all traffic to Node.js)...")
        set_traffic_split(0)
        reload_nginx()
        show_status()


if __name__ == "__main__":
    main()
