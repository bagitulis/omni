"""
NDJSON per-table operations for Omni Build System.

SRP: Single-table export and import logic.
Extracted from ndjson_sync.py for ~300 line compliance.
"""
import gzip
import json
from pathlib import Path
from typing import Callable, List, Tuple

from omni_build.logger import log_warning
from omni_build.ndjson_sql_helpers import build_batch_insert, build_single_upsert


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
) -> Tuple[int, int]:
    """Export a single table to NDJSON. Returns (exported_count, error_count)."""
    col_list = ', '.join(f'"{c}"' for c in all_columns)
    pk_order = ', '.join(f'"{c}"' for c in pk_columns) if pk_columns else '1'

    query = (
        f"SELECT row_to_json(t) FROM "
        f"(SELECT {col_list} FROM {schema}.{name} ORDER BY {pk_order}) t;"
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
    use_lfs = is_large or content_size > large_file_size_bytes

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


def import_table(
    schema: str,
    name: str,
    full_name: str,
    all_columns: List[str],
    pk_columns: List[str],
    filepath: Path,
    compressed: bool,
    psql_exec_fn: Callable[[str, int], Tuple[bool, str]],
) -> Tuple[int, int, int]:
    """Import a single NDJSON file. Returns (imported, skipped, errors)."""
    if compressed:
        with gzip.open(filepath, 'rt', encoding='utf-8') as f:
            lines = [line.strip() for line in f if line.strip()]
    else:
        with open(filepath, 'r', encoding='utf-8') as f:
            lines = [line.strip() for line in f if line.strip()]

    if not lines:
        return 0, 0, 0

    from decimal import Decimal
    first_row = json.loads(lines[0], parse_float=Decimal)
    available_cols = [c for c in all_columns if c in first_row]

    if not available_cols:
        log_warning(f"  {full_name}: no matching columns")
        return 0, 0, 1

    pk_cols = pk_columns
    non_pk_cols = [c for c in available_cols if c not in pk_cols]

    if not pk_cols:
        log_warning(f"  {full_name}: no primary key, skipping")
        return 0, len(lines), 0

    imported = 0
    error_count = 0
    batch_size = 1000

    for batch_start in range(0, len(lines), batch_size):
        batch = lines[batch_start:batch_start + batch_size]
        batch_sql = build_batch_insert(
            schema, name, available_cols, pk_cols, non_pk_cols, batch
        )

        ok, _ = psql_exec_fn(batch_sql, 120)
        if ok:
            imported += len(batch)
        else:
            for line in batch:
                row = json.loads(line, parse_float=Decimal)
                single_sql = build_single_upsert(
                    schema, name, available_cols, pk_cols, non_pk_cols, row
                )
                single_ok, _ = psql_exec_fn(single_sql, 30)
                if single_ok:
                    imported += 1
                else:
                    error_count += 1

    status = "OK" if error_count == 0 else "WARN"
    print(f"  [{status}] {full_name}: {imported:,}/{len(lines):,} rows")

    return imported, 0, error_count
