"""
Real-time Output Parser for error detection during builds.

SRP: Parse build output in real-time and detect errors as they happen.
"""
import re
from typing import Optional

from omni_build.logger import log_error, log_warning
from omni_build.models import ErrorPattern


class OutputParser:
    """
    Parses build output in real-time to detect errors.
    
    This allows us to detect and fix errors DURING build,
    not just after it fails.
    """
    
    def __init__(self, error_patterns: list[ErrorPattern]) -> None:
        """
        Initialize parser with error patterns.
        
        Args:
            error_patterns: List of error patterns to match
        """
        self.error_patterns = error_patterns
        self.detected_errors: list[ErrorPattern] = []
        self.output_buffer: list[str] = []
    
    def parse_line(self, line: str) -> Optional[ErrorPattern]:
        """
        Parse a single line of output for errors.
        
        Args:
            line: Output line to parse
            
        Returns:
            ErrorPattern if detected, None otherwise
        """
        self.output_buffer.append(line)
        
        # Check against all error patterns
        for pattern in self.error_patterns:
            if re.search(pattern.pattern, line, re.IGNORECASE):
                if pattern not in self.detected_errors:
                    self.detected_errors.append(pattern)
                    log_warning(f"Error detected in output: {pattern.description}")
                return pattern
        
        return None
    
    def get_buffered_output(self, last_n_lines: int = 50) -> str:
        """
        Get last N lines of buffered output.
        
        Args:
            last_n_lines: Number of lines to return
            
        Returns:
            Joined output string
        """
        return "\n".join(self.output_buffer[-last_n_lines:])
    
    def clear_buffer(self) -> None:
        """Clear output buffer and detected errors."""
        self.output_buffer.clear()
        self.detected_errors.clear()
    
    def get_most_critical_error(self) -> Optional[ErrorPattern]:
        """
        Get the most critical error detected.
        
        Returns:
            ErrorPattern with highest severity, or None
        """
        if not self.detected_errors:
            return None
        
        # Sort by severity (CRITICAL > HIGH > MEDIUM > LOW)
        severity_order = {"critical": 4, "high": 3, "medium": 2, "low": 1}
        
        sorted_errors = sorted(
            self.detected_errors,
            key=lambda e: severity_order.get(e.severity, 0),
            reverse=True,
        )
        
        return sorted_errors[0]
