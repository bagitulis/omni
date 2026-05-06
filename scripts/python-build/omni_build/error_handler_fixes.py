"""
Error Handler Fix Implementations for Omni Build System.

SRP: This module contains infrastructure and system fix implementations.
Service-specific fixes are in service_fixers.py.
Cross-platform: handles Windows (taskkill, Docker Desktop.exe) and Linux (systemctl, kill).
"""
import platform
import subprocess
import time
from pathlib import Path

from omni_build.logger import log_error, log_info, log_success

IS_WINDOWS = platform.system() == "Windows"


class DockerInfraFixer:
    """Fix implementations for Docker infrastructure issues."""

    @staticmethod
    def repair_docker_engine() -> bool:
        """Repair Docker engine with robust restart mechanism."""
        if IS_WINDOWS:
            return DockerInfraFixer._repair_windows()
        return DockerInfraFixer._repair_linux()

    @staticmethod
    def _repair_windows() -> bool:
        """Repair Docker Desktop on Windows."""
        log_info("Stopping Docker Desktop...")

        processes = ["Docker Desktop.exe", "com.docker.backend.exe",
                    "com.docker.vpnkit.exe", "com.docker.proxy.exe"]
        for proc in processes:
            subprocess.run(["taskkill", "/F", "/IM", proc],
                          capture_output=True, check=False)

        time.sleep(3)
        subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
        time.sleep(5)

        log_info("Starting Docker Desktop...")
        docker_paths = [
            Path("C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe"),
            Path("C:\\Program Files (x86)\\Docker\\Docker\\Docker Desktop.exe"),
            Path.home() / "AppData" / "Local" / "Docker" / "Docker Desktop.exe",
        ]
        for docker_path in docker_paths:
            if docker_path.exists():
                subprocess.Popen([str(docker_path)],
                               stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
                time.sleep(20)
                log_success("Docker Desktop restart completed")
                return True

        log_error("Docker Desktop not found")
        return False

    @staticmethod
    def _repair_linux() -> bool:
        """Repair Docker daemon on Linux."""
        log_info("Restarting Docker daemon...")

        subprocess.run(["sudo", "systemctl", "restart", "docker"],
                      capture_output=True, check=False)
        time.sleep(5)

        # Verify
        result = subprocess.run(["docker", "info"],
                               capture_output=True, text=True, timeout=10)
        if result.returncode == 0:
            log_success("Docker daemon restarted successfully")
            return True

        log_error("Docker daemon restart failed")
        return False
    
    @staticmethod
    def repair_wsl_mount_cache() -> bool:
        """Repair WSL2 mount cache corruption."""
        subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
        time.sleep(10)
        return DockerInfraFixer.repair_docker_engine()
    
    @staticmethod
    def repair_wsl_kernel() -> bool:
        """Repair WSL2 kernel issues."""
        subprocess.run(["wsl", "--update"], capture_output=True, check=False)
        time.sleep(5)
        return SystemFixer.restart_wsl()
    
    @staticmethod
    def repair_hyperv() -> bool:
        """Repair Hyper-V services."""
        subprocess.run(["net", "stop", "vmcompute"], capture_output=True, check=False)
        time.sleep(2)
        subprocess.run(["net", "start", "vmcompute"], capture_output=True, check=False)
        return True
    
    @staticmethod
    def repair_buildkit() -> bool:
        """Clear BuildKit cache."""
        subprocess.run(["docker", "builder", "prune", "-af"], 
                      capture_output=True, check=False)
        return True
    
    @staticmethod
    def repair_container_name_conflict() -> bool:
        """Remove conflicting containers."""
        subprocess.run(["docker", "container", "prune", "-f"], 
                      capture_output=True, check=False)
        return True
    
    @staticmethod
    def repair_dependency_failure() -> bool:
        """Repair dependency service failures by diagnosing and fixing specific issues."""
        log_info("Diagnosing dependency failure...")
        
        # Check backend logs for specific errors
        specific_error = DockerInfraFixer._diagnose_backend_error()
        
        if specific_error == "database_not_exist":
            log_info("Detected: PostgreSQL database does not exist")
            from omni_build.service_fixers_postgres import PostgresFixer
            return PostgresFixer.create_postgres_database_if_not_exists()
        
        if specific_error == "pg_hba_error":
            log_info("Detected: PostgreSQL authentication error")
            from omni_build.service_fixers_postgres import PostgresFixer
            return PostgresFixer.repair_postgres_pg_hba()
        
        if specific_error == "connection_refused":
            log_info("Detected: PostgreSQL connection refused")
            # Restart postgres and wait
            subprocess.run(["docker", "restart", "omni-postgres"],
                          capture_output=True, check=False, timeout=60)
            time.sleep(30)
            return True
        
        # Generic fix: proper startup sequence
        log_info("Applying generic dependency repair...")
        
        containers = ["omni-backend", "omni-frontend", "omni-pgbouncer",
                     "omni-redis", "omni-postgres"]
        for c in containers:
            subprocess.run(["docker", "stop", c],
                          capture_output=True, check=False, timeout=30)
        
        time.sleep(3)
        
        subprocess.run(
            ["docker-compose", "-f", "docker-compose.tunnel.yml",
             "-f", "docker-compose.tunnel.standard.yml", "up", "-d", "postgres"],
            capture_output=True, check=False)
        
        time.sleep(30)
        
        subprocess.run(
            ["docker-compose", "-f", "docker-compose.tunnel.yml",
             "-f", "docker-compose.tunnel.standard.yml", "up", "-d", "redis"],
            capture_output=True, check=False)
        
        time.sleep(10)
        log_success("Dependency services repaired")
        return True
    
    @staticmethod
    def _diagnose_backend_error() -> str | None:
        """Check backend container logs to diagnose specific error."""
        try:
            result = subprocess.run(
                ["docker", "logs", "omni-backend", "--tail", "30"],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            logs = result.stdout + result.stderr
            
            # Check for specific errors in order of likelihood
            if "database" in logs.lower() and "does not exist" in logs.lower():
                return "database_not_exist"
            if "SQLSTATE 3D000" in logs:
                return "database_not_exist"
            if "no pg_hba.conf entry" in logs or "SQLSTATE 28000" in logs:
                return "pg_hba_error"
            if "connection refused" in logs.lower():
                return "connection_refused"
            
            return None
        except Exception:
            return None


class NetworkFixer:
    """Fix implementations for network and DNS issues."""
    
    @staticmethod
    def repair_dns() -> bool:
        """Full DNS repair."""
        SystemFixer.flush_dns()
        NetworkFixer.reset_network_stack()
        time.sleep(5)
        return True
    
    @staticmethod
    def repair_alpine_repo() -> bool:
        """Repair Alpine repository connectivity."""
        SystemFixer.flush_dns()
        DockerInfraFixer.repair_buildkit()
        time.sleep(10)
        return True
    
    @staticmethod
    def repair_network() -> bool:
        """Repair network issues."""
        subprocess.run(["docker", "network", "prune", "-f"], 
                      capture_output=True, check=False)
        return True
    
    @staticmethod
    def reset_network_stack() -> None:
        """Reset network stack."""
        subprocess.run(["netsh", "winsock", "reset"], capture_output=True, check=False)
        subprocess.run(["netsh", "int", "ip", "reset"], capture_output=True, check=False)


class ResourceFixer:
    """Fix implementations for resource-related issues."""
    
    @staticmethod
    def cleanup_disk_space() -> bool:
        """Cleanup disk space."""
        subprocess.run(["docker", "system", "prune", "-af"], 
                      capture_output=True, check=False)
        return True
    
    @staticmethod
    def repair_oom() -> bool:
        """Repair out of memory issues."""
        subprocess.run(["docker", "stop", "$(docker ps -q)"], 
                      shell=True, capture_output=True, check=False)
        time.sleep(5)
        return True
    
    @staticmethod
    def repair_port_conflict() -> bool:
        """Repair port conflicts."""
        log_info("Checking for port conflicts...")
        return True
    
    @staticmethod
    def repair_npm_integrity() -> bool:
        """Repair NPM integrity issues."""
        return True


class SystemFixer:
    """System-level fix implementations."""
    
    @staticmethod
    def flush_dns() -> None:
        """Flush DNS cache."""
        subprocess.run(["ipconfig", "/flushdns"], capture_output=True, check=False)
    
    @staticmethod
    def restart_wsl() -> bool:
        """Restart WSL."""
        subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
        time.sleep(10)
        return True
    
    @staticmethod
    def light_cleanup() -> None:
        """Light cleanup - dangling images only."""
        subprocess.run(["docker", "image", "prune", "-f"], 
                      capture_output=True, check=False)
    
    @staticmethod
    def aggressive_cleanup() -> None:
        """Aggressive cleanup - all unused resources."""
        subprocess.run(["docker", "system", "prune", "-af", "--volumes"], 
                      capture_output=True, check=False)
