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
import shutil
import socket
import subprocess
import tempfile
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
        DatabaseConfig.psql_cmd(interactive=True) + ["-q", "-v", "ON_ERROR_STOP=1"],
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
        DatabaseConfig.psql_cmd() + ["-q", "-t", "-A", "-F", "\t", "-c", sql],
        capture_output=True, text=True, timeout=timeout,
        encoding='utf-8', errors='replace',
    )
    if result.returncode != 0:
        return False, result.stderr[:300]
    return True, result.stdout.strip()


def discover_tables() -> List[TableInfo]:
    """Discover all syncable tables with columns and PKs (dynamic, like Extensions)."""
    from omni_build.ndjson_sql_helpers import validate_identifier

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

    if not tables:
        return []

    # Bug #17: Batch discovery queries to avoid N+1
    schema_list = ','.join(f"'{s}'" for s, _ in tables)
    table_list = ','.join(f"'{n}'" for _, n in tables)

    # Columns query (batched)
    ok, col_output = _psql_query(f"""
        SELECT table_schema, table_name,
               string_agg(column_name, ',' ORDER BY ordinal_position)
        FROM information_schema.columns
        WHERE table_schema IN ({schema_list}) AND table_name IN ({table_list})
        GROUP BY table_schema, table_name;
    """)
    col_map: Dict[str, List[str]] = {}
    if ok:
        for line in col_output.split('\n'):
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) >= 3:
                key = f"{parts[0].strip()}.{parts[1].strip()}"
                col_map[key] = [c.strip() for c in parts[2].split(',') if c.strip()]

    # PK query (batched)
    ok, pk_output = _psql_query(f"""
        SELECT n.nspname, c.relname,
               string_agg(a.attname, ',' ORDER BY array_position(i.indkey, a.attnum))
        FROM pg_index i
        JOIN pg_class c ON c.oid = i.indrelid
        JOIN pg_namespace n ON n.oid = c.relnamespace
        JOIN pg_attribute a ON a.attrelid = i.indrelid AND a.attnum = ANY(i.indkey)
        WHERE i.indisprimary
          AND n.nspname IN ({schema_list})
          AND c.relname IN ({table_list})
        GROUP BY n.nspname, c.relname;
    """)
    pk_map: Dict[str, List[str]] = {}
    if ok:
        for line in pk_output.split('\n'):
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) >= 3:
                key = f"{parts[0].strip()}.{parts[1].strip()}"
                pk_map[key] = [c.strip() for c in parts[2].split(',') if c.strip()]

    # Row count query (batched)
    ok, count_output = _psql_query(f"""
        SELECT schemaname, relname, n_live_tup
        FROM pg_stat_user_tables
        WHERE schemaname IN ({schema_list}) AND relname IN ({table_list});
    """)
    count_map: Dict[str, int] = {}
    if ok:
        for line in count_output.split('\n'):
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) >= 3:
                key = f"{parts[0].strip()}.{parts[1].strip()}"
                val = parts[2].strip()
                count_map[key] = int(val) if val.isdigit() else 0

    # Bug #12: FK query — removed tc.table_schema = ccu.table_schema filter
    # to detect cross-schema FKs (e.g., tenant_x.orders -> system.users)
    ok, fk_output = _psql_query(f"""
        SELECT tc.table_schema, tc.table_name,
               string_agg(DISTINCT ccu.table_schema || '.' || ccu.table_name, ',')
        FROM information_schema.table_constraints tc
        JOIN information_schema.constraint_column_usage ccu
            ON tc.constraint_name = ccu.constraint_name
        WHERE tc.constraint_type = 'FOREIGN KEY'
          AND tc.table_schema IN ({schema_list})
          AND tc.table_name IN ({table_list})
          AND NOT (ccu.table_schema = tc.table_schema AND ccu.table_name = tc.table_name)
        GROUP BY tc.table_schema, tc.table_name;
    """)
    fk_map: Dict[str, List[str]] = {}
    if ok:
        for line in fk_output.split('\n'):
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) >= 3:
                key = f"{parts[0].strip()}.{parts[1].strip()}"
                fk_map[key] = [d.strip() for d in parts[2].split(',') if d.strip()]

    # Bug #11: Validate identifiers before building result
    result = []
    for schema, name in tables:
        if name.startswith('mv_'):
            continue
        key = f"{schema}.{name}"
        columns = col_map.get(key, [])
        if not columns:
            continue
        # Validate schema/table names (SQL injection prevention)
        try:
            validate_identifier(schema)
            validate_identifier(name)
        except ValueError:
            log_warning(f"  Skipping invalid identifier: {key}")
            continue
        pk_cols = pk_map.get(key, [])
        row_count = count_map.get(key, 0)
        deps = fk_map.get(key, [])
        result.append(TableInfo(
            schema=schema, name=name,
            pk_columns=pk_cols, all_columns=columns,
            row_count=row_count, deps=deps,
        ))

    return result


def _topo_sort(tables: List[TableInfo]) -> List[TableInfo]:
    """Topological sort by FK dependencies (parents first).

    Detects circular FK dependencies and breaks cycles with a warning.
    """
    by_name: Dict[str, TableInfo] = {t.name: t for t in tables}
    visited: set[str] = set()
    in_progress: set[str] = set()  # Cycle detection: nodes currently being visited
    sorted_tables: List[TableInfo] = []

    def visit(name: str):
        if name in visited:
            return
        if name in in_progress:
            # Circular dependency detected — break the cycle
            log_warning(f"Circular FK dependency detected involving '{name}', breaking cycle")
            return
        in_progress.add(name)
        t = by_name.get(name)
        if not t:
            in_progress.discard(name)
            return
        for dep in t.deps:
            visit(dep)
        in_progress.discard(name)
        visited.add(name)
        sorted_tables.append(t)

    for t in tables:
        visit(t.name)

    # Add any tables not reached by topo sort (no deps, not referenced)
    for t in tables:
        if t not in sorted_tables:
            sorted_tables.append(t)

    return sorted_tables


def _cleanup_orphan_files(sync_dir: Path, exported_keys: set[str]) -> int:
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

    # Bug #18: Skip tables without PK (can't be imported via upsert)
    exportable = []
    for t in tables:
        if not t.pk_columns:
            log_warning(f"  Skipping {t.full_name}: no primary key (cannot upsert)")
        else:
            exportable.append(t)
    tables = exportable

    log_info(f"Found {len(tables)} tables to export")

    results = []
    total_rows = 0

    # Group by schema
    schemas: dict[str, list[TableInfo]] = {}
    for t in tables:
        schemas.setdefault(t.schema, []).append(t)

    # Track exported table names for orphan cleanup
    exported_table_keys: set[str] = set()

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
        "machine": socket.gethostname(),
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
    """Export a single table to NDJSON (delegates to ndjson_table_ops).

    Each table is exported within a REPEATABLE READ transaction via _psql_exec
    (stdin-based) to ensure per-table consistency.
    """
    from omni_build.ndjson_table_ops import export_table

    def _rr_exec(sql: str, timeout: int) -> Tuple[bool, str]:
        """Wrap SQL in REPEATABLE READ transaction for consistent per-table export."""
        wrapped = (
            "BEGIN TRANSACTION ISOLATION LEVEL REPEATABLE READ;\n"
            f"{sql}\n"
            "COMMIT;"
        )
        return _psql_exec(wrapped, timeout)

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
        psql_exec_fn=_rr_exec,
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


def _create_pre_import_backup(sorted_tables: List[TableInfo], sync_dir: Path) -> Optional[Path]:
    """Create a temporary backup of tables that will be truncated.

    Uses pg_dump per-table to capture current state. Returns the temp
    directory path on success, or None if backup is skipped (empty tables).
    """
    backup_dir = Path(tempfile.mkdtemp(prefix='ndjson_import_backup_'))
    tables_backed_up = 0

    for table in sorted_tables:
        schema_dir = sync_dir / table.schema
        ndjson_file = schema_dir / f"{table.name}.ndjson"
        large_file = sync_dir / "_large" / f"{table.schema}.{table.name}.ndjson.gz"
        if not (ndjson_file.exists() or large_file.exists()):
            continue

        # Dump table data using pg_dump --data-only for this table
        dump_file = backup_dir / f"{table.schema}.{table.name}.sql"
        dump_cmd = DatabaseConfig.docker_exec_prefix() + [
            "pg_dump", "-U", DatabaseConfig.USER,
            "--data-only", "--no-owner", "--no-privileges",
            "-t", f"{table.schema}.{table.name}",
            DatabaseConfig.DATABASE,
        ]
        try:
            result = subprocess.run(
                dump_cmd, capture_output=True, text=True,
                timeout=120, encoding='utf-8', errors='replace',
            )
            if result.returncode == 0 and result.stdout.strip():
                dump_file.write_text(result.stdout, encoding='utf-8')
                tables_backed_up += 1
        except (subprocess.TimeoutExpired, OSError) as e:
            log_warning(f"  Backup of {table.full_name} failed: {e}")

    if tables_backed_up == 0:
        shutil.rmtree(backup_dir, ignore_errors=True)
        return None

    log_info(f"Pre-import backup created: {tables_backed_up} tables → {backup_dir}")
    return backup_dir


def _restore_from_backup(backup_dir: Path) -> bool:
    """Restore tables from pre-import backup after a failed import.

    Replays the pg_dump SQL files to restore data to pre-truncate state.
    """
    log_warning("Import failed — restoring from pre-import backup...")
    restore_errors = 0

    for dump_file in sorted(backup_dir.glob('*.sql')):
        # Filename: schema.table.sql
        parts = dump_file.stem.split('.', 1)
        if len(parts) != 2:
            continue
        table_ref = f"{parts[0]}.{parts[1]}"
        sql = dump_file.read_text(encoding='utf-8')
        if not sql.strip():
            continue

        # TRUNCATE first to avoid duplicates from partial import, then replay dump
        restore_sql = (
            "SET session_replication_role = 'replica';\n"
            f"TRUNCATE {table_ref};\n"
            + sql
        )
        ok, err = _psql_exec(restore_sql, timeout=120)
        if ok:
            log_info(f"  Restored: {table_ref}")
        else:
            log_error(f"  Failed to restore {table_ref}: {err}")
            restore_errors += 1

    if restore_errors > 0:
        log_error(
            f"Restore completed with {restore_errors} error(s). "
            f"Backup files preserved at: {backup_dir}"
        )
        return False

    log_success("Pre-import backup restored successfully")
    return True


def _cleanup_backup(backup_dir: Optional[Path]) -> None:
    """Remove temporary backup directory."""
    if backup_dir and backup_dir.exists():
        shutil.rmtree(backup_dir, ignore_errors=True)


def _reset_sequences(tables: List[TableInfo]) -> None:
    """Reset ALL sequences in each schema to max value of their linked column."""
    log_info("Resetting sequences...")
    reset_schemas: set[str] = set()
    for table in tables:
        reset_schemas.add(table.schema)

    for schema in sorted(reset_schemas):
        # Query pg_sequences to find ALL sequences in this schema
        ok, output = _psql_query(
            f"SELECT sequencename FROM pg_sequences "
            f"WHERE schemaname = '{schema}';"
        )
        if not ok or not output.strip():
            continue

        for line in output.split('\n'):
            seq_name = line.strip()
            if not seq_name:
                continue
            # Find the column this sequence is linked to via pg_depend + pg_attribute
            ok2, col_info = _psql_query(
                f"SELECT a.attname, c.relname, n.nspname "
                f"FROM pg_depend d "
                f"JOIN pg_class s ON s.oid = d.objid "
                f"JOIN pg_class c ON c.oid = d.refobjid "
                f"JOIN pg_namespace n ON n.oid = c.relnamespace "
                f"JOIN pg_attribute a ON a.attrelid = c.oid AND a.attnum = d.refobjsubid "
                f"WHERE s.relkind = 'S' "
                f"AND s.relnamespace = (SELECT oid FROM pg_namespace WHERE nspname = '{schema}') "
                f"AND s.relname = '{seq_name}' "
                f"AND d.deptype = 'a' "
                f"LIMIT 1;"
            )
            if not ok2 or not col_info.strip():
                continue
            parts = col_info.strip().split('\t')
            if len(parts) < 3:
                continue
            col_name, table_name, table_schema = parts[0], parts[1], parts[2]
            # Use quoted identifiers for safety (values from pg_catalog)
            # setval(seq, max_val, true) means value already used
            # When table is empty (max=0), pass false so nextval returns 1
            q_schema = f'"{table_schema}"'
            q_table = f'"{table_name}"'
            q_col = f'"{col_name}"'
            q_seq = f'"{schema}"."{seq_name}"'
            sql = (
                f"SELECT setval('{q_seq}', "
                f"COALESCE((SELECT MAX({q_col}) FROM {q_schema}.{q_table}), 0), "
                f"COALESCE((SELECT MAX({q_col}) FROM {q_schema}.{q_table}), 0) > 0);"
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

    # Bug #19: Validate manifest version
    manifest_version = manifest.get('version')
    if manifest_version != 2:
        log_error(f"Incompatible manifest version: {manifest_version} (expected 2)")
        return False, []

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

    # Pre-import safety: backup current data before destructive truncate
    log_info("Creating pre-import backup...")
    backup_dir = _create_pre_import_backup(sorted_tables, sync_dir)

    try:
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

    except Exception as exc:
        log_error(f"Import failed with exception: {exc}")
        if backup_dir:
            _restore_from_backup(backup_dir)
        else:
            log_warning("No pre-import backup available (tables were empty)")
        return False, []

    # Import succeeded — cleanup backup
    _cleanup_backup(backup_dir)

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
    # Query NOT NULL columns (excluding PK and columns with defaults)
    not_null_cols: list[str] = []
    if table.pk_columns:
        pk_in = ", ".join(f"'{c}'" for c in table.pk_columns)
        nn_sql = (
            f"SELECT column_name FROM information_schema.columns "
            f"WHERE table_schema = '{table.schema}' AND table_name = '{table.name}' "
            f"AND is_nullable = 'NO' "
            f"AND column_default IS NULL "
            f"AND column_name NOT IN ({pk_in})"
        )
        ok, nn_output = _psql_query(nn_sql)
        if ok and nn_output.strip():
            not_null_cols = [c.strip() for c in nn_output.split('\n') if c.strip()]

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
        not_null_columns=not_null_cols,
    )
    return SyncResult(table=table.full_name, imported=imported, skipped=skipped, errors=errors)
