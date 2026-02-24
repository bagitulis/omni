"""
Configuration module for Omni Build System.

SRP: This module ONLY handles configuration loading and validation.
No business logic, no Docker operations.
"""
import os
from pathlib import Path
from typing import Optional

from dotenv import load_dotenv
from pydantic import BaseModel, ConfigDict, Field

from omni_build.models import DockerComposeConfig, FrontendBuildConfig, SpecLevel, SystemRequirements


class Config(BaseModel):
    """Central configuration for build system."""
    
    # Project paths
    project_root: Path = Field(..., description="Root directory of the project")
    scripts_dir: Path = Field(..., description="Scripts directory")
    frontend_dir: Path = Field(..., description="Frontend source directory")
    backend_dir: Path = Field(..., description="Backend source directory")
    
    # Docker Compose files
    compose_base: str = Field(default="docker-compose.tunnel.yml")
    compose_lowspec: str = Field(default="docker-compose.tunnel.lowspec.yml")
    compose_standard: str = Field(default="docker-compose.tunnel.standard.yml")
    compose_highspec: str = Field(default="docker-compose.tunnel.highspec.yml")
    
    # Timeouts (seconds) - Optimized for reliability
    # Not too fast (causes false failures), not too slow (wastes time)
    docker_start_timeout: int = Field(default=150, description="Docker engine start (2.5 min)")
    docker_build_timeout: int = Field(default=1800, description="Docker build (30 min - full build needs time)")
    docker_deploy_timeout: int = Field(default=900, description="Deployment (15 min - includes health waits)")
    health_check_timeout: int = Field(default=180, description="Health checks (3 min - Go backend needs tenant init)")
    npm_install_timeout: int = Field(default=600, description="npm install (10 min - large node_modules)")
    
    # Retry configuration
    max_build_retries: int = Field(default=7, description="Maximum build retry attempts")
    max_deploy_retries: int = Field(default=6, description="Maximum deploy retry attempts")
    max_fix_levels: int = Field(default=6, description="Progressive fix escalation levels")
    
    # System requirements
    system_requirements: SystemRequirements = Field(default_factory=SystemRequirements)
    
    # Frontend configuration
    frontend_config: FrontendBuildConfig = Field(default_factory=FrontendBuildConfig)
    
    # Container names (for health checks)
    container_backend: str = Field(default="omni-backend")
    container_frontend: str = Field(default="omni-frontend")
    container_nginx: str = Field(default="omni-nginx")
    container_postgres: str = Field(default="omni-postgres")
    container_redis: str = Field(default="omni-redis")
    
    # Health check endpoints
    backend_health_url: str = Field(default="http://localhost:3000/api/health")
    frontend_health_url: str = Field(default="http://localhost:80/")
    nginx_health_url: str = Field(default="http://localhost:80/health")
    
    # Environment
    go_env: str = Field(default="production")
    timezone: str = Field(default="Asia/Jakarta")
    
    model_config = ConfigDict(arbitrary_types_allowed=True)
    
    @classmethod
    def from_env(cls, project_root: Optional[Path] = None) -> "Config":
        """
        Create configuration from environment variables.
        
        Args:
            project_root: Project root directory. If None, auto-detect from script location.
            
        Returns:
            Config instance with values loaded from .env and defaults.
        """
        if project_root is None:
            # Auto-detect: scripts/python-build/omni_build -> project root is 3 levels up
            script_dir = Path(__file__).parent  # omni_build/
            project_root = script_dir.parent.parent.parent  # scripts/python-build/ -> scripts/ -> omni/
        
        # Load .env file from project root
        env_file = project_root / ".env"
        if env_file.exists():
            load_dotenv(env_file)
        
        # Build configuration
        return cls(
            project_root=project_root,
            scripts_dir=project_root / "scripts" / "python-build",
            frontend_dir=project_root / "frontend",
            backend_dir=project_root / "backend",
            
            # Override from environment variables
            docker_start_timeout=int(os.getenv("DOCKER_START_TIMEOUT", "150")),
            docker_build_timeout=int(os.getenv("DOCKER_BUILD_TIMEOUT", "1800")),
            docker_deploy_timeout=int(os.getenv("DOCKER_DEPLOY_TIMEOUT", "900")),
            health_check_timeout=int(os.getenv("HEALTH_CHECK_TIMEOUT", "120")),
            npm_install_timeout=int(os.getenv("NPM_INSTALL_TIMEOUT", "600")),
            
            max_build_retries=int(os.getenv("MAX_BUILD_RETRIES", "7")),
            max_deploy_retries=int(os.getenv("MAX_DEPLOY_RETRIES", "6")),
            max_fix_levels=int(os.getenv("MAX_FIX_LEVELS", "6")),
            
            backend_health_url=os.getenv("BACKEND_HEALTH_URL", "http://localhost:3000/api/health"),
            frontend_health_url=os.getenv("FRONTEND_HEALTH_URL", "http://localhost:80/"),
            nginx_health_url=os.getenv("NGINX_HEALTH_URL", "http://localhost:80/health"),
            
            go_env=os.getenv("GO_ENV", "production"),
            timezone=os.getenv("TZ", "Asia/Jakarta"),
        )
    
    def get_compose_files(self, spec: SpecLevel) -> list[str]:
        """
        Get Docker Compose file list for given spec level.
        
        Args:
            spec: Specification level (lowspec/standard/highspec)
            
        Returns:
            List of compose file paths relative to project root.
        """
        base_file = str(self.project_root / self.compose_base)
        
        spec_map = {
            SpecLevel.LOWSPEC: self.compose_lowspec,
            SpecLevel.STANDARD: self.compose_standard,
            SpecLevel.HIGHSPEC: self.compose_highspec,
        }
        
        spec_file = str(self.project_root / spec_map[spec])
        
        return [base_file, spec_file]
    
    def get_docker_compose_config(self, spec: SpecLevel) -> DockerComposeConfig:
        """
        Get Docker Compose configuration for given spec level.
        
        Args:
            spec: Specification level
            
        Returns:
            DockerComposeConfig instance.
        """
        files = self.get_compose_files(spec)
        
        return DockerComposeConfig(
            base_file=files[0],
            spec_file=files[1],
            build_timeout=self.docker_build_timeout,
            deploy_timeout=self.docker_deploy_timeout,
            max_retries=self.max_build_retries,
        )
