"""
Subprocess utilities with UTF-8 encoding support for Windows.

This module provides wrappers around subprocess functions that ensure
UTF-8 encoding is used instead of the Windows default (cp1252), preventing
UnicodeDecodeError when Docker or other tools output non-ASCII characters.
"""
import subprocess
from typing import Any, Optional


def run(
    args: list[str],
    capture_output: bool = False,
    text: bool = False,
    timeout: Optional[float] = None,
    cwd: Optional[str] = None,
    check: bool = False,
    shell: bool = False,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """
    Run a subprocess command with UTF-8 encoding on Windows.
    
    This is a wrapper around subprocess.run() that automatically sets
    encoding='utf-8' and errors='replace' when text=True is used,
    preventing UnicodeDecodeError on Windows.
    
    Args:
        args: Command and arguments to run
        capture_output: If True, capture stdout and stderr
        text: If True, decode stdout/stderr as text (with UTF-8 encoding)
        timeout: Timeout in seconds
        cwd: Working directory
        check: If True, raise CalledProcessError on non-zero exit
        shell: If True, run through shell
        **kwargs: Additional arguments passed to subprocess.run()
    
    Returns:
        CompletedProcess instance
    """
    # When text=True is requested, use explicit UTF-8 encoding
    # with 'replace' error handling to avoid UnicodeDecodeError
    if text:
        kwargs.setdefault('encoding', 'utf-8')
        kwargs.setdefault('errors', 'replace')
    
    return subprocess.run(
        args,
        capture_output=capture_output,
        text=text,
        timeout=timeout,
        cwd=cwd,
        check=check,
        shell=shell,
        **kwargs,
    )


def run_utf8(
    args: list[str],
    timeout: Optional[float] = None,
    cwd: Optional[str] = None,
    check: bool = False,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """
    Convenience function to run a command and capture output as UTF-8 text.
    
    Equivalent to:
        subprocess.run(args, capture_output=True, text=True, 
                       encoding='utf-8', errors='replace', ...)
    
    Args:
        args: Command and arguments to run
        timeout: Timeout in seconds
        cwd: Working directory
        check: If True, raise CalledProcessError on non-zero exit
        **kwargs: Additional arguments passed to subprocess.run()
    
    Returns:
        CompletedProcess instance with stdout/stderr as strings
    """
    return run(
        args,
        capture_output=True,
        text=True,
        timeout=timeout,
        cwd=cwd,
        check=check,
        **kwargs,
    )


def run_silent(
    args: list[str],
    timeout: Optional[float] = None,
    cwd: Optional[str] = None,
    **kwargs: Any,
) -> subprocess.CompletedProcess:
    """
    Run a command silently, suppressing output.
    
    Args:
        args: Command and arguments to run
        timeout: Timeout in seconds
        cwd: Working directory
        **kwargs: Additional arguments passed to subprocess.run()
    
    Returns:
        CompletedProcess instance
    """
    return run(
        args,
        capture_output=True,
        text=True,
        timeout=timeout,
        cwd=cwd,
        check=False,
        **kwargs,
    )
