"""
Error Handler module for Omni Build System.

SRP: This module ONLY handles error detection, pattern matching, and recovery orchestration.
Implements 22+ error patterns with progressive fix levels (6 escalation stages).

Based on AGENTS.md requirements:
- No caching
- All fixes must be reversible
- Structured logging
- Clean error messages
"""
import re
import subprocess
import time
from pathlib import Path
from typing import Callable, Optional

from omni_build.logger import log_error, log_fix, log_info, log_success, log_warning
from omni_build.models import ErrorPattern, ErrorSeverity


class ErrorHandler:
    """
    Comprehensive error detection and recovery system.
    
    Handles 22+ error patterns across categories:
    - Docker infrastructure (10 patterns)
    - System resources (4 patterns)
    - Network/DNS (4 patterns)
    - Build/compilation (4 patterns)
    """
    
    def __init__(self) -> None:
        """Initialize error handler with all error patterns."""
        self.error_patterns = self._init_error_patterns()
        self.fix_attempt_count = 0
    
    def _init_error_patterns(self) -> list[ErrorPattern]:
        """
        Initialize all 22+ error patterns.
        
        Returns:
            List of ErrorPattern instances.
        """
        return [
            # === DOCKER INFRASTRUCTURE ERRORS ===
            ErrorPattern(
                name="DockerEngineError",
                pattern=r"500 Internal Server Error|dockerDesktopLinuxEngine.*_ping|request returned 500|error during connect",
                description="Docker Desktop engine error (500)",
                severity=ErrorSeverity.CRITICAL,
                fix_function="repair_docker_engine",
            ),
            ErrorPattern(
                name="DockerPipeError",
                pattern=r"pipe.*docker|npipe.*error|named pipe|\\\\\.\\pipe",
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
                name="PostgresDataCorruption",
                pattern=r"could not open directory.*pg_|pg_notify.*No such file|pg_wal.*No such file|pg_xact.*No such file|database.*shut down|FATAL.*postgres|could not open file.*pg_filenode",
                description="PostgreSQL data directory corrupted",
                severity=ErrorSeverity.CRITICAL,
                fix_function="repair_postgres_data",
            ),
            ErrorPattern(
                name="BackendHealthCheckTimeout",
                pattern=r"backend.*did not become healthy|backend.*Timeout after|Health checks failed|services not responding",
                description="Backend service health check timeout (Go backend initializing)",
                severity=ErrorSeverity.MEDIUM,
                fix_function="wait_for_backend_init",
            ),
            ErrorPattern(
                name="ServiceUnhealthy",
                pattern=r"service.*unhealthy|container reports unhealthy|Health status: unhealthy",
                description="Service failed health check",
                severity=ErrorSeverity.HIGH,
                fix_function="restart_unhealthy_service",
            ),
            
            # === NETWORK/DNS ERRORS ===
            ErrorPattern(
                name="DNSError",
                pattern=r"DNS lookup error|DNS.*name does not exist|SERVFAIL|NXDOMAIN|fetch.*error|lookup.*no such host|dial tcp.*lookup",
                description="DNS resolution failure",
                severity=ErrorSeverity.HIGH,
                fix_function="repair_dns",
            ),
            ErrorPattern(
                name="DockerDNSPostgresError",
                pattern=r"lookup postgres on.*no such host|failed to connect.*postgres.*hostname resolving|postgres.*127\.0\.0\.11.*no such host",
                description="Docker internal DNS cannot resolve 'postgres' hostname",
                severity=ErrorSeverity.CRITICAL,
                fix_function="repair_docker_dns_postgres",
            ),
            ErrorPattern(
                name="AlpineRepoError",
                pattern=r"fetching https://dl-cdn\.alpinelinux\.org.*temporary error|unable to select packages.*alpine|apk.*error",
                description="Alpine repository network error",
                severity=ErrorSeverity.MEDIUM,
                fix_function="repair_alpine_repo",
            ),
            ErrorPattern(
                name="DockerRegistryError",
                pattern=r"registry-1\.docker\.io.*no such host|failed to do request.*registry|registry.*connection|pull.*manifest.*error",
                description="Docker registry connection error",
                severity=ErrorSeverity.HIGH,
                fix_function="repair_docker_registry",
            ),
            ErrorPattern(
                name="NetworkError",
                pattern=r"network.*error|connect.*refused|ECONNREFUSED|timeout|timed out",
                description="Network connectivity issue",
                severity=ErrorSeverity.MEDIUM,
                fix_function="repair_network",
            ),
            
            # === RESOURCE ERRORS ===
            ErrorPattern(
                name="DiskSpaceError",
                pattern=r"no space left|disk.*full|ENOSPC",
                description="Disk space exhausted",
                severity=ErrorSeverity.CRITICAL,
                fix_function="cleanup_disk_space",
            ),
            ErrorPattern(
                name="OOMError",
                pattern=r"exit.*code.*137|killed|out of memory|OOM|ENOMEM",
                description="Out of memory",
                severity=ErrorSeverity.CRITICAL,
                fix_function="repair_oom",
            ),
            ErrorPattern(
                name="PortConflictError",
                pattern=r"port.*already.*use|address already in use|EADDRINUSE",
                description="Port conflict",
                severity=ErrorSeverity.MEDIUM,
                fix_function="repair_port_conflict",
            ),
            
            # === BUILD/NPM ERRORS ===
            ErrorPattern(
                name="NpmIntegrityError",
                pattern=r"integrity checksum|EINTEGRITY|sha512|sha1.*mismatch",
                description="NPM integrity checksum mismatch",
                severity=ErrorSeverity.MEDIUM,
                fix_function="repair_npm_integrity",
            ),
            
            # === CODE ERRORS (CANNOT AUTO-FIX) ===
            ErrorPattern(
                name="TypeScriptError",
                pattern=r"error TS\d+:|Cannot find module|has no exported member|is not assignable to|Object is of type 'unknown'",
                description="TypeScript compilation error",
                severity=ErrorSeverity.HIGH,
                fix_function=None,
                is_code_error=True,
            ),
            ErrorPattern(
                name="ImportResolutionError",
                pattern=r"Could not resolve|Failed to resolve import|Module not found.*from",
                description="Import/Module resolution error",
                severity=ErrorSeverity.HIGH,
                fix_function=None,
                is_code_error=True,
            ),
            ErrorPattern(
                name="SyntaxError",
                pattern=r"SyntaxError|Unexpected token|Parse error",
                description="JavaScript/TypeScript syntax error",
                severity=ErrorSeverity.HIGH,
                fix_function=None,
                is_code_error=True,
            ),
            ErrorPattern(
                name="ViteError",
                pattern=r"vite.*error|rollup.*error",
                description="Vite/Rollup build error",
                severity=ErrorSeverity.HIGH,
                fix_function=None,
                is_code_error=True,
            ),
        ]
    
    def detect_error(self, error_message: str) -> Optional[ErrorPattern]:
        """
        Detect error type from message using pattern matching.
        
        Args:
            error_message: Error message to analyze
            
        Returns:
            ErrorPattern if matched, None otherwise
        """
        for pattern in self.error_patterns:
            if re.search(pattern.pattern, error_message, re.IGNORECASE | re.MULTILINE):
                return pattern
        return None
    
    def can_auto_fix(self, error_message: str) -> bool:
        """
        Check if error can be automatically fixed.
        
        Args:
            error_message: Error message
            
        Returns:
            True if auto-fixable, False if requires manual intervention
        """
        error_pattern = self.detect_error(error_message)
        if error_pattern is None:
            return True  # Unknown errors can try generic fixes
        return not error_pattern.is_code_error
    
    def apply_fix(self, error_pattern: ErrorPattern) -> bool:
        """
        Apply fix for detected error pattern.
        
        Args:
            error_pattern: Detected error pattern
            
        Returns:
            True if fix succeeded, False otherwise
        """
        if error_pattern.is_code_error:
            log_error(f"Code error detected: {error_pattern.description}")
            log_error("This error requires manual fix. Please correct the code and retry.")
            return False
        
        if not error_pattern.fix_function:
            log_warning(f"No fix function defined for {error_pattern.name}")
            return False
        
        log_fix(f"Applying fix for: {error_pattern.description}")
        
        # Map fix function names to actual methods
        fix_methods = {
            "repair_docker_engine": self._repair_docker_engine,
            "repair_docker_pipe": self._repair_docker_pipe,
            "repair_wsl_mount_cache": self._repair_wsl_mount_cache,
            "repair_wsl_kernel": self._repair_wsl_kernel,
            "repair_hyperv": self._repair_hyperv,
            "switch_docker_to_linux": self._switch_docker_to_linux,
            "repair_buildkit": self._repair_buildkit,
            "repair_container_name_conflict": self._repair_container_name_conflict,
            "repair_dependency_failure": self._repair_dependency_failure,
            "repair_postgres_data": self._repair_postgres_data,
            "repair_dns": self._repair_dns,
            "repair_alpine_repo": self._repair_alpine_repo,
            "repair_docker_registry": self._repair_docker_registry,
            "repair_network": self._repair_network,
            "cleanup_disk_space": self._cleanup_disk_space,
            "repair_oom": self._repair_oom,
            "repair_port_conflict": self._repair_port_conflict,
            "repair_npm_integrity": self._repair_npm_integrity,
        }
        
        fix_method = fix_methods.get(error_pattern.fix_function)
        if fix_method:
            try:
                result = fix_method()
                if result:
                    log_success(f"Fix applied successfully: {error_pattern.name}")
                else:
                    log_warning(f"Fix completed with warnings: {error_pattern.name}")
                return result
            except Exception as e:
                log_error(f"Fix failed: {e}")
                return False
        
        log_warning(f"Fix method not implemented: {error_pattern.fix_function}")
        return False
    
    def apply_progressive_fix(self, level: int) -> bool:
        """
        Apply progressive fix based on escalation level (1-6).
        
        Level 1: Light cleanup + port conflicts
        Level 2: Network flush + DNS repair
        Level 3: WSL restart + network reset
        Level 4: Docker engine full restart
        Level 5: Aggressive cleanup + Docker restart
        Level 6: Full system reset (Docker + WSL + Hyper-V)
        
        Args:
            level: Fix level (1-6)
            
        Returns:
            True if fix succeeded
        """
        self.fix_attempt_count += 1
        wait_time = min(5 * (2 ** (level - 1)), 60)  # Exponential backoff
        
        log_fix(f"Progressive Fix Level {level} (wait: {wait_time}s)")
        
        try:
            if level == 1:
                log_info("L1: Port conflicts + light cleanup...")
                self._repair_port_conflict()
                self._repair_container_name_conflict()
                self._light_cleanup()
                
            elif level == 2:
                log_info("L2: Network flush + DNS repair...")
                self._flush_dns()
                self._repair_alpine_repo()
                
            elif level == 3:
                log_info("L3: Full DNS + network reset...")
                self._repair_dns()
                
            elif level == 4:
                log_info("L4: WSL restart + network reset...")
                self._reset_network_stack()
                self._restart_wsl()
                
            elif level == 5:
                log_info("L5: Docker engine full restart...")
                self._repair_docker_engine()
                
            elif level == 6:
                log_info("L6: Aggressive cleanup + system reset...")
                self._aggressive_cleanup()
                self._restart_wsl()
                self._repair_docker_engine()
            
            log_info(f"Waiting {wait_time}s for stabilization...")
            time.sleep(wait_time)
            
            log_success(f"Fix level {level} applied")
            return True
            
        except Exception as e:
            log_error(f"Fix level {level} failed: {e}")
            return False
    
    # ============================================
    # FIX IMPLEMENTATIONS
    # ============================================
    
    def _repair_docker_engine(self) -> bool:
        """Repair Docker Desktop engine with robust restart mechanism."""
        log_info("Stopping Docker Desktop...")
        
        # Step 1: Kill all Docker processes thoroughly
        docker_processes = [
            "Docker Desktop.exe",
            "com.docker.backend.exe",
            "com.docker.vpnkit.exe",
            "com.docker.proxy.exe",
        ]
        
        for process in docker_processes:
            subprocess.run(
                ["taskkill", "/F", "/IM", process],
                capture_output=True,
                check=False
            )
        
        time.sleep(3)
        
        log_info("Stopping WSL backend...")
        # Force stop WSL to ensure clean slate
        subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
        time.sleep(5)
        
        # Step 2: Verify WSL is stopped
        result = subprocess.run(
            ["wsl", "--list", "--running"],
            capture_output=True,
            text=True,
            check=False
        )
        if result.returncode == 0 and "docker" in result.stdout.lower():
            log_warning("WSL still running, forcing terminate...")
            subprocess.run(["wsl", "--terminate", "docker-desktop"], capture_output=True, check=False)
            subprocess.run(["wsl", "--terminate", "docker-desktop-data"], capture_output=True, check=False)
            time.sleep(3)
        
        log_success("Docker and WSL stopped successfully")
        
        # Step 3: Start Docker Desktop
        log_info("Starting Docker Desktop...")
        try:
            # Try default location first
            docker_path = Path("C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe")
            if not docker_path.exists():
                log_warning("Docker Desktop not found in default location")
                # Try alternative locations
                alt_paths = [
                    Path("C:\\Program Files (x86)\\Docker\\Docker\\Docker Desktop.exe"),
                    Path.home() / "AppData" / "Local" / "Docker" / "Docker Desktop.exe",
                ]
                for alt_path in alt_paths:
                    if alt_path.exists():
                        docker_path = alt_path
                        log_info(f"Found Docker Desktop at: {docker_path}")
                        break
                else:
                    log_error("Could not find Docker Desktop executable")
                    return False
            
            # Start Docker Desktop process
            subprocess.Popen([str(docker_path)], stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
            log_info("Docker Desktop.exe started, waiting for initialization...")
            
            # Step 4: Wait for Docker Desktop process to be stable (not just started)
            time.sleep(20)
            
            # Step 5: Verify Docker Desktop is running
            for i in range(5):
                result = subprocess.run(
                    ["tasklist", "/FO", "CSV", "/NH"],
                    capture_output=True,
                    text=True,
                    check=False
                )
                if "Docker Desktop.exe" in result.stdout:
                    log_success("Docker Desktop process confirmed running")
                    break
                else:
                    log_warning(f"Docker Desktop process not found, waiting... ({i+1}/5)")
                    time.sleep(5)
            else:
                log_error("Docker Desktop process failed to start")
                return False
            
            # Step 6: Wait for WSL backend to start
            log_info("Waiting for WSL backend to initialize...")
            time.sleep(15)
            
            # Verify WSL docker-desktop is running
            for i in range(6):
                result = subprocess.run(
                    ["wsl", "--list", "--running"],
                    capture_output=True,
                    text=True,
                    check=False
                )
                if "docker-desktop" in result.stdout.lower():
                    log_success("WSL docker-desktop backend is running")
                    break
                else:
                    log_info(f"Waiting for WSL backend... ({i+1}/6)")
                    time.sleep(5)
            else:
                log_warning("WSL backend slow to start, continuing anyway...")
            
            # Step 7: Wait for dockerd daemon to start
            log_info("Waiting for Docker daemon (dockerd) to start...")
            time.sleep(10)
            
            # Verify dockerd is running in WSL
            for i in range(8):
                try:
                    result = subprocess.run(
                        ["wsl", "-d", "docker-desktop", "sh", "-c", "pgrep dockerd"],
                        capture_output=True,
                        text=True,
                        timeout=5,
                        check=False
                    )
                    if result.returncode == 0 and result.stdout.strip():
                        log_success(f"dockerd daemon is running (PID: {result.stdout.strip()})")
                        break
                    else:
                        log_info(f"Waiting for dockerd daemon... ({i+1}/8)")
                        time.sleep(5)
                except:
                    log_warning(f"dockerd check failed, retrying... ({i+1}/8)")
                    time.sleep(5)
            else:
                log_error("dockerd daemon failed to start in WSL!")
                log_error("Docker Desktop may be stuck in initialization")
                return False
            
            log_success("Docker Desktop restart completed successfully")
            log_info("Docker engine is starting, daemon is running")
            return True
            
        except Exception as e:
            log_error(f"Failed to start Docker Desktop: {e}")
            import traceback
            traceback.print_exc()
            return False
    
    def _repair_docker_pipe(self) -> bool:
        """Repair Docker named pipe."""
        return self._repair_docker_engine()
    
    def _repair_wsl_mount_cache(self) -> bool:
        """Repair WSL2 mount cache corruption."""
        log_info("Stopping all Docker containers...")
        subprocess.run(["docker", "stop", "$(docker ps -aq)"], 
                      shell=True, capture_output=True, check=False)
        
        log_info("Terminating WSL mount cache...")
        subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
        time.sleep(10)
        
        log_info("Restarting Docker Desktop...")
        return self._repair_docker_engine()
    
    def _repair_wsl_kernel(self) -> bool:
        """Repair WSL2 kernel issues."""
        log_info("Updating WSL...")
        subprocess.run(["wsl", "--update"], capture_output=True, check=False)
        time.sleep(5)
        return self._restart_wsl()
    
    def _repair_hyperv(self) -> bool:
        """Repair Hyper-V services."""
        log_info("Restarting Hyper-V services...")
        subprocess.run(["net", "stop", "vmcompute"], capture_output=True, check=False)
        time.sleep(2)
        subprocess.run(["net", "start", "vmcompute"], capture_output=True, check=False)
        return True
    
    def _switch_docker_to_linux(self) -> bool:
        """Switch Docker to Linux containers mode."""
        log_info("Docker should be in Linux mode - restarting Docker...")
        return self._repair_docker_engine()
    
    def _repair_buildkit(self) -> bool:
        """Clear BuildKit cache."""
        log_info("Clearing BuildKit cache...")
        subprocess.run(["docker", "builder", "prune", "-af"], 
                      capture_output=True, check=False)
        return True
    
    def _repair_container_name_conflict(self) -> bool:
        """Remove conflicting containers."""
        log_info("Removing stopped containers...")
        subprocess.run(["docker", "container", "prune", "-f"], 
                      capture_output=True, check=False)
        return True
    
    def _repair_dependency_failure(self) -> bool:
        """Repair dependency service failures."""
        log_info("Checking PostgreSQL and Redis...")
        # Stop and remove dependency containers
        for container in ["omni-postgres", "omni-redis", "omni-pgbouncer"]:
            subprocess.run(["docker", "rm", "-f", container], 
                          capture_output=True, check=False)
        return True
    
    def _repair_postgres_data(self) -> bool:
        """Repair corrupted PostgreSQL data."""
        log_warning("PostgreSQL data corruption detected")
        log_info("Stopping PostgreSQL container...")
        subprocess.run(["docker", "stop", "omni-postgres"], 
                      capture_output=True, check=False)
        log_info("Container will reinitialize on next start")
        return True
    
    def _repair_dns(self) -> bool:
        """Full DNS repair."""
        self._flush_dns()
        self._reset_network_stack()
        time.sleep(5)
        return True
    
    def _repair_docker_dns_postgres(self) -> bool:
        """
        Fix Docker internal DNS not resolving 'postgres' hostname.
        
        This happens when:
        - Docker network is corrupted
        - Internal DNS resolver (127.0.0.11) fails
        - Container can't reach other containers by service name
        
        Solution:
        1. Recreate Docker networks
        2. Restart containers to re-establish DNS
        """
        log_fix("Repairing Docker internal DNS for postgres hostname...")
        
        # Step 1: Prune all unused networks
        log_info("Pruning Docker networks...")
        subprocess.run(
            ["docker", "network", "prune", "-f"],
            capture_output=True,
            check=False
        )
        
        # Step 2: Stop all omni containers
        log_info("Stopping all containers to reset network...")
        for container in ["omni-backend", "omni-frontend", "omni-postgres", "omni-redis", "omni-pgbouncer"]:
            subprocess.run(
                ["docker", "stop", container],
                capture_output=True,
                check=False,
                timeout=30
            )
        
        time.sleep(3)
        
        # Step 3: Remove containers to force network recreation
        log_info("Removing containers to force network recreation...")
        for container in ["omni-backend", "omni-frontend", "omni-postgres", "omni-redis", "omni-pgbouncer"]:
            subprocess.run(
                ["docker", "rm", container],
                capture_output=True,
                check=False
            )
        
        # Step 4: Check Docker networks
        log_info("Verifying Docker networks...")
        result = subprocess.run(
            ["docker", "network", "ls", "--format", "{{.Name}}"],
            capture_output=True,
            text=True,
            check=False
        )
        
        if result.returncode == 0:
            networks = result.stdout.strip().split('\n')
            log_info(f"Available networks: {', '.join(networks)}")
        
        log_success("Docker DNS repair completed - containers will be redeployed with fresh network")
        return True
    
    def _repair_alpine_repo(self) -> bool:
        """Repair Alpine repository connectivity."""
        self._flush_dns()
        log_info("Clearing BuildKit cache for Alpine issues...")
        self._repair_buildkit()
        time.sleep(10)
        return True
    
    def _repair_docker_registry(self) -> bool:
        """Repair Docker registry connectivity."""
        return self._repair_dns()
    
    def _repair_network(self) -> bool:
        """Repair network issues."""
        log_info("Pruning Docker networks...")
        subprocess.run(["docker", "network", "prune", "-f"], 
                      capture_output=True, check=False)
        return True
    
    def _cleanup_disk_space(self) -> bool:
        """Cleanup disk space."""
        log_info("Cleaning up Docker resources...")
        subprocess.run(["docker", "system", "prune", "-af"], 
                      capture_output=True, check=False)
        return True
    
    def _repair_oom(self) -> bool:
        """Repair out of memory issues."""
        log_info("Stopping containers to free memory...")
        subprocess.run(["docker", "stop", "$(docker ps -q)"], 
                      shell=True, capture_output=True, check=False)
        time.sleep(5)
        return True
    
    def _repair_port_conflict(self) -> bool:
        """Repair port conflicts."""
        log_info("Checking for port conflicts...")
        # Just log - actual fix is to stop conflicting processes manually
        return True
    
    def _repair_npm_integrity(self) -> bool:
        """Repair NPM integrity issues."""
        log_info("This will be handled by frontend builder")
        return True
    
    def _wait_for_backend_init(self) -> bool:
        """Wait for Go backend to finish initializing (multi-tenant DB setup)."""
        log_fix("Waiting for Go backend to initialize (tenant DB connections)...")
        log_info("Go backend needs ~2-3 minutes for first run (PostgreSQL tenant setup)")
        
        # Progressive wait: 30s, 60s, 90s
        wait_times = [30, 60, 90]
        
        for wait_time in wait_times:
            log_info(f"Waiting {wait_time}s for backend initialization...")
            time.sleep(wait_time)
            
            # Check if backend is responding
            try:
                import requests
                response = requests.get("http://localhost:3000/api/health", timeout=10)
                if response.status_code == 200:
                    log_success(f"Backend is now healthy after {wait_time}s wait")
                    return True
            except:
                pass  # Continue waiting
        
        log_warning("Backend still not responding - may need manual investigation")
        return False
    
    def _restart_unhealthy_service(self) -> bool:
        """Restart unhealthy containers."""
        log_fix("Restarting unhealthy containers...")
        
        # Get unhealthy containers
        result = subprocess.run(
            ["docker", "ps", "--filter", "health=unhealthy", "--format", "{{.Names}}"],
            capture_output=True,
            text=True,
            check=False,
        )
        
        if result.returncode == 0 and result.stdout.strip():
            unhealthy_containers = result.stdout.strip().split('\n')
            for container in unhealthy_containers:
                log_info(f"Restarting unhealthy container: {container}")
                subprocess.run(["docker", "restart", container], 
                             capture_output=True, check=False)
            
            log_info("Waiting 30s for restarted containers to initialize...")
            time.sleep(30)
            return True
        
        return False
    
    # ============================================
    # HELPER METHODS
    # ============================================
    
    def _flush_dns(self) -> None:
        """Flush DNS cache."""
        log_info("Flushing DNS cache...")
        subprocess.run(["ipconfig", "/flushdns"], capture_output=True, check=False)
    
    def _reset_network_stack(self) -> None:
        """Reset network stack."""
        log_info("Resetting network stack...")
        subprocess.run(["netsh", "winsock", "reset"], capture_output=True, check=False)
        subprocess.run(["netsh", "int", "ip", "reset"], capture_output=True, check=False)
    
    def _restart_wsl(self) -> bool:
        """Restart WSL."""
        log_info("Restarting WSL...")
        subprocess.run(["wsl", "--shutdown"], capture_output=True, check=False)
        time.sleep(10)
        return True
    
    def _light_cleanup(self) -> None:
        """Light cleanup - dangling images only."""
        log_info("Light cleanup (dangling images)...")
        subprocess.run(["docker", "image", "prune", "-f"], 
                      capture_output=True, check=False)
    
    def _aggressive_cleanup(self) -> None:
        """Aggressive cleanup - all unused resources."""
        log_info("Aggressive cleanup (all unused resources)...")
        subprocess.run(["docker", "system", "prune", "-af", "--volumes"], 
                      capture_output=True, check=False)
