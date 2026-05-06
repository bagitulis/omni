"""
Database Restore Core module for Omni Build System.

SRP: This module handles core database restore operations.
Schema and chunk operations are in database_restore_ops.py.
"""
import json
import subprocess
import time
from pathlib import Path
from typing import List, Optional, Tuple

from omni_build.config import Config
from omni_build.db_config import DatabaseConfig
from omni_build.logger import log_error, log_info, log_success, log_warning


class PostgresHealthChecker:
    """Handles PostgreSQL container health checks and recovery."""
    
    def __init__(self, config: Config):
        self.config = config
    
    def check_running(self) -> bool:
        """Check if PostgreSQL container is running and ready."""
        try:
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + [
                    "pg_isready", "-U", DatabaseConfig.USER, "-d", DatabaseConfig.DATABASE
                ],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def check_healthy(self) -> bool:
        """Check if PostgreSQL container is healthy."""
        try:
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Health.Status}}", DatabaseConfig.CONTAINER_NAME],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            return result.returncode == 0 and result.stdout.strip() == "healthy"
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def ensure_healthy(self) -> bool:
        """Quick check if PostgreSQL container is running and responsive."""
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
    
    def start_container(self) -> bool:
        """Start PostgreSQL container if stopped."""
        log_info("Starting PostgreSQL container...")
        try:
            result = subprocess.run(
                ["docker", "start", DatabaseConfig.CONTAINER_NAME],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=30,
            )
            if result.returncode == 0:
                log_success("PostgreSQL container started")
                return True
            else:
                log_error(f"Failed to start PostgreSQL: {result.stderr}")
                return False
        except Exception as e:
            log_error(f"Error starting PostgreSQL: {e}")
            return False
    
    def recover(self) -> bool:
        """Attempt to recover PostgreSQL after a crash."""
        log_info("Attempting PostgreSQL recovery...")
        
        try:
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Status}}", DatabaseConfig.CONTAINER_NAME],
                capture_output=True, text=True, encoding='utf-8', errors='replace', timeout=10
            )
            status = result.stdout.strip()
            
            if status == "exited":
                log_info("Container exited - attempting restart...")
                subprocess.run(["docker", "start", DatabaseConfig.CONTAINER_NAME], 
                              capture_output=True, timeout=30)
                
                for _ in range(30):
                    time.sleep(2)
                    if self.ensure_healthy():
                        log_success("PostgreSQL recovered successfully")
                        return True
                
                log_error("PostgreSQL failed to recover within 60 seconds")
                return False
            elif status == "running":
                for _ in range(10):
                    time.sleep(2)
                    if self.ensure_healthy():
                        return True
                return False
            else:
                log_error(f"Container in unexpected state: {status}")
                return False
                
        except Exception as e:
            log_error(f"Recovery failed: {e}")
            return False
    
    def wait_for_ready(self, timeout: int = 120) -> bool:
        """Wait for PostgreSQL container to be ready."""
        log_info(f"Waiting for PostgreSQL to be ready (max {timeout}s)...")
        
        start_time = time.time()
        check_interval = 5
        last_status = ""
        
        while time.time() - start_time < timeout:
            elapsed = int(time.time() - start_time)
            
            try:
                result = subprocess.run(
                    ["docker", "inspect", "--format", 
                     "{{.State.Status}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}",
                     DatabaseConfig.CONTAINER_NAME],
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=10,
                )
                
                if result.returncode != 0:
                    log_warning("PostgreSQL container not found, creating...")
                    self._start_with_compose()
                    time.sleep(check_interval)
                    continue
                
                parts = result.stdout.strip().split("|")
                running_status = parts[0] if parts else "unknown"
                health_status = parts[1] if len(parts) > 1 else "unknown"
                
                current_status = f"{running_status}/{health_status}"
                
                if current_status != last_status:
                    log_info(f"PostgreSQL: {current_status} ({elapsed}s)")
                    last_status = current_status
                
                if running_status != "running":
                    if running_status == "exited":
                        self.start_container()
                    time.sleep(check_interval)
                    continue
                
                if self.check_running():
                    try:
                        verify = subprocess.run(
                            DatabaseConfig.psql_cmd() + ["-c", "SELECT 1"],
                            capture_output=True, text=True, encoding='utf-8', errors='replace', timeout=10,
                        )
                        if verify.returncode == 0:
                            log_success(f"PostgreSQL is ready ({elapsed}s)")
                            return True
                    except Exception:
                        pass
                
            except Exception as e:
                log_warning(f"Error checking PostgreSQL: {e}")
            
            time.sleep(check_interval)
        
        log_error(f"PostgreSQL not ready after {timeout}s")
        return False
    
    def _start_with_compose(self) -> bool:
        """Start PostgreSQL using docker-compose."""
        try:
            from omni_build.subprocess_utils import get_compose_command
            cmd = get_compose_command() + [
                "-f", "docker-compose.tunnel.yml",
                "-f", "docker-compose.tunnel.standard.yml",
                "up", "-d", "postgres",
            ]
            result = subprocess.run(
                cmd,
                capture_output=True, text=True, encoding='utf-8', errors='replace', timeout=60,
            )
            return result.returncode == 0
        except Exception as e:
            log_error(f"Failed to start PostgreSQL with compose: {e}")
            return False


class DatabaseRestorer:
    """
    Handles database restoration from backups.
    
    Responsibilities:
    - Check if backup exists
    - Verify PostgreSQL container is running and healthy
    - Execute restore operations
    """
    
    def __init__(self, config: Config) -> None:
        self.config = config
        self.backup_dir = config.project_root / "backups" / "smart"
        self.data_dir = self.backup_dir
        self.manifest_file = self.backup_dir / "manifest.json"
        self._pg_checker = PostgresHealthChecker(config)
    
    def has_backup(self) -> bool:
        """Check if backup manifest exists."""
        return self.manifest_file.exists()
    
    def check_container_exists(self, container_name: str) -> bool:
        """Check if container exists (running or stopped)."""
        try:
            result = subprocess.run(
                ["docker", "inspect", container_name],
                capture_output=True, text=True, encoding='utf-8', errors='replace', timeout=10,
            )
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def check_postgres_running(self) -> bool:
        """Check if PostgreSQL container is running and ready."""
        return self._pg_checker.check_running()
    
    def check_postgres_healthy(self) -> bool:
        """Check if PostgreSQL container is healthy."""
        return self._pg_checker.check_healthy()
    
    def start_postgres_container(self) -> bool:
        """Start PostgreSQL container if stopped."""
        return self._pg_checker.start_container()
    
    def wait_for_postgres(self, timeout: int = 120) -> bool:
        """Wait for PostgreSQL container to be ready."""
        return self._pg_checker.wait_for_ready(timeout)
    
    def get_backup_info(self) -> Optional[dict]:
        """Get backup manifest information."""
        if not self.has_backup():
            return None
        
        try:
            with open(self.manifest_file, 'r', encoding='utf-8-sig') as f:
                return json.load(f)
        except Exception as e:
            log_warning(f"Could not read manifest: {e}")
            return None
    
    def print_backup_status(self) -> None:
        """Print backup status information."""
        info = self.get_backup_info()
        
        if not info:
            log_warning("No backup found in backups/")
            log_info("Run: python build.py backup")
            return
        
        log_info("Backup Status:")
        log_info(f"  Last backup: {info.get('exported_at', 'Unknown')}")
        log_info(f"  Version: {info.get('version', 'Unknown')}")
        
        schemas = info.get('schemas', {})
        total_tables = sum(len(s.get('tables', [])) for s in schemas.values())
        total_rows = sum(t.get('rows', 0) for s in schemas.values() for t in s.get('tables', []))
        
        log_info(f"  Schemas: {len(schemas)}")
        log_info(f"  Total tables: {total_tables}")
        log_info(f"  Total rows: {total_rows:,}")
    
    def restore(self, force: bool = True, keep_extra: bool = False, 
                timeout: int = 1800) -> Tuple[bool, str]:
        """Execute database restore from smart backup."""
        if not self.has_backup():
            return False, "No backup found. Run: python build.py backup"
        
        log_info("Ensuring PostgreSQL is ready...")
        if not self.wait_for_postgres(timeout=120):
            return False, "PostgreSQL not ready after 120s"
        
        info = self.get_backup_info()
        if not info:
            return False, "Could not read backup manifest"
        
        schemas = info.get('schemas', {})
        total_tables = sum(len(s.get('tables', [])) for s in schemas.values())
        total_rows = sum(t.get('rows', 0) for s in schemas.values() for t in s.get('tables', []))
        
        log_info("Restore Details:")
        log_info(f"  Backup from: {info.get('exported_at', 'Unknown')}")
        log_info(f"  Schemas: {len(schemas)}, Tables: {total_tables}, Rows: {total_rows:,}")
        
        log_info("Starting Python-native database restore...")
        
        # Import restore operations module
        from omni_build.database_restore_ops import DatabaseRestoreOps
        ops = DatabaseRestoreOps(self.config, self._pg_checker)
        
        # Terminate active connections to prevent lock conflicts
        ops.terminate_active_connections()

        schema_success = ops.recreate_schemas_from_backup(self.data_dir)
        if not schema_success:
            log_warning("Schema recreation had issues - continuing with data restore")
        
        success, message = ops.restore_tables(info, self.data_dir)

        # Run ANALYZE after successful restore to update statistics
        if success:
            ops.run_post_restore_analyze()

        return success, message
