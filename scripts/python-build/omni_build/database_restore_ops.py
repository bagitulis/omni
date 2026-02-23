"""
Database Restore Operations module for Omni Build System.

SRP: This module handles table and schema restore operations.
"""
import gzip
import subprocess
import time
from pathlib import Path
from typing import List, Tuple

from omni_build.config import Config
from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseRestoreOps:
    """Handles database restore operations for tables and schemas."""
    
    def __init__(self, config: Config, pg_checker) -> None:
        self.config = config
        self._pg_checker = pg_checker
    
    def recreate_schemas_from_backup(self, data_dir: Path) -> bool:
        """Recreate schemas from backup _schema dumps."""
        schema_dir = data_dir / "_schema"
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
                if not self._pg_checker.ensure_healthy():
                    log_error("PostgreSQL crashed - attempting recovery...")
                    if not self._pg_checker.recover():
                        return False
                
                drop_cmd = [
                    "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                    "-d", "omni_main", "-c", f"DROP SCHEMA IF EXISTS {schema_name} CASCADE;"
                ]
                subprocess.run(drop_cmd, capture_output=True, text=True, timeout=60)
                
                if not self._pg_checker.ensure_healthy():
                    log_error(f"PostgreSQL crashed during DROP SCHEMA {schema_name}")
                    if not self._pg_checker.recover():
                        return False
                    continue
                
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
                log_error(f"  Timeout recreating {schema_name}")
                if not self._pg_checker.recover():
                    return False
            except Exception as e:
                log_warning(f"  Failed to recreate {schema_name}: {e}")
        
        return True
    
    def restore_tables(self, manifest: dict, data_dir: Path) -> Tuple[bool, str]:
        """Restore tables from backup files."""
        schemas = manifest.get('schemas', {})
        restored_tables = 0
        restored_rows = 0
        errors: List[str] = []
        
        tables_to_restore: List[Tuple] = []
        total_expected_rows = 0
        
        for schema_name, schema_data in schemas.items():
            schema_dir = data_dir / schema_name
            if not schema_dir.exists():
                log_warning(f"Schema directory not found: {schema_dir}")
                continue
            
            for table_info in schema_data.get('tables', []):
                table_name = table_info.get('name')
                expected_rows = table_info.get('rows', 0)
                if expected_rows > 0:
                    tables_to_restore.append((schema_name, table_name, expected_rows, schema_dir))
                    total_expected_rows += expected_rows
        
        log_info(f"Clearing {len(tables_to_restore)} tables...")
        for schema_name, table_name, _, _ in tables_to_restore:
            clear_cmd = [
                "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                "-d", "omni_main", "-c", f"TRUNCATE {schema_name}.{table_name};"
            ]
            subprocess.run(clear_cmd, capture_output=True, timeout=30)
        
        for schema_name, table_name, expected_rows, schema_dir in tables_to_restore:
            if not self._pg_checker.ensure_healthy():
                log_warning(f"PostgreSQL not healthy before {table_name}")
                if not self._pg_checker.recover():
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
                    schema_name, table_name, sql_file, skip_truncate=True,
                    expected_rows=expected_rows
                )
            
            if success:
                restored_tables += 1
                restored_rows += rows
            else:
                errors.append(f"{schema_name}.{table_name}")
        
        if errors:
            error_msg = f"Failed to restore {len(errors)} tables: {', '.join(errors[:5])}"
            log_warning(error_msg)
            if len(errors) > len(tables_to_restore) * 0.1:
                return False, error_msg
        
        if restored_rows == 0 and total_expected_rows > 0:
            return False, "Restore completed but 0 rows restored"
        
        success_rate = (restored_rows / total_expected_rows * 100) if total_expected_rows > 0 else 100
        if success_rate < 90:
            log_warning(f"Restore success rate: {success_rate:.1f}%")
        
        log_success(f"Restore complete: {restored_tables} tables, {restored_rows:,} rows")
        return True, f"Restored {restored_tables} tables with {restored_rows:,} rows"
    
    def _restore_single_table(
        self, 
        schema: str, 
        table: str, 
        sql_file: Path, 
        skip_truncate: bool = False,
        max_retries: int = 3,
        expected_rows: int = 0
    ) -> Tuple[bool, int]:
        """Restore single table from .sql.gz file with retry logic."""
        last_error = ""
        
        for attempt in range(1, max_retries + 1):
            try:
                if not skip_truncate:
                    clear_cmd = [
                        "docker", "exec", "omni-postgres", "psql", "-U", "omni", 
                        "-d", "omni_main", "-c", f"TRUNCATE {schema}.{table};"
                    ]
                    subprocess.run(clear_cmd, capture_output=True, timeout=30)
                
                with gzip.open(sql_file, 'rb') as f:
                    sql_bytes = f.read()
                sql_content = sql_bytes.decode('utf-8')
                
                # Strip pg_dump's \restrict / \unrestrict meta-commands.
                # pg_dump --data-only generates these to disable/enable triggers
                # during restore, but psql doesn't recognize them as valid
                # commands, causing subsequent COPY statements to silently fail.
                cleaned_lines = []
                for line in sql_content.split('\n'):
                    if line.startswith('\\restrict') or line.startswith('\\unrestrict'):
                        continue
                    cleaned_lines.append(line)
                sql_content = '\n'.join(cleaned_lines)
                
                # Disable triggers (FK constraints) before restore to allow
                # insertion regardless of table ordering
                disable_cmd = [
                    "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                    "-d", "omni_main", "-c",
                    f"ALTER TABLE {schema}.{table} DISABLE TRIGGER ALL;"
                ]
                subprocess.run(disable_cmd, capture_output=True, timeout=30)
                
                result = subprocess.run(
                    ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main"],
                    input=sql_content,
                    capture_output=True,
                    text=True,
                    timeout=300,
                    encoding='utf-8',
                    errors='replace',
                )
                
                # Re-enable triggers after restore
                enable_cmd = [
                    "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                    "-d", "omni_main", "-c",
                    f"ALTER TABLE {schema}.{table} ENABLE TRIGGER ALL;"
                ]
                subprocess.run(enable_cmd, capture_output=True, timeout=30)
                
                if result.returncode == 0:
                    count_result = subprocess.run(
                        ["docker", "exec", "omni-postgres", "psql", "-U", "omni", 
                         "-d", "omni_main", "-t", "-c", f"SELECT COUNT(*) FROM {schema}.{table};"],
                        capture_output=True, text=True, timeout=30,
                    )
                    rows = int(count_result.stdout.strip()) if count_result.returncode == 0 else 0
                    
                    # Detect false positive: 0 rows restored but expected > 0
                    if rows == 0 and expected_rows > 0:
                        last_error = f"0 rows restored (expected {expected_rows:,})"
                        print(f"  [WARN] {schema}.{table}: {last_error}")
                        if attempt < max_retries:
                            time.sleep(2)
                        continue  # retry
                    
                    print(f"  [OK] {schema}.{table}: {rows:,} rows")
                    return True, rows
                else:
                    last_error = result.stderr[:100]
                    if attempt < max_retries:
                        time.sleep(2)
                    
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
    
    def _restore_chunked_files(
        self, 
        schema: str, 
        table: str, 
        chunk_files: list, 
        expected: int
    ) -> Tuple[bool, int]:
        """Restore large table from .chunk*.sql.gz files."""
        try:
            total_chunks = len(chunk_files)
            print(f"  [INFO] {schema}.{table}: restoring {total_chunks} chunks...")
            
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
                for attempt in range(3):
                    try:
                        with gzip.open(chunk_file, 'rb') as f:
                            tsv_bytes = f.read()
                        tsv_data = tsv_bytes.decode('utf-8')
                        
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
            
            count_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main", "-t", "-c", f"SELECT COUNT(*) FROM {schema}.{table};"],
                capture_output=True, text=True, timeout=30,
            )
            rows = int(count_result.stdout.strip()) if count_result.returncode == 0 else 0
            print(f"  [OK] {schema}.{table}: {rows:,}/{expected:,} rows")
            return True, rows
            
        except Exception as e:
            print(f"  [ERROR] {schema}.{table}: {e}")
            return False, 0
