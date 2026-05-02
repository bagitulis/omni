"""
SQL helper functions for NDJSON sync.

SRP: PostgreSQL literal escaping and upsert SQL generation.
Extracted from ndjson_sync.py for ~300 line compliance.
"""
import json
from decimal import Decimal
from typing import Any, List


def pg_literal(value: Any) -> str:
    """Convert Python value to PostgreSQL literal string."""
    if value is None:
        return "NULL"
    if isinstance(value, bool):
        return "TRUE" if value else "FALSE"
    if isinstance(value, (int, float, Decimal)):
        return str(value)
    if isinstance(value, (dict, list)):
        # JSON value — serialize and quote
        json_str = json.dumps(value, ensure_ascii=False)
        return "'" + json_str.replace("'", "''") + "'"
    # String value — escape single quotes
    s = str(value)
    return "'" + s.replace("'", "''") + "'"


def build_batch_insert(schema: str, table: str, cols: List[str],
                       pk_cols: List[str], non_pk_cols: List[str],
                       lines: List[str]) -> str:
    """Build batch INSERT ... ON CONFLICT DO UPDATE from NDJSON lines."""
    col_list = ', '.join(f'"{c}"' for c in cols)
    conflict_cols = ', '.join(f'"{c}"' for c in pk_cols)

    values_parts = []
    for line in lines:
        row = json.loads(line)
        vals = []
        for col in cols:
            val = row.get(col)
            vals.append(pg_literal(val))
        values_parts.append(f"({', '.join(vals)})")

    values_str = ',\n'.join(values_parts)

    if non_pk_cols:
        updates = ', '.join(f'"{c}" = EXCLUDED."{c}"' for c in non_pk_cols)
        return (
            f'INSERT INTO {schema}.{table} ({col_list})\n'
            f'VALUES {values_str}\n'
            f'ON CONFLICT ({conflict_cols}) DO UPDATE SET {updates};'
        )
    else:
        return (
            f'INSERT INTO {schema}.{table} ({col_list})\n'
            f'VALUES {values_str}\n'
            f'ON CONFLICT ({conflict_cols}) DO NOTHING;'
        )


def build_single_upsert(schema: str, table: str, cols: List[str],
                        pk_cols: List[str], non_pk_cols: List[str],
                        row: dict) -> str:
    """Build single-row upsert SQL."""
    col_list = ', '.join(f'"{c}"' for c in cols)
    conflict_cols = ', '.join(f'"{c}"' for c in pk_cols)
    vals = [pg_literal(row.get(c)) for c in cols]
    values_str = ', '.join(vals)

    if non_pk_cols:
        updates = ', '.join(f'"{c}" = EXCLUDED."{c}"' for c in non_pk_cols)
        return (
            f'INSERT INTO {schema}.{table} ({col_list}) '
            f'VALUES ({values_str}) '
            f'ON CONFLICT ({conflict_cols}) DO UPDATE SET {updates};'
        )
    else:
        return (
            f'INSERT INTO {schema}.{table} ({col_list}) '
            f'VALUES ({values_str}) '
            f'ON CONFLICT ({conflict_cols}) DO NOTHING;'
        )
