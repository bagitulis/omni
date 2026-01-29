"""
Docker Manager module for Omni Build System.

SRP: This module coordinates Docker operations using specialized sub-modules.
Actual implementations are in docker_desktop.py, docker_builder.py, docker_deployer.py.
"""
import subprocess
import time
from typing import Optional

from omni_build.config import Config
from omni_build.docker_builder import DockerBuilder
from omni_build.docker_deployer import DockerDeployer
from omni_build.docker_desktop import DockerDesktopManager
from omni_build.error_handler import ErrorHandler
from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.models import ContainerStatus, SpecLevel
from omni_build.output_parser import OutputParser


class DockerManager:
    """
    Manages Docker operations with retry logic and error recovery.
    
    Delegates to specialized sub-modules:
    - DockerDesktopManager: Docker Desktop process management
    - DockerBuilder: Image build operations
    - DockerDeployer: Container deployment operations
    """
    
    def __init__(self, config: Config, error_handler: ErrorHandler) -> None:
        self.config = config
        self.error_handler = error_handler
        self.output_parser = OutputParser(error_handler.error_patterns)
        
        # Specialized managers
        self._desktop = DockerDesktopManager()
        self._builder = DockerBuilder(config, error_handler, self.output_parser)
        self._deployer = DockerDeployer(config, error_handler)
    
    def check_docker_ready(self) -> bool:
        """
        Check if Docker Desktop is running and ready.
        AUTO-START if not running, AUTO-FIX if errors.
        """
        log_info("Checking Docker Desktop status...")
        
        if not self._desktop.is_running():
            log_info("Docker Desktop not running, starting it...")
            if not self._desktop.start():
                log_error("Failed to start Docker Desktop")
                return False
            
            log_info("Waiting for Docker Desktop cold start...")
            if not self._desktop.wait_for_ready(timeout_seconds=120):
                log_warning("Docker Desktop took too long, attempting repair...")
                return self._auto_fix_docker()
            
            log_success("Docker Desktop started successfully")
            return True
        
        log_info("Docker Desktop process found, checking engine...")
        
        if self._desktop.check_engine_ready():
            log_success("Docker engine is ready")
            return True
        
        log_info("Docker engine not ready, waiting...")
        if self._desktop.wait_for_ready(timeout_seconds=60):
            log_success("Docker engine is now ready")
            return True
        
        log_warning("Docker engine not responding, attempting repair...")
        return self._auto_fix_docker()
    
    def _auto_fix_docker(self) -> bool:
        """Auto-fix Docker issues by restarting engine."""
        log_info("Applying Docker auto-fix...")
        
        if self.error_handler._repair_docker_engine():
            log_success("Docker engine restarted successfully")
            
            log_info("Waiting for Docker API (max 90 seconds)...")
            
            max_attempts = 9
            consecutive_timeouts = 0
            
            for attempt in range(1, max_attempts + 1):
                time.sleep(10)
                elapsed_time = attempt * 10
                
                try:
                    result = subprocess.run(
                        ["docker", "info"],
                        capture_output=True,
                        text=True,
                        timeout=15,
                    )
                    
                    if result.returncode == 0:
                        log_success(f"Docker API ready after {elapsed_time}s")
                        
                        try:
                            verify = subprocess.run(
                                ["docker", "ps"],
                                capture_output=True,
                                timeout=10,
                            )
                            if verify.returncode == 0:
                                log_success("Docker fully operational")
                                return True
                        except Exception:
                            pass
                    
                    consecutive_timeouts = 0
                    log_info(f"Docker API initializing... ({elapsed_time}s)")
                    
                except subprocess.TimeoutExpired:
                    consecutive_timeouts += 1
                    if consecutive_timeouts >= 3:
                        log_error("Docker API not responding")
                        return False
                except Exception as e:
                    log_warning(f"Unexpected error: {e}")
            
            log_error(f"Docker API not ready after {max_attempts * 10}s")
            return False
        
        log_error("Failed to restart Docker engine")
        return False
    
    def check_linux_mode(self) -> bool:
        """Verify Docker is in Linux containers mode."""
        try:
            result = subprocess.run(
                ["docker", "version", "--format", "{{.Server.Os}}"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            
            if result.returncode == 0:
                os_type = result.stdout.strip().lower()
                if os_type == "linux":
                    log_success("Docker is in Linux containers mode")
                    return True
                else:
                    log_warning(f"Docker is in {os_type} mode - should be Linux")
                    return False
            
            return True  # Assume OK
            
        except Exception as e:
            log_warning(f"Could not verify Docker mode: {e}")
            return True
    
    # === DELEGATE TO SPECIALIZED MANAGERS ===
    
    def build_images(
        self,
        spec: SpecLevel,
        no_cache: bool = False,
        max_retries: Optional[int] = None,
    ) -> bool:
        """Build Docker images - delegates to DockerBuilder."""
        return self._builder.build_images(spec, no_cache, max_retries)
    
    def deploy_containers(
        self,
        spec: SpecLevel,
        max_retries: Optional[int] = None,
    ) -> bool:
        """Deploy containers - delegates to DockerDeployer."""
        return self._deployer.deploy_containers(spec, max_retries)
    
    # === CONTAINER LIFECYCLE OPERATIONS ===
    
    def stop_containers(self, spec: SpecLevel) -> bool:
        """Stop all containers for given spec."""
        compose_files = self.config.get_compose_files(spec)
        log_info("Stopping containers...")
        
        cmd = ["docker-compose"]
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.extend(["down", "--remove-orphans"])
        
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=60,
                cwd=str(self.config.project_root),
            )
            
            if result.returncode == 0:
                log_success("Containers stopped")
            else:
                log_warning(f"Stop completed with warnings: {result.stderr}")
            return True
                
        except Exception as e:
            log_error(f"Error stopping containers: {e}")
            return False
    
    def restart_containers(self, spec: SpecLevel) -> bool:
        """Restart containers using docker-compose restart."""
        compose_files = self.config.get_compose_files(spec)
        log_info("Restarting containers...")
        
        cmd = ["docker-compose"]
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.append("restart")
        
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=120,
                cwd=str(self.config.project_root),
            )
            
            if result.returncode == 0:
                log_success("Containers restarted")
                return True
            else:
                log_warning(f"Restart completed with warnings: {result.stderr}")
                return False
                
        except Exception as e:
            log_error(f"Error restarting containers: {e}")
            return False
    
    def check_containers_running(self, spec: SpecLevel) -> bool:
        """Check if main containers are running."""
        try:
            for container in [self.config.container_backend, self.config.container_frontend]:
                result = subprocess.run(
                    ["docker", "inspect", "--format", "{{.State.Status}}", container],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                if result.returncode != 0 or result.stdout.strip() != "running":
                    return False
            return True
        except Exception as e:
            log_error(f"Error checking container status: {e}")
            return False
    
    def get_container_status(self) -> list[ContainerStatus]:
        """Get status of all Omni containers."""
        try:
            result = subprocess.run(
                [
                    "docker", "ps", "-a",
                    "--filter", "name=omni-",
                    "--format", "{{.Names}}\t{{.Status}}\t{{.State}}\t{{.CreatedAt}}",
                ],
                capture_output=True,
                text=True,
                timeout=10,
            )
            
            if result.returncode != 0:
                return []
            
            containers = []
            for line in result.stdout.strip().split("\n"):
                if not line:
                    continue
                
                parts = line.split("\t")
                if len(parts) >= 4:
                    containers.append(ContainerStatus(
                        name=parts[0],
                        status=parts[1],
                        health=parts[2] if "healthy" in parts[1].lower() else None,
                        created=parts[3],
                    ))
            
            return containers
            
        except Exception as e:
            log_warning(f"Could not get container status: {e}")
            return []
