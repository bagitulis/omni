"""Container runtime abstraction for Docker and Podman.

SRP: Single entry point for all container runtime CLI operations.
Supports Docker and Podman with auto-detection and CONTAINER_RUNTIME env var override.
"""
from __future__ import annotations

import os
import platform
import shutil
import subprocess
from typing import Optional

from omni_build.subprocess_utils import run_silent

_runtime: Optional[ContainerRuntime] = None


def _real_home():
    """Get the real home directory, handling Linux pwd module."""
    try:
        import pwd
        from pathlib import Path
        return Path(pwd.getpwuid(os.getuid()).pw_dir)
    except (ImportError, KeyError):
        from pathlib import Path
        return Path.home()


def detect_runtime() -> str:
    """Detect which container runtime is available.

    Priority: CONTAINER_RUNTIME env var > podman > docker.
    Returns "docker" or "podman". Raises RuntimeError if none found.
    """
    env_runtime = os.environ.get("CONTAINER_RUNTIME", "").strip().lower()
    if env_runtime in ("docker", "podman"):
        return env_runtime

    for candidate in ("podman", "docker"):
        if shutil.which(candidate):
            try:
                result = run_silent([candidate, "info"], timeout=10)
                if result.returncode == 0:
                    return candidate
            except (subprocess.TimeoutExpired, FileNotFoundError):
                continue

    raise RuntimeError(
        "No container runtime found. Install Docker or Podman, "
        + "or set CONTAINER_RUNTIME=docker|podman in your environment."
    )


class ContainerRuntime:
    """Wraps Docker/Podman CLI operations.

    All methods return command arrays (list[str]) that callers pass to
    subprocess_utils.run() or subprocess.run(). Exceptions are is_running()
    and start() which execute commands internally to check state.
    """

    def __init__(self, runtime: str = "docker") -> None:
        self._runtime: str = runtime
        self._is_windows: bool = platform.system() == "Windows"

    @property
    def name(self) -> str:
        """Return the runtime name ('docker' or 'podman')."""
        return self._runtime

    def get_cli_command(self) -> list[str]:
        """Return the base CLI command array: ["docker"] or ["podman"]."""
        return [self._runtime]

    def get_compose_command(self) -> list[str]:
        """Return the compose command array.

        Detection: podman-compose > podman compose > docker-compose > docker compose.
        Raises RuntimeError if no compose variant is found.
        """
        if self._runtime == "podman":
            return self._detect_podman_compose()
        return self._detect_docker_compose()

    def _detect_podman_compose(self) -> list[str]:
        if shutil.which("podman-compose"):
            return ["podman-compose"]
        try:
            result = run_silent(["podman", "compose", "version"], timeout=10)
            if result.returncode == 0:
                return ["podman", "compose"]
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass
        raise RuntimeError(
            "Podman Compose is not available. Install podman-compose "
            + "or ensure `podman compose version` works."
        )

    def _detect_docker_compose(self) -> list[str]:
        if shutil.which("docker-compose"):
            return ["docker-compose"]
        try:
            result = run_silent(["docker", "compose", "version"], timeout=10)
            if result.returncode == 0:
                return ["docker", "compose"]
        except (subprocess.TimeoutExpired, FileNotFoundError):
            pass
        raise RuntimeError(
            "Docker Compose is not available. Install docker-compose-plugin "
            + "or ensure `docker compose version` works."
        )

    # === State checks (execute commands internally) ===

    def is_running(self) -> bool:
        """Check if the container runtime daemon is available."""
        try:
            result = run_silent(self.info(), timeout=10)
            return result.returncode == 0
        except (subprocess.TimeoutExpired, FileNotFoundError):
            return False

    def start(self) -> bool:
        """Start the container runtime (Docker Desktop / podman machine / systemctl)."""
        if self._runtime == "podman":
            return self._start_podman()
        return self._start_docker()

    def _start_podman(self) -> bool:
        if self._is_windows or platform.system() == "Darwin":
            try:
                result = run_silent(["podman", "machine", "start"], timeout=120)
                return result.returncode == 0
            except (subprocess.TimeoutExpired, FileNotFoundError):
                return False
        return self.is_running()

    def _start_docker(self) -> bool:
        if self._is_windows:
            return self._start_docker_desktop()
        # Linux: try systemctl then service
        for cmd in (["sudo", "systemctl", "start", "docker"],
                    ["sudo", "service", "docker", "start"]):
            if shutil.which(cmd[1]):
                try:
                    result = run_silent(cmd, timeout=30)
                    if result.returncode == 0:
                        return True
                except (subprocess.TimeoutExpired, FileNotFoundError):
                    continue
        return False

    def _start_docker_desktop(self) -> bool:
        import time
        from pathlib import Path

        docker_paths = [
            Path("C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe"),
            Path("C:\\Program Files (x86)\\Docker\\Docker\\Docker Desktop.exe"),
            _real_home() / "AppData" / "Local" / "Docker" / "Docker Desktop.exe",
        ]
        for docker_path in docker_paths:
            if docker_path.exists():
                try:
                    _proc = subprocess.Popen(
                        [str(docker_path)],
                        stdout=subprocess.DEVNULL,
                        stderr=subprocess.DEVNULL,
                    )  # noqa: S603
                    for _ in range(15):
                        time.sleep(2)
                        if self.is_running():
                            time.sleep(3)
                            return True
                except Exception:
                    continue
        return False

    def restart_docker_service(self) -> list[str]:
        """Return command to restart the container runtime service."""
        if self._runtime == "podman":
            return ["sudo", "systemctl", "--user", "restart", "podman"]
        return ["sudo", "systemctl", "restart", "docker"]

    # === CLI wrappers (return command arrays) ===

    def inspect(self, container: str, format_str: Optional[str] = None) -> list[str]:
        """Return inspect command. E.g. ["docker", "inspect", "--format", "{{.State.Status}}", "c"]"""
        cmd = self.get_cli_command() + ["inspect"]
        if format_str is not None:
            cmd.extend(["--format", format_str])
        cmd.append(container)
        return cmd

    def ps(self, args: Optional[list[str]] = None) -> list[str]:
        """Return ps command. E.g. ["docker", "ps", "-a", "--filter", "name=omni-"]"""
        cmd = self.get_cli_command() + ["ps"]
        if args:
            cmd.extend(args)
        return cmd

    def exec(self, container: str, *args: str, interactive: bool = False) -> list[str]:
        """Return exec command. interactive=True adds -i flag (for stdin piping)."""
        cmd = self.get_cli_command() + ["exec"]
        if interactive:
            cmd.append("-i")
        cmd.append(container)
        cmd.extend(args)
        return cmd

    def restart(self, container: str) -> list[str]:
        """Return restart command."""
        return self.get_cli_command() + ["restart", container]

    def start_container(self, container: str) -> list[str]:
        """Return start command."""
        return self.get_cli_command() + ["start", container]

    def stop(self, container: str) -> list[str]:
        """Return stop command."""
        return self.get_cli_command() + ["stop", container]

    def rm(self, container: str, force: bool = False) -> list[str]:
        """Return rm command. force=True adds -f flag."""
        cmd = self.get_cli_command() + ["rm"]
        if force:
            cmd.append("-f")
        cmd.append(container)
        return cmd

    def logs(self, container: str, tail: Optional[int] = None) -> list[str]:
        """Return logs command. tail=N adds --tail N."""
        cmd = self.get_cli_command() + ["logs"]
        if tail is not None:
            cmd.extend(["--tail", str(tail)])
        cmd.append(container)
        return cmd

    def cp(self, src: str, dest: str) -> list[str]:
        """Return cp command."""
        return self.get_cli_command() + ["cp", src, dest]

    def info(self) -> list[str]:
        """Return info command."""
        return self.get_cli_command() + ["info"]

    def version(self, format_str: Optional[str] = None) -> list[str]:
        """Return version command. format_str adds --format flag."""
        cmd = self.get_cli_command() + ["version"]
        if format_str is not None:
            cmd.extend(["--format", format_str])
        return cmd

    def network_prune(self, force: bool = True) -> list[str]:
        """Return network prune command."""
        cmd = self.get_cli_command() + ["network", "prune"]
        if force:
            cmd.append("-f")
        return cmd

    def container_prune(self, force: bool = True) -> list[str]:
        """Return container prune command."""
        cmd = self.get_cli_command() + ["container", "prune"]
        if force:
            cmd.append("-f")
        return cmd

    def image_prune(self, force: bool = True) -> list[str]:
        """Return image prune command."""
        cmd = self.get_cli_command() + ["image", "prune"]
        if force:
            cmd.append("-f")
        return cmd

    def system_prune(self, all: bool = False, volumes: bool = False) -> list[str]:
        """Return system prune command. all=True prunes all images, volumes=True includes volumes."""
        cmd = self.get_cli_command() + ["system", "prune"]
        if all and volumes:
            cmd.extend(["-af", "--volumes"])
        elif all:
            cmd.append("-af")
        else:
            cmd.append("-f")
        return cmd

    def builder_prune(self, all: bool = False) -> list[str]:
        """Return builder prune command. all=True prunes all build cache."""
        cmd = self.get_cli_command() + ["builder", "prune"]
        if all:
            cmd.append("-af")
        return cmd

    def command(self, args: list[str]) -> list[str]:
        """Generic fallback: build any runtime command."""
        return self.get_cli_command() + args


def get_runtime() -> ContainerRuntime:
    """Get or create the singleton ContainerRuntime instance."""
    global _runtime
    if _runtime is None:
        _runtime = ContainerRuntime(detect_runtime())
    return _runtime


def reset_runtime() -> None:
    """Reset the singleton instance (for testing)."""
    global _runtime
    _runtime = None
