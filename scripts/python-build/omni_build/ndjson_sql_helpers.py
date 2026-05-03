"""
SQL helper functions for NDJSON sync.

SRP: PostgreSQL literal escaping and upsert SQL generation.
Extracted from ndjson_sync.py for ~300 line compliance.
"""
import math
import json
import re
from decimal import Decimal
from typing import Any, List

# Valid PostgreSQL identifier pattern (prevents SQL injection)
_IDENT_RE = re.compile(r'^[a-zA-Z_][a-zA-Z0-9_]*$')


def validate_identifier(name: str) -> str:
    """Validate and quote a PostgreSQL identifier to prevent SQL injection."""
    if not _IDENT_RE.match(name):
        raise ValueError(f"Invalid SQL identifier: {name!r}")
    return f'"{name}"'


def pg_literal(value: Any) -> str:
    """Convert Python value to PostgreSQL literal string.

    Handles: None, bool, int, float, Decimal, dict/list (JSON), str.
    Escapes single quotes, backslashes, and strips NULL bytes.
    """
    if value is None:
        return "NULL"
    if isinstance(value, bool):
        return "TRUE" if value else "FALSE"
    if isinstance(value, (int, Decimal)):
        return str(value)
    if isinstance(value, float):
        if math.isnan(value):
            return "'NaN'::float"
        if math.isinf(value):
            return "'Infinity'::float" if value > 0 else "'-Infinity'::float"
        return repr(value)
    if isinstance(value, (dict, list)):
        # JSON value — serialize and quote
        json_str = json.dumps(value, ensure_ascii=False)
        return _escape_string(json_str)
    # String value
    s = str(value)
    return _escape_string(s)


def _escape_string(s: str) -> str:
    """Escape a string for PostgreSQL, handling backslashes and NULL bytes."""
    # Strip NULL bytes (PostgreSQL cannot store them in text)
    s = s.replace('\x00', '')
    # Escape single quotes
    escaped = s.replace("'", "''")
    # If string contains backslashes, use E'' escape syntax
    if '\\' in s:
        escaped = escaped.replace('\\', '\\\\')
        return "E'" + escaped + "'"
    return "'" + escaped + "'"


def build_batch_insert(schema: str, table: str, cols: List[str],
                       pk_cols: List[str], non_pk_cols: List[str],
                       lines: List[str]) -> str:
    """Build batch INSERT ... ON CONFLICT DO UPDATE from NDJSON lines.

    Prepends SET session_replication_role = 'replica' to disable FK triggers
    for the duration of this psql session.
    """
    col_list = ', '.join(validate_identifier(c) for c in cols)
    conflict_cols = ', '.join(validate_identifier(c) for c in pk_cols)

    values_parts = []
    for line in lines:
        row = json.loads(line, parse_float=Decimal)
        vals = [pg_literal(row.get(col)) for col in cols]
        values_parts.append(f"({', '.join(vals)})")

    values_str = ',\n'.join(values_parts)

    # Prepend FK disable for this session
    prefix = "SET session_replication_role = 'replica';\n"

    if non_pk_cols:
        updates = ', '.join(
            f'{validate_identifier(c)} = EXCLUDED.{validate_identifier(c)}'
            for c in non_pk_cols
        )
        return (
            f'{prefix}'
            f'INSERT INTO {schema}.{table} ({col_list})\n'
            f'VALUES {values_str}\n'
            f'ON CONFLICT ({conflict_cols}) DO UPDATE SET {updates};'
        )
    else:
        return (
            f'{prefix}'
            f'INSERT INTO {schema}.{table} ({col_list})\n'
            f'VALUES {values_str}\n'
            f'ON CONFLICT ({conflict_cols}) DO NOTHING;'
        )


def build_single_upsert(schema: str, table: str, cols: List[str],
                         pk_cols: List[str], non_pk_cols: List[str],
                         row: dict) -> str:
    """Build single-row upsert SQL with FK disable."""
    col_list = ', '.join(validate_identifier(c) for c in cols)
    conflict_cols = ', '.join(validate_identifier(c) for c in pk_cols)
    vals = [pg_literal(row.get(c)) for c in cols]
    values_str = ', '.join(vals)

    prefix = "SET session_replication_role = 'replica';\n"

    if non_pk_cols:
        updates = ', '.join(
            f'{validate_identifier(c)} = EXCLUDED.{validate_identifier(c)}'
            for c in non_pk_cols
        )
        return (
            f'{prefix}'
            f'INSERT INTO {schema}.{table} ({col_list}) '
            f'VALUES ({values_str}) '
            f'ON CONFLICT ({conflict_cols}) DO UPDATE SET {updates};'
        )
    else:
        return (
            f'{prefix}'
            f'INSERT INTO {schema}.{table} ({col_list}) '
            f'VALUES ({values_str}) '
            f'ON CONFLICT ({conflict_cols}) DO NOTHING;'
        )
