#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Base Report Generator Module
============================
Shared utilities and base classes for report generation.
Platform-agnostic.
"""

from datetime import datetime
from typing import Dict, Any, Optional
from dataclasses import dataclass


@dataclass
class ReportMetadata:
    """Report metadata container."""
    title: str
    generated_at: datetime
    platform: str = "generic"
    version: str = "1.0.0"
    author: str = "Omni Analytics"
    
    def to_dict(self) -> Dict[str, Any]:
        return {
            "title": self.title,
            "generated_at": self.generated_at.isoformat(),
            "platform": self.platform,
            "version": self.version,
            "author": self.author,
        }


class BaseReportGenerator:
    """
    Base class for report generators.
    
    Subclasses should implement:
    - generate(): Main generation method
    - _build_sections(): Build report sections
    """
    
    def __init__(self, title: str = "Report"):
        self.title = title
        self.metadata = ReportMetadata(
            title=title,
            generated_at=datetime.now()
        )
        self.sections = []
    
    def generate(self) -> str:
        """Generate report. Override in subclass."""
        raise NotImplementedError("Subclasses must implement generate()")
    
    def add_section(self, title: str, content: str) -> None:
        """Add a section to the report."""
        self.sections.append({
            "title": title,
            "content": content
        })
    
    def get_timestamp(self) -> str:
        """Get formatted timestamp."""
        return datetime.now().strftime("%Y-%m-%d %H:%M:%S")


# ============================================================================
# Formatting Utilities
# ============================================================================

def format_currency(
    value: float,
    prefix: str = "Rp",
    decimal_places: int = 0
) -> str:
    """
    Format number as Indonesian currency.
    
    Args:
        value: Numeric value
        prefix: Currency prefix (default "Rp")
        decimal_places: Number of decimal places
        
    Returns:
        str: Formatted currency string
    """
    if value >= 1_000_000_000:
        return f"{prefix} {value / 1_000_000_000:.{decimal_places}f}B"
    elif value >= 1_000_000:
        return f"{prefix} {value / 1_000_000:.{decimal_places}f}M"
    elif value >= 1_000:
        return f"{prefix} {value / 1_000:.{decimal_places}f}K"
    else:
        return f"{prefix} {value:,.{decimal_places}f}"


def format_percentage(
    value: float,
    decimal_places: int = 1,
    include_sign: bool = False
) -> str:
    """
    Format number as percentage.
    
    Args:
        value: Numeric value (0.1 = 10%)
        decimal_places: Number of decimal places
        include_sign: Include + for positive values
        
    Returns:
        str: Formatted percentage string
    """
    pct = value * 100
    sign = "+" if include_sign and pct > 0 else ""
    return f"{sign}{pct:.{decimal_places}f}%"


def format_number(
    value: float,
    decimal_places: int = 2,
    use_abbreviation: bool = True
) -> str:
    """
    Format number with optional abbreviation.
    
    Args:
        value: Numeric value
        decimal_places: Number of decimal places
        use_abbreviation: Use K/M/B abbreviations
        
    Returns:
        str: Formatted number string
    """
    if not use_abbreviation:
        return f"{value:,.{decimal_places}f}"
    
    if abs(value) >= 1_000_000_000:
        return f"{value / 1_000_000_000:.{decimal_places}f}B"
    elif abs(value) >= 1_000_000:
        return f"{value / 1_000_000:.{decimal_places}f}M"
    elif abs(value) >= 1_000:
        return f"{value / 1_000:.{decimal_places}f}K"
    else:
        return f"{value:.{decimal_places}f}"


def get_color_for_value(
    value: float,
    thresholds: Dict[str, float] = None
) -> str:
    """
    Get color code based on value thresholds.
    
    Args:
        value: Numeric value
        thresholds: Custom thresholds (default: good=70, warning=50)
        
    Returns:
        str: Color code (green/yellow/red)
    """
    if thresholds is None:
        thresholds = {"good": 70, "warning": 50}
    
    if value >= thresholds["good"]:
        return "green"
    elif value >= thresholds["warning"]:
        return "yellow"
    else:
        return "red"


def get_trend_arrow(trend: str) -> str:
    """
    Get arrow symbol for trend direction.
    
    Args:
        trend: Trend direction string
        
    Returns:
        str: Arrow symbol
    """
    arrows = {
        "increasing": "↑",
        "up": "↑",
        "positive": "↑",
        "stable": "→",
        "flat": "→",
        "decreasing": "↓",
        "down": "↓",
        "negative": "↓",
    }
    return arrows.get(trend.lower(), "•")
