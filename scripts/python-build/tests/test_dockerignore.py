"""
Regression test for the "tsc: Permission denied" build failure that
happens when the host's `node_modules/` gets copied into the container
via `COPY . .` and clobbers the container-side `npm ci` output
(no executable bit on Windows -> `sh: tsc: Permission denied`).

Rule: both frontend/ and backend/ MUST have a .dockerignore that
excludes at least node_modules (frontend) and the git dir. The backend
context also excludes local .venv/data/log noise to keep builds sane.
"""
from __future__ import annotations

from pathlib import Path

import pytest

ROOT = Path(__file__).resolve().parent.parent.parent.parent


@pytest.mark.parametrize(
    "context, must_ignore",
    [
        ("frontend", {"node_modules", "dist", ".git"}),
        ("backend", {".git"}),
    ],
)
def test_dockerignore_excludes_critical_paths(context: str, must_ignore: set[str]) -> None:
    dockerignore = ROOT / context / ".dockerignore"
    assert dockerignore.exists(), (
        f"{context}/.dockerignore is missing. Without it, `COPY . .` in "
        f"{context}/Dockerfile drags host cruft (node_modules, .git, "
        f"logs, envs) into the build context. On Windows + Podman this "
        f"clobbers container-side installs and strips the executable "
        f"bit (`sh: tsc: Permission denied`)."
    )
    lines = {
        line.strip().lstrip("/").rstrip("/")
        for line in dockerignore.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.lstrip().startswith("#")
    }
    missing = must_ignore - lines
    assert not missing, (
        f"{context}/.dockerignore must ignore {sorted(missing)}. "
        f"Currently ignores: {sorted(lines)}"
    )


def test_frontend_dockerignore_blocks_node_modules_first() -> None:
    """node_modules must appear before any negation ('!node_modules/...').
    We do not currently use negations, but if someone adds one they must
    not accidentally re-include node_modules."""
    lines = (ROOT / "frontend" / ".dockerignore").read_text(encoding="utf-8").splitlines()
    node_modules_seen = False
    for raw in lines:
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if line == "node_modules" or line.startswith("node_modules/"):
            node_modules_seen = True
        elif node_modules_seen and line.startswith("!"):
            assert "node_modules" not in line, (
                f"Negation would re-include node_modules: {line!r}"
            )
    assert node_modules_seen, "node_modules is not ignored in frontend/.dockerignore"
