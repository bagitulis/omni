"""
Fix Methods Registry for Omni Build System.

SRP: This module ONLY contains the mapping of fix function names to implementations.
"""
from omni_build.error_handler_fixes import (
    DockerInfraFixer,
    NetworkFixer,
    ResourceFixer,
)
from omni_build.service_fixers import (
    BackendFixer,
    CloudflaredFixer,
    NginxFixer,
    PgBouncerFixer,
    PgWebFixer,
    PostgresFixer,
    RedisFixer,
)
from omni_build.service_fixers_silent import (
    BackendSilentFixer,
    CloudflaredSilentFixer,
    NginxSilentFixer,
    PgBouncerSilentFixer,
    RedisSilentFixer,
    SystemSilentFixer,
)


def get_fix_methods() -> dict:
    """Get all fix method mappings."""
    return {
        # Docker infrastructure
        "repair_docker_engine": DockerInfraFixer.repair_docker_engine,
        "repair_docker_pipe": DockerInfraFixer.repair_docker_engine,
        "repair_wsl_mount_cache": DockerInfraFixer.repair_wsl_mount_cache,
        "repair_wsl_kernel": DockerInfraFixer.repair_wsl_kernel,
        "repair_hyperv": DockerInfraFixer.repair_hyperv,
        "switch_docker_to_linux": DockerInfraFixer.repair_docker_engine,
        "repair_buildkit": DockerInfraFixer.repair_buildkit,
        "repair_docker_credentials": DockerInfraFixer.repair_docker_credentials,
        "repair_container_name_conflict": DockerInfraFixer.repair_container_name_conflict,
        "repair_dependency_failure": DockerInfraFixer.repair_dependency_failure,
        
        # PostgreSQL
        "repair_postgres_data": PostgresFixer.repair_postgres_data,
        "repair_postgres_pg_hba": PostgresFixer.repair_postgres_pg_hba,
        "repair_postgres_database": PostgresFixer.repair_postgres_database,
        "create_postgres_database_if_not_exists": PostgresFixer.create_postgres_database_if_not_exists,
        "repair_docker_dns_postgres": PostgresFixer.repair_docker_dns_postgres,
        
        # Redis
        "restart_redis": RedisFixer.restart_redis,
        
        # Nginx
        "restart_nginx": NginxFixer.restart_nginx,
        "repair_nginx_upstream": NginxFixer.repair_nginx_upstream,
        
        # Cloudflared
        "restart_cloudflared": CloudflaredFixer.restart_cloudflared,
        
        # PgBouncer
        "restart_pgbouncer": PgBouncerFixer.restart_pgbouncer,
        
        # PgWeb
        "restart_pgweb": PgWebFixer.restart_pgweb,
        
        # Backend
        "wait_for_backend_init": BackendFixer.wait_for_backend_init,
        "restart_unhealthy_service": BackendFixer.restart_unhealthy_service,
        
        # Network/DNS
        "repair_dns": NetworkFixer.repair_dns,
        "repair_alpine_repo": NetworkFixer.repair_alpine_repo,
        "repair_docker_registry": NetworkFixer.repair_dns,
        "repair_network": NetworkFixer.repair_network,
        
        # Resources
        "cleanup_disk_space": ResourceFixer.cleanup_disk_space,
        "repair_oom": ResourceFixer.repair_oom,
        "repair_port_conflict": ResourceFixer.repair_port_conflict,
        
        # Build
        "repair_npm_integrity": ResourceFixer.repair_npm_integrity,
        
        # PostgreSQL Silent Errors
        "repair_postgres_io": PostgresFixer.repair_postgres_io,
        "repair_postgres_checkpoint": PostgresFixer.repair_postgres_checkpoint,
        "repair_postgres_wal": PostgresFixer.repair_postgres_wal,
        "repair_postgres_deadlock": PostgresFixer.repair_postgres_deadlock,
        "repair_postgres_connections": PostgresFixer.repair_postgres_connections,
        "repair_postgres_replication": PostgresFixer.repair_postgres_replication,
        
        # Redis Silent Errors
        "repair_redis_persistence": RedisSilentFixer.repair_redis_persistence,
        "log_redis_slow_query": RedisSilentFixer.log_redis_slow_query,
        "repair_redis_memory": RedisSilentFixer.repair_redis_memory,
        
        # Backend Silent Errors
        "restart_backend": BackendFixer.restart_backend,
        "repair_backend_db_pool": BackendSilentFixer.repair_backend_db_pool,
        "log_backend_slow_query": BackendSilentFixer.log_backend_slow_query,
        "log_backend_panic": BackendSilentFixer.log_backend_panic,
        
        # Nginx Silent Errors
        "log_nginx_slow_upstream": NginxSilentFixer.log_nginx_slow_upstream,
        "log_nginx_buffer_issue": NginxSilentFixer.log_nginx_buffer_issue,
        "log_nginx_ssl_issue": NginxSilentFixer.log_nginx_ssl_issue,
        
        # Cloudflared Silent Errors
        "log_cloudflared_reconnection": CloudflaredSilentFixer.log_cloudflared_reconnection,
        "log_cloudflared_quota": CloudflaredSilentFixer.log_cloudflared_quota,
        "repair_cloudflared_origin": CloudflaredSilentFixer.repair_cloudflared_origin,
        
        # PgBouncer Silent Errors
        "repair_pgbouncer_pool": PgBouncerSilentFixer.repair_pgbouncer_pool,
        "log_pgbouncer_timeout": PgBouncerSilentFixer.log_pgbouncer_timeout,
        "repair_pgbouncer_disconnect": PgBouncerSilentFixer.repair_pgbouncer_disconnect,
        
        # System Silent Errors
        "log_high_cpu": SystemSilentFixer.log_high_cpu,
        "repair_memory_pressure": SystemSilentFixer.repair_memory_pressure,
        "log_disk_latency": SystemSilentFixer.log_disk_latency,
    }
