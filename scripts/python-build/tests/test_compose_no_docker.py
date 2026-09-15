"""
Regression tests for the WinError-2 crash in subprocess_utils.get_compose_command
when the host has no `docker` on PATH (podman-only environment).

Bug: ensure_buildx_plugin() and has_compose_plugin() called subprocess.run
with ["docker", ...] unconditionally. On Windows with docker absent, that
raises FileNotFoundError (WinError 2) instead of returning cleanly.

Also: get_compose_command() falls back to the Docker legacy path even when
runtime was resolved to podman. That path is docker-only and cannot help.
"""
from __future__ import annotations

import subprocess
import sys
from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(ROOT))

from omni_build import container_runtime, subprocess_utils  # noqa: E402


@pytest.fixture(autouse=True)
def reset_runtime_singleton():
    """Ensure each test starts with a fresh runtime singleton."""
    container_runtime.reset_runtime()
    yield
    container_runtime.reset_runtime()


def _fake_which(monkeypatch: pytest.MonkeyPatch, present: set[str]) -> None:
    """Replace shutil.which so only names in `present` resolve."""
    import shutil as _shutil

    def fake_which(name, *args, **kwargs):
        if name in present:
            return f"C:\\fake\\{name}.exe"
        return None

    monkeypatch.setattr(container_runtime.shutil, "which", fake_which)
    # subprocess_utils imports shutil lazily inside get_compose_command()
    monkeypatch.setattr(_shutil, "which", fake_which)
    monkeypatch.setattr(subprocess_utils, "subprocess", subprocess)


def _fake_podman_available(monkeypatch: pytest.MonkeyPatch) -> None:
    """Make container_runtime.detect_runtime believe podman is up."""
    def fake_run_silent(args, timeout=None):
        return subprocess.CompletedProcess(args, 0, stdout="", stderr="")

    monkeypatch.setattr(container_runtime, "_run_silent", fake_run_silent)


def _explode_on_docker(monkeypatch: pytest.MonkeyPatch) -> None:
    """Force subprocess.run to raise FileNotFoundError when called with ['docker', ...]."""
    real_run = subprocess.run

    def fake_run(args, *pos, **kwargs):
        if isinstance(args, list) and args and args[0] == "docker":
            raise FileNotFoundError(2, "The system cannot find the file specified", "docker")
        return real_run(args, *pos, **kwargs)

    monkeypatch.setattr(subprocess_utils.subprocess, "run", fake_run)


def test_get_compose_command_returns_podman_when_docker_missing(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    """With CONTAINER_RUNTIME=podman and no docker CLI, we must NOT crash.

    Expected: returns a podman-compose command (["podman-compose"] or
    ["podman", "compose"]), and never invokes subprocess.run(["docker", ...]).
    """
    monkeypatch.setenv("CONTAINER_RUNTIME", "podman")
    _fake_which(monkeypatch, present={"podman", "podman-compose"})
    _fake_podman_available(monkeypatch)
    _explode_on_docker(monkeypatch)

    cmd = subprocess_utils.get_compose_command()
    assert cmd[0] in ("podman-compose", "podman"), (
        f"expected a podman variant, got {cmd!r}"
    )
    if cmd[0] == "podman":
        assert cmd[1] == "compose"


def test_get_compose_command_error_message_when_nothing_available(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    """When docker is missing AND podman-compose is missing, error must be
    a clear RuntimeError - never a bare FileNotFoundError (WinError 2)."""
    monkeypatch.setenv("CONTAINER_RUNTIME", "podman")
    _fake_which(monkeypatch, present={"podman"})  # podman but no compose
    # podman info succeeds (runtime detects), but `podman compose version` fails.
    def selective_run_silent(args, timeout=None):
        if args[:2] == ["podman", "info"]:
            return subprocess.CompletedProcess(args, 0, "", "")
        return subprocess.CompletedProcess(args, 127, "", "not found")
    monkeypatch.setattr(container_runtime, "_run_silent", selective_run_silent)
    _explode_on_docker(monkeypatch)

    with pytest.raises(RuntimeError) as ei:
        subprocess_utils.get_compose_command()
    assert "podman-compose" in str(ei.value).lower() or "podman compose" in str(ei.value).lower()


def test_get_compose_command_does_not_shell_out_to_docker_when_podman_active(
    monkeypatch: pytest.MonkeyPatch,
) -> None:
    """The legacy Docker fallback must not run `docker ...` when the
    active runtime is podman (regression for WinError 2 crash)."""
    monkeypatch.setenv("CONTAINER_RUNTIME", "podman")
    _fake_which(monkeypatch, present={"podman", "podman-compose"})
    _fake_podman_available(monkeypatch)
    _explode_on_docker(monkeypatch)  # blows up if code shells out to docker

    # Must not raise FileNotFoundError. Silence is proof.
    cmd = subprocess_utils.get_compose_command()
    assert cmd[0] in ("podman-compose", "podman")
