#!/bin/bash
# Omni Build System - Linux/Mac Wrapper
# Auto-setup and run Python CLI

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PYTHON_BUILD_DIR="$SCRIPT_DIR/scripts/python-build"
VENV_DIR="$PYTHON_BUILD_DIR/.venv"
VENV_PYTHON="$VENV_DIR/bin/python"

# Check if virtual environment exists
if [ ! -d "$VENV_DIR" ]; then
    echo "[SETUP] Virtual environment not found - running setup..."
    python3 "$PYTHON_BUILD_DIR/setup.py"
fi

# Run CLI with all arguments
"$VENV_PYTHON" -m omni_build.cli "$@"
