"""
CLI module for Omni Build System.

SRP: This module ONLY handles command-line interface and orchestration.
Delegates actual work to specialized modules (DockerManager, FrontendBuilder, etc).
"""
import sys
import time
from pathlib import Path

import click
from rich.console import Console
from rich.table import Table

from omni_build.config import Config
from omni_build.database_backup import DatabaseBackup
from omni_build.database_restorer import DatabaseRestorer
from omni_build.docker_manager import DockerManager
from omni_build.error_handler import ErrorHandler
from omni_build.frontend_builder import FrontendBuilder
from omni_build.health_checker import HealthChecker
from omni_build.logger import (
    confirm,
    log_error,
    log_info,
    log_success,
    log_warning,
)
from omni_build.models import BuildMode, BuildResult, SpecLevel

console = Console()


class BuildOrchestrator:
    """
    Orchestrates build process across all components.
    
    SRP: Coordinate between modules, no actual build logic.
    """
    
    def __init__(self, config: Config) -> None:
        """Initialize orchestrator with all required modules."""
        self.config = config
        self.error_handler = ErrorHandler()
        self.docker_manager = DockerManager(config, self.error_handler)
        self.frontend_builder = FrontendBuilder(config)
        self.health_checker = HealthChecker(config)
        self.database_restorer = DatabaseRestorer(config)
    
    def execute_build(
        self,
        mode: BuildMode,
        spec: SpecLevel,
        skip_frontend: bool = False,
        restore_db: bool = False,
    ) -> BuildResult:
        """
        Execute full build process.
        
        Args:
            mode: Build mode (quick/smart/full/validate/clean)
            spec: Specification level (lowspec/standard/highspec)
            skip_frontend: Skip frontend build
            restore_db: Restore database from backup after containers are up
            
        Returns:
            BuildResult with success status and metrics
        """
        start_time = time.time()
        errors = []
        warnings = []
        
        log_info(f"Starting build: mode={mode.value}, spec={spec.value}")
        
        # ============ VALIDATE MODE ============
        if mode == BuildMode.VALIDATE:
            log_info("Validation mode - checking environment only...")
            
            if not self.docker_manager.check_docker_ready():
                errors.append("Docker is not ready")
            
            if not self.docker_manager.check_linux_mode():
                warnings.append("Docker not in Linux mode")
            
            duration = time.time() - start_time
            
            if errors:
                log_error("Validation failed")
                return BuildResult(
                    success=False,
                    mode=mode,
                    spec=spec,
                    duration_seconds=duration,
                    errors=errors,
                    warnings=warnings,
                )
            
            log_success("Validation passed")
            return BuildResult(
                success=True,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                warnings=warnings,
            )
        
        # ============ CLEAN MODE ============
        if mode == BuildMode.CLEAN:
            log_info("Clean mode - stopping containers and cleaning up...")
            
            if not self.docker_manager.stop_containers(spec):
                errors.append("Failed to stop containers")
            
            log_info("Running Docker system cleanup...")
            self.error_handler._light_cleanup()
            
            duration = time.time() - start_time
            
            log_success("Cleanup completed")
            return BuildResult(
                success=len(errors) == 0,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                errors=errors,
            )
        
        # ============ PRE-FLIGHT CHECKS ============
        log_info("Running pre-flight checks...")
        
        if not self.docker_manager.check_docker_ready():
            errors.append("Docker is not ready")
            duration = time.time() - start_time
            return BuildResult(
                success=False,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                errors=errors,
            )
        
        if not self.docker_manager.check_linux_mode():
            warnings.append("Docker not in Linux mode - may cause issues")
        
        # ============ QUICK MODE ============
        if mode == BuildMode.QUICK:
            log_info("Quick mode - restarting containers only...")
            
            # Try 3 attempts with progressive fixes
            max_quick_attempts = 3
            
            for attempt in range(1, max_quick_attempts + 1):
                log_info(f"Quick restart attempt {attempt}/{max_quick_attempts}")
                
                # Attempt 1: Simple restart
                if attempt == 1:
                    log_info("Trying simple restart...")
                    if self.docker_manager.restart_containers(spec):
                        # Verify containers are actually running
                        time.sleep(3)
                        if self.docker_manager.check_containers_running(spec):
                            # Quick health check (60s timeout)
                            if self.health_checker.wait_for_all_services(timeout=60):
                                log_success("Quick restart succeeded!")
                                duration = time.time() - start_time
                                return BuildResult(
                                    success=True,
                                    mode=mode,
                                    spec=spec,
                                    duration_seconds=duration,
                                )
                
                # Attempt 2-3: Stop and deploy fresh
                log_info("Trying stop + deploy...")
                
                # Stop existing
                if not self.docker_manager.stop_containers(spec):
                    log_warning("Stop containers completed with warnings")
                
                # Deploy fresh
                if not self.docker_manager.deploy_containers(spec):
                    if attempt < max_quick_attempts:
                        log_warning(f"Deploy failed, will retry (attempt {attempt}/{max_quick_attempts})...")
                        
                        # Apply quick fixes between attempts
                        detected_errors = [
                            "Quick restart failed - containers not starting properly"
                        ]
                        
                        if attempt == 2:
                            # Try light fixes on attempt 2
                            log_info("Applying quick diagnostic fixes...")
                            self.error_handler._light_cleanup()
                            time.sleep(5)
                        
                        continue
                    else:
                        errors.append("Failed to deploy containers after 3 quick attempts")
                        break
                
                # Health check (short timeout for quick mode)
                log_info("Running health checks...")
                if self.health_checker.wait_for_all_services(timeout=60):
                    log_success("All services healthy after quick restart")
                    
                    duration = time.time() - start_time
                    log_success(f"Quick restart completed successfully in {duration:.1f}s")
                    
                    return BuildResult(
                        success=True,
                        mode=mode,
                        spec=spec,
                        duration_seconds=duration,
                    )
                else:
                    # Health check failed
                    if attempt < max_quick_attempts:
                        log_warning(f"Health check failed, will retry (attempt {attempt}/{max_quick_attempts})...")
                        time.sleep(5)
                        continue
            
            # All quick attempts failed - offer fallback
            duration = time.time() - start_time
            
            log_error("Quick restart failed after 3 attempts")
            log_error("Health checks did not pass - services not responding properly")
            
            return BuildResult(
                success=False,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                errors=["Quick restart failed - services failed health checks after 3 attempts"],
                warnings=["Consider using 'smart' mode for a fresh build"],
            )
        
        # ============ SMART/FULL MODE ============
        
        # Frontend build
        if not skip_frontend:
            print(f"\n{'='*60}")
            print(f"STEP 1: FRONTEND BUILD")
            print(f"{'='*60}\n")
            sys.stdout.flush()
            
            log_info("Building frontend...")
            
            force_install = (mode == BuildMode.FULL)
            
            if not self.frontend_builder.build(force_install=force_install):
                errors.append("Frontend build failed")
                duration = time.time() - start_time
                return BuildResult(
                    success=False,
                    mode=mode,
                    spec=spec,
                    duration_seconds=duration,
                    errors=errors,
                    warnings=warnings,
                )
            
            log_success("Frontend build completed")
        else:
            print(f"\n⏭️  SKIPPING FRONTEND BUILD (--skip-frontend flag)\n")
            sys.stdout.flush()
            log_info("Skipping frontend build (--skip-frontend)")
        
        # Docker build
        print(f"\n{'='*60}")
        print(f"STEP 2: DOCKER IMAGE BUILD")
        print(f"{'='*60}\n")
        sys.stdout.flush()
        
        log_info("Building Docker images...")
        
        no_cache = (mode == BuildMode.FULL)
        
        if not self.docker_manager.build_images(spec, no_cache=no_cache):
            errors.append("Docker build failed")
            duration = time.time() - start_time
            return BuildResult(
                success=False,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                errors=errors,
                warnings=warnings,
            )
        
        log_success("Docker images built")
        
        # Deploy containers
        log_info("Deploying containers...")
        
        if not self.docker_manager.deploy_containers(spec):
            errors.append("Container deployment failed")
            duration = time.time() - start_time
            return BuildResult(
                success=False,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                errors=errors,
                warnings=warnings,
            )
        
        log_success("Containers deployed")
        
        # ============ DATABASE RESTORE (if requested) ============
        if restore_db:
            print(f"\n{'='*60}")
            print(f"STEP 3: DATABASE RESTORE")
            print(f"{'='*60}\n")
            sys.stdout.flush()
            
            log_info("Restoring database from backup...")
            
            # Wait for PostgreSQL to be ready first
            if not self.database_restorer.wait_for_postgres(timeout=120):
                errors.append("PostgreSQL not ready for restore")
                warnings.append("Database restore skipped - continuing with build")
            else:
                success, message = self.database_restorer.restore(force=True)
                if success:
                    log_success("Database restored successfully!")
                else:
                    warnings.append(f"Database restore warning: {message}")
                    log_warning(f"Database restore issue: {message}")
                    log_warning("Continuing with build - database may need manual restore")
        
        # Health checks - CRITICAL: Must succeed or build fails
        log_info("Running health checks...")
        
        if self.health_checker.wait_for_all_services():
            log_success("All services healthy")
            
            # SUCCESS - All checks passed
            duration = time.time() - start_time
            log_success(f"Build completed successfully in {duration:.1f}s")
            
            return BuildResult(
                success=True,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                warnings=warnings,
            )
        else:
            # FAILURE - Health checks failed
            errors.append("Health checks failed - services not responding properly")
            duration = time.time() - start_time
            
            log_error(f"Build failed: Health checks did not pass after {duration:.1f}s")
            
            return BuildResult(
                success=False,
                mode=mode,
                spec=spec,
                duration_seconds=duration,
                errors=errors,
                warnings=warnings,
            )


# ============================================
# CLI COMMANDS
# ============================================

@click.group()
@click.version_option(version="1.0.0", prog_name="Omni Build System")
def cli():
    """
    Omni Build System - Python Edition
    
    Modern build automation with intelligent error recovery.
    """
    pass


@cli.command()
@click.option(
    "--spec",
    type=click.Choice(["lowspec", "standard", "highspec"], case_sensitive=False),
    default="standard",
    help="Server specification level (default: standard)",
)
@click.option(
    "--skip-frontend",
    is_flag=True,
    help="Skip frontend build",
)
@click.option(
    "--restore",
    is_flag=True,
    help="Restore database from backup after containers are up",
)
def smart(spec: str, skip_frontend: bool, restore: bool):
    """
    Smart build (RECOMMENDED) - Cache dependencies, rebuild code.
    
    This mode:
    - Checks if npm dependencies changed (hash-based)
    - Only reinstalls if package-lock.json changed
    - Rebuilds frontend code
    - Rebuilds Docker images with cache
    - Deploys containers
    - Optionally restores database from backup (--restore)
    - Runs health checks
    
    Typical time: 2-3 minutes
    """
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    spec_level = SpecLevel(spec.lower())
    
    # Check if restore requested and backup exists
    if restore:
        if not orchestrator.database_restorer.has_backup():
            log_error("No backup found! Cannot restore database.")
            log_info("Run .\\backups\\db-tools\\backup-smart.ps1 first to create backup.")
            sys.exit(1)
        
        orchestrator.database_restorer.print_backup_status()
    
    result = orchestrator.execute_build(
        mode=BuildMode.SMART,
        spec=spec_level,
        skip_frontend=skip_frontend,
        restore_db=restore,
    )
    
    _print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option(
    "--spec",
    type=click.Choice(["lowspec", "standard", "highspec"], case_sensitive=False),
    default="standard",
    help="Server specification level (default: standard)",
)
def quick(spec: str):
    """
    Quick restart - Restart containers only (no rebuild).
    
    This mode:
    - Stops existing containers
    - Starts containers again
    - Runs health checks
    
    Typical time: 10-20 seconds
    
    Use this when you only need to restart services without code changes.
    """
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    spec_level = SpecLevel(spec.lower())
    
    result = orchestrator.execute_build(
        mode=BuildMode.QUICK,
        spec=spec_level,
    )
    
    _print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option(
    "--spec",
    type=click.Choice(["lowspec", "standard", "highspec"], case_sensitive=False),
    default="standard",
    help="Server specification level (default: standard)",
)
@click.option(
    "--skip-frontend",
    is_flag=True,
    help="Skip frontend build",
)
@click.option(
    "--restore",
    is_flag=True,
    help="Restore database from backup after containers are up",
)
def full(spec: str, skip_frontend: bool, restore: bool):
    """
    Full rebuild - Clean build from scratch (no cache).
    
    This mode:
    - Reinstalls all npm dependencies
    - Rebuilds frontend from scratch
    - Rebuilds Docker images without cache
    - Deploys containers
    - Optionally restores database from backup (--restore)
    - Runs health checks
    
    Typical time: 5-10 minutes
    
    Use this when you have persistent build issues or need a clean slate.
    Use --restore when moving to a new PC or need fresh database from backup.
    """
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    spec_level = SpecLevel(spec.lower())
    
    # Check if restore requested and backup exists
    if restore:
        if not orchestrator.database_restorer.has_backup():
            log_error("No backup found! Cannot restore database.")
            log_info("Run .\\backups\\db-tools\\backup-smart.ps1 first to create backup.")
            sys.exit(1)
        
        orchestrator.database_restorer.print_backup_status()
        log_warning("Full rebuild with database restore will take 5-10 minutes")
    else:
        log_warning("Full rebuild will take 5-10 minutes and clear all caches")
    
    if not confirm("Continue with full rebuild?"):
        log_info("Cancelled by user")
        sys.exit(0)
    
    result = orchestrator.execute_build(
        mode=BuildMode.FULL,
        spec=spec_level,
        skip_frontend=skip_frontend,
        restore_db=restore,
    )
    
    _print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
def validate():
    """
    Validate environment - Check Docker and system requirements only.
    
    This mode:
    - Checks if Docker Desktop is running
    - Verifies Linux containers mode
    - Validates Docker engine is responsive
    
    Typical time: 5 seconds
    
    Use this to troubleshoot environment issues without building.
    """
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    result = orchestrator.execute_build(
        mode=BuildMode.VALIDATE,
        spec=SpecLevel.STANDARD,  # Spec doesn't matter for validation
    )
    
    _print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option(
    "--spec",
    type=click.Choice(["lowspec", "standard", "highspec"], case_sensitive=False),
    default="standard",
    help="Server specification level (default: standard)",
)
def clean(spec: str):
    """
    Cleanup - Stop containers and clean Docker resources.
    
    This mode:
    - Stops all containers
    - Removes stopped containers
    - Cleans up dangling images
    
    Typical time: 1-2 minutes
    
    Use this to free up disk space or before a full rebuild.
    """
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    spec_level = SpecLevel(spec.lower())
    
    result = orchestrator.execute_build(
        mode=BuildMode.CLEAN,
        spec=spec_level,
    )
    
    _print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
def status():
    """
    Show status of all Omni containers and services.
    """
    config = Config.from_env()
    docker_manager = DockerManager(config, ErrorHandler())
    health_checker = HealthChecker(config)
    
    log_info("Checking container status...")
    
    containers = docker_manager.get_container_status()
    
    if not containers:
        log_warning("No Omni containers found")
        sys.exit(0)
    
    # Create status table
    table = Table(title="Omni Container Status")
    table.add_column("Container", style="cyan")
    table.add_column("Status", style="yellow")
    table.add_column("Health", style="green")
    table.add_column("Created", style="blue")
    
    for container in containers:
        health = container.health or "-"
        table.add_row(
            container.name,
            container.status,
            health,
            container.created or "-",
        )
    
    console.print(table)
    
    # Check endpoint health
    log_info("\nChecking endpoint health...")
    health_results = health_checker.check_all_services()
    
    health_table = Table(title="Service Health Checks")
    health_table.add_column("Service", style="cyan")
    health_table.add_column("Status", style="yellow")
    health_table.add_column("Response Time", style="green")
    health_table.add_column("Error", style="red")
    
    for service, result in health_results.items():
        status_color = "green" if result.status == "healthy" else "red"
        response_time = f"{result.response_time_ms:.0f}ms" if result.response_time_ms else "-"
        error = result.error or "-"
        
        health_table.add_row(
            service,
            f"[{status_color}]{result.status}[/{status_color}]",
            response_time,
            error,
        )
    
    console.print(health_table)


@cli.command()
@click.option(
    "--force",
    is_flag=True,
    help="Skip confirmations (for automation)",
)
@click.option(
    "--keep-extra",
    is_flag=True,
    help="Keep tables not in backup (default: sync mode - deletes extra tables)",
)
def restore(force: bool, keep_extra: bool):
    """
    Restore database from backup (standalone command).
    
    This command:
    - Checks if PostgreSQL container is running
    - Restores database from backups/ directory
    - Recreates schema structure from backup
    - Imports all table data
    
    By default, this syncs database to match backup exactly.
    Use --keep-extra to preserve local-only tables.
    
    Typical time: 1-5 minutes depending on database size.
    """
    config = Config.from_env()
    restorer = DatabaseRestorer(config)
    
    if not restorer.has_backup():
        log_error("No backup found!")
        log_info("Run: python build.py backup")
        sys.exit(1)
    
    restorer.print_backup_status()
    
    if not force:
        if not confirm("Continue with database restore?"):
            log_info("Cancelled by user")
            sys.exit(0)
    
    # Check PostgreSQL
    if not restorer.check_postgres_running():
        log_error("PostgreSQL container is not running!")
        log_info("Start containers first: python build.py quick")
        sys.exit(1)
    
    success, message = restorer.restore(force=True, keep_extra=keep_extra)
    
    if success:
        log_success("Database restore completed!")
        sys.exit(0)
    else:
        log_error(f"Database restore failed: {message}")
        sys.exit(1)


@cli.command()
@click.option(
    "--force",
    is_flag=True,
    help="Force export all tables (ignore change detection)",
)
@click.option(
    "--dry-run",
    is_flag=True,
    help="Preview only, don't actually export",
)
def backup(force: bool, dry_run: bool):
    """
    Backup database (standalone command).
    
    This command:
    - Detects changed tables since last backup
    - Exports only changed tables (incremental)
    - Chunks large tables (>50k rows) for efficiency
    - Saves to backups/ directory
    
    Options:
    - --force: Export all tables regardless of changes
    - --dry-run: Preview what would be exported without doing it
    
    Typical time: 1-10 minutes depending on changes.
    """
    config = Config.from_env()
    backup_handler = DatabaseBackup(config)
    
    success, message = backup_handler.backup(force=force, dry_run=dry_run)
    
    if success:
        log_success(f"Backup completed: {message}")
        sys.exit(0)
    else:
        log_error(f"Backup failed: {message}")
        sys.exit(1)


def _print_result(result: BuildResult) -> None:
    """
    Print build result summary.
    
    Args:
        result: BuildResult to display
    """
    console.print("\n" + "=" * 60)
    
    if result.success:
        console.print("[bold green]BUILD SUCCESSFUL[/bold green]")
    else:
        console.print("[bold red]BUILD FAILED[/bold red]")
    
    console.print(f"Mode: {result.mode}")
    console.print(f"Spec: {result.spec}")
    console.print(f"Duration: {result.duration_seconds:.1f}s")
    
    if result.errors:
        console.print("\n[bold red]Errors:[/bold red]")
        for error in result.errors:
            console.print(f"  - {error}")
    
    if result.warnings:
        console.print("\n[bold yellow]Warnings:[/bold yellow]")
        for warning in result.warnings:
            console.print(f"  - {warning}")
    
    console.print("=" * 60 + "\n")


if __name__ == "__main__":
    cli()
