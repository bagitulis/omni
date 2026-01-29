"""
Database Backup Operations module for Omni Build System.

SRP: This module handles table and schema export operations for backup.
"""
import gzip
import json
import subprocess
import time
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple

from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseBackupOps:
    """Handles database export operations for backup."""
    
    LARGE_TABLE_THRESHOLD = 50000
    CHUNK_SIZE = 25000
    
    def __init__(self, backup_dir: Path) -> None:
        self.backup_dir = backup_dir
        self.schema_dir = backup_dir / "_schema"
        self.errors: List[str] = []
        self.warnings: List[str] = []
    
    def ensure_postgres_healthy(self) -> bool:
        """Quick check if PostgreSQL is responsive."""
        try:
            result = subprocess.run(
                ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni"],
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
                ["docker", "inspect", "--format", "{{.State.Status}}", "omni-postgres"],
                capture_output=True, text=True, timeout=10
            )
            status = result.stdout.strip()
            
            if status == "exited":
                log_info("Container exited - attempting restart...")
                subprocess.run(["docker", "start", "omni-postgres"], 
                              capture_output=True, timeout=30)
                
                for _ in range(30):
                    time.sleep(2)
                    if self.ensure_postgres_healthy():
                        log_success("PostgreSQL recovered successfully")
                        return True
                
                self.errors.append("PostgreSQL failed to recover within 60 seconds")
                return False
            elif status == "running":
                for _ in range(10):
                    time.sleep(2)
                    if self.ensure_postgres_healthy():
                        return True
                return False
            else:
                self.errors.append(f"Container in unexpected state: {status}")
                return False
                
        except Exception as e:
            self.errors.append(f"Recovery failed: {e}")
            return False
    
    def export_schema(self, schema: str) -> bool:
        """Export schema structure (DDL) to _schema directory."""
        self.schema_dir.mkdir(parents=True, exist_ok=True)
        schema_file = self.schema_dir / f"{schema}.sql.gz"
        
        try:
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
        
        if not self.ensure_postgres_healthy():
            log_warning(f"PostgreSQL not healthy before {table} export")
            if not self.recover_postgres():
                self.errors.append(f"Container crash before {schema}.{table}")
                return False
        
        if rows >= self.LARGE_TABLE_THRESHOLD:
            return self._export_table_chunked(schema, table, rows, schema_dir)
        else:
            return self._export_table_single(schema, table, schema_dir)
    
    def _export_table_single(self, schema: str, table: str, dest_dir: Path) -> bool:
        """Export small table as single gzipped SQL file."""
        table_file = dest_dir / f"{table}.sql.gz"
        
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
    
    def _export_table_chunked(
        self, schema: str, table: str, total_rows: int, dest_dir: Path
    ) -> bool:
        """Export large table in chunks using LIMIT/OFFSET."""
        print(f"  [INFO] {schema}.{table}: exporting {total_rows:,} rows in chunks...")
        
        old_file = dest_dir / f"{table}.sql.gz"
        if old_file.exists():
            old_file.unlink()
        for old_chunk in dest_dir.glob(f"{table}.chunk*.sql.gz"):
            old_chunk.unlink()
        
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
            
            schema_dir = self.backup_dir / item['schema']
            if schema_dir.exists():
                for chunk in schema_dir.glob(f"{item['table']}.chunk*.sql.gz"):
                    chunk.unlink()
                meta_file = schema_dir / f"{item['table']}.meta.json"
                if meta_file.exists():
                    meta_file.unlink()
