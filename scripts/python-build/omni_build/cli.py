"""
CLI Entry Point for Omni Build System.

SRP: This module ONLY provides the CLI entry point.
Commands are in commands.py, orchestration in orchestrator.py.
"""
from omni_build.commands import cli, print_result
from omni_build.orchestrator import BuildOrchestrator

# Re-export for backward compatibility
__all__ = ["cli", "BuildOrchestrator", "print_result"]

if __name__ == "__main__":
    cli()
