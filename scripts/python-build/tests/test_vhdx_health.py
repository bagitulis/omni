"""Tests for WSL VHDX bloat detection and compaction guidance.

Guards added after the 2026-09-21 storage audit. Two independent facts were
established live:

  1. Deleting images/containers/volumes inside the Podman VM does NOT shrink
     ext4.vhdx on the Windows host. The file is not sparse ("fsutil sparse
     queryflag" -> "This file is NOT set as sparse"), so Windows keeps every
     byte allocated. Measured: VHDX 16.73 GB while the VM only used 7.10 GB,
     leaving 9.63 GB stranded.

  2. Compaction needs Administrator rights and was never wired into build.py —
     scripts/compact-podman-vhd.ps1 existed but was referenced only in a
     comment. So a build could prune perfectly and still never return space.

These tests pin the pure decision logic so the reminder cannot silently rot.
They deliberately do NOT run compaction (that requires elevation).
"""
from __future__ import annotations

import importlib.util
import sys
from pathlib import Path

import pytest


_SCRIPT = Path(__file__).resolve().parents[2] / "vhdx_health.py"


def _load():
    spec = importlib.util.spec_from_file_location("vhdx_health_under_test", _SCRIPT)
    assert spec is not None and spec.loader is not None, f"cannot load {_SCRIPT}"
    module = importlib.util.module_from_spec(spec)
    sys.modules[spec.name] = module
    spec.loader.exec_module(module)
    return module


@pytest.fixture(scope="module")
def vhdx():
    return _load()


GB = 1024 ** 3


class TestDeadSpaceCalculation:
    """Dead space = allocated file size minus bytes the VM actually uses."""

    def test_detects_stranded_space(self, vhdx):
        # The exact live shape: 16.73 GB file, 7.10 GB used.
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        assert health.dead_bytes == pytest.approx(9.63 * GB, rel=0.01)

    def test_reports_healthy_when_fully_used(self, vhdx):
        health = vhdx.assess(file_bytes=10 * GB, used_bytes=10 * GB)
        assert health.dead_bytes == 0
        assert health.needs_compaction is False

    def test_never_reports_negative_dead_space(self, vhdx):
        """File size can briefly lag usage (sparse/optimised VHDX)."""
        health = vhdx.assess(file_bytes=5 * GB, used_bytes=9 * GB)
        assert health.dead_bytes == 0


class TestCompactionThreshold:
    """Only nag when the waste is genuinely worth an elevated action."""

    def test_below_threshold_does_not_nag(self, vhdx):
        health = vhdx.assess(
            file_bytes=10 * GB, used_bytes=int(9.5 * GB), threshold_bytes=1 * GB
        )
        assert health.dead_bytes == pytest.approx(0.5 * GB, rel=0.01)
        assert health.needs_compaction is False

    def test_above_threshold_nags(self, vhdx):
        health = vhdx.assess(
            file_bytes=16.73 * GB, used_bytes=int(7.10 * GB),
            threshold_bytes=1 * GB,
        )
        assert health.needs_compaction is True

    def test_live_audit_shape_needs_compaction_with_default_threshold(self, vhdx):
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        assert health.needs_compaction is True

    def test_threshold_boundary_is_exclusive(self, vhdx):
        """Exactly at the threshold is not yet worth compacting."""
        health = vhdx.assess(
            file_bytes=10 * GB, used_bytes=9 * GB, threshold_bytes=1 * GB
        )
        assert health.needs_compaction is False


class TestMessageRendering:
    """The reminder must state the numbers and the exact elevated command."""

    def test_message_includes_sizes_and_command(self, vhdx):
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")
        assert "9.6" in msg or "9.63" in msg, f"dead space missing from: {msg}"
        assert "compact-podman-vhd.ps1" in msg
        assert "podman-machine-default" in msg

    def test_message_warns_about_administrator(self, vhdx):
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")
        assert "dmin" in msg, "must warn that elevation is required"

    def test_no_message_when_healthy(self, vhdx):
        health = vhdx.assess(file_bytes=10 * GB, used_bytes=10 * GB)
        assert vhdx.render_reminder(health, distro="d") == ""

    def test_message_is_not_accidentally_multiplied(self, vhdx):
        """Guards a real Python trap: "=" * 62 adjacent to a string literal.

        `"=" * 62 + "\\n"` followed by more adjacent literals parses as
        ("=" * 62) + ("\\n..." * 62), duplicating the whole body 62 times.
        The rendered reminder must appear exactly once.
        """
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")

        assert msg.count("[STORAGE]") == 1, (
            f"reminder body duplicated {msg.count('[STORAGE]')}x"
        )
        assert msg.count("compact-podman-vhd.ps1") == 1, "command line duplicated"
        assert len(msg.splitlines()) < 30, (
            f"reminder should be compact, got {len(msg.splitlines())} lines"
        )

    def test_rule_line_length_is_correct(self, vhdx):
        """The separator must not balloon into one '=' per line."""
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")

        equals_only = [ln for ln in msg.splitlines() if ln and set(ln) == {"="}]
        assert equals_only, "expected at least one separator rule"
        for line in equals_only:
            assert len(line) >= 40, f"separator collapsed: {len(line)} chars"


class TestVhdxPathDiscovery:
    """The VHDX lives under the Podman machine dir, not the project."""

    def test_builds_expected_path(self, vhdx):
        from pathlib import Path as P

        path = vhdx.vhdx_path_for("podman-machine-default", home=P("/home/u"))
        assert str(path).replace("\\", "/").endswith(
            ".local/share/containers/podman/machine/wsl/wsldist/"
            "podman-machine-default/ext4.vhdx"
        )
