"""
Silent Error Fixer implementations for Omni Build System.

SRP: This module contains fix implementations for silent errors
(errors that occur while services are still running).
"""
import subprocess

from omni_build.logger import log_fix, log_info, log_success, log_warning


def _restart_container(container_name: str, timeout: int = 60) -> bool:
    """Restart a Docker container (local helper)."""
    try:
        log_info(f"Restarting {container_name}...")
        result = subprocess.run(
            ["docker", "restart", container_name],
            capture_output=True,
            text=True,
            encoding='utf-8',
            errors='replace',
            timeout=timeout,
        )
        if result.returncode == 0:
            log_success(f"{container_name} restarted")
            return True
        log_warning(f"Failed to restart {container_name}: {result.stderr}")
        return False
    except Exception:
        return False


class RedisSilentFixer:
    """Fix implementations for Redis silent errors."""
    
    @staticmethod
    def repair_redis_persistence() -> bool:
        """Repair Redis RDB/AOF persistence issues."""
        log_fix("Repairing Redis persistence...")
        
        # Try to trigger BGSAVE
        result = subprocess.run(
            ["docker", "exec", "omni-redis", "redis-cli", "BGSAVE"],
            capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=30
        )
        
        if "Background saving started" in result.stdout:
            log_success("Redis BGSAVE triggered successfully")
            return True
        
        log_warning("Redis persistence issue - restarting Redis...")
        return _restart_container("omni-redis")
    
    @staticmethod
    def log_redis_slow_query() -> bool:
        """Log Redis slow queries for investigation."""
        log_fix("Redis slow query detected...")
        log_warning("Check Redis SLOWLOG for slow commands")
        log_info("Run: docker exec omni-redis redis-cli SLOWLOG GET 10")
        return True
    
    @staticmethod
    def repair_redis_memory() -> bool:
        """Repair Redis memory fragmentation."""
        log_fix("Repairing Redis memory fragmentation...")
        
        # Try MEMORY DOCTOR
        result = subprocess.run(
            ["docker", "exec", "omni-redis", "redis-cli", "MEMORY", "DOCTOR"],
            capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=30
        )
        
        log_info(f"Redis MEMORY DOCTOR: {result.stdout[:200] if result.stdout else 'N/A'}")
        
        # If fragmentation is high, restart Redis to defragment
        log_warning("Restarting Redis to defragment memory...")
        return _restart_container("omni-redis")


class BackendSilentFixer:
    """Fix implementations for Backend silent errors."""
    
    @staticmethod
    def repair_backend_db_pool() -> bool:
        """Repair backend database connection pool exhaustion."""
        log_fix("Repairing backend DB connection pool...")
        
        # Restart backend to reset connection pool
        log_info("Restarting backend to reset connection pool...")
        return _restart_container("omni-backend", timeout=90)
    
    @staticmethod
    def log_backend_slow_query() -> bool:
        """Log backend slow queries for investigation."""
        log_fix("Backend slow query detected...")
        log_warning("Check backend logs for slow SQL queries")
        log_info("Run: docker logs omni-backend --tail 100 | grep -i slow")
        return True
    
    @staticmethod
    def log_backend_panic() -> bool:
        """Log backend panic/recover for investigation."""
        log_fix("Backend panic (recovered) detected...")
        log_warning("Go panic detected - check backend logs for stack trace")
        log_info("Run: docker logs omni-backend --tail 200 | grep -A 20 panic")
        return True


class NginxSilentFixer:
    """Fix implementations for Nginx silent errors."""
    
    @staticmethod
    def log_nginx_slow_upstream() -> bool:
        """Log Nginx slow upstream for investigation."""
        log_fix("Nginx slow upstream detected...")
        log_warning("Backend/Frontend responding slowly")
        log_info("Check backend health and database query performance")
        return True
    
    @staticmethod
    def log_nginx_buffer_issue() -> bool:
        """Log Nginx buffer overflow for investigation."""
        log_fix("Nginx buffer overflow detected...")
        log_warning("Request body too large - check client_max_body_size in nginx.conf")
        return True
    
    @staticmethod
    def log_nginx_ssl_issue() -> bool:
        """Log Nginx SSL issues for investigation."""
        log_fix("Nginx SSL error detected...")
        log_warning("SSL/TLS error - check certificate validity and configuration")
        return True


class CloudflaredSilentFixer:
    """Fix implementations for Cloudflared silent errors."""
    
    @staticmethod
    def log_cloudflared_reconnection() -> bool:
        """Log Cloudflared reconnection for monitoring."""
        log_fix("Cloudflare tunnel reconnecting...")
        log_info("Tunnel is reconnecting - this is usually temporary")
        return True
    
    @staticmethod
    def log_cloudflared_quota() -> bool:
        """Log Cloudflared quota exceeded for investigation."""
        log_fix("Cloudflare quota exceeded...")
        log_warning("Rate limit or quota exceeded - check Cloudflare dashboard")
        return True
    
    @staticmethod
    def repair_cloudflared_origin() -> bool:
        """Repair Cloudflared origin connection issues."""
        log_fix("Repairing Cloudflare origin connection...")
        
        # Check if backend and nginx are running
        for container in ["omni-backend", "omni-nginx"]:
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Status}}", container],
                capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=10
            )
            if result.returncode != 0 or result.stdout.strip() != "running":
                log_warning(f"{container} not running, restarting...")
                _restart_container(container)
        
        log_success("Origin services checked")
        return True


class PgBouncerSilentFixer:
    """Fix implementations for PgBouncer silent errors."""
    
    @staticmethod
    def repair_pgbouncer_pool() -> bool:
        """Repair PgBouncer pool saturation."""
        log_fix("Repairing PgBouncer pool saturation...")
        log_warning("Connection pool is saturated")
        
        # Restart PgBouncer to reset connections
        return _restart_container("omni-pgbouncer")
    
    @staticmethod
    def log_pgbouncer_timeout() -> bool:
        """Log PgBouncer query timeout for investigation."""
        log_fix("PgBouncer query timeout detected...")
        log_warning("Query timeout - check PostgreSQL performance")
        return True
    
    @staticmethod
    def repair_pgbouncer_disconnect() -> bool:
        """Repair PgBouncer server disconnection."""
        log_fix("Repairing PgBouncer server disconnection...")
        
        # Check PostgreSQL health first
        result = subprocess.run(
            ["docker", "exec", "omni-postgres", "pg_isready", "-U", "omni", "-d", "omni_main"],
            capture_output=True, text=True, encoding='utf-8', errors='replace', check=False, timeout=10
        )
        
        if result.returncode != 0:
            log_warning("PostgreSQL not ready, waiting...")
            return _restart_container("omni-postgres", timeout=90)
        
        return _restart_container("omni-pgbouncer")


class SystemSilentFixer:
    """Fix implementations for system-level silent errors."""
    
    @staticmethod
    def log_high_cpu() -> bool:
        """Log high CPU usage for investigation."""
        log_fix("High CPU usage detected...")
        log_warning("CPU usage is high - check container resource usage")
        log_info("Run: docker stats --no-stream")
        return True
    
    @staticmethod
    def repair_memory_pressure() -> bool:
        """Repair system memory pressure."""
        log_fix("Repairing memory pressure...")
        log_warning("System memory is under pressure")
        
        # Prune Docker to free memory
        subprocess.run(
            ["docker", "system", "prune", "-f"],
            capture_output=True, check=False, timeout=60
        )
        
        log_success("Docker system pruned to free memory")
        return True
    
    @staticmethod
    def log_disk_latency() -> bool:
        """Log disk latency for investigation."""
        log_fix("Disk latency detected...")
        log_warning("Disk I/O is slow - check storage performance")
        return True
