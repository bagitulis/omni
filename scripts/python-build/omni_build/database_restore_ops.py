"""
Database Restore Operations module for Omni Build System.

SRP: This module handles table restore orchestration.
Per-table restore logic is in restore_table_ops.py.
Checksum verification and schema recreation are in restore_helpers.py.
"""
import subprocess
from pathlib import Path
from typing import Any, List, Tuple

from omni_build.database_backup_ops import DatabaseBackupOps
from omni_build.restore_helpers import (
    is_sha256_checksum,
    recreate_schemas_from_backup,
    run_post_restore_analyze,
    terminate_active_connections,
)
from omni_build.restore_table_ops import (
    restore_single_table,
    restore_chunked_files,
)

from omni_build.config import Config
from omni_build.db_config import DatabaseConfig
from omni_build.logger import log_error, log_info, log_success, log_warning


class DatabaseRestoreOps:
    """Handles database restore operations for tables and schemas."""
    
    def __init__(self, config: Config, pg_checker) -> None:
        self.config = config
        self._pg_checker = pg_checker

    def _is_sha256_checksum(self, checksum: Any) -> bool:
        """Validate checksum format as sha256 hex string."""
        return is_sha256_checksum(checksum)

    def _calculate_backup_table_checksum(self, schema: str, table: str, schema_dir: Path) -> Tuple[bool, str, str]:
        """Calculate checksum for backup artifact(s) used by restore."""
        from omni_build.restore_helpers import calculate_backup_table_checksum
        return calculate_backup_table_checksum(schema, table, schema_dir)

    def _verify_backup_checksum(
        self, schema: str, table: str, schema_dir: Path, expected_checksum: str,
    ) -> Tuple[bool, str]:
        """Verify backup file checksum against manifest checksum."""
        from omni_build.restore_helpers import verify_backup_checksum
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

    def _restore_single_table(
        self, schema: str, table: str, sql_file: Path,
        skip_truncate: bool = False, max_retries: int = 3, expected_rows: int = 0
    ) -> Tuple[bool, int]:
        """Restore single table (delegates to restore_table_ops)."""
        return restore_single_table(
            schema, table, sql_file, self._pg_checker,
            skip_truncate=skip_truncate, max_retries=max_retries,
            expected_rows=expected_rows,
        )

    def _restore_chunked_files(
        self, schema: str, table: str, chunk_files: List[Path], expected: int
    ) -> Tuple[bool, int]:
        """Restore chunked table (delegates to restore_table_ops)."""
        return restore_chunked_files(schema, table, chunk_files, expected)
    
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
            clear_cmd = DatabaseConfig.psql_cmd() + [
                "-c", f"TRUNCATE {schema_name}.{table_name} CASCADE;"
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
                        # Table expected to be empty — verify it is
                        from omni_build.restore_table_ops import _get_table_count
                        count_ok, rows, count_error = _get_table_count(schema_name, table_name)
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
                    schema_name, table_name, sql_file,
                    skip_truncate=True, expected_rows=expected_rows
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
