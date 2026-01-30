"""
Health Checker module for Omni Build System.

SRP: This module ONLY handles health checks for containers and HTTP endpoints.
No Docker operations, no builds.
"""
import subprocess
import time
from typing import Optional

import requests

from omni_build.config import Config
from omni_build.logger import log_error, log_info, log_success, log_warning
from omni_build.models import HealthCheckResult


class HealthChecker:
    """
    Performs health checks on containers and HTTP endpoints.
    
    Responsibilities:
    - Check container health status
    - Check HTTP endpoint health
    - Wait for services to become healthy
    """
    
    def __init__(self, config: Config) -> None:
        """
        Initialize health checker.
        
        Args:
            config: Configuration instance
        """
        self.config = config
    
    def check_container_health(self, container_name: str) -> HealthCheckResult:
        """
        Check health status of a Docker container.
        
        Args:
            container_name: Name of container to check
            
        Returns:
            HealthCheckResult with status
        """
        try:
            # Get container health status
            result = subprocess.run(
                [
                    "docker",
                    "inspect",
                    "--format", "{{.State.Health.Status}}",
                    container_name,
                ],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            
            if result.returncode == 0:
                health_status = result.stdout.strip()
                
                if health_status == "healthy":
                    return HealthCheckResult(
                        service=container_name,
                        status="healthy",
                    )
                elif health_status == "unhealthy":
                    return HealthCheckResult(
                        service=container_name,
                        status="unhealthy",
                        error="Container reports unhealthy status",
                    )
                else:
                    # No health check defined or starting
                    return HealthCheckResult(
                        service=container_name,
                        status="unknown",
                        error=f"Health status: {health_status}",
                    )
            else:
                # Container not found or error
                return HealthCheckResult(
                    service=container_name,
                    status="unhealthy",
                    error=f"Container check failed: {result.stderr}",
                )
        
        except subprocess.TimeoutExpired:
            return HealthCheckResult(
                service=container_name,
                status="unhealthy",
                error="Health check timed out",
            )
        
        except Exception as e:
            return HealthCheckResult(
                service=container_name,
                status="unhealthy",
                error=str(e),
            )
    
    def check_endpoint_health(
        self,
        endpoint_url: str,
        service_name: str,
        timeout: int = 5,
    ) -> HealthCheckResult:
        """
        Check health of an HTTP endpoint.
        
        Args:
            endpoint_url: URL to check
            service_name: Service name for logging
            timeout: Request timeout in seconds
            
        Returns:
            HealthCheckResult with status and response time
        """
        try:
            start_time = time.time()
            
            response = requests.get(
                endpoint_url,
                timeout=timeout,
                allow_redirects=True,
            )
            
            response_time = (time.time() - start_time) * 1000  # Convert to ms
            
            if response.status_code == 200:
                return HealthCheckResult(
                    service=service_name,
                    endpoint=endpoint_url,
                    status="healthy",
                    response_time_ms=response_time,
                )
            else:
                return HealthCheckResult(
                    service=service_name,
                    endpoint=endpoint_url,
                    status="unhealthy",
                    response_time_ms=response_time,
                    error=f"HTTP {response.status_code}",
                )
        
        except requests.Timeout:
            return HealthCheckResult(
                service=service_name,
                endpoint=endpoint_url,
                status="unhealthy",
                error=f"Timeout after {timeout}s",
            )
        
        except requests.ConnectionError as e:
            return HealthCheckResult(
                service=service_name,
                endpoint=endpoint_url,
                status="unhealthy",
                error=f"Connection failed: {str(e)[:100]}",
            )
        
        except Exception as e:
            return HealthCheckResult(
                service=service_name,
                endpoint=endpoint_url,
                status="unhealthy",
                error=str(e)[:100],
            )
    
    def wait_for_healthy(
        self,
        container_name: Optional[str] = None,
        endpoint_url: Optional[str] = None,
        service_name: Optional[str] = None,
        timeout: Optional[int] = None,
    ) -> bool:
        """
        Wait for a service to become healthy.
        
        Args:
            container_name: Container name to check (if checking container)
            endpoint_url: Endpoint URL to check (if checking HTTP)
            service_name: Service name for logging
            timeout: Max wait time in seconds (uses config default if None)
            
        Returns:
            True if service became healthy, False if timeout
        """
        if timeout is None:
            timeout = self.config.health_check_timeout
        
        if service_name is None:
            service_name = container_name or endpoint_url or "service"
        
        log_info(f"Waiting for {service_name} to become healthy (timeout: {timeout}s)...")
        
        start_time = time.time()
        attempt = 0
        
        while (time.time() - start_time) < timeout:
            attempt += 1
            
            # Check health based on what was provided
            if container_name:
                result = self.check_container_health(container_name)
            elif endpoint_url:
                result = self.check_endpoint_health(endpoint_url, service_name)
            else:
                log_error("Must provide either container_name or endpoint_url")
                return False
            
            if result.status == "healthy":
                elapsed = time.time() - start_time
                log_success(f"{service_name} is healthy (took {elapsed:.1f}s, {attempt} attempts)")
                return True
            
            # Wait before retry
            time.sleep(5)
        
        # Timeout reached
        elapsed = time.time() - start_time
        log_error(f"{service_name} did not become healthy after {elapsed:.1f}s ({attempt} attempts)")
        return False
    
    def check_all_services(self) -> dict[str, HealthCheckResult]:
        """
        Check health of all configured services.
        
        Returns:
            Dictionary mapping service name to HealthCheckResult
        """
        results = {}
        
        # Check backend HTTP endpoint
        log_info("Checking backend health...")
        results["backend"] = self.check_endpoint_health(
            self.config.backend_health_url,
            "backend",
        )
        
        # CRITICAL: Also check backend logs for database errors
        if results["backend"].status == "healthy":
            log_info("Backend HTTP OK, verifying database connection...")
            db_check = self._check_backend_database_connection()
            if not db_check:
                log_error("Backend HTTP is OK but DATABASE connection FAILED!")
                results["backend"] = HealthCheckResult(
                    service="backend",
                    endpoint=self.config.backend_health_url,
                    status="unhealthy",
                    error="Backend can't connect to PostgreSQL - DNS or network issue",
                )
        
        # Check nginx
        log_info("Checking nginx health...")
        results["nginx"] = self.check_endpoint_health(
            self.config.nginx_health_url,
            "nginx",
        )
        
        # Check postgres container
        log_info("Checking postgres container...")
        results["postgres"] = self.check_container_health(
            self.config.container_postgres,
        )
        
        # Check redis container
        log_info("Checking redis container...")
        results["redis"] = self.check_container_health(
            self.config.container_redis,
        )
        
        return results
    
    def _check_backend_database_connection(self) -> bool:
        """
        Check backend logs for database connection errors.
        
        Returns:
            True if no database errors found, False otherwise
        """
        import subprocess
        
        try:
            # Get last 50 lines of backend logs
            result = subprocess.run(
                ["docker", "logs", "--tail", "50", self.config.container_backend],
                capture_output=True,
                text=True,
                encoding='utf-8',
                errors='replace',
                timeout=10,
            )
            
            if result.returncode != 0:
                log_warning("Could not read backend logs - cannot verify DB connection")
                return False  # Cannot verify = not OK
            
            logs = result.stdout + result.stderr
            
            # Check for critical database errors
            error_patterns = [
                "lookup postgres on.*no such host",
                "failed to connect.*postgres.*hostname resolving",
                "Error getting tenant list",
                "no such host",
                "connection refused.*postgres",
                "could not connect to server",
            ]
            
            import re
            for pattern in error_patterns:
                if re.search(pattern, logs, re.IGNORECASE):
                    log_error(f"Found database error in backend logs: {pattern}")
                    log_error("Backend is running but CAN'T CONNECT TO DATABASE!")
                    return False
            
            log_success("Backend database connection verified OK")
            return True
            
        except Exception as e:
            log_warning(f"Error checking backend logs: {e}")
            return False  # Cannot verify = not OK
    
    def wait_for_all_services(self, timeout: Optional[int] = None) -> bool:
        """
        Wait for all critical services to become healthy.
        
        Args:
            timeout: Max wait time in seconds per service
            
        Returns:
            True if all services became healthy, False otherwise
        """
        log_info("Waiting for all services to become healthy...")
        
        # Critical services that must be healthy
        services = [
            (self.config.backend_health_url, "backend", None),
            (None, "postgres", self.config.container_postgres),
            (None, "redis", self.config.container_redis),
        ]
        
        all_healthy = True
        
        for endpoint, name, container in services:
            if endpoint:
                healthy = self.wait_for_healthy(
                    endpoint_url=endpoint,
                    service_name=name,
                    timeout=timeout,
                )
            else:
                healthy = self.wait_for_healthy(
                    container_name=container,
                    service_name=name,
                    timeout=timeout,
                )
            
            if not healthy:
                all_healthy = False
                log_error(f"{name} failed to become healthy")
        
        if all_healthy:
            log_success("All services are healthy!")
        else:
            log_warning("Some services failed health checks")
        
        return all_healthy
