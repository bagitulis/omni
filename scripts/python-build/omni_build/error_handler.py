"""
Error Handler module for Omni Build System.

SRP: This module ONLY handles error detection and fix orchestration.
Pattern definitions are in error_patterns*.py.
Fix implementations are in error_handler_fixes.py and service_fixers*.py.
Fix method registry is in fix_methods_registry.py.
"""
import re
import time
from typing import Optional

from omni_build.error_handler_fixes import (
    DockerInfraFixer,
    NetworkFixer,
    ResourceFixer,
    SystemFixer,
)
from omni_build.error_patterns import get_all_error_patterns
from omni_build.fix_methods_registry import get_fix_methods
from omni_build.logger import log_error, log_fix, log_info, log_success, log_warning
from omni_build.models import ErrorPattern
from omni_build.service_fixers import (
    NginxFixer,
    PostgresFixer,
    RedisFixer,
)


class ErrorHandler:
    """
    Error detection and recovery orchestrator.
    
    Responsibilities:
    - Detect error patterns from logs/output
    - Dispatch fixes to appropriate service fixers
    - Apply progressive fix escalation
    """
    
    def __init__(self) -> None:
        """Initialize error handler with all error patterns."""
        self.error_patterns = get_all_error_patterns()
        self.fix_attempt_count = 0
        self.fix_methods = get_fix_methods()
    
    def detect_error(self, error_message: str) -> Optional[ErrorPattern]:
        """Detect error type from message using pattern matching."""
        for pattern in self.error_patterns:
            if re.search(pattern.pattern, error_message, re.IGNORECASE | re.MULTILINE):
                return pattern
        return None
    
    def can_auto_fix(self, error_message: str) -> bool:
        """Check if error can be automatically fixed."""
        error_pattern = self.detect_error(error_message)
        if error_pattern is None:
            return True
        return not error_pattern.is_code_error
    
    def apply_fix(self, error_pattern: ErrorPattern) -> bool:
        """Apply fix for detected error pattern."""
        if error_pattern.is_code_error:
            log_error(f"Code error detected: {error_pattern.description}")
            log_error("This error requires manual fix.")
            return False
        
        if not error_pattern.fix_function:
            log_warning(f"No fix function defined for {error_pattern.name}")
            return False
        
        log_fix(f"Applying fix for: {error_pattern.description}")
        
        fix_method = self.fix_methods.get(error_pattern.fix_function)
        if fix_method:
            try:
                result = fix_method()
                if result:
                    log_success(f"Fix applied: {error_pattern.name}")
                else:
                    log_warning(f"Fix completed with warnings: {error_pattern.name}")
                return result
            except Exception as e:
                log_error(f"Fix failed: {e}")
                return False
        
        log_warning(f"Fix method not implemented: {error_pattern.fix_function}")
        return False
    
    def apply_smart_fix(self, symptoms: dict) -> bool:
        """Apply SMART fix based on analyzed symptoms."""
        log_fix("Smart Fix - Analyzing symptoms...")
        
        fixes_applied = []
        
        if symptoms.get('docker_not_responding') or symptoms.get('docker_engine_error'):
            log_info("Detected: Docker engine issue -> Restarting Docker")
            DockerInfraFixer.repair_docker_engine()
            fixes_applied.append("docker_engine")
            time.sleep(30)
            return True
        
        if symptoms.get('wsl_error') or symptoms.get('mount_error'):
            log_info("Detected: WSL issue -> Restarting WSL")
            SystemFixer.restart_wsl()
            fixes_applied.append("wsl_restart")
            time.sleep(15)
        
        if symptoms.get('dns_error') or symptoms.get('network_error'):
            log_info("Detected: Network/DNS issue -> Flushing DNS")
            SystemFixer.flush_dns()
            NetworkFixer.repair_network()
            fixes_applied.append("network_dns")
            time.sleep(5)
        
        if symptoms.get('port_conflict') or symptoms.get('container_conflict'):
            log_info("Detected: Port/Container conflict -> Cleaning up")
            ResourceFixer.repair_port_conflict()
            DockerInfraFixer.repair_container_name_conflict()
            fixes_applied.append("container_cleanup")
        
        if symptoms.get('disk_full'):
            log_info("Detected: Disk space issue -> Aggressive cleanup")
            SystemFixer.aggressive_cleanup()
            fixes_applied.append("disk_cleanup")
        
        if symptoms.get('postgres_error'):
            log_info("Detected: PostgreSQL issue -> Repairing")
            PostgresFixer.repair_docker_dns_postgres()
            fixes_applied.append("postgres_repair")
        
        if symptoms.get('redis_error'):
            log_info("Detected: Redis issue -> Restarting")
            RedisFixer.restart_redis()
            fixes_applied.append("redis_restart")
        
        if symptoms.get('nginx_error'):
            log_info("Detected: Nginx issue -> Repairing upstream")
            NginxFixer.repair_nginx_upstream()
            fixes_applied.append("nginx_repair")
        
        if not fixes_applied:
            log_info("No specific symptoms -> Applying light cleanup")
            SystemFixer.light_cleanup()
            SystemFixer.flush_dns()
            fixes_applied.append("light_cleanup")
        
        log_success(f"Smart Fix applied: {', '.join(fixes_applied)}")
        time.sleep(5)
        return True
    
    def analyze_symptoms(self, error_output: str = "", container_status: Optional[dict] = None) -> dict:
        """Analyze error output and container status to determine symptoms."""
        symptoms = {}
        
        error_lower = error_output.lower() if error_output else ""
        
        if "500 internal server" in error_lower or "docker" in error_lower and "error" in error_lower:
            symptoms['docker_engine_error'] = True
        
        if "pipe" in error_lower or "npipe" in error_lower:
            symptoms['docker_not_responding'] = True
        
        if "wsl" in error_lower or "mount" in error_lower:
            symptoms['wsl_error'] = True
        
        if "/run/desktop/mnt" in error_lower:
            symptoms['mount_error'] = True
        
        if "dns" in error_lower or "no such host" in error_lower or "servfail" in error_lower:
            symptoms['dns_error'] = True
        
        if "network" in error_lower or "connection refused" in error_lower or "timeout" in error_lower:
            symptoms['network_error'] = True
        
        if "address already in use" in error_lower or "port" in error_lower and "conflict" in error_lower:
            symptoms['port_conflict'] = True
        
        if "container name" in error_lower and "already in use" in error_lower:
            symptoms['container_conflict'] = True
        
        if "no space left" in error_lower or "disk" in error_lower and "full" in error_lower:
            symptoms['disk_full'] = True
        
        if "postgres" in error_lower and ("error" in error_lower or "failed" in error_lower):
            symptoms['postgres_error'] = True
        
        if "redis" in error_lower and ("error" in error_lower or "refused" in error_lower):
            symptoms['redis_error'] = True
        
        if "nginx" in error_lower and ("upstream" in error_lower or "error" in error_lower):
            symptoms['nginx_error'] = True
        
        if container_status:
            for name, status in container_status.items():
                if status in ['exited', 'dead', 'not_found']:
                    symptoms['containers_down'] = True
                    break
        
        return symptoms
    
    def apply_progressive_fix(self, level: int) -> bool:
        """Apply progressive fix based on escalation level (1-6)."""
        self.fix_attempt_count += 1
        wait_time = min(5 * (2 ** (level - 1)), 60)
        
        log_fix(f"Progressive Fix Level {level} (wait: {wait_time}s)")
        
        try:
            if level == 1:
                log_info("L1: Port conflicts + light cleanup...")
                ResourceFixer.repair_port_conflict()
                DockerInfraFixer.repair_container_name_conflict()
                SystemFixer.light_cleanup()
            elif level == 2:
                log_info("L2: Network flush + DNS repair...")
                SystemFixer.flush_dns()
                NetworkFixer.repair_alpine_repo()
            elif level == 3:
                log_info("L3: Full DNS + network reset...")
                NetworkFixer.repair_dns()
            elif level == 4:
                log_info("L4: WSL restart + network reset...")
                NetworkFixer.reset_network_stack()
                SystemFixer.restart_wsl()
            elif level == 5:
                log_info("L5: Docker engine full restart...")
                DockerInfraFixer.repair_docker_engine()
            elif level == 6:
                log_info("L6: Aggressive cleanup + system reset...")
                SystemFixer.aggressive_cleanup()
                SystemFixer.restart_wsl()
                DockerInfraFixer.repair_docker_engine()
            
            log_info(f"Waiting {wait_time}s for stabilization...")
            time.sleep(wait_time)
            
            log_success(f"Fix level {level} applied")
            return True
            
        except Exception as e:
            log_error(f"Fix level {level} failed: {e}")
            return False
    
    # Legacy method aliases for backward compatibility
    def _repair_docker_engine(self) -> bool:
        return DockerInfraFixer.repair_docker_engine()
    
    def _light_cleanup(self) -> None:
        SystemFixer.light_cleanup()
    
    def _flush_dns(self) -> None:
        SystemFixer.flush_dns()
