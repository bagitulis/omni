"""
Restore helper utilities for Omni Build System.

SRP: Checksum verification and schema recreation helpers.
Extracted from database_restore_ops.py for ~300 line compliance.
"""
import gzip
import hashlib
import subprocess
from pathlib import Path
from typing import Any, Tuple

from omni_build.logger import log_error, log_info, log_success, log_warning


def update_hash_from_file(file_path: Path, hasher: "hashlib._Hash") -> None:
    """Update hasher from file bytes in chunks."""
    with open(file_path, 'rb') as f:
        while True:
            chunk = f.read(1024 * 1024)
            if not chunk:
                break
            hasher.update(chunk)


def is_sha256_checksum(checksum: Any) -> bool:
    """Validate checksum format as sha256 hex string."""
    if not isinstance(checksum, str):
        return False
    if len(checksum) != 64:
        return False
    return all(c in "0123456789abcdef" for c in checksum.lower())


def calculate_backup_table_checksum(
    schema: str, table: str, schema_dir: Path
) -> Tuple[bool, str, str]:
    """Calculate checksum for backup artifact(s) used by restore."""
    hasher = hashlib.sha256()
    table_file = schema_dir / f"{table}.sql.gz"
    chunk_files = sorted(schema_dir.glob(f"{table}.chunk*.sql.gz"))

    if table_file.exists():
        hasher.update(f"single:{schema}.{table}".encode('utf-8'))
        update_hash_from_file(table_file, hasher)
        return True, hasher.hexdigest(), ""

    if chunk_files:
        hasher.update(f"chunked:{schema}.{table}".encode('utf-8'))
        for chunk_file in chunk_files:
            hasher.update(chunk_file.name.encode('utf-8'))
            update_hash_from_file(chunk_file, hasher)
        return True, hasher.hexdigest(), ""

    return False, "", "backup artifact not found"


def verify_backup_checksum(
    schema: str,
    table: str,
    schema_dir: Path,
    expected_checksum: str,
) -> Tuple[bool, str]:
    """Verify backup file checksum against manifest checksum."""
    ok, actual_checksum, error_message = calculate_backup_table_checksum(
        schema, table, schema_dir
    )
    if not ok:
        return False, error_message

    if actual_checksum != expected_checksum:
        return False, (
            f"checksum mismatch: expected {expected_checksum[:12]}..., "
            f"actual {actual_checksum[:12]}..."
        )

    return True, ""


def recreate_schemas_from_backup(data_dir: Path, pg_checker) -> bool:
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
            if not pg_checker.ensure_healthy():
                log_error("PostgreSQL crashed - attempting recovery...")
                if not pg_checker.recover():
                    return False

            drop_cmd = [
                "docker", "exec", "omni-postgres", "psql", "-U", "omni",
                "-d", "omni_main", "-c",
                f"DROP SCHEMA IF EXISTS {schema_name} CASCADE;"
            ]
            subprocess.run(drop_cmd, capture_output=True, text=True, timeout=60)

            if not pg_checker.ensure_healthy():
                log_error(f"PostgreSQL crashed during DROP SCHEMA {schema_name}")
                if not pg_checker.recover():
                    return False
                continue

            with gzip.open(schema_file, 'rb') as f:
                sql_bytes = f.read()
            sql_content = sql_bytes.decode('utf-8')

            result = subprocess.run(
                ["docker", "exec", "-i", "omni-postgres", "psql", "-U", "omni",
                 "-d", "omni_main"],
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
            if not pg_checker.recover():
                return False
        except Exception as e:
            log_warning(f"  Failed to recreate {schema_name}: {e}")

    return True


def terminate_active_connections() -> bool:
    """Terminate active connections to prevent lock conflicts during restore."""
    try:
        result = subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
             "-d", "postgres", "-c",
             "SELECT pg_terminate_backend(pid) FROM pg_stat_activity "
             "WHERE datname = 'omni_main' AND pid <> pg_backend_pid() "
             "AND state = 'active';"],
            capture_output=True, text=True, timeout=30,
            encoding='utf-8', errors='replace',
        )
        if result.returncode == 0:
            log_info("Active connections terminated")
            return True
        return False
    except Exception:
        return False


def run_post_restore_analyze() -> bool:
    """Run ANALYZE on all restored tables to update query planner statistics."""
    try:
        log_info("Running ANALYZE on restored tables...")
        result = subprocess.run(
            ["docker", "exec", "omni-postgres", "psql", "-U", "omni",
             "-d", "omni_main", "-c", "ANALYZE;"],
            capture_output=True, text=True, timeout=300,
            encoding='utf-8', errors='replace',
        )
        if result.returncode == 0:
            log_success("ANALYZE completed — query planner statistics updated")
            return True
        else:
            log_warning(f"ANALYZE had issues: {result.stderr[:100]}")
            return False
    except Exception as e:
        log_warning(f"ANALYZE failed: {e}")
        return False
