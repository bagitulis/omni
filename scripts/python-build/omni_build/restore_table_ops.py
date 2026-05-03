"""
Per-table restore operations for database restore.

SRP: This module handles individual table restore logic (single + chunked).
The orchestration logic remains in DatabaseRestoreOps.
"""
import gzip
import subprocess
import time
from pathlib import Path
from typing import List, Tuple

from omni_build.db_config import DatabaseConfig


def restore_single_table(
    schema: str,
    table: str,
    sql_file: Path,
    pg_checker,
    skip_truncate: bool = False,
    max_retries: int = 3,
    expected_rows: int = 0
) -> Tuple[bool, int]:
    """Restore single table from .sql.gz file with retry logic."""
    last_error = ""
    
    for attempt in range(1, max_retries + 1):
        try:
            if not skip_truncate:
                clear_cmd = DatabaseConfig.psql_cmd() + [
                    "-c", f"TRUNCATE {schema}.{table};"
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
            disable_cmd = DatabaseConfig.psql_cmd() + [
                "-c",
                f"ALTER TABLE {schema}.{table} DISABLE TRIGGER ALL;"
            ]
            subprocess.run(disable_cmd, capture_output=True, timeout=30)
            
            result = subprocess.run(
                DatabaseConfig.psql_cmd(interactive=True) + ["-v", "ON_ERROR_STOP=1"],
                input=sql_content,
                capture_output=True,
                text=True,
                timeout=300,
                encoding='utf-8',
                errors='replace',
            )
            
            # Re-enable triggers after restore
            enable_cmd = DatabaseConfig.psql_cmd() + [
                "-c",
                f"ALTER TABLE {schema}.{table} ENABLE TRIGGER ALL;"
            ]
            subprocess.run(enable_cmd, capture_output=True, timeout=30)
            
            # Check for errors — psql with ON_ERROR_STOP=1 exits non-zero on SQL errors
            has_error = result.returncode != 0 or "ERROR" in (result.stderr or "")
            
            if not has_error:
                count_ok, rows, count_error = _get_table_count(schema, table)
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
                # If error is duplicate key, force TRUNCATE on next retry
                if "duplicate key" in last_error or "already exists" in last_error:
                    skip_truncate = False
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


def restore_chunked_files(
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
            DatabaseConfig.psql_cmd() + ["-t", "-c",
             f"SELECT string_agg(column_name, ', ' ORDER BY ordinal_position) FROM information_schema.columns WHERE table_schema = '{schema}' AND table_name = '{table}';"],
            capture_output=True, text=True, timeout=30
        )
        columns = col_result.stdout.strip() if col_result.returncode == 0 else ""
        
        if not columns:
            print(f"  [ERROR] {schema}.{table}: could not get column list")
            return False, 0
        
        # Detect column count from chunk data to handle schema evolution
        data_col_count = _detect_chunk_column_count(chunk_files)
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
            DatabaseConfig.psql_cmd() + [
                "-c",
                f"ALTER TABLE {schema}.{table} DISABLE TRIGGER ALL;"
            ],
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
                        DatabaseConfig.psql_cmd(interactive=True) + ["-v", "ON_ERROR_STOP=1"],
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
            DatabaseConfig.psql_cmd() + [
                "-c",
                f"ALTER TABLE {schema}.{table} ENABLE TRIGGER ALL;"
            ],
            capture_output=True, timeout=30
        )
        
        if failed_chunks > 0:
            print(f"  [ERROR] {failed_chunks}/{total_chunks} chunks failed")
            return False, 0

        count_ok, rows, count_error = _get_table_count(schema, table)
        if not count_ok:
            print(f"  [ERROR] {schema}.{table}: {count_error}")
            return False, 0

        # Return actual row count — caller handles mismatch auto-fix
        print(f"  [OK] {schema}.{table}: {rows:,}/{expected:,} rows")
        return True, rows
        
    except Exception as e:
        print(f"  [ERROR] {schema}.{table}: {e}")
        return False, 0


def _detect_chunk_column_count(chunk_files: List[Path]) -> int:
    """Detect the number of columns in chunk data by reading first line."""
    try:
        with gzip.open(chunk_files[0], 'rt', encoding='utf-8', errors='replace') as f:
            first_line = f.readline()
        if first_line.strip():
            return first_line.count('\t') + 1
    except Exception:
        pass
    return 0


def _get_table_count(schema: str, table: str) -> Tuple[bool, int, str]:
    """Get table row count from PostgreSQL."""
    count_result = subprocess.run(
        DatabaseConfig.psql_cmd() + [
            "-t", "-c", f"SELECT COUNT(*) FROM {schema}.{table};"
        ],
        capture_output=True,
        text=True,
        timeout=30,
        encoding='utf-8',
        errors='replace',
    )
    if count_result.returncode != 0:
        return False, 0, f"count query failed: {count_result.stderr[:100]}"

    rows = _parse_count(count_result.stdout)
    if rows < 0:
        return False, 0, "invalid row count output"

    return True, rows, ""


def _parse_count(raw_count: str) -> int:
    """Parse COUNT(*) output safely."""
    try:
        return int(raw_count.strip())
    except (TypeError, ValueError, AttributeError):
        return -1
