"""
Docker Desktop Management module for Omni Build System.

SRP: This module ONLY handles Docker Desktop process management.
Cross-platform: handles Windows (Docker Desktop.exe) and Linux (docker daemon/systemd).
"""
import os
import platform
import shutil
import subprocess
import time
from pathlib import Path
from typing import Optional
from omni_build.container_runtime import get_runtime


def _real_home():
    try:
        import pwd
        return Path(pwd.getpwuid(os.getuid()).pw_dir)
    except (ImportError, KeyError):
        return Path.home()

from omni_build.logger import log_error, log_info, log_success, log_warning

IS_WINDOWS = platform.system() == "Windows"
IS_LINUX = platform.system() == "Linux"


class DockerDesktopManager:
    """Manages Docker Desktop process (Windows) or Docker daemon (Linux)."""

    def is_running(self) -> bool:
        """
        Check if Docker is running.

        Windows: checks for Docker Desktop.exe process
        Linux: checks if docker daemon is responding
        """
        if IS_WINDOWS:
            return self._is_running_windows()
        return self._is_running_linux()

    def _is_daemon_running_but_no_perms(self) -> bool:
        """Check if Docker daemon is running but user lacks socket permissions.
        
        Uses sudo to check if the daemon is actually alive, to distinguish between
        'docker not installed/running' vs 'docker running but no group membership'.
        """
        try:
            result = subprocess.run(
                ["sudo", "docker", "info"],
                capture_output=True, text=True, timeout=10,
            )
            return result.returncode == 0
        except Exception:
            return False

    def _is_running_windows(self) -> bool:
        """Check if Docker or Podman is running on Windows."""
        # Try runtime abstraction first (works for both Docker and Podman)
        try:
            if get_runtime().is_running():
                log_info(f"{get_runtime().name} is running (via runtime check)")
                return True
        except Exception:
            pass

        # Fallback: check Docker Desktop.exe process directly
        try:
            result = subprocess.run(
                ["tasklist", "/FO", "CSV", "/NH"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                check=False
            )

            if "Docker Desktop.exe" in result.stdout:
                log_info("Docker Desktop process is running")
                return True

            log_info("No container runtime detected on Windows")
            return False

        except Exception as e:
            log_warning(f"Could not check container runtime process: {e}")
            return False

    def _is_running_linux(self) -> bool:
        """Check if container daemon is running on Linux."""
        try:
            result = subprocess.run(
                get_runtime().info(),
                capture_output=True,
                text=True,
                timeout=10,
            )
            if result.returncode == 0:
                log_info(f"{get_runtime().name} daemon is running")
                return True
            # Detect permission denied vs daemon not running
            stderr_lower = (result.stderr or "").lower()
            if "permission denied" in stderr_lower:
                log_warning("Permission denied — user not in docker/podman group")
            else:
                log_info(f"{get_runtime().name} daemon not responding")
            return False
        except Exception as e:
            log_warning(f"Could not check container daemon: {e}")
            return False

    def start(self) -> bool:
        """
        Start Docker.

        Windows: launches Docker Desktop.exe from known paths
        Linux: starts docker via systemctl or checks if already running

        Returns:
            True if started successfully
        """
        if IS_WINDOWS:
            return self._start_windows()
        return self._start_linux()

    def _start_windows(self) -> bool:
        """Start Docker or Podman on Windows."""
        runtime = get_runtime()

        # Podman: use podman machine start
        if runtime.name == "podman":
            log_info("Starting Podman machine...")
            try:
                result = subprocess.run(
                    ["podman", "machine", "start"],
                    capture_output=True,
                    text=True,
                    timeout=120,
                )
                if result.returncode == 0:
                    time.sleep(3)
                    if runtime.is_running():
                        log_success("Podman machine started")
                        return True
                log_warning("Podman machine start did not confirm running state")
            except Exception as e:
                log_error(f"Failed to start Podman machine: {e}")

        # Docker Desktop fallback
        docker_paths = [
            Path("C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe"),
            Path("C:\\Program Files (x86)\\Docker\\Docker\\Docker Desktop.exe"),
            _real_home() / "AppData" / "Local" / "Docker" / "Docker Desktop.exe",
        ]

        for docker_path in docker_paths:
            if docker_path.exists():
                log_info(f"Starting Docker Desktop from: {docker_path}")
                try:
                    subprocess.Popen(
                        [str(docker_path)],
                        stdout=subprocess.DEVNULL,
                        stderr=subprocess.DEVNULL
                    )

                    max_wait = 30
                    waited = 0
                    while waited < max_wait:
                        time.sleep(2)
                        waited += 2
                        if self.is_running():
                            time.sleep(5)
                            log_success("Docker Desktop process started")
                            return True

                    log_warning(f"Docker Desktop process not detected after {max_wait}s")
                    return False

                except Exception as e:
                    log_error(f"Failed to start Docker Desktop: {e}")
                    return False

        log_error("No container runtime executable found")
        return False

    def _start_linux(self) -> bool:
        """Start Docker daemon on Linux via systemctl."""
        # Check if already running
        if self._is_running_linux():
            return True

        # Check for permission issue first
        try:
            id_result = subprocess.run(
                ["id", "-nG"],
                capture_output=True, text=True, timeout=5,
            )
            groups = id_result.stdout.strip().split()
            if "docker" not in groups and self._is_daemon_running_but_no_perms():
                log_warning("User not in docker group. Run: sudo usermod -aG docker $USER && newgrp docker")
                return False
        except Exception:
            pass

        # Try systemctl
        try:
            log_info("Starting Docker daemon via systemctl...")
            result = subprocess.run(
                ["sudo", "systemctl", "start", "docker"],
                capture_output=True,
                text=True,
                timeout=30,
            )
            if result.returncode == 0:
                time.sleep(3)
                if self._is_running_linux():
                    log_success("Docker daemon started via systemctl")
                    return True
        except Exception as e:
            log_warning(f"systemctl start docker failed: {e}")

        # Try service command as fallback
        try:
            result = subprocess.run(
                ["sudo", "service", "docker", "start"],
                capture_output=True,
                text=True,
                timeout=30,
            )
            if result.returncode == 0:
                time.sleep(3)
                if self._is_running_linux():
                    log_success("Docker daemon started via service")
                    return True
        except Exception as e:
            log_warning(f"service docker start failed: {e}")

        log_error("Could not start Docker. Install: sudo pacman -S docker (Arch) or sudo apt install docker.io (Debian)")
        return False

    def wait_for_ready(
        self,
        timeout_seconds: int = 90,
        activity: str = "Waiting for Docker..."
    ) -> bool:
        """
        Wait for Docker to be ready with progressive backoff.

        Args:
            timeout_seconds: Maximum time to wait
            activity: Activity description for logging

        Returns:
            True if Docker became ready, False if timeout
        """
        log_info(f"{activity} (timeout: {timeout_seconds}s)")

        elapsed = 0
        interval = 3
        consecutive_ready = 0

        while elapsed < timeout_seconds:
            time.sleep(interval)
            elapsed += interval

            if elapsed > 30 and interval < 8:
                interval = 8
                log_info("Increasing check interval for efficiency...")

            try:
                result = subprocess.run(
                    get_runtime().info(),
                    capture_output=True,
                    text=True,
                    encoding='utf-8',
                    errors='replace',
                    timeout=10,
                )

                if result.returncode == 0:
                    consecutive_ready += 1
                    log_info(f"{get_runtime().name} responding ({consecutive_ready}/2 consecutive checks)")

                    if consecutive_ready >= 2:
                        log_success(f"{get_runtime().name} ready after {elapsed}s")
                        return True
                else:
                    consecutive_ready = 0

            except subprocess.TimeoutExpired:
                consecutive_ready = 0
                log_info(f"{get_runtime().name} not responding yet ({elapsed}s elapsed)...")
            except Exception as e:
                consecutive_ready = 0
                log_warning(f"Container runtime check error: {e}")

        log_error(f"Docker did not become ready after {timeout_seconds}s")
        return False

    def check_engine_ready(self) -> bool:
        """Quick check if container engine is responding."""
        try:
            result = subprocess.run(
                get_runtime().info(),
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            return result.returncode == 0
        except Exception:
            return False
