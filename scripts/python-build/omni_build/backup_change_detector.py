"""
Backup change detection module for Omni Build System.

SRP: This module only decides which tables need export.
"""

from pathlib import Path
from typing import Any, Dict, List, Optional, Tuple

from omni_build.database_backup_ops import DatabaseBackupOps


class BackupChangeDetector:
    """Detect changed/unchanged/deleted tables for smart backup."""

    def __init__(self, backup_dir: Path) -> None:
        self._ops = DatabaseBackupOps(backup_dir)

    @staticmethod
    def _is_sha256_checksum(checksum: str) -> bool:
        if len(checksum) != 64:
            return False
        return all(c in "0123456789abcdef" for c in checksum.lower())

    @staticmethod
    def _parse_change_vector(value: Any) -> Optional[Tuple[int, int, int, int]]:
        if not isinstance(value, str):
            return None
        parts = value.split(":")
        if len(parts) != 4:
            return None
        try:
            return (
                int(parts[0]),
                int(parts[1]),
                int(parts[2]),
                int(parts[3]),
            )
        except ValueError:
            return None

    def detect_changes(
        self,
        tables: List[Dict[str, Any]],
        prev_manifest: Optional[Dict[str, Any]],
        force: bool = False,
    ) -> Tuple[List[Dict[str, Any]], List[Dict[str, Any]], List[Dict[str, Any]]]:
        to_export: List[Dict[str, Any]] = []
        unchanged: List[Dict[str, Any]] = []
        deleted: List[Dict[str, Any]] = []

        for table in tables:
            needs_export = True
            reason = "new"

            if not force and prev_manifest:
                prev_schema = prev_manifest.get("schemas", {}).get(table["schema"])
                if prev_schema:
                    prev_tables = prev_schema.get("tables", [])
                    prev_table = next((t for t in prev_tables if t["name"] == table["table"]), None)
                    if prev_table:
                        prev_rows = prev_table.get("rows")
                        if prev_rows != table["rows"]:
                            reason = f"rows: {prev_rows} -> {table['rows']}"
                        else:
                            prev_vector = self._parse_change_vector(prev_table.get("change_vector"))
                            current_vector = self._parse_change_vector(table.get("change_vector"))

                            if prev_vector and current_vector:
                                if any(cur < prev for cur, prev in zip(current_vector, prev_vector)):
                                    reason = "stats_reset_or_rollover"
                                elif current_vector == prev_vector:
                                    needs_export = False
                                    unchanged.append(table)
                                else:
                                    reason = (
                                        "change_vector: "
                                        f"{prev_table.get('change_vector')} -> {table.get('change_vector')}"
                                    )
                            else:
                                reason = "legacy_manifest_checksum_verify"

                            if needs_export and reason in (
                                "stats_reset_or_rollover",
                                "legacy_manifest_checksum_verify",
                            ):
                                prev_checksum = prev_table.get("checksum", "")
                                if isinstance(prev_checksum, str) and self._is_sha256_checksum(prev_checksum):
                                    live_checksum = self._ops.calculate_live_table_checksum(
                                        table["schema"],
                                        table["table"],
                                        table["rows"],
                                    )
                                    if live_checksum and live_checksum == prev_checksum:
                                        needs_export = False
                                        unchanged.append(table)
                                    elif live_checksum:
                                        reason = "checksum_changed"
                                    else:
                                        reason = "checksum_verify_failed"
                                else:
                                    reason = "missing_previous_checksum"

            if needs_export:
                table["reason"] = reason
                to_export.append(table)

        if prev_manifest:
            for schema_name, schema_data in prev_manifest.get("schemas", {}).items():
                for prev_table in schema_data.get("tables", []):
                    exists = any(
                        t["schema"] == schema_name and t["table"] == prev_table["name"] for t in tables
                    )
                    if not exists:
                        deleted.append({"schema": schema_name, "table": prev_table["name"]})

        return to_export, unchanged, deleted
