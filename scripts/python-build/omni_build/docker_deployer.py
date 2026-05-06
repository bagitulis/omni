"""
Docker Deploy Operations module for Omni Build System.

SRP: This module ONLY handles Docker container deployment operations.
"""
import subprocess
import time
from typing import Optional

from omni_build.config import Config
from omni_build.error_handler import ErrorHandler
from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.models import SpecLevel


class DockerDeployer:
    """Handles Docker container deployment with retry logic."""
    
    def __init__(self, config: Config, error_handler: ErrorHandler) -> None:
        self.config = config
        self.error_handler = error_handler
    
    def deploy_containers(
        self,
        spec: SpecLevel,
        max_retries: Optional[int] = None,
    ) -> bool:
        """
        Deploy containers with docker-compose up.
        
        Args:
            spec: Specification level
            max_retries: Maximum retry attempts
            
        Returns:
            True if deployment succeeded, False otherwise
        """
        if max_retries is None:
            max_retries = self.config.max_deploy_retries
        
        compose_files = self.config.get_compose_files(spec)
        log_info(f"Deploying containers (spec: {spec.value})...")
        
        from omni_build.subprocess_utils import get_compose_command
        cmd = get_compose_command()
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.extend(["up", "-d", "--remove-orphans"])
        
        for attempt in range(1, max_retries + 1):
            log_info(f"Deploy attempt {attempt}/{max_retries}...")
            
            try:
                result = subprocess.run(
                    cmd,
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=self.config.docker_deploy_timeout,
                    cwd=str(self.config.project_root),
                )
                
                error_output = result.stderr + "\n" + result.stdout
                log_info(f"Docker compose completed (exit code: {result.returncode})")
                
                log_info("Checking container status...")
                time.sleep(5)
                
                core_status = self._check_core_containers_status()
                
                if core_status['postgres_running'] and core_status['backend_exists']:
                    log_info("Core containers running, waiting for health...")
                    
                    if self._wait_for_all_containers_healthy(spec, timeout=240):
                        log_success("All containers are healthy!")
                        self._restart_nginx_after_deploy(spec)
                        return True
                    else:
                        log_warning("Some containers not healthy after 240s")
                        self._log_container_health_status()
                        
                        if attempt < max_retries:
                            if self._wait_for_all_containers_healthy(spec, timeout=120):
                                log_success("Containers finally healthy!")
                                self._restart_nginx_after_deploy(spec)
                                return True
                
                if result.returncode == 0:
                    log_warning("Docker compose succeeded but containers not healthy")
                    self._log_container_health_status()
                    
                    if attempt < max_retries:
                        log_info("Restarting unhealthy containers...")
                        self._restart_unhealthy_containers()
                        continue
                
                log_error(f"Deploy attempt {attempt} failed")
                
                error_pattern = self.error_handler.detect_error(error_output)
                
                if error_pattern:
                    log_warning(f"Error detected: {error_pattern.description}")
                    if attempt < max_retries:
                        if self.error_handler.apply_fix(error_pattern):
                            continue
                
                if attempt < max_retries:
                    fix_level = min(attempt, self.config.max_fix_levels)
                    log_info(f"Applying progressive fix level {fix_level}...")
                    self.error_handler.apply_progressive_fix(fix_level)
                else:
                    log_error("Max retries reached - deployment failed")
                    self._log_container_health_status()
                    return False
                
            except subprocess.TimeoutExpired:
                log_error(f"Deploy timed out after {self.config.docker_deploy_timeout}s")
                if attempt < max_retries:
                    self.error_handler.apply_progressive_fix(3)
                else:
                    return False
            
            except Exception as e:
                log_error(f"Deploy error: {e}")
                if attempt >= max_retries:
                    return False
        
        return False
    
    def _check_core_containers_status(self) -> dict:
        """Check status of core containers."""
        status = {
            'postgres_running': False,
            'redis_running': False,
            'backend_exists': False,
            'frontend_exists': False,
        }
        
        try:
            containers = [
                ('omni-postgres', 'postgres_running'),
                ('omni-redis', 'redis_running'),
                ('omni-backend', 'backend_exists'),
                ('omni-frontend', 'frontend_exists'),
            ]
            for container, key in containers:
                result = subprocess.run(
                    ["docker", "inspect", "--format", "{{.State.Status}}", container],
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=10,
                )
                if result.returncode == 0:
                    state = result.stdout.strip()
                    if state in ['running', 'starting']:
                        status[key] = True
        except Exception as e:
            log_warning(f"Error checking container status: {e}")
        
        return status
    
    def _restart_unhealthy_containers(self) -> None:
        """Restart any unhealthy containers."""
        try:
            result = subprocess.run(
                ["docker", "ps", "--filter", "health=unhealthy", "--format", "{{.Names}}"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            if result.returncode == 0 and result.stdout.strip():
                unhealthy = result.stdout.strip().split('\n')
                for container in unhealthy:
                    log_info(f"Restarting unhealthy container: {container}")
                    subprocess.run(
                        ["docker", "restart", container], 
                        capture_output=True, 
                        timeout=30
                    )
                time.sleep(10)
        except Exception as e:
            log_warning(f"Error restarting unhealthy containers: {e}")
    
    def _wait_for_all_containers_healthy(
        self, 
        spec: SpecLevel, 
        timeout: int = 180
    ) -> bool:
        """Wait for all critical containers to be healthy."""
        critical_containers = [
            ("omni-postgres", True),
            ("omni-redis", True),
            ("omni-backend", True),
            ("omni-frontend", False),
        ]
        
        start_time = time.time()
        check_interval = 10
        
        while (time.time() - start_time) < timeout:
            all_healthy = True
            
            for container, must_be_healthy in critical_containers:
                try:
                    result = subprocess.run(
                        ["docker", "inspect", "--format", 
                         "{{.State.Status}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}",
                         container],
                        capture_output=True,
                        text=True,
                        encoding='utf-8',
                        errors='replace',
                        timeout=10,
                    )
                    
                    if result.returncode != 0:
                        log_warning(f"{container}: not found")
                        all_healthy = False
                        continue
                    
                    parts = result.stdout.strip().split("|")
                    running_status = parts[0] if parts else "unknown"
                    health_status = parts[1] if len(parts) > 1 else "unknown"
                    
                    if running_status != "running":
                        log_warning(f"{container}: {running_status}")
                        all_healthy = False
                        continue
                    
                    if must_be_healthy and health_status not in ["healthy", "no-healthcheck"]:
                        if health_status == "starting":
                            log_info(f"{container}: starting...")
                        else:
                            log_warning(f"{container}: {health_status}")
                        all_healthy = False
                        
                except Exception as e:
                    log_warning(f"Error checking {container}: {e}")
                    all_healthy = False
            
            if all_healthy:
                return True
            
            elapsed = int(time.time() - start_time)
            log_info(f"Waiting for containers... ({elapsed}s / {timeout}s)")
            time.sleep(check_interval)
        
        return False
    
    def _log_container_health_status(self) -> None:
        """Log detailed health status of all containers."""
        log_info("Container health status:")
        try:
            result = subprocess.run(
                ["docker", "ps", "-a", "--filter", "name=omni-",
                 "--format", "  {{.Names}}: {{.Status}}"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            if result.returncode == 0 and result.stdout.strip():
                print(result.stdout)
        except Exception as e:
            log_warning(f"Could not get container status: {e}")
    
    def _restart_nginx_after_deploy(self, spec: SpecLevel) -> bool:
        """Restart nginx to refresh DNS cache."""
        try:
            compose_files = self.config.get_compose_files(spec)
            log_info("Restarting nginx to refresh DNS cache...")
            
            cmd = get_compose_command()
            for file in compose_files:
                cmd.extend(["-f", file])
            cmd.extend(["restart", "nginx"])
            
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=30,
                cwd=str(self.config.project_root),
            )
            
            if result.returncode == 0:
                log_success("Nginx restarted - DNS cache refreshed")
                return True
            else:
                log_warning(f"Nginx restart failed: {result.stderr}")
                return False
                
        except Exception as e:
            log_warning(f"Error restarting nginx: {e}")
            return False
