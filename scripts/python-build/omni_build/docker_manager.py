"""
Docker Manager module for Omni Build System.

SRP: This module ONLY handles Docker operations (build, deploy, status checks).
No error detection logic (uses ErrorHandler), no frontend builds, no health checks.
"""
import subprocess
import sys
import time
from pathlib import Path
from typing import Optional

from omni_build.config import Config
from omni_build.error_handler import ErrorHandler
from omni_build.logger import (
    create_progress,
    log_error,
    log_info,
    log_success,
    log_warning,
)
from omni_build.models import BuildMode, BuildResult, ContainerStatus, SpecLevel, ErrorPattern
from omni_build.output_parser import OutputParser


class DockerManager:
    """
    Manages Docker operations with retry logic and error recovery.
    
    Responsibilities:
    - Check Docker readiness
    - Build Docker images
    - Deploy containers
    - Stop/cleanup containers
    - Get container status
    """
    
    def __init__(self, config: Config, error_handler: ErrorHandler) -> None:
        """
        Initialize Docker manager.
        
        Args:
            config: Configuration instance
            error_handler: Error handler for auto-fix
        """
        self.config = config
        self.error_handler = error_handler
        self.output_parser = OutputParser(error_handler.error_patterns)
    
    def check_docker_ready(self) -> bool:
        """
        Check if Docker Desktop is running and ready.
        If not ready, AUTO-FIX by restarting Docker.
        
        Returns:
            True if Docker is ready, False otherwise
        """
        log_info("Checking Docker Desktop status...")
        
        try:
            # Simple check like PowerShell: docker info with redirect stderr
            result = subprocess.run(
                ["docker", "info"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            
            if result.returncode == 0:
                log_success("Docker Desktop is ready")
                return True
            
            # Docker not ready - AUTO-FIX
            log_warning("Docker Desktop is not ready - attempting auto-fix...")
            return self._auto_fix_docker()
            
        except subprocess.TimeoutExpired:
            log_warning("Docker command timed out - Docker engine may be hanging")
            log_info("Auto-fixing: Restarting Docker engine...")
            return self._auto_fix_docker()
        except FileNotFoundError:
            log_error("Docker CLI not found - Docker Desktop may not be installed")
            log_error("Please install Docker Desktop: https://www.docker.com/products/docker-desktop")
            return False
    
    def _auto_fix_docker(self) -> bool:
        """
        Auto-fix Docker issues by restarting engine.
        
        Returns:
            True if Docker is ready after fix
        """
        log_info("Applying Docker auto-fix...")
        
        # Use error handler's repair method
        if self.error_handler._repair_docker_engine():
            log_success("Docker engine restarted")
            
            # Wait for Docker to be ready - INCREASED TIMEOUT (match PowerShell: 120s)
            log_info("Waiting for Docker to initialize (max 2 minutes)...")
            for attempt in range(1, 13):  # Max 120 seconds (12 * 10s)
                time.sleep(10)
                
                try:
                    result = subprocess.run(
                        ["docker", "info"],
                        capture_output=True,
                        timeout=5,
                    )
                    
                    if result.returncode == 0:
                        log_success(f"Docker ready after {attempt * 10}s")
                        return True
                    
                    log_info(f"Docker still initializing (attempt {attempt}/12)...")
                    
                except subprocess.TimeoutExpired:
                    log_info(f"Docker still starting (attempt {attempt}/12)...")
                    continue
            
            log_error("Docker did not become ready after 2 minutes")
            return False
        
        log_error("Failed to restart Docker engine")
        return False
    
    def check_linux_mode(self) -> bool:
        """
        Verify Docker is in Linux containers mode.
        
        Returns:
            True if in Linux mode, False otherwise
        """
        try:
            # Use docker version like PowerShell
            result = subprocess.run(
                ["docker", "version", "--format", "{{.Server.Os}}"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            
            if result.returncode == 0:
                os_type = result.stdout.strip().lower()
                if os_type == "linux":
                    log_success("Docker is in Linux containers mode")
                    return True
                else:
                    log_warning(f"Docker is in {os_type} mode - should be Linux mode")
                    log_warning("Please switch to Linux containers in Docker Desktop")
                    return False
            
            # Could not check - assume OK (like PowerShell)
            return True
            
        except Exception as e:
            log_warning(f"Could not verify Docker mode: {e}")
            return True  # Assume OK if we can't check
    
    def build_images(
        self,
        spec: SpecLevel,
        no_cache: bool = False,
        max_retries: Optional[int] = None,
    ) -> bool:
        """
        Build Docker images with INTELLIGENT retry logic and real-time error detection.
        
        This is the SMART auto-fix version:
        1. Stream output in real-time
        2. Detect errors AS THEY HAPPEN
        3. Apply specific fix for detected error
        4. Retry automatically
        5. Loop until success
        
        Args:
            spec: Specification level (lowspec/standard/highspec)
            no_cache: Force rebuild without cache
            max_retries: Maximum retry attempts (uses config default if None)
            
        Returns:
            True if build succeeded, False otherwise
        """
        if max_retries is None:
            max_retries = self.config.max_build_retries
        
        compose_files = self.config.get_compose_files(spec)
        log_info(f"Building images (spec: {spec.value}, no-cache: {no_cache})...")
        
        # Build docker-compose command
        cmd = ["docker-compose"]
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.append("build")
        
        if no_cache:
            cmd.append("--no-cache")
        
        # INTELLIGENT RETRY LOOP with error-specific fixes
        for attempt in range(1, max_retries + 1):
            print(f"\n{'='*60}")
            print(f"BUILD ATTEMPT {attempt}/{max_retries}")
            print(f"{'='*60}\n")
            sys.stdout.flush()
            
            log_info(f"Starting build attempt {attempt}/{max_retries}...")
            
            self.output_parser.clear_buffer()
            
            process: Optional[subprocess.Popen] = None
            start_time = time.time()
            last_output_time = time.time()
            line_count = 0
            
            try:
                # Run with REAL-TIME output streaming
                process = subprocess.Popen(
                    cmd,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.STDOUT,
                    text=True,
                    bufsize=1,  # Line buffered
                    universal_newlines=True,
                    cwd=str(self.config.project_root),
                )
                
                # Stream output and detect errors in REAL-TIME
                error_detected: Optional[ErrorPattern] = None
                
                print(f"[{time.strftime('%H:%M:%S')}] Build started...\n")
                sys.stdout.flush()
                
                if process.stdout:
                    for line in process.stdout:
                        if line:
                            line_count += 1
                            elapsed = int(time.time() - start_time)
                            
                            # Print line with timestamp every 10 lines
                            if line_count % 10 == 0:
                                print(f"[{elapsed}s] {line.rstrip()}")
                            else:
                                print(line.rstrip())
                            
                            sys.stdout.flush()  # FORCE FLUSH to show immediately
                            
                            # Parse line for errors
                            detected = self.output_parser.parse_line(line)
                            if detected and not error_detected:
                                error_detected = detected
                                print(f"\n⚠️  ERROR DETECTED: {detected.description}\n")
                                sys.stdout.flush()
                            
                            last_output_time = time.time()
                
                process.wait(timeout=self.config.docker_build_timeout)
                
                elapsed = int(time.time() - start_time)
                print(f"\n[{elapsed}s] Build process completed with exit code: {process.returncode}\n")
                sys.stdout.flush()
                
                # Check exit code
                if process.returncode == 0:
                    log_success(f"Build completed successfully (attempt {attempt}, took {elapsed}s)")
                    return True
                
                # Build failed
                print(f"\n{'!'*60}")
                print(f"BUILD FAILED - Exit code: {process.returncode}")
                print(f"{'!'*60}\n")
                sys.stdout.flush()
                
                log_error(f"Build failed (attempt {attempt}, exit code: {process.returncode})")
                
                # Get most critical error detected during streaming
                if not error_detected:
                    error_detected = self.output_parser.get_most_critical_error()
                
                if error_detected:
                    print(f"\n🔍 Critical error identified: {error_detected.description}")
                    print(f"   Severity: {error_detected.severity}")
                    print(f"   Pattern: {error_detected.name}\n")
                    sys.stdout.flush()
                    
                    log_warning(f"Critical error: {error_detected.description}")
                    
                    # Check if code error (cannot auto-fix)
                    if error_detected.is_code_error:
                        print(f"\n❌ CODE ERROR - Cannot auto-fix!")
                        print(f"   Please fix the code errors manually and retry.\n")
                        sys.stdout.flush()
                        
                        log_error("CODE ERROR DETECTED - Cannot auto-fix")
                        log_error("Please fix the code errors and retry manually")
                        log_error("\nLast 20 lines of output:")
                        print(self.output_parser.get_buffered_output(20))
                        return False
                    
                    # Apply error-specific fix
                    if attempt < max_retries:
                        print(f"\n🔧 Applying fix for: {error_detected.name}")
                        print(f"   Fix function: {error_detected.fix_function}\n")
                        sys.stdout.flush()
                        
                        log_info(f"Applying fix for: {error_detected.name}")
                        if self.error_handler.apply_fix(error_detected):
                            print(f"✅ Fix applied successfully, retrying...\n")
                            sys.stdout.flush()
                            log_success("Fix applied, retrying build...")
                            continue
                        else:
                            print(f"⚠️  Fix failed or not applicable, trying progressive fix...\n")
                            sys.stdout.flush()
                            log_warning("Fix failed or not applicable, trying progressive fix...")
                
                # No specific error detected or fix failed - use progressive fix
                if attempt < max_retries:
                    fix_level = min(attempt, self.config.max_fix_levels)
                    
                    print(f"\n🔄 No specific error fix - using progressive fix level {fix_level}")
                    print(f"   This will apply {fix_level} escalation steps...\n")
                    sys.stdout.flush()
                    
                    log_info(f"Applying progressive fix level {fix_level}...")
                    self.error_handler.apply_progressive_fix(fix_level)
                else:
                    # Max retries reached
                    print(f"\n❌ MAX RETRIES REACHED ({max_retries} attempts)")
                    print(f"   Build failed after all retry attempts.\n")
                    sys.stdout.flush()
                    
                    log_error("Max retries reached - build failed")
                    log_error("\nLast 50 lines of output:")
                    print(self.output_parser.get_buffered_output(50))
                    return False
                
            except subprocess.TimeoutExpired:
                elapsed = int(time.time() - start_time)
                print(f"\n⏱️  BUILD TIMEOUT after {elapsed}s (max: {self.config.docker_build_timeout}s)\n")
                sys.stdout.flush()
                
                log_error(f"Build timed out after {self.config.docker_build_timeout}s")
                if process:
                    process.kill()
                
                if attempt < max_retries:
                    print(f"🔄 Retrying with cleanup (attempt {attempt+1}/{max_retries})...\n")
                    sys.stdout.flush()
                    log_info("Retrying with cleanup...")
                    self.error_handler.apply_progressive_fix(2)
                else:
                    return False
            
            except Exception as e:
                print(f"\n❌ Unexpected error: {e}\n")
                sys.stdout.flush()
                log_error(f"Build error: {e}")
                if attempt >= max_retries:
                    return False
        
        return False
    
    def deploy_containers(
        self,
        spec: SpecLevel,
        max_retries: Optional[int] = None,
    ) -> bool:
        """
        Deploy containers with docker-compose up.
        
        Args:
            spec: Specification level
            max_retries: Maximum retry attempts (uses config default if None)
            
        Returns:
            True if deployment succeeded, False otherwise
        """
        if max_retries is None:
            max_retries = self.config.max_deploy_retries
        
        compose_files = self.config.get_compose_files(spec)
        log_info(f"Deploying containers (spec: {spec.value})...")
        
        # Build docker-compose command
        cmd = ["docker-compose"]
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.extend(["up", "-d", "--build", "--remove-orphans"])
        
        # Retry loop with progressive fixes
        for attempt in range(1, max_retries + 1):
            log_info(f"Deploy attempt {attempt}/{max_retries}...")
            
            try:
                result = subprocess.run(
                    cmd,
                    capture_output=True,
                    text=True,
                    timeout=self.config.docker_deploy_timeout,
                    cwd=str(self.config.project_root),
                )
                
                if result.returncode == 0:
                    log_success(f"Deployment completed (attempt {attempt})")
                    return True
                
                # Deploy failed - analyze error
                error_output = result.stderr + "\n" + result.stdout
                log_error(f"Deploy failed (attempt {attempt}): {result.returncode}")
                
                # Detect error pattern
                error_pattern = self.error_handler.detect_error(error_output)
                
                if error_pattern:
                    log_warning(f"Error detected: {error_pattern.description}")
                    
                    # Try to fix
                    if attempt < max_retries:
                        if self.error_handler.apply_fix(error_pattern):
                            log_info("Fix applied, retrying...")
                            continue
                
                # Apply progressive fix
                if attempt < max_retries:
                    fix_level = min(attempt, self.config.max_fix_levels)
                    log_info(f"Applying progressive fix level {fix_level}...")
                    self.error_handler.apply_progressive_fix(fix_level)
                else:
                    log_error("Max retries reached - deployment failed")
                    log_error(error_output[:1000])
                    return False
                
            except subprocess.TimeoutExpired:
                log_error(f"Deploy timed out after {self.config.docker_deploy_timeout}s")
                if attempt < max_retries:
                    log_info("Retrying with cleanup...")
                    self.error_handler.apply_progressive_fix(3)
                else:
                    return False
            
            except Exception as e:
                log_error(f"Deploy error: {e}")
                if attempt >= max_retries:
                    return False
        
        return False
    
    def stop_containers(self, spec: SpecLevel) -> bool:
        """
        Stop all containers for given spec.
        
        Args:
            spec: Specification level
            
        Returns:
            True if stopped successfully
        """
        compose_files = self.config.get_compose_files(spec)
        log_info("Stopping containers...")
        
        cmd = ["docker-compose"]
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.extend(["down", "--remove-orphans"])
        
        try:
            result = subprocess.run(
                cmd,
                capture_output=True,
                text=True,
                timeout=60,
                cwd=str(self.config.project_root),
            )
            
            if result.returncode == 0:
                log_success("Containers stopped")
                return True
            else:
                log_warning(f"Stop completed with warnings: {result.stderr}")
                return True  # Still return True as containers are likely stopped
                
        except Exception as e:
            log_error(f"Error stopping containers: {e}")
            return False
    
    def get_container_status(self) -> list[ContainerStatus]:
        """
        Get status of all Omni containers.
        
        Returns:
            List of ContainerStatus instances
        """
        try:
            result = subprocess.run(
                [
                    "docker",
                    "ps",
                    "-a",
                    "--filter", "name=omni-",
                    "--format", "{{.Names}}\t{{.Status}}\t{{.State}}\t{{.CreatedAt}}",
                ],
                capture_output=True,
                text=True,
                timeout=10,
            )
            
            if result.returncode != 0:
                return []
            
            containers = []
            for line in result.stdout.strip().split("\n"):
                if not line:
                    continue
                
                parts = line.split("\t")
                if len(parts) >= 4:
                    containers.append(ContainerStatus(
                        name=parts[0],
                        status=parts[1],
                        health=parts[2] if "healthy" in parts[1].lower() else None,
                        created=parts[3],
                    ))
            
            return containers
            
        except Exception as e:
            log_warning(f"Could not get container status: {e}")
            return []
