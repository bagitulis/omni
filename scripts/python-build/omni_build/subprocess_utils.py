"""
Subprocess utilities with UTF-8 encoding support for Windows.

This module provides wrappers around subprocess functions that ensure
UTF-8 encoding is used instead of the Windows default (cp1252), preventing
UnicodeDecodeError when Docker or other tools output non-ASCII characters.
"""
import os
import subprocess
from typing import Any, Optional
from omni_build.container_runtime import get_runtime


def _real_home():
    try:
        import pwd
        from pathlib import Path
        return Path(pwd.getpwuid(os.getuid()).pw_dir)
    except (ImportError, KeyError):
        from pathlib import Path
        return Path.home()


def run(
    args: list[str],
    capture_output: bool = False,
    text: bool = False,
    timeout: Optional[float] = None,
    cwd: Optional[str] = None,
    check: bool = False,
    shell: bool = False,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """
    Run a subprocess command with UTF-8 encoding on Windows.
    
    This is a wrapper around subprocess.run() that automatically sets
    encoding='utf-8' and errors='replace' when text=True is used,
    preventing UnicodeDecodeError on Windows.
    
    Args:
        args: Command and arguments to run
        capture_output: If True, capture stdout and stderr
        text: If True, decode stdout/stderr as text (with UTF-8 encoding)
        timeout: Timeout in seconds
        cwd: Working directory
        check: If True, raise CalledProcessError on non-zero exit
        shell: If True, run through shell
        **kwargs: Additional arguments passed to subprocess.run()
    
    Returns:
        CompletedProcess instance
    """
    # When text=True is requested, use explicit UTF-8 encoding
    # with 'replace' error handling to avoid UnicodeDecodeError
    if text:
        kwargs.setdefault('encoding', 'utf-8')
        kwargs.setdefault('errors', 'replace')
    
    return subprocess.run(
        args,
        capture_output=capture_output,
        text=text,
        timeout=timeout,
        cwd=cwd,
        check=check,
        shell=shell,
        **kwargs,
    )


def run_utf8(
    args: list[str],
    timeout: Optional[float] = None,
    cwd: Optional[str] = None,
    check: bool = False,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """
    Convenience function to run a command and capture output as UTF-8 text.
    
    Equivalent to:
        subprocess.run(args, capture_output=True, text=True, 
                       encoding='utf-8', errors='replace', ...)
    
    Args:
        args: Command and arguments to run
        timeout: Timeout in seconds
        cwd: Working directory
        check: If True, raise CalledProcessError on non-zero exit
        **kwargs: Additional arguments passed to subprocess.run()
    
    Returns:
        CompletedProcess instance with stdout/stderr as strings
    """
    return run(
        args,
        capture_output=True,
        text=True,
        timeout=timeout,
        cwd=cwd,
        check=check,
        **kwargs,
    )


def run_silent(
    args: list[str],
    timeout: Optional[float] = None,
    cwd: Optional[str] = None,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """
    Run a command silently, suppressing output.
    
    Args:
        args: Command and arguments to run
        timeout: Timeout in seconds
        cwd: Working directory
        **kwargs: Additional arguments passed to subprocess.run()
    
    Returns:
        CompletedProcess instance
    """
    return run(
        args,
        capture_output=True,
        text=True,
        timeout=timeout,
        cwd=cwd,
        check=False,
        **kwargs,
    )


def get_compose_command() -> list[str]:
    """
    Get the correct compose command for this system.

    Tries the unified container_runtime (Podman + Docker) first,
    then falls back to legacy Docker-only detection for backward
    compatibility.

    Returns:
        A list of strings forming the compose command.

    Raises:
        RuntimeError if no compose variant is found.
    """
    # The unified runtime is the source of truth. If the active runtime is
    # podman, the docker-only legacy path below cannot help and would crash
    # on hosts without a `docker` CLI (WinError 2 on Windows). Propagate the
    # runtime's RuntimeError so callers see a clear "install podman-compose"
    # message instead of a mysterious FileNotFoundError.
    try:
        runtime = get_runtime()
    except RuntimeError:
        runtime = None
    if runtime is not None:
        try:
            return runtime.get_compose_command()
        except RuntimeError:
            if runtime.name == "podman":
                raise

    # Fall back to legacy Docker-only detection
    import os
    import platform
    import shutil
    import stat
    import urllib.error
    import urllib.request
    from pathlib import Path

    # If docker itself is missing, skip the legacy path entirely. Every helper
    # below shells out to `docker ...`, which raises FileNotFoundError on
    # Windows when docker.exe is absent.
    if not shutil.which("docker"):
        raise RuntimeError(
            "Docker Compose is not available and the `docker` CLI was not "
            "found on PATH. Install Docker Desktop, or set "
            "CONTAINER_RUNTIME=podman and install podman-compose."
        )

    def repair_docker_cli_config() -> None:
        config_path = _real_home() / ".docker" / "config.json"
        if not config_path.exists():
            return
        try:
            import json

            data = json.loads(config_path.read_text())
            changed = False
            creds_store = data.get("credsStore")
            if creds_store and not shutil.which(f"docker-credential-{creds_store}"):
                data.pop("credsStore", None)
                changed = True
            if data.get("features", {}).get("hooks") == "true":
                data.setdefault("features", {})["hooks"] = "false"
                changed = True
            if changed:
                backup = config_path.with_suffix(".json.autofix-backup")
                if not backup.exists():
                    backup.write_text(config_path.read_text())
                config_path.write_text(json.dumps(data, indent=2) + "\n")
        except (OSError, ValueError):
            return

    def ensure_buildx_plugin() -> None:
        if subprocess.run(["docker", "buildx", "version"], capture_output=True, text=True, timeout=10).returncode == 0:
            return
        if system != "linux" or not buildx_arch:
            return
        plugin_path = _real_home() / ".docker" / "cli-plugins" / "docker-buildx"
        try:
            if plugin_path.is_symlink() and not plugin_path.exists():
                plugin_path.unlink()
            api = "https://api.github.com/repos/docker/buildx/releases/latest"
            import json

            release = json.loads(urllib.request.urlopen(api, timeout=30).read().decode("utf-8"))
            suffix = f"linux-{buildx_arch}"
            url = next(
                asset["browser_download_url"]
                for asset in release["assets"]
                if asset["name"].endswith(suffix)
            )
            plugin_path.parent.mkdir(parents=True, exist_ok=True)
            with urllib.request.urlopen(url, timeout=120) as response:
                plugin_path.write_bytes(response.read())
            plugin_path.chmod(plugin_path.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
        except (OSError, StopIteration, KeyError, ValueError, urllib.error.URLError, TimeoutError):
            return

    system = platform.system().lower()
    machine = platform.machine().lower()
    arch_map = {
        "x86_64": "x86_64",
        "amd64": "x86_64",
        "aarch64": "aarch64",
        "arm64": "aarch64",
    }
    buildx_arch_map = {"x86_64": "amd64", "amd64": "amd64", "aarch64": "arm64", "arm64": "arm64"}
    arch = arch_map.get(machine)
    buildx_arch = buildx_arch_map.get(machine)

    repair_docker_cli_config()
    ensure_buildx_plugin()

    # Prefer standalone docker-compose (backward compat).
    if shutil.which("docker-compose"):
        return ["docker-compose"]

    def has_compose_plugin() -> bool:
        result = subprocess.run(
            ["docker", "compose", "version"],
            capture_output=True,
            text=True,
            timeout=10,
        )
        return result.returncode == 0

    if has_compose_plugin():
        return ["docker", "compose"]

    system = platform.system().lower()
    machine = platform.machine().lower()
    arch_map = {
        "x86_64": "x86_64",
        "amd64": "x86_64",
        "aarch64": "aarch64",
        "arm64": "aarch64",
    }
    arch = arch_map.get(machine)

    if system == "linux" and arch:
        plugin_dir = _real_home() / ".docker" / "cli-plugins"
        plugin_path = plugin_dir / "docker-compose"
        url = f"https://github.com/docker/compose/releases/latest/download/docker-compose-linux-{arch}"
        try:
            plugin_dir.mkdir(parents=True, exist_ok=True)
            if plugin_path.is_symlink() and not plugin_path.exists():
                plugin_path.unlink()
            with urllib.request.urlopen(url, timeout=60) as response:
                plugin_path.write_bytes(response.read())
            plugin_path.chmod(plugin_path.stat().st_mode | stat.S_IXUSR | stat.S_IXGRP | stat.S_IXOTH)
            os.environ["PATH"] = f"{plugin_dir}{os.pathsep}" + os.environ.get("PATH", "")
            if has_compose_plugin():
                return ["docker", "compose"]
        except (OSError, urllib.error.URLError, TimeoutError, subprocess.SubprocessError):
            pass

    raise RuntimeError(
        "Docker Compose is not available. Install docker-compose-plugin or ensure "
        "`docker compose version` works, then rerun build.py."
    )

