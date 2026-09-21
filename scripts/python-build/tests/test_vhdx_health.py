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
from unittest.mock import patch

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


class TestUnsafeFlagWarning:
    """--allow-unsafe must never be recommended without its data risk.

    WSL itself prints this when asked to set a distro sparse:

        "Sparse VHD support is currently disabled due to potential data
         corruption. To force a distribution to use a sparse VHD, please run:
         wsl.exe --manage <DistributionName> --set-sparse true --allow-unsafe"

    Microsoft disables sparse VHD by default for a reason, and the flag name
    ("--allow-unsafe") says so. Presenting it as a plain equal alternative to
    the safe offline compaction would mislead the operator, so the reminder
    must mark the risk and present the ps1 script as the recommended path.
    """

    def test_reminder_marks_data_corruption_risk(self, vhdx):
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")

        lowered = msg.lower()
        assert "corruption" in lowered, (
            f"reminder must warn about the data-corruption risk, got:\n{msg}"
        )

    def test_reminder_presents_safe_script_as_recommended(self, vhdx):
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")

        lowered = msg.lower()
        assert "recommend" in lowered, (
            f"reminder must name the recommended (safe) path, got:\n{msg}"
        )
        # The recommended path must be the offline compaction script.
        recommended_idx = lowered.index("recommend")
        script_idx = msg.index("compact-podman-vhd.ps1")
        assert script_idx > recommended_idx, (
            "compact-podman-vhd.ps1 must appear under the recommendation"
        )

    def test_unsafe_flag_is_grouped_under_a_warning(self, vhdx):
        """The --allow-unsafe line must be preceded by its own warning label."""
        health = vhdx.assess(file_bytes=int(16.73 * GB), used_bytes=int(7.10 * GB))
        msg = vhdx.render_reminder(health, distro="podman-machine-default")

        lines = msg.splitlines()
        flag_lines = [i for i, ln in enumerate(lines) if "--allow-unsafe" in ln]
        assert flag_lines, "reminder should still document the alternative"

        for idx in flag_lines:
            # Look in the few lines above the flag for an explicit warning.
            window = "\n".join(lines[max(0, idx - 6):idx + 2]).lower()
            assert "warn" in window or "unsafe" in window or "corrupt" in window, (
                f"--allow-unsafe must be labelled as risky, window was:\n{window}"
            )

    def test_non_elevated_guidance_also_warns(self, vhdx, capsys):
        """run_compaction's printed guidance must carry the same warning."""
        with patch.object(vhdx, "is_elevated", return_value=False), \
             patch.object(vhdx.sys, "platform", "win32"):
            vhdx.run_compaction(Path("."), distro="podman-machine-default")

        out = capsys.readouterr().out
        assert "compact-podman-vhd.ps1" in out, out
        if "--allow-unsafe" in out:
            assert "corruption" in out.lower() or "unsafe" in out.lower(), (
                f"non-elevated guidance mentions --allow-unsafe without a "
                f"warning:\n{out}"
            )


class TestVhdxPathDiscovery:
    """The VHDX lives under the Podman machine dir, not the project."""

    def test_builds_expected_path(self, vhdx):
        from pathlib import Path as P

        path = vhdx.vhdx_path_for("podman-machine-default", home=P("/home/u"))
        assert str(path).replace("\\", "/").endswith(
            ".local/share/containers/podman/machine/wsl/wsldist/"
            "podman-machine-default/ext4.vhdx"
        )


class TestCompactionScriptRestartPolicy:
    """The ps1 must not resurrect a machine the operator had stopped.

    Observed live 2026-09-21: the operator ran the script on a stopped machine,
    and it finished by unconditionally starting the VM and all 7 omni-*
    containers. The machine was left running when the operator wanted it off,
    requiring a manual stop afterwards.

    These are text assertions on a PowerShell script: it cannot be executed
    here (it needs Administrator), so we pin the contract instead of behaviour.
    """

    @staticmethod
    def _script() -> str:
        path = Path(__file__).resolve().parents[2] / "compact-podman-vhd.ps1"
        assert path.exists(), f"missing {path}"
        return path.read_text(encoding="utf-8")

    def test_supports_no_restart_switch(self):
        assert "NoRestart" in self._script(), (
            "script must expose a -NoRestart switch so operators can compact "
            "without the machine coming back up"
        )

    def test_records_initial_machine_state(self):
        """Restart must be conditional on the machine having been up before."""
        script = self._script()
        assert "wasRunning" in script or "WasRunning" in script, (
            "script must capture whether the machine was running beforehand"
        )

    def test_restart_is_guarded_not_unconditional(self):
        """A bare `podman machine start` outside any condition is the bug."""
        script = self._script()
        lines = script.splitlines()
        for idx, line in enumerate(lines):
            stripped = line.strip()
            if not stripped.startswith("podman machine start"):
                continue
            # Accept it only when nested inside an if/else block.
            indent = len(line) - len(line.lstrip())
            assert indent > 0, (
                f"L{idx + 1}: `podman machine start` must not sit at top level "
                f"(it restarts a machine the operator stopped): {line}"
            )

    def test_restart_is_gated_on_was_running(self):
        """The gate must reference the recorded state, not merely -NoRestart.

        Indentation alone is too weak: a guarded-if could test the wrong
        thing. The decision variable must combine the recorded state with the
        switch.
        """
        script = self._script()
        assert "$shouldRestart = $wasRunning -and (-not $NoRestart)" in script, (
            "restart must be gated on BOTH the recorded state and -NoRestart"
        )

    def test_stopped_machine_path_exits_before_starting_containers(self):
        """The 'stay stopped' branch must exit, never fall through."""
        script = self._script()
        # Between the not-$shouldRestart branch and the restart write-out
        # there must be an exit, so control cannot reach the start loop.
        gate_idx = script.index("if (-not $shouldRestart)")
        restart_idx = script.index("Restoring state: machine was running")
        between = script[gate_idx:restart_idx]
        assert "exit 0" in between, (
            "the stay-stopped branch must exit before the restart section"
        )

    def test_only_previously_running_containers_are_restarted(self):
        """Restarting all 7 omni-* containers was the original bug."""
        script = self._script()
        restart_idx = script.index("Restoring state: machine was running")
        tail = script[restart_idx:]
        assert "foreach ($c in $runningContainers)" in tail, (
            "restart must iterate the containers recorded as running"
        )
        assert "foreach ($c in $containers)" not in tail, (
            "restart must not iterate a hardcoded full container list"
        )

    def test_documents_the_no_restart_usage(self):
        script = self._script().lower()
        assert "norestart" in script and "usage" in script, (
            "the header usage block should document -NoRestart"
        )
