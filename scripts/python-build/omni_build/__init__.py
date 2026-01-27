"""
Omni Build System - Smart Docker orchestration with auto-fix capabilities.

This package provides:
- Automated build and deployment for multi-service Docker applications
- Comprehensive error detection and recovery (22+ error patterns)
- Progressive fix levels (6 escalation levels)
- Frontend build automation
- Health checking and verification
- Structured logging with Rich console output

Usage:
    from omni_build.config import Config
    from omni_build.cli import BuildOrchestrator
    
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    success = orchestrator.execute_build(mode="smart", spec="standard")
"""

__version__ = "1.0.0"
__author__ = "Omni Team"

from omni_build.config import Config
from omni_build.docker_manager import DockerManager
from omni_build.error_handler import ErrorHandler
from omni_build.frontend_builder import FrontendBuilder
from omni_build.health_checker import HealthChecker
from omni_build.models import BuildMode, BuildResult, ErrorPattern, SpecLevel

__all__ = [
    "Config",
    "DockerManager",
    "ErrorHandler",
    "FrontendBuilder",
    "HealthChecker",
    "BuildMode",
    "BuildResult",
    "ErrorPattern",
    "SpecLevel",
]
