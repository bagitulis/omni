"""Podman retention policy: what may be removed, and the removal itself.

Split out of scripts/podman-retention.py. The policy lives here; the CLI in
scripts/podman-retention.py is a thin wrapper.

Policy (conservative, safe defaults):
  - Keep the current `:latest` tag for each Omni image (backend, frontend).
  - Keep the N most-recent dangling images per repository (default: 3).
  - Never touch images in use by a running or exited container.
  - Never touch third-party base images (postgres, redis, nginx, ...).
  - Never remove a running container, an unlabelled container, or a
    compose-managed omni-* container.
  - Never remove a named volume (`omni_omni-pgdata` IS the database).

Containers and volumes were added after the 2026-09-21 storage audit. The
image-only policy silently defeated itself: a live audit found 17 leaked
testcontainers containers and 52 orphaned anonymous volumes while the script
still reported "retention already satisfied". Because image pruning skips
anything a container references, every leaked container pinned its image as
"in use" and no space was reclaimed.
"""
from __future__ import annotations

import sys
from collections import defaultdict
from typing import Iterable

from omni_build.podman_inventory import (
    ANON_VOLUME_RE,
    PODMAN,
    REMOVABLE_STATES,
    TESTCONTAINERS_LABEL,
    Ctr,
    Img,
    Vol,
    human_bytes,
    in_use_image_ids,
    run_podman,
)

OMNI_REPOS = ("localhost/omni_backend", "localhost/omni_frontend")


def select_pruneable_images(
    imgs: list[Img],
    keep: int,
    aggressive: bool,
    in_use: set[str] | None = None,
) -> list[Img]:
    """Return images eligible for removal.

    `in_use` is injectable so callers (and tests) need not probe podman.
    """
    if in_use is None:
        in_use = in_use_image_ids()

    by_repo: dict[str, list[Img]] = defaultdict(list)
    for img in imgs:
        by_repo[img.repo].append(img)

    victims: list[Img] = []
    for repo, group in by_repo.items():
        group.sort(key=lambda x: x.created_ns, reverse=True)

        # Third-party base images are tagged and shared across rebuilds.
        if not (repo == "<none>" or repo in OMNI_REPOS):
            continue

        if aggressive:
            for img in group:
                if img.tag == "latest" and repo in OMNI_REPOS:
                    continue
                if img.id in in_use:
                    continue
                victims.append(img)
            continue

        kept = 0
        for img in group:
            if img.tag == "latest" and repo in OMNI_REPOS:
                continue
            if img.id in in_use:
                continue
            if kept < keep:
                kept += 1
                continue
            victims.append(img)

    return victims


def select_pruneable_containers(containers: list[Ctr]) -> list[Ctr]:
    """Return stopped testcontainers containers safe to remove.

    Deliberately narrow: only containers that are (a) not running and
    (b) explicitly labelled by testcontainers. Compose-managed omni-*
    containers carry different labels and are never selected, and an
    unlabelled stopped container may be a user's manual work.
    """
    return [
        ctr for ctr in containers
        if ctr.state in REMOVABLE_STATES
        and ctr.labels.get(TESTCONTAINERS_LABEL) == "true"
    ]


def select_pruneable_volumes(volumes: list[Vol], in_use: set[str]) -> list[Vol]:
    """Return anonymous volumes that no surviving container references.

    Named volumes are never selected: `omni_omni-pgdata` holds the Postgres
    data directory, so removing it would destroy the database.
    """
    return [
        vol for vol in volumes
        if ANON_VOLUME_RE.match(vol.name) and vol.name not in in_use
    ]


def _last_error(err: str) -> str:
    return err.strip().splitlines()[-1] if err.strip() else "unknown"


def prune_images(victims: Iterable[Img], dry_run: bool) -> tuple[int, int]:
    """Delete images. Returns (deleted_count, freed_bytes)."""
    deleted = 0
    freed = 0
    for img in victims:
        label = f"{img.repo}:{img.tag}" if img.repo != "<none>" else "<dangling>"
        if dry_run:
            print(f"  [DRY] would remove {img.id[:12]} {label} ({human_bytes(img.size_bytes)})")
            deleted += 1
            freed += img.size_bytes
            continue
        rc, _, err = run_podman("rmi", "-f", img.id)
        if rc == 0:
            print(f"  [OK]  removed {img.id[:12]} {label} ({human_bytes(img.size_bytes)})")
            deleted += 1
            freed += img.size_bytes
        else:
            # A concurrent container may have grabbed it — skip loudly.
            print(f"  [SKIP] {img.id[:12]}: {_last_error(err)}")
    return deleted, freed


def prune_containers(victims: Iterable[Ctr], dry_run: bool) -> tuple[int, int]:
    """Delete containers. Returns (deleted_count, freed_bytes).

    Containers themselves free little (writable layer only); the real win is
    that removing them unpins their images so image pruning can proceed.
    """
    deleted = 0
    for ctr in victims:
        if dry_run:
            print(f"  [DRY] would remove container {ctr.id[:12]} ({ctr.name})")
            deleted += 1
            continue
        rc, _, err = run_podman("rm", "-f", ctr.id)
        if rc == 0:
            print(f"  [OK]  removed container {ctr.id[:12]} ({ctr.name})")
            deleted += 1
        else:
            print(f"  [SKIP] container {ctr.id[:12]}: {_last_error(err)}")
    return deleted, 0


def prune_volumes(victims: Iterable[Vol], dry_run: bool) -> tuple[int, int]:
    """Delete volumes. Returns (deleted_count, freed_bytes).

    Freed bytes are always 0: podman does not expose a per-volume size, so the
    caller reports the reclaimed total separately.
    """
    deleted = 0
    for vol in victims:
        if dry_run:
            print(f"  [DRY] would remove anonymous volume {vol.name[:12]}")
            deleted += 1
            continue
        rc, _, err = run_podman("volume", "rm", "-f", vol.name)
        if rc == 0:
            print(f"  [OK]  removed anonymous volume {vol.name[:12]}")
            deleted += 1
        else:
            print(f"  [SKIP] volume {vol.name[:12]}: {_last_error(err)}")
    return deleted, 0


def ensure_podman_available() -> bool:
    """Return True when the podman CLI can be resolved."""
    import shutil

    return shutil.which("podman") is not None or PODMAN != "podman"


def is_runtime_reachable() -> bool:
    """Return True when podman can actually talk to a container host.

    The CLI can be installed while the Podman machine is stopped, in which case
    every listing command fails. That is not an error worth alarming about: a
    stopped machine cannot be accumulating anything, so retention has nothing
    to do. Verified live 2026-09-21: `podman ps` exits 1 with
    "Cannot connect to Podman ... actively refused it".
    """
    rc, _, _ = run_podman("info")
    return rc == 0


def run_retention(
    keep: int,
    dry_run: bool,
    aggressive: bool,
    skip_volumes: bool,
) -> int:
    """Run all retention phases in order. Returns a process exit code.

    Order matters for correctness, not just tidiness: leaked containers must
    go before images (otherwise their images stay pinned) and the volume phase
    must evaluate references AFTER the container phase, treating the removed
    containers as already gone.
    """
    from omni_build.podman_inventory import (
        human_bytes as _human,
        list_containers,
        list_images,
        list_volumes,
        in_use_volume_names,
        system_df,
    )

    if not ensure_podman_available():
        print("[FAIL] podman not found in PATH", file=sys.stderr)
        return 1

    # A stopped machine cannot accumulate anything, so treat an unreachable
    # runtime as "nothing to do" rather than an error (this is a best-effort
    # post-build hook — see build.py).
    if not is_runtime_reachable():
        print("[SKIP] podman runtime not reachable (machine stopped?) — "
              "retention skipped.")
        return 0

    print(f"=== Podman retention (keep={keep}, aggressive={aggressive}, dry_run={dry_run}) ===")
    df = system_df()
    if df:
        print()
        print("=== podman system df ===")
        for line in df.splitlines():
            print(f"  {line}")

    total_removed = 0

    # --- Phase 1: leaked testcontainers containers -------------------------
    print()
    print("--- Phase 1: leaked testcontainers containers ---")
    containers = list_containers()
    ctr_victims = select_pruneable_containers(containers)
    if ctr_victims:
        print(f"Selecting {len(ctr_victims)} container(s) to remove:")
        removed, _ = prune_containers(ctr_victims, dry_run)
        total_removed += removed
    else:
        print("No leaked testcontainers containers found.")

    # --- Phase 2: orphaned anonymous volumes ------------------------------
    print()
    if skip_volumes:
        print("--- Phase 2: orphaned anonymous volumes (skipped) ---")
    else:
        print("--- Phase 2: orphaned anonymous volumes ---")
        # Survivors = every container except the ones just removed. In
        # dry-run nothing was removed, so compute the same set.
        removed_ids = {c.id for c in ctr_victims}
        survivors = [c.id for c in containers if c.id not in removed_ids]
        in_use = in_use_volume_names(survivors)
        vol_victims = select_pruneable_volumes(list_volumes(), in_use)
        if vol_victims:
            print(f"Selecting {len(vol_victims)} anonymous volume(s) to remove:")
            removed, _ = prune_volumes(vol_victims, dry_run)
            total_removed += removed
        else:
            print("No orphaned anonymous volumes found.")

    # --- Phase 3: dangling / stale images ---------------------------------
    print()
    print("--- Phase 3: dangling and stale images ---")
    imgs = list_images()
    print(f"Found {len(imgs)} image entries")
    img_victims = select_pruneable_images(imgs, keep, aggressive)
    freed = 0
    if img_victims:
        print(f"Selecting {len(img_victims)} image(s) to remove:")
        removed, freed = prune_images(img_victims, dry_run)
        total_removed += removed
    else:
        print("No images to remove — retention already satisfied.")

    print()
    if dry_run:
        freed_msg = f", would free {_human(freed)}" if freed else ""
        print(f"=== DRY RUN: {total_removed} resource(s) would be removed{freed_msg} ===")
    else:
        freed_msg = f", freed {_human(freed)}" if freed else ""
        print(f"=== {total_removed} resource(s) removed{freed_msg} ===")
        if total_removed > 0:
            print()
            print("NOTE: freed bytes stay trapped in the WSL VHDX until it is compacted.")
            print("      Run: python scripts/vhdx_health.py")
            print("      Then (elevated): scripts/compact-podman-vhd.ps1")
    return 0
