"""
NDJSON Sync — Constraint management (secondary UNIQUE constraints).

Extracted from ndjson_sync.py for maintainability.
Functions here manage discovery, drop, recreate, and crash-recovery
of secondary UNIQUE constraints during import.
"""
import json
from pathlib import Path
from typing import Callable, Dict, List, Tuple

from omni_build.logger import log_error, log_info, log_success, log_warning

# Type alias for the psql helper functions passed via dependency injection
PsqlExecFn = Callable[[str, int], Tuple[bool, str]]
PsqlQueryFn = Callable[[str, int], Tuple[bool, str]]


def _discover_secondary_unique_constraints(schema_list: str, table_list: str,
                                           psql_query_fn: PsqlQueryFn
                                           ) -> Dict[str, List[Tuple[str, str]]]:
    """Discover secondary UNIQUE constraints (not PK) for all tables.

    Returns dict mapping 'schema.table' -> [(constraint_name, definition), ...].
    Uses pg_get_constraintdef() so recreation DDL is exact (handles expressions,
    partial indexes, etc.).
    """
    ok, output = psql_query_fn(
        f"SELECT n.nspname, c.relname, con.conname, "
        f"pg_get_constraintdef(con.oid) "
        f"FROM pg_constraint con "
        f"JOIN pg_class c ON c.oid = con.conrelid "
        f"JOIN pg_namespace n ON n.oid = c.relnamespace "
        f"WHERE con.contype = 'u' "
        f"AND n.nspname IN ({schema_list}) "
        f"AND c.relname IN ({table_list})",
        60
    )
    result: Dict[str, List[Tuple[str, str]]] = {}
    if ok and output.strip():
        for line in output.split('\n'):
            if not line.strip():
                continue
            parts = line.split('\t')
            if len(parts) >= 4:
                key = f"{parts[0].strip()}.{parts[1].strip()}"
                con_name = parts[2].strip()
                con_def = parts[3].strip()
                result.setdefault(key, []).append((con_name, con_def))
    return result


def _save_constraint_manifest(constraints: Dict[str, List[Tuple[str, str]]],
                              sync_dir: Path) -> None:
    """Save constraint definitions to a recovery manifest file.

    This ensures constraints can be recreated even if the process crashes
    between dropping and recreating them.
    """
    manifest = {}
    for table_key, cons in constraints.items():
        manifest[table_key] = [{"name": name, "definition": defn} for name, defn in cons]
    manifest_path = sync_dir / "_constraints_dropped.json"
    with open(manifest_path, 'w', encoding='utf-8') as f:
        json.dump(manifest, f, indent=2)


def _load_constraint_manifest(sync_dir: Path) -> Dict[str, List[Tuple[str, str]]]:
    """Load constraint definitions from recovery manifest (if exists).

    Returns empty dict if no manifest found.
    """
    manifest_path = sync_dir / "_constraints_dropped.json"
    if not manifest_path.exists():
        return {}
    try:
        with open(manifest_path, 'r', encoding='utf-8') as f:
            raw = json.load(f)
        result: Dict[str, List[Tuple[str, str]]] = {}
        for table_key, cons in raw.items():
            result[table_key] = [(c["name"], c["definition"]) for c in cons]
        return result
    except (json.JSONDecodeError, KeyError, TypeError):
        return {}


def _delete_constraint_manifest(sync_dir: Path) -> None:
    """Delete the constraint recovery manifest after successful recreation."""
    manifest_path = sync_dir / "_constraints_dropped.json"
    if manifest_path.exists():
        manifest_path.unlink()


def _recover_constraints_if_needed(sync_dir: Path,
                                   psql_exec_fn: PsqlExecFn) -> None:
    """Auto-recover constraints from a previous crashed import.

    If _constraints_dropped.json exists, it means a previous import crashed
    after dropping constraints but before recreating them. Recreate now.
    """
    saved = _load_constraint_manifest(sync_dir)
    if not saved:
        return
    total = sum(len(v) for v in saved.values())
    log_warning(
        f"Recovering {total} UNIQUE constraint(s) from previous crashed import..."
    )
    failures = _recreate_secondary_unique_constraints(saved, psql_exec_fn)
    if failures == 0:
        log_success(f"Recovered all {total} constraint(s) successfully")
        _delete_constraint_manifest(sync_dir)
    else:
        log_error(
            f"{failures} constraint(s) failed to recover. "
            f"Manual intervention may be needed."
        )


def _drop_secondary_unique_constraints(
    constraints: Dict[str, List[Tuple[str, str]]],
    psql_exec_fn: PsqlExecFn
) -> int:
    """Drop secondary UNIQUE constraints before import.

    Returns number of constraints dropped.
    """
    dropped = 0
    for table_key, cons in constraints.items():
        for con_name, _ in cons:
            ok, err = psql_exec_fn(
                f"ALTER TABLE {table_key} DROP CONSTRAINT IF EXISTS \"{con_name}\";",
                30
            )
            if ok:
                dropped += 1
            else:
                log_warning(f"  Failed to drop constraint {con_name} on {table_key}: {err}")
    return dropped


def _recreate_secondary_unique_constraints(
    constraints: Dict[str, List[Tuple[str, str]]],
    psql_exec_fn: PsqlExecFn
) -> int:
    """Recreate secondary UNIQUE constraints after import.

    Returns number of constraints that failed to recreate.
    """
    failures = 0
    for table_key, cons in constraints.items():
        for con_name, con_def in cons:
            ok, err = psql_exec_fn(
                f"ALTER TABLE {table_key} ADD CONSTRAINT \"{con_name}\" {con_def};",
                120
            )
            if not ok:
                failures += 1
                log_error(
                    f"  Failed to recreate constraint {con_name} on {table_key}: {err}"
                )
    return failures
