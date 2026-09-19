"""TDD tests for DockerDeployer compose up command construction.

Regression guard for the 2026-09-19 deploy hang: podman-compose does NOT
recreate an existing container whose service definition is unchanged from
its point of view, even when the container was created by an older compose
file revision that had no healthcheck. Services with
`depends_on: service_healthy` then wait forever and every deploy attempt
times out.

Adding `--force-recreate` guarantees every `up` rebuilds all containers
from the current images + compose config, so healthchecks always exist.
"""
from unittest.mock import MagicMock, patch

import pytest

from omni_build.docker_deployer import DockerDeployer
from omni_build.models import SpecLevel


@pytest.fixture
def deployer() -> DockerDeployer:
    config = MagicMock()
    config.get_compose_files.return_value = ["a.yml", "b.yml"]
    config.max_deploy_retries = 3
    config.docker_deploy_timeout = 900
    config.project_root = "."
    return DockerDeployer(config, error_handler=MagicMock())


class TestDeployComposeCommand:
    def test_up_command_includes_force_recreate(self, deployer: DockerDeployer):
        """Regression: stale containers without healthcheck must be recreated."""
        captured: dict = {}

        def fake_run(cmd, **kwargs):
            captured["cmd"] = cmd
            return MagicMock(returncode=0, stdout="", stderr="")

        with patch("omni_build.docker_deployer.subprocess.run", side_effect=fake_run), \
             patch("omni_build.docker_deployer.get_compose_command", return_value=["podman-compose"]), \
             patch("omni_build.docker_deployer.time.sleep"), \
             patch.object(DockerDeployer, "_check_core_containers_status",
                          return_value={"postgres_running": True, "backend_exists": True}), \
             patch.object(DockerDeployer, "_wait_for_all_containers_healthy", return_value=True), \
             patch.object(DockerDeployer, "_restart_nginx_after_deploy"):
            assert deployer.deploy_containers(SpecLevel.STANDARD) is True

        cmd = captured["cmd"]
        assert "up" in cmd
        assert "-d" in cmd
        assert "--remove-orphans" in cmd
        assert "--force-recreate" in cmd
