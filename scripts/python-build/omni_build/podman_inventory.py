"""Podman inventory: typed listing of images, containers and volumes.

Split out of scripts/podman-retention.py so the CLI stays a thin wrapper and
each module keeps a single responsibility (project ~300-line rule).

Wire-format notes (verified live 2026-09-21 against Podman 5.5.2 on Windows):

  * `podman images --format json` emits a JSON ARRAY.
  * `podman ps --all --format {{json .}}` and
    `podman volume ls --format {{json .}}` emit JSON LINES (one object per
    line). Calling json.loads() on the latter raises
    "Extra data: line 2 column 1", which is why _parse_json_records exists.
  * `ps` reports Mounts as bare PATH STRINGS, so a named volume is
    indistinguishable from an anonymous one there. `volume ls` MountCount
    reports 0 even for a volume a running container mounts. Only
    `podman inspect` exposes mounts as objects carrying `Name`, so volume
    reference detection must use it.
"""
from __future__ import annotations

import json
import re
import shutil
import subprocess
import sys
from dataclasses import dataclass

PODMAN = shutil.which("podman") or "podman"

# Label every testcontainers container carries (testcontainers-go/java/...).
TESTCONTAINERS_LABEL = "org.testcontainers"

# Anonymous volumes are named with a 64-char lowercase hex string.
ANON_VOLUME_RE = re.compile(r"^[0-9a-f]{64}$")

# Container states safe to remove (stopped/exited). Anything else is kept.
REMOVABLE_STATES = {"exited", "stopped", "created", "configured", "dead"}


@dataclass(frozen=True)
class Img:
    id: str
    repo: str
    tag: str
    size_bytes: int
    created_ns: int  # unix seconds *1e9 for stable sort


@dataclass(frozen=True)
class Ctr:
    id: str
    name: str
    state: str
    image: str
    labels: dict

    @property
    def is_testcontainers(self) -> bool:
        return self.labels.get(TESTCONTAINERS_LABEL) == "true"


@dataclass(frozen=True)
class Vol:
    name: str
    labels: dict


def run_podman(*args: str) -> tuple[int, str, str]:
    """Run a podman subcommand. Returns (returncode, stdout, stderr)."""
    proc = subprocess.run([PODMAN, *args], capture_output=True, text=True)
    return proc.returncode, proc.stdout, proc.stderr


def parse_json_records(out: str) -> list[dict]:
    """Parse podman output that may be a JSON array OR JSON lines.

    See the module docstring: `images` emits an array while `ps` and
    `volume ls` emit one object per line.
    """
    text = (out or "").strip()
    if not text:
        return []
    try:
        parsed = json.loads(text)
    except json.JSONDecodeError:
        records: list[dict] = []
        for line in text.splitlines():
            line = line.strip()
            if not line:
                continue
            try:
                item = json.loads(line)
            except json.JSONDecodeError:
                continue
            if isinstance(item, dict):
                records.append(item)
        return records
    if isinstance(parsed, dict):
        return [parsed]
    if isinstance(parsed, list):
        return [r for r in parsed if isinstance(r, dict)]
    return []


def list_images() -> list[Img]:
    """Return every image, tagged or not, with repo/tag/size/created."""
    rc, out, err = run_podman("images", "--all", "--format", "json")
    if rc != 0:
        print(f"[FAIL] podman images: {err.strip()}", file=sys.stderr)
        sys.exit(1)
    imgs: list[Img] = []
    for row in parse_json_records(out):
        names: list[str] = row.get("Names") or []
        if isinstance(names, str):
            names = [names]
        size = int(row.get("Size", 0) or 0)
        created = int(row.get("Created", 0) or 0)
        image_id = row.get("Id", "")
        if names:
            for full in names:
                repo, tag = full.rsplit(":", 1) if ":" in full else (full, "latest")
                imgs.append(Img(image_id, repo, tag, size, created))
        else:
            # Rows without Names are dangling.
            imgs.append(Img(image_id, "<none>", "<none>", size, created))
    return imgs


def list_containers() -> list[Ctr]:
    """Return every container with state and labels."""
    rc, out, err = run_podman("ps", "--all", "--format", "{{json .}}")
    if rc != 0:
        print(f"[FAIL] podman ps: {err.strip()}", file=sys.stderr)
        sys.exit(1)
    ctrs: list[Ctr] = []
    for row in parse_json_records(out):
        names = row.get("Names") or []
        if isinstance(names, str):
            names = [names]
        ctrs.append(Ctr(
            id=row.get("Id", ""),
            name=names[0] if names else "",
            state=(row.get("State") or "").lower(),
            image=row.get("Image") or "",
            labels=row.get("Labels") or {},
        ))
    return ctrs


def list_volumes() -> list[Vol]:
    """Return every volume with its labels."""
    rc, out, err = run_podman("volume", "ls", "--format", "{{json .}}")
    if rc != 0:
        print(f"[FAIL] podman volume ls: {err.strip()}", file=sys.stderr)
        sys.exit(1)
    return [
        Vol(name=row.get("Name", ""), labels=row.get("Labels") or {})
        for row in parse_json_records(out)
        if row.get("Name")
    ]


def in_use_image_ids() -> set[str]:
    """IDs of images referenced by any container (running or stopped)."""
    rc, out, _ = run_podman("ps", "--all", "--format", "{{.ImageID}}")
    if rc != 0:
        return set()
    return {line.strip() for line in out.splitlines() if line.strip()}


def in_use_volume_names(container_ids: list[str]) -> set[str]:
    """Names of volumes mounted by the given containers.

    Uses `podman inspect` for the reasons in the module docstring. A stopped
    container still holds its volume; removing a referenced anonymous volume
    would make a resumable container start from an empty volume.
    """
    if not container_ids:
        return set()
    rc, out, _ = run_podman("inspect", *container_ids)
    if rc != 0:
        return set()
    used: set[str] = set()
    for row in parse_json_records(out):
        for mount in row.get("Mounts") or []:
            if not isinstance(mount, dict):
                continue
            if (mount.get("Type") or "volume") != "volume":
                continue
            name = mount.get("Name")
            if name:
                used.add(name)
    return used


def human_bytes(n: int) -> str:
    """Render a byte count using binary units."""
    value = float(n)
    for unit in ("B", "KB", "MB", "GB", "TB"):
        if value < 1024:
            return f"{value:.1f}{unit}"
        value /= 1024
    return f"{value:.1f}PB"


def system_df() -> str:
    """Return `podman system df` output, or "" when unavailable."""
    rc, out, _ = run_podman("system", "df")
    return out.strip() if rc == 0 else ""
