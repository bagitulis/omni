"""
Error Pattern definitions for Omni Build System.

SRP: This module defines infrastructure error patterns (Docker, Postgres, Redis).
Service patterns are in error_patterns_services.py.
"""
from omni_build.models import ErrorPattern, ErrorSeverity

# Import service patterns from separate modules
from omni_build.error_patterns_services import (
    get_backend_patterns,
    get_cloudflared_patterns,
    get_nginx_patterns,
    get_pgbouncer_patterns,
    get_pgweb_patterns,
)
from omni_build.error_patterns_system import (
    get_build_patterns,
    get_code_error_patterns,
    get_network_dns_patterns,
    get_resource_patterns,
)


def get_docker_infrastructure_patterns() -> list[ErrorPattern]:
    """Docker infrastructure error patterns."""
    return [
        ErrorPattern(
            name="DockerEngineError",
            pattern=r"500 Internal Server Error|dockerDesktopLinuxEngine.*_ping|request returned 500|error during connect",
            description="Docker Desktop engine error (500)",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_docker_engine",
        ),
        ErrorPattern(
            name="DockerPipeError",
            pattern=r"pipe.*docker|npipe.*error|named pipe|\\\\\\\\.\\pipe",
            description="Docker pipe connection error",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_docker_pipe",
        ),
        ErrorPattern(
            name="WslMountCacheError",
            pattern=r"/run/desktop/mnt/host|mount source path|mkdir.*mnt|creating mount.*path|file exists.*mnt|mnt/host.*file exists",
            description="WSL2 mount cache corruption",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_wsl_mount_cache",
        ),
        ErrorPattern(
            name="WslKernelError",
            pattern=r"wsl.*kernel|wsl2.*error|vmlinux|WSL.*failed|docker-desktop.*distro",
            description="WSL2 kernel error",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_wsl_kernel",
        ),
        ErrorPattern(
            name="HyperVError",
            pattern=r"hyperv|hyper-v|virtualization|vmcompute|hv_sock",
            description="Hyper-V virtualization issue",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_hyperv",
        ),
        ErrorPattern(
            name="ManifestError",
            pattern=r"no matching manifest|manifest.*not found|platform.*not supported",
            description="Docker manifest/platform error",
            severity=ErrorSeverity.MEDIUM,
            fix_function="switch_docker_to_linux",
        ),
        ErrorPattern(
            name="DockerCredentialError",
            pattern=r"error getting credentials|no usernames for|credsStore.*desktop|credential.*helper",
            description="Docker credential helper misconfigured (common after OS migration)",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_docker_credentials",
        ),
        ErrorPattern(
            name="BuildKitError",
            pattern=r"buildkit|builder.*error|failed to solve",
            description="BuildKit cache or build issue",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_buildkit",
        ),
        ErrorPattern(
            name="ContainerNameConflict",
            pattern=r"container name.*already in use|Conflict.*container name",
            description="Container name conflict",
            severity=ErrorSeverity.LOW,
            fix_function="repair_container_name_conflict",
        ),
        ErrorPattern(
            name="DependencyFailedToStart",
            pattern=r"dependency.*failed to start|depends_on.*failed|service.*unhealthy",
            description="Dependency service failed to start",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_dependency_failure",
        ),
        ErrorPattern(
            name="ServiceUnhealthy",
            pattern=r"service.*unhealthy|container reports unhealthy|Health status: unhealthy",
            description="Service failed health check",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_unhealthy_service",
        ),
    ]


def get_postgres_patterns() -> list[ErrorPattern]:
    """PostgreSQL error patterns."""
    return [
        ErrorPattern(
            name="PostgresPgHbaError",
            pattern=r"no pg_hba\.conf entry|FATAL.*no pg_hba\.conf|SQLSTATE 28000",
            description="PostgreSQL pg_hba.conf authentication denied",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_postgres_pg_hba",
        ),
        ErrorPattern(
            name="PostgresDatabaseNotExist",
            pattern=r"database.*does not exist|FATAL.*database.*does not exist|SQLSTATE 3D000",
            description="PostgreSQL database does not exist",
            severity=ErrorSeverity.CRITICAL,
            fix_function="create_postgres_database_if_not_exists",
        ),
        ErrorPattern(
            name="PostgresOwnershipError",
            pattern=r"data directory.*wrong ownership|must be started by the user that owns the data directory",
            description="PostgreSQL data directory ownership error (NTFS/exFAT filesystem)",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_postgres_ownership",
        ),
        ErrorPattern(
            name="PostgresDataCorruption",
            pattern=r"could not open directory.*pg_|pg_notify.*No such file|pg_wal.*No such file|pg_xact.*No such file|database.*shut down|FATAL.*postgres|could not open file.*pg_filenode",
            description="PostgreSQL data directory corrupted",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_postgres_data",
        ),
        ErrorPattern(
            name="DockerDNSPostgresError",
            pattern=r"lookup postgres on.*no such host|failed to connect.*postgres.*hostname resolving|postgres.*127\.0\.0\.11.*no such host",
            description="Docker internal DNS cannot resolve 'postgres' hostname",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_docker_dns_postgres",
        ),
        # === SILENT ERRORS - Service tetap running tapi ada masalah ===
        ErrorPattern(
            name="PostgresIOError",
            pattern=r"could not (read|write|fsync|extend|open) (?:block|file)|Input/output error|I/O error|pread failed|pwrite failed|read error|write failed",
            description="PostgreSQL I/O error - disk/storage issue saat service running",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_postgres_io",
        ),
        ErrorPattern(
            name="PostgresCheckpointError",
            pattern=r"checkpoint.*failed|checkpoint.*error|checkpointer.*error|could not flush dirty data|PANIC.*checkpoint",
            description="PostgreSQL checkpoint failure - data mungkin tidak tersimpan",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_postgres_checkpoint",
        ),
        ErrorPattern(
            name="PostgresWALError",
            pattern=r"wal.*error|could not.*wal|wal.*corruption|wal.*failed|requested WAL segment.*has already been removed|could not open file.*pg_wal",
            description="PostgreSQL WAL (Write-Ahead Log) error",
            severity=ErrorSeverity.CRITICAL,
            fix_function="repair_postgres_wal",
        ),
        ErrorPattern(
            name="PostgresDeadlock",
            pattern=r"deadlock detected|process.*still waiting|lock.*not available|could not obtain lock",
            description="PostgreSQL deadlock atau lock timeout",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_postgres_deadlock",
        ),
        ErrorPattern(
            name="PostgresConnectionExhausted",
            pattern=r"too many connections|FATAL.*connection limit|connection slot|remaining connection slots|max_connections",
            description="PostgreSQL connection pool habis",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_postgres_connections",
        ),
        ErrorPattern(
            name="PostgresReplicationLag",
            pattern=r"replication.*lag|streaming replication.*behind|recovery.*still|standby.*behind",
            description="PostgreSQL replication lag (jika ada replica)",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_postgres_replication",
        ),
    ]


def get_redis_patterns() -> list[ErrorPattern]:
    """Redis error patterns."""
    return [
        ErrorPattern(
            name="RedisConnectionRefused",
            pattern=r"redis.*connection refused|ECONNREFUSED.*6379|connect ECONNREFUSED.*redis|dial tcp.*redis.*connection refused",
            description="Redis connection refused - service not running",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_redis",
        ),
        ErrorPattern(
            name="RedisAuthError",
            pattern=r"redis.*NOAUTH|redis.*authentication|ERR AUTH|WRONGPASS",
            description="Redis authentication failed",
            severity=ErrorSeverity.MEDIUM,
            fix_function="restart_redis",
        ),
        ErrorPattern(
            name="RedisMaxMemory",
            pattern=r"redis.*OOM|redis.*maxmemory|NOSCRIPT.*memory|command not allowed when used memory",
            description="Redis out of memory",
            severity=ErrorSeverity.HIGH,
            fix_function="restart_redis",
        ),
        # === SILENT ERRORS - Service tetap running tapi ada masalah ===
        ErrorPattern(
            name="RedisRDBError",
            pattern=r"MISCONF.*RDB|RDB.*error|can't save in background|Background saving error|bgsave.*failed",
            description="Redis RDB persistence failure - data mungkin tidak tersimpan",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_redis_persistence",
        ),
        ErrorPattern(
            name="RedisAOFError",
            pattern=r"AOF.*error|can't open.*appendonly|AOF rewrite.*failed|Bad file format reading.*aof|appendonly.*failed",
            description="Redis AOF persistence failure",
            severity=ErrorSeverity.HIGH,
            fix_function="repair_redis_persistence",
        ),
        ErrorPattern(
            name="RedisSlowLog",
            pattern=r"slowlog.*command|command took.*ms|Slow command|execution time.*exceeded",
            description="Redis slow query detected - potential performance issue",
            severity=ErrorSeverity.MEDIUM,
            fix_function="log_redis_slow_query",
        ),
        ErrorPattern(
            name="RedisMemoryFragmentation",
            pattern=r"mem_fragmentation_ratio.*[2-9]|memory fragmentation|fragmentation.*high|used_memory_rss.*exceeds",
            description="Redis memory fragmentation tinggi",
            severity=ErrorSeverity.MEDIUM,
            fix_function="repair_redis_memory",
        ),
    ]


def get_all_error_patterns() -> list[ErrorPattern]:
    """Get all error patterns combined."""
    return (
        get_docker_infrastructure_patterns() +
        get_postgres_patterns() +
        get_redis_patterns() +
        get_nginx_patterns() +
        get_cloudflared_patterns() +
        get_pgbouncer_patterns() +
        get_pgweb_patterns() +
        get_backend_patterns() +
        get_network_dns_patterns() +
        get_resource_patterns() +
        get_build_patterns() +
        get_code_error_patterns()
    )
