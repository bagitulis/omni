"""
NDJSON per-table operations for Omni Build System.

SRP: Single-table export and import logic.
Extracted from ndjson_sync.py for ~300 line compliance.
"""
import gzip
import json
from pathlib import Path
from typing import Callable, List, Optional, Tuple

from omni_build.logger import log_error, log_warning
from omni_build.ndjson_sql_helpers import (
    build_batch_insert, build_batch_plain_insert, build_single_upsert,
)


def export_table(
    schema: str,
    name: str,
    full_name: str,
    all_columns: List[str],
    pk_columns: List[str],
    is_large: bool,
    dest_dir: Path,
    large_file_size_bytes: int,
    psql_query_fn: Callable[[str, int], Tuple[bool, str]],
    psql_exec_fn: Optional[Callable[[str, int], Tuple[bool, str]]] = None,
) -> Tuple[int, int]:
    """Export a single table to NDJSON. Returns (exported_count, error_count).

    Uses SELECT row_to_json() via psql_query_fn (-t -A) for clean JSON output.
    """
    col_list = ', '.join('"' + c + '"' for c in all_columns)
    pk_order = ', '.join('"' + c + '"' for c in pk_columns) if pk_columns else '1'

    # Use psql with -t -A flags for clean JSON output (no headers).
    # NOTE: COPY TO STDOUT double-escapes backslashes, corrupting JSON.
    query = (
        f"SELECT row_to_json(t) FROM "
        f"(SELECT {col_list} FROM {schema}.{name} ORDER BY {pk_order}) t;"
    )

    # For large tables, stream directly to file to avoid MemoryError
    if is_large:
        return _export_table_streaming(
            schema, name, full_name, query, dest_dir,
            large_file_size_bytes, psql_query_fn
        )

    ok, output = psql_query_fn(query, 300)

    if not ok:
        log_warning(f"  Export {full_name} failed: {output[:100]}")
        return 0, 1

    lines = [line for line in output.split('\n') if line.strip()]
    count = len(lines)

    if count == 0:
        filepath = dest_dir / f"{name}.ndjson"
        filepath.write_text("", encoding='utf-8')
        return 0, 0

    content = '\n'.join(lines) + '\n'
    content_size = len(content.encode('utf-8'))
    use_lfs = content_size > large_file_size_bytes

    if use_lfs:
        large_dir = dest_dir.parent / "_large"
        large_dir.mkdir(parents=True, exist_ok=True)
        filename = f"{schema}.{name}.ndjson.gz"
        filepath = large_dir / filename
        with gzip.open(filepath, 'wt', encoding='utf-8') as f:
            f.write(content)
        size_mb = filepath.stat().st_size / (1024 * 1024)
        print(f"  [OK] {full_name}: {count:,} rows ({size_mb:.1f} MB, compressed)")
        plain_file = dest_dir / f"{name}.ndjson"
        if plain_file.exists():
            plain_file.unlink()
    else:
        filepath = dest_dir / f"{name}.ndjson"
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"  [OK] {full_name}: {count:,} rows")
        large_dir = dest_dir.parent / "_large"
        compressed_file = large_dir / f"{schema}.{name}.ndjson.gz"
        if compressed_file.exists():
            compressed_file.unlink()

    return count, 0


def _export_table_streaming(
    schema: str, name: str, full_name: str, query: str,
    dest_dir: Path, large_file_size_bytes: int,
    psql_query_fn: Callable[[str, int], Tuple[bool, str]],
) -> Tuple[int, int]:
    """Export a large table by streaming psql output directly to gzip file.

    Avoids loading entire result set into memory.
    """
    import subprocess as sp
    from omni_build.db_config import DatabaseConfig

    large_dir = dest_dir.parent / "_large"
    large_dir.mkdir(parents=True, exist_ok=True)
    filename = f"{schema}.{name}.ndjson.gz"
    filepath = large_dir / filename

    # Stream psql output line-by-line to gzip file
    cmd = DatabaseConfig.psql_cmd(interactive=True) + ["-q", "-t", "-A"]
    try:
        proc = sp.Popen(
            cmd, stdin=sp.PIPE, stdout=sp.PIPE, stderr=sp.PIPE,
            text=True, encoding='utf-8', errors='replace'
        )
        proc.stdin.write(query)
        proc.stdin.close()

        count = 0
        with gzip.open(filepath, 'wt', encoding='utf-8') as f:
            for line in proc.stdout:
                stripped = line.rstrip('\n')
                if stripped:
                    f.write(stripped + '\n')
                    count += 1

        proc.wait(timeout=60)
        if proc.returncode != 0:
            err = proc.stderr.read()[:200]
            log_warning(f"  Export {full_name} failed: {err}")
            return 0, 1

    except sp.TimeoutExpired:
        proc.kill()
        log_warning(f"  Export {full_name} timed out")
        return 0, 1
    except Exception as e:
        log_warning(f"  Export {full_name} error: {e}")
        return 0, 1

    if count == 0:
        filepath.unlink(missing_ok=True)
        plain_file = dest_dir / f"{name}.ndjson"
        plain_file.write_text("", encoding='utf-8')
        return 0, 0

    size_mb = filepath.stat().st_size / (1024 * 1024)
    print(f"  [OK] {full_name}: {count:,} rows ({size_mb:.1f} MB, compressed)")

    # Remove plain file if exists
    plain_file = dest_dir / f"{name}.ndjson"
    if plain_file.exists():
        plain_file.unlink()

    return count, 0


def _open_ndjson(filepath: Path, compressed: bool):
    """Open an NDJSON file (plain or gzip) for line-by-line reading."""
    if compressed:
        return gzip.open(filepath, 'rt', encoding='utf-8')
    return open(filepath, 'r', encoding='utf-8')


def _stream_batches(filepath: Path, compressed: bool, batch_size: int):
    """Yield batches of non-empty lines from an NDJSON file without loading all into memory."""
    batch: List[str] = []
    with _open_ndjson(filepath, compressed) as f:
        for line in f:
            stripped = line.strip()
            if not stripped:
                continue
            batch.append(stripped)
            if len(batch) >= batch_size:
                yield batch
                batch = []
    if batch:
        yield batch


def import_table(
    schema: str,
    name: str,
    full_name: str,
    all_columns: List[str],
    pk_columns: List[str],
    filepath: Path,
    compressed: bool,
    psql_exec_fn: Callable[[str, int], Tuple[bool, str]],
    not_null_columns: Optional[List[str]] = None,
    use_plain_insert: bool = False,
) -> Tuple[int, int, int]:
    from decimal import Decimal

    # Pass 1: Stream through file to discover union of column keys
    # Only parses JSON keys, does not store full lines in memory
    ndjson_cols: set[str] = set()
    total_lines = 0
    with _open_ndjson(filepath, compressed) as f:
        for line in f:
            stripped = line.strip()
            if not stripped:
                continue
            total_lines += 1
            row_keys = json.loads(stripped).keys()
            ndjson_cols.update(row_keys)

    if total_lines == 0:
        return 0, 0, 0

    # Bug #9: Warn about columns in NDJSON that don't exist in target table
    target_col_set = set(all_columns)
    extra_cols = ndjson_cols - target_col_set
    if extra_cols:
        log_warning(
            f"  {full_name}: {len(extra_cols)} column(s) in NDJSON not in target table "
            f"(will be skipped): {', '.join(sorted(extra_cols))}"
        )

    available_cols = [c for c in all_columns if c in ndjson_cols]

    if not available_cols:
        log_warning(f"  {full_name}: no matching columns")
        return 0, 0, 1

    # Bug #8: Detect NOT NULL columns missing from NDJSON
    if not_null_columns:
        available_set = set(available_cols)
        missing_not_null = [c for c in not_null_columns if c not in available_set]
        if missing_not_null:
            log_error(
                f"  {full_name}: NOT NULL column(s) missing from NDJSON data: "
                f"{', '.join(missing_not_null)}. Import would fail with NOT NULL violation."
            )
            return 0, 0, 1

    pk_cols = pk_columns
    non_pk_cols = [c for c in available_cols if c not in pk_cols]

    if not pk_cols:
        log_warning(f"  {full_name}: no primary key, skipping")
        return 0, total_lines, 0

    # Pass 2: Stream file again in batches for import (memory-efficient)
    imported = 0
    error_count = 0
    batch_size = 1000
    max_batch_bytes = 10 * 1024 * 1024  # 10 MB

    for batch in _stream_batches(filepath, compressed, batch_size):
        if use_plain_insert:
            batch_sql = build_batch_plain_insert(schema, name, available_cols, batch)
        else:
            batch_sql = build_batch_insert(
                schema, name, available_cols, pk_cols, non_pk_cols, batch
            )

        # Bug #15: Dynamic batch size — reduce if SQL exceeds 10MB
        if len(batch_sql.encode('utf-8')) > max_batch_bytes and len(batch) > 1:
            half = len(batch) // 2
            for sub_batch in [batch[:half], batch[half:]]:
                if use_plain_insert:
                    sub_sql = build_batch_plain_insert(schema, name, available_cols, sub_batch)
                else:
                    sub_sql = build_batch_insert(
                        schema, name, available_cols, pk_cols, non_pk_cols, sub_batch
                    )
                ok, err = psql_exec_fn(sub_sql, 120)
                if ok:
                    imported += len(sub_batch)
                else:
                    for line in sub_batch:
                        row = json.loads(line, parse_float=Decimal)
                        single_sql = build_single_upsert(
                            schema, name, available_cols, pk_cols, non_pk_cols, row
                        )
                        single_ok, single_err = psql_exec_fn(single_sql, 30)
                        if single_ok:
                            imported += 1
                        else:
                            error_count += 1
                            if error_count <= 5:
                                log_warning(f"    Row error: {single_err[:200]}")
            continue

        ok, err = psql_exec_fn(batch_sql, 120)
        if ok:
            imported += len(batch)
        else:
            # Fallback: try one-by-one for failed batch
            if error_count == 0:
                log_warning(f"  Batch failed for {full_name}, retrying row-by-row...")
                log_warning(f"    Batch error: {err[:200]}")
            for line in batch:
                row = json.loads(line, parse_float=Decimal)
                single_sql = build_single_upsert(
                    schema, name, available_cols, pk_cols, non_pk_cols, row
                )
                single_ok, single_err = psql_exec_fn(single_sql, 30)
                if single_ok:
                    imported += 1
                else:
                    error_count += 1
                    if error_count <= 5:
                        log_warning(f"    Row error: {single_err[:200]}")

    status = "OK" if error_count == 0 else "WARN"
    print(f"  [{status}] {full_name}: {imported:,}/{total_lines:,} rows")

    return imported, 0, error_count
