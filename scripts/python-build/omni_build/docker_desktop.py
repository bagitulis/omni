"""
Docker Desktop Management module for Omni Build System.

SRP: This module ONLY handles Docker Desktop process management on Windows.
"""
import subprocess
import time
from pathlib import Path
from typing import Optional

from omni_build.logger import log_error, log_info, log_success, log_warning


class DockerDesktopManager:
    """Manages Docker Desktop process on Windows."""
    
    def is_running(self) -> bool:
        """
        Check if Docker Desktop process is running.
        
        Returns:
            True if Docker Desktop.exe process is found
        """
        try:
            result = subprocess.run(
                ["tasklist", "/FO", "CSV", "/NH"],
                capture_output=True,
                text=True,
                check=False
            )
            
            if "Docker Desktop.exe" in result.stdout:
                log_info("Docker Desktop process is running")
                return True
            
            log_info("Docker Desktop process not found")
            return False
            
        except Exception as e:
            log_warning(f"Could not check Docker Desktop process: {e}")
            return False
    
    def start(self) -> bool:
        """
        Start Docker Desktop application.
        
        Returns:
            True if started successfully
        """
        docker_paths = [
            Path("C:\\Program Files\\Docker\\Docker\\Docker Desktop.exe"),
            Path("C:\\Program Files (x86)\\Docker\\Docker\\Docker Desktop.exe"),
            Path.home() / "AppData" / "Local" / "Docker" / "Docker Desktop.exe",
        ]
        
        for docker_path in docker_paths:
            if docker_path.exists():
                log_info(f"Starting Docker Desktop from: {docker_path}")
                try:
                    subprocess.Popen(
                        [str(docker_path)], 
                        stdout=subprocess.DEVNULL, 
                        stderr=subprocess.DEVNULL
                    )
                    
                    max_wait = 30
                    waited = 0
                    while waited < max_wait:
                        time.sleep(2)
                        waited += 2
                        if self.is_running():
                            time.sleep(5)
                            log_success("Docker Desktop process started")
                            return True
                    
                    log_warning(f"Docker Desktop process not detected after {max_wait}s")
                    return True  # Continue anyway
                    
                except Exception as e:
                    log_error(f"Failed to start Docker Desktop: {e}")
                    return False
        
        log_error("Docker Desktop executable not found in standard locations")
        return False
    
    def wait_for_ready(
        self, 
        timeout_seconds: int = 90, 
        activity: str = "Waiting for Docker..."
    ) -> bool:
        """
        Wait for Docker to be ready with progressive backoff.
        
        Args:
            timeout_seconds: Maximum time to wait
            activity: Activity description for logging
            
        Returns:
            True if Docker became ready, False if timeout
        """
        log_info(f"{activity} (timeout: {timeout_seconds}s)")
        
        elapsed = 0
        interval = 3
        consecutive_ready = 0
        
        while elapsed < timeout_seconds:
            time.sleep(interval)
            elapsed += interval
            
            if elapsed > 30 and interval < 8:
                interval = 8
                log_info("Increasing check interval for efficiency...")
            
            try:
                result = subprocess.run(
                    ["docker", "info"],
                    capture_output=True,
                    text=True,
                    timeout=10,
                )
                
                if result.returncode == 0:
                    consecutive_ready += 1
                    log_info(f"Docker responding ({consecutive_ready}/2 consecutive checks)")
                    
                    if consecutive_ready >= 2:
                        log_success(f"Docker ready after {elapsed}s")
                        return True
                else:
                    consecutive_ready = 0
                    
            except subprocess.TimeoutExpired:
                consecutive_ready = 0
                log_info(f"Docker not responding yet ({elapsed}s elapsed)...")
            except Exception as e:
                consecutive_ready = 0
                log_warning(f"Docker check error: {e}")
        
        log_error(f"Docker did not become ready after {timeout_seconds}s")
        return False
    
    def check_engine_ready(self) -> bool:
        """Quick check if Docker engine is responding."""
        try:
            result = subprocess.run(
                ["docker", "info"],
                capture_output=True,
                text=True,
                timeout=10,
            )
            return result.returncode == 0
        except Exception:
            return False
