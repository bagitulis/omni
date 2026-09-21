"""Retention-gap regression tests for the Podman retention policy.

Guards added after the 2026-09-21 storage audit (C: 94% full, 14.3 GB free):

The original retention script only pruned IMAGES. Evidence from the live
Podman VM showed it reported "Nothing to prune — retention already
satisfied." while the machine still held:

  - 17 leaked testcontainers containers (label org.testcontainers=true,
    all Exited, 2026-09-15) that kept their images pinned as "in use";
  - 52 orphaned anonymous volumes (64-hex names) never referenced again.

Because pruning images skips anything a container references, leaked
containers silently defeat the whole retention policy. These tests pin the
two missing prune paths AND their hard safety rails:

  1. NEVER remove a container that is running.
  2. NEVER remove a container without the testcontainers label
     (omni-* compose containers must survive).
  3. NEVER remove an anonymous volume that any surviving container
     references (a run in progress would lose its data mid-test).
  4. NEVER remove a named volume (omni_omni-pgdata IS the database).

The policy lives in omni_build/podman_retention.py; the typed podman
inventory in omni_build/podman_inventory.py.
"""
from __future__ import annotations

import json
import sys
from pathlib import Path
from unittest.mock import patch

# Make omni_build importable regardless of the invocation directory.
_PYTHON_BUILD = Path(__file__).resolve().parents[1]
if str(_PYTHON_BUILD) not in sys.path:
    sys.path.insert(0, str(_PYTHON_BUILD))

from omni_build import podman_inventory as inv  # noqa: E402
from omni_build import podman_retention as pol  # noqa: E402


def _ctr(cid: str, name: str, state: str, labels: dict | None = None):
    """Build a Ctr the way list_containers() would."""
    return inv.Ctr(
        id=cid, name=name, state=state,
        image="docker.io/library/postgres:15-alpine",
        labels=labels or {},
    )


def _vol(name: str, labels: dict | None = None):
    """Build a Vol the way list_volumes() would."""
    return inv.Vol(name=name, labels=labels or {})


# ---------------------------------------------------------------------------
# Fixtures: fake podman JSON payloads (wire format)
# ---------------------------------------------------------------------------

def _container_json(
    cid: str,
    name: str,
    state: str,
    labels: dict | None = None,
    mounts: list | None = None,
) -> dict:
    return {
        "Id": cid,
        "Names": [name],
        "Image": "docker.io/library/postgres:15-alpine",
        "State": state,
        "Labels": labels or {},
        "Mounts": mounts or [],
    }


def _leaked_json(cid: str, name: str) -> dict:
    """A stopped testcontainers container — the leak this PR fixes."""
    return _container_json(
        cid, name, "exited",
        {"org.testcontainers": "true", "org.testcontainers.lang": "go"},
    )


def _volume_json(name: str, labels: dict | None = None) -> dict:
    return {"Name": name, "Driver": "local", "Labels": labels or {}}


ANON_A = "a" * 64
ANON_B = "b" * 64


# ---------------------------------------------------------------------------
# Wire parsing
# ---------------------------------------------------------------------------

class TestParseJsonRecords:
    """podman mixes two output shapes; both must parse."""

    def test_parses_json_array(self):
        assert inv.parse_json_records('[{"a": 1}, {"b": 2}]') == [{"a": 1}, {"b": 2}]

    def test_parses_json_lines(self):
        assert inv.parse_json_records('{"a": 1}\n{"b": 2}\n') == [{"a": 1}, {"b": 2}]

    def test_empty_output_yields_nothing(self):
        assert inv.parse_json_records("") == []
        assert inv.parse_json_records("   ") == []

    def test_skips_unparseable_lines(self):
        assert inv.parse_json_records('{"a": 1}\nnot json\n{"b": 2}') == [
            {"a": 1}, {"b": 2},
        ]

    def test_single_object_is_wrapped(self):
        assert inv.parse_json_records('{"a": 1}') == [{"a": 1}]


class TestListContainersParsing:
    """`podman ps --format {{json .}}` emits JSON LINES, not a JSON array."""

    def test_parses_labels_and_state(self):
        payload = json.dumps(_leaked_json("c1", "leaked"))
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            ctrs = inv.list_containers()

        assert len(ctrs) == 1
        assert ctrs[0].name == "leaked"
        assert ctrs[0].state == "exited"
        assert ctrs[0].is_testcontainers is True

    def test_parses_multiple_json_lines(self):
        """Two containers on two lines — a plain json.loads() fails here."""
        payload = "\n".join([
            json.dumps(_leaked_json("c1", "a")),
            json.dumps(_leaked_json("c2", "b")),
        ])
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            ctrs = inv.list_containers()

        assert [c.name for c in ctrs] == ["a", "b"]

    def test_normalises_names_string_to_list(self):
        payload = json.dumps({
            "Id": "c1", "Names": "single-name", "State": "Exited", "Labels": {},
        })
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            ctrs = inv.list_containers()

        assert ctrs[0].name == "single-name"
        assert ctrs[0].state == "exited"

    def test_handles_missing_labels(self):
        payload = json.dumps({"Id": "c1", "Names": ["n"], "State": "exited"})
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            ctrs = inv.list_containers()

        assert ctrs[0].labels == {}
        assert ctrs[0].is_testcontainers is False


class TestListVolumesParsing:
    """`podman volume ls --format {{json .}}` also emits JSON lines."""

    def test_parses_names(self):
        payload = "\n".join([
            json.dumps(_volume_json(ANON_A)),
            json.dumps(_volume_json("omni_omni-pgdata")),
        ])
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            vols = inv.list_volumes()

        assert {v.name for v in vols} == {ANON_A, "omni_omni-pgdata"}

    def test_handles_json_array(self):
        payload = json.dumps([_volume_json(ANON_A)])
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            vols = inv.list_volumes()

        assert [v.name for v in vols] == [ANON_A]

    def test_empty_output_yields_no_volumes(self):
        with patch.object(inv, "run_podman", return_value=(0, "", "")):
            assert inv.list_volumes() == []


# ---------------------------------------------------------------------------
# Container pruning
# ---------------------------------------------------------------------------

class TestSelectPruneableContainers:
    """Only stopped, testcontainers-labelled containers may be removed."""

    def test_selects_stopped_testcontainers(self):
        victims = pol.select_pruneable_containers([
            _ctr("c1", "vigorous_jackson", "exited", {"org.testcontainers": "true"}),
        ])
        assert [c.name for c in victims] == ["vigorous_jackson"]

    def test_never_selects_running_testcontainers(self):
        containers = [
            _ctr("c1", "busy_test", "running", {"org.testcontainers": "true"}),
        ]
        assert pol.select_pruneable_containers(containers) == []

    def test_never_selects_omni_compose_containers(self):
        """omni-postgres/redis/pgbouncer carry compose labels, not testcontainers."""
        containers = [
            _ctr("c1", "omni-postgres", "exited",
                 {"com.docker.compose.project": "omni"}),
            _ctr("c2", "omni-redis", "exited",
                 {"io.podman.compose.project": "omni"}),
        ]
        assert pol.select_pruneable_containers(containers) == []

    def test_never_selects_unlabelled_stopped_containers(self):
        """An unlabelled stopped container may be a user's manual work."""
        assert pol.select_pruneable_containers(
            [_ctr("c1", "my-debug-box", "exited", {})]
        ) == []

    def test_unknown_state_is_ignored(self):
        containers = [
            _ctr("c1", "weird", "unknown", {"org.testcontainers": "true"}),
            _ctr("c2", "normal", "exited", {"org.testcontainers": "true"}),
        ]
        victims = pol.select_pruneable_containers(containers)
        assert [c.name for c in victims] == ["normal"]

    def test_selects_all_seventeen_from_live_audit_shape(self):
        containers = [
            _ctr(f"c{i}", f"name{i}", "exited", {"org.testcontainers": "true"})
            for i in range(17)
        ]
        assert len(pol.select_pruneable_containers(containers)) == 17


class TestPruneContainers:
    """Removal runs `podman rm <id>` and reports counts."""

    def test_dry_run_reports_without_calling_podman(self):
        victims = pol.select_pruneable_containers(
            [_ctr("c1", "leaked", "exited", {"org.testcontainers": "true"})]
        )

        with patch.object(pol, "run_podman") as mock_run:
            removed, _ = pol.prune_containers(victims, dry_run=True)

        assert removed == 1
        mock_run.assert_not_called()

    def test_removes_each_victim(self):
        victims = pol.select_pruneable_containers([
            _ctr("c1", "a", "exited", {"org.testcontainers": "true"}),
            _ctr("c2", "b", "exited", {"org.testcontainers": "true"}),
        ])

        with patch.object(pol, "run_podman", return_value=(0, "", "")) as mock_run:
            removed, _ = pol.prune_containers(victims, dry_run=False)

        assert removed == 2
        assert mock_run.call_count == 2
        assert mock_run.call_args_list[0].args == ("rm", "-f", "c1")

    def test_partial_failure_is_counted_not_raised(self):
        victims = pol.select_pruneable_containers([
            _ctr("c1", "a", "exited", {"org.testcontainers": "true"}),
            _ctr("c2", "b", "exited", {"org.testcontainers": "true"}),
        ])
        responses = [(0, "", ""), (1, "", "already gone")]

        with patch.object(pol, "run_podman", side_effect=responses):
            removed, _ = pol.prune_containers(victims, dry_run=False)

        assert removed == 1


# ---------------------------------------------------------------------------
# Anonymous-volume pruning
# ---------------------------------------------------------------------------

class TestSelectPruneableVolumes:
    """Only anonymous volumes referenced by nothing may be removed."""

    def test_selects_unreferenced_anonymous_volumes(self):
        victims = pol.select_pruneable_volumes(
            [_vol(ANON_A), _vol(ANON_B)], in_use=set()
        )
        assert {v.name for v in victims} == {ANON_A, ANON_B}

    def test_never_selects_named_volumes(self):
        """omni_omni-pgdata IS the database — removing it destroys all data."""
        volumes = [_vol("omni_omni-pgdata"), _vol("omni_redis_data")]
        assert pol.select_pruneable_volumes(volumes, in_use=set()) == []

    def test_never_selects_referenced_anonymous_volumes(self):
        """A live test run would lose its data mid-import."""
        volumes = [_vol(ANON_A), _vol(ANON_B)]
        victims = pol.select_pruneable_volumes(volumes, in_use={ANON_A})
        assert [v.name for v in victims] == [ANON_B]

    def test_ignores_non_hex_volume_names(self):
        volumes = [_vol("my-project-cache"), _vol("data")]
        assert pol.select_pruneable_volumes(volumes, in_use=set()) == []

    def test_ignores_short_hex_names(self):
        """A 32-char name is not the 64-char anonymous-volume shape."""
        assert pol.select_pruneable_volumes([_vol("a" * 32)], in_use=set()) == []


class TestPruneVolumes:
    """Removal runs `podman volume rm <name>`."""

    def test_dry_run_reports_without_calling_podman(self):
        victims = pol.select_pruneable_volumes([_vol(ANON_A)], in_use=set())

        with patch.object(pol, "run_podman") as mock_run:
            removed, _ = pol.prune_volumes(victims, dry_run=True)

        assert removed == 1
        mock_run.assert_not_called()

    def test_removes_each_victim(self):
        victims = pol.select_pruneable_volumes(
            [_vol(ANON_A), _vol(ANON_B)], in_use=set()
        )

        with patch.object(pol, "run_podman", return_value=(0, "", "")) as mock_run:
            removed, _ = pol.prune_volumes(victims, dry_run=False)

        assert removed == 2
        assert mock_run.call_args_list[0].args == ("volume", "rm", "-f", ANON_A)


# ---------------------------------------------------------------------------
# Volume reference detection
# ---------------------------------------------------------------------------

class TestInUseVolumeNames:
    """Reference detection must come from `podman inspect`.

    `ps --format {{json .}}` reports Mounts as bare path strings, which cannot
    distinguish a named volume from an anonymous one, and `volume ls`
    MountCount reports 0 even for a mounted volume (verified live).
    """

    def test_parses_volume_type_mounts(self):
        payload = json.dumps([
            {"Name": "omni-postgres", "Mounts": [
                {"Type": "volume", "Name": "omni_omni-pgdata"},
                {"Type": "bind", "Source": "C:/x", "Name": ""},
            ]},
            {"Name": "vigorous_jackson", "Mounts": [
                {"Type": "volume", "Name": ANON_A},
            ]},
        ])
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            used = inv.in_use_volume_names(["c1", "c2"])

        assert used == {"omni_omni-pgdata", ANON_A}

    def test_ignores_bind_mounts_without_name(self):
        payload = json.dumps([{"Mounts": [{"Type": "bind", "Source": "/tmp/x"}]}])
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            assert inv.in_use_volume_names(["c1"]) == set()

    def test_returns_empty_on_podman_failure(self):
        with patch.object(inv, "run_podman", return_value=(1, "", "boom")):
            assert inv.in_use_volume_names(["c1"]) == set()

    def test_returns_empty_when_no_containers(self):
        with patch.object(inv, "run_podman") as mock_run:
            assert inv.in_use_volume_names([]) == set()
        mock_run.assert_not_called()

    def test_handles_missing_mounts_key(self):
        payload = json.dumps([{"Name": "c1"}])
        with patch.object(inv, "run_podman", return_value=(0, payload, "")):
            assert inv.in_use_volume_names(["c1"]) == set()

    def test_inspects_all_ids_in_one_call(self):
        with patch.object(inv, "run_podman", return_value=(0, "[]", "")) as mock_run:
            inv.in_use_volume_names(["c1", "c2", "c3"])

        assert mock_run.call_count == 1, "should batch into a single inspect call"
        args = mock_run.call_args.args
        assert args[:1] == ("inspect",)
        assert set(args[1:]) == {"c1", "c2", "c3"}


# ---------------------------------------------------------------------------
# CLI wiring: run_retention must cover all resource types
# ---------------------------------------------------------------------------

class TestRunRetention:
    """run_retention must prune containers and volumes, not only images."""

    def _run(self, keep=3, dry_run=False, aggressive=False, skip_volumes=False,
             containers=None, volumes=None, inspected=None, calls=None):
        """Run run_retention against a simulated podman."""
        containers = containers if containers is not None else []
        volumes = volumes if volumes is not None else []
        inspected = inspected if inspected is not None else []

        def fake_run(*args: str):
            if calls is not None:
                calls.append(args)
            if args[:1] == ("images",):
                return (0, "[]", "")
            if args[:1] == ("ps",):
                if "{{json .}}" in args:
                    return (0, "\n".join(json.dumps(c) for c in containers), "")
                return (0, "", "")
            if args[:1] == ("volume",):
                return (0, "\n".join(json.dumps(v) for v in volumes), "")
            if args[:1] == ("inspect",):
                return (0, json.dumps(inspected), "")
            return (0, "", "")

        with patch.object(inv, "run_podman", side_effect=fake_run), \
             patch.object(pol, "run_podman", side_effect=fake_run):
            return pol.run_retention(
                keep=keep, dry_run=dry_run,
                aggressive=aggressive, skip_volumes=skip_volumes,
            )

    def test_dry_run_succeeds(self):
        rc = self._run(
            dry_run=True,
            containers=[_leaked_json("c1", "leaked")],
            volumes=[_volume_json(ANON_A)],
        )
        assert rc == 0

    def test_dry_run_does_not_delete_anything(self):
        """The whole point: --dry-run must never reach podman rm/volume rm."""
        calls: list[tuple] = []
        self._run(
            dry_run=True, calls=calls,
            containers=[_leaked_json("c1", "leaked")],
            volumes=[_volume_json(ANON_A)],
        )
        for call in calls:
            assert "rm" not in call, f"dry run must not delete, saw: {call}"
            assert call[:2] != ("volume", "rm"), f"dry run must not delete: {call}"

    def test_real_run_prunes_stopped_testcontainers(self):
        calls: list[tuple] = []
        self._run(calls=calls, containers=[_leaked_json("c1", "leaked")])
        assert ("rm", "-f", "c1") in calls, (
            f"real run must remove the leaked container, saw: {calls}"
        )

    def test_never_removes_omni_container_in_real_run(self):
        calls: list[tuple] = []
        self._run(
            calls=calls,
            containers=[_container_json(
                "omni1", "omni-postgres", "exited",
                {"com.docker.compose.project": "omni"},
            )],
        )
        assert not any(c[:1] == ("rm",) for c in calls), (
            f"omni-postgres must never be removed, saw: {calls}"
        )

    def test_real_run_reclaims_volume_held_by_removed_container(self):
        """A volume mounted only by a leaked container must be reclaimed."""
        calls: list[tuple] = []
        self._run(
            calls=calls,
            containers=[_leaked_json("c1", "leaked")],
            volumes=[_volume_json(ANON_A)],
            # The leaked container still references ANON_A right now, but it is
            # a phase-1 victim, so the volume must still be reclaimed.
            inspected=[{"Mounts": [{"Type": "volume", "Name": ANON_A}]}],
        )
        assert ("volume", "rm", "-f", ANON_A) in calls, (
            f"volume freed by container removal must be reclaimed, saw: {calls}"
        )

    def test_never_removes_pgdata_volume_in_real_run(self):
        """omni_omni-pgdata must survive even when a removed container mounts it."""
        calls: list[tuple] = []
        self._run(
            calls=calls,
            containers=[_leaked_json("c1", "leaked")],
            volumes=[_volume_json("omni_omni-pgdata")],
            inspected=[{"Mounts": [{"Type": "volume", "Name": "omni_omni-pgdata"}]}],
        )
        assert not any(c[:2] == ("volume", "rm") for c in calls), (
            f"pgdata must never be removed, saw: {calls}"
        )

    def test_skip_volumes_leaves_volumes_alone(self):
        calls: list[tuple] = []
        self._run(
            skip_volumes=True, calls=calls, containers=[],
            volumes=[_volume_json(ANON_A)],
        )
        assert not any(c[:2] == ("volume", "rm") for c in calls)

    def test_reports_unavailable_podman(self):
        with patch.object(pol, "ensure_podman_available", return_value=False):
            rc = pol.run_retention(
                keep=3, dry_run=True, aggressive=False, skip_volumes=False,
            )
        assert rc == 1

    def test_stopped_machine_is_not_an_error(self):
        """A stopped machine cannot accumulate space, so skip cleanly.

        Verified live 2026-09-21: with the machine stopped, `podman ps` exits 1
        with "Cannot connect to Podman ... actively refused it". Treating that
        as a failure would report a scary [FAIL] after every build on a machine
        whose VM is simply off.
        """
        calls: list[tuple] = []

        def fake_run(*args: str):
            calls.append(args)
            return (1, "", "Cannot connect to Podman.")

        with patch.object(inv, "run_podman", side_effect=fake_run), \
             patch.object(pol, "run_podman", side_effect=fake_run):
            rc = pol.run_retention(
                keep=3, dry_run=True, aggressive=False, skip_volumes=False,
            )

        assert rc == 0, "an unreachable runtime must not fail the build"
        assert not any(c[:1] == ("rm",) for c in calls)
        assert not any(c[:2] == ("volume", "rm") for c in calls)

    def test_reachable_runtime_proceeds_to_prune(self):
        """Guard against the skip check swallowing real work."""
        calls: list[tuple] = []

        def fake_run(*args: str):
            calls.append(args)
            if args[:1] == ("info",):
                return (0, "host info", "")
            if args[:1] == ("images",):
                return (0, "[]", "")
            if args[:1] == ("ps",):
                if "{{json .}}" in args:
                    return (0, json.dumps(_leaked_json("c1", "leaked")), "")
                return (0, "", "")
            if args[:1] == ("volume",):
                return (0, "", "")
            if args[:1] == ("inspect",):
                return (0, "[]", "")
            return (0, "", "")

        with patch.object(inv, "run_podman", side_effect=fake_run), \
             patch.object(pol, "run_podman", side_effect=fake_run):
            pol.run_retention(
                keep=3, dry_run=False, aggressive=False, skip_volumes=False,
            )

        assert ("rm", "-f", "c1") in calls, (
            f"a reachable runtime must still prune, saw: {calls}"
        )


class TestImageSelectionSafety:
    """The pre-existing image rules must keep holding after the refactor."""

    def test_aggressive_still_protects_named_volumes(self):
        victims = pol.select_pruneable_volumes(
            [_vol("omni_omni-pgdata"), _vol(ANON_A)], in_use=set()
        )
        assert {v.name for v in victims} == {ANON_A}

    def test_never_prunes_images_referenced_by_containers(self):
        img = inv.Img("img1", "localhost/omni_backend", "latest", 100, 1)
        victims = pol.select_pruneable_images([img], keep=0, aggressive=True,
                                              in_use={"img1"})
        assert victims == []

    def test_keeps_latest_tag(self):
        img = inv.Img("img1", "localhost/omni_backend", "latest", 100, 1)
        victims = pol.select_pruneable_images([img], keep=0, aggressive=False,
                                              in_use=set())
        assert victims == []

    def test_never_prunes_third_party_base_images(self):
        img = inv.Img("img1", "docker.io/library/postgres", "15-alpine", 100, 1)
        victims = pol.select_pruneable_images([img], keep=0, aggressive=True,
                                              in_use=set())
        assert victims == []
