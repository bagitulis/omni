"""
Live checksum calculation module for Omni Build System.

SRP: This module calculates checksums from live PostgreSQL data streams.
"""
import hashlib
import subprocess
from typing import Optional

from omni_build.db_config import DatabaseConfig


class LiveChecksumCalculator:
    """Calculate checksums from live PostgreSQL data without writing backup files."""
    
    LARGE_TABLE_THRESHOLD = DatabaseConfig.LARGE_TABLE_THRESHOLD
    CHUNK_SIZE = DatabaseConfig.CHUNK_SIZE
    
    def __init__(self, get_primary_key_fn) -> None:
        """
        Initialize calculator with primary key resolver.
        
        Args:
            get_primary_key_fn: Function(schema, table) -> str that resolves primary key column
        """
        self._get_primary_key = get_primary_key_fn
    
    def _calculate_live_single_checksum(self, schema: str, table: str) -> str:
        """Calculate checksum from live PostgreSQL dump stream (single-file strategy)."""
        _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
        cmd = (
            f"pg_dump -U {_u} -d {_d} --table={schema}.{table} "
            "--data-only --no-owner --no-acl 2>/dev/null | gzip -n"
        )
        result = subprocess.run(
            DatabaseConfig.docker_exec_prefix() + ["sh", "-c", cmd],
            capture_output=True,
            timeout=300,
        )

        if result.returncode != 0 or not result.stdout:
            return ""

        hasher = hashlib.sha256()
        hasher.update(f"single:{schema}.{table}".encode("utf-8"))
        hasher.update(result.stdout)
        return hasher.hexdigest()

    def _calculate_live_chunked_checksum(self, schema: str, table: str, total_rows: int) -> str:
        """Calculate checksum from live PostgreSQL stream (chunked strategy)."""
        if total_rows <= 0:
            return ""

        order_by = self._get_primary_key(schema, table)
        hasher = hashlib.sha256()
        hasher.update(f"chunked:{schema}.{table}".encode("utf-8"))

        offset = 0
        chunk_num = 0
        while offset < total_rows:
            chunk_name = f"{table}.chunk{chunk_num:03d}.sql.gz"
            _u, _d = DatabaseConfig.USER, DatabaseConfig.DATABASE
            copy_cmd = (
                f'psql -U {_u} -d {_d} -c "COPY (SELECT * FROM {schema}.{table} '
                f'ORDER BY {order_by} LIMIT {self.CHUNK_SIZE} OFFSET {offset}) TO STDOUT" '
                '2>/dev/null | gzip -n'
            )
            result = subprocess.run(
                DatabaseConfig.docker_exec_prefix() + ["sh", "-c", copy_cmd],
                capture_output=True,
                timeout=600,
            )

            if result.returncode != 0 or not result.stdout:
                return ""

            hasher.update(chunk_name.encode("utf-8"))
            hasher.update(result.stdout)
            offset += self.CHUNK_SIZE
            chunk_num += 1

        return hasher.hexdigest()

    def calculate_live_table_checksum(self, schema: str, table: str, rows: int) -> str:
        """Calculate checksum directly from live PostgreSQL data without writing backup files."""
        try:
            if rows >= self.LARGE_TABLE_THRESHOLD:
                return self._calculate_live_chunked_checksum(schema, table, rows)
            return self._calculate_live_single_checksum(schema, table)
        except Exception:
            return ""
