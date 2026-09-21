#!/usr/bin/env python3
"""WSL VHDX health: detect stranded space and explain how to reclaim it.

Why this exists (2026-09-21 storage audit):

Deleting images, containers and volumes inside the Podman VM does NOT shrink
`ext4.vhdx` on the Windows host. The file is not sparse, so Windows keeps every
allocated byte. Measured live: the VHDX was 16.73 GB while the VM used only
7.10 GB — 9.63 GB stranded and invisible to every in-VM prune.

Consequence: `build.py` could prune perfectly and never return a single byte to
the host. The recovery path (`scripts/compact-podman-vhd.ps1`, or
`wsl --manage <distro> --set-sparse true --allow-unsafe`) needs Administrator
rights, so it can never run silently during a build. This module supplies the
decision logic so build.py can *tell* the operator when it is worth doing.

Pure functions only — nothing here elevates, compacts, or mutates state.

Usage (from build.py):
    from vhdx_health import assess, render_reminder, vhdx_path_for
"""
from __future__ import annotations

import shutil
import subprocess
import sys
from dataclasses import dataclass
from pathlib import Path

# Nag only when at least this much is stranded; an elevated compaction is not
# worth prompting for a few hundred MB.
DEFAULT_THRESHOLD_BYTES = 1 * 1024 ** 3  # 1 GiB

DEFAULT_DISTRO = "podman-machine-default"

_GIB = 1024 ** 3


@dataclass(frozen=True)
class VhdxHealth:
    """Result of comparing the VHDX file size against real VM usage."""

    file_bytes: int
    used_bytes: int
    dead_bytes: int
    threshold_bytes: int
    needs_compaction: bool


def _human(n: int) -> str:
    value = float(n)
    for unit in ("B", "KB", "MB", "GB", "TB"):
        if value < 1024:
            return f"{value:.2f}{unit}"
        value /= 1024
    return f"{value:.2f}PB"


def assess(
    file_bytes: int,
    used_bytes: int,
    threshold_bytes: int = DEFAULT_THRESHOLD_BYTES,
) -> VhdxHealth:
    """Compare allocated VHDX size with bytes the guest filesystem uses.

    Dead space is clamped at zero: a sparse or mid-optimisation VHDX can
    legitimately report a file size below current usage, which is not waste.
    """
    dead = max(0, file_bytes - used_bytes)
    return VhdxHealth(
        file_bytes=file_bytes,
        used_bytes=used_bytes,
        dead_bytes=dead,
        threshold_bytes=threshold_bytes,
        # Strictly greater: exactly at the threshold is not yet worth an
        # elevated action.
        needs_compaction=dead > threshold_bytes,
    )


def render_reminder(health: VhdxHealth, distro: str = DEFAULT_DISTRO) -> str:
    """Return the operator-facing reminder, or "" when nothing is stranded."""
    if not health.needs_compaction:
        return ""
    rule = "=" * 62
    # Build with explicit lines: adjacent string literals next to a
    # multiplication ("=" * 62 + "\n" "...") silently duplicate the body.
    lines = [
        "",
        rule,
        "  [STORAGE] Podman VM disk has stranded space",
        rule,
        f"  VHDX file size : {_human(health.file_bytes)}",
        f"  Actually used  : {_human(health.used_bytes)}",
        f"  Stranded       : {_human(health.dead_bytes)}",
        "",
        "  Pruning inside the VM cannot return this space: ext4.vhdx is not",
        "  sparse, so Windows keeps it allocated.",
        "",
        "  RECOMMENDED (safe) - from an ADMIN PowerShell:",
        "    scripts/compact-podman-vhd.ps1",
        "",
        "  ALTERNATIVE (instant, but UNSAFE):",
        f"    wsl --manage {distro} --set-sparse true --allow-unsafe",
        "    WARNING: WSL disables sparse VHD by default due to potential DATA",
        "    CORRUPTION. Use the alternative only if you accept that risk.",
        "",
        "  Administrator rights are required, so this cannot run automatically.",
        rule,
        "",
    ]
    return "\n".join(lines)


def vhdx_path_for(distro: str = DEFAULT_DISTRO, home: Path | None = None) -> Path:
    """Return the on-disk VHDX path for a Podman WSL machine."""
    base = home if home is not None else Path.home()
    return (
        base / ".local" / "share" / "containers" / "podman" / "machine"
        / "wsl" / "wsldist" / distro / "ext4.vhdx"
    )


def _vm_used_bytes() -> int | None:
    """Bytes used inside the Podman VM, or None if it cannot be determined."""
    podman = shutil.which("podman")
    if not podman:
        return None
    try:
        result = subprocess.run(
            [podman, "machine", "ssh", "df -B1 --output=used / | tail -1"],
            capture_output=True, text=True, timeout=60,
        )
    except (subprocess.TimeoutExpired, OSError):
        return None
    if result.returncode != 0:
        return None
    try:
        return int(result.stdout.strip().splitlines()[-1].strip())
    except (ValueError, IndexError):
        return None


def check_current(
    distro: str = DEFAULT_DISTRO,
    threshold_bytes: int = DEFAULT_THRESHOLD_BYTES,
) -> VhdxHealth | None:
    """Assess the live VHDX. Returns None if the VM is stopped or unavailable.

    A stopped machine is reported as None rather than "healthy": its file size
    yields no signal about the guest filesystem.
    """
    path = vhdx_path_for(distro)
    if not path.exists():
        return None
    used = _vm_used_bytes()
    if used is None:
        return None
    return assess(path.stat().st_size, used, threshold_bytes)


def main() -> int:
    health = check_current()
    if health is None:
        print("VHDX health: unavailable (machine stopped, or not a WSL Podman setup)")
        return 0
    print(f"VHDX: {_human(health.file_bytes)} file vs {_human(health.used_bytes)} used")
    print(f"Stranded: {_human(health.dead_bytes)}")
    reminder = render_reminder(health)
    if reminder:
        print(reminder)
    else:
        print("No compaction needed.")
    return 0


def is_elevated() -> bool:
    """True when the current process has Administrator rights (Windows)."""
    if sys.platform != "win32":
        return False
    try:
        import ctypes

        return bool(ctypes.windll.shell32.IsUserAnAdmin())
    except Exception:
        return False


def run_compaction(
    project_root: Path,
    distro: str = DEFAULT_DISTRO,
    no_restart: bool = True,
) -> bool:
    """Attempt VHDX compaction, degrading cleanly when not elevated.

    Compaction requires Administrator (Optimize-VHD / diskpart), so this can
    never succeed silently from an ordinary build shell. Instead of failing the
    build, it prints the exact elevated command the operator must run.

    `no_restart` defaults to True: the script otherwise starts the VM and all
    omni-* containers afterwards, which surprises an operator who had the
    machine stopped (observed live 2026-09-21).

    Returns True only when compaction actually ran and reported success.
    """
    script = project_root / "scripts" / "compact-podman-vhd.ps1"

    if sys.platform != "win32":
        print("  [SKIP] VHDX compaction is a Windows/WSL-only operation.")
        return False

    if not is_elevated():
        print("  [SKIP] Not running as Administrator.")
        print("         Run in an ADMIN PowerShell (safe, offline compaction):")
        print(f"           {script}")
        print("         Alternative (instant, but UNSAFE - WSL disables sparse")
        print("         VHD by default due to potential data corruption):")
        print(f"           wsl --manage {distro} --set-sparse true --allow-unsafe")
        return False

    if not script.exists():
        print(f"  [WARN] Compaction script not found: {script}")
        return False

    # Always pass through: the build shell must not silently resurrect a
    # machine the operator stopped.
    cmd = ["pwsh", "-NoProfile", "-File", str(script)]
    if no_restart:
        cmd.append("-NoRestart")

    print("  [VHDX] Compacting (this takes 2-5 minutes)...")
    result = subprocess.run(cmd, check=False)
    if result.returncode == 0:
        print("  [OK] VHDX compaction finished.")
        return True
    print(f"  [WARN] Compaction exited with code {result.returncode}")
    return False


def report_and_maybe_compact(project_root: Path, compact_requested: bool) -> None:
    """Print VHDX health; compact only when explicitly requested AND elevated."""
    try:
        health = check_current()
    except Exception as exc:  # never let a health probe fail a build
        print(f"[WARN] VHDX health check skipped: {exc}")
        return

    if health is None:
        return

    reminder = render_reminder(health)
    if not reminder:
        return

    print(reminder)
    if compact_requested:
        run_compaction(project_root)
    else:
        print("  (re-run with --compact-vhdx to attempt it automatically "
              "when elevated)\n")


if __name__ == "__main__":
    sys.exit(main())
