"""
Database Restore module for Omni Build System.

SRP: This module ONLY handles database restore operations.
Uses native Python implementation for cross-platform compatibility.
"""
import json
import subprocess
import sys
import time
from pathlib import Path
from typing import Optional, Tuple, List

from omni_build.config import Config
from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseRestorer:
    """
    Handles database restoration from backups.
    
    Responsibilities:
    - Check if backup exists
    - Verify PostgreSQL container is running and healthy
    - Wait for PostgreSQL to be ready with retry logic
    - Execute restore-smart.ps1 script
    - Auto-start PostgreSQL container if needed
    - Report restore status
    """
    
    def __init__(self, config: Config) -> None:
        """Initialize database restorer with config."""
        self.config = config
        self.backup_dir = config.project_root / "backups"
        # Check new location first (backups/), fallback to old (backups/smart/)
        new_manifest = self.backup_dir / "manifest.json"
        old_manifest = self.backup_dir / "smart" / "manifest.json"
        
        if new_manifest.exists():
            # New structure: backups/manifest.json, backups/_schema/, backups/tenant_*/
            self.data_dir = self.backup_dir
            self.manifest_file = new_manifest
        else:
            # Old structure: backups/smart/manifest.json
            self.data_dir = self.backup_dir / "smart"
            self.manifest_file = old_manifest
    
    def has_backup(self) -> bool:
        """Check if backup manifest exists."""
        return self.manifest_file.exists()
    
    def check_container_exists(self, container_name: str) -> bool:
        """Check if container exists (running or stopped)."""
        try:
            result = subprocess.run(
                ["docker", "inspect", container_name],
                capture_output=True,
                text=True,
                timeout=10,
            )
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def check_postgres_running(self) -> bool:
        """Check if PostgreSQL container is running and ready."""
        try:
            result = subprocess.run(
                ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def check_postgres_healthy(self) -> bool:
        """Check if PostgreSQL container is healthy."""
        try:
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Health.Status}}", "omni-postgres"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            return result.returncode == 0 and result.stdout.strip() == "healthy"
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False
    
    def start_postgres_container(self) -> bool:
        """Start PostgreSQL container if stopped."""
        log_info("Starting PostgreSQL container...")
        try:
            result = subprocess.run(
                ["docker", "start", "omni-postgres"],
                capture_output=True,
                text=True,
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
    
    def wait_for_postgres(self, timeout: int = 120) -> bool:
        """
        Wait for PostgreSQL container to be ready with robust retry logic.
        
        Args:
            timeout: Maximum wait time in seconds (default: 120s for fresh start)
            
        Returns:
            True if PostgreSQL is ready, False if timeout
        """
        log_info(f"Waiting for PostgreSQL to be ready (max {timeout}s)...")
        
        start_time = time.time()
        check_interval = 5
        last_status = ""
        
        while time.time() - start_time < timeout:
            elapsed = int(time.time() - start_time)
            
            # Check container status
            try:
                # Check if container is running
                result = subprocess.run(
                    ["docker", "inspect", "--format", 
                     "{{.State.Status}}|{{if .State.Health}}{{.State.Health.Status}}{{else}}no-healthcheck{{end}}",
                     "omni-postgres"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                
                if result.returncode != 0:
                    # Container doesn't exist, try to start with docker-compose
                    log_warning("PostgreSQL container not found, creating...")
                    self._start_postgres_with_compose()
                    time.sleep(check_interval)
                    continue
                
                status_parts = result.stdout.strip().split("|")
                running_status = status_parts[0] if len(status_parts) > 0 else "unknown"
                health_status = status_parts[1] if len(status_parts) > 1 else "unknown"
                
                current_status = f"{running_status}/{health_status}"
                
                if current_status != last_status:
                    log_info(f"PostgreSQL: {current_status} ({elapsed}s)")
                    last_status = current_status
                
                # Container not running - try to start it
                if running_status != "running":
                    if running_status == "exited":
                        self.start_postgres_container()
                    time.sleep(check_interval)
                    continue
                
                # Check if ready to accept connections
                if self.check_postgres_running():
                    # Verify with actual query
                    try:
                        verify = subprocess.run(
                            ["docker", "exec", "omni-postgres", 
                             "psql", "-U", "omni", "-d", "omni_main", "-c", "SELECT 1"],
                            capture_output=True,
                            text=True,
                            timeout=10,
                        )
                        if verify.returncode == 0:
                            log_success(f"PostgreSQL is ready ({elapsed}s)")
                            return True
                    except:
                        pass
                
            except Exception as e:
                log_warning(f"Error checking PostgreSQL: {e}")
            
            time.sleep(check_interval)
        
        log_error(f"PostgreSQL not ready after {timeout}s")
        return False
    
    def _start_postgres_with_compose(self) -> bool:
        """Start PostgreSQL using docker-compose."""
        try:
            result = subprocess.run(
                ["docker-compose", "-f", "docker-compose.tunnel.yml", 
                 "-f", "docker-compose.tunnel.standard.yml",
                 "up", "-d", "postgres"],
                capture_output=True,
                text=True,
                timeout=60,
                cwd=str(self.config.project_root),
            )
            return result.returncode == 0
        except Exception as e:
            log_error(f"Failed to start PostgreSQL with compose: {e}")
            return False
    
    def restore(self, force: bool = True, keep_extra: bool = False, timeout: int = 1800) -> Tuple[bool, str]:
        """
        Execute database restore from smart backup using native Python.
        
        Args:
            force: Skip confirmations (for automation)
            keep_extra: Keep tables not in backup (default: False = sync mode)
            timeout: Timeout in seconds (default: 30 minutes)
            
        Returns:
            Tuple of (success, message)
        """
        if not self.has_backup():
            return False, "No backup found. Run: python build.py backup"
        
        # Wait for PostgreSQL to be ready
        log_info("Ensuring PostgreSQL is ready...")
        if not self.wait_for_postgres(timeout=120):
            return False, "PostgreSQL not ready after 120s"
        
        # Get backup info
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
        
        # Step 1: Recreate schemas from _schema dumps (ensures matching schema)
        schema_success = self._recreate_schemas_from_backup()
        if not schema_success:
            log_warning("Schema recreation had issues - continuing with data restore")
        
        return self._restore_native(info)
    
    def _recreate_schemas_from_backup(self) -> bool:
        """Recreate schemas from backup _schema dumps to match backup data structure."""
        import gzip
        
        schema_dir = self.data_dir / "_schema"
        if not schema_dir.exists():
            log_warning("No _schema directory found - using existing database schema")
            return True
        
        schema_files = list(schema_dir.glob("*.sql.gz"))
        if not schema_files:
            return True
        
        log_info(f"Recreating {len(schema_files)} schemas from backup...")
        
        for schema_file in schema_files:
            schema_name = schema_file.stem.replace('.sql', '')
            log_info(f"  Recreating schema: {schema_name}")
            
            try:
                # Check PostgreSQL is still healthy before each schema operation
                if not self._ensure_postgres_healthy():
                    log_error("PostgreSQL crashed - attempting recovery...")
                    if not self._recover_postgres():
                        return False
                
                # Drop existing schema with timeout and error handling
                drop_cmd = ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                           "-d", "omni_main", "-c", f"DROP SCHEMA IF EXISTS {schema_name} CASCADE;"]
                drop_result = subprocess.run(drop_cmd, capture_output=True, text=True, timeout=60)
                
                # Check if container crashed during DROP
                if not self._ensure_postgres_healthy():
                    log_error(f"PostgreSQL crashed during DROP SCHEMA {schema_name}")
                    if not self._recover_postgres():
                        return False
                    # Skip this schema recreation since container was reset
                    continue
                
                # Recreate from dump with UTF-8
                with gzip.open(schema_file, 'rb') as f:
                    sql_bytes = f.read()
                sql_content = sql_bytes.decode('utf-8')
                
                result = subprocess.run(
                    ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main"],
                    input=sql_content,
                    capture_output=True,
                    text=True,
                    timeout=120,
                    encoding='utf-8',
                    errors='replace',
                )
                
                if result.returncode != 0:
                    log_warning(f"  Schema {schema_name} errors: {result.stderr[:200]}")
                else:
                    log_success(f"  Schema {schema_name} recreated")
                    
            except subprocess.TimeoutExpired:
                log_error(f"  Timeout recreating {schema_name} - checking container health...")
                if not self._recover_postgres():
                    return False
            except Exception as e:
                log_warning(f"  Failed to recreate {schema_name}: {e}")
        
        return True
    
    def _ensure_postgres_healthy(self) -> bool:
        """Quick check if PostgreSQL container is running and responsive."""
        try:
            result = subprocess.run(
                ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni"],
                capture_output=True, text=True, timeout=5
            )
            return result.returncode == 0
        except Exception:
            return False
    
    def _recover_postgres(self) -> bool:
        """Attempt to recover PostgreSQL after a crash."""
        log_info("Attempting PostgreSQL recovery...")
        
        # Check container status
        try:
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Status}}", "omni-postgres"],
                capture_output=True, text=True, timeout=10
            )
            status = result.stdout.strip()
            
            if status == "exited":
                log_info("Container exited - attempting restart...")
                subprocess.run(["docker", "start", "omni-postgres"], 
                              capture_output=True, timeout=30)
                
                # Wait for healthy
                for i in range(30):
                    time.sleep(2)
                    if self._ensure_postgres_healthy():
                        log_success("PostgreSQL recovered successfully")
                        return True
                
                log_error("PostgreSQL failed to recover within 60 seconds")
                return False
            elif status == "running":
                # Container running but not responding - may need more time
                for i in range(10):
                    time.sleep(2)
                    if self._ensure_postgres_healthy():
                        return True
                return False
            else:
                log_error(f"Container in unexpected state: {status}")
                return False
                
        except Exception as e:
            log_error(f"Recovery failed: {e}")
            return False
    
    def _restore_native(self, manifest: dict) -> Tuple[bool, str]:
        """Native Python restore without PowerShell dependency."""
        import gzip
        
        schemas = manifest.get('schemas', {})
        restored_tables = 0
        restored_rows = 0
        errors: List[str] = []
        
        # Collect all tables to restore first
        tables_to_restore: List[Tuple] = []
        total_expected_rows = 0
        for schema_name, schema_data in schemas.items():
            schema_dir = self.data_dir / schema_name
            if not schema_dir.exists():
                log_warning(f"Schema directory not found: {schema_dir}")
                continue
            
            for table_info in schema_data.get('tables', []):
                table_name = table_info.get('name')
                expected_rows = table_info.get('rows', 0)
                if expected_rows > 0:
                    tables_to_restore.append((schema_name, table_name, expected_rows, schema_dir))
                    total_expected_rows += expected_rows
        
        # Clear all tables first (without CASCADE to avoid cross-table issues)
        log_info(f"Clearing {len(tables_to_restore)} tables...")
        for schema_name, table_name, _, _ in tables_to_restore:
            clear_cmd = ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                        "-d", "omni_main", "-c", f"TRUNCATE {schema_name}.{table_name};"]
            subprocess.run(clear_cmd, capture_output=True, timeout=30)
        
        # Now restore all tables with container health checks
        for schema_name, table_name, expected_rows, schema_dir in tables_to_restore:
            # Check container health before each table restore
            if not self._ensure_postgres_healthy():
                log_warning(f"PostgreSQL not healthy before {table_name} - attempting recovery...")
                if not self._recover_postgres():
                    errors.append(f"{schema_name}.{table_name} (container crash)")
                    continue
            
            chunk_files = sorted(schema_dir.glob(f"{table_name}.chunk*.sql.gz"))
            
            if chunk_files:
                success, rows = self._restore_chunked_files(
                    schema_name, table_name, chunk_files, expected_rows
                )
            else:
                sql_file = schema_dir / f"{table_name}.sql.gz"
                if not sql_file.exists():
                    log_warning(f"  Skip: {table_name} (file not found)")
                    continue
                success, rows = self._restore_single_table(
                    schema_name, table_name, sql_file, skip_truncate=True
                )
            
            if success:
                restored_tables += 1
                restored_rows += rows
            else:
                errors.append(f"{schema_name}.{table_name}")
        
        if errors:
            error_msg = f"Failed to restore {len(errors)} tables: {', '.join(errors[:5])}"
            log_warning(error_msg)
            # Return False if more than 10% of tables failed
            if len(errors) > len(tables_to_restore) * 0.1:
                return False, error_msg
        
        # Validate restore success
        if restored_rows == 0 and total_expected_rows > 0:
            return False, "Restore completed but 0 rows restored - possible schema mismatch"
        
        success_rate = (restored_rows / total_expected_rows * 100) if total_expected_rows > 0 else 100
        if success_rate < 90:
            log_warning(f"Restore success rate: {success_rate:.1f}% ({restored_rows:,}/{total_expected_rows:,} rows)")
        
        log_success(f"Restore complete: {restored_tables} tables, {restored_rows:,} rows")
        return True, f"Restored {restored_tables} tables with {restored_rows:,} rows"
    
    def _restore_single_table(self, schema: str, table: str, sql_file: Path, skip_truncate: bool = False, max_retries: int = 3) -> Tuple[bool, int]:
        """Restore single table from .sql.gz file with retry logic."""
        import gzip
        
        last_error = ""
        for attempt in range(1, max_retries + 1):
            try:
                if not skip_truncate:
                    # Clear table first
                    clear_cmd = ["docker", "exec", "omni-postgres", "psql", "-U", "omni", 
                                "-d", "omni_main", "-c", f"TRUNCATE {schema}.{table};"]
                    subprocess.run(clear_cmd, capture_output=True, timeout=30)
                
                # Read with binary and decode as UTF-8 (fix Windows encoding issue)
                with gzip.open(sql_file, 'rb') as f:
                    sql_bytes = f.read()
                sql_content = sql_bytes.decode('utf-8')
                
                # Execute via docker exec psql with explicit UTF-8 encoding
                result = subprocess.run(
                    ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main"],
                    input=sql_content,
                    capture_output=True,
                    text=True,
                    timeout=300,
                    encoding='utf-8',
                    errors='replace',
                )
                
                if result.returncode == 0:
                    # Count restored rows
                    count_result = subprocess.run(
                        ["docker", "exec", "omni-postgres", "psql", "-U", "omni", 
                         "-d", "omni_main", "-t", "-c", f"SELECT COUNT(*) FROM {schema}.{table};"],
                        capture_output=True,
                        text=True,
                        timeout=30,
                    )
                    rows = int(count_result.stdout.strip()) if count_result.returncode == 0 else 0
                    print(f"  [OK] {schema}.{table}: {rows:,} rows")
                    return True, rows
                else:
                    last_error = result.stderr[:100]
                    if attempt < max_retries:
                        time.sleep(2)  # Wait before retry
                    
            except subprocess.TimeoutExpired:
                last_error = "Timeout expired"
                if attempt < max_retries:
                    time.sleep(5)
            except Exception as e:
                last_error = str(e)
                if attempt < max_retries:
                    time.sleep(2)
        
        print(f"  [FAIL] {schema}.{table}: {last_error}")
        return False, 0
    
    def _restore_chunked_files(self, schema: str, table: str, chunk_files: list, expected: int) -> Tuple[bool, int]:
        """Restore large table from .chunk000.sql.gz files (raw TSV format) with retry."""
        import gzip
        
        try:
            # Table already truncated in batch before restore
            total_chunks = len(chunk_files)
            print(f"  [INFO] {schema}.{table}: restoring {total_chunks} chunks...")
            
            # Get column list for COPY command
            col_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main", "-t", "-c",
                 f"SELECT string_agg(column_name, ', ' ORDER BY ordinal_position) FROM information_schema.columns WHERE table_schema = '{schema}' AND table_name = '{table}';"],
                capture_output=True, text=True, timeout=30
            )
            columns = col_result.stdout.strip() if col_result.returncode == 0 else ""
            
            if not columns:
                print(f"  [ERROR] {schema}.{table}: could not get column list")
                return False, 0
            
            failed_chunks = 0
            for i, chunk_file in enumerate(chunk_files, 1):
                success = False
                for attempt in range(3):  # Retry each chunk up to 3 times
                    try:
                        # Read raw TSV data
                        with gzip.open(chunk_file, 'rb') as f:
                            tsv_bytes = f.read()
                        tsv_data = tsv_bytes.decode('utf-8')
                        
                        # Build COPY command with data
                        copy_sql = f"COPY {schema}.{table} ({columns}) FROM stdin;\n{tsv_data}\\.\n"
                        
                        result = subprocess.run(
                            ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main"],
                            input=copy_sql,
                            capture_output=True,
                            text=True,
                            timeout=600,
                            encoding='utf-8',
                            errors='replace',
                        )
                        
                        if result.returncode == 0:
                            success = True
                            break
                        else:
                            if attempt < 2:
                                time.sleep(2)
                    except subprocess.TimeoutExpired:
                        if attempt < 2:
                            time.sleep(5)
                    except Exception:
                        if attempt < 2:
                            time.sleep(2)
                
                if not success:
                    failed_chunks += 1
                    print(f"  [WARN] Chunk {i}/{total_chunks} failed after 3 attempts")
                elif i % 5 == 0 or i == total_chunks:
                    print(f"  [INFO] {schema}.{table}: {i}/{total_chunks} chunks done")
            
            if failed_chunks > 0:
                print(f"  [WARN] {failed_chunks}/{total_chunks} chunks failed")
            
            # Count final rows
            count_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main", "-t", "-c", f"SELECT COUNT(*) FROM {schema}.{table};"],
                capture_output=True,
                text=True,
                timeout=30,
            )
            rows = int(count_result.stdout.strip()) if count_result.returncode == 0 else 0
            print(f"  [OK] {schema}.{table}: {rows:,}/{expected:,} rows")
            return True, rows
            
        except Exception as e:
            print(f"  [ERROR] {schema}.{table}: {e}")
            return False, 0
    
    def get_backup_info(self) -> Optional[dict]:
        """Get backup manifest information."""
        if not self.has_backup():
            return None
        
        try:
            # Use utf-8-sig to handle BOM (created by PowerShell)
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
        total_tables = 0
        total_rows = 0
        for schema_name, schema_data in schemas.items():
            tables = schema_data.get('tables', [])
            total_tables += len(tables)
            for table in tables:
                total_rows += table.get('rows', 0)
        
        log_info(f"  Schemas: {len(schemas)}")
        log_info(f"  Total tables: {total_tables}")
        log_info(f"  Total rows: {total_rows:,}")
