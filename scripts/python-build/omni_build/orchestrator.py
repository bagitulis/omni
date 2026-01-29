"""
Build Orchestrator for Omni Build System.

SRP: This module ONLY handles build orchestration logic.
Helper methods are in orchestrator_helpers.py.
"""
import sys
import time

from omni_build.config import Config
from omni_build.database_restorer import DatabaseRestorer
from omni_build.docker_manager import DockerManager
from omni_build.error_handler import ErrorHandler
from omni_build.frontend_builder import FrontendBuilder
from omni_build.health_checker import HealthChecker
from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.models import BuildMode, BuildResult, SpecLevel
from omni_build.orchestrator_helpers import ErrorTracker, OrchestratorHelpers


class BuildOrchestrator:
    """Orchestrates build process across all components."""
    
    def __init__(self, config: Config) -> None:
        self.config = config
        self.error_handler = ErrorHandler()
        self.docker_manager = DockerManager(config, self.error_handler)
        self.frontend_builder = FrontendBuilder(config)
        self.health_checker = HealthChecker(config)
        self.database_restorer = DatabaseRestorer(config)
        self.error_tracker = ErrorTracker(max_same_error=5)
        
        self._helpers = OrchestratorHelpers(
            config, self.docker_manager, self.health_checker, self.error_handler
        )
    
    def execute_build(
        self,
        mode: BuildMode,
        spec: SpecLevel,
        skip_frontend: bool = False,
        restore_db: bool = False,
    ) -> BuildResult:
        """Execute full build process."""
        start_time = time.time()
        self.error_tracker.reset()
        
        log_info(f"Starting build: mode={mode.value}, spec={spec.value}")
        
        if mode == BuildMode.VALIDATE:
            return self._execute_validate(spec, start_time)
        elif mode == BuildMode.CLEAN:
            return self._execute_clean(spec, start_time)
        elif mode == BuildMode.QUICKFIX:
            return self._execute_quickfix(spec, start_time)
        elif mode == BuildMode.QUICK:
            return self._execute_quick(spec, start_time)
        else:
            return self._execute_smart_full(mode, spec, skip_frontend, restore_db, start_time)
    
    def _execute_validate(self, spec: SpecLevel, start_time: float) -> BuildResult:
        """Execute validation mode."""
        errors, warnings = [], []
        
        if not self._helpers.safe_docker_check():
            errors.append("Docker is not ready")
        if not self.docker_manager.check_linux_mode():
            warnings.append("Docker not in Linux mode")
        
        return BuildResult(
            success=len(errors) == 0,
            mode=BuildMode.VALIDATE,
            spec=spec,
            duration_seconds=time.time() - start_time,
            errors=errors,
            warnings=warnings,
        )
    
    def _execute_clean(self, spec: SpecLevel, start_time: float) -> BuildResult:
        """Execute clean mode."""
        log_info("Clean mode - stopping containers...")
        
        errors = []
        if not self.docker_manager.stop_containers(spec):
            errors.append("Failed to stop containers")
        
        self.error_handler._light_cleanup()
        log_success("Cleanup completed")
        
        return BuildResult(
            success=len(errors) == 0,
            mode=BuildMode.CLEAN,
            spec=spec,
            duration_seconds=time.time() - start_time,
            errors=errors,
        )
    
    def _execute_quickfix(self, spec: SpecLevel, start_time: float) -> BuildResult:
        """Execute Quick Fix mode with smart diagnosis."""
        log_info("Quick Fix mode - smart diagnosis...")
        
        print(f"\n{'='*60}")
        print(f"QUICK FIX - Smart Diagnosis & Fix")
        print(f"{'='*60}\n")
        sys.stdout.flush()
        
        problematic = self._helpers.diagnose_services()
        
        if not problematic:
            if self._helpers.verify_all_services_truly_healthy():
                log_success("All services verified healthy!")
                return BuildResult(
                    success=True, mode=BuildMode.QUICKFIX, spec=spec,
                    duration_seconds=time.time() - start_time,
                )
            else:
                log_warning("Services appear OK but verification failed")
                problematic = ["backend"]
        
        log_warning(f"Problematic: {', '.join(problematic)}")
        
        max_attempts = 5
        for attempt in range(1, max_attempts + 1):
            print(f"\n--- Attempt {attempt}/{max_attempts} ---\n")
            sys.stdout.flush()
            
            error_logs = self._helpers.collect_service_logs(problematic)
            error_pattern = self._helpers.detect_service_errors(problematic)
            
            error_name = error_pattern.name if error_pattern else "UnknownError"
            
            if not self.error_tracker.record_error(error_name):
                repeated = self.error_tracker.get_repeated_error()
                log_error(f"\n{'!'*60}")
                log_error(f"SAME ERROR REPEATED {self.error_tracker.max_same_error}x: {repeated}")
                log_error(f"This likely requires MANUAL intervention!")
                log_error(f"{'!'*60}\n")
                
                return BuildResult(
                    success=False, mode=BuildMode.QUICKFIX, spec=spec,
                    duration_seconds=time.time() - start_time,
                    attempts=attempt,
                    errors=[f"Error '{repeated}' repeated {self.error_tracker.max_same_error}x - requires manual fix"],
                    warnings=problematic,
                )
            
            if error_pattern:
                log_info(f"Detected: {error_pattern.name}")
                self.error_handler.apply_fix(error_pattern)
            else:
                symptoms = self.error_handler.analyze_symptoms(error_logs)
                if symptoms:
                    log_info(f"Symptoms: {list(symptoms.keys())}")
                    self.error_handler.apply_smart_fix(symptoms)
                else:
                    for service in problematic:
                        self._helpers.restart_service(service)
            
            time.sleep(5)
            
            for service in problematic:
                self._helpers.restart_service(service)
            
            time.sleep(10)
            
            remaining = self._helpers.diagnose_services()
            if not remaining:
                if self._helpers.verify_all_services_truly_healthy():
                    log_success(f"Quick Fix succeeded after {attempt} attempt(s)!")
                    return BuildResult(
                        success=True, mode=BuildMode.QUICKFIX, spec=spec,
                        duration_seconds=time.time() - start_time,
                        attempts=attempt,
                    )
                else:
                    log_warning("Services appear OK but health verification failed")
            
            problematic = remaining or problematic
        
        return BuildResult(
            success=False, mode=BuildMode.QUICKFIX, spec=spec,
            duration_seconds=time.time() - start_time,
            attempts=max_attempts,
            errors=["Quick Fix failed - consider Smart Build"],
            warnings=problematic,
        )
    
    def _execute_quick(self, spec: SpecLevel, start_time: float) -> BuildResult:
        """Execute quick restart mode."""
        log_info("Quick mode - restarting containers...")
        
        for attempt in range(1, 4):
            log_info(f"Attempt {attempt}/3")
            
            if attempt == 1 and self.docker_manager.restart_containers(spec):
                time.sleep(3)
                if self._helpers.verify_all_services_truly_healthy():
                    return BuildResult(
                        success=True, mode=BuildMode.QUICK, spec=spec,
                        duration_seconds=time.time() - start_time,
                    )
            
            self.docker_manager.stop_containers(spec)
            if self.docker_manager.deploy_containers(spec):
                if self._helpers.verify_all_services_truly_healthy():
                    return BuildResult(
                        success=True, mode=BuildMode.QUICK, spec=spec,
                        duration_seconds=time.time() - start_time,
                    )
            
            if attempt == 2:
                self.error_handler._light_cleanup()
                time.sleep(5)
        
        return BuildResult(
            success=False, mode=BuildMode.QUICK, spec=spec,
            duration_seconds=time.time() - start_time,
            errors=["Quick restart failed after 3 attempts"],
            warnings=["Consider 'smart' mode"],
        )
    
    def _execute_smart_full(
        self,
        mode: BuildMode,
        spec: SpecLevel,
        skip_frontend: bool,
        restore_db: bool,
        start_time: float,
    ) -> BuildResult:
        """Execute smart or full build mode."""
        errors, warnings = [], []
        
        if not self._helpers.safe_docker_check():
            return BuildResult(
                success=False, mode=mode, spec=spec,
                duration_seconds=time.time() - start_time,
                errors=["Docker is not ready"],
            )
        
        if not self.docker_manager.check_linux_mode():
            warnings.append("Docker not in Linux mode")
        
        if not skip_frontend:
            print(f"\n{'='*60}\nSTEP 1: FRONTEND BUILD\n{'='*60}\n")
            sys.stdout.flush()
            
            if not self.frontend_builder.build(force_install=(mode == BuildMode.FULL)):
                return BuildResult(
                    success=False, mode=mode, spec=spec,
                    duration_seconds=time.time() - start_time,
                    errors=["Frontend build failed"],
                )
            log_success("Frontend build completed")
        
        print(f"\n{'='*60}\nSTEP 2: DOCKER IMAGE BUILD\n{'='*60}\n")
        sys.stdout.flush()
        
        if not self.docker_manager.build_images(spec, no_cache=(mode == BuildMode.FULL)):
            return BuildResult(
                success=False, mode=mode, spec=spec,
                duration_seconds=time.time() - start_time,
                errors=["Docker build failed"],
            )
        
        if not self.docker_manager.deploy_containers(spec):
            return BuildResult(
                success=False, mode=mode, spec=spec,
                duration_seconds=time.time() - start_time,
                errors=["Deployment failed"],
            )
        
        if restore_db:
            print(f"\n{'='*60}\nSTEP 3: DATABASE RESTORE\n{'='*60}\n")
            if self.database_restorer.wait_for_postgres(timeout=120):
                success, msg = self.database_restorer.restore(force=True)
                if not success:
                    warnings.append(f"DB restore: {msg}")
        
        if self._helpers.verify_all_services_truly_healthy():
            log_success("All services verified healthy")
            return BuildResult(
                success=True, mode=mode, spec=spec,
                duration_seconds=time.time() - start_time,
                warnings=warnings,
            )
        
        return BuildResult(
            success=False, mode=mode, spec=spec,
            duration_seconds=time.time() - start_time,
            errors=["Health verification failed"],
            warnings=warnings,
        )
