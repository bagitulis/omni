"""
Backup Manifest Management module.

SRP: This module handles backup manifest operations (load, save, status).
"""
import json
from datetime import datetime
from pathlib import Path
from typing import Any, Dict, List, Optional

from omni_build.logger import log_error, log_info, log_warning


class BackupManifest:
    """
    Handles backup manifest operations.
    
    Features:
    - Load/save manifest JSON
    - Checksum and row count tracking
    - Status reporting
    """
    
    def __init__(self, backup_dir: Path) -> None:
        self.backup_dir = backup_dir
        self.manifest_file = backup_dir / "manifest.json"
    
    def load(self) -> Optional[Dict[str, Any]]:
        """Load previous backup manifest if exists."""
        if not self.manifest_file.exists():
            return None
        
        try:
            with open(self.manifest_file, 'r', encoding='utf-8-sig') as f:
                return json.load(f)
        except Exception:
            return None
    
    def _get_previous_checksum(self, prev_manifest: Optional[Dict[str, Any]], schema: str, table: str) -> str:
        """Get checksum from previous manifest if available."""
        if not prev_manifest:
            return ""

        schema_data = prev_manifest.get('schemas', {}).get(schema, {})
        for prev_table in schema_data.get('tables', []):
            if prev_table.get('name') == table:
                checksum = prev_table.get('checksum', '')
                return checksum if isinstance(checksum, str) else ""

        return ""

    def _get_previous_rows(self, prev_manifest: Optional[Dict[str, Any]], schema: str, table: str) -> int:
        """Get row count from previous manifest if available."""
        if not prev_manifest:
            return -1

        schema_data = prev_manifest.get('schemas', {}).get(schema, {})
        for prev_table in schema_data.get('tables', []):
            if prev_table.get('name') == table:
                rows = prev_table.get('rows', -1)
                return rows if isinstance(rows, int) and rows >= 0 else -1

        return -1

    def _is_sha256_checksum(self, checksum: str) -> bool:
        """Validate checksum format as sha256 hex string."""
        if len(checksum) != 64:
            return False
        return all(c in "0123456789abcdef" for c in checksum.lower())

    def save_manifest(
        self,
        tables: List[Dict[str, Any]],
        prev_manifest: Optional[Dict[str, Any]] = None,
        generated_checksums: Optional[Dict[str, str]] = None,
        generated_rows: Optional[Dict[str, int]] = None,
    ) -> bool:
        """Save backup manifest."""
        schemas: Dict[str, Dict[str, List[Dict[str, Any]]]] = {}
        generated_checksums = generated_checksums or {}
        generated_rows = generated_rows or {}
        missing_checksums: List[str] = []
        missing_rows: List[str] = []
        
        for table in tables:
            schema = table['schema']
            if schema not in schemas:
                schemas[schema] = {'tables': []}

            table_name = table['table']
            table_key = f"{schema}.{table_name}"

            row_count = generated_rows.get(table_key, -1)
            if row_count < 0:
                row_count = self._get_previous_rows(prev_manifest, schema, table_name)
            if row_count < 0:
                source_rows = table.get('rows', 0)
                row_count = source_rows if isinstance(source_rows, int) and source_rows >= 0 else 0
                if row_count > 0:
                    missing_rows.append(table_key)

            checksum = generated_checksums.get(f"{schema}.{table_name}", "")
            if not checksum:
                checksum = self._get_previous_checksum(prev_manifest, schema, table_name)

            if not checksum:
                if row_count > 0:
                    missing_checksums.append(f"{schema}.{table_name}")
                checksum = "zero-rows"

            schemas[schema]['tables'].append({
                'name': table_name,
                'rows': row_count,
                'checksum': checksum,
                'change_vector': table.get('change_vector', ''),
            })

        if missing_rows:
            log_warning(
                "Manifest rows fallback to analyzed counts for table(s): "
                + ", ".join(missing_rows[:5])
            )

        if missing_checksums:
            log_error(
                "Cannot save manifest: missing checksum for non-empty table(s): "
                + ", ".join(missing_checksums[:5])
            )
            return False
        
        manifest = {
            'version': 4,
            'exported_at': datetime.now().strftime("%Y-%m-%d %H:%M:%S"),
            'schemas': schemas
        }
        
        try:
            with open(self.manifest_file, 'w', encoding='utf-8') as f:
                json.dump(manifest, f, indent=2)
            return True
        except Exception:
            return False
    
    def print_status(self) -> None:
        """Print backup status information."""
        info = self.load()
        
        if not info:
            log_warning("No backup found in backups/")
            log_info("Run: python build.py backup")
            return
        
        log_info("Backup Status:")
        log_info(f"  Last backup: {info.get('exported_at', 'Unknown')}")
        log_info(f"  Version: {info.get('version', 'Unknown')}")
        
        schemas = info.get('schemas', {})
        total_tables = sum(len(s.get('tables', [])) for s in schemas.values())
        total_rows = sum(t.get('rows', 0) for s in schemas.values() for t in s.get('tables', []))
        
        log_info(f"  Schemas: {len(schemas)}")
        log_info(f"  Total tables: {total_tables}")
        log_info(f"  Total rows: {total_rows:,}")
