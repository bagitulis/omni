"""
Logging module for Omni Build System.

SRP: This module ONLY handles logging and console output.
Uses Rich library for beautiful, structured logging.
"""
import logging
import sys
from typing import Any, Optional

from rich.console import Console
from rich.logging import RichHandler
from rich.progress import BarColumn, Progress, SpinnerColumn, TextColumn, TimeElapsedColumn
from rich.theme import Theme


# Custom theme for Omni Build System
OMNI_THEME = Theme({
    "info": "cyan",
    "success": "bold green",
    "warning": "yellow",
    "error": "bold red",
    "critical": "bold white on red",
    "step": "bold blue",
    "fix": "bold magenta",
    "retry": "yellow",
})

# Global console instance
console = Console(theme=OMNI_THEME)


def setup_logger(name: str = "omni_build", level: int = logging.INFO) -> logging.Logger:
    """
    Setup logger with Rich handler.
    
    Args:
        name: Logger name
        level: Logging level (default: INFO)
        
    Returns:
        Configured logger instance
    """
    logger = logging.getLogger(name)
    logger.setLevel(level)
    
    # Remove existing handlers to avoid duplicates
    logger.handlers.clear()
    
    # Rich handler for beautiful console output
    rich_handler = RichHandler(
        console=console,
        show_time=True,
        show_path=False,
        markup=True,
        rich_tracebacks=True,
        tracebacks_show_locals=True,
    )
    rich_handler.setFormatter(logging.Formatter("%(message)s"))
    
    logger.addHandler(rich_handler)
    
    return logger


# Global logger instance
logger = setup_logger()


def log_info(message: str) -> None:
    """Log info message with [INFO] prefix."""
    console.print(f"[info][INFO][/info] {message}")


def log_success(message: str) -> None:
    """Log success message with [OK] prefix."""
    console.print(f"[success][OK][/success] {message}")


def log_warning(message: str) -> None:
    """Log warning message with [WARN] prefix."""
    console.print(f"[warning][WARN][/warning] {message}")


def log_error(message: str) -> None:
    """Log error message with [ERROR] prefix."""
    console.print(f"[error][ERROR][/error] {message}")


def log_critical(message: str) -> None:
    """Log critical error message."""
    console.print(f"[critical][CRITICAL][/critical] {message}")


def log_step(step: str, message: str) -> None:
    """Log build step with [STEP] prefix."""
    console.print(f"[step][{step}][/step] {message}")


def log_fix(message: str) -> None:
    """Log auto-fix message with [AUTO-FIX] prefix."""
    console.print(f"[fix][AUTO-FIX][/fix] {message}")


def log_retry(attempt: int, max_attempts: int, message: str = "Retrying...") -> None:
    """Log retry attempt with [RETRY] prefix."""
    console.print(f"[retry][RETRY {attempt}/{max_attempts}][/retry] {message}")


def log_header(title: str, char: str = "=", width: int = 60) -> None:
    """Print a header with centered title."""
    console.print()
    console.print(char * width, style="bold cyan")
    console.print(f"  {title}".ljust(width - 2), style="bold cyan")
    console.print(char * width, style="bold cyan")
    console.print()


def log_separator(width: int = 60) -> None:
    """Print a separator line."""
    console.print("─" * width, style="dim")


def create_progress() -> Progress:
    """
    Create a Rich progress bar.
    
    Returns:
        Progress instance for tracking long-running operations.
    """
    return Progress(
        SpinnerColumn(),
        TextColumn("[progress.description]{task.description}"),
        BarColumn(),
        TextColumn("[progress.percentage]{task.percentage:>3.0f}%"),
        TimeElapsedColumn(),
        console=console,
    )


def print_build_result(
    success: bool,
    mode: str,
    spec: str,
    duration: float,
    containers: int = 0,
    health_checks: int = 0,
) -> None:
    """
    Print build result summary.
    
    Args:
        success: Whether build succeeded
        mode: Build mode used
        spec: Spec level used
        duration: Total duration in seconds
        containers: Number of containers deployed
        health_checks: Number of health checks passed
    """
    log_separator()
    
    if success:
        console.print("\n[success]BUILD COMPLETED SUCCESSFULLY[/success]\n", justify="center")
        console.print(f"  Mode: {mode}")
        console.print(f"  Spec: {spec}")
        console.print(f"  Duration: {duration:.1f}s")
        if containers > 0:
            console.print(f"  Containers: {containers}")
        if health_checks > 0:
            console.print(f"  Health Checks: {health_checks}/{health_checks}")
    else:
        console.print("\n[error]BUILD FAILED[/error]\n", justify="center")
        console.print(f"  Mode: {mode}")
        console.print(f"  Spec: {spec}")
        console.print(f"  Duration: {duration:.1f}s")
    
    log_separator()


def print_error_details(errors: list[str], warnings: list[str]) -> None:
    """
    Print error and warning details.
    
    Args:
        errors: List of error messages
        warnings: List of warning messages
    """
    if errors:
        console.print("\n[error]Errors:[/error]")
        for error in errors:
            console.print(f"  - {error}", style="error")
    
    if warnings:
        console.print("\n[warning]Warnings:[/warning]")
        for warning in warnings:
            console.print(f"  - {warning}", style="warning")


def prompt_user(question: str, choices: list[str], default: Optional[str] = None) -> str:
    """
    Prompt user for input with choices.
    
    Args:
        question: Question to ask
        choices: List of valid choices
        default: Default choice if user presses Enter
        
    Returns:
        User's choice
    """
    choices_str = "/".join(choices)
    if default:
        choices_str = choices_str.replace(default, f"[bold]{default}[/bold]")
        prompt_text = f"{question} [{choices_str}] (default: {default}): "
    else:
        prompt_text = f"{question} [{choices_str}]: "
    
    while True:
        console.print(prompt_text, end="")
        user_input = input().strip().upper()
        
        if not user_input and default:
            return default.upper()
        
        if user_input in [c.upper() for c in choices]:
            return user_input
        
        console.print(f"[warning]Invalid choice. Please enter one of: {choices_str}[/warning]")


def confirm(question: str, default: bool = False) -> bool:
    """
    Ask user for yes/no confirmation.
    
    Args:
        question: Question to ask
        default: Default answer
        
    Returns:
        True if yes, False if no
    """
    choices = ["Y", "N"]
    default_choice = "Y" if default else "N"
    
    result = prompt_user(question, choices, default_choice)
    return result == "Y"
