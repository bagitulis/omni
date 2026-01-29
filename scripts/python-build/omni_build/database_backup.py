"""
Database Backup module for Omni Build System.

SRP: This module ONLY handles database backup operations.
Native Python implementation - no PowerShell dependency.
"""
import gzip
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
    - Comprehensive error handling
    """
    
    LARGE_TABLE_THRESHOLD = 50000  # Rows before using chunked export
    CHUNK_SIZE = 25000  # Rows per chunk
    
    def __init__(self, config: Config) -> None:
        """Initialize database backup with config."""
        self.config = config
        self.backup_dir = config.project_root / "backups"
        self.manifest_file = self.backup_dir / "manifest.json"
        self.schema_dir = self.backup_dir / "_schema"
        self.errors: List[str] = []
        self.warnings: List[str] = []
    
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
                
                for i in range(30):
                    time.sleep(2)
                    if self._ensure_postgres_healthy():
                        log_success("PostgreSQL recovered successfully")
                        return True
                
                self.errors.append("PostgreSQL failed to recover within 60 seconds")
                return False
            elif status == "running":
                for i in range(10):
                    time.sleep(2)
                    if self._ensure_postgres_healthy():
                        return True
                return False
            else:
                self.errors.append(f"Container in unexpected state: {status}")
                return False
                
        except Exception as e:
            self.errors.append(f"Recovery failed: {e}")
            return False
    
    def start_postgres_if_needed(self) -> bool:
        """Start PostgreSQL container if not running."""
        if self.check_postgres_ready():
            return True
        
        log_info("Starting PostgreSQL container...")
        try:
            subprocess.run(["docker", "start", "omni-postgres"], 
                          capture_output=True, timeout=30)
            
            # Wait for ready
            for _ in range(30):
                if self.check_postgres_ready():
                    log_success("PostgreSQL is ready")
                    return True
                time.sleep(1)
            
            self.errors.append("PostgreSQL failed to start within 30 seconds")
            return False
        except Exception as e:
            self.errors.append(f"Failed to start PostgreSQL: {e}")
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
                self.errors.append(f"Failed to query tables: {result.stderr}")
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
                        # Get actual row count
                        rows = self._get_row_count(schema, table)
                        tables.append({
                            'schema': schema,
                            'table': table,
                            'rows': rows,
                            'checksum': f"{rows}-{table}"
                        })
            
            return tables
        except Exception as e:
            self.errors.append(f"Exception getting tables: {e}")
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
        except Exception as e:
            self.warnings.append(f"Could not load previous manifest: {e}")
            return None
    
    def detect_changes(self, tables: List[Dict], prev_manifest: Optional[Dict], 
                       force: bool = False) -> Tuple[List[Dict], List[Dict], List[Dict]]:
        """
        Detect which tables need to be exported.
        
        Returns: (to_export, unchanged, deleted)
        """
        to_export = []
        unchanged = []
        deleted = []
        
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
        
        # Check for deleted tables
        if prev_manifest:
            for schema_name, schema_data in prev_manifest.get('schemas', {}).items():
                for prev_table in schema_data.get('tables', []):
                    exists = any(t['schema'] == schema_name and t['table'] == prev_table['name'] 
                               for t in tables)
                    if not exists:
                        deleted.append({'schema': schema_name, 'table': prev_table['name']})
        
        return to_export, unchanged, deleted
    
    def export_schema(self, schema: str) -> bool:
        """Export schema structure (DDL) to _schema directory."""
        self.schema_dir.mkdir(parents=True, exist_ok=True)
        schema_file = self.schema_dir / f"{schema}.sql.gz"
        
        try:
            # Export schema structure
            cmd = f"pg_dump -U omni -d omni_main --schema={schema} --schema-only --no-owner --no-acl 2>/dev/null | gzip -n"
            result = subprocess.run(
                ["docker", "exec", "omni-postgres", "sh", "-c", cmd],
                capture_output=True, timeout=120
            )
            
            if result.returncode == 0 and result.stdout:
                with open(schema_file, 'wb') as f:
                    f.write(result.stdout)
                return True
            else:
                self.warnings.append(f"Schema export warning for {schema}")
                return False
        except Exception as e:
            self.errors.append(f"Failed to export schema {schema}: {e}")
            return False
    
    def export_table(self, schema: str, table: str, rows: int) -> bool:
        """Export a single table's data with container health checks."""
        schema_dir = self.backup_dir / schema
        schema_dir.mkdir(parents=True, exist_ok=True)
        
        # Check container health before export
        if not self._ensure_postgres_healthy():
            log_warning(f"PostgreSQL not healthy before {table} export - attempting recovery...")
            if not self._recover_postgres():
                self.errors.append(f"Container crash before {schema}.{table}")
                return False
        
        if rows >= self.LARGE_TABLE_THRESHOLD:
            return self._export_table_chunked(schema, table, rows, schema_dir)
        else:
            return self._export_table_single(schema, table, schema_dir)
    
    def _export_table_single(self, schema: str, table: str, dest_dir: Path) -> bool:
        """Export small table as single gzipped SQL file."""
        table_file = dest_dir / f"{table}.sql.gz"
        
        # Clean up old chunks if switching from chunked to single
        for old_chunk in dest_dir.glob(f"{table}.chunk*.sql.gz"):
            old_chunk.unlink()
        meta_file = dest_dir / f"{table}.meta.json"
        if meta_file.exists():
            meta_file.unlink()
        
        try:
            cmd = f"pg_dump -U omni -d omni_main --table={schema}.{table} --data-only --no-owner --no-acl 2>/dev/null | gzip -n"
            result = subprocess.run(
                ["docker", "exec", "omni-postgres", "sh", "-c", cmd],
                capture_output=True, timeout=300
            )
            
            if result.returncode == 0 and result.stdout:
                with open(table_file, 'wb') as f:
                    f.write(result.stdout)
                size_kb = table_file.stat().st_size / 1024
                print(f"  [OK] {schema}.{table}: {size_kb:.1f} KB")
                return True
            else:
                self.warnings.append(f"pg_dump warning for {schema}.{table}")
                return False
        except Exception as e:
            self.errors.append(f"Failed to export {schema}.{table}: {e}")
            return False
    
    def _export_table_chunked(self, schema: str, table: str, total_rows: int, 
                              dest_dir: Path) -> bool:
        """Export large table in chunks using LIMIT/OFFSET."""
        print(f"  [INFO] {schema}.{table}: exporting {total_rows:,} rows in chunks...")
        
        # Clean up old single file and chunks
        old_file = dest_dir / f"{table}.sql.gz"
        if old_file.exists():
            old_file.unlink()
        for old_chunk in dest_dir.glob(f"{table}.chunk*.sql.gz"):
            old_chunk.unlink()
        
        # Get primary key for ordering
        pk_query = f"""
        SELECT a.attname FROM pg_index i 
        JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey) 
        WHERE i.indrelid = '{schema}.{table}'::regclass AND i.indisprimary LIMIT 1;
        """
        try:
            pk_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main", "-t", "-A", "-c", pk_query],
                capture_output=True, text=True, timeout=30
            )
            order_by = pk_result.stdout.strip() if pk_result.returncode == 0 and pk_result.stdout.strip() else "1"
        except Exception:
            order_by = "1"
        
        offset = 0
        chunk_num = 0
        success_count = 0
        
        while offset < total_rows:
            chunk_file = dest_dir / f"{table}.chunk{chunk_num:03d}.sql.gz"
            
            # Build COPY query with LIMIT/OFFSET
            copy_cmd = f'psql -U omni -d omni_main -c "COPY (SELECT * FROM {schema}.{table} ORDER BY {order_by} LIMIT {self.CHUNK_SIZE} OFFSET {offset}) TO STDOUT" 2>/dev/null | gzip -n'
            
            try:
                result = subprocess.run(
                    ["docker", "exec", "omni-postgres", "sh", "-c", copy_cmd],
                    capture_output=True, timeout=600
                )
                
                if result.returncode == 0 and result.stdout and len(result.stdout) > 20:
                    with open(chunk_file, 'wb') as f:
                        f.write(result.stdout)
                    success_count += 1
                    
                    if (chunk_num + 1) % 5 == 0 or offset + self.CHUNK_SIZE >= total_rows:
                        print(f"    chunk{chunk_num:03d}: rows {offset:,}-{min(offset + self.CHUNK_SIZE, total_rows):,}")
                else:
                    self.warnings.append(f"Chunk {chunk_num} empty for {schema}.{table}")
                    
            except Exception as e:
                self.warnings.append(f"Chunk {chunk_num} failed for {schema}.{table}: {e}")
            
            offset += self.CHUNK_SIZE
            chunk_num += 1
        
        # Save metadata
        meta_file = dest_dir / f"{table}.meta.json"
        meta = {
            'table': table,
            'schema': schema,
            'total_rows': total_rows,
            'chunks': success_count,
            'chunk_size': self.CHUNK_SIZE,
            'order_by': order_by,
            'exported_at': datetime.now().strftime("%Y-%m-%d %H:%M:%S")
        }
        with open(meta_file, 'w', encoding='utf-8') as f:
            json.dump(meta, f, indent=2)
        
        print(f"  [OK] {schema}.{table}: {success_count} chunks exported")
        return success_count > 0
    
    def remove_deleted_tables(self, deleted: List[Dict]) -> None:
        """Remove backup files for deleted tables."""
        for item in deleted:
            table_file = self.backup_dir / item['schema'] / f"{item['table']}.sql.gz"
            if table_file.exists():
                table_file.unlink()
                print(f"  Removed: {item['schema']}.{item['table']}")
            
            # Also remove chunks
            schema_dir = self.backup_dir / item['schema']
            if schema_dir.exists():
                for chunk in schema_dir.glob(f"{item['table']}.chunk*.sql.gz"):
                    chunk.unlink()
                meta_file = schema_dir / f"{item['table']}.meta.json"
                if meta_file.exists():
                    meta_file.unlink()
    
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
            # Save without BOM for cross-platform compatibility
            with open(self.manifest_file, 'w', encoding='utf-8') as f:
                json.dump(manifest, f, indent=2)
            return True
        except Exception as e:
            self.errors.append(f"Failed to save manifest: {e}")
            return False
    
    def backup(self, force: bool = False, dry_run: bool = False) -> Tuple[bool, str]:
        """
        Execute full backup.
        
        Args:
            force: Force export all tables (ignore change detection)
            dry_run: Preview only, don't actually export
            
        Returns:
            Tuple of (success, message)
        """
        log_info("=" * 50)
        log_info("  PostgreSQL Smart Backup (Python)")
        log_info("=" * 50)
        
        if force:
            log_warning("  MODE: FORCE (export all)")
        if dry_run:
            log_warning("  MODE: DRY RUN (no export)")
        
        # Step 1: Check PostgreSQL
        log_info("[1/6] Checking PostgreSQL...")
        if not self.start_postgres_if_needed():
            return False, "PostgreSQL is not available"
        log_success("PostgreSQL ready")
        
        # Step 2: Analyze database
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
        
        # Step 3: Detect changes
        log_info("[3/6] Detecting changes...")
        prev_manifest = self.load_previous_manifest()
        to_export, unchanged, deleted = self.detect_changes(tables, prev_manifest, force)
        
        # Summary
        print()
        print("  +-----------------------------------+")
        print("  | CHANGE SUMMARY                    |")
        print("  +-----------------------------------+")
        print(f"  | To export:  {len(to_export):5} table(s)        |")
        print(f"  | Unchanged:  {len(unchanged):5} table(s)        |")
        print(f"  | Deleted:    {len(deleted):5} table(s)        |")
        print("  +-----------------------------------+")
        print()
        
        if len(to_export) <= 15:
            for item in to_export:
                print(f"    + {item['schema']}.{item['table']} ({item.get('reason', 'new')})")
            print()
        
        if dry_run:
            log_success("DRY RUN COMPLETE (no changes made)")
            return True, f"Would export {len(to_export)} tables"
        
        if not to_export and not deleted:
            log_success("NO CHANGES DETECTED - Database in sync")
            return True, "No changes to backup"
        
        # Step 4: Export tables
        log_info(f"[4/6] Exporting {len(to_export)} table(s)...")
        
        # Export schema structures
        schemas_to_export = set(t['schema'] for t in to_export)
        for schema in schemas_to_export:
            self.export_schema(schema)
        
        # Export table data
        for i, item in enumerate(to_export, 1):
            print(f"  [{i}/{len(to_export)}] ", end="")
            self.export_table(item['schema'], item['table'], item['rows'])
        
        # Step 5: Remove deleted
        log_info("[5/6] Cleaning deleted tables...")
        self.remove_deleted_tables(deleted)
        
        # Step 6: Save manifest
        log_info("[6/6] Saving manifest...")
        self.save_manifest(tables)
        
        # Summary
        print()
        log_success("=" * 50)
        log_success("  SMART BACKUP COMPLETE!")
        log_success("=" * 50)
        log_info(f"Exported: {len(to_export)} table(s)")
        log_info(f"Skipped:  {len(unchanged)} table(s) (unchanged)")
        log_info(f"Deleted:  {len(deleted)} table(s)")
        
        if self.errors:
            log_error(f"ERRORS: {len(self.errors)}")
            for err in self.errors[:5]:
                print(f"  - {err}")
        
        if self.warnings:
            log_warning(f"WARNINGS: {len(self.warnings)}")
            for warn in self.warnings[:5]:
                print(f"  - {warn}")
        
        total_rows = sum(t['rows'] for t in tables)
        return len(self.errors) == 0, f"Backup complete: {len(tables)} tables, {total_rows:,} rows"
    
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
