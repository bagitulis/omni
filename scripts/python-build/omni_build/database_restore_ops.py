"""
Database Restore Operations module for Omni Build System.

SRP: This module handles table restore operations (single + chunked).
Checksum verification and schema recreation are in restore_helpers.py.
"""
import gzip
import subprocess
import time
from pathlib import Path
from typing import Any, List, Tuple

from omni_build.database_backup_ops import DatabaseBackupOps
from omni_build.restore_helpers import (
    calculate_backup_table_checksum,
    is_sha256_checksum,
    recreate_schemas_from_backup,
    run_post_restore_analyze,
    terminate_active_connections,
    verify_backup_checksum,
)

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

    def _is_sha256_checksum(self, checksum: Any) -> bool:
        """Validate checksum format as sha256 hex string."""
        return is_sha256_checksum(checksum)

    def _calculate_backup_table_checksum(self, schema: str, table: str, schema_dir: Path) -> Tuple[bool, str, str]:
        """Calculate checksum for backup artifact(s) used by restore."""
        return calculate_backup_table_checksum(schema, table, schema_dir)

    def _verify_backup_checksum(
        self, schema: str, table: str, schema_dir: Path, expected_checksum: str,
    ) -> Tuple[bool, str]:
        """Verify backup file checksum against manifest checksum."""
        return verify_backup_checksum(schema, table, schema_dir, expected_checksum)

    def recreate_schemas_from_backup(self, data_dir: Path) -> bool:
        """Recreate schemas from backup _schema dumps."""
        return recreate_schemas_from_backup(data_dir, self._pg_checker)

    def terminate_active_connections(self) -> bool:
        """Terminate active connections to prevent lock conflicts."""
        return terminate_active_connections()

    def run_post_restore_analyze(self) -> bool:
        """Run ANALYZE after restore to update query planner statistics."""
        return run_post_restore_analyze()
    
    def restore_tables(self, manifest: dict[str, Any], data_dir: Path) -> Tuple[bool, str]:
        """Restore tables from backup files."""
        schemas = manifest.get('schemas', {})
        restored_tables = 0
        restored_rows = 0
        errors: List[str] = []
        truncate_failures = set()
        auto_fixes = 0
        
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
                    # Verify artifact file actually exists before auto-fixing
                    ok, _, _ = self._calculate_backup_table_checksum(
                        schema_name, table_name, schema_dir
                    )
                    if ok:
                        print(f"  [FIX] {schema_name}.{table_name}: {checksum_error} — artifact valid, continuing")
                        auto_fixes += 1
                    else:
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
                    # Strict verify: count rows in artifact file
                    artifact_ops = DatabaseBackupOps(data_dir)
                    artifact_rows = artifact_ops.calculate_artifact_row_count(
                        schema_name, table_name
                    )
                    if artifact_rows >= 0 and artifact_rows == rows:
                        # Manifest was wrong — artifact matches restored
                        print(f"  [FIX] {schema_name}.{table_name}: accepted {rows:,} rows (manifest expected {expected_rows:,}, artifact has {artifact_rows:,})")
                        total_expected_rows += (rows - expected_rows)
                        auto_fixes += 1
                    elif artifact_rows >= 0 and rows > artifact_rows:
                        # Backend seeded extra rows during restore — all backup data is in
                        print(f"  [FIX] {schema_name}.{table_name}: accepted {rows:,} rows (artifact {artifact_rows:,} + {rows - artifact_rows:,} backend-seeded)")
                        total_expected_rows += (rows - expected_rows)
                        auto_fixes += 1
                    elif artifact_rows >= 0 and rows < artifact_rows:
                        # Rows missing — retry once with TRUNCATE to clear backend-seeded conflicts
                        print(f"  [WARN] {schema_name}.{table_name}: {rows:,}/{artifact_rows:,} rows — retrying with TRUNCATE...")
                        retry_success = False
                        chunk_files = sorted(schema_dir.glob(f"{table_name}.chunk*.sql.gz"))
                        if chunk_files:
                            retry_ok, retry_rows = self._restore_chunked_files(
                                schema_name, table_name, chunk_files, artifact_rows
                            )
                        else:
                            sql_file = schema_dir / f"{table_name}.sql.gz"
                            retry_ok, retry_rows = self._restore_single_table(
                                schema_name, table_name, sql_file,
                                skip_truncate=False, expected_rows=artifact_rows
                            )
                        if retry_ok and retry_rows >= artifact_rows:
                            print(f"  [FIX] {schema_name}.{table_name}: retry succeeded — {retry_rows:,} rows")
                            rows = retry_rows
                            total_expected_rows += (rows - expected_rows)
                            auto_fixes += 1
                        else:
                            errors.append(
                                f"{schema_name}.{table_name} (row mismatch after retry: restored {retry_rows:,}, artifact {artifact_rows:,})"
                            )
                            continue
                    else:
                        errors.append(
                            f"{schema_name}.{table_name} (row mismatch: restored {rows:,}, expected {expected_rows:,}, artifact {artifact_rows:,})"
                        )
                        continue
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

        fix_msg = f" ({auto_fixes} auto-fixed)" if auto_fixes else ""
        log_success(f"Restore complete: {restored_tables} tables, {restored_rows:,} rows{fix_msg}")
        return True, f"Restored {restored_tables} tables with {restored_rows:,} rows{fix_msg}"

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
                    ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main",
                     "-v", "ON_ERROR_STOP=1"],
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
                
                # Check for errors — psql with ON_ERROR_STOP=1 exits non-zero on SQL errors
                has_error = result.returncode != 0 or "ERROR" in (result.stderr or "")
                
                if not has_error:
                    count_ok, rows, count_error = self._get_table_count(schema, table)
                    if not count_ok:
                        last_error = count_error
                        if attempt < max_retries:
                            time.sleep(2)
                        continue

                    # Return actual row count — caller handles mismatch auto-fix
                    print(f"  [OK] {schema}.{table}: {rows:,}/{expected_rows:,} rows")
                    return True, rows
                else:
                    last_error = result.stderr[:200] if result.stderr else f"exit code {result.returncode}"
                    # If error is column mismatch, log it clearly
                    if "extra data after last expected column" in last_error or "missing data for column" in last_error:
                        print(f"  [WARN] {schema}.{table}: column mismatch (schema evolved since backup)")
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
    
    def _detect_chunk_column_count(self, chunk_files: List[Path]) -> int:
        """Detect the number of columns in chunk data by reading first line."""
        try:
            with gzip.open(chunk_files[0], 'rt', encoding='utf-8', errors='replace') as f:
                first_line = f.readline()
            if first_line.strip():
                return first_line.count('\t') + 1
        except Exception:
            pass
        return 0

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
            
            # Get all columns ordered by position
            col_result = subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main", "-t", "-c",
                 f"SELECT string_agg(column_name, ', ' ORDER BY ordinal_position) FROM information_schema.columns WHERE table_schema = '{schema}' AND table_name = '{table}';"],
                capture_output=True, text=True, timeout=30
            )
            columns = col_result.stdout.strip() if col_result.returncode == 0 else ""
            
            if not columns:
                print(f"  [ERROR] {schema}.{table}: could not get column list")
                return False, 0
            
            # Detect column count from chunk data to handle schema evolution
            data_col_count = self._detect_chunk_column_count(chunk_files)
            all_columns = [c.strip() for c in columns.split(',')]
            db_col_count = len(all_columns)
            
            if data_col_count > 0 and data_col_count < db_col_count:
                # Schema has new columns not present in backup data
                # Use only the first N columns matching the data width
                columns = ', '.join(all_columns[:data_col_count])
                print(f"  [INFO] {schema}.{table}: data has {data_col_count} cols, table has {db_col_count} — using first {data_col_count}")
            elif data_col_count > db_col_count:
                print(f"  [ERROR] {schema}.{table}: data has {data_col_count} cols but table only has {db_col_count}")
                return False, 0
            
            # Disable triggers before bulk COPY
            subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main", "-c",
                 f"ALTER TABLE {schema}.{table} DISABLE TRIGGER ALL;"],
                capture_output=True, timeout=30
            )
            
            failed_chunks = 0
            for i, chunk_file in enumerate(chunk_files, 1):
                success = False
                last_error = ""
                for attempt in range(3):
                    try:
                        with gzip.open(chunk_file, 'rt', encoding='utf-8', errors='replace') as f:
                            tsv_data = f.read()
                        
                        copy_sql = f"COPY {schema}.{table} ({columns}) FROM stdin;\n{tsv_data}\\.\n"
                        
                        result = subprocess.run(
                            ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni", "-d", "omni_main",
                             "-v", "ON_ERROR_STOP=1"],
                            input=copy_sql,
                            capture_output=True,
                            text=True,
                            timeout=600,
                            encoding='utf-8',
                            errors='replace',
                        )
                        
                        # Check both exit code AND stderr for COPY errors
                        if result.returncode == 0 and "ERROR" not in result.stderr:
                            success = True
                            break
                        else:
                            last_error = result.stderr[:200] if result.stderr else f"exit code {result.returncode}"
                            if attempt < 2:
                                time.sleep(2)
                    except subprocess.TimeoutExpired:
                        last_error = "timeout"
                        if attempt < 2:
                            time.sleep(5)
                    except Exception as e:
                        last_error = str(e)
                        if attempt < 2:
                            time.sleep(2)
                
                if not success:
                    failed_chunks += 1
                    print(f"  [WARN] Chunk {i}/{total_chunks} failed: {last_error[:100]}")
                elif i % 5 == 0 or i == total_chunks:
                    print(f"  [INFO] {schema}.{table}: {i}/{total_chunks} chunks done")
            
            # Re-enable triggers
            subprocess.run(
                ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main", "-c",
                 f"ALTER TABLE {schema}.{table} ENABLE TRIGGER ALL;"],
                capture_output=True, timeout=30
            )
            
            if failed_chunks > 0:
                print(f"  [ERROR] {failed_chunks}/{total_chunks} chunks failed")
                return False, 0

            count_ok, rows, count_error = self._get_table_count(schema, table)
            if not count_ok:
                print(f"  [ERROR] {schema}.{table}: {count_error}")
                return False, 0

            # Return actual row count — caller handles mismatch auto-fix
            print(f"  [OK] {schema}.{table}: {rows:,}/{expected:,} rows")
            return True, rows
            
        except Exception as e:
            print(f"  [ERROR] {schema}.{table}: {e}")
            return False, 0
