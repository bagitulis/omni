"""
Docker Build Operations module for Omni Build System.

SRP: This module ONLY handles Docker image build operations.
"""
import subprocess
import sys
import time
import io
import os

# Fix Windows console encoding for Docker build output (contains Unicode ✓/✅)
if sys.platform == 'win32' and not os.environ.get('PYTHONIOENCODING'):
    try:
        sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')
        sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8', errors='replace')
    except (AttributeError, TypeError):
        pass  # Already wrapped or no buffer
from typing import Optional

from omni_build.config import Config
from omni_build.error_handler import ErrorHandler
from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.models import ErrorPattern, SpecLevel
from omni_build.output_parser import OutputParser


class DockerBuilder:
    """Handles Docker image build operations with retry logic."""
    
    def __init__(
        self, 
        config: Config, 
        error_handler: ErrorHandler,
        output_parser: OutputParser
    ) -> None:
        self.config = config
        self.error_handler = error_handler
        self.output_parser = output_parser
    
    def build_images(
        self,
        spec: SpecLevel,
        no_cache: bool = False,
        max_retries: Optional[int] = None,
    ) -> bool:
        """
        Build Docker images with INTELLIGENT retry logic and real-time error detection.
        
        Args:
            spec: Specification level (lowspec/standard/highspec)
            no_cache: Force rebuild without cache
            max_retries: Maximum retry attempts
            
        Returns:
            True if build succeeded, False otherwise
        """
        if max_retries is None:
            max_retries = self.config.max_build_retries
        
        compose_files = self.config.get_compose_files(spec)
        log_info(f"Building images (spec: {spec.value}, no-cache: {no_cache})...")
        
        from omni_build.subprocess_utils import get_compose_command
        cmd = get_compose_command()
        for file in compose_files:
            cmd.extend(["-f", file])
        cmd.append("build")
        
        # Add cache-busting build arg to force rebuild when source changes
        # This ensures Docker detects Go source changes on Windows+WSL2
        cmd.extend(["--build-arg", f"CACHEBUST={int(time.time())}"])
        
        if no_cache:
            cmd.append("--no-cache")
        for attempt in range(1, max_retries + 1):
            print(f"\n{'='*60}")
            print(f"BUILD ATTEMPT {attempt}/{max_retries}")
            print(f"{'='*60}\n")
            sys.stdout.flush()
            
            log_info(f"Starting build attempt {attempt}/{max_retries}...")
            
            self.output_parser.clear_buffer()
            
            process: Optional[subprocess.Popen] = None
            start_time = time.time()
            line_count = 0
            
            try:
                process = subprocess.Popen(
                    cmd,
                    stdout=subprocess.PIPE,
                    stderr=subprocess.STDOUT,
                    text=True,
                    bufsize=1,
                    encoding='utf-8',
                    errors='replace',
                    cwd=str(self.config.project_root),
                )
                
                error_detected: Optional[ErrorPattern] = None
                
                print(f"[{time.strftime('%H:%M:%S')}] Build started...\n")
                sys.stdout.flush()
                
                if process.stdout:
                    for line in process.stdout:
                        if line:
                            line_count += 1
                            elapsed = int(time.time() - start_time)
                            
                            if line_count % 10 == 0:
                                print(f"[{elapsed}s] {line.rstrip()}")
                            else:
                                print(line.rstrip())
                            
                            sys.stdout.flush()
                            
                            detected = self.output_parser.parse_line(line)
                            if detected and not error_detected:
                                error_detected = detected
                                print(f"\n  ERROR DETECTED: {detected.description}\n")
                                sys.stdout.flush()
                
                process.wait(timeout=self.config.docker_build_timeout)
                
                elapsed = int(time.time() - start_time)
                print(f"\n[{elapsed}s] Build completed with exit code: {process.returncode}\n")
                sys.stdout.flush()
                
                if process.returncode == 0:
                    log_success(f"Build completed (attempt {attempt}, took {elapsed}s)")
                    return True
                
                # Build failed - handle error
                success = self._handle_build_failure(
                    attempt, max_retries, error_detected, process.returncode
                )
                if not success and attempt >= max_retries:
                    return False
                    
            except subprocess.TimeoutExpired:
                elapsed = int(time.time() - start_time)
                log_error(f"Build timed out after {self.config.docker_build_timeout}s")
                if process:
                    process.kill()
                
                if attempt < max_retries:
                    log_info("Retrying with cleanup...")
                    self.error_handler.apply_progressive_fix(2)
                else:
                    return False
            
            except Exception as e:
                log_error(f"Build error: {e}")
                if attempt >= max_retries:
                    return False
        
        return False
    
    def _handle_build_failure(
        self,
        attempt: int,
        max_retries: int,
        error_detected: Optional[ErrorPattern],
        return_code: int,
    ) -> bool:
        """Handle build failure and apply fixes."""
        print(f"\n{'!'*60}")
        print(f"BUILD FAILED - Exit code: {return_code}")
        print(f"{'!'*60}\n")
        sys.stdout.flush()
        
        log_error(f"Build failed (attempt {attempt}, exit code: {return_code})")
        
        if not error_detected:
            error_detected = self.output_parser.get_most_critical_error()
        
        if error_detected:
            print(f"\n  Critical error: {error_detected.description}")
            print(f"   Severity: {error_detected.severity}")
            sys.stdout.flush()
            
            if error_detected.is_code_error:
                print(f"\n  CODE ERROR - Cannot auto-fix!")
                log_error("CODE ERROR DETECTED - Please fix manually")
                print(self.output_parser.get_buffered_output(20))
                return False
            
            if attempt < max_retries:
                log_info(f"Applying fix for: {error_detected.name}")
                if self.error_handler.apply_fix(error_detected):
                    log_success("Fix applied, retrying...")
                    return True
                else:
                    log_warning("Fix failed, trying progressive fix...")
        
        # No specific error or fix failed
        if attempt < max_retries:
            fix_level = min(attempt, self.config.max_fix_levels)
            log_info(f"Applying progressive fix level {fix_level}...")
            self.error_handler.apply_progressive_fix(fix_level)
            return True
        else:
            log_error(f"MAX RETRIES REACHED ({max_retries} attempts)")
            print(self.output_parser.get_buffered_output(50))
            return False
