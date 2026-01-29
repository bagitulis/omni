"""
Service Fixer implementations for Omni Build System.

SRP: This module contains base and lightweight service fixers.
PostgreSQL fixes are in service_fixers_postgres.py.
Silent error fixes are in service_fixers_silent.py.
"""
import subprocess
import time

from omni_build.logger import log_error, log_fix, log_info, log_success, log_warning


class ServiceFixer:
    """
    Base class for service-specific fix implementations.
    
    Implements DRY principle - common operations are shared.
    """
    
    @staticmethod
    def restart_container(container_name: str, timeout: int = 60) -> bool:
        """Restart a Docker container."""
        try:
            log_info(f"Restarting {container_name}...")
            result = subprocess.run(
                ["docker", "restart", container_name],
                capture_output=True,
                text=True,
                timeout=timeout,
            )
            if result.returncode == 0:
                log_success(f"{container_name} restarted")
                return True
            log_warning(f"Failed to restart {container_name}: {result.stderr}")
            return False
        except Exception as e:
            log_error(f"Error restarting {container_name}: {e}")
            return False
    
    @staticmethod
    def stop_container(container_name: str, timeout: int = 30) -> bool:
        """Stop a Docker container."""
        try:
            subprocess.run(
                ["docker", "stop", container_name],
                capture_output=True,
                timeout=timeout,
                check=False
            )
            return True
        except Exception:
            return False
    
    @staticmethod
    def remove_container(container_name: str) -> bool:
        """Remove a Docker container."""
        try:
            subprocess.run(
                ["docker", "rm", "-f", container_name],
                capture_output=True,
                check=False
            )
            return True
        except Exception:
            return False


class RedisFixer(ServiceFixer):
    """Fix implementations for Redis service."""
    
    @staticmethod
    def restart_redis() -> bool:
        """Restart Redis container."""
        log_fix("Restarting Redis service...")
        return ServiceFixer.restart_container("omni-redis")


class NginxFixer(ServiceFixer):
    """Fix implementations for Nginx service."""
    
    @staticmethod
    def restart_nginx() -> bool:
        """Restart Nginx container."""
        log_fix("Restarting Nginx service...")
        return ServiceFixer.restart_container("omni-nginx")
    
    @staticmethod
    def repair_nginx_upstream() -> bool:
        """Repair Nginx upstream issues by restarting nginx after backends."""
        log_fix("Repairing Nginx upstream configuration...")
        
        log_info("Checking upstream services...")
        
        for container in ["omni-backend", "omni-frontend"]:
            result = subprocess.run(
                ["docker", "inspect", "--format", "{{.State.Status}}", container],
                capture_output=True,
                text=True,
                timeout=10,
                check=False
            )
            if result.returncode != 0 or result.stdout.strip() != "running":
                log_warning(f"{container} is not running, restarting...")
                ServiceFixer.restart_container(container)
        
        time.sleep(5)
        return NginxFixer.restart_nginx()


class CloudflaredFixer(ServiceFixer):
    """Fix implementations for Cloudflare tunnel service."""
    
    @staticmethod
    def restart_cloudflared() -> bool:
        """Restart Cloudflared tunnel container."""
        log_fix("Restarting Cloudflare tunnel...")
        return ServiceFixer.restart_container("omni-cloudflared", timeout=90)


class PgBouncerFixer(ServiceFixer):
    """Fix implementations for PgBouncer connection pooler."""
    
    @staticmethod
    def restart_pgbouncer() -> bool:
        """Restart PgBouncer container."""
        log_fix("Restarting PgBouncer connection pooler...")
        
        result = subprocess.run(
            ["docker", "inspect", "--format", "{{.State.Health.Status}}", "omni-postgres"],
            capture_output=True,
            text=True,
            timeout=10,
            check=False
        )
        
        if result.returncode == 0 and result.stdout.strip() == "healthy":
            log_info("PostgreSQL is healthy, restarting PgBouncer...")
        else:
            log_warning("PostgreSQL not healthy, waiting 10s before PgBouncer restart...")
            time.sleep(10)
        
        return ServiceFixer.restart_container("omni-pgbouncer")


class PgWebFixer(ServiceFixer):
    """Fix implementations for PgWeb database viewer."""
    
    @staticmethod
    def restart_pgweb() -> bool:
        """Restart PgWeb container."""
        log_fix("Restarting PgWeb database viewer...")
        return ServiceFixer.restart_container("omni-pgweb")


class BackendFixer(ServiceFixer):
    """Fix implementations for Go backend service."""
    
    @staticmethod
    def wait_for_backend_init() -> bool:
        """Wait for Go backend to finish initializing."""
        log_fix("Waiting for Go backend to initialize (tenant DB connections)...")
        log_info("Go backend needs ~2-3 minutes for first run (PostgreSQL tenant setup)")
        
        wait_times = [30, 60, 90]
        
        for wait_time in wait_times:
            log_info(f"Waiting {wait_time}s for backend initialization...")
            time.sleep(wait_time)
            
            try:
                import requests
                response = requests.get("http://localhost:3000/api/health", timeout=10)
                if response.status_code == 200:
                    log_success(f"Backend is now healthy after {wait_time}s wait")
                    return True
            except Exception:
                pass
        
        log_warning("Backend still not responding - may need manual investigation")
        return False
    
    @staticmethod
    def restart_unhealthy_service() -> bool:
        """Restart unhealthy containers."""
        log_fix("Restarting unhealthy containers...")
        
        result = subprocess.run(
            ["docker", "ps", "--filter", "health=unhealthy", "--format", "{{.Names}}"],
            capture_output=True,
            text=True,
            check=False,
        )
        
        if result.returncode == 0 and result.stdout.strip():
            containers = result.stdout.strip().split('\n')
            for container in containers:
                log_info(f"Restarting unhealthy: {container}")
                ServiceFixer.restart_container(container)
            
            log_info("Waiting 30s for restarted containers to initialize...")
            time.sleep(30)
            return True
        
        return False
    
    @staticmethod
    def restart_backend() -> bool:
        """Restart backend container."""
        log_fix("Restarting Go backend...")
        return ServiceFixer.restart_container("omni-backend", timeout=90)


# Re-export PostgresFixer from separate module for backward compatibility
from omni_build.service_fixers_postgres import PostgresFixer

__all__ = [
    'ServiceFixer',
    'RedisFixer', 
    'NginxFixer',
    'CloudflaredFixer',
    'PgBouncerFixer',
    'PgWebFixer',
    'BackendFixer',
    'PostgresFixer',
]
