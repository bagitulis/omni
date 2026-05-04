"""Unit tests for NDJSON sync import guardrails."""
import json
from pathlib import Path

import pytest

from omni_build import ndjson_sync
from omni_build.ndjson_sync import TableInfo, import_all


def test_import_all_accepts_manifest_version_one_before_db_work(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    sync_dir = tmp_path / "backups" / "sync"
    sync_dir.mkdir(parents=True)
    _ = (sync_dir / "manifest.json").write_text(
        json.dumps({"version": 1, "tables": {}, "total_rows": 0}),
        encoding="utf-8",
    )

    discover_called = False

    def fake_discover_tables() -> list[TableInfo]:
        nonlocal discover_called
        discover_called = True
        return []

    def fake_materialize_lfs_pointers(project_root: Path, sync_root: Path) -> bool:
        _ = (project_root, sync_root)
        return True

    monkeypatch.setattr(ndjson_sync, "materialize_lfs_pointers", fake_materialize_lfs_pointers)
    monkeypatch.setattr(ndjson_sync, "discover_tables", fake_discover_tables)

    success, results = import_all(tmp_path)

    assert success is False
    assert results == []
    assert discover_called is True


def test_import_all_rejects_unknown_manifest_version_before_db_work(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    sync_dir = tmp_path / "backups" / "sync"
    sync_dir.mkdir(parents=True)
    _ = (sync_dir / "manifest.json").write_text(
        json.dumps({"version": 999, "tables": {}, "total_rows": 0}),
        encoding="utf-8",
    )

    discover_called = False

    def fake_discover_tables() -> list[TableInfo]:
        nonlocal discover_called
        discover_called = True
        return []

    def fake_materialize_lfs_pointers(project_root: Path, sync_root: Path) -> bool:
        _ = (project_root, sync_root)
        return True

    monkeypatch.setattr(ndjson_sync, "materialize_lfs_pointers", fake_materialize_lfs_pointers)
    monkeypatch.setattr(ndjson_sync, "discover_tables", fake_discover_tables)

    success, results = import_all(tmp_path)

    assert success is False
    assert results == []
    assert discover_called is False
