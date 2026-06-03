"""
PostgreSQL health check and recovery module for Omni Build System.

SRP: This module handles PostgreSQL container health monitoring and recovery.
"""
import subprocess
import time
from typing import Callable, List

from omni_build.db_config import DatabaseConfig
from omni_build.container_runtime import get_runtime
from omni_build.logger import log_info, log_success


class PostgresHealthManager:
    """Manage PostgreSQL container health checks and recovery."""
    
    def __init__(self, error_callback: Callable[[str], None]) -> None:
        """
        Initialize health manager.
        
        Args:
            error_callback: Function to call when errors occur (e.g., append to errors list)
        """
        self._add_error = error_callback
    
    def ensure_postgres_healthy(self) -> bool:
        """Quick check if PostgreSQL is responsive."""
        try:
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + [
                    "pg_isready", "-U", DatabaseConfig.USER, "-d", DatabaseConfig.DATABASE
                ],
                capture_output=True, text=True, timeout=5
            )
            return result.returncode == 0
        except Exception:
            return False
    
    def recover_postgres(self) -> bool:
        """Attempt to recover PostgreSQL after a crash."""
        log_info("Attempting PostgreSQL recovery...")
        
        try:
            result = subprocess.run(
                get_runtime().inspect(DatabaseConfig.CONTAINER_NAME, "{{.State.Status}}")
                capture_output=True, text=True, timeout=10
            )
            status = result.stdout.strip()
            
            if status == "exited":
                log_info("Container exited - attempting restart...")
                subprocess.run(get_runtime().start_container(DatabaseConfig.CONTAINER_NAME),
                              capture_output=True, timeout=30)
                
                for _ in range(30):
                    time.sleep(2)
                    if self.ensure_postgres_healthy():
                        log_success("PostgreSQL recovered successfully")
                        return True
                
                self._add_error("PostgreSQL failed to recover within 60 seconds")
                return False
            elif status == "running":
                for _ in range(10):
                    time.sleep(2)
                    if self.ensure_postgres_healthy():
                        return True
                return False
            else:
                self._add_error(f"Container in unexpected state: {status}")
                return False
                
        except Exception as e:
            self._add_error(f"Recovery failed: {e}")
            return False
