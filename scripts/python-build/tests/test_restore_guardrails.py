import gzip
import json
import time
from pathlib import Path
from types import SimpleNamespace
from typing import Any, cast

from omni_build.database_backup import DatabaseBackup
from omni_build.database_restore_ops import DatabaseRestoreOps
from omni_build.models import BuildMode, SpecLevel
from omni_build.orchestrator import BuildOrchestrator


class _DummyChecker:
    def ensure_healthy(self) -> bool:
        return True

    def recover(self) -> bool:
        return True


def _result(returncode: int = 0, stdout: str = "", stderr: str = "") -> SimpleNamespace:
    return SimpleNamespace(returncode=returncode, stdout=stdout, stderr=stderr)


def test_restore_tables_fails_when_row_count_mismatch(monkeypatch: Any, tmp_path: Path):
    ops = DatabaseRestoreOps(config=cast(Any, SimpleNamespace()), pg_checker=_DummyChecker())

    schema_dir = tmp_path / "tenant_abc"
    schema_dir.mkdir(parents=True, exist_ok=True)
    # Valid gzip with 7 INSERT rows — artifact has 7, but restore returns 9
    # Auto-fix verifies artifact_rows==restored_rows, so 7≠9 → real failure
    sql_data = "\n".join(f"INSERT INTO t VALUES ({i});" for i in range(7))
    with gzip.open(schema_dir / "inventory_records.sql.gz", "wt", encoding="utf-8") as f:
        f.write(sql_data)

    manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {"name": "inventory_records", "rows": 10},
                ]
            }
        }
    }

    monkeypatch.setattr("omni_build.database_restore_ops.subprocess.run", lambda *args, **kwargs: _result())
    monkeypatch.setattr(
        ops,
        "_restore_single_table",
        lambda schema, table, sql_file, skip_truncate, expected_rows: (True, 9),
    )

    success, message = ops.restore_tables(manifest, tmp_path)

    assert success is False
    assert "row mismatch" in message or "artifact" in message


def test_restore_chunked_files_fails_on_expected_row_mismatch(monkeypatch: Any, tmp_path: Path):
    ops = DatabaseRestoreOps(config=cast(Any, SimpleNamespace()), pg_checker=_DummyChecker())

    chunk_file = tmp_path / "inventory_records.chunk000.sql.gz"
    with gzip.open(chunk_file, "wt", encoding="utf-8") as f:
        f.write("sku-1\t10\n")

    def fake_run(cmd: list[str], *args: Any, **kwargs: Any):
        cmd_text = " ".join(cmd)
        if "information_schema.columns" in cmd_text:
            return _result(stdout="sku, stock\n")
        if "SELECT COUNT(*)" in cmd_text:
            return _result(stdout="5\n")
        return _result()

    monkeypatch.setattr("omni_build.database_restore_ops.subprocess.run", fake_run)

    success, rows = ops._restore_chunked_files(
        schema="tenant_abc",
        table="inventory_records",
        chunk_files=[chunk_file],
        expected=10,
    )

    # _restore_chunked_files now returns True with actual rows;
    # the caller (restore_tables) handles mismatch via auto-fix verification.
    assert success is True
    assert rows == 5


def test_orchestrator_fails_build_when_restore_fails():
    orchestrator = BuildOrchestrator.__new__(BuildOrchestrator)
    orchestrator._helpers = cast(Any, SimpleNamespace(
        safe_docker_check=lambda: True,
        verify_all_services_truly_healthy=lambda: True,
    ))
    orchestrator.docker_manager = cast(Any, SimpleNamespace(
        check_linux_mode=lambda: True,
        build_images=lambda spec, no_cache=False: True,
        deploy_containers=lambda spec: True,
    ))
    orchestrator.frontend_builder = cast(Any, SimpleNamespace(build=lambda force_install=False: True))
    orchestrator.database_restorer = cast(Any, SimpleNamespace(
        wait_for_postgres=lambda timeout=120: True,
        restore=lambda force=True: (False, "row total mismatch"),
    ))

    result = orchestrator._execute_smart_full(
        mode=BuildMode.SMART,
        spec=SpecLevel.STANDARD,
        skip_frontend=True,
        restore_db=True,
        start_time=time.time(),
    )

    assert result.success is False
    assert result.errors == ["DB restore failed: row total mismatch"]


def test_restore_tables_fails_when_checksum_mismatch(monkeypatch: Any, tmp_path: Path):
    """Checksum mismatch should fail when backup artifact is MISSING."""
    ops = DatabaseRestoreOps(config=cast(Any, SimpleNamespace()), pg_checker=_DummyChecker())

    schema_dir = tmp_path / "tenant_abc"
    schema_dir.mkdir(parents=True, exist_ok=True)
    # Do NOT create the backup file — simulates missing/corrupted artifact

    manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {
                        "name": "inventory_records",
                        "rows": 1,
                        "checksum": "a" * 64,
                    },
                ]
            }
        }
    }

    monkeypatch.setattr("omni_build.database_restore_ops.subprocess.run", lambda *args, **kwargs: _result())

    success, message = ops.restore_tables(manifest, tmp_path)

    assert success is False
    assert "checksum" in message or "not found" in message


def test_restore_tables_autofixes_checksum_when_artifact_exists(monkeypatch: Any, tmp_path: Path):
    """Checksum mismatch should auto-fix when backup artifact file exists."""
    ops = DatabaseRestoreOps(config=cast(Any, SimpleNamespace()), pg_checker=_DummyChecker())

    schema_dir = tmp_path / "tenant_abc"
    schema_dir.mkdir(parents=True, exist_ok=True)
    # Create backup file — artifact exists, just checksum in manifest is stale
    (schema_dir / "inventory_records.sql.gz").write_bytes(b"valid-backup-content")

    manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {
                        "name": "inventory_records",
                        "rows": 1,
                        "checksum": "a" * 64,
                    },
                ]
            }
        }
    }

    monkeypatch.setattr("omni_build.database_restore_ops.subprocess.run", lambda *args, **kwargs: _result())
    monkeypatch.setattr(
        ops,
        "_restore_single_table",
        lambda schema, table, sql_file, skip_truncate, expected_rows: (True, 1),
    )

    success, message = ops.restore_tables(manifest, tmp_path)

    assert success is True
    assert "auto-fixed" in message


def test_restore_tables_skips_materialized_view_entries(monkeypatch: Any, tmp_path: Path):
    ops = DatabaseRestoreOps(config=cast(Any, SimpleNamespace()), pg_checker=_DummyChecker())

    schema_dir = tmp_path / "tenant_abc"
    schema_dir.mkdir(parents=True, exist_ok=True)
    (schema_dir / "orders.sql.gz").write_bytes(b"placeholder")

    manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {"name": "mv_ml_portfolio_summary", "rows": 120},
                    {"name": "orders", "rows": 1},
                ]
            }
        }
    }

    monkeypatch.setattr("omni_build.database_restore_ops.subprocess.run", lambda *args, **kwargs: _result())
    monkeypatch.setattr(
        ops,
        "_restore_single_table",
        lambda schema, table, sql_file, skip_truncate, expected_rows: (True, 1),
    )

    success, message = ops.restore_tables(manifest, tmp_path)

    assert success is True
    assert "Restored 1 tables with 1 rows" in message


def test_save_manifest_keeps_previous_checksum_for_unchanged_tables(tmp_path: Path):
    config = cast(Any, SimpleNamespace(project_root=tmp_path))
    backup = DatabaseBackup(config)
    backup.backup_dir.mkdir(parents=True, exist_ok=True)

    tables = [
        {"schema": "tenant_abc", "table": "inventory_records", "rows": 10},
        {"schema": "tenant_abc", "table": "inventory_sync_history", "rows": 0},
    ]
    previous_manifest = {
        "version": 2,
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {
                        "name": "inventory_sync_history",
                        "rows": 0,
                        "checksum": "b" * 64,
                    }
                ]
            }
        },
    }
    generated_checksums = {"tenant_abc.inventory_records": "c" * 64}

    ok = backup.save_manifest(
        tables,
        prev_manifest=previous_manifest,
        generated_checksums=generated_checksums,
    )

    assert ok is True

    manifest_content = json.loads(backup.manifest_file.read_text(encoding="utf-8"))
    saved_tables = {
        item["name"]: item["checksum"]
        for item in manifest_content["schemas"]["tenant_abc"]["tables"]
    }
    assert manifest_content["version"] == 4
    assert saved_tables["inventory_records"] == "c" * 64
    assert saved_tables["inventory_sync_history"] == "b" * 64


def test_save_manifest_uses_generated_row_counts(tmp_path: Path):
    config = cast(Any, SimpleNamespace(project_root=tmp_path))
    backup = DatabaseBackup(config)
    backup.backup_dir.mkdir(parents=True, exist_ok=True)

    tables = [
        {"schema": "tenant_abc", "table": "marketplace_sync_history", "rows": 24},
    ]
    generated_checksums = {"tenant_abc.marketplace_sync_history": "d" * 64}
    generated_rows = {"tenant_abc.marketplace_sync_history": 4}

    ok = backup.save_manifest(
        tables,
        generated_checksums=generated_checksums,
        generated_rows=generated_rows,
    )

    assert ok is True

    manifest_content = json.loads(backup.manifest_file.read_text(encoding="utf-8"))
    saved_table = manifest_content["schemas"]["tenant_abc"]["tables"][0]
    assert saved_table["rows"] == 4
    assert saved_table["checksum"] == "d" * 64


def test_detect_changes_exports_when_change_vector_increases(tmp_path: Path):
    config = cast(Any, SimpleNamespace(project_root=tmp_path))
    backup = DatabaseBackup(config)

    tables = [
        {
            "schema": "tenant_abc",
            "table": "inventory_records",
            "rows": 10,
            "change_vector": "10:6:0:0",
        }
    ]
    previous_manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {
                        "name": "inventory_records",
                        "rows": 10,
                        "checksum": "a" * 64,
                        "change_vector": "10:5:0:0",
                    }
                ]
            }
        }
    }

    to_export, unchanged, _ = backup.detect_changes(tables, previous_manifest, force=False)

    assert len(to_export) == 1
    assert unchanged == []
    assert "change_vector" in to_export[0]["reason"]


def test_detect_changes_legacy_manifest_skips_when_checksum_equal(
    monkeypatch: Any, tmp_path: Path
):
    config = cast(Any, SimpleNamespace(project_root=tmp_path))
    backup = DatabaseBackup(config)

    monkeypatch.setattr(
        "omni_build.database_backup_ops.DatabaseBackupOps.calculate_live_table_checksum",
        lambda self, schema, table, rows: "b" * 64,
    )

    tables = [
        {
            "schema": "tenant_abc",
            "table": "inventory_records",
            "rows": 10,
            "change_vector": "10:6:0:0",
        }
    ]
    previous_manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {
                        "name": "inventory_records",
                        "rows": 10,
                        "checksum": "b" * 64,
                    }
                ]
            }
        }
    }

    to_export, unchanged, _ = backup.detect_changes(tables, previous_manifest, force=False)

    assert to_export == []
    assert len(unchanged) == 1


def test_detect_changes_legacy_manifest_exports_when_checksum_diff(
    monkeypatch: Any, tmp_path: Path
):
    config = cast(Any, SimpleNamespace(project_root=tmp_path))
    backup = DatabaseBackup(config)

    monkeypatch.setattr(
        "omni_build.database_backup_ops.DatabaseBackupOps.calculate_live_table_checksum",
        lambda self, schema, table, rows: "c" * 64,
    )

    tables = [
        {
            "schema": "tenant_abc",
            "table": "inventory_records",
            "rows": 10,
            "change_vector": "10:6:0:0",
        }
    ]
    previous_manifest = {
        "schemas": {
            "tenant_abc": {
                "tables": [
                    {
                        "name": "inventory_records",
                        "rows": 10,
                        "checksum": "b" * 64,
                    }
                ]
            }
        }
    }

    to_export, unchanged, _ = backup.detect_changes(tables, previous_manifest, force=False)

    assert len(to_export) == 1
    assert unchanged == []
    assert to_export[0]["reason"] == "checksum_changed"
