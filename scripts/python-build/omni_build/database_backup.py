"""
Database Backup module for Omni Build System.

SRP: This module handles database backup coordination.
Export operations are in database_backup_ops.py.
"""
import json
import re
import shutil
import subprocess
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from omni_build.backup_change_detector import BackupChangeDetector
from omni_build.backup_manifest import BackupManifest
from omni_build.config import Config
from omni_build.db_config import DatabaseConfig
from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseBackup:
    """
    Handles database backup operations.
    
    Features:
    - Change detection (only backup changed tables)
    - Checksum-based comparison  
    - Chunked export for large tables (>50k rows)
    - Schema structure export
    """
    
    def __init__(self, config: Config) -> None:
        self.config = config
        self.backup_dir = config.project_root / "backups" / "smart"
        self._manifest = BackupManifest(self.backup_dir)
    
    def check_postgres_ready(self) -> bool:
        """Check if PostgreSQL is running and ready."""
        try:
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + [
                    "pg_isready", "-U", DatabaseConfig.USER, "-d", DatabaseConfig.DATABASE
                ],
                capture_output=True, text=True, timeout=10
            )
            return result.returncode == 0
        except Exception:
            return False
    
    def start_postgres_if_needed(self) -> bool:
        """Start PostgreSQL container if not running."""
        if self.check_postgres_ready():
            return True
        
        log_info("Starting PostgreSQL container...")
        try:
            subprocess.run(["docker", "start", DatabaseConfig.CONTAINER_NAME], 
                          capture_output=True, timeout=30)
            
            for _ in range(30):
                if self.check_postgres_ready():
                    log_success("PostgreSQL is ready")
                    return True
                time.sleep(1)
            
            return False
        except Exception as e:
            log_error(f"Failed to start PostgreSQL: {e}")
            return False
    
    def check_pg_version_compatibility(self) -> Tuple[bool, str]:
        """Verify pg_dump version >= PostgreSQL server version."""
        try:
            dump_result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + ["pg_dump", "--version"],
                capture_output=True, text=True, timeout=10
            )
            if dump_result.returncode != 0:
                return True, ""  # Skip check if can't determine
            
            server_result = subprocess.run(
                DatabaseConfig.psql_cmd() + ["-t", "-A", "-c", "SHOW server_version"],
                capture_output=True, text=True, timeout=10
            )
            if server_result.returncode != 0:
                return True, ""
            
            dump_match = re.search(r'(\d+)', dump_result.stdout)
            server_match = re.search(r'^(\d+)', server_result.stdout.strip())
            
            if dump_match and server_match:
                dump_major = int(dump_match.group(1))
                server_major = int(server_match.group(1))
                if dump_major < server_major:
                    return False, f"pg_dump v{dump_major} < server v{server_major}"
            
            return True, ""
        except Exception as e:
            log_warning(f"Version check skipped: {e}")
            return True, ""
    
    def check_disk_space(self) -> Tuple[bool, str]:
        """Verify sufficient disk space (2x database size)."""
        try:
            _db = DatabaseConfig.DATABASE
            result = subprocess.run(
                DatabaseConfig.psql_cmd() + [
                    "-t", "-A", "-c",
                    f"SELECT pg_database_size('{_db}')"
                ],
                capture_output=True, text=True, timeout=30
            )
            if result.returncode != 0:
                return True, ""
            
            db_size = int(result.stdout.strip())
            
            stat = shutil.disk_usage(self.backup_dir)
            available = stat.free
            required = db_size * 2
            
            if available < required:
                avail_gb = available / (1024**3)
                req_gb = required / (1024**3)
                return False, f"Disk space: {avail_gb:.1f}GB available < {req_gb:.1f}GB required"
            
            return True, ""
        except Exception as e:
            log_warning(f"Disk space check skipped: {e}")
            return True, ""
    
    def get_all_tables(self) -> List[Dict[str, Any]]:
        """Get all tables with row counts from tenant and system schemas."""
        query = """
        SELECT schemaname, relname,
               COALESCE(n_tup_ins, 0)::bigint,
               COALESCE(n_tup_upd, 0)::bigint,
               COALESCE(n_tup_del, 0)::bigint,
               COALESCE(n_tup_hot_upd, 0)::bigint
        FROM pg_stat_user_tables 
        WHERE schemaname LIKE 'tenant_%' OR schemaname = 'system'
        ORDER BY schemaname, relname;
        """
        
        try:
            result = subprocess.run(
                DatabaseConfig.psql_cmd() + ["-t", "-A", "-F", "|", "-c", query],
                capture_output=True, text=True, timeout=60
            )
            
            if result.returncode != 0:
                return []
            
            tables = []
            for line in result.stdout.strip().split('\n'):
                if not line.strip():
                    continue
                parts = line.split('|')
                if len(parts) >= 6:
                    schema = parts[0].strip()
                    table = parts[1].strip()
                    if schema and table:
                        n_tup_ins = int(parts[2].strip() or 0)
                        n_tup_upd = int(parts[3].strip() or 0)
                        n_tup_del = int(parts[4].strip() or 0)
                        n_tup_hot_upd = int(parts[5].strip() or 0)
                        rows = self._get_row_count(schema, table)
                        tables.append({
                            'schema': schema,
                            'table': table,
                            'rows': rows,
                            'change_vector': f"{n_tup_ins}:{n_tup_upd}:{n_tup_del}:{n_tup_hot_upd}",
                        })
            
            return tables
        except Exception:
            return []
    
    def _get_row_count(self, schema: str, table: str) -> int:
        """Get actual row count for a table."""
        try:
            result = subprocess.run(
                DatabaseConfig.psql_cmd() + [
                    "-t", "-A", "-c",
                    f'SELECT COUNT(*) FROM "{schema}"."{table}"'
                ],
                capture_output=True, text=True, timeout=60
            )
            if result.returncode == 0:
                return int(result.stdout.strip())
        except Exception:
            pass
        return 0
    
    def detect_changes(
        self, 
        tables: List[Dict[str, Any]], 
        prev_manifest: Optional[Dict[str, Any]], 
        force: bool = False
    ) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]], List[Dict[str, Any]]]:
        """Detect which tables need to be exported."""
        detector = BackupChangeDetector(self.backup_dir)
        return detector.detect_changes(tables, prev_manifest, force)
    
    def backup(self, force: bool = False, dry_run: bool = False) -> Tuple[bool, str]:
        """Execute full backup."""
        log_info("=" * 50)
        log_info("  PostgreSQL Smart Backup (Python)")
        log_info("=" * 50)
        
        if force:
            log_warning("  MODE: FORCE (export all)")
        if dry_run:
            log_warning("  MODE: DRY RUN (no export)")
        
        log_info("[1/6] Checking PostgreSQL...")
        if not self.start_postgres_if_needed():
            return False, "PostgreSQL is not available"
        log_success("PostgreSQL ready")
        
        # Version compatibility check
        log_info("[1.5/6] Checking version compatibility...")
        version_ok, version_msg = self.check_pg_version_compatibility()
        if not version_ok:
            return False, f"Version mismatch: {version_msg}"
        log_success("Version compatible")
        
        # Disk space check
        log_info("[1.6/6] Checking disk space...")
        space_ok, space_msg = self.check_disk_space()
        if not space_ok:
            return False, f"Insufficient space: {space_msg}"
        log_success("Disk space sufficient")
        
        log_info("[2/6] Collecting table metadata...")
        
        tables = self.get_all_tables()
        if not tables:
            return False, "No tables found!"
        
        schemas = set(t['schema'] for t in tables)
        log_success(f"Found {len(tables)} table(s) in {len(schemas)} schema(s)")
        
        log_info("[3/6] Detecting changes...")
        prev_manifest = self._manifest.load()
        to_export, unchanged, deleted = self.detect_changes(tables, prev_manifest, force)
        
        print()
        print("  +-----------------------------------+")
        print("  | CHANGE SUMMARY                    |")
        print("  +-----------------------------------+")
        print(f"  | To export:  {len(to_export):5} table(s)        |")
        print(f"  | Unchanged:  {len(unchanged):5} table(s)        |")
        print(f"  | Deleted:    {len(deleted):5} table(s)        |")
        print("  +-----------------------------------+")
        print()
        
        if dry_run:
            log_success("DRY RUN COMPLETE (no changes made)")
            return True, f"Would export {len(to_export)} tables"
        
        if not to_export and not deleted:
            log_success("NO CHANGES DETECTED - Database in sync")
            return True, "No changes to backup"
        
        # Import operations module
        from omni_build.database_backup_ops import DatabaseBackupOps
        ops = DatabaseBackupOps(self.backup_dir)
        
        log_info(f"[4/6] Exporting {len(to_export)} table(s)...")
        
        schemas_to_export = set(t['schema'] for t in to_export)
        for schema in schemas_to_export:
            ops.export_schema(schema)
        
        for i, item in enumerate(to_export, 1):
            print(f"  [{i}/{len(to_export)}] ", end="")
            ops.export_table(item['schema'], item['table'], item['rows'])
        
        # Backup globals (roles, permissions)
        ops.backup_globals()
        
        log_info("[5/6] Cleaning deleted tables...")
        ops.remove_deleted_tables(deleted)

        manifest_checksums: Dict[str, str] = dict(ops.table_checksums)
        manifest_rows: Dict[str, int] = dict(ops.table_row_counts)
        artifact_ops = DatabaseBackupOps(self.backup_dir)
        for table in tables:
            key = f"{table['schema']}.{table['table']}"
            if key not in manifest_rows:
                # Always use artifact row count as ground truth — previous
                # manifest rows can be stale if data changed between backups.
                artifact_rows = artifact_ops.calculate_artifact_row_count(table['schema'], table['table'])
                if artifact_rows >= 0:
                    manifest_rows[key] = artifact_rows
                else:
                    previous_rows = self._manifest._get_previous_rows(prev_manifest, table['schema'], table['table'])
                    if previous_rows >= 0:
                        manifest_rows[key] = previous_rows

            if key in manifest_checksums:
                continue

            previous_checksum = self._manifest._get_previous_checksum(prev_manifest, table['schema'], table['table'])
            if previous_checksum and self._manifest._is_sha256_checksum(previous_checksum):
                manifest_checksums[key] = previous_checksum
                continue

            artifact_checksum = artifact_ops.calculate_artifact_checksum(table['schema'], table['table'])
            if artifact_checksum:
                manifest_checksums[key] = artifact_checksum
        
        log_info("[6/6] Saving manifest...")
        if not self._manifest.save_manifest(
            tables,
            prev_manifest=prev_manifest,
            generated_checksums=manifest_checksums,
            generated_rows=manifest_rows,
        ):
            return False, "Failed to save manifest with verified checksums"
        
        print()
        log_success("=" * 50)
        log_success("  SMART BACKUP COMPLETE!")
        log_success("=" * 50)
        
        total_rows = sum(t['rows'] for t in tables)
        return len(ops.errors) == 0, f"Backup complete: {len(tables)} tables, {total_rows:,} rows"
    
    def get_backup_info(self) -> Optional[Dict[str, Any]]:
        """Get current backup manifest info."""
        return self._manifest.load()
    
    def print_backup_status(self) -> None:
        """Print backup status information."""
        self._manifest.print_status()
