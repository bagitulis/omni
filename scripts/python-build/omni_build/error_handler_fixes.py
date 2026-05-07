"""
Error Handler Fix Implementations for Omni Build System.

SRP: This module contains infrastructure and system fix implementations.
Service-specific fixes are in service_fixers.py.
Cross-platform: handles Windows (taskkill, Docker Desktop.exe) and Linux (systemctl, kill).
"""
import json
import platform
import shutil
import subprocess
import time
from pathlib import Path

from omni_build.logger import log_error, log_info, log_success, log_warning

IS_WINDOWS = platform.system() == "Windows"
IS_LINUX = platform.system() == "Linux"
IS_MACOS = platform.system() == "Darwin"

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

        # Try systemctl first (most common)
        restarted = False
        if shutil.which("systemctl"):
            result = subprocess.run(["sudo", "systemctl", "restart", "docker"],
                                   capture_output=True, check=False)
            restarted = result.returncode == 0

        # Fallback to service command
        if not restarted and shutil.which("service"):
            result = subprocess.run(["sudo", "service", "docker", "restart"],
                                   capture_output=True, check=False)
            restarted = result.returncode == 0

        if not restarted:
            log_error("Could not restart Docker daemon (no systemctl or service)")
            return False

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
        """Repair WSL2 mount cache corruption (Windows only, no-op on Linux)."""
        if IS_WINDOWS:
            subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
            time.sleep(10)
        return DockerInfraFixer.repair_docker_engine()
    
    @staticmethod
    def repair_wsl_kernel() -> bool:
        """Repair WSL2 kernel issues (Windows only, no-op on Linux)."""
        if IS_WINDOWS:
            subprocess.run(["wsl", "--update"], capture_output=True, check=False)
        time.sleep(5)
        return SystemFixer.restart_wsl()
    
    @staticmethod
    def repair_hyperv() -> bool:
        """Repair Hyper-V services (Windows only, no-op on Linux)."""
        if not IS_WINDOWS:
            return DockerInfraFixer.repair_docker_engine()
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
    def repair_docker_credentials() -> bool:
        """
        Fix Docker credential helper misconfiguration.
        
        Common after migrating from Windows (Docker Desktop) to Linux.
        The 'credsStore: desktop' entry references a helper that doesn't exist on Linux.
        """
        docker_config = Path.home() / ".docker" / "config.json"
        
        if not docker_config.exists():
            docker_config.parent.mkdir(parents=True, exist_ok=True)
            docker_config.write_text(json.dumps({"auths": {}}, indent=2))
            log_info("Created fresh ~/.docker/config.json")
            return True
        
        try:
            config = json.loads(docker_config.read_text())
            changed = False
            
            # Remove credsStore if it points to a non-existent helper
            if "credsStore" in config:
                creds_store = config["credsStore"]
                helper_name = f"docker-credential-{creds_store}"
                if not shutil.which(helper_name):
                    log_info(f"Removing invalid credsStore '{creds_store}' (helper not found)")
                    del config["credsStore"]
                    changed = True
            
            # Remove credHelpers entries that point to non-existent helpers
            if "credHelpers" in config:
                invalid = []
                for registry, helper in config["credHelpers"].items():
                    if not shutil.which(f"docker-credential-{helper}"):
                        invalid.append(registry)
                for reg in invalid:
                    del config["credHelpers"][reg]
                    changed = True
                if not config["credHelpers"]:
                    del config["credHelpers"]
                    changed = True
            
            # Clean up empty auths entries
            if "auths" in config:
                empty_auths = [k for k, v in config["auths"].items() if not v]
                for k in empty_auths:
                    del config["auths"][k]
                    changed = True
            
            # Fix currentContext if it references desktop-linux
            if config.get("currentContext") == "desktop-linux":
                config["currentContext"] = "default"
                changed = True
            
            if changed:
                docker_config.write_text(json.dumps(config, indent=2))
                log_success("Fixed ~/.docker/config.json")
            else:
                log_info("Docker config looks OK, clearing builder cache...")
                subprocess.run(["docker", "builder", "prune", "-af"],
                              capture_output=True, check=False)
            
            return True
        except Exception as e:
            log_error(f"Failed to fix docker config: {e}")
            # Nuclear option: reset config
            docker_config.write_text(json.dumps({"auths": {}}, indent=2))
            log_info("Reset ~/.docker/config.json to defaults")
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
        
        # Check postgres container directly first (most common root cause)
        postgres_error = DockerInfraFixer._diagnose_postgres_container()
        if postgres_error:
            return postgres_error
        
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
        
        from omni_build.subprocess_utils import get_compose_command
        compose = get_compose_command()
        
        # Use standard compose files (base + standard spec)
        compose_base = "docker-compose.tunnel.yml"
        compose_spec = "docker-compose.tunnel.standard.yml"
        
        subprocess.run(
            compose + ["-f", compose_base,
             "-f", compose_spec, "up", "-d", "postgres"],
            capture_output=True, check=False)
        
        time.sleep(30)
        
        subprocess.run(
            compose + ["-f", compose_base,
             "-f", compose_spec, "up", "-d", "redis"],
            capture_output=True, check=False)
        
        time.sleep(10)
        log_success("Dependency services repaired")
        return True
    
    @staticmethod
    def _diagnose_postgres_container() -> bool:
        """Check postgres container logs for root cause. Returns True if fix applied."""
        try:
            result = subprocess.run(
                ["docker", "logs", "omni-postgres", "--tail", "10"],
                capture_output=True, text=True, encoding='utf-8',
                errors='replace', timeout=10,
            )
            logs = result.stdout + "\n" + result.stderr
            
            # Check for ownership error (NTFS filesystem)
            if "wrong ownership" in logs or "must be started by the user that owns" in logs:
                log_info("Detected: PostgreSQL ownership error (NTFS filesystem)")
                from omni_build.service_fixers_postgres import PostgresFixer
                return PostgresFixer.repair_postgres_ownership()
            
            # Check for data corruption
            if "could not open" in logs and "pg_" in logs:
                log_info("Detected: PostgreSQL data corruption")
                from omni_build.service_fixers_postgres import PostgresFixer
                return PostgresFixer.repair_postgres_data()
            
        except Exception:
            pass
        
        return False
    
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
        """Reset network stack. Cross-platform."""
        if IS_WINDOWS:
            # netsh commands may require admin privileges — fail silently if not elevated
            subprocess.run(["netsh", "winsock", "reset"], capture_output=True, check=False)
            subprocess.run(["netsh", "int", "ip", "reset"], capture_output=True, check=False)
        else:
            # Linux: restart Docker's network
            subprocess.run(["docker", "network", "prune", "-f"], capture_output=True, check=False)
            if shutil.which("systemctl"):
                subprocess.run(["sudo", "systemctl", "restart", "docker"], capture_output=True, check=False)
            elif shutil.which("service"):
                subprocess.run(["sudo", "service", "docker", "restart"], capture_output=True, check=False)


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
        """Repair out of memory issues by stopping all containers, then restarting Docker."""
        log_info("Stopping all containers to free memory...")
        try:
            result = subprocess.run(
                ["docker", "ps", "-q"],
                capture_output=True, text=True, timeout=10,
            )
            if result.returncode == 0 and result.stdout.strip():
                for container_id in result.stdout.strip().split('\n'):
                    container_id = container_id.strip()
                    if container_id:
                        subprocess.run(
                            ["docker", "stop", container_id],
                            capture_output=True, check=False, timeout=30,
                        )
        except Exception as e:
            log_warning(f"Error stopping containers: {e}")
        
        time.sleep(5)
        
        # Prune stopped containers and dangling images to free memory
        subprocess.run(["docker", "container", "prune", "-f"],
                      capture_output=True, check=False)
        subprocess.run(["docker", "image", "prune", "-f"],
                      capture_output=True, check=False)
        
        log_info("Restarting Docker engine after OOM cleanup...")
        return DockerInfraFixer.repair_docker_engine()
    
    @staticmethod
    def repair_port_conflict() -> bool:
        """Repair port conflicts by finding and reporting conflicting processes."""
        log_info("Checking for port conflicts...")
        
        # Check common ports used by Omni
        ports_to_check = [80, 443, 3000, 5432, 6379]
        conflicts_found = False
        
        for port in ports_to_check:
            try:
                if IS_WINDOWS:
                    result = subprocess.run(
                        ["netstat", "-ano"],
                        capture_output=True, text=True, timeout=10,
                    )
                    for line in result.stdout.split('\n'):
                        if f":{port} " in line and "LISTENING" in line:
                            log_warning(f"Port {port} is in use: {line.strip()}")
                            conflicts_found = True
                else:
                    result = subprocess.run(
                        ["ss", "-tlnp"],
                        capture_output=True, text=True, timeout=10,
                    )
                    for line in result.stdout.split('\n'):
                        if f":{port} " in line:
                            log_warning(f"Port {port} is in use: {line.strip()}")
                            conflicts_found = True
            except Exception:
                pass
        
        if conflicts_found:
            log_info("Stopping Omni containers to release ports...")
            subprocess.run(["docker", "container", "prune", "-f"],
                          capture_output=True, check=False)
        else:
            log_info("No port conflicts detected")
        
        return True
    
    @staticmethod
    def repair_npm_integrity() -> bool:
        """Repair NPM integrity issues by cleaning cache and node_modules."""
        log_info("Repairing NPM integrity...")
        
        # Find npm
        npm_cmd = "npm.cmd" if IS_WINDOWS else "npm"
        npm_path = shutil.which(npm_cmd) or shutil.which("npm")
        
        if npm_path:
            subprocess.run(
                [npm_path, "cache", "clean", "--force"],
                capture_output=True, check=False, timeout=60,
            )
            log_success("NPM cache cleaned")
        else:
            log_warning("npm not found in PATH, skipping cache clean")
        
        return True


class SystemFixer:
    """System-level fix implementations."""
    
    @staticmethod
    def flush_dns() -> None:
        """Flush DNS cache. Cross-platform."""
        if IS_WINDOWS:
            subprocess.run(["ipconfig", "/flushdns"], capture_output=True, check=False)
        elif IS_MACOS:
            subprocess.run(["sudo", "dscacheutil", "-flushcache"], capture_output=True, check=False)
        else:
            # Linux: try systemd-resolve first, then resolvectl
            if shutil.which("systemd-resolve"):
                subprocess.run(["sudo", "systemd-resolve", "--flush-caches"], capture_output=True, check=False)
            elif shutil.which("resolvectl"):
                subprocess.run(["sudo", "resolvectl", "flush-caches"], capture_output=True, check=False)
    
    @staticmethod
    def restart_wsl() -> bool:
        """Restart WSL (Windows only). On Linux, restart Docker daemon instead."""
        if IS_WINDOWS:
            subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
            time.sleep(10)
        else:
            # Linux: restart docker daemon instead
            if shutil.which("systemctl"):
                subprocess.run(["sudo", "systemctl", "restart", "docker"], capture_output=True, check=False)
            elif shutil.which("service"):
                subprocess.run(["sudo", "service", "docker", "restart"], capture_output=True, check=False)
            time.sleep(5)
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
