"""Git LFS helpers for NDJSON sync imports.

SRP: this module only detects Git LFS pointer files and materializes them.
"""
import subprocess
from collections.abc import Sequence
from os import environ
from pathlib import Path

from omni_build.logger import log_error, log_fix, log_info, log_success

LFS_POINTER_PREFIX = b"version https://git-lfs"


def is_lfs_pointer(filepath: Path) -> bool:
    """Return True when filepath contains an unresolved Git LFS pointer."""
    with open(filepath, "rb") as file_handle:
        return file_handle.read(30).startswith(LFS_POINTER_PREFIX)


def find_lfs_pointers(sync_dir: Path) -> list[Path]:
    """Find unresolved Git LFS pointer files in the NDJSON large-file directory."""
    large_dir = sync_dir / "_large"
    if not large_dir.exists():
        return []

    return [filepath for filepath in large_dir.glob("*.ndjson.gz") if is_lfs_pointer(filepath)]


def materialize_lfs_pointers(project_root: Path, sync_dir: Path) -> bool:
    """Auto-fix unresolved Git LFS pointers before NDJSON import starts."""
    pointer_files = find_lfs_pointers(sync_dir)
    if not pointer_files:
        return True

    log_fix(f"Git LFS pointer detected in {len(pointer_files)} NDJSON large file(s)")
    _ensure_lfs_attributes(project_root)
    include_paths = [_to_git_path(filepath, project_root) for filepath in pointer_files]

    if not _run_git_lfs_install(project_root):
        return False

    if not _run_git_lfs_pull(project_root, include_paths):
        return False

    remaining = find_lfs_pointers(sync_dir)
    if remaining:
        for filepath in remaining:
            log_error(f"Git LFS pointer still unresolved: {filepath.name}")
        log_error("Auto-fix failed. Run: git lfs install && git lfs pull")
        return False

    log_success("Git LFS files materialized successfully")
    return True


def _to_git_path(filepath: Path, project_root: Path) -> str:
    """Convert a filesystem path to a repo-relative Git path with forward slashes."""
    return filepath.resolve().relative_to(project_root.resolve()).as_posix()


def _ensure_lfs_attributes(project_root: Path) -> None:
    """Ensure NDJSON large-file backups have a Git LFS attribute rule."""
    attributes_path = project_root / ".gitattributes"
    rule = "backups/sync/_large/*.ndjson.gz filter=lfs diff=lfs merge=lfs -text"
    if not attributes_path.exists():
        _ = attributes_path.write_text(rule + "\n", encoding="utf-8")
        log_fix("Created .gitattributes rule for NDJSON Git LFS files")
        return

    content = attributes_path.read_text(encoding="utf-8")
    if rule in content.splitlines():
        return

    separator = "" if content.endswith("\n") else "\n"
    with open(attributes_path, "a", encoding="utf-8") as file_handle:
        _ = file_handle.write(separator + rule + "\n")
    log_fix("Added .gitattributes rule for NDJSON Git LFS files")


def _run_git_lfs_install(project_root: Path) -> bool:
    """Install Git LFS filters for this repository."""
    log_info("Preparing Git LFS for this repository...")
    return _run_git_command(project_root, ["lfs", "install", "--local"], timeout=60)


def _run_git_lfs_pull(project_root: Path, include_paths: Sequence[str]) -> bool:
    """Pull the specific missing LFS objects instead of the whole LFS history."""
    include_arg = "--include=" + ",".join(include_paths)
    log_info(f"Pulling {len(include_paths)} Git LFS object(s)...")
    return _run_git_command(project_root, ["lfs", "pull", include_arg], timeout=180)


def _run_git_command(project_root: Path, args: Sequence[str], timeout: int) -> bool:
    """Run a Git command safely without shell expansion."""
    try:
        result = subprocess.run(
            ["git", *args],
            cwd=project_root,
            capture_output=True,
            env=_git_env(),
            text=True,
            timeout=timeout,
            encoding="utf-8",
            errors="replace",
        )
    except FileNotFoundError:
        log_error("Git executable not found; install Git and Git LFS, then retry sync import")
        return False
    except subprocess.TimeoutExpired:
        log_error(f"Git command timed out: git {' '.join(args)}")
        return False

    if result.returncode == 0:
        return True

    output = (result.stderr or result.stdout or "unknown error").strip()
    log_error(f"Git command failed: git {' '.join(args)}")
    log_error(output[:500])
    return False


def _git_env() -> dict[str, str]:
    """Return environment variables that keep Git LFS auto-fix non-interactive."""
    env = dict(environ)
    env["GIT_TERMINAL_PROMPT"] = "0"
    env["GCM_INTERACTIVE"] = "never"
    return env
