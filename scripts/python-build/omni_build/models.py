"""
Data models for Omni Build System.

SRP: This module ONLY defines data structures and validation.
No business logic, no I/O operations.
"""
from enum import Enum
from typing import Callable, Optional

from pydantic import BaseModel, ConfigDict, Field


class BuildMode(str, Enum):
    """Build execution modes."""
    QUICKFIX = "quickfix"  # Fix service issues (~30-60s) - FIRST OPTION
    QUICK = "quick"      # Restart only (~10s) - DEPRECATED, use QUICKFIX
    SMART = "smart"      # Cache deps, rebuild code (~2-3m) [RECOMMENDED for code changes]
    FULL = "full"        # Clean rebuild (~5-10m)
    VALIDATE = "validate"  # Check only (~5s)
    CLEAN = "clean"      # Cleanup only (~1-2m)


class SpecLevel(str, Enum):
    """Server resource specification levels."""
    LOWSPEC = "lowspec"      # 2GB RAM - VPS kecil
    STANDARD = "standard"    # 4GB RAM - Production [RECOMMENDED]
    HIGHSPEC = "highspec"    # 8GB+ RAM - High traffic


class ErrorSeverity(str, Enum):
    """Error severity levels."""
    LOW = "low"
    MEDIUM = "medium"
    HIGH = "high"
    CRITICAL = "critical"


class ErrorPattern(BaseModel):
    """Error pattern definition for auto-detection."""
    name: str = Field(..., description="Error pattern name (e.g., 'AlpineRepoError')")
    pattern: str = Field(..., description="Regex pattern to match error message")
    description: str = Field(..., description="Human-readable error description")
    severity: ErrorSeverity = Field(default=ErrorSeverity.MEDIUM)
    fix_function: Optional[str] = Field(default=None, description="Fix function name to call")
    is_code_error: bool = Field(default=False, description="Cannot auto-fix (requires manual intervention)")
    
    model_config = ConfigDict(use_enum_values=True)


class BuildResult(BaseModel):
    """Result of a build operation."""
    success: bool = Field(..., description="Whether build succeeded")
    mode: BuildMode = Field(..., description="Build mode executed")
    spec: SpecLevel = Field(..., description="Spec level used")
    duration_seconds: float = Field(..., description="Total execution time")
    errors: list[str] = Field(default_factory=list, description="Error messages encountered")
    warnings: list[str] = Field(default_factory=list, description="Warning messages")
    attempts: int = Field(default=1, description="Number of retry attempts")
    containers_deployed: int = Field(default=0, description="Number of containers successfully deployed")
    health_checks_passed: int = Field(default=0, description="Number of health checks passed")
    
    model_config = ConfigDict(use_enum_values=True)


class ContainerStatus(BaseModel):
    """Docker container status."""
    name: str = Field(..., description="Container name")
    status: str = Field(..., description="Container status (running, exited, etc)")
    health: Optional[str] = Field(None, description="Health status")
    created: Optional[str] = Field(None, description="Creation timestamp")
    
    
class HealthCheckResult(BaseModel):
    """Health check result for a service."""
    service: str = Field(..., description="Service name")
    endpoint: Optional[str] = Field(default=None, description="Health endpoint URL")
    status: str = Field(..., description="Status (healthy, unhealthy, unknown)")
    response_time_ms: Optional[float] = Field(default=None, description="Response time in milliseconds")
    error: Optional[str] = Field(default=None, description="Error message if unhealthy")


class DockerComposeConfig(BaseModel):
    """Docker Compose configuration."""
    base_file: str = Field(default="docker-compose.tunnel.yml", description="Base compose file")
    spec_file: str = Field(..., description="Spec-specific override file")
    services: list[str] = Field(default_factory=list, description="Services to build/deploy")
    build_timeout: int = Field(default=1200, description="Build timeout in seconds")
    deploy_timeout: int = Field(default=600, description="Deploy timeout in seconds")
    max_retries: int = Field(default=7, description="Maximum retry attempts")


class FrontendBuildConfig(BaseModel):
    """Frontend build configuration."""
    directory: str = Field(default="frontend", description="Frontend directory path")
    package_manager: str = Field(default="npm", description="Package manager (npm, pnpm, yarn)")
    build_command: str = Field(default="npm run build", description="Build command")
    install_command: str = Field(default="npm ci --legacy-peer-deps", description="Install command")
    output_directory: str = Field(default="dist", description="Build output directory")
    max_retries: int = Field(default=3, description="Maximum build retry attempts")
    cache_enabled: bool = Field(default=True, description="Whether to check for unchanged dependencies")


class SystemRequirements(BaseModel):
    """System requirements validation."""
    min_disk_gb: float = Field(default=5.0, description="Minimum free disk space in GB")
    min_memory_gb: float = Field(default=1.0, description="Minimum free memory in GB")
    required_ports: list[int] = Field(
        default_factory=lambda: [80, 443, 3000, 5432, 6379],
        description="Ports that must be available"
    )
    docker_required: bool = Field(default=True, description="Docker Desktop must be running")
    docker_linux_mode: bool = Field(default=True, description="Docker must be in Linux mode")
