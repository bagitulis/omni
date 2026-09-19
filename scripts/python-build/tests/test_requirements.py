"""Guardrail tests: declared dependencies must match runtime capabilities.

The podman migration (commit 32433b32) added ContainerRuntime support for
podman-compose, but requirements.txt must also declare the dependency —
otherwise a fresh podman-only host passes unit tests (mocked) yet fails at
real `podman compose` invocation.
"""
import re
from pathlib import Path

REQUIREMENTS = Path(__file__).resolve().parents[1] / "requirements.txt"


def _requirements_text() -> str:
    return REQUIREMENTS.read_text(encoding="utf-8")


def _requirement_names() -> list[str]:
    names = []
    for raw in _requirements_text().splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        # Strip env markers and version specifiers: "podman-compose>=1.0; python_version>'3.8'"
        name = re.split(r"[<>=!;\[\s]", line, maxsplit=1)[0]
        names.append(name.lower().replace("_", "-"))
    return names


class TestPodmanComposeDependency:
    """podman runtime support must ship its compose provider dependency."""

    def test_requirements_declares_podman_compose(self):
        assert "podman-compose" in _requirement_names(), (
            "requirements.txt must declare podman-compose: ContainerRuntime "
            "detects it as the compose provider on podman-only hosts, but a "
            "fresh setup without it fails `podman compose version`."
        )
