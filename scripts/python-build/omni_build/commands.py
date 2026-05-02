"""
CLI Commands for Omni Build System.

SRP: This module ONLY defines CLI commands.
Build logic is in orchestrator.py.
"""
import sys

import click
from rich.console import Console
from rich.table import Table

from omni_build.config import Config
from omni_build.database_backup import DatabaseBackup
from omni_build.database_restorer import DatabaseRestorer
from omni_build.docker_manager import DockerManager
from omni_build.error_handler import ErrorHandler
from omni_build.health_checker import HealthChecker
from omni_build.logger import confirm, log_error, log_info, log_success, log_warning
from omni_build.models import BuildMode, BuildResult, SpecLevel
from omni_build.orchestrator import BuildOrchestrator

console = Console()


def print_result(result: BuildResult) -> None:
    """Print build result summary."""
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


@click.group()
@click.version_option(version="1.0.0", prog_name="Omni Build System")
def cli():
    """Omni Build System - Python Edition"""
    pass


@cli.command()
@click.option("--spec", type=click.Choice(["lowspec", "standard", "highspec"]), default="standard")
def quickfix(spec: str):
    """Quick Fix - Fix service issues without rebuild (FIRST OPTION)."""
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    result = orchestrator.execute_build(
        mode=BuildMode.QUICKFIX,
        spec=SpecLevel(spec.lower()),
    )
    
    print_result(result)
    
    if not result.success:
        console.print("\n[yellow]Quick Fix failed. Run Smart Build?[/yellow]")
        try:
            choice = input("Run Smart Build? [Y/n]: ").strip().lower()
            if choice in ('', 'y', 'yes'):
                console.print("\n[cyan]Starting Smart Build...[/cyan]\n")
                smart_result = orchestrator.execute_build(
                    mode=BuildMode.SMART,
                    spec=SpecLevel(spec.lower()),
                )
                print_result(smart_result)
                sys.exit(0 if smart_result.success else 1)
        except (EOFError, KeyboardInterrupt):
            console.print("\nCancelled.")
    
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option("--spec", type=click.Choice(["lowspec", "standard", "highspec"]), default="standard")
@click.option("--skip-frontend", is_flag=True)
@click.option("--restore", is_flag=True)
def smart(spec: str, skip_frontend: bool, restore: bool):
    """Smart build (RECOMMENDED) - Cache dependencies, rebuild code."""
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    if restore:
        if not orchestrator.database_restorer.has_backup():
            log_error("No backup found!")
            sys.exit(1)
        orchestrator.database_restorer.print_backup_status()
    
    result = orchestrator.execute_build(
        mode=BuildMode.SMART,
        spec=SpecLevel(spec.lower()),
        skip_frontend=skip_frontend,
        restore_db=restore,
    )
    
    print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option("--spec", type=click.Choice(["lowspec", "standard", "highspec"]), default="standard")
def quick(spec: str):
    """Quick restart - Restart containers only (DEPRECATED: use quickfix)."""
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    result = orchestrator.execute_build(
        mode=BuildMode.QUICK,
        spec=SpecLevel(spec.lower()),
    )
    
    print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option("--spec", type=click.Choice(["lowspec", "standard", "highspec"]), default="standard")
@click.option("--skip-frontend", is_flag=True)
@click.option("--restore", is_flag=True)
def full(spec: str, skip_frontend: bool, restore: bool):
    """Full rebuild - Clean build from scratch (no cache)."""
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    if restore:
        if not orchestrator.database_restorer.has_backup():
            log_error("No backup found!")
            sys.exit(1)
        orchestrator.database_restorer.print_backup_status()
        log_warning("Full rebuild with database restore will take 5-10 minutes")
    else:
        log_warning("Full rebuild will take 5-10 minutes")
    
    if not confirm("Continue with full rebuild?"):
        log_info("Cancelled")
        sys.exit(0)
    
    result = orchestrator.execute_build(
        mode=BuildMode.FULL,
        spec=SpecLevel(spec.lower()),
        skip_frontend=skip_frontend,
        restore_db=restore,
    )
    
    print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
def validate():
    """Validate environment - Check Docker and system requirements."""
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    result = orchestrator.execute_build(
        mode=BuildMode.VALIDATE,
        spec=SpecLevel.STANDARD,
    )
    
    print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
@click.option("--spec", type=click.Choice(["lowspec", "standard", "highspec"]), default="standard")
def clean(spec: str):
    """Cleanup - Stop containers and clean Docker resources."""
    config = Config.from_env()
    orchestrator = BuildOrchestrator(config)
    
    result = orchestrator.execute_build(
        mode=BuildMode.CLEAN,
        spec=SpecLevel(spec.lower()),
    )
    
    print_result(result)
    sys.exit(0 if result.success else 1)


@cli.command()
def status():
    """Show status of all Omni containers and services."""
    config = Config.from_env()
    docker_manager = DockerManager(config, ErrorHandler())
    health_checker = HealthChecker(config)
    
    log_info("Checking container status...")
    
    containers = docker_manager.get_container_status()
    
    if not containers:
        log_warning("No Omni containers found")
        sys.exit(0)
    
    table = Table(title="Omni Container Status")
    table.add_column("Container", style="cyan")
    table.add_column("Status", style="yellow")
    table.add_column("Health", style="green")
    table.add_column("Created", style="blue")
    
    for container in containers:
        table.add_row(
            container.name,
            container.status,
            container.health or "-",
            container.created or "-",
        )
    
    console.print(table)
    
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
        
        health_table.add_row(
            service,
            f"[{status_color}]{result.status}[/{status_color}]",
            response_time,
            result.error or "-",
        )
    
    console.print(health_table)


@cli.command()
@click.option("--force", is_flag=True)
@click.option("--keep-extra", is_flag=True)
def restore(force: bool, keep_extra: bool):
    """Restore database from backup."""
    config = Config.from_env()
    restorer = DatabaseRestorer(config)
    
    if not restorer.has_backup():
        log_error("No backup found!")
        sys.exit(1)
    
    restorer.print_backup_status()
    
    if not force:
        if not confirm("Continue with database restore?"):
            log_info("Cancelled")
            sys.exit(0)
    
    if not restorer.check_postgres_running():
        log_error("PostgreSQL container is not running!")
        sys.exit(1)
    
    success, message = restorer.restore(force=True, keep_extra=keep_extra)
    
    if success:
        log_success("Database restore completed!")
    else:
        log_error(f"Database restore failed: {message}")
    
    sys.exit(0 if success else 1)


@cli.command()
@click.option("--force", is_flag=True)
@click.option("--dry-run", is_flag=True)
@click.option("--sync", is_flag=True, help="Sync to OneDrive after backup")
def backup(force: bool, dry_run: bool, sync: bool):
    """Backup database."""
    config = Config.from_env()
    backup_handler = DatabaseBackup(config)
    
    success, message = backup_handler.backup(force=force, dry_run=dry_run)
    
    if success:
        log_success(f"Backup completed: {message}")
        
        if sync and not dry_run:
            log_info("Syncing to OneDrive...")
            from omni_build.onedrive_sync import OneDriveSync
            sync_handler = OneDriveSync(config.project_root)
            sync_success, sync_message = sync_handler.sync_to_cloud()
            if sync_success:
                log_success(f"Sync completed: {sync_message}")
            else:
                log_warning(f"Sync issue: {sync_message}")
    else:
        log_error(f"Backup failed: {message}")
    
    sys.exit(0 if success else 1)


@cli.command()
def sync_restore():
    """Sync from OneDrive and restore database."""
    config = Config.from_env()
    
    log_info("Syncing from OneDrive...")
    from omni_build.onedrive_sync import OneDriveSync
    sync_handler = OneDriveSync(config.project_root)
    sync_success, sync_message = sync_handler.sync_from_cloud()
    
    if not sync_success:
        log_error(f"Sync failed: {sync_message}")
        sys.exit(1)
    
    log_success(f"Sync completed: {sync_message}")
    
    # Now restore
    restorer = DatabaseRestorer(config)
    
    if not restorer.has_backup():
        log_error("No backup found after sync!")
        sys.exit(1)
    
    restorer.print_backup_status()
    
    if not confirm("Continue with database restore?"):
        log_info("Cancelled")
        sys.exit(0)
    
    if not restorer.check_postgres_running():
        log_error("PostgreSQL container is not running!")
        sys.exit(1)
    
    success, message = restorer.restore(force=True)
    
    if success:
        log_success("Database restore completed!")
    else:
        log_error(f"Database restore failed: {message}")
    
    sys.exit(0 if success else 1)
