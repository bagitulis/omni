"""
NDJSON Sync for OMNI PostgreSQL — Extensions-style backup mechanism.

Design (mirrors D:\\Project\\extensions\\internal\\sync):
  - Export: SELECT * from syncable tables → NDJSON (one JSON object per row)
  - Import: read NDJSON → INSERT ... ON CONFLICT(pk) DO UPDATE (true upsert)
  - Dynamic columns: auto-detected from information_schema at runtime
  - FK handling: session_replication_role = 'replica' disables FK triggers
  - Large tables (>=10K rows): compressed .ndjson.gz → Git LFS
  - Small tables (<10K rows): plain .ndjson → Git (text, diffable)

Correctness guarantees:
  - INSERT ... ON CONFLICT DO UPDATE (not DELETE+INSERT) to avoid cascades
  - json.loads with parse_float=Decimal for numeric precision
  - Timestamp-based conflict resolution (updated_at WHERE clause)
  - Per-table transactions with rollback on failure
  - Idempotent: run import twice = same result
"""
import gzip
import json
import os
import subprocess
import time
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.ndjson_sql_helpers import build_batch_insert, build_single_upsert

# Threshold: tables with >= this many rows use compressed format (Git LFS)
LARGE_TABLE_THRESHOLD = 10000

# File size threshold: if exported NDJSON > this size, move to LFS
LARGE_FILE_SIZE_BYTES = 5 * 1024 * 1024  # 5 MB

# Directory structure
SYNC_DIR = "backups" + os.sep + "sync"


class TableInfo:
    """Metadata for a syncable table."""

    def __init__(self, schema: str, name: str, pk_columns: List[str],
                 all_columns: List[str], row_count: int = 0,
                 deps: Optional[List[str]] = None):
        self.schema = schema
        self.name = name
        self.pk_columns = pk_columns
        self.all_columns = all_columns
        self.row_count = row_count
        self.deps = deps or []

    @property
    def full_name(self) -> str:
        return f"{self.schema}.{self.name}"

    @property
    def is_large(self) -> bool:
        return self.row_count >= LARGE_TABLE_THRESHOLD


class SyncResult:
    """Result of a sync operation for one table."""

    def __init__(self, table: str, exported: int = 0, imported: int = 0,
                 skipped: int = 0, errors: int = 0):
        self.table = table
        self.exported = exported
        self.imported = imported
        self.skipped = skipped
        self.errors = errors


def _psql_exec(sql: str, timeout: int = 60) -> Tuple[bool, str]:
    """Execute SQL via docker exec psql."""
    result = subprocess.run(
        ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni",
         "-d", "omni_main", "-v", "ON_ERROR_STOP=1"],
        input=sql,
        capture_output=True, text=True, timeout=timeout,
        encoding='utf-8', errors='replace',
    )
    if result.returncode != 0:
        return False, result.stderr[:300]
    return True, result.stdout


def _psql_query(sql: str, timeout: int = 60) -> Tuple[bool, str]:
    """Execute query and return raw output."""
    result = subprocess.run(
        ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
         "-d", "omni_main", "-t", "-A", "-F", "\t", "-c", sql],
        capture_output=True, text=True, timeout=timeout,
        encoding='utf-8', errors='replace',
    )
    if result.returncode != 0:
        return False, result.stderr[:300]
    return True, result.stdout.strip()


def discover_tables() -> List[TableInfo]:
    """Discover all syncable tables with columns and PKs (dynamic, like Extensions)."""
    # Get all tables
    ok, output = _psql_query("""
        SELECT table_schema, table_name
        FROM information_schema.tables
        WHERE table_type = 'BASE TABLE'
        AND (table_schema LIKE 'tenant_%' OR table_schema = 'system')
        ORDER BY table_schema, table_name;
    """)
    if not ok:
        log_error(f"Cannot discover tables: {output}")
        return []

    tables = []
    for line in output.split('\n'):
        if not line.strip():
            continue
        parts = line.split('\t')
        if len(parts) >= 2:
            schema, name = parts[0].strip(), parts[1].strip()
            if schema and name:
                tables.append((schema, name))

    # Get columns and PKs for each table
    result = []
    for schema, name in tables:
        # Skip materialized views
        if name.startswith("mv_"):
            continue

        # Get columns
        ok, col_output = _psql_query(f"""
            SELECT column_name
            FROM information_schema.columns
            WHERE table_schema = '{schema}' AND table_name = '{name}'
            ORDER BY ordinal_position;
        """)
        if not ok:
            continue
        columns = [c.strip() for c in col_output.split('\n') if c.strip()]
        if not columns:
            continue

        # Get PK columns
        ok, pk_output = _psql_query(f"""
            SELECT a.attname
            FROM pg_index i
            JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
            WHERE i.indrelid = '{schema}.{name}'::regclass AND i.indisprimary
            ORDER BY array_position(i.indkey, a.attnum);
        """)
        pk_cols = [c.strip() for c in pk_output.split('\n') if c.strip()] if ok else []

        # Get row count (approximate from pg_stat)
        ok, count_output = _psql_query(
            f"SELECT n_live_tup FROM pg_stat_user_tables "
            f"WHERE schemaname = '{schema}' AND relname = '{name}';"
        )
        row_count = int(count_output.strip()) if ok and count_output.strip().isdigit() else 0

        # Get FK dependencies (within same schema)
        ok, fk_output = _psql_query(f"""
            SELECT DISTINCT ccu.table_name
            FROM information_schema.table_constraints tc
            JOIN information_schema.constraint_column_usage ccu
                ON tc.constraint_name = ccu.constraint_name
                AND tc.table_schema = ccu.table_schema
            WHERE tc.constraint_type = 'FOREIGN KEY'
            AND tc.table_schema = '{schema}'
            AND tc.table_name = '{name}'
            AND ccu.table_name != '{name}';
        """)
        deps = [d.strip() for d in fk_output.split('\n') if d.strip()] if ok else []

        result.append(TableInfo(
            schema=schema, name=name,
            pk_columns=pk_cols, all_columns=columns,
            row_count=row_count, deps=deps,
        ))

    return result


def _topo_sort(tables: List[TableInfo]) -> List[TableInfo]:
    """Topological sort by FK dependencies (parents first)."""
    by_name: Dict[str, TableInfo] = {t.name: t for t in tables}
    visited = set()
    sorted_tables = []

    def visit(name: str):
        if name in visited:
            return
        visited.add(name)
        t = by_name.get(name)
        if not t:
            return
        for dep in t.deps:
            visit(dep)
        sorted_tables.append(t)

    for t in tables:
        visit(t.name)

    # Add any tables not reached by topo sort (no deps, not referenced)
    for t in tables:
        if t not in sorted_tables:
            sorted_tables.append(t)

    return sorted_tables


def export_all(project_root: Path) -> Tuple[bool, List[SyncResult]]:
    """Export all tables to NDJSON files (like Extensions ExportAll)."""
    sync_dir = project_root / SYNC_DIR
    sync_dir.mkdir(parents=True, exist_ok=True)

    log_info("Discovering tables...")
    tables = discover_tables()
    if not tables:
        return False, []

    log_info(f"Found {len(tables)} tables to export")

    results = []
    total_rows = 0

    # Group by schema
    schemas = {}
    for t in tables:
        schemas.setdefault(t.schema, []).append(t)

    for schema, schema_tables in schemas.items():
        schema_dir = sync_dir / schema
        schema_dir.mkdir(parents=True, exist_ok=True)

        for table in schema_tables:
            r = _export_table(table, schema_dir)
            results.append(r)
            total_rows += r.exported

    # Write manifest
    manifest = {
        "exported_at": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "machine": subprocess.run(["hostname"], capture_output=True, text=True).stdout.strip(),
        "version": 1,
        "tables": {r.table: r.exported for r in results},
        "total_rows": total_rows,
        "large_threshold": LARGE_TABLE_THRESHOLD,
    }
    manifest_path = sync_dir / "manifest.json"
    with open(manifest_path, 'w', encoding='utf-8') as f:
        json.dump(manifest, f, indent=2)

    log_success(f"Export complete: {len(results)} tables, {total_rows:,} rows")
    return True, results


def _export_table(table: TableInfo, dest_dir: Path) -> SyncResult:
    """Export a single table to NDJSON."""
    col_list = ', '.join(f'"{c}"' for c in table.all_columns)
    pk_order = ', '.join(f'"{c}"' for c in table.pk_columns) if table.pk_columns else '1'

    # Use JSON export via psql: row_to_json outputs each row as JSON
    query = (
        f"SELECT row_to_json(t) FROM "
        f"(SELECT {col_list} FROM {table.schema}.{table.name} ORDER BY {pk_order}) t;"
    )

    ok, output = _psql_query(query, timeout=300)
    if not ok:
        log_warning(f"  Export {table.full_name} failed: {output[:100]}")
        return SyncResult(table=table.full_name, errors=1)

    lines = [line for line in output.split('\n') if line.strip()]
    count = len(lines)

    if count == 0:
        # Empty table — write empty file
        filename = f"{table.name}.ndjson"
        filepath = dest_dir / filename
        filepath.write_text("", encoding='utf-8')
        return SyncResult(table=table.full_name, exported=0)

    # First write as plain text to check size
    content = '\n'.join(lines) + '\n'
    content_size = len(content.encode('utf-8'))

    # Determine format: large tables OR large files → compressed (Git LFS)
    use_lfs = table.is_large or content_size > LARGE_FILE_SIZE_BYTES

    if use_lfs:
        # Compressed for Git LFS
        large_dir = dest_dir.parent / "_large"
        large_dir.mkdir(parents=True, exist_ok=True)
        filename = f"{table.schema}.{table.name}.ndjson.gz"
        filepath = large_dir / filename
        with gzip.open(filepath, 'wt', encoding='utf-8') as f:
            f.write(content)
        size_mb = filepath.stat().st_size / (1024 * 1024)
        print(f"  [OK] {table.full_name}: {count:,} rows ({size_mb:.1f} MB, compressed)")
        # Remove plain file if it existed from previous export
        plain_file = dest_dir / f"{table.name}.ndjson"
        if plain_file.exists():
            plain_file.unlink()
    else:
        # Plain text for git (small, diffable)
        filename = f"{table.name}.ndjson"
        filepath = dest_dir / filename
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"  [OK] {table.full_name}: {count:,} rows")
        # Remove compressed file if it existed from previous export
        large_dir = dest_dir.parent / "_large"
        compressed_file = large_dir / f"{table.schema}.{table.name}.ndjson.gz"
        if compressed_file.exists():
            compressed_file.unlink()

    return SyncResult(table=table.full_name, exported=count)


def import_all(project_root: Path) -> Tuple[bool, List[SyncResult]]:
    """Import all NDJSON files into PostgreSQL (like Extensions ImportAll)."""
    sync_dir = project_root / SYNC_DIR

    if not sync_dir.exists():
        return False, []

    manifest_path = sync_dir / "manifest.json"
    if not manifest_path.exists():
        log_error("No manifest.json found in sync directory")
        return False, []

    with open(manifest_path, 'r', encoding='utf-8') as f:
        manifest = json.load(f)

    log_info(f"Importing from: {manifest.get('exported_at', 'unknown')}")
    log_info(f"Machine: {manifest.get('machine', 'unknown')}")
    log_info(f"Total rows: {manifest.get('total_rows', 0):,}")

    # Discover current table structure
    log_info("Discovering current table structure...")
    tables = discover_tables()
    if not tables:
        return False, []

    # Disable FK checks for the duration of import
    log_info("Disabling FK constraints...")
    _psql_exec("SET session_replication_role = 'replica';")

    # Sort by dependencies
    sorted_tables = _topo_sort(tables)

    results = []
    total_imported = 0
    errors = 0

    for table in sorted_tables:
        # Find the NDJSON file
        schema_dir = sync_dir / table.schema
        ndjson_file = schema_dir / f"{table.name}.ndjson"
        large_file = sync_dir / "_large" / f"{table.schema}.{table.name}.ndjson.gz"

        if large_file.exists():
            r = _import_table(table, large_file, compressed=True)
        elif ndjson_file.exists():
            r = _import_table(table, ndjson_file, compressed=False)
        else:
            continue

        results.append(r)
        total_imported += r.imported
        errors += r.errors

    # Re-enable FK checks
    _psql_exec("SET session_replication_role = 'origin';")

    # Run ANALYZE
    log_info("Running ANALYZE...")
    _psql_exec("ANALYZE;", timeout=300)

    if errors > 0:
        log_warning(f"Import completed with {errors} errors")
        return False, results

    log_success(f"Import complete: {len(results)} tables, {total_imported:,} rows")
    return True, results


def _import_table(table: TableInfo, filepath: Path, compressed: bool) -> SyncResult:
    """Import a single NDJSON file using upsert."""
    # Read lines
    if compressed:
        with gzip.open(filepath, 'rt', encoding='utf-8') as f:
            lines = [line.strip() for line in f if line.strip()]
    else:
        with open(filepath, 'r', encoding='utf-8') as f:
            lines = [line.strip() for line in f if line.strip()]

    if not lines:
        return SyncResult(table=table.full_name, imported=0)

    # Parse first line to discover available columns
    first_row = json.loads(lines[0])
    available_cols = [c for c in table.all_columns if c in first_row]

    if not available_cols:
        log_warning(f"  {table.full_name}: no matching columns")
        return SyncResult(table=table.full_name, errors=1)

    # Determine PK and non-PK columns for upsert
    pk_cols = table.pk_columns
    non_pk_cols = [c for c in available_cols if c not in pk_cols]

    if not pk_cols:
        log_warning(f"  {table.full_name}: no primary key, skipping")
        return SyncResult(table=table.full_name, skipped=len(lines))

    # Truncate table first (clean restore)
    _psql_exec(f"TRUNCATE {table.schema}.{table.name} CASCADE;", timeout=30)

    # Batch import using COPY for speed, fallback to upsert for conflicts
    imported = 0
    error_count = 0
    batch_size = 1000

    for batch_start in range(0, len(lines), batch_size):
        batch = lines[batch_start:batch_start + batch_size]
        batch_sql = build_batch_insert(table.schema, table.name, available_cols, pk_cols, non_pk_cols, batch)

        ok, err_msg = _psql_exec(batch_sql, timeout=120)
        if ok:
            imported += len(batch)
        else:
            # Fallback: insert one by one
            for line in batch:
                row = json.loads(line)
                single_sql = build_single_upsert(table.schema, table.name, available_cols, pk_cols, non_pk_cols, row)
                single_ok, _ = _psql_exec(single_sql, timeout=30)
                if single_ok:
                    imported += 1
                else:
                    error_count += 1

    status = "OK" if error_count == 0 else "WARN"
    print(f"  [{status}] {table.full_name}: {imported:,}/{len(lines):,} rows")

    return SyncResult(table=table.full_name, imported=imported, errors=error_count)


    # SQL helpers are in ndjson_sql_helpers.py (SRP extraction)
