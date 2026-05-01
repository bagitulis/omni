"""
Database configuration constants for Omni Build System.

SRP: Centralized database configuration to avoid hardcoded values.
"""
import os
from typing import List


class DatabaseConfig:
    """Centralized database configuration."""

    CONTAINER_NAME = os.getenv("POSTGRES_CONTAINER", "omni-postgres")
    USER = os.getenv("POSTGRES_USER", "omni")
    DATABASE = os.getenv("POSTGRES_DB", "omni_main")

    # Thresholds
    LARGE_TABLE_THRESHOLD = int(os.getenv("LARGE_TABLE_THRESHOLD", "50000"))
    CHUNK_SIZE = int(os.getenv("CHUNK_SIZE", "25000"))

    # Timeouts (seconds)
    BACKUP_TIMEOUT = int(os.getenv("BACKUP_TIMEOUT", "300"))
    RESTORE_TIMEOUT = int(os.getenv("RESTORE_TIMEOUT", "600"))
    HEALTH_CHECK_TIMEOUT = int(os.getenv("HEALTH_CHECK_TIMEOUT", "10"))

    @classmethod
    def docker_exec_prefix(cls) -> List[str]:
        """Return the docker exec prefix for PostgreSQL commands."""
        return ["docker", "exec", cls.CONTAINER_NAME]

    @classmethod
    def psql_cmd(cls, database: str = "") -> List[str]:
        """Return psql command with user and database."""
        db = database or cls.DATABASE
        return cls.docker_exec_prefix() + ["psql", "-U", cls.USER, "-d", db]
