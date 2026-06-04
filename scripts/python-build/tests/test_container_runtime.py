"""Unit tests for ContainerRuntime abstraction layer."""
from __future__ import annotations

import os
from unittest.mock import MagicMock, patch

import pytest

from omni_build.container_runtime import (
    ContainerRuntime,
    detect_runtime,
    get_runtime,
    reset_runtime,
)


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture(autouse=True)
def _reset_singleton():
    """Reset the singleton before each test."""
    reset_runtime()
    yield
    reset_runtime()


# ---------------------------------------------------------------------------
# Tests for detect_runtime
# ---------------------------------------------------------------------------

class TestDetectRuntime:
    """Tests for the detect_runtime() function."""

    def test_env_var_docker(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": "docker"}):
            assert detect_runtime() == "docker"

    def test_env_var_podman(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": "podman"}):
            assert detect_runtime() == "podman"

    def test_env_var_case_insensitive(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": "DOCKER"}):
            assert detect_runtime() == "docker"

    def test_env_var_strips_whitespace(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": "  podman  "}):
            assert detect_runtime() == "podman"

    def test_auto_detect_docker_first(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": ""}, clear=False):
            with patch("omni_build.container_runtime.shutil.which") as mock_which:
                mock_which.side_effect = lambda cmd: cmd == "docker"
                with patch("omni_build.container_runtime._run_silent") as mock_run:
                    mock_run.return_value = MagicMock(returncode=0)
                    assert detect_runtime() == "docker"

    def test_auto_detect_podman_fallback(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": ""}, clear=False):
            with patch("omni_build.container_runtime.shutil.which") as mock_which:
                mock_which.side_effect = lambda cmd: cmd == "podman"
                with patch("omni_build.container_runtime._run_silent") as mock_run:
                    mock_run.return_value = MagicMock(returncode=0)
                    assert detect_runtime() == "podman"

    def test_auto_detect_docker_first_when_both_available(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": ""}, clear=False):
            with patch("omni_build.container_runtime.shutil.which") as mock_which:
                mock_which.side_effect = lambda cmd: cmd in ("docker", "podman")
                with patch("omni_build.container_runtime._run_silent") as mock_run:
                    mock_run.return_value = MagicMock(returncode=0)
                    assert detect_runtime() == "docker"
    def test_auto_detect_raises_when_none_found(self):
        with patch.dict(os.environ, {"CONTAINER_RUNTIME": ""}, clear=False):
            with patch("omni_build.container_runtime.shutil.which", return_value=None):
                with pytest.raises(RuntimeError, match="No container runtime found"):
                    detect_runtime()


# ---------------------------------------------------------------------------
# Tests for ContainerRuntime CLI wrappers
# ---------------------------------------------------------------------------

class TestContainerRuntimeCLI:
    """Tests that all CLI wrapper methods return correct command arrays."""

    def test_get_cli_command_docker(self):
        rt = ContainerRuntime("docker")
        assert rt.get_cli_command() == ["docker"]

    def test_get_cli_command_podman(self):
        rt = ContainerRuntime("podman")
        assert rt.get_cli_command() == ["podman"]

    def test_name_property(self):
        rt = ContainerRuntime("podman")
        assert rt.name == "podman"

    def test_inspect_basic(self):
        rt = ContainerRuntime("docker")
        assert rt.inspect("my-container") == ["docker", "inspect", "my-container"]

    def test_inspect_with_format(self):
        rt = ContainerRuntime("docker")
        assert rt.inspect("c", "{{.State.Status}}") == [
            "docker", "inspect", "--format", "{{.State.Status}}", "c"
        ]

    def test_ps_no_args(self):
        rt = ContainerRuntime("docker")
        assert rt.ps() == ["docker", "ps"]

    def test_ps_with_args(self):
        rt = ContainerRuntime("docker")
        assert rt.ps(["-a", "--filter", "name=omni-"]) == [
            "docker", "ps", "-a", "--filter", "name=omni-"
        ]

    def test_exec_no_interactive(self):
        rt = ContainerRuntime("docker")
        assert rt.exec("c", "psql", "-U", "user") == [
            "docker", "exec", "c", "psql", "-U", "user"
        ]

    def test_exec_interactive(self):
        rt = ContainerRuntime("docker")
        assert rt.exec("c", "psql", "-U", "user", interactive=True) == [
            "docker", "exec", "-i", "c", "psql", "-U", "user"
        ]

    def test_restart(self):
        rt = ContainerRuntime("docker")
        assert rt.restart("c") == ["docker", "restart", "c"]

    def test_start_container(self):
        rt = ContainerRuntime("docker")
        assert rt.start_container("c") == ["docker", "start", "c"]

    def test_stop(self):
        rt = ContainerRuntime("docker")
        assert rt.stop("c") == ["docker", "stop", "c"]

    def test_rm_no_force(self):
        rt = ContainerRuntime("docker")
        assert rt.rm("c") == ["docker", "rm", "c"]

    def test_rm_force(self):
        rt = ContainerRuntime("docker")
        assert rt.rm("c", force=True) == ["docker", "rm", "-f", "c"]

    def test_logs_no_tail(self):
        rt = ContainerRuntime("docker")
        assert rt.logs("c") == ["docker", "logs", "c"]

    def test_logs_with_tail(self):
        rt = ContainerRuntime("docker")
        assert rt.logs("c", tail=50) == ["docker", "logs", "--tail", "50", "c"]

    def test_cp(self):
        rt = ContainerRuntime("docker")
        assert rt.cp("file.txt", "c:/tmp/") == ["docker", "cp", "file.txt", "c:/tmp/"]

    def test_info(self):
        rt = ContainerRuntime("docker")
        assert rt.info() == ["docker", "info"]

    def test_version_no_format(self):
        rt = ContainerRuntime("docker")
        assert rt.version() == ["docker", "version"]

    def test_version_with_format(self):
        rt = ContainerRuntime("docker")
        assert rt.version("{{.Server.Os}}") == [
            "docker", "version", "--format", "{{.Server.Os}}"
        ]

    def test_network_prune(self):
        rt = ContainerRuntime("docker")
        assert rt.network_prune() == ["docker", "network", "prune", "-f"]

    def test_container_prune(self):
        rt = ContainerRuntime("docker")
        assert rt.container_prune() == ["docker", "container", "prune", "-f"]

    def test_image_prune(self):
        rt = ContainerRuntime("docker")
        assert rt.image_prune() == ["docker", "image", "prune", "-f"]

    def test_system_prune_basic(self):
        rt = ContainerRuntime("docker")
        assert rt.system_prune() == ["docker", "system", "prune", "-f"]

    def test_system_prune_all(self):
        rt = ContainerRuntime("docker")
        assert rt.system_prune(all=True) == ["docker", "system", "prune", "-af"]

    def test_system_prune_all_volumes(self):
        rt = ContainerRuntime("docker")
        assert rt.system_prune(all=True, volumes=True) == [
            "docker", "system", "prune", "-af", "--volumes"
        ]

    def test_builder_prune(self):
        rt = ContainerRuntime("docker")
        assert rt.builder_prune() == ["docker", "builder", "prune"]

    def test_builder_prune_all(self):
        rt = ContainerRuntime("docker")
        assert rt.builder_prune(all=True) == ["docker", "builder", "prune", "-af"]

    def test_command(self):
        rt = ContainerRuntime("docker")
        assert rt.command(["buildx", "version"]) == ["docker", "buildx", "version"]

    def test_podman_commands(self):
        """Verify all methods use podman prefix when runtime is podman."""
        rt = ContainerRuntime("podman")
        assert rt.info() == ["podman", "info"]
        assert rt.ps() == ["podman", "ps"]
        assert rt.exec("c", "cmd") == ["podman", "exec", "c", "cmd"]
        assert rt.restart("c") == ["podman", "restart", "c"]
        assert rt.stop("c") == ["podman", "stop", "c"]
        assert rt.rm("c", force=True) == ["podman", "rm", "-f", "c"]


# ---------------------------------------------------------------------------
# Tests for compose command detection
# ---------------------------------------------------------------------------

class TestComposeCommand:
    """Tests for get_compose_command() with different runtimes."""

    def test_docker_compose_standalone(self):
        rt = ContainerRuntime("docker")
        with patch("omni_build.container_runtime.shutil.which") as mock_which:
            mock_which.side_effect = lambda cmd: cmd == "docker-compose"
            assert rt.get_compose_command() == ["docker-compose"]

    def test_docker_compose_plugin(self):
        rt = ContainerRuntime("docker")
        with patch("omni_build.container_runtime.shutil.which", return_value=None):
            with patch("omni_build.container_runtime._run_silent") as mock_run:
                mock_run.return_value = MagicMock(returncode=0)
                assert rt.get_compose_command() == ["docker", "compose"]

    def test_podman_compose_standalone(self):
        rt = ContainerRuntime("podman")
        with patch("omni_build.container_runtime.shutil.which") as mock_which:
            mock_which.side_effect = lambda cmd: cmd == "podman-compose"
            assert rt.get_compose_command() == ["podman-compose"]

    def test_podman_compose_plugin(self):
        rt = ContainerRuntime("podman")
        with patch("omni_build.container_runtime.shutil.which", return_value=None):
            with patch("omni_build.container_runtime._run_silent") as mock_run:
                mock_run.return_value = MagicMock(returncode=0)
                assert rt.get_compose_command() == ["podman", "compose"]

    def test_raises_when_no_compose(self):
        rt = ContainerRuntime("docker")
        with patch("omni_build.container_runtime.shutil.which", return_value=None):
            with patch("omni_build.container_runtime._run_silent") as mock_run:
                mock_run.return_value = MagicMock(returncode=1)
                with pytest.raises(RuntimeError, match="not available"):
                    rt.get_compose_command()


# ---------------------------------------------------------------------------
# Tests for restart_docker_service
# ---------------------------------------------------------------------------

class TestRestartDockerService:
    """Tests for restart_docker_service()."""

    def test_docker(self):
        rt = ContainerRuntime("docker")
        assert rt.restart_docker_service() == ["sudo", "systemctl", "restart", "docker"]

    def test_podman(self):
        rt = ContainerRuntime("podman")
        assert rt.restart_docker_service() == ["sudo", "systemctl", "--user", "restart", "podman"]


# ---------------------------------------------------------------------------
# Tests for is_running
# ---------------------------------------------------------------------------

class TestIsRunning:
    """Tests for is_running()."""

    def test_returns_true_when_running(self):
        rt = ContainerRuntime("docker")
        with patch("omni_build.container_runtime._run_silent") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            assert rt.is_running() is True

    def test_returns_false_when_not_running(self):
        rt = ContainerRuntime("docker")
        with patch("omni_build.container_runtime._run_silent") as mock_run:
            mock_run.return_value = MagicMock(returncode=1)
            assert rt.is_running() is False

    def test_returns_false_on_timeout(self):
        import subprocess
        rt = ContainerRuntime("docker")
        with patch("omni_build.container_runtime._run_silent", side_effect=subprocess.TimeoutExpired("docker", 10)):
            assert rt.is_running() is False


# ---------------------------------------------------------------------------
# Tests for singleton
# ---------------------------------------------------------------------------

class TestSingleton:
    """Tests for get_runtime() singleton pattern."""

    def test_creates_instance(self):
        with patch("omni_build.container_runtime.detect_runtime", return_value="docker"):
            rt = get_runtime()
            assert isinstance(rt, ContainerRuntime)
            assert rt.name == "docker"

    def test_returns_same_instance(self):
        with patch("omni_build.container_runtime.detect_runtime", return_value="docker"):
            rt1 = get_runtime()
            rt2 = get_runtime()
            assert rt1 is rt2

    def test_reset_creates_new_instance(self):
        with patch("omni_build.container_runtime.detect_runtime", return_value="docker"):
            rt1 = get_runtime()
            reset_runtime()
            with patch("omni_build.container_runtime.detect_runtime", return_value="podman"):
                rt2 = get_runtime()
            assert rt1 is not rt2
