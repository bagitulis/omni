"""Unit tests for _cleanup_orphan_files stale large file handling."""
import gzip
import json
from pathlib import Path
from unittest.mock import patch

import pytest

from omni_build.ndjson_sync import _cleanup_orphan_files, _gzip_has_content


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _write_empty_gz(path: Path) -> None:
    """Write an empty gzip file (header + CRC + size only)."""
    path.parent.mkdir(parents=True, exist_ok=True)
    with gzip.open(path, 'wt', encoding='utf-8') as f:
        pass  # write nothing


def _write_ndjson_gz(path: Path, lines: int = 5) -> None:
    """Write a gzip file with actual NDJSON content."""
    path.parent.mkdir(parents=True, exist_ok=True)
    with gzip.open(path, 'wt', encoding='utf-8') as f:
        for i in range(lines):
            f.write(json.dumps({"id": i, "name": f"row_{i}"}) + '\n')


def _write_manifest(sync_dir: Path, tables: dict) -> None:
    """Write a manifest.json with the given tables dict."""
    manifest = {"tables": tables, "version": 3}
    with open(sync_dir / "manifest.json", 'w', encoding='utf-8') as f:
        json.dump(manifest, f)


# ---------------------------------------------------------------------------
# Tests for _gzip_has_content
# ---------------------------------------------------------------------------

def test_gzip_has_content_returns_false_for_empty_file(tmp_path: Path) -> None:
    gz = tmp_path / "test.ndjson.gz"
    _write_empty_gz(gz)
    assert _gzip_has_content(gz) is False


def test_gzip_has_content_returns_true_for_file_with_data(tmp_path: Path) -> None:
    gz = tmp_path / "test.ndjson.gz"
    _write_ndjson_gz(gz, lines=3)
    assert _gzip_has_content(gz) is True


def test_gzip_has_content_returns_true_for_unreadable_file(tmp_path: Path) -> None:
    gz = tmp_path / "test.ndjson.gz"
    gz.parent.mkdir(parents=True, exist_ok=True)
    gz.write_bytes(b"not a gzip file at all")
    # Exception caught → returns True (safe default)
    assert _gzip_has_content(gz) is True


# ---------------------------------------------------------------------------
# Tests for _cleanup_orphan_files — existing orphan behavior (regression)
# ---------------------------------------------------------------------------

def test_cleanup_removes_orphan_large_file(tmp_path: Path) -> None:
    """Large file for a table NOT in exported_keys is removed (existing behavior)."""
    sync_dir = tmp_path / "sync"
    schema_dir = sync_dir / "public"
    schema_dir.mkdir(parents=True)
    large_dir = sync_dir / "_large"
    large_dir.mkdir()

    # Create orphan large file
    orphan = large_dir / "public.dropped_table.ndjson.gz"
    _write_ndjson_gz(orphan, lines=3)

    exported_keys = set()  # empty — table doesn't exist in export
    _write_manifest(sync_dir, {})

    removed = _cleanup_orphan_files(sync_dir, exported_keys)
    assert removed == 1
    assert not orphan.exists()


def test_cleanup_keeps_large_file_for_table_with_data(tmp_path: Path) -> None:
    """Large file for a table with non-zero rows is NOT removed."""
    sync_dir = tmp_path / "sync"
    large_dir = sync_dir / "_large"
    large_dir.mkdir(parents=True)

    large_file = large_dir / "public.big_table.ndjson.gz"
    _write_ndjson_gz(large_file, lines=5)

    exported_keys = {"public.big_table"}
    _write_manifest(sync_dir, {"public.big_table": 50000})

    removed = _cleanup_orphan_files(sync_dir, exported_keys)
    assert removed == 0
    assert large_file.exists()


# ---------------------------------------------------------------------------
# Tests for _cleanup_orphan_files — stale large file for 0-row tables
# ---------------------------------------------------------------------------

def test_cleanup_removes_stale_large_file_zero_rows_empty_gzip(tmp_path: Path) -> None:
    """Stale large file is removed when manifest says 0 rows, gzip is empty, plain exists."""
    sync_dir = tmp_path / "sync"
    schema_dir = sync_dir / "public"
    schema_dir.mkdir(parents=True)
    large_dir = sync_dir / "_large"
    large_dir.mkdir()

    stale_gz = large_dir / "public.tiktok_ads_creative_data.ndjson.gz"
    _write_empty_gz(stale_gz)

    # Create the plain .ndjson (0-row export creates empty file)
    plain = schema_dir / "tiktok_ads_creative_data.ndjson"
    plain.write_text("", encoding='utf-8')

    exported_keys = {"public.tiktok_ads_creative_data"}
    _write_manifest(sync_dir, {"public.tiktok_ads_creative_data": 0})

    removed = _cleanup_orphan_files(sync_dir, exported_keys)
    assert removed == 1
    assert not stale_gz.exists()
    assert plain.exists()  # plain file is NOT touched


def test_cleanup_warns_for_stale_large_file_with_data(tmp_path: Path) -> None:
    """Stale large file with actual content is NOT deleted — WARNING is logged."""
    sync_dir = tmp_path / "sync"
    schema_dir = sync_dir / "public"
    schema_dir.mkdir(parents=True)
    large_dir = sync_dir / "_large"
    large_dir.mkdir()

    stale_gz = large_dir / "public.big_table.ndjson.gz"
    _write_ndjson_gz(stale_gz, lines=10)

    plain = schema_dir / "big_table.ndjson"
    plain.write_text("", encoding='utf-8')

    exported_keys = {"public.big_table"}
    _write_manifest(sync_dir, {"public.big_table": 0})

    with patch('omni_build.ndjson_sync.log_warning') as mock_warn:
        removed = _cleanup_orphan_files(sync_dir, exported_keys)

    assert removed == 0
    assert stale_gz.exists()  # file preserved
    mock_warn.assert_called_once()
    assert "data integrity" in mock_warn.call_args[0][0]


def test_cleanup_skips_stale_check_when_no_plain_file(tmp_path: Path) -> None:
    """If no plain .ndjson file exists, stale check is skipped even with 0 rows."""
    sync_dir = tmp_path / "sync"
    large_dir = sync_dir / "_large"
    large_dir.mkdir(parents=True)

    stale_gz = large_dir / "public.orphan.ndjson.gz"
    _write_empty_gz(stale_gz)

    # No plain file exists, no schema dir
    exported_keys = {"public.orphan"}
    _write_manifest(sync_dir, {"public.orphan": 0})

    removed = _cleanup_orphan_files(sync_dir, exported_keys)
    assert removed == 0
    assert stale_gz.exists()


def test_cleanup_skips_when_no_manifest(tmp_path: Path) -> None:
    """If no manifest.json exists, stale large file check is skipped gracefully."""
    sync_dir = tmp_path / "sync"
    large_dir = sync_dir / "_large"
    large_dir.mkdir(parents=True)

    gz = large_dir / "public.table.ndjson.gz"
    _write_ndjson_gz(gz, lines=5)

    schema_dir = sync_dir / "public"
    schema_dir.mkdir()
    plain = schema_dir / "table.ndjson"
    plain.write_text("", encoding='utf-8')

    exported_keys = {"public.table"}
    # No manifest → manifest_tables empty → row_count defaults to -1 → no stale check

    removed = _cleanup_orphan_files(sync_dir, exported_keys)
    assert removed == 0
    assert gz.exists()


def test_cleanup_handles_both_orphan_and_stale(tmp_path: Path) -> None:
    """Both orphan and stale files are handled in one call."""
    sync_dir = tmp_path / "sync"
    large_dir = sync_dir / "_large"
    large_dir.mkdir(parents=True)
    schema_dir = sync_dir / "public"
    schema_dir.mkdir()

    # Orphan (table not in exported_keys)
    orphan_gz = large_dir / "public.dropped.ndjson.gz"
    _write_ndjson_gz(orphan_gz, lines=3)

    # Stale (table in exported_keys, 0 rows, empty gzip, plain exists)
    stale_gz = large_dir / "public.empty_now.ndjson.gz"
    _write_empty_gz(stale_gz)
    plain = schema_dir / "empty_now.ndjson"
    plain.write_text("", encoding='utf-8')

    exported_keys = {"public.empty_now"}
    _write_manifest(sync_dir, {"public.empty_now": 0})

    removed = _cleanup_orphan_files(sync_dir, exported_keys)
    assert removed == 2
    assert not orphan_gz.exists()
    assert not stale_gz.exists()
