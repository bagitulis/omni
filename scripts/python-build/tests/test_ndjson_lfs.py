"""Unit tests for Git LFS pointer auto-fix in NDJSON sync import."""
import subprocess
from pathlib import Path

import pytest

from omni_build.ndjson_lfs import (
    find_lfs_pointers,
    materialize_lfs_pointers,
)


def _write_lfs_pointer(filepath: Path) -> None:
    filepath.parent.mkdir(parents=True, exist_ok=True)
    _ = filepath.write_bytes(
        b"version https://git-lfs.github.com/spec/v1\n"
        + b"oid sha256:1234567890abcdef\n"
        + b"size 12345\n"
    )


def test_find_lfs_pointers_detects_only_large_ndjson_pointers(tmp_path: Path) -> None:
    sync_dir = tmp_path / "backups" / "sync"
    pointer_file = sync_dir / "_large" / "tenant.table.ndjson.gz"
    real_file = sync_dir / "_large" / "tenant.other.ndjson.gz"
    plain_pointer = sync_dir / "tenant" / "table.ndjson"

    _write_lfs_pointer(pointer_file)
    _ = real_file.write_bytes(b"\x1f\x8b\x08real gzip data")
    _write_lfs_pointer(plain_pointer)

    assert find_lfs_pointers(sync_dir) == [pointer_file]


def test_materialize_lfs_pointers_runs_targeted_pull_with_git_paths(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    project_root = tmp_path
    sync_dir = project_root / "backups" / "sync"
    pointer_file = sync_dir / "_large" / "tenant.table.ndjson.gz"
    _write_lfs_pointer(pointer_file)

    commands: list[list[str]] = []

    def fake_run(
        args: list[str],
        cwd: Path,
        capture_output: bool,
        env: dict[str, str],
        text: bool,
        timeout: int,
        encoding: str,
        errors: str,
    ) -> subprocess.CompletedProcess[str]:
        _ = timeout
        commands.append(args)
        assert cwd == project_root
        assert capture_output is True
        assert env["GIT_TERMINAL_PROMPT"] == "0"
        assert env["GCM_INTERACTIVE"] == "never"
        assert text is True
        assert encoding == "utf-8"
        assert errors == "replace"
        if args[:3] == ["git", "lfs", "pull"]:
            _ = pointer_file.write_bytes(b"\x1f\x8b\x08materialized gzip data")
        return subprocess.CompletedProcess(args, 0, stdout="ok", stderr="")

    monkeypatch.setattr(subprocess, "run", fake_run)

    assert materialize_lfs_pointers(project_root, sync_dir) is True
    assert commands == [
        ["git", "lfs", "install", "--local"],
        ["git", "lfs", "pull", "--include=backups/sync/_large/tenant.table.ndjson.gz"],
    ]
    assert (project_root / ".gitattributes").read_text(encoding="utf-8") == (
        "backups/sync/_large/*.ndjson.gz filter=lfs diff=lfs merge=lfs -text\n"
    )


def test_materialize_lfs_pointers_fails_when_git_lfs_pull_fails(
    tmp_path: Path,
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    project_root = tmp_path
    sync_dir = project_root / "backups" / "sync"
    pointer_file = sync_dir / "_large" / "tenant.table.ndjson.gz"
    _write_lfs_pointer(pointer_file)

    def fake_run(
        args: list[str],
        cwd: Path,
        capture_output: bool,
        env: dict[str, str],
        text: bool,
        timeout: int,
        encoding: str,
        errors: str,
    ) -> subprocess.CompletedProcess[str]:
        _ = (cwd, capture_output, env, text, timeout, encoding, errors)
        if args[:3] == ["git", "lfs", "pull"]:
            return subprocess.CompletedProcess(args, 2, stdout="", stderr="missing object")
        return subprocess.CompletedProcess(args, 0, stdout="ok", stderr="")

    monkeypatch.setattr(subprocess, "run", fake_run)

    assert materialize_lfs_pointers(project_root, sync_dir) is False
