#!/usr/bin/env python3
"""Podman retention CLI for the Omni stack.

This is a thin wrapper: the policy lives in
scripts/python-build/omni_build/podman_retention.py and the typed podman
inventory in omni_build/podman_inventory.py.

Why retention has more than one phase (2026-09-21 storage audit):
an image-only policy reported "nothing to prune" while the machine actually
held 17 leaked testcontainers containers and 52 orphaned anonymous volumes.
Image pruning skips any image a container references, so leaked containers
pinned their images as "in use" and no space was ever reclaimed. The WSL VM
disk reached 94% full with ~9.6 GB stranded in ext4.vhdx.

Phases (order matters — see run_retention):
  1. Remove STOPPED containers labelled `org.testcontainers=true` only.
     Compose-managed omni-* and unlabelled containers are never touched.
  2. Remove anonymous volumes (64-hex names) that no SURVIVING container
     references. Named volumes are never touched — `omni_omni-pgdata` IS the
     database.
  3. Prune dangling/stale images, now that phase 1 has unpinned them.

Reclaiming the freed bytes on the Windows host additionally requires compaction
(`scripts/compact-podman-vhd.ps1`, elevated) because ext4.vhdx is not sparse.

Usage:
    python scripts/podman-retention.py                 # default keep=3
    python scripts/podman-retention.py --keep 5        # keep 5 dangling per repo
    python scripts/podman-retention.py --dry-run       # report only, no changes
    python scripts/podman-retention.py --aggressive    # keep only :latest
    python scripts/podman-retention.py --skip-volumes  # containers+images only

Exit codes: 0 clean, 1 podman unavailable, 2 partial failure.
"""
from __future__ import annotations

import argparse
import sys
from pathlib import Path

# The omni_build package lives beside scripts/, mirroring how build.py imports it.
_PYTHON_BUILD = Path(__file__).resolve().parent / "python-build"
if str(_PYTHON_BUILD) not in sys.path:
    sys.path.insert(0, str(_PYTHON_BUILD))

from omni_build import podman_retention  # noqa: E402


def main() -> int:
    parser = argparse.ArgumentParser(description="Podman retention for Omni")
    parser.add_argument(
        "--keep", type=int, default=3,
        help="Keep N most-recent dangling images per Omni repo (default 3)",
    )
    parser.add_argument(
        "--dry-run", action="store_true", help="Report only, don't delete",
    )
    parser.add_argument(
        "--aggressive", action="store_true",
        help="Keep only :latest on Omni repos; nuke everything else",
    )
    parser.add_argument(
        "--skip-volumes", action="store_true",
        help="Do not remove orphaned anonymous volumes",
    )
    args = parser.parse_args()

    return podman_retention.run_retention(
        keep=args.keep,
        dry_run=args.dry_run,
        aggressive=args.aggressive,
        skip_volumes=args.skip_volumes,
    )


if __name__ == "__main__":
    sys.exit(main())
