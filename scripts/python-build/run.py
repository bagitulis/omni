#!/usr/bin/env python3
"""
Standalone launcher for Omni Build System.
Adds scripts/python-build to sys.path so no pip install -e is needed.
"""
import sys
from pathlib import Path

# Add the omni_build package to Python path
SCRIPT_DIR = Path(__file__).parent
sys.path.insert(0, str(SCRIPT_DIR))

# Import and run CLI
from omni_build.cli import cli

if __name__ == "__main__":
    cli()
