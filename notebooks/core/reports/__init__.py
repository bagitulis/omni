#!/usr/bin/env python
# -*- coding: utf-8 -*-
"""
Reports Subpackage
==================
Report generation utilities.
Platform-agnostic base classes.

Note: Most report generators are platform-specific and live in
the respective platform folders (tiktok_ads/intelligence/, etc.)
This module provides shared utilities and base classes.
"""

from .base import (
    BaseReportGenerator,
    ReportMetadata,
    format_currency,
    format_percentage,
    format_number,
)

__all__ = [
    "BaseReportGenerator",
    "ReportMetadata",
    "format_currency",
    "format_percentage",
    "format_number",
]
