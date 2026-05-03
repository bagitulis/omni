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
import json
import os
import subprocess
from datetime import datetime
from pathlib import Path
from typing import Dict, List, Optional, Tuple

from omni_build.db_config import DatabaseConfig
from omni_build.logger import log_error, log_info, log_success, log_warning
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
        DatabaseConfig.psql_cmd(interactive=True) + ["-v", "ON_ERROR_STOP=1"],
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
        DatabaseConfig.psql_cmd() + ["-t", "-A", "-F", "\t", "-c", sql],
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


def _cleanup_orphan_files(sync_dir: Path, exported_keys: set) -> int:
    """Remove NDJSON files for tables that no longer exist in the database.

    After export, any .ndjson or .ndjson.gz file that doesn't correspond to
    a currently-exported table is orphaned (table was dropped/renamed).
    Removing these prevents stale data from being imported on another machine.
    """
    removed = 0

    # Check schema directories for orphan .ndjson files
    for schema_dir in sync_dir.iterdir():
        if not schema_dir.is_dir():
            continue
        # Skip special directories
        if schema_dir.name.startswith('_') or schema_dir.name == 'manifest.json':
            continue

        schema_name = schema_dir.name
        for ndjson_file in list(schema_dir.glob('*.ndjson')):
            table_name = ndjson_file.stem  # e.g. 'orders' from 'orders.ndjson'
            key = f"{schema_name}.{table_name}"
            if key not in exported_keys:
                ndjson_file.unlink()
                removed += 1
                log_info(f"  Removed orphan: {schema_name}/{ndjson_file.name}")

    # Check _large directory for orphan compressed files
    large_dir = sync_dir / '_large'
    if large_dir.exists():
        for gz_file in list(large_dir.glob('*.ndjson.gz')):
            # Filename format: schema.table.ndjson.gz
            parts = gz_file.stem.replace('.ndjson', '').split('.', 1)
            if len(parts) == 2:
                key = f"{parts[0]}.{parts[1]}"
                if key not in exported_keys:
                    gz_file.unlink()
                    removed += 1
                    log_info(f"  Removed orphan: _large/{gz_file.name}")

    # Remove empty schema directories
    for schema_dir in list(sync_dir.iterdir()):
        if schema_dir.is_dir() and not schema_dir.name.startswith('_'):
            remaining = list(schema_dir.iterdir())
            if not remaining:
                schema_dir.rmdir()
                log_info(f"  Removed empty directory: {schema_dir.name}/")

    return removed

def export_all(project_root: Path) -> Tuple[bool, List[SyncResult]]:
    """Export all tables to NDJSON files (like Extensions ExportAll)."""
    sync_dir = project_root / SYNC_DIR
    sync_dir.mkdir(parents=True, exist_ok=True)

    # Run ANALYZE first for accurate row counts
    _psql_exec("ANALYZE;", timeout=300)

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

    # Track exported table names for orphan cleanup
    exported_table_keys: set = set()

    for schema, schema_tables in schemas.items():
        schema_dir = sync_dir / schema
        schema_dir.mkdir(parents=True, exist_ok=True)

        for table in schema_tables:
            r = _export_table(table, schema_dir)
            results.append(r)
            total_rows += r.exported
            exported_table_keys.add(f"{schema}.{table.name}")

    # Cleanup orphan NDJSON files (tables that no longer exist in DB)
    orphans_removed = _cleanup_orphan_files(sync_dir, exported_table_keys)
    if orphans_removed > 0:
        log_info(f"Cleaned up {orphans_removed} orphan file(s) from previous exports")

    # Write manifest
    manifest = {
        "exported_at": datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
        "machine": subprocess.run(["hostname"], capture_output=True, text=True).stdout.strip(),
        "version": 2,
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
    """Export a single table to NDJSON (delegates to ndjson_table_ops)."""
    from omni_build.ndjson_table_ops import export_table
    exported, errors = export_table(
        schema=table.schema,
        name=table.name,
        full_name=table.full_name,
        all_columns=table.all_columns,
        pk_columns=table.pk_columns,
        is_large=table.is_large,
        dest_dir=dest_dir,
        large_file_size_bytes=LARGE_FILE_SIZE_BYTES,
        psql_query_fn=_psql_query,
    )
    return SyncResult(table=table.full_name, exported=exported, errors=errors)

def _check_lfs_pointers(sync_dir: Path) -> bool:
    """Check if LFS files are actual data (not unresolved pointers)."""
    large_dir = sync_dir / "_large"
    if not large_dir.exists():
        return True
    for f in large_dir.glob("*.ndjson.gz"):
        with open(f, 'rb') as fh:
            header = fh.read(30)
            if header.startswith(b'version https://git-lfs'):
                log_error(f"Git LFS pointer detected: {f.name}")
                log_error("Run: git lfs install && git lfs pull")
                return False
    return True


def _reset_sequences(tables: List[TableInfo]) -> None:
    """Reset all sequences to max(pk) + 1 after import."""
    log_info("Resetting sequences...")
    for table in tables:
        if len(table.pk_columns) != 1:
            continue
        pk = table.pk_columns[0]
        # Only reset if PK is likely a serial/identity column
        sql = (
            f"DO $$ BEGIN "
            f"IF EXISTS (SELECT pg_get_serial_sequence('{table.schema}.{table.name}', '{pk}')) THEN "
            f"PERFORM setval("
            f"pg_get_serial_sequence('{table.schema}.{table.name}', '{pk}'), "
            f"COALESCE((SELECT MAX(\"{pk}\") FROM {table.schema}.{table.name}), 0) + 1, "
            f"false); "
            f"END IF; "
            f"END $$;"
        )
        _psql_exec(sql, timeout=10)


def import_all(project_root: Path) -> Tuple[bool, List[SyncResult]]:
    """Import all NDJSON files into PostgreSQL (like Extensions ImportAll)."""
    sync_dir = project_root / SYNC_DIR

    if not sync_dir.exists():
        return False, []

    manifest_path = sync_dir / "manifest.json"
    if not manifest_path.exists():
        log_error("No manifest.json found in sync directory")
        return False, []

    # Check for unresolved LFS pointers
    if not _check_lfs_pointers(sync_dir):
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

    # Build lookup of known tables for orphan detection
    known_table_keys = {f"{t.schema}.{t.name}" for t in tables}

    # Detect NDJSON files that have no matching table in DB (orphan warning)
    manifest_tables = manifest.get('tables', {})
    orphan_files = [k for k in manifest_tables if k not in known_table_keys]
    if orphan_files:
        log_warning(
            f"{len(orphan_files)} table(s) in sync files have no matching DB table "
            f"(skipped): {', '.join(orphan_files[:5])}"
        )

    # Sort by dependencies (parents first)
    sorted_tables = _topo_sort(tables)

    # Phase 1: TRUNCATE all tables in REVERSE order (children first)
    # This avoids CASCADE which could destroy data in unrelated tables
    log_info("Truncating tables (reverse dependency order)...")
    for table in reversed(sorted_tables):
        schema_dir = sync_dir / table.schema
        ndjson_file = schema_dir / f"{table.name}.ndjson"
        large_file = sync_dir / "_large" / f"{table.schema}.{table.name}.ndjson.gz"
        if ndjson_file.exists() or large_file.exists():
            # Use session_replication_role to skip FK checks during truncate
            _psql_exec(
                f"SET session_replication_role = 'replica';\n"
                f"TRUNCATE {table.schema}.{table.name};",
                timeout=30
            )

    # Phase 2: Import all tables in dependency order (parents first)
    results = []
    total_imported = 0
    errors = 0

    for table in sorted_tables:
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

    # Phase 3: Reset sequences to max(pk) + 1
    _reset_sequences(sorted_tables)

    # Phase 4: Run ANALYZE for query planner
    log_info("Running ANALYZE...")
    _psql_exec("ANALYZE;", timeout=300)

    if errors > 0:
        log_warning(f"Import completed with {errors} errors")
        return False, results

    log_success(f"Import complete: {len(results)} tables, {total_imported:,} rows")
    return True, results


def _import_table(table: TableInfo, filepath: Path, compressed: bool) -> SyncResult:
    """Import a single NDJSON file (delegates to ndjson_table_ops)."""
    from omni_build.ndjson_table_ops import import_table
    imported, skipped, errors = import_table(
        schema=table.schema,
        name=table.name,
        full_name=table.full_name,
        all_columns=table.all_columns,
        pk_columns=table.pk_columns,
        filepath=filepath,
        compressed=compressed,
        psql_exec_fn=_psql_exec,
    )
    return SyncResult(table=table.full_name, imported=imported, skipped=skipped, errors=errors)
