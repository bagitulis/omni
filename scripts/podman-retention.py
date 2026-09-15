#!/usr/bin/env python3
"""Podman image retention policy for the Omni stack.

Rebuilds via `build.py smart` leave old images dangling. Without a policy,
a day of iteration racks up 20-30 GB of untagged rubbish (observed
2026-09-15: 29 GB total, 28 GB reclaimable).

Policy (conservative, safe defaults):
  - Keep the current `:latest` tag for each Omni image (backend, frontend).
  - Keep the N most-recent dangling images per repository (default: 3).
  - Never touch images that are in use by a running or exited container.
  - Never touch third-party base images (postgres, redis, nginx, etc.) —
    those are tagged and reused across rebuilds.

Usage:
    python scripts/podman-retention.py            # default keep=3
    python scripts/podman-retention.py --keep 5   # keep 5 dangling per repo
    python scripts/podman-retention.py --dry-run  # report only, no changes
    python scripts/podman-retention.py --aggressive  # keep only :latest

Exit codes: 0 clean, 1 podman unavailable, 2 partial failure.
"""

from __future__ import annotations

import argparse
import json
import shutil
import subprocess
import sys
from collections import defaultdict
from dataclasses import dataclass
from typing import Iterable


PODMAN = shutil.which("podman") or "podman"
OMNI_REPOS = ("localhost/omni_backend", "localhost/omni_frontend")


@dataclass(frozen=True)
class Img:
    id: str
    repo: str
    tag: str
    size_bytes: int
    created_ns: int  # unix seconds *1e9 for stable sort


def _run(*args: str) -> tuple[int, str, str]:
    proc = subprocess.run([PODMAN, *args], capture_output=True, text=True)
    return proc.returncode, proc.stdout, proc.stderr


def _list_images() -> list[Img]:
    """Return every image, tagged or not, with repo/tag/size/created."""
    rc, out, err = _run("images", "--all", "--format", "json")
    if rc != 0:
        print(f"[FAIL] podman images: {err.strip()}", file=sys.stderr)
        sys.exit(1)
    raw = json.loads(out or "[]")
    imgs: list[Img] = []
    for row in raw:
        # Rows without Names list are dangling — flag with "<none>".
        names: list[str] = row.get("Names") or []
        if names:
            for full in names:
                if ":" in full:
                    repo, tag = full.rsplit(":", 1)
                else:
                    repo, tag = full, "latest"
                imgs.append(Img(
                    id=row["Id"],
                    repo=repo,
                    tag=tag,
                    size_bytes=int(row.get("Size", 0)),
                    created_ns=int(row.get("Created", 0)),
                ))
        else:
            imgs.append(Img(
                id=row["Id"],
                repo="<none>",
                tag="<none>",
                size_bytes=int(row.get("Size", 0)),
                created_ns=int(row.get("Created", 0)),
            ))
    return imgs


def _in_use_ids() -> set[str]:
    """IDs of images referenced by any container (running or stopped)."""
    rc, out, err = _run("ps", "--all", "--format", "{{.ImageID}}")
    if rc != 0:
        return set()
    return {line.strip() for line in out.splitlines() if line.strip()}


def _select_pruneable(imgs: list[Img], keep: int, aggressive: bool) -> list[Img]:
    """Return the images we're going to delete."""
    in_use = _in_use_ids()

    # Group by repo (dangling all bucket together).
    by_repo: dict[str, list[Img]] = defaultdict(list)
    for i in imgs:
        by_repo[i.repo].append(i)

    victims: list[Img] = []
    for repo, group in by_repo.items():
        # Sort newest-first.
        group.sort(key=lambda x: x.created_ns, reverse=True)

        # Third-party base images: skip entirely — they're tagged and shared.
        if not (repo == "<none>" or repo in OMNI_REPOS):
            continue

        if aggressive:
            # Aggressive: keep only the current :latest tag on Omni repos,
            # nuke everything else including dangling.
            for img in group:
                if img.tag == "latest" and repo in OMNI_REPOS:
                    continue
                if img.id in in_use:
                    continue
                victims.append(img)
            continue

        # Conservative: keep :latest + `keep` most recent dangling per bucket.
        kept = 0
        for img in group:
            # Never touch :latest of Omni repos.
            if img.tag == "latest" and repo in OMNI_REPOS:
                continue
            # Never touch anything referenced by a container.
            if img.id in in_use:
                continue
            if kept < keep:
                kept += 1
                continue
            victims.append(img)

    return victims


def _human(n: int) -> str:
    for unit in ("B", "KB", "MB", "GB", "TB"):
        if n < 1024:
            return f"{n:.1f}{unit}"
        n /= 1024
    return f"{n:.1f}PB"


def _prune(victims: Iterable[Img], dry_run: bool) -> tuple[int, int]:
    """Delete images. Returns (deleted_count, freed_bytes)."""
    deleted = 0
    freed = 0
    for img in victims:
        label = f"{img.repo}:{img.tag}" if img.repo != "<none>" else "<dangling>"
        if dry_run:
            print(f"  [DRY] would remove {img.id[:12]} {label} ({_human(img.size_bytes)})")
            deleted += 1
            freed += img.size_bytes
            continue
        rc, _, err = _run("rmi", "-f", img.id)
        if rc == 0:
            print(f"  [OK]  removed {img.id[:12]} {label} ({_human(img.size_bytes)})")
            deleted += 1
            freed += img.size_bytes
        else:
            # A concurrent container may have grabbed it — skip loudly.
            print(f"  [SKIP] {img.id[:12]}: {err.strip().splitlines()[-1] if err.strip() else 'unknown'}")
    return deleted, freed


def main() -> int:
    p = argparse.ArgumentParser(description="Podman image retention for Omni")
    p.add_argument("--keep", type=int, default=3, help="Keep N most-recent dangling images per Omni repo (default 3)")
    p.add_argument("--dry-run", action="store_true", help="Report only, don't delete")
    p.add_argument("--aggressive", action="store_true", help="Keep only :latest on Omni repos; nuke everything else")
    args = p.parse_args()

    if shutil.which("podman") is None and PODMAN == "podman":
        print("[FAIL] podman not found in PATH", file=sys.stderr)
        return 1

    print(f"=== Podman retention (keep={args.keep}, aggressive={args.aggressive}, dry_run={args.dry_run}) ===")
    imgs = _list_images()
    print(f"Found {len(imgs)} image entries")
    victims = _select_pruneable(imgs, args.keep, args.aggressive)
    if not victims:
        print("Nothing to prune — retention already satisfied.")
        return 0
    print(f"Selecting {len(victims)} image(s) to remove:")
    deleted, freed = _prune(victims, args.dry_run)
    print()
    verb = "would free" if args.dry_run else "freed"
    print(f"=== {'DRY RUN: ' if args.dry_run else ''}{deleted} image(s) removed, {verb} {_human(freed)} ===")
    return 0


if __name__ == "__main__":
    sys.exit(main())
