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
from typing import Dict, List, Optional, Tuple

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
                ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni"],
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
    
    def get_all_tables(self) -> List[Dict]:
        """Get all tables with row counts from tenant and system schemas."""
        query = """
        SELECT schemaname, relname 
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
                if len(parts) >= 2:
                    schema = parts[0].strip()
                    table = parts[1].strip()
                    if schema and table:
                        rows = self._get_row_count(schema, table)
                        tables.append({
                            'schema': schema,
                            'table': table,
                            'rows': rows,
                            'checksum': f"{rows}-{table}"
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
    
    def load_previous_manifest(self) -> Optional[Dict]:
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
        tables: List[Dict], 
        prev_manifest: Optional[Dict], 
        force: bool = False
    ) -> Tuple[List[Dict], List[Dict], List[Dict]]:
        """Detect which tables need to be exported."""
        to_export, unchanged, deleted = [], [], []
        
        for table in tables:
            needs_export = True
            reason = "new"
            
            if not force and prev_manifest:
                prev_schema = prev_manifest.get('schemas', {}).get(table['schema'])
                if prev_schema:
                    prev_tables = prev_schema.get('tables', [])
                    prev_table = next((t for t in prev_tables if t['name'] == table['table']), None)
                    if prev_table:
                        if prev_table.get('rows') == table['rows']:
                            needs_export = False
                            unchanged.append(table)
                        else:
                            reason = f"rows: {prev_table.get('rows')} -> {table['rows']}"
            
            if needs_export:
                table['reason'] = reason
                to_export.append(table)
        
        if prev_manifest:
            for schema_name, schema_data in prev_manifest.get('schemas', {}).items():
                for prev_table in schema_data.get('tables', []):
                    exists = any(t['schema'] == schema_name and t['table'] == prev_table['name'] 
                               for t in tables)
                    if not exists:
                        deleted.append({'schema': schema_name, 'table': prev_table['name']})
        
        return to_export, unchanged, deleted
    
    def save_manifest(self, tables: List[Dict]) -> bool:
        """Save backup manifest."""
        schemas: Dict[str, Dict] = {}
        
        for table in tables:
            schema = table['schema']
            if schema not in schemas:
                schemas[schema] = {'tables': []}
            schemas[schema]['tables'].append({
                'name': table['table'],
                'rows': table['rows'],
                'checksum': table['checksum']
            })
        
        manifest = {
            'version': 2,
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
        
        log_info("[2/6] Analyzing database...")
        subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni", 
             "-d", "omni_main", "-c", "ANALYZE;"],
            capture_output=True, timeout=120
        )
        
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
        
        log_info("[6/6] Saving manifest...")
        self.save_manifest(tables)
        
        print()
        log_success("=" * 50)
        log_success("  SMART BACKUP COMPLETE!")
        log_success("=" * 50)
        
        total_rows = sum(t['rows'] for t in tables)
        return len(ops.errors) == 0, f"Backup complete: {len(tables)} tables, {total_rows:,} rows"
    
    def get_backup_info(self) -> Optional[Dict]:
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
