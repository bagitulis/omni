"""
Database Backup module for Omni Build System.

SRP: This module handles database backup coordination.
Export operations are in database_backup_ops.py.
"""
import json
import subprocess
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from omni_build.backup_change_detector import BackupChangeDetector
from omni_build.config import Config
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
        self.manifest_file = self.backup_dir / "manifest.json"
    
    def check_postgres_ready(self) -> bool:
        """Check if PostgreSQL is running and ready."""
        try:
            result = subprocess.run(
                ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni", "-d", "omni_main"],
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
            subprocess.run(["docker", "start", "omni-postgres"], 
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
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni", 
                 "-d", "omni_main", "-t", "-A", "-F", "|", "-c", query],
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
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main", "-t", "-A", "-c", 
                 f'SELECT COUNT(*) FROM "{schema}"."{table}"'],
                capture_output=True, text=True, timeout=60
            )
            if result.returncode == 0:
                return int(result.stdout.strip())
        except Exception:
            pass
        return 0
    
    def load_previous_manifest(self) -> Optional[Dict[str, Any]]:
        """Load previous backup manifest if exists."""
        if not self.manifest_file.exists():
            return None
        
        try:
            with open(self.manifest_file, 'r', encoding='utf-8-sig') as f:
                return json.load(f)
        except Exception:
            return None
    
    def detect_changes(
        self, 
        tables: List[Dict[str, Any]], 
        prev_manifest: Optional[Dict[str, Any]], 
        force: bool = False
    ) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]], List[Dict[str, Any]]]:
        """Detect which tables need to be exported."""
        detector = BackupChangeDetector(self.backup_dir)
        return detector.detect_changes(tables, prev_manifest, force)
    
    def _get_previous_checksum(self, prev_manifest: Optional[Dict[str, Any]], schema: str, table: str) -> str:
        """Get checksum from previous manifest if available."""
        if not prev_manifest:
            return ""

        schema_data = prev_manifest.get('schemas', {}).get(schema, {})
        for prev_table in schema_data.get('tables', []):
            if prev_table.get('name') == table:
                checksum = prev_table.get('checksum', '')
                return checksum if isinstance(checksum, str) else ""

        return ""

    def _get_previous_rows(self, prev_manifest: Optional[Dict[str, Any]], schema: str, table: str) -> int:
        """Get row count from previous manifest if available."""
        if not prev_manifest:
            return -1

        schema_data = prev_manifest.get('schemas', {}).get(schema, {})
        for prev_table in schema_data.get('tables', []):
            if prev_table.get('name') == table:
                rows = prev_table.get('rows', -1)
                return rows if isinstance(rows, int) and rows >= 0 else -1

        return -1

    def _is_sha256_checksum(self, checksum: str) -> bool:
        """Validate checksum format as sha256 hex string."""
        if len(checksum) != 64:
            return False
        return all(c in "0123456789abcdef" for c in checksum.lower())

    def save_manifest(
        self,
        tables: List[Dict[str, Any]],
        prev_manifest: Optional[Dict[str, Any]] = None,
        generated_checksums: Optional[Dict[str, str]] = None,
        generated_rows: Optional[Dict[str, int]] = None,
    ) -> bool:
        """Save backup manifest."""
        schemas: Dict[str, Dict[str, List[Dict[str, Any]]]] = {}
        generated_checksums = generated_checksums or {}
        generated_rows = generated_rows or {}
        missing_checksums: List[str] = []
        missing_rows: List[str] = []
        
        for table in tables:
            schema = table['schema']
            if schema not in schemas:
                schemas[schema] = {'tables': []}

            table_name = table['table']
            table_key = f"{schema}.{table_name}"

            row_count = generated_rows.get(table_key, -1)
            if row_count < 0:
                row_count = self._get_previous_rows(prev_manifest, schema, table_name)
            if row_count < 0:
                source_rows = table.get('rows', 0)
                row_count = source_rows if isinstance(source_rows, int) and source_rows >= 0 else 0
                if row_count > 0:
                    missing_rows.append(table_key)

            checksum = generated_checksums.get(f"{schema}.{table_name}", "")
            if not checksum:
                checksum = self._get_previous_checksum(prev_manifest, schema, table_name)

            if not checksum:
                if row_count > 0:
                    missing_checksums.append(f"{schema}.{table_name}")
                checksum = "zero-rows"

            schemas[schema]['tables'].append({
                'name': table_name,
                'rows': row_count,
                'checksum': checksum,
                'change_vector': table.get('change_vector', ''),
            })

        if missing_rows:
            log_warning(
                "Manifest rows fallback to analyzed counts for table(s): "
                + ", ".join(missing_rows[:5])
            )

        if missing_checksums:
            log_error(
                "Cannot save manifest: missing checksum for non-empty table(s): "
                + ", ".join(missing_checksums[:5])
            )
            return False
        
        manifest = {
            'version': 4,
            'exported_at': datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            'schemas': schemas
        }
        
        try:
            with open(self.manifest_file, 'w', encoding='utf-8') as f:
                json.dump(manifest, f, indent=2)
            return True
        except Exception:
            return False
    
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
        
        log_info("[2/6] Collecting table metadata...")
        
        tables = self.get_all_tables()
        if not tables:
            return False, "No tables found!"
        
        schemas = set(t['schema'] for t in tables)
        log_success(f"Found {len(tables)} table(s) in {len(schemas)} schema(s)")
        
        log_info("[3/6] Detecting changes...")
        prev_manifest = self.load_previous_manifest()
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
                    previous_rows = self._get_previous_rows(prev_manifest, table['schema'], table['table'])
                    if previous_rows >= 0:
                        manifest_rows[key] = previous_rows

            if key in manifest_checksums:
                continue

            previous_checksum = self._get_previous_checksum(prev_manifest, table['schema'], table['table'])
            if previous_checksum and self._is_sha256_checksum(previous_checksum):
                manifest_checksums[key] = previous_checksum
                continue

            artifact_checksum = artifact_ops.calculate_artifact_checksum(table['schema'], table['table'])
            if artifact_checksum:
                manifest_checksums[key] = artifact_checksum
        
        log_info("[6/6] Saving manifest...")
        if not self.save_manifest(
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
        return self.load_previous_manifest()
    
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
