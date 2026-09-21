"""Data-safety regression tests.

Guards added after the 2026-09-19 data-safety audit:

1. aggressive_cleanup must NEVER include --volumes: it is auto-invoked by
   error escalation (L6) and disk-full symptom handling, and wiping unused
   named volumes machine-wide would destroy the Postgres data volume.
2. create_postgres_database_if_not_exists must NOT silently fall back to
   repair_postgres_database (compose down + data rmtree) on exception.
3. import_all must keep the pre-import backup when the import finishes
   with errors (tables are left truncated/partially imported; the backup
   is the only local snapshot of the pre-import state).
"""
from pathlib import Path
from types import SimpleNamespace
from unittest.mock import patch

import pytest

from omni_build import ndjson_sync
from omni_build.container_runtime import ContainerRuntime


class TestAggressiveCleanupNeverTouchesVolumes:
    """system prune --volumes is machine-wide and would kill omni-pgdata."""

    def test_aggressive_cleanup_does_not_pass_volumes(self):
        from omni_build.error_handler_fixes import SystemFixer

        with patch("omni_build.error_handler_fixes.get_runtime") as mock_gr, \
             patch("omni_build.error_handler_fixes.subprocess.run") as mock_run:
            runtime = ContainerRuntime("podman")
            mock_gr.return_value = runtime
            SystemFixer.aggressive_cleanup()

        cmd = mock_run.call_args.args[0]
        assert "--volumes" not in cmd, (
            f"aggressive_cleanup must never prune volumes, got: {cmd}"
        )

    def test_aggressive_cleanup_still_prunes_unused_images(self):
        from omni_build.error_handler_fixes import SystemFixer

        with patch("omni_build.error_handler_fixes.get_runtime") as mock_gr, \
             patch("omni_build.error_handler_fixes.subprocess.run") as mock_run:
            mock_gr.return_value = ContainerRuntime("podman")
            SystemFixer.aggressive_cleanup()

        cmd = mock_run.call_args.args[0]
        assert "-af" in cmd


class TestPostgresFixerNoSilentReset:
    """A transient error must not trigger a destructive data reset."""

    def test_create_db_failure_does_not_call_repair(self):
        from omni_build.service_fixers_postgres import PostgresFixer

        with patch.object(
            PostgresFixer, "repair_postgres_database"
        ) as mock_repair, \
             patch(
            "omni_build.service_fixers_postgres.get_runtime"
        ) as mock_gr, \
             patch(
            "omni_build.service_fixers_postgres.subprocess.run"
        ) as mock_run:
            mock_gr.return_value = ContainerRuntime("podman")
            # Every command fails (transient error), including CREATE DATABASE.
            mock_run.return_value = SimpleNamespace(
                returncode=1, stdout="", stderr="transient failure",
            )

            result = PostgresFixer.create_postgres_database_if_not_exists()

        assert result is False
        mock_repair.assert_not_called()


class TestImportAllKeepsBackupOnErrors:
    """The pre-import snapshot must survive a failed import."""

    def test_backup_kept_when_import_reports_errors(self, tmp_path: Path):
        sync_dir = tmp_path / "backups" / "sync"
        sync_dir.mkdir(parents=True)
        (sync_dir / "manifest.json").write_text(
            '{"version": 2, "tables": {}, "total_rows": 0}',
            encoding="utf-8",
        )

        created_dirs: list[Path] = []

        def fake_create_backup(tables, sync_root):
            d = tmp_path / "pre-import-backup-test"
            d.mkdir()
            created_dirs.append(d)
            return d

        # One table with a NDJSON artifact that fails to import after truncate.
        table = SimpleNamespace(
            schema="tenant_x", name="t1", full_name="tenant_x.t1",
            pk_columns=["id"], has_serial_pk=False,
        )
        table_dir = sync_dir / "tenant_x"
        table_dir.mkdir()
        (table_dir / "t1.ndjson").write_text('{"id": 1}\n', encoding="utf-8")
        errors_result = SimpleNamespace(errors=1, imported=0, table=table)

        with patch.object(ndjson_sync, "materialize_lfs_pointers", return_value=True), \
             patch.object(ndjson_sync, "discover_tables", return_value=[table]), \
             patch.object(ndjson_sync, "_topo_sort", return_value=[table]), \
             patch.object(ndjson_sync, "_discover_secondary_unique_constraints", return_value={}), \
             patch.object(ndjson_sync, "_create_pre_import_backup", side_effect=fake_create_backup), \
             patch.object(ndjson_sync, "_psql_exec", return_value=(True, "")), \
             patch.object(ndjson_sync, "_psql_query", return_value=(True, "")), \
             patch.object(
                 ndjson_sync, "_import_table", return_value=errors_result
             ):
            success, results = ndjson_sync.import_all(tmp_path)

        assert success is False
        assert len(created_dirs) == 1
        assert created_dirs[0].exists(), (
            "pre-import backup must be kept when the import finishes with errors"
        )

    def test_import_all_does_not_require_a_live_runtime(self, tmp_path: Path):
        """This test mocks every DB call, so it must not need a running runtime.

        Regression note (2026-09-21): the assertions above only held while the
        Podman machine happened to be running. With the machine stopped,
        `DatabaseConfig.docker_exec_prefix()` raised
        "No container runtime found" from inside a fully-mocked code path, so
        the suite failed purely because a VM was off. The test must exercise
        its own mocks, not the host's container runtime.
        """
        sync_dir = tmp_path / "backups" / "sync"
        sync_dir.mkdir(parents=True)
        (sync_dir / "manifest.json").write_text(
            '{"version": 2, "tables": {}, "total_rows": 0}',
            encoding="utf-8",
        )

        table = SimpleNamespace(
            schema="tenant_x", name="t1", full_name="tenant_x.t1",
            pk_columns=["id"], has_serial_pk=False,
        )

        # Simulate a host with NO reachable container runtime at all.
        with patch.object(
            ndjson_sync, "materialize_lfs_pointers", return_value=True,
        ), patch.object(
            ndjson_sync, "discover_tables", return_value=[table],
        ), patch.object(
            ndjson_sync, "_topo_sort", return_value=[table],
        ), patch.object(
            ndjson_sync, "_discover_secondary_unique_constraints", return_value={},
        ), patch.object(
            ndjson_sync, "_create_pre_import_backup",
            side_effect=lambda tables, sync_root: tmp_path,
        ), patch.object(
            ndjson_sync, "_psql_exec", return_value=(True, ""),
        ), patch.object(
            ndjson_sync, "_psql_query", return_value=(True, ""),
        ), patch.object(
            ndjson_sync, "_import_table",
            return_value=SimpleNamespace(errors=0, imported=1, table=table),
        ):
            success, _ = ndjson_sync.import_all(tmp_path)

        assert success is True, (
            "import_all must succeed on fully-mocked calls even when no "
            "container runtime is reachable"
        )
