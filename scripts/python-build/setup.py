#!/usr/bin/env python3
"""
Setup script for Omni Build System
Auto-creates venv and installs dependencies
"""
import os
import subprocess
import sys
from pathlib import Path

SCRIPT_DIR = Path(__file__).parent
VENV_DIR = SCRIPT_DIR / ".venv"
REQUIREMENTS = SCRIPT_DIR / "requirements.txt"


def create_venv() -> None:
    """Create virtual environment if it doesn't exist."""
    if VENV_DIR.exists():
        print(f"[OK] Virtual environment already exists: {VENV_DIR}")
        return
        
    print(f"[INFO] Creating virtual environment at {VENV_DIR}...")
    subprocess.run([sys.executable, "-m", "venv", str(VENV_DIR)], check=True)
    print("[OK] Virtual environment created successfully")


def get_pip_executable() -> Path:
    """Get pip executable path from venv."""
    if sys.platform == "win32":
        return VENV_DIR / "Scripts" / "pip.exe"
    return VENV_DIR / "bin" / "pip"


def install_dependencies() -> None:
    """Install dependencies from requirements.txt."""
    pip_exe = get_pip_executable()
    
    if not pip_exe.exists():
        raise FileNotFoundError(f"pip not found at {pip_exe}")
    
    print(f"[INFO] Installing dependencies from {REQUIREMENTS}...")
    
    # Try to upgrade pip (skip if fails - not critical)
    try:
        subprocess.run(
            [str(pip_exe), "install", "--upgrade", "pip"],
            check=False,  # Don't fail if pip upgrade fails
            capture_output=True,
        )
    except Exception:
        pass  # Pip upgrade not critical
    
    # Install requirements
    subprocess.run(
        [str(pip_exe), "install", "-r", str(REQUIREMENTS)],
        check=True
    )
    print("[OK] Dependencies installed successfully")


def main() -> None:
    """Main setup function."""
    print("=" * 60)
    print("  Omni Build System - Setup")
    print("=" * 60)
    print()
    
    try:
        # Step 1: Create venv
        create_venv()
        
        # Step 2: Install dependencies
        install_dependencies()
        
        print()
        print("=" * 60)
        print("  Setup Complete!")
        print("=" * 60)
        print()
        print("To use the build system:")
        if sys.platform == "win32":
            print("  .\\build.bat smart")
        else:
            print("  ./build.sh smart")
        print()
        
    except Exception as e:
        print(f"\n[ERROR] Setup failed: {e}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
