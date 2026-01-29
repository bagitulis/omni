"""
Build Orchestrator Helpers for Omni Build System.

SRP: This module contains helper functions and verification logic for build orchestration.
"""
import subprocess
import time
from typing import Optional

from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.models import ErrorPattern


class ErrorTracker:
    """Track repeated errors to detect stuck situations."""
    
    def __init__(self, max_same_error: int = 5):
        self.max_same_error = max_same_error
        self.error_history: list[str] = []
        self.last_error: Optional[str] = None
        self.same_error_count: int = 0
    
    def record_error(self, error_name: str) -> bool:
        """
        Record an error occurrence.
        
        Returns:
            True if should continue, False if same error repeated too many times
        """
        self.error_history.append(error_name)
        
        if error_name == self.last_error:
            self.same_error_count += 1
        else:
            self.last_error = error_name
            self.same_error_count = 1
        
        if self.same_error_count >= self.max_same_error:
            return False
        
        return True
    
    def get_repeated_error(self) -> Optional[str]:
        """Get the error that was repeated too many times."""
        if self.same_error_count >= self.max_same_error:
            return self.last_error
        return None
    
    def reset(self):
        """Reset error tracking."""
        self.error_history = []
        self.last_error = None
        self.same_error_count = 0


# All 8 services with their container names
ALL_SERVICES = [
    ("omni-postgres", "postgres", True),
    ("omni-redis", "redis", True),
    ("omni-backend", "backend", True),
    ("omni-frontend", "frontend", True),
    ("omni-nginx", "nginx", True),
    ("omni-cloudflared", "cloudflared", False),
    ("omni-pgbouncer", "pgbouncer", False),
    ("omni-pgweb", "pgweb", False),
]


class OrchestratorHelpers:
    """Helper methods for build orchestration."""
    
    def __init__(self, config, docker_manager, health_checker, error_handler):
        self.config = config
        self.docker_manager = docker_manager
        self.health_checker = health_checker
        self.error_handler = error_handler
    
    def safe_docker_check(self) -> bool:
        """Check Docker with connection retry handling."""
        for attempt in range(3):
            try:
                return self.docker_manager.check_docker_ready()
            except Exception as e:
                if attempt < 2:
                    log_warning(f"Docker check failed, retrying... ({attempt+1}/3)")
                    time.sleep(5)
                else:
                    log_error(f"Docker check failed: {e}")
                    return False
        return False
    
    def verify_all_services_truly_healthy(self) -> bool:
        """
        Double-check all services are TRULY healthy.
        Prevents false positive where status says OK but actually not working.
        """
        log_info("Verifying services are truly healthy...")
        
        if not self.health_checker.wait_for_all_services(timeout=90):
            log_warning("Container health check failed")
            return False
        
        try:
            import requests
            response = requests.get(
                self.config.backend_health_url,
                timeout=10
            )
            if response.status_code != 200:
                log_warning(f"Backend health returned {response.status_code}")
                return False
            
            try:
                data = response.json()
                if not isinstance(data, dict):
                    log_warning("Backend health response is not valid JSON object")
                    return False
            except Exception:
                pass
            
        except Exception as e:
            if "ConnectionError" in type(e).__name__:
                log_warning("Backend not reachable")
            elif "Timeout" in type(e).__name__:
                log_warning("Backend health timeout")
            else:
                log_warning(f"Backend health check error: {e}")
            return False
        
        if not self._verify_backend_db_connection():
            log_warning("Backend database connection verification failed")
            return False
        
        log_success("All services verified truly healthy!")
        return True
    
    def _verify_backend_db_connection(self) -> bool:
        """Verify backend can connect to database by checking logs."""
        try:
            result = subprocess.run(
                ["docker", "logs", "--tail", "30", "omni-backend"],
                capture_output=True, text=True, timeout=10,
            )
            
            if result.returncode != 0:
                return True
            
            logs = result.stdout + result.stderr
            
            db_errors = [
                "no such host",
                "connection refused",
                "failed to connect",
                "database error",
                "FATAL",
            ]
            
            logs_lower = logs.lower()
            for error in db_errors:
                if error.lower() in logs_lower:
                    log_warning(f"Found DB error in logs: {error}")
                    return False
            
            return True
            
        except Exception:
            return True
    
    def diagnose_services(self) -> list[str]:
        """Diagnose which services are problematic."""
        problematic = []
        
        for container, name, is_critical in ALL_SERVICES:
            try:
                result = subprocess.run(
                    ["docker", "inspect", "--format",
                     "{{.State.Status}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}",
                     container],
                    capture_output=True, text=True, timeout=10,
                )
                
                if result.returncode != 0:
                    if is_critical:
                        problematic.append(name)
                    continue
                
                parts = result.stdout.strip().split("|")
                status = parts[0] if parts else "unknown"
                health = parts[1] if len(parts) > 1 else "unknown"
                
                if status != "running":
                    if is_critical:
                        problematic.append(name)
                elif health == "unhealthy":
                    problematic.append(name)
                    
            except subprocess.TimeoutExpired:
                log_warning(f"{name}: timeout checking status")
                if is_critical:
                    problematic.append(name)
            except Exception:
                if is_critical:
                    problematic.append(name)
        
        return problematic
    
    def detect_service_errors(self, services: list[str]) -> Optional[ErrorPattern]:
        """Detect error patterns from service logs."""
        logs = self.collect_service_logs(services)
        return self.error_handler.detect_error(logs) if logs else None
    
    def collect_service_logs(self, services: list[str]) -> str:
        """Collect logs from services with connection retry."""
        container_map = {s[1]: s[0] for s in ALL_SERVICES}
        combined = ""
        
        for service in services:
            container = container_map.get(service, f"omni-{service}")
            for attempt in range(2):
                try:
                    result = subprocess.run(
                        ["docker", "logs", "--tail", "50", container],
                        capture_output=True, text=True, timeout=10,
                    )
                    if result.returncode == 0:
                        combined += f"=== {service} ===\n{result.stdout}{result.stderr}\n"
                    break
                except subprocess.TimeoutExpired:
                    if attempt == 0:
                        time.sleep(2)
                        continue
                except Exception:
                    break
        
        return combined
    
    def restart_service(self, service: str) -> bool:
        """Restart a service with connection retry."""
        container_map = {s[1]: s[0] for s in ALL_SERVICES}
        container = container_map.get(service, f"omni-{service}")
        
        for attempt in range(2):
            try:
                log_info(f"Restarting {service}...")
                result = subprocess.run(
                    ["docker", "restart", container],
                    capture_output=True, text=True, timeout=60,
                )
                if result.returncode == 0:
                    log_success(f"{service} restarted")
                    return True
                return False
            except subprocess.TimeoutExpired:
                if attempt == 0:
                    log_warning(f"Restart {service} timeout, retrying...")
                    continue
                return False
            except Exception as e:
                log_error(f"Error restarting {service}: {e}")
                return False
        
        return False
