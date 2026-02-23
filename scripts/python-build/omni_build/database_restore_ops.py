"""
Database Restore Operations module for Omni Build System.

SRP: This module handles table and schema restore operations.
"""
import gzip
import hashlib
import subprocess
import time
from pathlib import Path
from typing import Any, List, Tuple

from omni_build.config import Config
from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseRestoreOps:
    """Handles database restore operations for tables and schemas."""
    
    def __init__(self, config: Config, pg_checker) -> None:
        self.config = config
        self._pg_checker = pg_checker

    def _parse_count(self, raw_count: str) -> int:
        """Parse COUNT(*) output safely."""
        try:
            return int(raw_count.strip())
        except (TypeError, ValueError, AttributeError):
            return -1

    def _get_table_count(self, schema: str, table: str) -> Tuple[bool, int, str]:
        """Get table row count from PostgreSQL."""
        count_result = subprocess.run(
            [
                "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                "-d", "omni_main", "-t", "-c", f"SELECT COUNT(*) FROM {schema}.{table};"
            ],
            capture_output=True,
            text=True,
            timeout=30,
            encoding='utf-8',
            errors='replace',
        )
        if count_result.returncode != 0:
            return False, 0, f"count query failed: {count_result.stderr[:100]}"

        rows = self._parse_count(count_result.stdout)
        if rows < 0:
            return False, 0, "invalid row count output"

        return True, rows, ""

    def _update_hash_from_file(self, file_path: Path, hasher: "hashlib._Hash") -> None:
        """Update hasher from file bytes in chunks."""
        with open(file_path, 'rb') as f:
            while True:
                chunk = f.read(1024 * 1024)
                if not chunk:
                    break
                hasher.update(chunk)

    def _is_sha256_checksum(self, checksum: Any) -> bool:
        """Validate checksum format as sha256 hex string."""
        if not isinstance(checksum, str):
            return False
        if len(checksum) != 64:
            return False
        return all(c in "0123456789abcdef" for c in checksum.lower())

    def _calculate_backup_table_checksum(self, schema: str, table: str, schema_dir: Path) -> Tuple[bool, str, str]:
        """Calculate checksum for backup artifact(s) used by restore."""
        hasher = hashlib.sha256()
        table_file = schema_dir / f"{table}.sql.gz"
        chunk_files = sorted(schema_dir.glob(f"{table}.chunk*.sql.gz"))

        if table_file.exists():
            hasher.update(f"single:{schema}.{table}".encode('utf-8'))
            self._update_hash_from_file(table_file, hasher)
            return True, hasher.hexdigest(), ""

        if chunk_files:
            hasher.update(f"chunked:{schema}.{table}".encode('utf-8'))
            for chunk_file in chunk_files:
                hasher.update(chunk_file.name.encode('utf-8'))
                self._update_hash_from_file(chunk_file, hasher)
            return True, hasher.hexdigest(), ""

        return False, "", "backup artifact not found"

    def _verify_backup_checksum(
        self,
        schema: str,
        table: str,
        schema_dir: Path,
        expected_checksum: str,
    ) -> Tuple[bool, str]:
        """Verify backup file checksum against manifest checksum."""
        ok, actual_checksum, error_message = self._calculate_backup_table_checksum(schema, table, schema_dir)
        if not ok:
            return False, error_message

        if actual_checksum != expected_checksum:
            return False, (
                f"checksum mismatch: expected {expected_checksum[:12]}..., "
                f"actual {actual_checksum[:12]}..."
            )

        return True, ""
    
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
    
    def restore_tables(self, manifest: dict[str, Any], data_dir: Path) -> Tuple[bool, str]:
        """Restore tables from backup files."""
        schemas = manifest.get('schemas', {})
        restored_tables = 0
        restored_rows = 0
        errors: List[str] = []
        truncate_failures = set()
        
        tables_to_restore: List[Tuple[str, str, int, str, Path]] = []
        total_expected_rows = 0
        
        for schema_name, schema_data in schemas.items():
            schema_dir = data_dir / schema_name
            if not schema_dir.exists():
                log_warning(f"Schema directory not found: {schema_dir}")
                continue
            
            for table_info in schema_data.get('tables', []):
                table_name = table_info.get('name')
                expected_rows = table_info.get('rows', 0)
                if not table_name:
                    continue

                # Materialized views are schema objects, not restorable table data.
                if table_name.startswith("mv_"):
                    continue

                expected_rows = expected_rows if isinstance(expected_rows, int) and expected_rows >= 0 else 0

                checksum = table_info.get('checksum')
                expected_checksum = checksum if self._is_sha256_checksum(checksum) else ""
                tables_to_restore.append((schema_name, table_name, expected_rows, expected_checksum, schema_dir))
                total_expected_rows += expected_rows
        
        log_info(f"Clearing {len(tables_to_restore)} tables...")
        for schema_name, table_name, _, _, _ in tables_to_restore:
            clear_cmd = [
                "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                "-d", "omni_main", "-c", f"TRUNCATE {schema_name}.{table_name} CASCADE;"
            ]
            clear_result = subprocess.run(
                clear_cmd,
                capture_output=True,
                text=True,
                timeout=30,
                encoding='utf-8',
                errors='replace',
            )
            if clear_result.returncode != 0:
                truncate_failures.add((schema_name, table_name))
                errors.append(
                    f"{schema_name}.{table_name} (truncate failed: {clear_result.stderr[:100]})"
                )
        
        for schema_name, table_name, expected_rows, expected_checksum, schema_dir in tables_to_restore:
            if (schema_name, table_name) in truncate_failures:
                continue

            if expected_checksum:
                checksum_ok, checksum_error = self._verify_backup_checksum(
                    schema_name,
                    table_name,
                    schema_dir,
                    expected_checksum,
                )
                if not checksum_ok:
                    errors.append(f"{schema_name}.{table_name} ({checksum_error})")
                    continue

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
                    if expected_rows == 0:
                        count_ok, rows, count_error = self._get_table_count(schema_name, table_name)
                        if not count_ok:
                            errors.append(f"{schema_name}.{table_name} ({count_error})")
                        elif rows != 0:
                            errors.append(
                                f"{schema_name}.{table_name} (row mismatch: restored {rows:,}, expected 0)"
                            )
                        else:
                            restored_tables += 1
                        continue

                    errors.append(f"{schema_name}.{table_name} (backup file not found)")
                    continue
                success, rows = self._restore_single_table(
                    schema_name, table_name, sql_file, skip_truncate=True,
                    expected_rows=expected_rows
                )

            if success:
                if rows != expected_rows:
                    errors.append(
                        f"{schema_name}.{table_name} (row mismatch: restored {rows:,}, expected {expected_rows:,})"
                    )
                else:
                    restored_tables += 1
                    restored_rows += rows
            else:
                errors.append(f"{schema_name}.{table_name}")

        if errors:
            error_msg = f"Restore validation failed for {len(errors)} table(s): {', '.join(errors[:5])}"
            log_error(error_msg)
            return False, error_msg

        if restored_rows == 0 and total_expected_rows > 0:
            return False, "Restore completed but 0 rows restored"

        if restored_rows != total_expected_rows:
            return False, (
                "Restore row total mismatch: "
                f"restored {restored_rows:,}, expected {total_expected_rows:,}"
            )

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
                    count_ok, rows, count_error = self._get_table_count(schema, table)
                    if not count_ok:
                        last_error = count_error
                        if attempt < max_retries:
                            time.sleep(2)
                        continue

                    if rows != expected_rows:
                        last_error = f"row mismatch: restored {rows:,}, expected {expected_rows:,}"
                        print(f"  [WARN] {schema}.{table}: {last_error}")
                        if attempt < max_retries:
                            time.sleep(2)
                        continue

                    print(f"  [OK] {schema}.{table}: {rows:,}/{expected_rows:,} rows")
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
        chunk_files: List[Path], 
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
                        with gzip.open(chunk_file, 'rt', encoding='utf-8', errors='replace') as f:
                            tsv_data = f.read()
                        
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
                print(f"  [ERROR] {failed_chunks}/{total_chunks} chunks failed")
                return False, 0

            count_ok, rows, count_error = self._get_table_count(schema, table)
            if not count_ok:
                print(f"  [ERROR] {schema}.{table}: {count_error}")
                return False, 0

            if rows != expected:
                print(f"  [ERROR] {schema}.{table}: row mismatch {rows:,}/{expected:,}")
                return False, rows

            print(f"  [OK] {schema}.{table}: {rows:,}/{expected:,} rows")
            return True, rows
            
        except Exception as e:
            print(f"  [ERROR] {schema}.{table}: {e}")
            return False, 0
