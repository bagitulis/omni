"""
Database Backup Operations module for Omni Build System.

SRP: This module handles table and schema export operations for backup.
"""
import gzip
import hashlib
import json
import subprocess
import time
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Tuple

from omni_build.db_config import DatabaseConfig
from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseBackupOps:
    """Handles database export operations for backup."""
    
    LARGE_TABLE_THRESHOLD = DatabaseConfig.LARGE_TABLE_THRESHOLD
    CHUNK_SIZE = DatabaseConfig.CHUNK_SIZE
    
    def __init__(self, backup_dir: Path) -> None:
        self.backup_dir = backup_dir
        self.schema_dir = backup_dir / "_schema"
        self.errors: List[str] = []
        self.warnings: List[str] = []
        self.table_checksums: Dict[str, str] = {}
        self.table_row_counts: Dict[str, int] = {}

    def _update_hash_from_file(self, file_path: Path, hasher: "hashlib._Hash") -> None:
        """Update hasher from file bytes in chunks."""
        with open(file_path, 'rb') as f:
            while True:
                chunk = f.read(1024 * 1024)
                if not chunk:
                    break
                hasher.update(chunk)

    def _calculate_table_checksum(self, schema: str, table: str, schema_dir: Path) -> str:
        """Calculate deterministic checksum for backup artifact(s)."""
        hasher = hashlib.sha256()
        table_file = schema_dir / f"{table}.sql.gz"
        chunk_files = sorted(schema_dir.glob(f"{table}.chunk*.sql.gz"))

        if table_file.exists():
            hasher.update(f"single:{schema}.{table}".encode('utf-8'))
            self._update_hash_from_file(table_file, hasher)
            return hasher.hexdigest()

        if chunk_files:
            hasher.update(f"chunked:{schema}.{table}".encode('utf-8'))
            for chunk_file in chunk_files:
                hasher.update(chunk_file.name.encode('utf-8'))
                self._update_hash_from_file(chunk_file, hasher)
            return hasher.hexdigest()

        return ""

    def calculate_artifact_checksum(self, schema: str, table: str) -> str:
        """Calculate checksum for existing backup artifact(s)."""
        schema_dir = self.backup_dir / schema
        return self._calculate_table_checksum(schema, table, schema_dir)

    def _count_rows_in_single_dump(self, table_file: Path) -> int:
        """Count rows stored in a pg_dump data-only .sql.gz file."""
        copy_rows = 0
        insert_rows = 0
        in_copy = False

        with gzip.open(table_file, 'rt', encoding='utf-8', errors='replace') as f:
            for line in f:
                if line.startswith('COPY '):
                    in_copy = True
                    continue

                if in_copy:
                    if line.strip() == "\\.":
                        in_copy = False
                    else:
                        copy_rows += 1
                    continue

                if line.startswith('INSERT INTO '):
                    insert_rows += 1

        return copy_rows + insert_rows

    def _count_rows_in_chunk_dump(self, chunk_files: List[Path]) -> int:
        """Count rows stored in chunked .sql.gz files."""
        total_rows = 0
        for chunk_file in chunk_files:
            with gzip.open(chunk_file, 'rt', encoding='utf-8', errors='replace') as f:
                for line in f:
                    if line.strip():
                        total_rows += 1
        return total_rows

    def _get_primary_key_column(self, schema: str, table: str) -> str:
        """Resolve primary key column for deterministic chunk ordering."""
        pk_query = f"""
        SELECT a.attname FROM pg_index i
        JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
        WHERE i.indrelid = '{schema}.{table}'::regclass AND i.indisprimary LIMIT 1;
        """

        try:
            pk_result = subprocess.run(
                DatabaseConfig.psql_cmd() + ["-t", "-A", "-c", pk_query],
                capture_output=True,
                text=True,
                timeout=30,
            )
            if pk_result.returncode == 0 and pk_result.stdout.strip():
                return pk_result.stdout.strip()
        except Exception:
            pass

        return "1"

    def _calculate_live_single_checksum(self, schema: str, table: str) -> str:
        """Calculate checksum from live PostgreSQL dump stream (single-file strategy)."""
        _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
        cmd = (
            f"pg_dump -U {_u} -d {_d} --table={schema}.{table} "
            "--data-only --no-owner --no-acl 2>/dev/null | gzip -n"
        )
        result = subprocess.run(
            DatabaseConfig.docker_exec_prefix() + ["sh", "-c", cmd],
            capture_output=True,
            timeout=300,
        )

        if result.returncode != 0 or not result.stdout:
            return ""

        hasher = hashlib.sha256()
        hasher.update(f"single:{schema}.{table}".encode("utf-8"))
        hasher.update(result.stdout)
        return hasher.hexdigest()

    def _calculate_live_chunked_checksum(self, schema: str, table: str, total_rows: int) -> str:
        """Calculate checksum from live PostgreSQL stream (chunked strategy)."""
        if total_rows <= 0:
            return ""

        order_by = self._get_primary_key_column(schema, table)
        hasher = hashlib.sha256()
        hasher.update(f"chunked:{schema}.{table}".encode("utf-8"))

        offset = 0
        chunk_num = 0
        while offset < total_rows:
            chunk_name = f"{table}.chunk{chunk_num:03d}.sql.gz"
            _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
            copy_cmd = (
                f'psql -U {_u} -d {_d} -c "COPY (SELECT * FROM {schema}.{table} '
                f'ORDER BY {order_by} LIMIT {self.CHUNK_SIZE} OFFSET {offset}) TO STDOUT" '
                '2>/dev/null | gzip -n'
            )
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + ["sh", "-c", copy_cmd],
                capture_output=True,
                timeout=600,
            )

            if result.returncode != 0 or not result.stdout:
                return ""

            hasher.update(chunk_name.encode("utf-8"))
            hasher.update(result.stdout)
            offset += self.CHUNK_SIZE
            chunk_num += 1

        return hasher.hexdigest()

    def calculate_live_table_checksum(self, schema: str, table: str, rows: int) -> str:
        """Calculate checksum directly from live PostgreSQL data without writing backup files."""
        try:
            if rows >= self.LARGE_TABLE_THRESHOLD:
                return self._calculate_live_chunked_checksum(schema, table, rows)
            return self._calculate_live_single_checksum(schema, table)
        except Exception:
            return ""

    def calculate_artifact_row_count(self, schema: str, table: str) -> int:
        """Calculate row count from existing backup artifact(s)."""
        schema_dir = self.backup_dir / schema
        table_file = schema_dir / f"{table}.sql.gz"
        chunk_files = sorted(schema_dir.glob(f"{table}.chunk*.sql.gz"))

        if table_file.exists():
            return self._count_rows_in_single_dump(table_file)

        if chunk_files:
            return self._count_rows_in_chunk_dump(chunk_files)

        return -1
    
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
                ["docker", "inspect", "--format", "{{.State.Status}}", DatabaseConfig.CONTAINER_NAME],
                capture_output=True, text=True, timeout=10
            )
            status = result.stdout.strip()
            
            if status == "exited":
                log_info("Container exited - attempting restart...")
                subprocess.run(["docker", "start", DatabaseConfig.CONTAINER_NAME], 
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
        schema_file_plain = self.schema_dir / f"{schema}.sql"
        
        try:
            _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
            _prefix = DatabaseConfig.docker_exec_prefix()
            # Export compressed (for restore)
            cmd_gz = f"pg_dump -U {_u} -d {_d} --schema={schema} --schema-only --no-owner --no-acl 2>/dev/null | gzip -n"
            result_gz = subprocess.run(
                _prefix + ["sh", "-c", cmd_gz],
                capture_output=True, timeout=120
            )
            
            # Export plain text (for git diff)
            cmd_plain = f"pg_dump -U {_u} -d {_d} --schema={schema} --schema-only --no-owner --no-acl 2>/dev/null"
            result_plain = subprocess.run(
                _prefix + ["sh", "-c", cmd_plain],
                capture_output=True, timeout=120
            )
            
            if result_gz.returncode == 0 and result_gz.stdout:
                with open(schema_file, 'wb') as f:
                    f.write(result_gz.stdout)
            
            if result_plain.returncode == 0 and result_plain.stdout:
                with open(schema_file_plain, 'wb') as f:
                    f.write(result_plain.stdout)
            
            if result_gz.returncode == 0 and result_gz.stdout:
                return True
            else:
                self.warnings.append(f"Schema export warning for {schema}")
                return False
        except Exception as e:
            self.errors.append(f"Failed to export schema {schema}: {e}")
            return False
    
    def log_backup_metrics(self, tables_count: int, total_rows: int, 
                           duration_seconds: float, exported_count: int) -> None:
        """Append backup metrics to metrics log file."""
        metrics_file = self.backup_dir / "metrics.jsonl"
        metrics = {
            "timestamp": datetime.now().isoformat(),
            "tables_total": tables_count,
            "tables_exported": exported_count,
            "total_rows": total_rows,
            "duration_seconds": round(duration_seconds, 1),
            "errors": len(self.errors),
            "warnings": len(self.warnings),
        }
        try:
            with open(metrics_file, 'a', encoding='utf-8') as f:
                f.write(json.dumps(metrics) + '\n')
        except Exception:
            pass  # Metrics logging is best-effort
    
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
            success = self._export_table_chunked(schema, table, rows, schema_dir)
        else:
            success = self._export_table_single(schema, table, schema_dir)

        if not success:
            self.table_checksums.pop(f"{schema}.{table}", None)
            return False

        checksum = self._calculate_table_checksum(schema, table, schema_dir)
        if not checksum:
            self.warnings.append(f"Checksum skipped: backup artifact missing for {schema}.{table}")
            self.table_checksums.pop(f"{schema}.{table}", None)
            return False

        row_count = self.calculate_artifact_row_count(schema, table)
        if row_count < 0:
            self.warnings.append(f"Row count skipped: backup artifact missing for {schema}.{table}")
            self.table_checksums.pop(f"{schema}.{table}", None)
            return False

        self.table_checksums[f"{schema}.{table}"] = checksum
        self.table_row_counts[f"{schema}.{table}"] = row_count
        return True
    
    def _export_table_single(self, schema: str, table: str, dest_dir: Path) -> bool:
        """Export small table as single gzipped SQL file."""
        table_file = dest_dir / f"{table}.sql.gz"
        
        for old_chunk in dest_dir.glob(f"{table}.chunk*.sql.gz"):
            old_chunk.unlink()
        meta_file = dest_dir / f"{table}.meta.json"
        if meta_file.exists():
            meta_file.unlink()
        
        try:
            _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
            cmd = f"pg_dump -U {_u} -d {_d} --table={schema}.{table} --data-only --no-owner --no-acl 2>/dev/null | gzip -n"
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + ["sh", "-c", cmd],
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
        
        order_by = self._get_primary_key_column(schema, table)
        
        offset = 0
        chunk_num = 0
        success_count = 0
        
        while offset < total_rows:
            chunk_file = dest_dir / f"{table}.chunk{chunk_num:03d}.sql.gz"
            
            _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
            copy_cmd = f'psql -U {_u} -d {_d} -c "COPY (SELECT * FROM {schema}.{table} ORDER BY {order_by} LIMIT {self.CHUNK_SIZE} OFFSET {offset}) TO STDOUT" 2>/dev/null | gzip -n'
            
            try:
                result = subprocess.run(
                    DatabaseConfig.docker_exec_prefix() + ["sh", "-c", copy_cmd],
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
    
    def remove_deleted_tables(self, deleted: List[Dict[str, str]]) -> None:
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
    
    def backup_globals(self) -> bool:
        """Backup global objects (roles, tablespaces) - not included in pg_dump."""
        try:
            globals_file = self.schema_dir / "_globals.sql.gz"
            self.schema_dir.mkdir(parents=True, exist_ok=True)
            
            cmd = f"pg_dumpall -U {DatabaseConfig.USER} --globals-only --no-role-passwords 2>/dev/null | gzip -n"
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + ["sh", "-c", cmd],
                capture_output=True, timeout=120
            )
            
            if result.returncode == 0 and result.stdout:
                with open(globals_file, 'wb') as f:
                    f.write(result.stdout)
                return True
            return False
        except Exception as e:
            self.warnings.append(f"Globals backup failed: {e}")
            return False
